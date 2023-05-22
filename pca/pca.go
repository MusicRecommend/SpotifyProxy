package pca

import (
	"errors"
	"log"

	"github.com/zmb3/spotify/v2"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
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

func NormalizeAudioFeatures(features *spotify.AudioFeatures, dim int) []float64 {
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
func CalcPCA(y *mat.Dense, dim int, artistAudioFeatures []*spotify.AudioFeatures) (plots []MusicPlot, err error) {
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
		dense := mat.NewDense(dim, 1, NormalizeAudioFeatures(feature, dim))
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
