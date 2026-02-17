package flowmeter

// computePacketLen fills packet length statistics: fwd/bwd/total min, max, mean, std,
// variance, avg packet size, and avg segment sizes per direction.
// All length stats use payload size (TCP/UDP payload only), matching CICFlowMeter.
// CIC double-counts the first packet payload in flowLengthStats; we replicate for flow-level PacketLen.
func computePacketLen(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) == 0 {
		return
	}
	var fwdLengths, bwdLengths, allLengths []float64
	firstPayload := float64(packets[0].PayloadSize)
	allLengths = append(allLengths, firstPayload, firstPayload)
	for i := 1; i < len(packets); i++ {
		p := packets[i]
		l := float64(p.PayloadSize)
		allLengths = append(allLengths, l)
		if p.Direction == Forward {
			fwdLengths = append(fwdLengths, l)
		} else {
			bwdLengths = append(bwdLengths, l)
		}
	}
	if packets[0].Direction == Forward {
		fwdLengths = append([]float64{firstPayload}, fwdLengths...)
	} else {
		bwdLengths = append([]float64{firstPayload}, bwdLengths...)
	}

	if len(fwdLengths) > 0 {
		f.FwdPacketLen = StatsFromValues(fwdLengths)
		f.AvgFwdSegmentSize = f.FwdPacketLen.Mean
	}
	if len(bwdLengths) > 0 {
		f.BwdPacketLen = StatsFromValues(bwdLengths)
		f.AvgBwdSegmentSize = f.BwdPacketLen.Mean
	}

	f.PacketLen = StatsFromValues(allLengths)
	f.PacketLenVar = Variance(allLengths)
	f.MinPacketLen = int(f.PacketLen.Min)
	f.MaxPacketLen = int(f.PacketLen.Max)
	f.PacketLenMean = f.PacketLen.Mean
	f.PacketLenStd = f.PacketLen.Std

	var flowSum float64
	for _, v := range allLengths {
		flowSum += v
	}
	f.AvgPacketSize = flowSum / float64(len(packets))
}
