package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_IAT_FlowIAT(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Packets at 0, 1s, 2s -> flow IATs = 1e6 µs, 1e6 µs -> mean 1e6, min 1e6, max 1e6, std 0 (CICFlowMeter units)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	oneSecUs := float64(time.Second.Microseconds())
	if f.FlowIAT.Mean != oneSecUs || f.FlowIAT.Min != oneSecUs || f.FlowIAT.Max != oneSecUs || f.FlowIAT.Std != 0 {
		t.Errorf("FlowIAT: expected mean=min=max=%.0f std=0, got mean=%f min=%f max=%f std=%f", oneSecUs, f.FlowIAT.Mean, f.FlowIAT.Min, f.FlowIAT.Max, f.FlowIAT.Std)
	}
}

func TestProcessPackets_IAT_FwdBwdIAT(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Fwd at 0, 2s -> one fwd IAT = 2s. Bwd at 1s -> no bwd IAT (single bwd packet)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Fwd IAT: 2s between first and second fwd packet = 2e6 µs
	twoSecUs := float64((2 * time.Second).Microseconds())
	if f.FwdIAT.Mean != twoSecUs || f.FwdIAT.Min != twoSecUs || f.FwdIAT.Max != twoSecUs {
		t.Errorf("FwdIAT: expected %.0f throughout, got mean=%f min=%f max=%f", twoSecUs, f.FwdIAT.Mean, f.FwdIAT.Min, f.FwdIAT.Max)
	}
	if f.FwdIATTotal != 2*time.Second {
		t.Errorf("FwdIATTotal: expected 2s, got %v", f.FwdIATTotal)
	}
	// Bwd: only one packet -> no IAT, zero stats and zero total
	if f.BwdIAT.Mean != 0 || f.BwdIATTotal != 0 {
		t.Errorf("BwdIAT: expected 0 (single bwd packet), got mean=%f total=%v", f.BwdIAT.Mean, f.BwdIATTotal)
	}
}

func TestProcessPackets_IAT_SinglePacket(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FlowIAT.Mean != 0 || f.FwdIATTotal != 0 || f.BwdIATTotal != 0 {
		t.Errorf("single packet: expected zero IAT stats and totals, got FlowIAT.Mean=%f FwdIATTotal=%v BwdIATTotal=%v", f.FlowIAT.Mean, f.FwdIATTotal, f.BwdIATTotal)
	}
}

func TestProcessPackets_IAT_TwoPackets(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(500 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// One flow IAT = 500ms = 500000 µs
	fiveHundredMsUs := float64((500 * time.Millisecond).Microseconds())
	if f.FlowIAT.Mean != fiveHundredMsUs || f.FlowIAT.Min != fiveHundredMsUs || f.FlowIAT.Max != fiveHundredMsUs {
		t.Errorf("FlowIAT: expected %.0f µs, got mean=%f min=%f max=%f", fiveHundredMsUs, f.FlowIAT.Mean, f.FlowIAT.Min, f.FlowIAT.Max)
	}
	if f.FwdIATTotal != 500*time.Millisecond {
		t.Errorf("FwdIATTotal: expected 500ms, got %v", f.FwdIATTotal)
	}
}
