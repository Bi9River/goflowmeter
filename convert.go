package flowmeter

import (
	"fmt"
	"sort"
	"time"
)

// RawPacket is a packet as seen on the wire (no direction). Callers that read PCAP
// or other sources fill this; ConvertToPacketInfo assigns direction and normalizes
// the 5-tuple so the flowmeter sees one flow per connection.
type RawPacket struct {
	Timestamp   time.Time
	HeaderLen   int
	PayloadSize int
	TCPWindow   uint16   // TCP window size (from TCP header); 0 for non-TCP or if unknown
	SrcIP       string
	DstIP       string
	SrcPort     uint16
	DstPort     uint16
	Protocol    uint8
	FIN         bool
	SYN         bool
	RST         bool
	PSH         bool
	ACK         bool
	URG         bool
	CWR         bool
	ECE         bool
}

// canonicalFlowKey returns a string key that is the same for (A,B,sp,dp) and (B,A,dp,sp).
func canonicalFlowKey(srcIP, dstIP string, srcPort, dstPort uint16, protocol uint8) string {
	c := CanonicalFlowKey(FlowKey{SrcIP: srcIP, DstIP: dstIP, SrcPort: srcPort, DstPort: dstPort, Protocol: protocol})
	return fmt.Sprintf("%s|%s|%d|%d|%d", c.SrcIP, c.DstIP, c.SrcPort, c.DstPort, c.Protocol)
}

// CanonicalFlowKey returns the flow key in CICFlowMeter format: smaller IP first (lexicographic),
// then smaller port if IPs equal. Use this so FlowKey matches CIC's Flow ID / Src-Dst column order.
func CanonicalFlowKey(k FlowKey) FlowKey {
	ipMin, ipMax := k.SrcIP, k.DstIP
	pMin, pMax := k.SrcPort, k.DstPort
	if k.SrcIP > k.DstIP {
		ipMin, ipMax = k.DstIP, k.SrcIP
		pMin, pMax = k.DstPort, k.SrcPort
	} else if k.SrcIP == k.DstIP && k.SrcPort > k.DstPort {
		pMin, pMax = k.DstPort, k.SrcPort
	}
	return FlowKey{SrcIP: ipMin, DstIP: ipMax, SrcPort: pMin, DstPort: pMax, Protocol: k.Protocol}
}

// flowIdentity is the normalized 5-tuple we use for all PacketInfo in one flow.
type flowIdentity struct {
	SrcIP    string
	DstIP    string
	SrcPort  uint16
	DstPort  uint16
	Protocol uint8
}

// forwardEndpoint is (SrcIP, SrcPort) of the first packet in the flow.
type forwardEndpoint struct {
	IP   string
	Port uint16
}

// ConvertToPacketInfo converts raw packets (e.g. from a PCAP reader) into PacketInfo
// with direction assigned by the first-packet rule and a consistent 5-tuple per flow.
// Packets from the same connection (A↔B) are normalized to one flow key; Direction
// is Forward if the packet was sent by the flow's first sender, else Backward.
func ConvertToPacketInfo(raw []RawPacket) []PacketInfo {
	if len(raw) == 0 {
		return nil
	}
	// Group by canonical flow key
	type flowState struct {
		identity flowIdentity
		forward  forwardEndpoint
		packets  []RawPacket
	}
	byFlow := make(map[string]*flowState)
	for i := range raw {
		r := &raw[i]
		key := canonicalFlowKey(r.SrcIP, r.DstIP, r.SrcPort, r.DstPort, r.Protocol)
		if byFlow[key] == nil {
			byFlow[key] = &flowState{packets: nil}
		}
		byFlow[key].packets = append(byFlow[key].packets, raw[i])
	}
	// Sort each flow by time; first packet defines forward endpoint; identity is canonical (CIC format).
	for _, state := range byFlow {
		sort.Slice(state.packets, func(i, j int) bool {
			return state.packets[i].Timestamp.Before(state.packets[j].Timestamp)
		})
		first := state.packets[0]
		can := CanonicalFlowKey(FlowKey{SrcIP: first.SrcIP, DstIP: first.DstIP, SrcPort: first.SrcPort, DstPort: first.DstPort, Protocol: first.Protocol})
		state.identity = flowIdentity{can.SrcIP, can.DstIP, can.SrcPort, can.DstPort, can.Protocol}
		state.forward = forwardEndpoint{first.SrcIP, first.SrcPort}
	}
	// Build []PacketInfo with normalized 5-tuple and direction
	out := make([]PacketInfo, 0, len(raw))
	for _, state := range byFlow {
		id := state.identity
		fwd := state.forward
		for _, r := range state.packets {
			dir := Backward
			if r.SrcIP == fwd.IP && r.SrcPort == fwd.Port {
				dir = Forward
			}
			out = append(out, PacketInfo{
				Timestamp:   r.Timestamp,
				Direction:   dir,
				HeaderLen:   r.HeaderLen,
				PayloadSize: r.PayloadSize,
				TCPWindow:   r.TCPWindow,
				SrcIP:       id.SrcIP,
				DstIP:       id.DstIP,
				SrcPort:     id.SrcPort,
				DstPort:     id.DstPort,
				Protocol:    id.Protocol,
				FIN:         r.FIN,
				SYN:         r.SYN,
				RST:         r.RST,
				PSH:         r.PSH,
				ACK:         r.ACK,
				URG:         r.URG,
				CWR:         r.CWR,
				ECE:         r.ECE,
			})
		}
	}
	return out
}
