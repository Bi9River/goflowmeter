package flowmeter

// computeCounts fills total forward/backward packet counts and byte totals.
// Byte totals use payload size (TCP/UDP payload only), matching CICFlowMeter.
func computeCounts(packets []PacketInfo, f *FlowFeatures) {
	for _, p := range packets {
		if p.Direction == Forward {
			f.TotalFwdPackets++
			f.TotalFwdBytes += int64(p.PayloadSize)
		} else {
			f.TotalBwdPackets++
			f.TotalBwdBytes += int64(p.PayloadSize)
		}
	}
}
