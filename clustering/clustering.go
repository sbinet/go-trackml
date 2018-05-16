// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clustering

import (
	trackml "github.com/sbinet/go-trackml"
)

// Classifier clusters hits.
type Classifier interface {
	Predict(hits []trackml.Hit) ([]int, error)
}

func New(nWorkers, nbinsR0Inv, nbinsGamma, nbinsTheta, minHits int) Classifier {
	switch {
	case nWorkers > 1:
		return &pcluster{
			nWorkers:   nWorkers,
			nbinsR0Inv: nbinsR0Inv,
			nbinsGamma: nbinsGamma,
			nbinsTheta: nbinsTheta,
			minHits:    minHits,
		}
	default:
		return &scluster{
			nbinsR0Inv: nbinsR0Inv,
			nbinsGamma: nbinsGamma,
			nbinsTheta: nbinsTheta,
			minHits:    minHits,
		}
	}
}
