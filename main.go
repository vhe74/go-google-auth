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
	serveTemplate(c, "templates/index.html", gin.H{})
}

func protectedHandler(c *gin.Context) {

	if authSce.checkAuth(c) {
		serveTemplate(c, "templates/protected.html", gin.H{})
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

}

func serveTemplate(c *gin.Context, templateFile string, data map[string]any) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
