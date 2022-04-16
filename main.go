package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func hello(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "hello"})
}

type Firebase struct {
	Auth *auth.Client
}

func NewFirebase() (*Firebase, error) {
	firebaseInst := new(Firebase)
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	authInst, err := app.Auth(context.Background())
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	firebaseInst.Auth = authInst
	
	return firebaseInst, nil
}

type Authentication struct {
	firebase Firebase
}

func NewAuthentication(firebase *Firebase) *Authentication {
	authInst := new(Authentication)
	authInst.firebase = *firebase
	return authInst
}

func authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("#### authenticate ####")
		app, err := NewFirebase()
		if err != nil {
			log.Println(err.Error())
			c.Abort()
			return
		}

		auth := NewAuthentication(app)

		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Authorization is empty")
			c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := auth.firebase.Auth.VerifyIDToken(context.Background(), tokenStr)
		if err != nil {
			log.Printf("authentication failed: %s", err.Error())
			c.Abort()
			return
		}

		res := token.Claims["user_id"].(string)
		c.JSON(http.StatusOK, gin.H{"token": res})
	}
}

func main() {
	engine := gin.Default()
	engine.Use(authenticate())
	engine.GET("/hello", hello)

	if err := engine.Run(":8080"); err != nil {
		log.Println(err.Error())
		return
	}
}