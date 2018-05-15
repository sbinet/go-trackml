// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
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
	log.SetPrefix("trkml: ")

	ncpus := flag.Int("ncpus", 1, "number of goroutines to use for the prediction")
	profCPU := flag.Bool("prof-cpu", false, "enable CPU profiling")
	profMEM := flag.Bool("prof-mem", false, "enable MEM profiling")

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
		fname = "./example_standard/dataset/event000000200"
	}

	log.Printf("processing [%s]...", fname)
	evt, err := trackml.ReadMcEvent(fname)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("processing [%s]... [done]", fname)

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

	runFromJSON := false
	var labels []int
	if !runFromJSON {
		labels, err = model.Predict(evt.Hits)
		if err != nil {
			log.Fatal(err)
		}
		{
			out, err := os.Create("go.labels.json")
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()
			err = json.NewEncoder(out).Encode(labels)
			if err != nil {
				log.Fatal(err)
			}
			err = out.Close()
			if err != nil {
				log.Fatal(err)
			}
		}

	} else {
		out, err := os.Open("labels.json")
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		err = json.NewDecoder(out).Decode(&labels)
		if err != nil {
			log.Fatal(err)
		}
	}

	score := trackml.Score(evt, labels)
	log.Printf("score for event %v: %v", evt.ID, score)

	var scores []float64
	ds, err := trackml.NewDataset("./example_standard/dataset", 0, 5, nil)
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

	log.Printf("mean score: %v", stat.Mean(scores, nil))
}
