// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hough

import (
	"reflect"
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestDigitize(t *testing.T) {
	xs := []float64{0.2, 6.4, 3.0, 1.6}
	inds := make([]int, len(xs))
	digitizeCol(inds, xs, 5, 0, 10)
	want := []int{1, 3, 2, 1}
	if !reflect.DeepEqual(inds, want) {
		t.Fatalf("digitize error\ngot = %v\nwant= %v", inds, want)
	}
}
func TestDigitize2(t *testing.T) {
	xs := []float64{0.2, 6.4, 3.0, 1.6}
	bins := floats.Span(make([]float64, 5), 0, 10) // -> [0 2.5 5 7.5 10]
	inds := make([]int, 0, len(xs))
	for _, v := range xs {
		i := floats.NearestIdx(bins, v)
		if i > 0 && v <= bins[i] {
			i--
		}
		inds = append(inds, i)
	}
	want := []int{0, 2, 1, 0}
	if !reflect.DeepEqual(inds, want) {
		t.Fatalf("digitize error\ngot = %v\nwant= %v", inds, want)
	}
}
func TestDigitize3(t *testing.T) {
	xs := []float64{0.2, 6.4, 3.0, 1.6}
	bins := []float64{0.0, 1.0, 2.5, 4.0, 10.0}
	inds := make([]int, 0, len(xs))
	for _, v := range xs {
		inds = append(inds, floats.NearestIdx(bins, v))
	}
	want := []int{0, 3, 2, 1}
	if !reflect.DeepEqual(inds, want) {
		t.Fatalf("digitize error\ngot = %v\nwant= %v", inds, want)
	}
}
