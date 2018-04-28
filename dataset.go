// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trackml

import (
	"io"

	"github.com/pkg/errors"
	"go-hep.org/x/hep/csvutil"
)

type Cell struct {
	HitID int
	Ch0   int
	Ch1   int
	Value float64
}

type Hit struct {
	HitID    int
	X, Y, Z  float64
	VolumeID int
	LayerID  int
	ModuleID int
}

type Particle struct {
	ID         int
	Vx, Vy, Vz float64
	Px, Py, Pz float64
	Q          int
	NHits      int
}

type Truth struct {
	HitID      int
	PID        int
	Tx, Ty, Tz float64
	Px, Py, Pz float64
	Weight     float64
}

func ReadEvent(fname string) ([]Hit, []Cell, []Particle, []Truth, error) {
	var (
		hits  []Hit
		cells []Cell
		parts []Particle
		mcs   []Truth
		err   error
	)

	hits, err = ReadHits(fname + "-hits.csv")
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "could not read hits")
	}

	cells, err = ReadCells(fname + "-cells.csv")
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "could not read cells")
	}

	parts, err = ReadParticles(fname + "-particles.csv")
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "could not read particles")
	}

	mcs, err = ReadMcTruth(fname + "-truth.csv")
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "could not read truth")
	}

	return hits, cells, parts, mcs, err
}

func ReadHits(fname string) ([]Hit, error) {
	var hits []Hit
	tbl, err := csvutil.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open CSV file")
	}
	defer tbl.Close()

	rows, err := tbl.ReadRows(1, -1) // skip header
	if err != nil {
		return nil, errors.Wrapf(err, "could not create row iterator")
	}
	defer rows.Close()

	for rows.Next() {
		var hit Hit
		err := rows.Scan(&hit)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read row")
		}
		hits = append(hits, hit)
	}

	if err := rows.Err(); err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "error during row iteration")
	}

	return hits, nil
}

func ReadCells(fname string) ([]Cell, error) {
	var cells []Cell
	tbl, err := csvutil.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open CSV file")
	}
	defer tbl.Close()

	rows, err := tbl.ReadRows(1, -1) // skip header
	if err != nil {
		return nil, errors.Wrapf(err, "could not create row iterator")
	}
	defer rows.Close()

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read row")
		}
		cells = append(cells, cell)
	}

	if err := rows.Err(); err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "error during row iteration")
	}

	return cells, nil
}

func ReadParticles(fname string) ([]Particle, error) {
	var ps []Particle
	tbl, err := csvutil.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open CSV file")
	}
	defer tbl.Close()

	rows, err := tbl.ReadRows(1, -1) // skip header
	if err != nil {
		return nil, errors.Wrapf(err, "could not create row iterator")
	}
	defer rows.Close()

	for rows.Next() {
		var p Particle
		err := rows.Scan(&p)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read row")
		}
		ps = append(ps, p)
	}

	if err := rows.Err(); err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "error during row iteration")
	}

	return ps, nil
}

func ReadMcTruth(fname string) ([]Truth, error) {
	var mcs []Truth
	tbl, err := csvutil.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open CSV file")
	}
	defer tbl.Close()

	rows, err := tbl.ReadRows(1, -1) // skip header
	if err != nil {
		return nil, errors.Wrapf(err, "could not create row iterator")
	}
	defer rows.Close()

	for rows.Next() {
		var mc Truth
		err := rows.Scan(&mc)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read row")
		}
		mcs = append(mcs, mc)
	}

	if err := rows.Err(); err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "error during row iteration")
	}

	return mcs, nil
}
