package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Basic_SingleFlow(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Flow bytes/s use payload (CICFlowMeter); set PayloadSize to match intended bytes.
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 100, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(time.Second), PayloadSize: 200, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), PayloadSize: 300, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Duration: 2 seconds = 2e6 microseconds
	if f.FlowDurationUs != 2_000_000 {
		t.Errorf("FlowDurationUs: expected 2000000, got %d", f.FlowDurationUs)
	}
	// 600 bytes / 2 sec = 300 bytes/s
	if f.FlowBytesPerSec != 300 {
		t.Errorf("FlowBytesPerSec: expected 300, got %f", f.FlowBytesPerSec)
	}
	// 3 packets / 2 sec = 1.5 packets/s
	if f.FlowPacketsPerSec != 1.5 {
		t.Errorf("FlowPacketsPerSec: expected 1.5, got %f", f.FlowPacketsPerSec)
	}
}

func TestProcessPackets_Basic_SinglePacket(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 64, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FlowDurationUs != 0 {
		t.Errorf("FlowDurationUs: expected 0 for single packet, got %d", f.FlowDurationUs)
	}
	// Duration 0: match CIC — flow and Fwd/Bwd packets-per-sec are 0
	if f.FlowBytesPerSec != 0 {
		t.Errorf("FlowBytesPerSec: expected 0 (CIC), got %f", f.FlowBytesPerSec)
	}
	if f.FlowPacketsPerSec != 0 {
		t.Errorf("FlowPacketsPerSec: expected 0 (CIC), got %f", f.FlowPacketsPerSec)
	}
}

func TestProcessPackets_Basic_Empty(t *testing.T) {
	pairs := ProcessPacketsWithKeys(nil)
	if pairs != nil {
		t.Errorf("expected nil for nil input, got len=%d", len(pairs))
	}
	pairs = ProcessPacketsWithKeys([]PacketInfo{})
	if pairs != nil {
		t.Errorf("expected nil for empty input, got len=%d", len(pairs))
	}
}

func TestProcessPackets_Basic_TwoFlows(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(time.Second), PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6},
		{Timestamp: base, PayloadSize: 200, SrcIP: "3.3.3.3", DstIP: "4.4.4.4", SrcPort: 22222, DstPort: 443, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 flows, got %d", len(pairs))
	}
	// Each flow should have correct basic stats; we don't assert order
	var foundOne, foundTwo bool
	for _, p := range pairs {
		f := p.Features
		if f.FlowDurationUs == 1_000_000 && f.FlowBytesPerSec == 200 && f.FlowPacketsPerSec == 2 {
			foundOne = true
		}
		if f.FlowDurationUs == 0 && f.FlowBytesPerSec == 0 && f.FlowPacketsPerSec == 0 {
			foundTwo = true
		}
	}
	if !foundOne || !foundTwo {
		t.Errorf("expected one flow with (1s, 200 B/s, 2 pkt/s) and one with (0s, 0 B/s, 0 pkt/s per CIC); foundOne=%v foundTwo=%v", foundOne, foundTwo)
	}
}
