package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Ratio_Mixed(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 2 fwd, 3 bwd -> ratio 3/2 = 1 (CIC integer)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(3 * time.Millisecond), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(4 * time.Millisecond), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.DownUpRatio != 1 {
		t.Errorf("DownUpRatio: expected 1 (CIC integer 3/2), got %f", f.DownUpRatio)
	}
}

func TestProcessPackets_Ratio_ForwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.DownUpRatio != 0 {
		t.Errorf("DownUpRatio: expected 0 (no bwd), got %f", f.DownUpRatio)
	}
}

func TestProcessPackets_Ratio_BackwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// TotalFwdPackets == 0 -> ratio stays 0
	if f.DownUpRatio != 0 {
		t.Errorf("DownUpRatio: expected 0 (no fwd), got %f", f.DownUpRatio)
	}
}
