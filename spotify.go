package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"proxy/pca"
	"time"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gonum.org/v1/gonum/mat"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2/clientcredentials"
)

func (sh *SpotifyHandler) updateClient() error {
	token, err := sh.config.Token(sh.ctx)
	if err != nil {
		return err
	}
	httpClient := spotifyauth.New().Client(sh.ctx, token)
	sh.client = spotify.New(httpClient)
	return nil
}
func NewSpotifyHandler(config *clientcredentials.Config, ctx context.Context) (error, SpotifyHandler) {
	sh := SpotifyHandler{
		config: config,
		ctx:    ctx,
	}
	// TODOここを分離
	go func() {
		t := time.NewTicker(55 * time.Minute) // 55分おきに通知
		for {
			select {
			case <-t.C:
				err := sh.updateClient()
				if err != nil {
					fmt.Printf("%s", err.Error())
				}
			}
		}
		t.Stop() // タイマを止める。
	}()
	err := sh.updateClient()
	return err, sh
}

type SpotifyHandler struct {
	client *spotify.Client
	config *clientcredentials.Config
	ctx    context.Context
}

// TODO 検索の際にアーティスト以外にも対応
// チェックする前にtokenが切れてないことを確認する。
func (sh SpotifyHandler) search() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 何も指定していない場合はアーティスト名がaから始まるものについて出力
		searchQuery := c.Query("q")
		if searchQuery == "" {
			searchQuery = "a"
		}
		searchResult, err := sh.client.Search(sh.ctx, searchQuery, spotify.SearchTypeArtist)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, searchResult)
	}
}

func TokenProxy() func(c *gin.Context) {
	return func(c *gin.Context) {
		remote, err := url.Parse("https://accounts.spotify.com/api/token")
		if err != nil {
			fmt.Println(err.Error())
			return
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

// TODO 最終的にはデータの取得のみを行うように
func (sh SpotifyHandler) PCA() func(c *gin.Context) {
	dim := 9
	return func(c *gin.Context) {
		// 何も指定していない場合はアーティスト名がaから始まるものについて出力
		playlistID := c.Query("playlistID")
		artistID := c.Query("artistID")
		playlist, err := sh.client.GetPlaylist(sh.ctx, spotify.ID(playlistID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
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

		artistTracks, artistAudioFeatures, err := ArtistTracksandFeatures(sh.client, sh.ctx, spotify.ID(artistID))
		if err != nil {
			c.JSON(err.(spotify.Error).Status, err)
			return
		}

		AudioFeatures, err := sh.client.GetAudioFeatures(sh.ctx, playlistIDs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// 楽曲情報を取得
		// データの正規化
		y := mat.NewDense(len(playlistIDs), dim, nil)
		for i, feature := range AudioFeatures {
			dense := mat.NewDense(dim, 1, pca.NormalizeAudioFeatures(feature, dim))
			y.SetRow(i, mat.Col(nil, 0, dense))
		}
		// pcaの実行
		plots, err := pca.CalcPCA(y, dim, artistAudioFeatures)
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
		c.JSON(http.StatusOK, pca.MusicPlots{Plots: plots})
	}
}
