package flowmeter

import (
	"testing"
	"time"
)

// TestIO_BareMinimum runs the full pipeline (RawPacket → ConvertToPacketInfo → ProcessPacketsWithKeys)
// and asserts the result without any file I/O.
func TestIO_BareMinimum(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Stage 1: Raw packets (as from a PCAP reader) — no direction; both sides of a connection as seen on wire.
	raw := []RawPacket{
		{Timestamp: base, HeaderLen: 40, PayloadSize: 60, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6, SYN: true},
		{Timestamp: base.Add(10 * time.Millisecond), HeaderLen: 40, PayloadSize: 40, SrcIP: "2.2.2.2", DstIP: "1.1.1.1", SrcPort: 80, DstPort: 11111, Protocol: 6, SYN: true, ACK: true},
		{Timestamp: base.Add(20 * time.Millisecond), HeaderLen: 40, PayloadSize: 160, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6, PSH: true, ACK: true},
		{Timestamp: base.Add(30 * time.Millisecond), HeaderLen: 40, PayloadSize: 50, SrcIP: "2.2.2.2", DstIP: "1.1.1.1", SrcPort: 80, DstPort: 11111, Protocol: 6, ACK: true},
		{Timestamp: base.Add(40 * time.Millisecond), HeaderLen: 40, PayloadSize: 24, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6, FIN: true, ACK: true},
		{Timestamp: base.Add(time.Second), HeaderLen: 40, PayloadSize: 60, SrcIP: "3.3.3.3", DstIP: "4.4.4.4", SrcPort: 22222, DstPort: 443, Protocol: 6},
		{Timestamp: base.Add(time.Second + 50*time.Millisecond), HeaderLen: 40, PayloadSize: 80, SrcIP: "3.3.3.3", DstIP: "4.4.4.4", SrcPort: 22222, DstPort: 443, Protocol: 6},
	}

	// Stage 2: Converter assigns direction and normalizes 5-tuple per flow.
	packets := ConvertToPacketInfo(raw)
	if len(packets) != len(raw) {
		t.Errorf("ConvertToPacketInfo: got %d packets, want %d", len(packets), len(raw))
	}

	// Stage 3: Flowmeter computes feature vectors per flow.
	pairs := ProcessPacketsWithKeys(packets)
	// Two flows: 1.1.1.1:11111 <-> 2.2.2.2:80 (5 packets) and 3.3.3.3:22222 -> 4.4.4.4:443 (2 packets).
	if len(pairs) != 2 {
		t.Fatalf("ProcessPacketsWithKeys: got %d flows, want 2", len(pairs))
	}
	// Sanity check: first flow has 5 packets (3 fwd + 2 bwd), second has 2.
	var totalFwd, totalBwd int
	for _, p := range pairs {
		totalFwd += p.Features.TotalFwdPackets
		totalBwd += p.Features.TotalBwdPackets
	}
	if totalFwd+totalBwd != len(packets) {
		t.Errorf("total packets across flows: got %d fwd + %d bwd = %d, want %d", totalFwd, totalBwd, totalFwd+totalBwd, len(packets))
	}
}
