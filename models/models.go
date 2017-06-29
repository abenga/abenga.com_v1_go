package models

import (
	"time"
)

import (
	"google.golang.org/appengine/datastore"
)

type Author struct {
	Key					*datastore.Key
	Email				string
	FirstName			string
	LastName			string
	OtherNames			string
	BioMD				string
	BioHTML				string
}

type LoginSession struct {
	AuthorEmail		string
	SessionID		string
	DateStarted		time.Time
}

type PostSeries struct {
	Title			string
	JoinedTitle		string
	AbstractMD		string `datastore:",noindex"`
	AbstractHTML	string `datastore:",noindex"`
	Author			*datastore.Key
	DateAdded		time.Time
	LastEdited		time.Time
	Tags			[]string
	CoverImage		string
}

type Post struct {
	Title				string
	JoinedTitle			string
	DateAdded			time.Time
	LastEdited			time.Time
	YearAdded			int
	MonthAdded			int
	DayAdded			int
	Author				*datastore.Key
	Tags				[]string
	AbstractMD			string `datastore:",noindex"`
	AbstractHTML		string `datastore:",noindex"`
	BodyMD				string `datastore:",noindex"`
	BodyHTML			string `datastore:",noindex"`
	Series				*datastore.Key
	PositionInSeries	int
	ReferencesMD		string `datastore:",noindex"`
	ReferencesHTML		string `datastore:",noindex"`
	CoverImage			string
}
