// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trackml

import (
	"encoding/csv"
	"strconv"

	"github.com/pkg/errors"
)

// Submission creates a CSV file ready for submission to Kaggle
type Submission struct {
	w *csv.Writer
}

func NewSubmission(w *csv.Writer) *Submission {
	return &Submission{w: w}
}

func (sub *Submission) Close() error {
	sub.w.Flush()
	return sub.w.Error()
}

func (sub *Submission) Append(evt Event, trkIDs []int) error {
	defer sub.w.Flush()

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
		err := sub.w.Write(rec[:])
		if err != nil {
			return errors.Wrapf(err, "could not write row %d of event %v", i, evt.ID)
		}
	}

	return sub.w.Error()
}
