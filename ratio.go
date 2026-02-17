package flowmeter

// computeRatio sets DownUpRatio per CIC: integer division then double.
func computeRatio(packets []PacketInfo, f *FlowFeatures) {
	if f.TotalFwdPackets > 0 {
		f.DownUpRatio = float64(f.TotalBwdPackets / f.TotalFwdPackets)
	}
}
