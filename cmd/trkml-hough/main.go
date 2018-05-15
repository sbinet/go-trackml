// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// trkml-hough is a simple TrackML example using a Hough transform to make predictions,
// similar to the Jupyter notebook from https://github.com/LAL/trackml-library.
//
//
// Usage:
//
//   $> trkml-hough [OPTIONS] <path-to-dataset>
//
// Examples:
//
//   $> trkml-hough ./example_standard/dataset/event000000200
//   $> trkml-hough -npcus=+1 ./example_standard/dataset/event000000200
//   $> trkml-hough -npcus=-1 ./example_standard/dataset/event000000200
//
// Options:
//
//   -ncpus int
//     	number of goroutines to use for the prediction (default 1)
//   -prof-cpu
//     	enable CPU profiling
//   -prof-mem
//     	enable MEM profiling
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/pkg/profile"
	"github.com/sbinet/go-trackml"
	"github.com/sbinet/go-trackml/clustering"
	"gonum.org/v1/gonum/stat"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("trkml-hough: ")

	ncpus := flag.Int("ncpus", 1, "number of goroutines to use for the prediction")
	profCPU := flag.Bool("prof-cpu", false, "enable CPU profiling")
	profMEM := flag.Bool("prof-mem", false, "enable MEM profiling")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `trkml-hough uses a Hough transform to make predictions.

Usage:

  $> trkml-hough [OPTIONS] <path-to-dataset>

Examples:

  $> trkml-hough ./example_standard/dataset/event000000200
  $> trkml-hough -npcus=+1 ./example_standard/dataset/event000000200
  $> trkml-hough -npcus=-1 ./example_standard/dataset/event000000200

Options:

`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *ncpus <= 0 {
		*ncpus = runtime.NumCPU() + 1
	}
	switch {
	case *profCPU:
		defer profile.Start(profile.CPUProfile).Stop()
	case *profMEM:
		defer profile.Start(profile.MemProfile).Stop()
	}

	fname := flag.Arg(0)
	if fname == "" {
		flag.Usage()
		log.Fatalf("missing path to event dataset")
	}

	log.Printf("loading [%s]...", fname)
	evt, err := trackml.ReadMcEvent(fname)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loading [%s]... [done]", fname)

	//	log.Printf("hits:  %d", len(evt.Hits))
	//	log.Printf("cells: %d", len(evt.Cells))
	//	log.Printf("parts: %d", len(evt.Ps))
	//	log.Printf("truth: %d", len(evt.Mcs))

	const (
		nbinsR0Inv = 200
		nbinsGamma = 500
		nbinsTheta = 500
	)

	model := clustering.New(*ncpus, nbinsR0Inv, nbinsGamma, nbinsTheta, 9)

	var labels []int
	labels, err = model.Predict(evt.Hits)
	if err != nil {
		log.Fatal(err)
	}

	score := trackml.Score(evt, labels)
	log.Printf("score for event %v: %v", evt.ID, score)

	dsname := "./example_standard/dataset"
	log.Printf("loading the whole dataset %q...", dsname)
	var scores []float64
	ds, err := trackml.NewDataset(dsname, 0, 5, nil)
	if err != nil {
		log.Fatal(err)
	}
	for ds.Next() {
		evt = ds.Event()

		var labels []int
		labels, err = model.Predict(evt.Hits)
		if err != nil {
			log.Fatal(err)
		}

		score := trackml.Score(evt, labels)
		log.Printf("score for event %v: %v", evt.ID, score)
		scores = append(scores, score)
		evt.Delete()
	}
	if err := ds.Err(); err != nil {
		log.Fatal(err)
	}
	log.Printf("loading the whole dataset %q... [done]", dsname)

	log.Printf("mean score: %v", stat.Mean(scores, nil))
}
