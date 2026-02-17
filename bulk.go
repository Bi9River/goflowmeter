package flowmeter

const bulkIdleThresholdUs = 1_000_000 // 1 second in microseconds

// bulkDirState holds state for computing bulk stats in one direction.
type bulkDirState struct {
	start, lastTs                        int64
	pktHelper, stateCount                int
	pktTotal, sizeTotal, durUs, sizeHelper int64
}

// updateBulkDir updates state for one packet in the given direction.
// otherLastTs is the timestamp of the most recent packet in the opposite direction;
// if otherLastTs > state.start the current bulk is reset (cross-direction interrupt).
func updateBulkDir(state *bulkDirState, ts, size int64, otherLastTs int64) {
	if otherLastTs > state.start {
		state.start = 0
	}
	if size <= 0 {
		return
	}
	if state.start == 0 {
		state.start = ts
		state.lastTs = ts
		state.pktHelper = 1
		state.sizeHelper = size
		return
	}
	if (ts - state.lastTs) > bulkIdleThresholdUs {
		state.start = ts
		state.lastTs = ts
		state.pktHelper = 1
		state.sizeHelper = size
		return
	}
	state.pktHelper++
	state.sizeHelper += size
	if state.pktHelper == 4 {
		state.stateCount++
		state.pktTotal += 4
		state.sizeTotal += state.sizeHelper
		state.durUs += ts - state.start
	} else if state.pktHelper > 4 {
		state.pktTotal++
		state.sizeTotal += size
		state.durUs += ts - state.lastTs
	}
	state.lastTs = ts
}

// computeBulk fills Fwd/Bwd avg bytes per bulk, avg packets per bulk, and avg bulk rate.
// A bulk is a run of at least 4 same-direction packets with payload > 0 and gap <= 1s.
// Packets must be sorted by timestamp.
func computeBulk(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) == 0 {
		return
	}
	var fState, bState bulkDirState
	var lastFwdTs, lastBwdTs int64

	for _, p := range packets {
		ts := p.Timestamp.UnixMicro()
		size := int64(p.PayloadSize)
		if p.Direction == Forward {
			updateBulkDir(&fState, ts, size, lastBwdTs)
			lastFwdTs = ts
		} else {
			updateBulkDir(&bState, ts, size, lastFwdTs)
			lastBwdTs = ts
		}
	}

	if fState.stateCount > 0 {
		f.FwdAvgBytesPerBulk = float64(fState.sizeTotal) / float64(fState.stateCount)
		f.FwdAvgPacketsPerBulk = float64(fState.pktTotal) / float64(fState.stateCount)
		if fState.durUs > 0 {
			f.FwdAvgBulkRate = float64(fState.sizeTotal) / (float64(fState.durUs) / 1e6)
		}
	}
	if bState.stateCount > 0 {
		f.BwdAvgBytesPerBulk = float64(bState.sizeTotal) / float64(bState.stateCount)
		f.BwdAvgPacketsPerBulk = float64(bState.pktTotal) / float64(bState.stateCount)
		if bState.durUs > 0 {
			f.BwdAvgBulkRate = float64(bState.sizeTotal) / (float64(bState.durUs) / 1e6)
		}
	}
}
