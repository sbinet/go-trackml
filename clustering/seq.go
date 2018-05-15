// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clustering

import (
	"math"

	trackml "github.com/sbinet/go-trackml"
	"github.com/sbinet/go-trackml/hough"
	"gonum.org/v1/gonum/floats"
)

// spred predicts clusters hits.
type spred struct {
	nbinsR0Inv int
	nbinsGamma int
	nbinsTheta int
	minHits    int
}

// Predict clusters hits.
func (pred *spred) Predict(hits []trackml.Hit) ([]int, error) {
	h := hough.New(hits)

	theta := make([]float64, pred.nbinsTheta)
	floats.Span(theta, -math.Pi, +math.Pi)

	var tracks [][]int
	for _, v := range theta {
		tracks = h.Calc(tracks, v, pred.nbinsR0Inv, pred.nbinsGamma, pred.minHits)
	}

	trackID := 0
	labels := make([]int, len(hits))
	used := make(map[int]struct{}, len(hits))
	for _, hits := range tracks[:] {
		slice := make([]int, 0, len(hits))
		for _, hit := range hits {
			if _, dup := used[hit]; !dup {
				slice = append(slice, hit)
			}
		}
		if len(slice) >= pred.minHits {
			for _, v := range slice {
				labels[v] = trackID
				used[v] = struct{}{}
			}
			trackID++
		}
	}
	return labels, nil
}
