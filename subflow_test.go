package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Subflow_NoGap(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// No gap >1s; CIC returns 0 for subflow features.
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(100 * time.Millisecond), PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(200 * time.Millisecond), PayloadSize: 100, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(300 * time.Millisecond), PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(400 * time.Millisecond), PayloadSize: 100, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.SubflowFwdPackets != 0 || f.SubflowBwdPackets != 0 {
		t.Errorf("no gap: CIC subflow=0, got %f %f", f.SubflowFwdPackets, f.SubflowBwdPackets)
	}
	if f.SubflowFwdBytes != 0 || f.SubflowBwdBytes != 0 {
		t.Errorf("no gap: CIC subflow bytes=0, got %f %f", f.SubflowFwdBytes, f.SubflowBwdBytes)
	}
}

func TestProcessPackets_Subflow_OneGap(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 2 packets, then gap >1s, then 2 packets: 2 subflows. Subflow bytes use payload.
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), PayloadSize: 100, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), PayloadSize: 200, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2001 * time.Millisecond), PayloadSize: 200, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// 1 gap: CIC divides by gaps=1 -> 2 fwd, 2 bwd, 300 bytes each
	if f.SubflowFwdPackets != 2 || f.SubflowBwdPackets != 2 {
		t.Errorf("one gap: CIC total/gaps=2, got %f %f", f.SubflowFwdPackets, f.SubflowBwdPackets)
	}
	if f.SubflowFwdBytes != 300 || f.SubflowBwdBytes != 300 {
		t.Errorf("one gap: CIC bytes/gaps=300, got %f %f", f.SubflowFwdBytes, f.SubflowBwdBytes)
	}
}

func TestProcessPackets_Subflow_SinglePacket(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// computeSubflow returns early for len < 2, so subflow fields stay 0
	if f.SubflowFwdPackets != 0 || f.SubflowFwdBytes != 0 {
		t.Errorf("single packet: subflow should be 0, got %f %f", f.SubflowFwdPackets, f.SubflowFwdBytes)
	}
}
