package flowmeter

import (
	"testing"
	"time"
)

func TestConvertToPacketInfo_OneFlow(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	raw := []RawPacket{
		{Timestamp: base, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), SrcIP: "2.2.2.2", DstIP: "1.1.1.1", SrcPort: 80, DstPort: 11111, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 11111, DstPort: 80, Protocol: 6},
	}
	packets := ConvertToPacketInfo(raw)
	if len(packets) != 3 {
		t.Fatalf("expected 3 packets, got %d", len(packets))
	}
	// First packet (1.1.1.1:11111 -> 2.2.2.2:80) defines flow; all PacketInfo should have that 5-tuple
	for i, p := range packets {
		if p.SrcIP != "1.1.1.1" || p.DstIP != "2.2.2.2" || p.SrcPort != 11111 || p.DstPort != 80 {
			t.Errorf("packet %d: expected flow 1.1.1.1:11111 -> 2.2.2.2:80, got %s:%d -> %s:%d", i, p.SrcIP, p.SrcPort, p.DstIP, p.DstPort)
		}
	}
	// Packet 0: from 1.1.1.1 -> Forward. Packet 1: from 2.2.2.2 -> Backward. Packet 2: from 1.1.1.1 -> Forward
	if packets[0].Direction != Forward || packets[1].Direction != Backward || packets[2].Direction != Forward {
		t.Errorf("directions: expected Fwd,Bwd,Fwd got %v %v %v", packets[0].Direction, packets[1].Direction, packets[2].Direction)
	}
}

func TestConvertToPacketInfo_Empty(t *testing.T) {
	packets := ConvertToPacketInfo(nil)
	if packets != nil {
		t.Errorf("expected nil, got len %d", len(packets))
	}
	packets = ConvertToPacketInfo([]RawPacket{})
	if packets != nil {
		t.Errorf("expected nil for empty slice, got len %d", len(packets))
	}
}
