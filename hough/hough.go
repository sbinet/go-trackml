// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hough

import (
	"encoding/json"
	"log"
	"math"
	"sort"

	"github.com/pkg/errors"
	"github.com/sbinet/go-trackml"
	"gonum.org/v1/gonum/floats"
)

var (
	errLenMismatch = errors.Errorf("hough: length mismatch")
)

type Hough struct {
	HitIDs []int `json:"HitID"`
	X      []float64
	Y      []float64
	Z      []float64
	R      []float64
	Phi    []float64
	R0Inv  []float64
	Gamma  []float64

	R0InvDigi  []int
	GammaDigi  []int
	ComboDigi  []int
	ComboDigiN []int `json:"ComboDigiCounts"`
}

func New(hits []trackml.Hit) *Hough {
	var (
		ids = make([]int, len(hits))
		xs  = make([]float64, len(hits))
		ys  = make([]float64, len(hits))
		zs  = make([]float64, len(hits))
	)
	for i, hit := range hits {
		ids[i] = hit.HitID
		xs[i] = hit.X
		ys[i] = hit.Y
		zs[i] = hit.Z
	}
	rs, phis := cart2Cyl(xs, ys)

	N := len(hits)
	hough := Hough{
		HitIDs:     ids,
		X:          xs,
		Y:          ys,
		Z:          zs,
		R:          rs,
		Phi:        phis,
		R0Inv:      make([]float64, N),
		Gamma:      make([]float64, N),
		R0InvDigi:  make([]int, N),
		GammaDigi:  make([]int, N),
		ComboDigi:  make([]int, N),
		ComboDigiN: make([]int, N),
	}

	//	f, err := os.Create("list.go.txt")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	hough.FF = f
	return &hough
}

func (hough *Hough) Calc(tracks [][]int, theta float64, nbinsR0Inv, nbinsGamma, minHits int) [][]int {
	if nbinsR0Inv <= 0 {
		nbinsR0Inv = 200
	}
	if nbinsGamma <= 0 {
		nbinsGamma = 100
	}

	for i := range hough.ComboDigi {
		hough.ComboDigi[i] = 0
		hough.ComboDigiN[i] = 0
	}

	for i := range hough.HitIDs {
		rinv := 1 / hough.R[i]
		r0inv := 2 * math.Cos(hough.Phi[i]-theta) * rinv
		gamma := hough.Z[i] * rinv
		hough.R0Inv[i] = r0inv
		hough.Gamma[i] = gamma
	}

	digitizeCol(hough.R0InvDigi, hough.R0Inv, nbinsR0Inv, -0.02, +0.02) // Tune it
	digitizeCol(hough.GammaDigi, hough.Gamma, nbinsGamma, -50.0, +50.0) // Tune it
	hough.combineDigi(hough.ComboDigi, [][]int{hough.R0InvDigi, hough.GammaDigi})

	hough.countDigi(hough.ComboDigiN)
	hough.fiducialCut(hough.ComboDigiN, hough.R0InvDigi, nbinsR0Inv)
	hough.fiducialCut(hough.ComboDigiN, hough.GammaDigi, nbinsGamma)

	set := make(map[int][]int)
	for i, digi := range hough.ComboDigi {
		if hough.ComboDigiN[i] < minHits {
			continue
		}
		// id := hough.HitIDs[i]
		//		log.Printf("hit[%d]=%v digi-n=%v digi=%v", i, id, hough.ComboDigiN[i], digi)
		set[digi] = append(set[digi], i)
	}
	if len(set) > 0 {
		trks := make([][]int, 0, len(set))
		for _, trk := range set {
			sort.Ints(trk)
			trks = append(trks, trk)
			//	if trk[0] == 36461 {
			//		log.Printf("digi=%v trk=%v", digi, trk)
			//		log.Printf("digi=%v id=%v %v", digi, hough.ComboDigi[trk[0]], hough.ComboDigiN[trk[0]])
			//	}
		}
		//fmt.Fprintf(hough.FF, "theta=%v tracks=%v\n", theta, len(trks))
		//		log.Printf(">>> theta=%v trks=%v", theta, trks)
		tracks = append(tracks, trks...)
	}
	return tracks
}

func (h *Hough) Dump(fname string) {
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	err = enc.Encode(h)
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}

func cart2Cyl(xs, ys []float64) (rs, phis []float64) {
	rs = make([]float64, len(xs))
	phis = make([]float64, len(xs))

	for i := range xs {
		x := xs[i]
		y := ys[i]
		rs[i] = math.Sqrt(x*x + y*y)
		phis[i] = math.Atan2(y, x)
	}
	return rs, phis
}

func digitizeCol(dst []int, vs []float64, nbins int, min, max float64) {
	if len(dst) != len(vs) {
		panic(errLenMismatch)
	}
	if math.IsInf(min, -1) {
		min = floats.Min(vs)
	}
	if math.IsInf(max, +1) {
		max = floats.Max(vs)
	}
	bins := floats.Span(make([]float64, nbins), min, max)
	for i, v := range vs {
		idx := sort.SearchFloat64s(bins, v)
		dst[i] = idx
	}
}

func (h *Hough) combineDigi(dst []int, cols [][]int) {
	if dst == nil {
		dst = make([]int, len(h.HitIDs))
	}
	for i, col := range cols {
		for j, digi := range col {
			dst[j] += digi * ipow(10, i*5)
		}
	}
}

func (h *Hough) countDigi(dst []int) {
	if dst == nil {
		dst = make([]int, len(h.HitIDs))
	}
	if len(dst) != len(h.HitIDs) {
		panic(errLenMismatch)
	}
	set := make(map[int]int, len(h.HitIDs)/2)
	for _, v := range h.ComboDigi {
		set[v]++
	}
	for i, v := range h.ComboDigi {
		dst[i] = set[v]
	}
}

func (h *Hough) fiducialCut(dst, col []int, n int) {
	if len(dst) != len(col) {
		panic(errLenMismatch)
	}
	for i, digi := range col {
		o := digi != 0 && digi != n
		var v int
		switch o {
		case true:
			v = 1
		case false:
			v = 0
		}
		dst[i] *= v
	}
}

func ipow(v, n int) int {
	switch n {
	case 0:
		return 1
	case 1:
		return v
	case 2:
		return v * v
	case 3:
		return v * v * v
	case 4:
		return v * v * v * v
	case 5:
		return v * v * v * v * v
	}
	o := 1
	for i := 0; i < n; i++ {
		o *= v
	}
	return o
}
