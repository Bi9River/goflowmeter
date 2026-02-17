package flowmeter

// computeRates fills per-direction packet rates (Fwd/Bwd packets/s), act_data_pkt_forward
// (count of forward packets with payload >= 1 byte), and min segment size in forward direction.
// MinSegSizeFwd is the minimum header length (bytes) among forward packets, matching CICFlowMeter.
func computeRates(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) == 0 {
		return
	}
	durSec := float64(f.FlowDurationUs) / 1e6
	if durSec > 0 {
		f.FwdPacketsPerSec = float64(f.TotalFwdPackets) / durSec
		f.BwdPacketsPerSec = float64(f.TotalBwdPackets) / durSec
	}
	// When duration is 0, leave at 0 to match CIC getfPktsPerSecond/getbPktsPerSecond (they return 0).
	first := true
	for _, p := range packets {
		if p.Direction != Forward {
			continue
		}
		if p.PayloadSize >= 1 {
			f.ActDataPktFwd++
		}
		if first {
			f.MinSegSizeFwd = p.HeaderLen
			first = false
		} else if p.HeaderLen < f.MinSegSizeFwd {
			f.MinSegSizeFwd = p.HeaderLen
		}
	}
}
