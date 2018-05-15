// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clustering

import (
	"math"
	"sync"

	trackml "github.com/sbinet/go-trackml"
	"github.com/sbinet/go-trackml/hough"
	"gonum.org/v1/gonum/floats"
)

// ppred clusters hits in parallel via its Predict method.
type ppred struct {
	nWorkers   int
	nbinsR0Inv int
	nbinsGamma int
	nbinsTheta int
	minHits    int
}

// Predict clusters hits.
func (pred *ppred) Predict(hits []trackml.Hit) ([]int, error) {
	workers := make([]worker, pred.nWorkers)
	for i := range workers {
		workers[i].tracks = make(map[int][][]int)
	}
	type index struct {
		i     int
		theta float64
	}
	ch := make(chan index, len(workers))
	var grp sync.WaitGroup
	grp.Add(len(workers))

	for i := range workers {
		wrk := &workers[i]
		wrk.pred = pred
		wrk.h = hough.New(hits)
		go func(wrk *worker) {
			defer grp.Done()
			for v := range ch {
				wrk.run(v.i, v.theta)
			}
		}(wrk)
	}

	theta := make([]float64, pred.nbinsTheta)
	floats.Span(theta, -math.Pi, +math.Pi)
	for i, v := range theta {
		ch <- index{i, v}
	}
	close(ch)
	grp.Wait()

	ntracks := 0
	for _, wrk := range workers {
		for _, trks := range wrk.tracks {
			ntracks += len(trks)
		}
	}
	tracks := make([][]int, 0, ntracks)
	for i := range theta {
		for _, wrk := range workers {
			if trks, ok := wrk.tracks[i]; ok {
				tracks = append(tracks, trks...)
			}
		}
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

type worker struct {
	pred   *ppred
	h      *hough.Hough
	tracks map[int][][]int
}

func (wrk *worker) run(i int, theta float64) {
	wrk.tracks[i] = wrk.h.Calc(wrk.tracks[i], theta, wrk.pred.nbinsR0Inv, wrk.pred.nbinsGamma, wrk.pred.minHits)
}
