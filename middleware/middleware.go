package middleware

import (
	"net/http"
)

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
		ctx := appengine.NewContext(r)
		sessionID, _ := x.GetActiveSession(w, r)

		pageData := new(x.PageData)
		pageData.Misc = make(map[string]interface{})

		if sessionID != "" {
			loginSessions := make([]models.LoginSession, 0)
			q := datastore.NewQuery("LoginSession").Filter("SessionID = ", sessionID)
			_, err := q.GetAll(ctx, &loginSessions)
			if err == nil && len(loginSessions) == 1 {
				q := datastore.NewQuery("Author").Filter("Email = ", loginSessions[0].AuthorEmail)
				authors := make([]models.Author, 0)
				if keys, err := q.GetAll(ctx, &authors); err == nil {
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
		pageData := context.Get(r, "PageData").(*x.PageData)

		if pageData.Author != nil {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("Location", "/author/sign_in/")
			w.WriteHeader(http.StatusFound)
		}
		context.Set(r, "PageData", pageData)
	}
	return http.HandlerFunc(handler)
}
