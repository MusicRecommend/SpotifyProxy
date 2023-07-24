package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// clientを作成している部分の分離
func main() {
	r := gin.Default()
	var ctx context.Context
	var config *clientcredentials.Config
	ctx = context.Background()

	config = &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	err, spotifyHandler := NewSpotifyHandler(config, ctx)
	if err != nil {
		fmt.Printf("spotify Clientの取得に失敗:%v", err)
		os.Exit(1)
	} else {
		fmt.Println("start application")
	}
	apiToken := r.RouterGroup.Group("/api")
	apiToken.POST("/token", TokenProxy())

	r.Use(cors.New(setCors()))
	spotify := r.RouterGroup.Group("/v1")
	spotify.GET("/search", spotifyHandler.search())
	spotify.GET("/track", spotifyHandler.getTrack())
	spotify.GET("/pca", spotifyHandler.PCA())
	r.Run(":" + os.Getenv("API_PORT"))
}
