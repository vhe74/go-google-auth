package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type AuthService struct {
	domain   string
	sessions sync.Map
}

func (as *AuthService) initialize() {
	log.Println("Initializing  Auth")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	clientCallbackURL := os.Getenv("CLIENT_CALLBACK_URL")
	sessionSecret := os.Getenv("SESSION_SECRET")
	domain := os.Getenv("DOMAIN")
	if clientID == "" || clientSecret == "" || clientCallbackURL == "" || sessionSecret == "" || domain == "" {
		log.Fatal("Environment variables (CLIENT_ID, CLIENT_SECRET, CLIENT_CALLBACK_URL, SESSION_KEY, SESSION_SECRET, DOMAIN) are required")
	}

	goth.UseProviders(
		google.New(clientID, clientSecret, clientCallbackURL),
	)

	as.domain = domain
	log.Printf("%+v", as)
}

func (as *AuthService) signInWithProviderHandler(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (as *AuthService) callbackHandler(c *gin.Context) {

	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	data, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("User Logged in  : %+s\n", data.Email)

	sessId := uuid.New().String()

	as.sessions.Store(sessId, data.Email)
	//log.Printf("%+v", as)

	c.SetCookie("SESSIONID", sessId, 3600, "/", as.domain, false, true)
	c.SetCookie("USERID", data.Email, 3600, "/", as.domain, false, true)

	c.Redirect(http.StatusTemporaryRedirect, "/success")
}

func (as *AuthService) successHandler(c *gin.Context) {

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
      <div style="
          background-color: #fff;
          padding: 40px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
          text-align: center;
      ">
          <h1 style="
              color: #333;
              margin-bottom: 20px;
          ">You have Successfull signed in!</h1>
          
          </div>
      </div>
  `)))
}

func (as *AuthService) logoutHandler(c *gin.Context) {
	cookieUserID, err := c.Cookie("USERID")
	if err != nil {
		log.Println("No cookie found")
	} else {
		log.Printf("User Logged out  : %+s\n", cookieUserID)
	}

	cookieSessionId, err := c.Cookie("SESSIONID")
	if err != nil {
		log.Println("No cookie found")
	}

	as.sessions.Delete(cookieSessionId)

	//log.Printf("%+v", as)

	c.SetCookie("SESSIONID", "", 5, "/", as.domain, false, true)
	c.SetCookie("USERID", "", 5, "/", as.domain, false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (as *AuthService) checkAuth(c *gin.Context) bool {
	cookie, err := c.Cookie("SESSIONID")
	if err != nil {
		log.Println("No cookie found")
		return false
	}

	cookieUser, err := c.Cookie("USERID")
	if err != nil {
		log.Println("No cookie (user) found")
		return false
	}

	//check cookie content with session database

	value, ok := as.sessions.Load(cookie)
	if ok != true {
		log.Println("No session found for cookie")
		return false
	}

	if value != cookieUser {
		log.Printf("User (%s) and Session (%s) in cookies don't match\n", cookieUser, cookie)
		return false
	}

	log.Printf("Session found for %s -> %s", cookie, value)
	return true
}
