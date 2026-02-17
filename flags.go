package flowmeter

// computeFlags fills TCP flag counts per direction (PSH, URG), header length
// per direction, and flow-wide counts for FIN, SYN, RST, PSH, ACK, URG, CWR, ECE.
func computeFlags(packets []PacketInfo, f *FlowFeatures) {
	for _, p := range packets {
		if p.Direction == Forward {
			if p.PSH {
				f.FwdPSHFlag++
			}
			if p.URG {
				f.FwdURGFlag++
			}
			f.FwdHeaderLen += int64(p.HeaderLen)
		} else {
			if p.PSH {
				f.BwdPSHFlag++
			}
			if p.URG {
				f.BwdURGFlag++
			}
			f.BwdHeaderLen += int64(p.HeaderLen)
		}
		if p.FIN {
			f.FIN++
		}
		if p.SYN {
			f.SYN++
		}
		if p.RST {
			f.RST++
		}
		if p.PSH {
			f.PSH++
		}
		if p.ACK {
			f.ACK++
		}
		if p.URG {
			f.URG++
		}
		if p.CWR {
			f.CWR++
		}
		if p.ECE {
			f.ECE++
		}
	}
}
