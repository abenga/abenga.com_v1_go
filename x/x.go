package x

import (
	"html/template"
)

import (
	"models"
)

type PageData struct {
	Author     *models.Author
	Misc       map[string]interface{}
	PostSeries map[string]template.HTML
	Post       map[string]template.HTML
	LogoutURL  string
}
