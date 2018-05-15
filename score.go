// Copyright 2018 The go-trackml Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trackml

import (
	"log"
	"sort"
)

// Score computes the TrackML event score for a single event.
func Score(evt Event, trkIDs []int) float64 {
	sum := 0.0
	trks := analyzeTracks(evt.Mcs, evt.Hits, trkIDs)
	for _, trk := range trks {
		var (
			majHits   = float64(trk.MajHits)
			nhits     = float64(trk.Hits)
			majPHits  = float64(trk.MajPHits)
			purityRec = majHits / nhits
			purityMaj = majHits / majPHits
			goodTrk   = 0.5 < purityRec && 0.5 < purityMaj
		)
		if goodTrk {
			sum += trk.MajWeight
		}
	}

	return sum
}

func extractMcHitIDs(mcs []Truth) []int {
	ids := make([]int, len(mcs))
	for i, mc := range mcs {
		ids[i] = mc.HitID
	}
	return ids
}

func extractHitIDs(hits []Hit) []int {
	ids := make([]int, len(hits))
	for i, hit := range hits {
		ids[i] = hit.HitID
	}
	return ids
}

type ntuple struct {
	mcs    []Truth
	hits   []Hit
	trkIDs []int
}

func (nt ntuple) print(beg, end int) {
	mids := make([]int, len(nt.mcs[beg:end]))
	pids := make([]int, len(nt.mcs[beg:end]))
	hids := make([]int, len(nt.hits[beg:end]))
	tids := make([]int, len(nt.trkIDs[beg:end]))
	wgts := make([]float64, len(nt.mcs[beg:end]))

	for i := range mids {
		j := i + beg
		mids[i] = nt.mcs[j].HitID
		pids[i] = nt.mcs[j].PID
		hids[i] = nt.hits[j].HitID
		tids[i] = nt.trkIDs[j]
		wgts[i] = nt.mcs[j].Weight
	}

	log.Printf("nt.slice[%d:%d]", beg, end)
	for i := range mids {
		log.Printf("%v\t%v\t%v\t%v\t%v", mids[i], hids[i], pids[i], wgts[i], tids[i])
	}
}

func (nt ntuple) display(pid int) {
	var (
		mids []int
		pids []int
		hids []int
		tids []int
		wgts []float64
	)

	for i, mc := range nt.mcs {
		if mc.PID != pid {
			continue
		}
		mids = append(mids, mc.HitID)
		pids = append(pids, mc.PID)
		hids = append(hids, nt.hits[i].HitID)
		tids = append(tids, nt.trkIDs[i])
		wgts = append(wgts, mc.Weight)
	}

	log.Printf("nt.display(%d)", pid)
	for i := range mids {
		log.Printf("%v\t%v\t%v\t%v\t%v", mids[i], hids[i], pids[i], wgts[i], tids[i])
	}
}

func analyzeTracks(mcs []Truth, hits []Hit, trkIDs []int) []recTrack {
	// compute the true number of hits for each particle id
	pids := make(map[int]int, len(mcs))
	totalWeight := 0.0
	for _, mc := range mcs {
		pids[mc.PID]++
		totalWeight += mc.Weight
	}

	nt := ntuple{mcs, hits, trkIDs}
	//	log.Printf("===========================")
	//	nt.print(0, 10)
	//	log.Printf("---------------------------")
	//	nt.print(len(mcs)-10, len(mcs))
	//	log.Printf("===========================")

	invTotWeight := 1 / totalWeight

	sort.Sort(byTrackAndPID{nt})
	defer sort.Sort(byHitID{nt})

	//	log.Printf("")
	//	log.Printf("===========================")
	//	nt.print(0, 10)
	//	log.Printf("---------------------------")
	//	nt.print(len(mcs)-30, len(mcs))
	//	log.Printf("===========================")

	//	log.Printf("+++++++++++++++++++++++++++")
	//	nt.display(117108158840700928)
	//	log.Printf("+++++++++++++++++++++++++++")

	type tuple struct {
		pid    int
		nhits  int
		weight float64
	}

	var (
		trks []recTrack

		recTrkID = -1
		recHits  = 0

		cur = tuple{-1, 0, 0}
		maj = tuple{-1, 0, 0}
	)

	for i := range trkIDs {
		tid := trkIDs[i]
		pid := mcs[i].PID

		// reached the next track: need to finalize the current one
		if recTrkID != -1 && recTrkID != tid {
			if maj.nhits < cur.nhits {
				maj = cur
			}
			trks = append(trks, recTrack{
				ID:        recTrkID,
				Hits:      recHits,
				MajPID:    maj.pid,
				MajPHits:  pids[maj.pid],
				MajHits:   maj.nhits,
				MajWeight: maj.weight * invTotWeight,
			})
		}

		// set running values for next track (or first)
		if recTrkID != tid {
			recTrkID = tid
			recHits = 1
			cur.pid = pid
			cur.nhits = 1
			cur.weight = mcs[i].Weight
			maj.pid = -1
			maj.nhits = 0
			maj.weight = 0
			continue
		}

		// hit is part of the current reconstructed track
		recHits++

		// reached new particle within the same reconstructed track
		if cur.pid != pid {
			// check whether the last particle has more hits than the majority one
			if maj.nhits < cur.nhits {
				maj = cur
			}
			// reset running values for the current particle
			cur.pid = pid
			cur.nhits = 1
			cur.weight = mcs[i].Weight
		} else {
			// hit belongs to the same particle within the same reconstructed track
			cur.nhits++
			cur.weight += mcs[i].Weight
		}
	}

	// last track not handled inside the above loop
	if maj.nhits < cur.nhits {
		maj = cur
	}
	trks = append(trks, recTrack{
		ID:        recTrkID,
		Hits:      recHits,
		MajPID:    maj.pid,
		MajPHits:  pids[maj.pid],
		MajHits:   maj.nhits,
		MajWeight: maj.weight * invTotWeight,
	})
	return trks
}

type recTrack struct {
	ID        int
	Hits      int
	MajPID    int
	MajPHits  int
	MajHits   int
	MajWeight float64
}

type byTrackAndPID struct {
	nt ntuple
}

func (evt byTrackAndPID) Len() int { return len(evt.nt.hits) }
func (evt byTrackAndPID) Swap(i, j int) {
	evt.nt.hits[i], evt.nt.hits[j] = evt.nt.hits[j], evt.nt.hits[i]
	evt.nt.mcs[i], evt.nt.mcs[j] = evt.nt.mcs[j], evt.nt.mcs[i]
	evt.nt.trkIDs[i], evt.nt.trkIDs[j] = evt.nt.trkIDs[j], evt.nt.trkIDs[i]
}

func (evt byTrackAndPID) Less(i, j int) bool {
	itrk := evt.nt.trkIDs[i]
	jtrk := evt.nt.trkIDs[j]
	ipid := evt.nt.mcs[i].PID
	jpid := evt.nt.mcs[j].PID
	switch {
	case itrk < jtrk:
		return true
	case itrk == jtrk:
		switch {
		case ipid < jpid:
			return true
		case ipid == jpid:
			return evt.nt.mcs[i].HitID < evt.nt.mcs[j].HitID
		}
	}
	return false
}

type byHitID struct {
	nt ntuple
}

func (evt byHitID) Len() int { return len(evt.nt.hits) }
func (evt byHitID) Swap(i, j int) {
	evt.nt.hits[i], evt.nt.hits[j] = evt.nt.hits[j], evt.nt.hits[i]
	evt.nt.mcs[i], evt.nt.mcs[j] = evt.nt.mcs[j], evt.nt.mcs[i]
	evt.nt.trkIDs[i], evt.nt.trkIDs[j] = evt.nt.trkIDs[j], evt.nt.trkIDs[i]
}
func (evt byHitID) Less(i, j int) bool { return evt.nt.hits[i].HitID < evt.nt.hits[j].HitID }
