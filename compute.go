package flowmeter

import "sort"

// FlowWithKey pairs a flow key with its computed features (for per-IP aggregation by SrcIP).
type FlowWithKey struct {
	Key      FlowKey
	Features FlowFeatures
}

// ProcessPacketsWithKeys groups packets by flow (5-tuple), then computes flow features
// for each flow. Call once per time window's packet set. Returns one FlowWithKey per flow
// so the caller always knows which features belong to which flow; order is undefined.
func ProcessPacketsWithKeys(packets []PacketInfo) []FlowWithKey {
	if len(packets) == 0 {
		return nil
	}
	// Group by canonical flow key (CICFlowMeter format: smaller IP first, then smaller port)
	byFlow := make(map[FlowKey][]PacketInfo)
	for _, p := range packets {
		k := CanonicalFlowKey(p.Key())
		byFlow[k] = append(byFlow[k], p)
	}
	out := make([]FlowWithKey, 0, len(byFlow))
	for key, flowPackets := range byFlow {
		// Sort by time for duration, IAT, and other time-based features
		sort.Slice(flowPackets, func(i, j int) bool {
			return flowPackets[i].Timestamp.Before(flowPackets[j].Timestamp)
		})
		f := FlowFeatures{}
		computeBasic(flowPackets, &f)
		computeCounts(flowPackets, &f)
		computePacketLen(flowPackets, &f)
		computeIAT(flowPackets, &f)
		computeFlags(flowPackets, &f)
		computeRates(flowPackets, &f)
		computeRatio(flowPackets, &f)
		computeBulk(flowPackets, &f)
		computeSubflow(flowPackets, &f)
		computeActiveIdle(flowPackets, &f)
		computeInitWin(flowPackets, &f)
		out = append(out, FlowWithKey{Key: key, Features: f})
	}
	return out
}
