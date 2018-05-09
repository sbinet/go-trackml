// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/sbinet/go-trackml"
	"github.com/sbinet/go-trackml/hough"
	"gonum.org/v1/gonum/floats"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("trkml: ")

	flag.Parse()

	fname := flag.Arg(0)
	if fname == "" {
		fname = "./example_standard/dataset/event000000200"
	}

	log.Printf("processing [%s]...", fname)
	hits, cells, parts, mcs, err := trackml.ReadEvent(fname)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("processing [%s]... [done]", fname)

	//hits = hits[5000:10000]
	//hits = hits[:10000]
	//	hits = hits[:5000]
	log.Printf("hits:  %d", len(hits))
	log.Printf("cells: %d", len(cells))
	log.Printf("parts: %d", len(parts))
	log.Printf("truth: %d", len(mcs))

	const (
		nbinsR0Inv = 200
		nbinsGamma = 500
		nbinsTheta = 500
	)

	if true && false {
		theta := -2.3105501079508097
		minHits := 9
		hh := hough.New(hits)
		tracks := hh.Calc(nil, theta, nbinsR0Inv, nbinsGamma, minHits)
		//log.Fatalf("tracks: %v\n\nlen=%d", tracks, len(tracks))
		ff, err := os.Create("list.txt")
		if err != nil {
			log.Fatal(err)
		}
		for _, tt := range tracks {
			fmt.Fprintf(ff, "%v\n", tt)
		}
		ff.Close()
		{
			ff, err := os.Create("go.json")
			if err != nil {
				log.Fatal(err)
			}
			defer ff.Close()
			err = json.NewEncoder(ff).Encode(hh)
			if err != nil {
				log.Fatal(err)
			}
			ff.Close()
		}
		log.Fatalf("boo")
	}

	model := Clusterer{
		nWorkers:   runtime.NumCPU() + 1,
		nbinsR0Inv: nbinsR0Inv,
		nbinsGamma: nbinsGamma,
		nbinsTheta: nbinsTheta,
		minHits:    9,
	}
	//model.nWorkers = 1

	labels, err := model.Predict(hits)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("labels: %d", len(labels))
	log.Printf("labels: %v", labels[:3])
}

type Clusterer struct {
	nWorkers   int
	nbinsR0Inv int
	nbinsGamma int
	nbinsTheta int
	minHits    int
}

func (cs *Clusterer) Predict(hits []trackml.Hit) ([]int, error) {
	workers := make([]worker, cs.nWorkers)
	ch := make(chan float64, len(workers))
	var grp sync.WaitGroup
	grp.Add(len(workers))

	for i := range workers {
		wrk := &workers[i]
		wrk.cs = cs
		wrk.h = hough.New(hits)
		go func(wrk *worker) {
			defer grp.Done()
			for v := range ch {
				wrk.run(v)
			}
		}(wrk)
	}

	theta := make([]float64, cs.nbinsTheta)
	floats.Span(theta, -math.Pi, +math.Pi)
	for _, v := range theta {
		//		log.Printf("theta: %v", v)
		ch <- v
	}
	close(ch)
	grp.Wait()

	ntracks := 0
	for _, wrk := range workers {
		ntracks += len(wrk.tracks)
	}
	tracks := make([][]int, 0, ntracks)
	for _, wrk := range workers {
		tracks = append(tracks, wrk.tracks...)
	}

	log.Printf("ntracks: %d", len(tracks))
	//	workers[0].h.FF.Close()

	/*
		ff, err := os.Create("list.go.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer ff.Close()
		for i, trks := range tracks {
			log.Printf("trk[%d] = %v", i, trks)
			fmt.Fprintf(ff, "trk[%d] = %v\n", i, trks)
		}
		ff.Close()
	*/

	trackID := 0
	labels := make([]int, len(hits))
	used := make(map[int]struct{}, len(hits))
	for _, hits := range tracks[:] {
		//		log.Printf("%03d track=%v", ii, hits)
		slice := make([]int, 0, len(hits))
		for _, hit := range hits {
			if _, dup := used[hit]; !dup {
				slice = append(slice, hit)
			}
		}
		//		log.Printf(">>> slice=%v => %v", slice, len(slice) >= cs.minHits)
		if len(slice) >= cs.minHits {
			for _, v := range slice {
				labels[v] = trackID
				used[v] = struct{}{}
			}
			trackID++
		}
	}

	//	log.Printf("labels=%v", labels[:20])

	//	{
	//		f, err := os.Create("labels.txt")
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		defer f.Close()
	//		json.NewEncoder(f).Encode(labels)
	//		//		fmt.Fprintf(f, "%v\n", labels)
	//		f.Close()
	//	}
	return labels, nil
}

type worker struct {
	cs     *Clusterer
	h      *hough.Hough
	tracks [][]int
}

func (wrk *worker) run(theta float64) {
	wrk.tracks = wrk.h.Calc(wrk.tracks, theta, wrk.cs.nbinsR0Inv, wrk.cs.nbinsGamma, wrk.cs.minHits)
}
