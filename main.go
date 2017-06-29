package abenga

import (
	"net/http"
)

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"


)

import (
	mw "middleware"
	"views"
	author "views/author"
)

func init() {
	r := mux.NewRouter()
	plainMid := alice.New(context.ClearHandler, mw.Initialize)
	authMid := alice.New(context.ClearHandler, mw.Initialize, mw.CheckAuth)

	// Viewable without login
	r.Handle("/", plainMid.ThenFunc(views.Index))
	r.Handle("/postseries/{seriesjtitle}/", plainMid.ThenFunc(views.PostSeries))
	r.Handle("/post/{year}/{month}/{day}/{jtitle}/", plainMid.ThenFunc(views.Post))
	r.Handle("/author/sign_in/", plainMid.ThenFunc(author.SignIn))
	r.Handle("/author/sign_in_google_callback/", plainMid.ThenFunc(author.GoogleCallback))


	// Author pages - viewable only by authorized author.
	r.Handle("/author/", authMid.ThenFunc(author.Home))
	r.Handle("/author/newpost/", authMid.ThenFunc(author.NewPost))
	r.Handle("/author/newpostseries/", authMid.ThenFunc(author.NewPostSeries))
	r.Handle("/author/postseries/{seriesjtitle}/newpost/", authMid.ThenFunc(author.NewPostInSeries))
	r.Handle("/author/postseries/{seriesjtitle}/edit/", authMid.ThenFunc(author.EditSeries))
	r.Handle("/author/post/{year}/{month}/{day}/{jtitle}/edit/", authMid.ThenFunc(author.EditPost))

	http.Handle("/", r)
}
