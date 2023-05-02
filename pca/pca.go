package pca

import (
	"gonum.org/v1/gonum/mat"
)

func PCA(parameters []float64) [][]float64 {

	data := mat.NewDense(len(parameters), len(parameters[0]), parameters)
	pca := mat.PCABasis(nil, data, nil)
	transformedData := mat.DenseCopyOf(data)
	transformedData.Product(data, pca.VectorsTo(nil))

	// 圧縮されたデータを取得
	compressedData := transformedData.Slice(0, len(parameters), 0, 2)
	return mat.Formatted(compressedData)
}
