package flowmeter

const subflowIdleThresholdUs = 1_000_000 // 1 second in microseconds

// computeSubflow fills average packets and bytes per subflow in each direction.
// A subflow boundary is a gap > 1s between consecutive packets (any direction).
// CIC uses divisor = number of gaps (sfCount); when gaps == 0 they return 0.
// Packets must be sorted by timestamp.
func computeSubflow(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) < 2 {
		return
	}
	gaps := 0
	lastTs := packets[0].Timestamp.UnixMicro()
	for i := 1; i < len(packets); i++ {
		ts := packets[i].Timestamp.UnixMicro()
		if (ts - lastTs) > subflowIdleThresholdUs {
			gaps++
		}
		lastTs = ts
	}
	if gaps <= 0 {
		return // CIC: getSflow_* return 0 when sfCount <= 0
	}
	f.SubflowFwdPackets = float64(f.TotalFwdPackets) / float64(gaps)
	f.SubflowFwdBytes = float64(f.TotalFwdBytes) / float64(gaps)
	f.SubflowBwdPackets = float64(f.TotalBwdPackets) / float64(gaps)
	f.SubflowBwdBytes = float64(f.TotalBwdBytes) / float64(gaps)
}
