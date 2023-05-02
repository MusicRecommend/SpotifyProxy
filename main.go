package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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
func TokenProxy() func(c *gin.Context) {
	return func(c *gin.Context) {
		remote, err := url.Parse("https://accounts.spotify.com/api/token")
		if err != nil {
			panic(err)
		}
		token := os.Getenv("SPOTIFY_BASIC_KEY")
		proxy := httputil.NewSingleHostReverseProxy(remote)
		c.Request.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = remote.Path
		}
		proxy.ServeHTTP(c.Writer, c.Request)
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
	apiToken := r.RouterGroup.Group("/api")
	apiToken.POST("/token", TokenProxy())
	r.Use(cors.New(cors.Config{
		// アクセスを許可したいアクセス元
		AllowOrigins: []string{
			os.Getenv("ALLOW_ORIGIN"),
		},
		// アクセスを許可したいHTTPメソッド(以下の例だとPUTやDELETEはアクセスできません)
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
		},
		// 許可したいHTTPリクエストヘッダ
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		// cookieなどの情報を必要とするかどうか
		AllowCredentials: true,
		// preflightリクエストの結果をキャッシュする時間
		MaxAge: 24 * time.Hour,
	}))
	spotify := r.RouterGroup.Group("/v1")
	spotify.GET("/search", search(client, ctx))
	spotify.GET("/me/playlists", playlists())

	r.Run(":" + os.Getenv("API_PORT"))
}
