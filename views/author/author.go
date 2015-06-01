package admin

import (
	"errors"
	// "fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
)

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

import (
	"models"
	"x"
)

var homeTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/home.html"))
var registerTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/register.html"))
var newPostSeriesTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/newpostseries.html"))
var editPostSeriesTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/editpostseries.html"))
var newPostTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/newpost.html"))
var editPostTmpl = template.Must(template.ParseFiles("templates/base.html", "templates/author/editpost.html"))

// Author home page.
func Home(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	series := make([]models.PostSeries, 0, 10) // ** Plural series
	q := datastore.NewQuery("PostSeries").Filter("Author =", pageData.Author.Key).Order("-DateAdded").Limit(10)
	if n, err := q.Count(c); err == nil {
		log.Printf("Number of series:%v", n)
		if keys, err := q.GetAll(c, &series); err == nil {
			var postSeries []map[string]template.HTML
			for i, s := range series {
				t := make(map[string]template.HTML)
				t["Title"] = template.HTML(s.Title)
				t["JoinedTitle"] = template.HTML(s.JoinedTitle)
				t["AbstractHTML"] = template.HTML(s.AbstractHTML)
				t["DateAdded"] = template.HTML(s.DateAdded.Format("January 2, 2006"))
				q := datastore.NewQuery("Post").Filter("Series =", keys[i])
				if n, err := q.Count(c); err == nil {
					t["NumberOfPosts"] = template.HTML(strconv.Itoa(n))
				} else {
					t["NumberOfPosts"] = template.HTML("0")
				}
				postSeries = append(postSeries, t)
			}
			pageData.Misc["AuthorPostSeries"] = postSeries
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	posts := make([]models.Post, 0, 20)
	q = datastore.NewQuery("Post").Filter("Author =", pageData.Author.Key).Order("-DateAdded").Limit(20)
	if _, err := q.Count(c); err == nil {
		if _, err := q.GetAll(c, &posts); err == nil {
			var authorPosts []map[string]template.HTML
			for _, p := range posts {
				post := make(map[string]template.HTML)
				post["Title"] = template.HTML(p.Title)
				post["JoinedTitle"] = template.HTML(p.JoinedTitle)
				post["AbstractHTML"] = template.HTML(p.AbstractHTML)
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
					}
				}
				authorPosts = append(authorPosts, post)
			}
			pageData.Misc["AuthorPosts"] = authorPosts
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := homeTmpl.Execute(w, pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// http.Error(w, errors.New("What...").Error(), http.StatusInternalServerError)
}

// Register new author.
func Register(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)

	if r.Method == "GET" {
		// var author models.Author
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u != nil {
			author := new(models.Author)
			author.Email = u.Email
			pageData.Author = author
			if err := registerTmpl.Execute(w, pageData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			url, err := user.LoginURL(c, r.URL.String())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusFound)
		}
	} else if r.Method == "POST" {
		var author models.Author
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u != nil {
			author.Email = u.Email
			if firstname := r.PostFormValue("FirstName"); firstname != "" {
				author.FirstName = firstname
			}
			if lastname := r.PostFormValue("LastName"); lastname != "" {
				author.LastName = lastname
			}
			if othernames := r.PostFormValue("OtherNames"); othernames != "" {
				author.OtherNames = othernames
			}
			if biomd := r.PostFormValue("BioMD"); biomd != "" {
				author.BioMD = biomd
				author.BioHTML = string(blackfriday.MarkdownBasic([]byte(biomd)))
			}
			if _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Author", nil), &author); err == nil {
				w.Header().Set("Location", "/author/")
				w.WriteHeader(http.StatusFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

// Add a new post series.
func NewPostSeries(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	if r.Method == "GET" {
		if err := newPostSeriesTmpl.Execute(w, pageData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		var series models.PostSeries
		if title := r.PostFormValue("Title"); title != "" {
			series.Title = title
			series.JoinedTitle = strings.Join(strings.Split(strings.ToLower(title), " "), "-")
		}
		if abstractmd := r.PostFormValue("AbstractMD"); abstractmd != "" {
			series.AbstractMD = abstractmd
			series.AbstractHTML = string(blackfriday.MarkdownBasic([]byte(abstractmd)))
		}
		if tagstr := r.PostFormValue("Tags"); tagstr != "" {
			tags := strings.Split(tagstr, ",")
			for i, tag := range tags {
				tags[i] = strings.Trim(tag, " \t")
			}
			series.Tags = tags
		}
		series.Author = pageData.Author.Key
		series.DateAdded = time.Now()

		if _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "PostSeries", nil), &series); err == nil {
			http.Redirect(w, r, "/author/", http.StatusFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Add a new post.
func NewPost(w http.ResponseWriter, r *http.Request) {
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	if r.Method == "GET" {
		if err := newPostTmpl.Execute(w, pageData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		var post models.Post
		if title := r.PostFormValue("Title"); title != "" {
			post.Title = title
			post.JoinedTitle = strings.Join(strings.Split(strings.ToLower(title), " "), "-")
		}
		if abstractmd := r.PostFormValue("AbstractMD"); abstractmd != "" {
			post.AbstractMD = abstractmd
			post.AbstractHTML = string(blackfriday.MarkdownBasic([]byte(abstractmd)))
		}
		if bodymd := r.PostFormValue("BodyMD"); bodymd != "" {
			post.BodyMD = bodymd
			post.BodyHTML = string(blackfriday.MarkdownBasic([]byte(bodymd)))
		}
		if tagstr := r.PostFormValue("Tags"); tagstr != "" {
			tags := strings.Split(tagstr, ",")
			for i, tag := range tags {
				tags[i] = strings.Trim(tag, " \t")
			}
			post.Tags = tags
		}

		post.Author = pageData.Author.Key
		post.DateAdded = time.Now()
		post.YearAdded = time.Now().Year()
		post.MonthAdded = int(time.Now().Month())
		post.DayAdded = time.Now().Day()

		if _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Post", nil), &post); err == nil {
			http.Redirect(w, r, "/author/", http.StatusFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Add a new post to series.
func NewPostInSeries(w http.ResponseWriter, r *http.Request) {
	pagevars := mux.Vars(r)
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	pageData.Misc["SeriesJTitle"] = pagevars["seriesjtitle"]
	if r.Method == "GET" {
		if err := newPostTmpl.Execute(w, pageData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		q := datastore.NewQuery("PostSeries").Filter("JoinedTitle = ", pagevars["seriesjtitle"]).Limit(10)
		if n, err := q.Count(c); err == nil && n == 1 {
			var postseries = make([]models.PostSeries, 0, 1)
			if keys, err := q.GetAll(c, &postseries); err == nil {
				serieskey := keys[0]
				var post models.Post
				post.Series = serieskey
				if title := r.PostFormValue("Title"); title != "" {
					post.Title = title
					post.JoinedTitle = strings.Join(strings.Split(strings.ToLower(title), " "), "-")
				}
				if abstractmd := r.PostFormValue("AbstractMD"); abstractmd != "" {
					post.AbstractMD = abstractmd
					post.AbstractHTML = string(blackfriday.MarkdownBasic([]byte(abstractmd)))
				}
				if bodymd := r.PostFormValue("BodyMD"); bodymd != "" {
					post.BodyMD = bodymd
					post.BodyHTML = string(blackfriday.MarkdownBasic([]byte(bodymd)))
				}
				if tagstr := r.PostFormValue("Tags"); tagstr != "" {
					tags := strings.Split(tagstr, ",")
					for i, tag := range tags {
						tags[i] = strings.Trim(tag, " \t")
					}
					post.Tags = tags
				}
				q := datastore.NewQuery("Post").Filter("Series = ", serieskey)
				if n, err := q.Count(c); err == nil {
					post.PositionInSeries = n + 1
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				post.Author = pageData.Author.Key
				post.DateAdded = time.Now()
				post.YearAdded = time.Now().Year()
				post.MonthAdded = int(time.Now().Month())
				post.DayAdded = time.Now().Day()

				if _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Post", nil), &post); err == nil {
					http.Redirect(w, r, "/author/", http.StatusFound)
					return
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func EditSeries(w http.ResponseWriter, r *http.Request) {
	pagevars := mux.Vars(r)
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)
	q := datastore.NewQuery("PostSeries").Filter("JoinedTitle = ", pagevars["seriesjtitle"])
	if n, err := q.Count(c); err == nil && n == 1 {
		t := q.Run(c)
		var s models.PostSeries
		k, err := t.Next(&s)
		if err == nil {
			postseries := make(map[string]template.HTML)

			postseries["Title"] = template.HTML(s.Title)
			postseries["JoinedTitle"] = template.HTML(s.JoinedTitle)
			postseries["AbstractMD"] = template.HTML(s.AbstractMD)
			// postseries["AbstractHTML"] = template.HTML(s.AbstractHTML)
			postseries["Tags"] = template.HTML(strings.Join(s.Tags, ", "))
			// postseries["DateAdded"] = template.HTML(s.DateAdded.Format("January 2, 2006"))

			pageData.Misc["PostSeries"] = postseries
		}
		if r.Method == "GET" {
			if err := editPostSeriesTmpl.Execute(w, pageData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if r.Method == "POST" {
			if title := r.PostFormValue("Title"); title != "" {
				if s.Title != title {
					s.Title = title
					s.JoinedTitle = strings.Join(strings.Split(strings.ToLower(title), " "), "-")
				}
			}
			if abstractmd := r.PostFormValue("AbstractMD"); abstractmd != "" {
				if s.AbstractMD != abstractmd {
					s.AbstractMD = abstractmd
					s.AbstractHTML = string(blackfriday.MarkdownBasic([]byte(abstractmd)))
				}
			}
			if tagstr := r.PostFormValue("Tags"); tagstr != "" {
				tags := strings.Split(tagstr, ",")
				for i, tag := range tags {
					tags[i] = strings.Trim(tag, " \t")
					inTags := false
					for _, tag := range s.Tags {
						if tags[i] == tag {
							inTags = true
						}
					}
					if !inTags {
						s.Tags = append(s.Tags, tag)
					}
				}
			}
			s.LastEdited = time.Now()
			if _, err := datastore.Put(c, k, &s); err == nil {
				w.Header().Set("Location", "/author/")
				w.WriteHeader(http.StatusFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Error(w, errors.New("There was an error retrieving the post series details.").Error(), http.StatusInternalServerError)
			return
		}
	}
}

func EditPost(w http.ResponseWriter, r *http.Request) {
	pagevars := mux.Vars(r)
	pageData := context.Get(r, "PageData").(*x.PageData)
	c := appengine.NewContext(r)

	// Retrieve Post from Data Store.
	year, _ := strconv.Atoi(pagevars["year"])
	month, _ := strconv.Atoi(pagevars["month"])
	day, _ := strconv.Atoi(pagevars["day"])
	q := datastore.NewQuery("Post").
		Filter("JoinedTitle = ", pagevars["jtitle"]).
		Filter("Author = ", pageData.Author.Key).
		Filter("YearAdded = ", year).
		Filter("MonthAdded = ", month).
		Filter("DayAdded = ", day)
	if n, err := q.Count(c); err == nil && n == 1 {
		t := q.Run(c)
		var p models.Post
		k, err := t.Next(&p)
		if err == nil {
			post := make(map[string]template.HTML)

			post["Title"] = template.HTML(p.Title)
			post["JoinedTitle"] = template.HTML(p.JoinedTitle)
			post["AbstractHTML"] = template.HTML(p.AbstractHTML)
			post["AbstractMD"] = template.HTML(p.AbstractMD)
			post["BodyHTML"] = template.HTML(p.BodyHTML)
			post["BodyMD"] = template.HTML(p.BodyMD)
			post["Tags"] = template.HTML(strings.Join(p.Tags, ", "))
			post["DateAdded"] = template.HTML(p.DateAdded.Format("January 2, 2006"))
			post["YearAdded"] = template.HTML(p.DateAdded.Format("2006"))
			post["MonthAdded"] = template.HTML(p.DateAdded.Format("1"))
			post["DayAdded"] = template.HTML(p.DateAdded.Format("2"))

			pageData.Post = post
		}
		if r.Method == "GET" {
			if err := editPostTmpl.Execute(w, pageData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if r.Method == "POST" {
			if title := r.PostFormValue("Title"); title != "" {
				if p.Title != title {
					p.Title = title
					p.JoinedTitle = strings.Join(strings.Split(strings.ToLower(title), " "), "-")
				}
			}
			if abstractmd := r.PostFormValue("AbstractMD"); abstractmd != "" {
				if p.AbstractMD != abstractmd {
					p.AbstractMD = abstractmd
					p.AbstractHTML = string(blackfriday.MarkdownBasic([]byte(abstractmd)))
				}
			}
			if bodymd := r.PostFormValue("BodyMD"); bodymd != "" {
				if p.BodyMD != bodymd {
					p.BodyMD = bodymd
					p.BodyHTML = string(blackfriday.MarkdownBasic([]byte(bodymd)))
				}
			}
			if tagstr := r.PostFormValue("Tags"); tagstr != "" {
				tags := strings.Split(tagstr, ",")
				for i, tag := range tags {
					tags[i] = strings.Trim(tag, " \t")
					inTags := false
					for _, tag := range p.Tags {
						if tags[i] == tag {
							inTags = true
						}
					}
					if !inTags {
						p.Tags = append(p.Tags, tag)
					}
				}
			}
			p.LastEdited = time.Now()
			if _, err := datastore.Put(c, k, &p); err == nil {
				w.Header().Set("Location", "/author/")
				w.WriteHeader(http.StatusFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}
