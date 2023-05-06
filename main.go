package main

import (
	"context"
	"errors"
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
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"

	"github.com/gin-gonic/gin"
)

type MusicPlot struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	IsPlaylist bool    `json:"is_playlist"`
}
type MusicPlots struct {
	Plots []MusicPlot `json:"plots"`
}

func normalizeAudioFeatures(features *spotify.AudioFeatures, dim int) []float64 {
	arr := make([]float64, dim)
	arr[0] = float64(features.Acousticness)
	arr[1] = float64(features.Danceability)
	arr[2] = float64(features.Energy)
	arr[3] = float64(features.Instrumentalness)
	// arr[4] = float64((features.Key + 1) / 12.0)
	arr[4] = float64(features.Liveness)
	arr[5] = float64((features.Loudness + 60) / 60.0)
	//arr[7] = float64(features.Mode)
	arr[6] = float64(features.Speechiness)
	// 正規化の方法を考える
	arr[7] = float64(features.Tempo / 200.0)
	//arr[10] = float64((features.TimeSignature - 3) / 5.0)
	arr[8] = float64(features.Valence)
	return arr
}

func ArtistTracksandFeatures(client *spotify.Client, ctx context.Context, artistID spotify.ID) (artistTracks []spotify.FullTrack, artistAudioFeatures []*spotify.AudioFeatures, err error) {
	artistTracks, err = client.GetArtistsTopTracks(ctx, spotify.ID(artistID), "JP")
	if err != nil {
		return
	}
	var artistTrackIDs []spotify.ID
	for _, track := range artistTracks {
		artistTrackIDs = append(artistTrackIDs, track.ID)
	}
	artistAudioFeatures, err = client.GetAudioFeatures(ctx, artistTrackIDs...)
	return
}
func PCA(client *spotify.Client, ctx context.Context) func(c *gin.Context) {
	dim := 9
	return func(c *gin.Context) {
		// 何も指定していない場合はアーティスト名がaから始まるものについて出力
		playlistID := c.Query("playlistID")
		artistID := c.Query("artistID")
		playlist, err := client.GetPlaylist(ctx, spotify.ID(playlistID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}
		var playlistIDs []spotify.ID
		if len(playlist.Tracks.Tracks) == 0 {
			c.JSON(http.StatusBadRequest, errors.New("プレイリスト内に楽曲がありません。"))
			return
		}
		for _, track := range playlist.Tracks.Tracks {
			playlistIDs = append(playlistIDs, track.Track.ID)
		}
		// アーティストのトップの楽曲と楽曲情報を取得

		artistTracks, artistAudioFeatures, err := ArtistTracksandFeatures(client, ctx, spotify.ID(artistID))
		if err != nil {
			c.JSON(err.(spotify.Error).Status, err)
			return
		}

		AudioFeatures, err := client.GetAudioFeatures(ctx, playlistIDs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// 楽曲情報を取得
		// データの正規化
		y := mat.NewDense(len(playlistIDs), dim, nil)
		for i, feature := range AudioFeatures {
			dense := mat.NewDense(dim, 1, normalizeAudioFeatures(feature, dim))
			y.SetRow(i, mat.Col(nil, 0, dense))
		}
		// pcaの実行
		plots, err := calcPCA(y, dim, artistAudioFeatures)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		// データの整形、レスポンス
		for i := 0; i < len(playlistIDs); i++ {
			plots[i].ID = string(playlistIDs[i])
			plots[i].Name = playlist.Tracks.Tracks[i].Track.Name
			plots[i].IsPlaylist = true
		}
		for i := len(playlistIDs); i < len(playlistIDs)+len(artistAudioFeatures); i++ {
			j := i - len(playlistIDs)
			plots[i].ID = string(artistTracks[j].ID)
			plots[i].Name = artistTracks[j].Name
			plots[i].IsPlaylist = false
		}
		c.JSON(http.StatusOK, MusicPlots{Plots: plots})
	}
}
func calcPCA(y *mat.Dense, dim int, artistAudioFeatures []*spotify.AudioFeatures) (plots []MusicPlot, err error) {
	var pc stat.PC
	ok := pc.PrincipalComponents(y, nil)
	if !ok {
		log.Fatal("PCA fails")
		err = errors.New("PCAの実行に失敗しました。")
		return
	}
	// 分析後の次元数
	afterDim := 2
	var proj mat.Dense
	var vec mat.Dense
	artistsNum := len(artistAudioFeatures)
	pc.VectorsTo(&vec)
	y_ := mat.NewDense(artistsNum, dim, nil)
	for i, feature := range artistAudioFeatures {
		dense := mat.NewDense(dim, 1, normalizeAudioFeatures(feature, dim))
		y_.SetRow(i, mat.Col(nil, 0, dense))
	}
	playlistsNum, _ := y.Dims()
	allFeatures := mat.NewDense(playlistsNum+artistsNum, dim, nil)
	allFeatures.Stack(y, y_)
	proj.Mul(allFeatures, vec.Slice(0, dim, 0, afterDim))
	plots = make([]MusicPlot, playlistsNum+artistsNum)
	for i := 0; i < len(plots); i++ {
		plots[i].X = proj.At(i, 0)
		plots[i].Y = proj.At(i, 1)
	}
	return

}

// TODO 検索の際にアーティスト以外にも対応
func search(client *spotify.Client, ctx context.Context) func(c *gin.Context) {

	return func(c *gin.Context) {
		// 何も指定していない場合はアーティスト名がaから始まるものについて出力
		searchQuery := c.Query("q")
		if searchQuery == "" {
			searchQuery = "a"
		}
		searchResult, err := client.Search(ctx, searchQuery, spotify.SearchTypeArtist)
		if err != nil {
			log.Fatal(err)
			c.JSON(err.(spotify.Error).Status, err.Error())
			return
		}
		c.JSON(http.StatusOK, searchResult)
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
func setCors() cors.Config {
	return cors.Config{
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
	r.Use(cors.New(setCors()))
	spotify := r.RouterGroup.Group("/v1")
	spotify.GET("/search", search(client, ctx))
	spotify.GET("/pca", PCA(client, ctx))
	r.Run(":" + os.Getenv("API_PORT"))
}
