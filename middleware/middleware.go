package middleware

import (
	"net/http"
)

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
)

import (
	"github.com/gorilla/context"
)

import (
	"models"
	"x"
)

// Set up the page variables struct. If the request is being made by an
// authenticated author. If this is true, we set the PageData author field.
func Initialize(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		pageData := new(x.PageData)
		pageData.Misc = make(map[string]interface{})
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u != nil {
			logouturl, _ := user.LogoutURL(c, "/")
			pageData.LogoutURL = logouturl
			q := datastore.NewQuery("Author").Filter("Email = ", u.Email)
			authors := make([]models.Author, 0)
			if n, err := q.Count(c); err == nil && n == 1 {
				if keys, err := q.GetAll(c, &authors); err == nil {
					if len(authors) == 1 {
						pageData.Author = &authors[0]
						pageData.Author.Key = keys[0]
					}
				}
			}
		}
		context.Set(r, "PageData", pageData)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

// If the author has not been authorized, redirect the user to a login page.
func CheckAuth(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		loginurl, _ := user.LoginURL(c, r.URL.String())
		pageData := context.Get(r, "PageData").(*x.PageData)
		if pageData.Author != nil {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("Location", loginurl)
			w.WriteHeader(http.StatusFound)
		}
		context.Set(r, "PageData", pageData)
	}
	return http.HandlerFunc(handler)
}
