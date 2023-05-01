package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/gin-gonic/gin"
)

// TODO 検索の際にアーティスト以外にも対応
func search(client *spotify.Client, ctx context.Context) func(c *gin.Context) {

	return func(c *gin.Context) {
		searchResult, err := client.Search(ctx, c.Query("q"), spotify.SearchTypeArtist)
		if err != nil {
			log.Fatal(err)
			c.JSON(err.(spotify.Error).Status, err.Error())
			return
		}
		c.JSON(http.StatusOK, searchResult)
	}
}

func playlists() func(c *gin.Context) {

	return func(c *gin.Context) {

	}
}

// clientを作成している部分の分離
func main() {
	r := gin.Default()

	ctx := context.Background()

	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	spotify := r.RouterGroup.Group("/v1")
	spotify.GET("/search", search(client, ctx))
	spotify.GET("/me/playlists", playlists())

	r.Run(":8000")
}
