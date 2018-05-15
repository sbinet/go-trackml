// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package trackml exposes facilities to ease handling of TrackML datasets.
package trackml

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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

// Event stores informations about a complete HEP event.
type Event struct {
	ID    int        // event id
	Hits  []Hit      // collection of hits for this event
	Cells []Cell     // collection of cells for this event
	Ps    []Particle // collection of reconstructed particles for this event
	Mcs   []Truth    // Monte-Carlo truth for this event
}

// Delete zeroes all internal data of an Event and
// prepares that Event to be collected by the Garbage Collector.
func (evt *Event) Delete() {
	evt.Hits = nil
	evt.Cells = nil
	evt.Ps = nil
	evt.Mcs = nil
}

// ReadMcEvent reads a complete Event value from the given path+prefix,
// including Monte-Carlo informations.
func ReadMcEvent(fname string) (Event, error) {
	var (
		evt Event
		err error
	)

	evt, err = ReadEvent(fname)
	if err != nil {
		return evt, errors.Wrapf(err, "could not read hits")
	}

	evt.Mcs, err = readMcTruth(fname + "-truth.csv")
	if err != nil {
		return evt, errors.Wrapf(err, "could not read truth")
	}

	return evt, err
}

// ReadEvent reads a complete Event value from the given path+prefix,
// but without the Monte-Carlo informations.
func ReadEvent(fname string) (Event, error) {
	var (
		evt Event
		err error
	)

	id := filepath.Base(fname)
	id = strings.TrimLeft(id, "event")
	evt.ID, err = strconv.Atoi(id)
	if err != nil {
		return evt, errors.Wrapf(err, "could not infer event ID")
	}

	evt.Hits, err = readHits(fname + "-hits.csv")
	if err != nil {
		return evt, errors.Wrapf(err, "could not read hits")
	}

	evt.Cells, err = readCells(fname + "-cells.csv")
	if err != nil {
		return evt, errors.Wrapf(err, "could not read cells")
	}

	evt.Ps, err = readParticles(fname + "-particles.csv")
	if err != nil {
		return evt, errors.Wrapf(err, "could not read particles")
	}

	return evt, err
}

func readHits(fname string) ([]Hit, error) {
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

func readCells(fname string) ([]Cell, error) {
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

func readParticles(fname string) ([]Particle, error) {
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

func readMcTruth(fname string) ([]Truth, error) {
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

// EventReader is a function to read an event from a path
type EventReader func(path string) (Event, error)

// Dataset is an Event container.
//
// Dataset logically contains many Events, iterating throught the list of
// Events via the Next method.
//
// Example:
//
//   ds, err := NewDataset("./example_standard/dataset", 0, -1, nil)
//   for ds.Next() {
//       evt := ds.Event()
//   }
//   if err := ds.Err(); err != nil {
//       panic(err)
//   }
//
type Dataset struct {
	path  string
	names []string

	readEvent EventReader

	cur int
	evt Event
	err error
}

func (ds *Dataset) Next() bool {
	if ds.err != nil {
		return false
	}

	ds.cur++
	if ds.cur >= len(ds.names) {
		return false
	}
	evt, err := ds.readEvent(ds.names[ds.cur])
	if err != nil {
		ds.err = err
		return false
	}
	ds.evt = evt
	return true
}

// Event returns the current event from the dataset.
// The returned value is valid until a call to Next.
func (ds *Dataset) Event() Event {
	return ds.evt
}

func (ds *Dataset) Err() error {
	return ds.err
}

// NewDataset returns the list of datasets from name, a directory or zip file,
// containing many events data.
//
// beg and end control the number of events to iterate over.
//
// The returned Dataset will use the reader function to load events from a path.
// If reader is nil, ReadMcEvent is used.
func NewDataset(name string, beg, end int, reader EventReader) (Dataset, error) {
	if reader == nil {
		reader = ReadMcEvent
	}
	ds := Dataset{
		path:      name,
		cur:       -1,
		readEvent: reader,
	}

	f, err := os.Open(name)
	if err != nil {
		return ds, errors.WithStack(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return ds, errors.WithStack(err)
	}

	var names []string
	switch {
	case fi.IsDir():
		names, err = filepath.Glob(filepath.Join(name, "*-hits.csv"))
		// names, err = f.Readdirnames(-1)
		if err != nil {
			return ds, err
		}
		sort.Strings(names)
		ds.names = names
		ds.names = ds.names[beg:]
		if end == -1 || end > len(ds.names) {
			end = len(ds.names)
		}
		ds.names = ds.names[:end]
		for i, n := range ds.names {
			ds.names[i] = n[:len(n)-len("-hits.csv")]
		}
	default:
		return ds, errors.Wrapf(err, "could not handle path %q", name)
	}
	return ds, nil
}
