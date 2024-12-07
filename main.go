package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

/*
Reference = https://permify.co/post/implement-oauth-2-golang-app/
*/

var authSce = AuthService{}

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file failed to load!")
	}

	authSce.initialize()

	r.GET("/", homeHandler)
	r.GET("/protected", protectedHandler)

	r.GET("/auth/:provider", authSce.signInWithProviderHandler)
	r.GET("/auth/:provider/callback", authSce.callbackHandler)
	r.GET("/success", authSce.successHandler)
	r.GET("/logout", authSce.logoutHandler)

	r.Run(":8080")
}

func homeHandler(c *gin.Context) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(c.Writer, gin.H{})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func protectedHandler(c *gin.Context) {

	if authSce.checkAuth(c) {
		tmpl, err := template.ParseFiles("templates/protected.html")
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(c.Writer, gin.H{})
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

}
