package x

import (
	"errors"
	"crypto/rand"
	"encoding/base64"
	"html/template"
)

import (
	"models"
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"net/http"
	"github.com/gorilla/sessions"
)

type PageData struct {
	Author     *models.Author
	Misc       map[string]interface{}
	PostSeries map[string]template.HTML
	Post       map[string]template.HTML
	LogoutURL  string
}

type Configuration struct {
    ClientID 		string `yaml:"client_id"`
    ClientSecret 	string `yaml:"client_secret"`
	AuthURI			string `yaml:"auth_uri"`
	TokenURI		string `yaml:"token_uri"`
}

type FlashMessage struct {
	Type string
	Text string
}

var SessionStore = sessions.NewCookieStore([]byte("27116426b315ea1719a488ebc48fa00a7eb334abaa7712b0a6378a9dad62f9b5"))


func (c *Configuration) GetConfiguration() *Configuration {
    yamlFile, err := ioutil.ReadFile("config.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }
    return c
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err == nil {
		return b, nil
	} else {
		return nil, err
	}
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}


//
func CreateOauthRandomString(w http.ResponseWriter, r *http.Request) (string, error) {
	session, _ := SessionStore.Get(r, "abenga.com")
	oauth_string, err := GenerateRandomString(32);
	if err != nil {
		return "", err
	}
	session.Values["oauth_string"] = oauth_string
	if err := session.Save(r, w); err == nil {
		return oauth_string, nil
	} else {
		return "", err
	}
}

//
func GetSavedOauthRandomString(w http.ResponseWriter, r *http.Request) (string, error) {
	session, _ := SessionStore.Get(r, "abenga.com")
	if len(session.Values) > 0 {
		return session.Values["oauth_string"].(string), nil
	} else {
		return "", errors.New("Could not retrieve session id")
	}
}

//
func GetSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	session, _ := SessionStore.Get(r, "abenga.com")
	sessionID, err := GenerateRandomString(32);
	if err != nil {
		return "", err
	}
	session.Values["SessionID"] = sessionID
	if err := session.Save(r, w); err == nil {
		return sessionID, nil
	} else {
		return "", err
	}
}

//
func GetActiveSession(w http.ResponseWriter, r *http.Request) (string, error) {
	session, _ := SessionStore.Get(r, "abenga.com")
	if len(session.Values) > 0 {
		return string(session.Values["SessionID"].(string)), nil
	} else {
		return "", errors.New("Could not retrieve session ID!")
	}
}