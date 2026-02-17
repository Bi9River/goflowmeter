package flowmeter

// computeInitWin fills initial window bytes (InitWinBytesFwd, InitWinBytesBwd) to match
// CICFlowMeter: InitWinBytesFwd = TCP window of the first forward packet;
// InitWinBytesBwd = TCP window of the last backward packet. For non-TCP, TCPWindow is 0.
// Packets must be sorted by timestamp.
func computeInitWin(packets []PacketInfo, f *FlowFeatures) {
	for _, p := range packets {
		if p.Direction == Forward {
			f.InitWinBytesFwd = int64(p.TCPWindow)
			break
		}
	}
	for i := len(packets) - 1; i >= 0; i-- {
		if packets[i].Direction == Backward {
			f.InitWinBytesBwd = int64(packets[i].TCPWindow)
			break
		}
	}
}
