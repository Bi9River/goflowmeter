package flowmeter

import "time"

// iatForDirection filters packets by direction and returns IAT deltas in microseconds
// and total duration between consecutive same-direction packets. Caller must pass
// packets sorted by timestamp.
func iatForDirection(packets []PacketInfo, dir Direction) (deltasUs []float64, total time.Duration) {
	var dirPackets []PacketInfo
	for _, p := range packets {
		if p.Direction == dir {
			dirPackets = append(dirPackets, p)
		}
	}
	if len(dirPackets) < 2 {
		return nil, 0
	}
	for i := 1; i < len(dirPackets); i++ {
		d := dirPackets[i].Timestamp.Sub(dirPackets[i-1].Timestamp)
		total += d
		deltasUs = append(deltasUs, float64(d.Microseconds()))
	}
	return deltasUs, total
}

// computeIAT fills inter-arrival time statistics: flow-wide, forward, and backward.
// Packets must be sorted by timestamp. IAT = time between consecutive packets
// (per direction for Fwd/Bwd). Single-packet flows get zero IAT stats.
// All IAT values (Mean, Std, Min, Max) are in microseconds, matching CICFlowMeter.
func computeIAT(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) < 2 {
		return
	}
	// Flow IAT: consecutive packets (already time-ordered), in microseconds
	var flowDeltas []float64
	for i := 1; i < len(packets); i++ {
		d := packets[i].Timestamp.Sub(packets[i-1].Timestamp)
		flowDeltas = append(flowDeltas, float64(d.Microseconds()))
	}
	f.FlowIAT = StatsFromValues(flowDeltas)

	fwdDeltas, fwdTotal := iatForDirection(packets, Forward)
	if len(fwdDeltas) > 0 {
		f.FwdIATTotal = fwdTotal
		f.FwdIAT = StatsFromValues(fwdDeltas)
	}
	bwdDeltas, bwdTotal := iatForDirection(packets, Backward)
	if len(bwdDeltas) > 0 {
		f.BwdIATTotal = bwdTotal
		f.BwdIAT = StatsFromValues(bwdDeltas)
	}
}
