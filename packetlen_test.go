package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_PacketLen_SingleFlow(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// All packet length stats use payload (CICFlowMeter). Fwd payloads: 50, 150. Bwd: 250.
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 50, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), PayloadSize: 150, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), PayloadSize: 250, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Overall payload: 50, 150, 250 -> min=50, max=250, mean=150
	if f.MinPacketLen != 50 || f.MaxPacketLen != 250 {
		t.Errorf("MinPacketLen/MaxPacketLen: expected 50/250, got %d/%d", f.MinPacketLen, f.MaxPacketLen)
	}
	if f.PacketLenMean != 125 {
		t.Errorf("PacketLenMean: expected 125 (CIC), got %f", f.PacketLenMean)
	}
	if f.AvgPacketSize != 500.0/3.0 {
		t.Errorf("AvgPacketSize: expected 500/3 (CIC), got %f", f.AvgPacketSize)
	}
	// Fwd payloads: 50, 150 -> mean 100, min 50, max 150
	if f.FwdPacketLen.Min != 50 || f.FwdPacketLen.Max != 150 || f.FwdPacketLen.Mean != 100 {
		t.Errorf("FwdPacketLen: expected min=50 max=150 mean=100, got min=%f max=%f mean=%f", f.FwdPacketLen.Min, f.FwdPacketLen.Max, f.FwdPacketLen.Mean)
	}
	if f.AvgFwdSegmentSize != 100 {
		t.Errorf("AvgFwdSegmentSize: expected 100, got %f", f.AvgFwdSegmentSize)
	}
	// Bwd payload: 250 only
	if f.BwdPacketLen.Min != 250 || f.BwdPacketLen.Max != 250 || f.BwdPacketLen.Mean != 250 {
		t.Errorf("BwdPacketLen: expected 250 throughout, got min=%f max=%f mean=%f", f.BwdPacketLen.Min, f.BwdPacketLen.Max, f.BwdPacketLen.Mean)
	}
	if f.AvgBwdSegmentSize != 250 {
		t.Errorf("AvgBwdSegmentSize: expected 250, got %f", f.AvgBwdSegmentSize)
	}
}

func TestProcessPackets_PacketLen_StdAndVar(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Packet length stats use payload. Payloads 10, 20, 30 -> mean 20, variance 100, std 10
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 10, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), PayloadSize: 20, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), PayloadSize: 30, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.PacketLen.Std < 9.5 || f.PacketLen.Std > 9.6 {
		t.Errorf("PacketLen.Std: expected ~9.57 (CIC), got %f", f.PacketLen.Std)
	}
	if f.PacketLenVar < 91 || f.PacketLenVar > 92 {
		t.Errorf("PacketLenVar: expected ~91.67 (CIC), got %f", f.PacketLenVar)
	}
}

func TestProcessPackets_PacketLen_SinglePacket(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 40, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.PacketLen.Std != 0 || f.PacketLenVar != 0 {
		t.Errorf("single packet: expected Std=0 Var=0, got Std=%f Var=%f", f.PacketLen.Std, f.PacketLenVar)
	}
	if f.AvgFwdSegmentSize != 40 {
		t.Errorf("AvgFwdSegmentSize: expected 40, got %f", f.AvgFwdSegmentSize)
	}
	// Bwd has no packets -> AvgBwdSegmentSize and BwdPacketLen stay 0
	if f.AvgBwdSegmentSize != 0 {
		t.Errorf("AvgBwdSegmentSize: expected 0 (no bwd packets), got %f", f.AvgBwdSegmentSize)
	}
}
