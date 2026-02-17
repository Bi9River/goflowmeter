package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Counts_SingleFlow(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// TotalFwdBytes/TotalBwdBytes use payload (CICFlowMeter).
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(time.Second), PayloadSize: 200, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), PayloadSize: 300, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(3 * time.Second), PayloadSize: 50, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.TotalFwdPackets != 2 {
		t.Errorf("TotalFwdPackets: expected 2, got %d", f.TotalFwdPackets)
	}
	if f.TotalBwdPackets != 2 {
		t.Errorf("TotalBwdPackets: expected 2, got %d", f.TotalBwdPackets)
	}
	if f.TotalFwdBytes != 300 {
		t.Errorf("TotalFwdBytes: expected 300, got %d", f.TotalFwdBytes)
	}
	if f.TotalBwdBytes != 350 {
		t.Errorf("TotalBwdBytes: expected 350, got %d", f.TotalBwdBytes)
	}
}

func TestProcessPackets_Counts_ForwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 64, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), PayloadSize: 128, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.TotalFwdPackets != 2 || f.TotalBwdPackets != 0 {
		t.Errorf("expected 2 fwd, 0 bwd; got fwd=%d bwd=%d", f.TotalFwdPackets, f.TotalBwdPackets)
	}
	if f.TotalFwdBytes != 192 || f.TotalBwdBytes != 0 {
		t.Errorf("expected TotalFwdBytes=192 TotalBwdBytes=0; got %d, %d", f.TotalFwdBytes, f.TotalBwdBytes)
	}
}

func TestProcessPackets_Counts_BackwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 256, Direction: Backward, SrcIP: "192.168.1.1", DstIP: "192.168.1.2", SrcPort: 443, DstPort: 50000, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.TotalFwdPackets != 0 || f.TotalBwdPackets != 1 {
		t.Errorf("expected 0 fwd, 1 bwd; got fwd=%d bwd=%d", f.TotalFwdPackets, f.TotalBwdPackets)
	}
	if f.TotalFwdBytes != 0 || f.TotalBwdBytes != 256 {
		t.Errorf("expected TotalFwdBytes=0 TotalBwdBytes=256; got %d, %d", f.TotalFwdBytes, f.TotalBwdBytes)
	}
}
