package flowmeter

import "time"

// Direction is forward (initiator→responder) or backward (responder→initiator).
// Caller assigns direction when building PacketInfo (e.g. first packet of flow = forward).
type Direction int

const (
	Forward Direction = iota
	Backward
)

// FlowKey identifies a flow (5-tuple). Comparable for use as map key.
type FlowKey struct {
	SrcIP     string
	DstIP     string
	SrcPort   uint16
	DstPort   uint16
	Protocol  uint8 // e.g. 6=TCP, 17=UDP
}

// PacketInfo is the minimal input for flow feature extraction.
// Caller builds this from PCAP (or other source) and passes packets per window.
type PacketInfo struct {
	Timestamp   time.Time
	Direction   Direction
	HeaderLen   int      // TCP or UDP header length in bytes
	PayloadSize int      // TCP or UDP payload size in bytes
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

// Key returns the flow key for this packet (for grouping).
func (p PacketInfo) Key() FlowKey {
	return FlowKey{
		SrcIP:    p.SrcIP,
		DstIP:    p.DstIP,
		SrcPort:  p.SrcPort,
		DstPort:  p.DstPort,
		Protocol: p.Protocol,
	}
}

// Stats holds min, max, mean, std for reuse across feature groups.
type Stats struct {
	Min  float64
	Max  float64
	Mean float64
	Std  float64
}

// FlowFeatures holds CICFlowMeter-style features for one flow.
// Submodules fill their subset; unused fields remain zero.
type FlowFeatures struct {
	// Basic (basic.go)
	FlowDurationUs   int64
	FlowBytesPerSec  float64
	FlowPacketsPerSec float64

	// Counts (counts.go)
	TotalFwdPackets int
	TotalBwdPackets int
	TotalFwdBytes   int64
	TotalBwdBytes   int64

	// Packet length (packetlen.go)
	FwdPacketLen  Stats
	BwdPacketLen  Stats
	PacketLen     Stats
	MinPacketLen  int
	MaxPacketLen  int
	PacketLenMean float64
	PacketLenStd  float64
	PacketLenVar  float64
	AvgPacketSize float64
	AvgFwdSegmentSize float64
	AvgBwdSegmentSize float64

	// IAT (iat.go)
	FlowIAT     Stats
	FwdIAT      Stats
	BwdIAT      Stats
	FwdIATTotal time.Duration
	BwdIATTotal time.Duration

	// Flags (flags.go)
	FwdPSHFlag   int
	BwdPSHFlag   int
	FwdURGFlag   int
	BwdURGFlag   int
	FwdHeaderLen int64
	BwdHeaderLen int64
	FIN          int
	SYN          int
	RST          int
	PSH          int
	ACK          int
	URG          int
	CWR          int
	ECE          int

	// Rates (rates.go) and bulk/subflow/initwin/ratio/activeidle
	FwdPacketsPerSec  float64
	BwdPacketsPerSec  float64
	DownUpRatio       float64
	FwdAvgBytesPerBulk   float64
	FwdAvgPacketsPerBulk float64
	FwdAvgBulkRate       float64
	BwdAvgBytesPerBulk   float64
	BwdAvgPacketsPerBulk float64
	BwdAvgBulkRate       float64
	SubflowFwdPackets    float64
	SubflowFwdBytes      float64
	SubflowBwdPackets    float64
	SubflowBwdBytes      float64
	InitWinBytesFwd      int64
	InitWinBytesBwd      int64
	ActDataPktFwd        int
	MinSegSizeFwd        int
	ActiveTime           Stats
	IdleTime             Stats
}
