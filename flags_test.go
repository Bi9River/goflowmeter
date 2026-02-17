package flowmeter

import (
	"testing"
	"time"
)

func TestProcessPackets_Flags_PerDirection(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, HeaderLen: 40, PSH: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Second), Direction: Forward, HeaderLen: 40, URG: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Second), Direction: Backward, HeaderLen: 32, PSH: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FwdPSHFlag != 1 || f.FwdURGFlag != 1 {
		t.Errorf("FwdPSHFlag=1 FwdURGFlag=1: got %d %d", f.FwdPSHFlag, f.FwdURGFlag)
	}
	if f.BwdPSHFlag != 1 || f.BwdURGFlag != 0 {
		t.Errorf("BwdPSHFlag=1 BwdURGFlag=0: got %d %d", f.BwdPSHFlag, f.BwdURGFlag)
	}
	if f.FwdHeaderLen != 80 || f.BwdHeaderLen != 32 {
		t.Errorf("FwdHeaderLen=80 BwdHeaderLen=32: got %d %d", f.FwdHeaderLen, f.BwdHeaderLen)
	}
}

func TestProcessPackets_Flags_FlowWide(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SYN: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(time.Millisecond), Direction: Backward, SYN: true, ACK: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(2 * time.Millisecond), Direction: Forward, ACK: true, PSH: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
		{Timestamp: base.Add(3 * time.Millisecond), Direction: Forward, FIN: true, ACK: true, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 1, DstPort: 2, Protocol: 6},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FIN != 1 || f.SYN != 2 || f.ACK != 3 || f.PSH != 1 {
		t.Errorf("FIN=1 SYN=2 ACK=3 PSH=1: got FIN=%d SYN=%d ACK=%d PSH=%d", f.FIN, f.SYN, f.ACK, f.PSH)
	}
	if f.RST != 0 || f.URG != 0 || f.CWR != 0 || f.ECE != 0 {
		t.Errorf("RST/URG/CWR/ECE should be 0: got %d %d %d %d", f.RST, f.URG, f.CWR, f.ECE)
	}
}

func TestProcessPackets_Flags_AllZero(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	packets := []PacketInfo{
		{Timestamp: base, Direction: Forward, SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 1, DstPort: 2, Protocol: 17},
	}
	pairs := ProcessPacketsWithKeys(packets)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(pairs))
	}
	f := pairs[0].Features
	if f.FwdPSHFlag != 0 || f.BwdPSHFlag != 0 || f.FwdHeaderLen != 0 {
		t.Errorf("UDP-like packet: expected zero PSH and header; got FwdPSH=%d BwdPSH=%d FwdHeaderLen=%d", f.FwdPSHFlag, f.BwdPSHFlag, f.FwdHeaderLen)
	}
}
