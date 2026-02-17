package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_InitWin_ZeroWithoutTCPWindow(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), Direction: Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// TCPWindow not set (0); init window bytes stay 0
	if f.InitWinBytesFwd != 0 || f.InitWinBytesBwd != 0 {
		t.Errorf("InitWinBytesFwd and InitWinBytesBwd should be 0 without TCP window in PacketInfo: got %d %d", f.InitWinBytesFwd, f.InitWinBytesBwd)
	}
}

func TestProcessPackets_InitWin_ForwardOnly(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.InitWinBytesFwd != 0 || f.InitWinBytesBwd != 0 {
		t.Errorf("expected 0: got Fwd=%d Bwd=%d", f.InitWinBytesFwd, f.InitWinBytesBwd)
	}
}

func TestProcessPackets_InitWin_FirstFwdLastBwd(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// CICFlowMeter: InitWinBytesFwd = first forward packet's TCP window;
	// InitWinBytesBwd = last backward packet's TCP window.
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, TCPWindow: 65535, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), Direction: Backward, TCPWindow: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), Direction: Forward, TCPWindow: 32768, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(3 * time.Millisecond), Direction: Backward, TCPWindow: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.InitWinBytesFwd != 65535 {
		t.Errorf("InitWinBytesFwd: expected 65535 (first forward), got %d", f.InitWinBytesFwd)
	}
	if f.InitWinBytesBwd != 200 {
		t.Errorf("InitWinBytesBwd: expected 200 (last backward), got %d", f.InitWinBytesBwd)
	}
}
