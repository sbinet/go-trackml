// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clustering

import (
	trackml "github.com/sbinet/go-trackml"
)

// Predictor predicts clusters of hits.
type Predictor interface {
	Predict(hits []trackml.Hit) ([]int, error)
}

func New(nWorkers, nbinsR0Inv, nbinsGamma, nbinsTheta, minHits int) Predictor {
	switch {
	case nWorkers > 1:
		return &ppred{
			nWorkers:   nWorkers,
			nbinsR0Inv: nbinsR0Inv,
			nbinsGamma: nbinsGamma,
			nbinsTheta: nbinsTheta,
			minHits:    minHits,
		}
	default:
		return &spred{
			nbinsR0Inv: nbinsR0Inv,
			nbinsGamma: nbinsGamma,
			nbinsTheta: nbinsTheta,
			minHits:    minHits,
		}
	}
}
