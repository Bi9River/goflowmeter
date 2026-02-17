package flowmeter

const activeIdleThresholdUs = 1_000_000 // 1 second in microseconds

// computeActiveIdle fills ActiveTime and IdleTime (min, mean, max, std).
// Active = continuous period with no gap > 1s between packets. Idle = gap > 1s.
// Durations are in microseconds, matching CICFlowMeter. Packets must be sorted by timestamp.
func computeActiveIdle(packets []PacketInfo, f *FlowFeatures) {
	if len(packets) < 2 {
		return
	}
	var activeDurations, idleDurations []float64
	startActiveUs := packets[0].Timestamp.UnixMicro()
	endActiveUs := packets[0].Timestamp.UnixMicro()

	for i := 1; i < len(packets); i++ {
		tsUs := packets[i].Timestamp.UnixMicro()
		gapUs := tsUs - endActiveUs
		if gapUs > activeIdleThresholdUs {
			if endActiveUs-startActiveUs > 0 {
				activeDurations = append(activeDurations, float64(endActiveUs-startActiveUs))
			}
			idleDurations = append(idleDurations, float64(gapUs))
			startActiveUs = tsUs
			endActiveUs = tsUs
		} else {
			endActiveUs = tsUs
		}
	}
	if endActiveUs-startActiveUs > 0 {
		activeDurations = append(activeDurations, float64(endActiveUs-startActiveUs))
	}

	if len(activeDurations) > 0 {
		f.ActiveTime = StatsFromValues(activeDurations)
	}
	if len(idleDurations) > 0 {
		f.IdleTime = StatsFromValues(idleDurations)
	}
}
