package views

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

import (
	"models"
	"x"
)

func Index(w http.ResponseWriter, r *http.Request) {
	var indexTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))

	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	postModels := make([]models.Post, 0, 12)
	q := datastore.NewQuery("Post").Order("-DateAdded").Limit(12)

	keys, _ := q.GetAll(c, &postModels)
	posts := make([]map[string]template.HTML, 0)
	for i, p := range postModels {
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
			datastore.Get(c, p.Series, &series)
			post["PositionInSeries"] = template.HTML(strconv.Itoa(p.PositionInSeries))
			post["SeriesTitle"] = template.HTML(series.Title)
			post["SeriesJoinedTitle"] = template.HTML(series.JoinedTitle)
		}
		posts = append(posts, post)
	}
	pageData.Misc["Posts"] = posts
	err := indexTmpl.Execute(w, pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostSeries(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	pageVars := mux.Vars(r)

	posts := make([]map[string]template.HTML, 0)
	q := datastore.NewQuery("PostSeries").Filter("JoinedTitle = ", pageVars["seriesjtitle"]).Limit(10)

	var postSeriesTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/postSeries.html"))

	var postSeries = make([]models.PostSeries, 0, 1)
	keys, err := q.GetAll(c, &postSeries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(keys) != 1 {
		http.Error(w, errors.New("There was an error retrieving the post series data.").Error(), http.StatusInternalServerError)
	}
	seriesKey := keys[0]
	s := make(map[string]template.HTML)

	s["Title"] = template.HTML(postSeries[0].Title)
	s["JoinedTitle"] = template.HTML(postSeries[0].JoinedTitle)
	s["AbstractMD"] = template.HTML(postSeries[0].AbstractMD)
	s["AbstractHTML"] = template.HTML(postSeries[0].AbstractHTML)
	s["Tags"] = template.HTML(strings.Join(postSeries[0].Tags, ", "))

	pageData.Misc["PostSeries"] = s

	postModels := make([]models.Post, 0)

	q = datastore.NewQuery("Post").Filter("Series = ", seriesKey).Order("-DateAdded")

	postKeys, err := q.GetAll(c, &postModels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, p := range postModels {
		post := make(map[string]template.HTML)
		post["PositionInSeries"] = template.HTML(strconv.Itoa(p.PositionInSeries))
		post["Title"] = template.HTML(p.Title)
		post["JoinedTitle"] = template.HTML(p.JoinedTitle)
		post["AbstractHTML"] = template.HTML(p.AbstractHTML)
		post["DateAdded"] = template.HTML(p.DateAdded.Format("January 2, 2006"))
		post["YearAdded"] = template.HTML(p.DateAdded.Format("2006"))
		post["MonthAdded"] = template.HTML(p.DateAdded.Format("1"))
		post["DayAdded"] = template.HTML(p.DateAdded.Format("2"))
		post["Key"] = template.HTML(postKeys[i].String())
		if p.Series != nil {
			var series models.PostSeries
			if err = datastore.Get(c, p.Series, &series); err == nil {
				post["SeriesTitle"] = template.HTML(series.Title)
				post["SeriesJoinedTitle"] = template.HTML(series.JoinedTitle)
			}
		}
		posts = append(posts, post)
	}
	if err := postSeriesTmpl.Execute(w, pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Post(w http.ResponseWriter, r *http.Request) {

	var postTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/post.html"))

	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	pageVars := mux.Vars(r)

	year, _ := strconv.Atoi(pageVars["year"])
	month, _ := strconv.Atoi(pageVars["month"])
	day, _ := strconv.Atoi(pageVars["day"])

	q := datastore.NewQuery("Post").
		Filter("JoinedTitle = ", pageVars["jtitle"]).
		Filter("YearAdded = ", year).
		Filter("MonthAdded = ", month).
		Filter("DayAdded = ", day)

	t := q.Run(c)
	var p models.Post
	_, err := t.Next(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
					var previousPost models.Post
					// log.Printf("PositionInSeries = %v", p.PositionInSeries-1)
					q := datastore.NewQuery("Post").
						Filter("Series = ", p.Series).
						Filter("PositionInSeries = ", p.PositionInSeries-1)
					if n, err := q.Count(c); err == nil && n == 1 {
						t = q.Run(c)
						t.Next(&previousPost)
						post["PreviousPostPositionInSeries"] = template.HTML(strconv.Itoa(previousPost.PositionInSeries))
						post["PreviousPostTitle"] = template.HTML(previousPost.Title)
						post["PreviousPostJoinedTitle"] = template.HTML(previousPost.JoinedTitle)
						post["PreviousPostYearAdded"] = template.HTML(previousPost.DateAdded.Format("2006"))
						post["PreviousPostMonthAdded"] = template.HTML(previousPost.DateAdded.Format("1"))
						post["PreviousPostDayAdded"] = template.HTML(previousPost.DateAdded.Format("2"))
					}
				}
				q := datastore.NewQuery("Post").
					Filter("Series = ", p.Series).
					Filter("PositionInSeries = ", p.PositionInSeries+1)
				if n, err := q.Count(c); err == nil && n == 1 {
					var nextPost models.Post
					t = q.Run(c)
					t.Next(&nextPost)
					post["NextPostPositionInSeries"] = template.HTML(strconv.Itoa(nextPost.PositionInSeries))
					post["NextPostTitle"] = template.HTML(nextPost.Title)
					post["NextPostJoinedTitle"] = template.HTML(nextPost.JoinedTitle)
					post["NextPostYearAdded"] = template.HTML(nextPost.DateAdded.Format("2006"))
					post["NextPostMonthAdded"] = template.HTML(nextPost.DateAdded.Format("1"))
					post["NextPostDayAdded"] = template.HTML(nextPost.DateAdded.Format("2"))
				}
			}
		}
		pageData.Post = post
	}

	if err := postTmpl.Execute(w, pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
