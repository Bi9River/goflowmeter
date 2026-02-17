package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Rates_FwdBwdPacketsPerSec(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 2 fwd, 1 bwd over 2 seconds -> 1 fwd/s, 0.5 bwd/s
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FwdPacketsPerSec != 1 || f.BwdPacketsPerSec != 0.5 {
		t.Errorf("FwdPacketsPerSec=1 BwdPacketsPerSec=0.5: got %f %f", f.FwdPacketsPerSec, f.BwdPacketsPerSec)
	}
}

func TestProcessPackets_Rates_ActDataPktFwdAndMinSegSize(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// MinSegSizeFwd = min header length in forward direction (CICFlowMeter). HeaderLen 0, 40, 40 -> min 0.
	packets := []PacketInfo{
		{Timestamp: base, HeaderLen: 0, Direction: Forward, PayloadSize: 0, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), HeaderLen: 40, Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), HeaderLen: 40, Direction: Forward, PayloadSize: 50, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// ActDataPktFwd: packets with PayloadSize >= 1 -> 2 (100 and 50)
	if f.ActDataPktFwd != 2 {
		t.Errorf("ActDataPktFwd: expected 2, got %d", f.ActDataPktFwd)
	}
	// MinSegSizeFwd: min header among forward packets = 0
	if f.MinSegSizeFwd != 0 {
		t.Errorf("MinSegSizeFwd: expected 0, got %d", f.MinSegSizeFwd)
	}
}

func TestProcessPackets_Rates_MinSegSizeForwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// MinSegSizeFwd = min header length in forward direction (CICFlowMeter).
	packets := []PacketInfo{
		{Timestamp: base, PayloadSize: 200, HeaderLen: 60, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), PayloadSize: 80, HeaderLen: 80, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), PayloadSize: 10, HeaderLen: 40, Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Min header among forward packets: 60, 80 -> 60
	if f.MinSegSizeFwd != 60 {
		t.Errorf("MinSegSizeFwd: expected 60 (min forward header), got %d", f.MinSegSizeFwd)
	}
}

func TestProcessPackets_Rates_NoForwardPackets(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Backward, PayloadSize: 50, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.ActDataPktFwd != 0 || f.MinSegSizeFwd != 0 {
		t.Errorf("no forward packets: ActDataPktFwd=0 MinSegSizeFwd=0, got %d %d", f.ActDataPktFwd, f.MinSegSizeFwd)
	}
}
