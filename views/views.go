package views

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

import (
	"appengine"
	"appengine/datastore"
)

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

import (
	"models"
	"x"
)

var indexTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
var postSeriesTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/postseries.html"))
var postTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/post.html"))

func Index(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)

	postmodels := make([]models.Post, 0, 12)
	q := datastore.NewQuery("Post").Order("-DateAdded").Limit(12)
	if n, err := q.Count(c); err == nil && n > 0 {
		if keys, err := q.GetAll(c, &postmodels); err == nil {
			posts := make([]map[string]template.HTML, 0)
			for i, p := range postmodels {
				post := make(map[string]template.HTML)
				post["Title"] = template.HTML(p.Title)
				post["JoinedTitle"] = template.HTML(p.JoinedTitle)
				post["PositionInSeries"] = template.HTML(p.PositionInSeries)
				post["AbstractHTML"] = template.HTML(p.AbstractHTML)
				post["DateAdded"] = template.HTML(p.DateAdded.Format("January 2, 2006"))
				post["YearAdded"] = template.HTML(p.DateAdded.Format("2006"))
				post["MonthAdded"] = template.HTML(p.DateAdded.Format("1"))
				post["DayAdded"] = template.HTML(p.DateAdded.Format("2"))
				post["Key"] = template.HTML(keys[i].String())
				if p.Series != nil {
					var series models.PostSeries
					if err = datastore.Get(c, p.Series, &series); err == nil {
						post["PositionInSeries"] = template.HTML(strconv.Itoa(p.PositionInSeries))
						post["SeriesTitle"] = template.HTML(series.Title)
						post["SeriesJoinedTitle"] = template.HTML(series.JoinedTitle)
					}
				}
				posts = append(posts, post)
			}
			pageData.Misc["Posts"] = posts
		}
	}

	if err := indexTmpl.Execute(w, pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostSeries(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	pagevars := mux.Vars(r)
	posts := make([]map[string]template.HTML, 0)
	q := datastore.NewQuery("PostSeries").Filter("JoinedTitle = ", pagevars["seriesjtitle"]).Limit(10)
	if n, err := q.Count(c); err == nil && n == 1 {
		var postseries = make([]models.PostSeries, 0, 1)
		if keys, err := q.GetAll(c, &postseries); err == nil {
			serieskey := keys[0]
			s := make(map[string]template.HTML)

			s["Title"] = template.HTML(postseries[0].Title)
			s["JoinedTitle"] = template.HTML(postseries[0].JoinedTitle)
			s["AbstractMD"] = template.HTML(postseries[0].AbstractMD)
			s["AbstractHTML"] = template.HTML(postseries[0].AbstractHTML)
			s["Tags"] = template.HTML(strings.Join(postseries[0].Tags, ", "))

			pageData.Misc["PostSeries"] = s

			postmodels := make([]models.Post, 0)
			q := datastore.NewQuery("Post").Filter("Series = ", serieskey).Order("-DateAdded")
			if _, err := q.Count(c); err == nil {
				n, err = q.Count(c)
				if keys, err := q.GetAll(c, &postmodels); err == nil {
					// posts := make([]map[string]template.HTML, 0)
					for i, p := range postmodels {
						post := make(map[string]template.HTML)
						post["PositionInSeries"] = template.HTML(strconv.Itoa(p.PositionInSeries))
						post["Title"] = template.HTML(p.Title)
						post["JoinedTitle"] = template.HTML(p.JoinedTitle)
						post["AbstractHTML"] = template.HTML(p.AbstractHTML)
						post["DateAdded"] = template.HTML(p.DateAdded.Format("January 2, 2006"))
						post["YearAdded"] = template.HTML(p.DateAdded.Format("2006"))
						post["MonthAdded"] = template.HTML(p.DateAdded.Format("1"))
						post["DayAdded"] = template.HTML(p.DateAdded.Format("2"))
						post["Key"] = template.HTML(keys[i].String())
						if p.Series != nil {
							var series models.PostSeries
							if err = datastore.Get(c, p.Series, &series); err == nil {
								post["SeriesTitle"] = template.HTML(series.Title)
								post["SeriesJoinedTitle"] = template.HTML(series.JoinedTitle)
							}
						}
						posts = append(posts, post)
					}
					pageData.Misc["Posts"] = posts
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				} else {
					pageData.Misc["Posts"] = nil
				}
			}
			if err := postSeriesTmpl.Execute(w, pageData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		return
	} else {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Error(w, errors.New("There was an error retrieving the post series data.").Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Post(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	pagevars := mux.Vars(r)
	year, _ := strconv.Atoi(pagevars["year"])
	month, _ := strconv.Atoi(pagevars["month"])
	day, _ := strconv.Atoi(pagevars["day"])
	q := datastore.NewQuery("Post").
		Filter("JoinedTitle = ", pagevars["jtitle"]).
		Filter("YearAdded = ", year).
		Filter("MonthAdded = ", month).
		Filter("DayAdded = ", day)
	if n, err := q.Count(c); err == nil && n == 1 {
		t := q.Run(c)
		var p models.Post
		_, err := t.Next(&p)
		if err == nil {
			post := make(map[string]template.HTML)
			post["Title"] = template.HTML(p.Title)
			post["JoinedTitle"] = template.HTML(p.JoinedTitle)
			post["AbstractHTML"] = template.HTML(p.AbstractHTML)
			post["BodyHTML"] = template.HTML(p.BodyHTML)
			post["DateAdded"] = template.HTML(p.DateAdded.Format("January 2, 2006"))
			post["YearAdded"] = template.HTML(p.DateAdded.Format("2006"))
			post["MonthAdded"] = template.HTML(p.DateAdded.Format("1"))
			post["DayAdded"] = template.HTML(p.DateAdded.Format("2"))
			if p.Series != nil {
				var series models.PostSeries
				if err = datastore.Get(c, p.Series, &series); err == nil {
					post["PositionInSeries"] = template.HTML(strconv.Itoa(p.PositionInSeries))
					post["SeriesTitle"] = template.HTML(series.Title)
					post["SeriesJoinedTitle"] = template.HTML(series.JoinedTitle)
					post["SeriesAbstract"] = template.HTML(series.AbstractHTML)
					if p.PositionInSeries > 1 {
						var prevpost models.Post
						// log.Printf("PositionInSeries = %v", p.PositionInSeries-1)
						q := datastore.NewQuery("Post").
							Filter("Series = ", p.Series).
							Filter("PositionInSeries = ", p.PositionInSeries-1)
						if n, err := q.Count(c); err == nil && n == 1 {
							t = q.Run(c)
							t.Next(&prevpost)
							post["PreviousPostPositionInSeries"] = template.HTML(strconv.Itoa(prevpost.PositionInSeries))
							post["PreviousPostTitle"] = template.HTML(prevpost.Title)
							post["PreviousPostJoinedTitle"] = template.HTML(prevpost.JoinedTitle)
							post["PreviousPostYearAdded"] = template.HTML(prevpost.DateAdded.Format("2006"))
							post["PreviousPostMonthAdded"] = template.HTML(prevpost.DateAdded.Format("1"))
							post["PreviousPostDayAdded"] = template.HTML(prevpost.DateAdded.Format("2"))
						}
					}
					q := datastore.NewQuery("Post").
						Filter("Series = ", p.Series).
						Filter("PositionInSeries = ", p.PositionInSeries+1)
					if n, err := q.Count(c); err == nil && n == 1 {
						var nextpost models.Post
						t = q.Run(c)
						t.Next(&nextpost)
						post["NextPostPositionInSeries"] = template.HTML(strconv.Itoa(nextpost.PositionInSeries))
						post["NextPostTitle"] = template.HTML(nextpost.Title)
						post["NextPostJoinedTitle"] = template.HTML(nextpost.JoinedTitle)
						post["NextPostYearAdded"] = template.HTML(nextpost.DateAdded.Format("2006"))
						post["NextPostMonthAdded"] = template.HTML(nextpost.DateAdded.Format("1"))
						post["NextPostDayAdded"] = template.HTML(nextpost.DateAdded.Format("2"))
					}
				}
			}
			pageData.Post = post
		}
		if err := postTmpl.Execute(w, pageData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, errors.New("There was an error retrieving the post.").Error(), http.StatusInternalServerError)
		}
		return
	}
}
