// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trackml

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// Submission creates a CSV file ready for submission to Kaggle
type Submission struct {
	f   *os.File
	gw  *gzip.Writer
	csv *csv.Writer
}

func NewSubmission() (*Submission, error) {
	f, err := os.Create("submission.csv.gz")
	if err != nil {
		return nil, err
	}
	sub := &Submission{f: f}
	sub.gw = gzip.NewWriter(f)
	sub.csv = csv.NewWriter(sub.gw)

	err = sub.csv.Write([]string{"event_id", "hit_id", "track_id"})
	if err != nil {
		return nil, err
	}
	sub.csv.Flush()
	return sub, sub.csv.Error()
}

func (sub *Submission) Close() error {
	sub.csv.Flush()
	err1 := sub.csv.Error()
	err2 := sub.gw.Close()
	err3 := sub.f.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	return nil
}

func (sub *Submission) Append(evt Event, trkIDs []int) error {
	defer sub.csv.Flush()

	if len(evt.Hits) != len(trkIDs) {
		return errors.Errorf("length mismatch")
	}
	var (
		rec   [3]string
		evtid = strconv.Itoa(evt.ID)
	)
	for i, tid := range trkIDs {
		rec[0] = evtid
		rec[1] = strconv.Itoa(evt.Hits[i].HitID)
		rec[2] = strconv.Itoa(tid)
		err := sub.csv.Write(rec[:])
		if err != nil {
			return errors.Wrapf(err, "could not write row %d of event %v", i, evt.ID)
		}
	}

	return sub.csv.Error()
}
