package flowmeter

// computeBasic fills flow duration, flow bytes/s, and flow packets/s.
func computeBasic(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) == 0 {
		return
	}
	first := packets[0].Timestamp
	last := packets[len(packets)-1].Timestamp
	dur := last.Sub(first)
	f.FlowDurationUs = dur.Microseconds()

	// Flow bytes = sum of payload (TCP/UDP payload only), matching CICFlowMeter.
	var totalBytes int64
	for _, p := range packets {
		totalBytes += int64(p.PayloadSize)
	}

	durSec := float64(f.FlowDurationUs) / 1e6
	if durSec > 0 {
		f.FlowBytesPerSec = float64(totalBytes) / durSec
		f.FlowPacketsPerSec = float64(len(packets)) / durSec
	}
	// When duration is 0 (single packet or same timestamp), leave rates at 0 to match CIC getfPktsPerSecond/getbPktsPerSecond.
}
