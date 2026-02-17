package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_ActiveIdle_NoGap(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 3 packets within 1s: one active period, no idle
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(100 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(200 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// One active period: 0 to 200ms = 200000 µs (CICFlowMeter units)
	activeUs := float64((200 * time.Millisecond).Microseconds())
	if f.ActiveTime.Mean != activeUs || f.ActiveTime.Min != activeUs || f.ActiveTime.Max != activeUs {
		t.Errorf("ActiveTime: expected %.0f µs, got mean=%f min=%f max=%f", activeUs, f.ActiveTime.Mean, f.ActiveTime.Min, f.ActiveTime.Max)
	}
	if f.IdleTime.Mean != 0 && f.IdleTime.Min != 0 {
		t.Errorf("IdleTime: expected 0 (no idle), got mean=%f min=%f", f.IdleTime.Mean, f.IdleTime.Min)
	}
}

func TestProcessPackets_ActiveIdle_OneGap(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Packets at 0, 0.1s; gap 2s; packets at 2s, 2.1s. Active: 0.1s and 0.1s. Idle: 2s (minus 0.1s? no - gap is 2s - 0.1s = 1.9s from end 0.1 to next 2)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(100 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2100 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2200 * time.Millisecond), Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Active: first period 0.1s, second period 0.1s = 100000 µs each (CICFlowMeter units)
	activeSegUs := float64((100 * time.Millisecond).Microseconds())
	if f.ActiveTime.Mean != activeSegUs || f.ActiveTime.Min != activeSegUs || f.ActiveTime.Max != activeSegUs {
		t.Errorf("ActiveTime: expected %.0f µs, got mean=%f min=%f max=%f", activeSegUs, f.ActiveTime.Mean, f.ActiveTime.Min, f.ActiveTime.Max)
	}
	// Idle: one gap from 0.1s to 2.1s = 2.0s = 2000000 µs
	idleUs := float64((2 * time.Second).Microseconds())
	if f.IdleTime.Mean != idleUs || f.IdleTime.Min != idleUs || f.IdleTime.Max != idleUs {
		t.Errorf("IdleTime: expected %.0f µs, got mean=%f min=%f max=%f", idleUs, f.IdleTime.Mean, f.IdleTime.Min, f.IdleTime.Max)
	}
}

func TestProcessPackets_ActiveIdle_SinglePacket(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// len < 2: no active/idle computed, zero stats
	if f.ActiveTime.Mean != 0 || f.IdleTime.Mean != 0 {
		t.Errorf("single packet: expected zero active/idle, got Active.Mean=%f Idle.Mean=%f", f.ActiveTime.Mean, f.IdleTime.Mean)
	}
}
