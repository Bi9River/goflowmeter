package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Bulk_ForwardBulk(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 5 consecutive forward packets with payload, <1s gap: one bulk, 5 packets, total payload 100+200+300+400+500=1500
	packets := []PacketInfo{
		{Timestamp: base,  Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(100 * time.Millisecond), Direction: Forward, PayloadSize: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(200 * time.Millisecond), Direction: Forward, PayloadSize: 300, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(300 * time.Millisecond),  Direction: Forward, PayloadSize: 400, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(400 * time.Millisecond),  Direction: Forward, PayloadSize: 500, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FwdAvgPacketsPerBulk != 5 {
		t.Errorf("FwdAvgPacketsPerBulk: expected 5, got %f", f.FwdAvgPacketsPerBulk)
	}
	if f.FwdAvgBytesPerBulk != 1500 {
		t.Errorf("FwdAvgBytesPerBulk: expected 1500, got %f", f.FwdAvgBytesPerBulk)
	}
	// Rate = 1500 bytes / 0.4 s = 3750 bytes/s
	if f.FwdAvgBulkRate != 3750 {
		t.Errorf("FwdAvgBulkRate: expected 3750, got %f", f.FwdAvgBulkRate)
	}
}

func TestProcessPackets_Bulk_NoBulkWhenLessThanFour(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base,  Direction: Forward, PayloadSize: 50, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond),  Direction: Forward, PayloadSize: 50, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond),  Direction: Forward, PayloadSize: 50, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FwdAvgBytesPerBulk != 0 || f.FwdAvgPacketsPerBulk != 0 || f.FwdAvgBulkRate != 0 {
		t.Errorf("expected zero bulk (only 3 packets): got bytes=%f pkts=%f rate=%f", f.FwdAvgBytesPerBulk, f.FwdAvgPacketsPerBulk, f.FwdAvgBulkRate)
	}
}

func TestProcessPackets_Bulk_IdleResetsBulk(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 4 packets, then gap >1s, then 4 more: two bulks
	packets := []PacketInfo{
		{Timestamp: base,  Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(10 * time.Millisecond),  Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(20 * time.Millisecond),  Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(30 * time.Millisecond),  Direction: Forward, PayloadSize: 100, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second),  Direction: Forward, PayloadSize: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2010 * time.Millisecond),  Direction: Forward, PayloadSize: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2020 * time.Millisecond),  Direction: Forward, PayloadSize: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2030 * time.Millisecond),  Direction: Forward, PayloadSize: 200, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	// Two bulks: first 4 pkts 400 bytes, second 4 pkts 800 bytes. Avg bytes per bulk = (400+800)/2 = 600
	if f.FwdAvgBytesPerBulk != 600 {
		t.Errorf("FwdAvgBytesPerBulk: expected 600 (two bulks 400 and 800), got %f", f.FwdAvgBytesPerBulk)
	}
	if f.FwdAvgPacketsPerBulk != 4 {
		t.Errorf("FwdAvgPacketsPerBulk: expected 4, got %f", f.FwdAvgPacketsPerBulk)
	}
}
