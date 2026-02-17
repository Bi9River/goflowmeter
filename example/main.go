// This example builds a slice of PacketInfo (one time window) with two flows,
// runs the flowmeter, and prints flow features in CICFlowMeter output order and naming.
// Feature order and names match CICFlowMeter CSV columns 8–84 (FlowFeature enum).
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Bi9River/goflowmeter"
)

func main() {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Two flows in one window: interleaved packets.
	// Flow 1: 1.1.1.1:12345 -> 2.2.2.2:80 (TCP)
	// Flow 2: 3.3.3.3:22222 -> 4.4.4.4:443 (TCP)
	packets := []flowmeter.PacketInfo{
		{Timestamp: base, HeaderLen: 40, PayloadSize: 60, Direction: flowmeter.Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6, SYN: true},
		{Timestamp: base.Add(100 * time.Millisecond), HeaderLen: 40, PayloadSize: 40, Direction: flowmeter.Forward, SrcIP: "3.3.3.3", DstIP: "4.4.4.4", SrcPort: 22222, DstPort: 443, Protocol: 6},
		{Timestamp: base.Add(time.Second), HeaderLen: 40, PayloadSize: 160, Direction: flowmeter.Forward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6, PSH: true, ACK: true},
		{Timestamp: base.Add(1100 * time.Millisecond), HeaderLen: 40, PayloadSize: 460, Direction: flowmeter.Backward, SrcIP: "3.3.3.3", DstIP: "4.4.4.4", SrcPort: 22222, DstPort: 443, Protocol: 6, ACK: true},
		{Timestamp: base.Add(2 * time.Second), HeaderLen: 40, PayloadSize: 110, Direction: flowmeter.Backward, SrcIP: "1.1.1.1", DstIP: "2.2.2.2", SrcPort: 12345, DstPort: 80, Protocol: 6, FIN: true, ACK: true},
	}

	pairs := flowmeter.ProcessPacketsWithKeys(packets)

	fmt.Printf("Input: %d packets -> Output: %d flow(s)\n\n", len(packets), len(pairs))

	// featureTable returns (names, values) in CICFlowMeter CSV column order (8–84)
	names, vals := featureTableCICOrder(pairs)
	colWidth := 14
	nameWidth := 28
	fmt.Printf("%-*s", nameWidth, "Feature")
	for i := range pairs {
		fmt.Printf(" %*s", colWidth, formatFlowID(pairs[i].Key))
	}
	fmt.Println()
	for i := 0; i < nameWidth+len(pairs)*colWidth+len(pairs); i++ {
		fmt.Print("-")
	}
	fmt.Println()
	for i := range names {
		fmt.Printf("%-*s", nameWidth, names[i])
		for _, v := range vals[i] {
			fmt.Printf(" %*v", colWidth, v)
		}
		fmt.Println()
	}

	// CSV-style: CICFlowMeter header (Flow ID + columns 8–84) then one line per flow
	fmt.Println()
	fmt.Println("CSV-style (CICFlowMeter header + Flow ID + columns 8–84):")
	header := "Flow ID," + strings.Join(names, ",")
	fmt.Println(header)
	for idx, p := range pairs {
		row := formatCSVRow(p, names, vals, idx)
		fmt.Println(row)
	}
}

// formatFlowID returns CIC-style flow ID: SrcIP-DstIP-SrcPort-DstPort-Protocol
func formatFlowID(k flowmeter.FlowKey) string {
	return fmt.Sprintf("%s-%s-%d-%d-%d", k.SrcIP, k.DstIP, k.SrcPort, k.DstPort, k.Protocol)
}

func formatCSVRow(p flowmeter.FlowWithKey, names []string, vals [][]interface{}, flowIdx int) string {
	row := formatFlowID(p.Key)
	for i := range names {
		if flowIdx < len(vals[i]) {
			row += fmt.Sprintf(",%v", vals[i][flowIdx])
		}
	}
	return row
}

// featureTableCICOrder returns feature names and values in CICFlowMeter CSV order (columns 8–84).
// Stat order: Fwd/Bwd packet length = Max, Min, Mean, Std; Flow/Fwd/Bwd IAT and Active/Idle = Mean, Std, Max, Min.
func featureTableCICOrder(pairs []flowmeter.FlowWithKey) (names []string, values [][]interface{}) {
	if len(pairs) == 0 {
		return nil, nil
	}
	nFlow := len(pairs)
	add := func(name string, getVal func(flowmeter.FlowFeatures) interface{}) {
		names = append(names, name)
		row := make([]interface{}, nFlow)
		for j := range pairs {
			row[j] = getVal(pairs[j].Features)
		}
		values = append(values, row)
	}

	// 8–12
	add("Flow Duration", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowDurationUs })
	add("Total Fwd Packet", func(fl flowmeter.FlowFeatures) interface{} { return fl.TotalFwdPackets })
	add("Total Bwd packets", func(fl flowmeter.FlowFeatures) interface{} { return fl.TotalBwdPackets })
	add("Total Length of Fwd Packet", func(fl flowmeter.FlowFeatures) interface{} { return fl.TotalFwdBytes })
	add("Total Length of Bwd Packet", func(fl flowmeter.FlowFeatures) interface{} { return fl.TotalBwdBytes })
	// 13–16 Fwd Packet Length: Max, Min, Mean, Std
	add("Fwd Packet Length Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPacketLen.Max })
	add("Fwd Packet Length Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPacketLen.Min })
	add("Fwd Packet Length Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPacketLen.Mean })
	add("Fwd Packet Length Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPacketLen.Std })
	// 17–20 Bwd Packet Length: Max, Min, Mean, Std
	add("Bwd Packet Length Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPacketLen.Max })
	add("Bwd Packet Length Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPacketLen.Min })
	add("Bwd Packet Length Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPacketLen.Mean })
	add("Bwd Packet Length Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPacketLen.Std })
	// 21–22
	add("Flow Bytes/s", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowBytesPerSec })
	add("Flow Packets/s", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowPacketsPerSec })
	// 23–26 Flow IAT: Mean, Std, Max, Min
	add("Flow IAT Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowIAT.Mean })
	add("Flow IAT Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowIAT.Std })
	add("Flow IAT Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowIAT.Max })
	add("Flow IAT Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.FlowIAT.Min })
	// 27–31 Fwd IAT Total (µs), Mean, Std, Max, Min
	add("Fwd IAT Total", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdIATTotal.Microseconds() })
	add("Fwd IAT Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdIAT.Mean })
	add("Fwd IAT Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdIAT.Std })
	add("Fwd IAT Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdIAT.Max })
	add("Fwd IAT Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdIAT.Min })
	// 32–36 Bwd IAT Total (µs), Mean, Std, Max, Min
	add("Bwd IAT Total", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdIATTotal.Microseconds() })
	add("Bwd IAT Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdIAT.Mean })
	add("Bwd IAT Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdIAT.Std })
	add("Bwd IAT Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdIAT.Max })
	add("Bwd IAT Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdIAT.Min })
	// 37–42
	add("Fwd PSH Flags", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPSHFlag })
	add("Bwd PSH Flags", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPSHFlag })
	add("Fwd URG Flags", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdURGFlag })
	add("Bwd URG Flags", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdURGFlag })
	add("Fwd Header Length", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdHeaderLen })
	add("Bwd Header Length", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdHeaderLen })
	// 43–44
	add("Fwd Packets/s", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdPacketsPerSec })
	add("Bwd Packets/s", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdPacketsPerSec })
	// 45–49 Packet length (all packets): Min, Max, Mean, Std, Variance
	add("Packet Length Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.MinPacketLen })
	add("Packet Length Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.MaxPacketLen })
	add("Packet Length Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.PacketLenMean })
	add("Packet Length Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.PacketLenStd })
	add("Packet Length Variance", func(fl flowmeter.FlowFeatures) interface{} { return fl.PacketLenVar })
	// 50–57
	add("FIN Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.FIN })
	add("SYN Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.SYN })
	add("RST Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.RST })
	add("PSH Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.PSH })
	add("ACK Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.ACK })
	add("URG Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.URG })
	add("CWR Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.CWR })
	add("ECE Flag Count", func(fl flowmeter.FlowFeatures) interface{} { return fl.ECE })
	// 58–61 (62 is duplicate in CIC, skipped)
	add("Down/Up Ratio", func(fl flowmeter.FlowFeatures) interface{} { return fl.DownUpRatio })
	add("Average Packet Size", func(fl flowmeter.FlowFeatures) interface{} { return fl.AvgPacketSize })
	add("Fwd Segment Size Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.AvgFwdSegmentSize })
	add("Bwd Segment Size Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.AvgBwdSegmentSize })
	// 63–68
	add("Fwd Bytes/Bulk Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdAvgBytesPerBulk })
	add("Fwd Packet/Bulk Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdAvgPacketsPerBulk })
	add("Fwd Bulk Rate Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.FwdAvgBulkRate })
	add("Bwd Bytes/Bulk Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdAvgBytesPerBulk })
	add("Bwd Packet/Bulk Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdAvgPacketsPerBulk })
	add("Bwd Bulk Rate Avg", func(fl flowmeter.FlowFeatures) interface{} { return fl.BwdAvgBulkRate })
	// 69–72
	add("Subflow Fwd Packets", func(fl flowmeter.FlowFeatures) interface{} { return fl.SubflowFwdPackets })
	add("Subflow Fwd Bytes", func(fl flowmeter.FlowFeatures) interface{} { return fl.SubflowFwdBytes })
	add("Subflow Bwd Packets", func(fl flowmeter.FlowFeatures) interface{} { return fl.SubflowBwdPackets })
	add("Subflow Bwd Bytes", func(fl flowmeter.FlowFeatures) interface{} { return fl.SubflowBwdBytes })
	// 73–76
	add("FWD Init Win Bytes", func(fl flowmeter.FlowFeatures) interface{} { return fl.InitWinBytesFwd })
	add("Bwd Init Win Bytes", func(fl flowmeter.FlowFeatures) interface{} { return fl.InitWinBytesBwd })
	add("Fwd Act Data Pkts", func(fl flowmeter.FlowFeatures) interface{} { return fl.ActDataPktFwd })
	add("Fwd Seg Size Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.MinSegSizeFwd })
	// 77–80 Active: Mean, Std, Max, Min
	add("Active Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.ActiveTime.Mean })
	add("Active Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.ActiveTime.Std })
	add("Active Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.ActiveTime.Max })
	add("Active Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.ActiveTime.Min })
	// 81–84 Idle: Mean, Std, Max, Min
	add("Idle Mean", func(fl flowmeter.FlowFeatures) interface{} { return fl.IdleTime.Mean })
	add("Idle Std", func(fl flowmeter.FlowFeatures) interface{} { return fl.IdleTime.Std })
	add("Idle Max", func(fl flowmeter.FlowFeatures) interface{} { return fl.IdleTime.Max })
	add("Idle Min", func(fl flowmeter.FlowFeatures) interface{} { return fl.IdleTime.Min })

	return names, values
}
