// Package controller contains all the controllers
// of the application
package controller

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/prakashjegan/golangexercise/database/model"
	"github.com/prakashjegan/golangexercise/handler"
	"github.com/prakashjegan/golangexercise/lib/renderer"

	"gopkg.in/danilopolani/gocialite.v1"
)

var Gocial = gocialite.NewDispatcher()

// CreateUserAuth - POST /register
func CreateUserAuth(c *gin.Context) {
	auth := model.Auth{}

	// bind JSON
	if err := c.ShouldBindJSON(&auth); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUserAuth(auth)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Show homepage with login URL
func IndexHandler(c *gin.Context) {
	// c.Writer.Write([]byte("<html><head><title>Gocialite example</title></head><body>" +
	// 	"<a href='/auth/github'><button>Login with GitHub</button></a><br>" +
	// 	"<a href='/auth/linkedin'><button>Login with LinkedIn</button></a><br>" +
	// 	"<a href='/auth/facebook'><button>Login with Facebook</button></a><br>" +
	// 	"<a href='/auth/google'><button>Login with Google</button></a><br>" +
	// 	"<a href='/auth/bitbucket'><button>Login with Bitbucket</button></a><br>" +
	// 	"<a href='/auth/amazon'><button>Login with Amazon</button></a><br>" +
	// 	"<a href='/auth/amazon'><button>Login with Slack</button></a><br>" +
	// 	"</body></html>"))

	c.Writer.Write([]byte("<html><body>" +
		//"<a href='/auth/linkedin'><button>Login with LinkedIn</button></a><br>" +
		"<a href='/api/v1/loginAuth/auth/facebook'><button>Login with Facebook</button></a><br>" +
		"<a href='/api/v1/loginAuth/auth/google'><button>Login with Google</button></a><br>" +
		"</body></html>"))

}

// Redirect to correct oAuth URL
func RedirectHandler(c *gin.Context) {
	// Retrieve provider from route
	provider := c.Param("provider")

	// In this case we use a map to store our secrets, but you can use dotenv or your framework configuration
	// for example, in revel you could use revel.Config.StringDefault(provider + "_clientID", "") etc.
	providerSecrets := map[string]map[string]string{
		// "github": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/github/callback",
		// },
		// "linkedin": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/linkedin/callback",
		// },
		"facebook": {
			"clientID":     "",
			"clientSecret": "",
			"redirectURL":  "https://api.XXXXXX.XXX/api/v1/loginAuth/auth/facebook/callback",
		},
		"google": {
			"clientID":     "",
			"clientSecret": "",
			"redirectURL":  "https://api.XXXXXX.XXX/api/v1/loginAuth/auth/google/callback",
			//"redirectURL": "http://localhost:8999/api/v1/loginAuth/auth/google/callback",
		},
		// "bitbucket": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/bitbucket/callback",
		// },
		// "amazon": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/amazon/callback",
		// },
		// "slack": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/slack/callback",
		// },
		// "asana": {
		// 	"clientID":     "xxxxxxxxxxxxxx",
		// 	"clientSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		// 	"redirectURL":  "https://api.XXXXXX.XXX/auth/asana/callback",
		// },
	}

	providerScopes := map[string][]string{
		// "github":    []string{"public_repo"},
		// "linkedin":  []string{},
		"facebook": []string{},
		"google":   []string{},
		// "bitbucket": []string{},
		// "amazon":    []string{},
		// "slack":     []string{},
		// "asana":     []string{},
	}

	providerData := providerSecrets[provider]
	actualScopes := providerScopes[provider]
	authURL, err := Gocial.New().
		Driver(provider).
		Scopes(actualScopes).
		Redirect(
			providerData["clientID"],
			providerData["clientSecret"],
			providerData["redirectURL"],
		)

	// Check for errors (usually driver not valid)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	// Redirect with authURL
	c.Redirect(http.StatusFound, authURL)
}

// Handle callback of provider
func CallbackHandler(c *gin.Context) {
	// Retrieve query params for state and code
	state := c.Query("state")
	code := c.Query("code")
	provider := c.Param("provider")

	// Handle callback and check for errors
	user, token, err := Gocial.Handle(state, code)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	// Print in terminal user information
	fmt.Printf("%#v", token)
	fmt.Printf("%#v", user)
	fmt.Printf("%#v", provider)
	// If no errors, show provider name
	c.Writer.Write([]byte("Hi, " + user.FullName))
}
