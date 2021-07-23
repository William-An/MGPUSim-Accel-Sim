package emu

import (
	"fmt"
	"log"

	"gitlab.com/akita/mgpusim/v2/insts"
	"gitlab.com/akita/mgpusim/v2/kernels"
	"gitlab.com/akita/util/v2/ca"
)

// A Wavefront in the emu package is a wrapper for the kernels.Wavefront
type Wavefront struct {
	*kernels.Wavefront

	pid ca.PID

	Completed  bool
	AtBarrier  bool
	inst       *insts.Inst
	scratchpad Scratchpad

	PC       uint64
	Exec     uint64
	SCC      byte
	VCC      uint64
	M0       uint32
	SRegFile []byte
	VRegFile []byte
	LDS      []byte

	// Accel-Sim
	// naive way to get prev reg values for self-assigned load insts
	PrevSRegFile []byte
	PrevVRegFile []byte
}

// NewWavefront returns the Wavefront that wraps the nativeWf
func NewWavefront(nativeWf *kernels.Wavefront) *Wavefront {
	wf := new(Wavefront)
	wf.Wavefront = nativeWf

	wf.SRegFile = make([]byte, 4*102)
	wf.VRegFile = make([]byte, 4*64*256)
	wf.scratchpad = make([]byte, 4096)

	wf.PrevSRegFile = make([]byte, 4*102)
	wf.PrevVRegFile = make([]byte, 4*64*256)
	return wf
}

// Inst returns the instruction that the wavefront is executing
func (wf *Wavefront) Inst() *insts.Inst {
	return wf.inst
}

// Scratchpad returns the scratchpad that is associated with the wavefront
func (wf *Wavefront) Scratchpad() Scratchpad {
	return wf.scratchpad
}

// PID returns pid
func (wf *Wavefront) PID() ca.PID {
	return wf.pid
}

// Create snapshot of wavefront of scalar regs
func (wf *Wavefront) SRegSaveSnapShot() {
	tmp := make([]byte, 4*102)
	copy(tmp, wf.SRegFile)
	wf.PrevSRegFile = tmp
}

// Create snapshot of wavefront of vector regs
func (wf *Wavefront) VRegSaveSnapShot() {
	tmp := make([]byte, 4*64*256)
	copy(tmp, wf.VRegFile)
	wf.PrevVRegFile = tmp
}

// SRegValue returns s(i)'s value
func (wf *Wavefront) SRegValue(i int) uint32 {
	return insts.BytesToUint32(wf.SRegFile[i*4 : i*4+4])
}

// VRegValue returns the value of v(i) of a certain lain
func (wf *Wavefront) VRegValue(lane int, i int) uint32 {
	offset := lane*1024 + i*4
	return insts.BytesToUint32(wf.VRegFile[offset : offset+4])
}

// ReadReg returns the raw register value
//nolint:gocyclo
func (wf *Wavefront) _ReadReg(reg *insts.Reg, regCount int, laneID int, sRegFile []byte, vRegFile []byte) []byte {
	numBytes := reg.ByteSize
	if regCount >= 2 {
		numBytes *= regCount
	}

	// There are some concerns in terms of reading VCC and EXEC (64 or 32? And how to decide?)
	var value = make([]byte, numBytes)
	if reg.IsSReg() {
		offset := reg.RegIndex() * 4
		copy(value, sRegFile[offset:offset+numBytes])
	} else if reg.IsVReg() {
		offset := laneID*256*4 + reg.RegIndex()*4
		copy(value, vRegFile[offset:offset+numBytes])
	} else if reg.RegType == insts.SCC {
		value[0] = wf.SCC
	} else if reg.RegType == insts.VCC {
		copy(value, insts.Uint64ToBytes(wf.VCC))
	} else if reg.RegType == insts.VCCLO && regCount == 1 {
		copy(value, insts.Uint32ToBytes(uint32(wf.VCC)))
	} else if reg.RegType == insts.VCCHI && regCount == 1 {
		copy(value, insts.Uint32ToBytes(uint32(wf.VCC>>32)))
	} else if reg.RegType == insts.VCCLO && regCount == 2 {
		copy(value, insts.Uint64ToBytes(wf.VCC))
	} else if reg.RegType == insts.EXEC {
		copy(value, insts.Uint64ToBytes(wf.Exec))
	} else if reg.RegType == insts.EXECLO && regCount == 2 {
		copy(value, insts.Uint64ToBytes(wf.Exec))
	} else if reg.RegType == insts.M0 {
		copy(value, insts.Uint32ToBytes(wf.M0))
	} else {
		log.Panicf("Register type %s not supported", reg.Name)
	}

	return value
}

func (wf *Wavefront) ReadReg(reg *insts.Reg, regCount int, laneID int) []byte {
	return wf._ReadReg(reg, regCount, laneID, wf.SRegFile, wf.VRegFile)
}

func (wf *Wavefront) ReadPrevReg(reg *insts.Reg, regCount int, laneID int) []byte {
	return wf._ReadReg(reg, regCount, laneID, wf.PrevSRegFile, wf.PrevVRegFile)
}

// WriteReg returns the raw register value
//nolint:gocyclo
func (wf *Wavefront) WriteReg(
	reg *insts.Reg,
	regCount int,
	laneID int,
	data []byte,
) {
	numBytes := reg.ByteSize
	if regCount >= 2 {
		numBytes *= regCount
	}

	if reg.IsSReg() {
		offset := reg.RegIndex() * 4
		copy(wf.SRegFile[offset:offset+numBytes], data)
	} else if reg.IsVReg() {
		offset := laneID*256*4 + reg.RegIndex()*4
		copy(wf.VRegFile[offset:offset+numBytes], data)
	} else if reg.RegType == insts.SCC {
		wf.SCC = data[0]
	} else if reg.RegType == insts.VCC {
		wf.VCC = insts.BytesToUint64(data)
	} else if reg.RegType == insts.VCCLO && regCount == 2 {
		wf.VCC = insts.BytesToUint64(data)
	} else if reg.RegType == insts.VCCLO && regCount == 1 {
		wf.VCC &= uint64(0x00000000ffffffff)
		wf.VCC |= uint64(insts.BytesToUint32(data))
	} else if reg.RegType == insts.VCCHI && regCount == 1 {
		wf.VCC &= uint64(0xffffffff00000000)
		wf.VCC |= uint64(insts.BytesToUint32(data)) << 32
	} else if reg.RegType == insts.EXEC {
		wf.Exec = insts.BytesToUint64(data)
	} else if reg.RegType == insts.EXECLO && regCount == 2 {
		wf.Exec = insts.BytesToUint64(data)
	} else if reg.RegType == insts.M0 {
		wf.M0 = insts.BytesToUint32(data)
	} else {
		log.Panicf("Register type %s not supported", reg.Name)
	}
}

// Take the work items addresses and convert to compressed address string
func (wf *Wavefront) compressedMemoryAddr() string {
	if !wf.inst.IsMemInst() {
		// 0 mem width
		return "0"
	}

	// Check if Dst regs and addr regs of inst overlap in a load op
	// if overlap, means that some regs of addr might get overwritten, thus read from prev regfile
	isRegOverlap := false
	if wf.inst.IsLoadInst() && wf.inst.Addr != nil &&
		wf.inst.Dst != nil &&
		wf.inst.Addr.OperandType == insts.RegOperand &&
		wf.inst.Dst.OperandType == insts.RegOperand {
		dstStart := wf.inst.Dst.Register.RegIndex()
		dstEnd := dstStart + wf.inst.Dst.RegCount
		addrStart := wf.inst.Addr.Register.RegIndex()
		addrEnd := addrStart + wf.inst.Dst.RegCount

		isRegOverlap = (dstStart <= addrStart && dstEnd >= addrStart) ||
			(addrStart <= dstStart && addrEnd >= dstStart)
	}
	var workItemAddrs [64]uint64

	// Get address from vregs
	for laneID := 0; laneID < 64; laneID++ {
		if wf.inst.FormatType == insts.SMEM {
			// SMEM ops
			// Treat SMEM op differently
			regIdx := wf.inst.Base.Register.RegIndex()
			regCount := wf.inst.Base.RegCount
			var offset uint64 = uint64(wf.inst.Offset.IntValue)
			if !wf.inst.Imm { // Read offset reg val
				offset = insts.BytesToUint64(wf.ReadReg(insts.SReg(wf.inst.Offset.Register.RegIndex()), 1, laneID))
			}

			if isRegOverlap {
				workItemAddrs[laneID] = insts.BytesToUint64(
					wf.ReadPrevReg(insts.SReg(regIdx), regCount, laneID)) + offset
			} else {
				workItemAddrs[laneID] = insts.BytesToUint64(
					wf.ReadReg(insts.SReg(regIdx), regCount, laneID)) + offset
			}

		} else if wf.inst.FormatType == insts.DS {
			// DS Ops, data share ops on local memory
			// TODO View as consecutive memory address despite of different offest values
			regIdx := wf.inst.Addr.Register.RegIndex()

			if isRegOverlap {
				workItemAddrs[laneID] = uint64(insts.BytesToUint32(
					wf.ReadPrevReg(insts.VReg(regIdx), 1, laneID)))
			} else {
				workItemAddrs[laneID] = uint64(insts.BytesToUint32(
					wf.ReadReg(insts.VReg(regIdx), 1, laneID)))
			}

		} else if wf.inst.Addr.OperandType != insts.RegOperand {
			// Literal address
			workItemAddrs[laneID] = uint64(wf.inst.Addr.IntValue)
		} else if wf.inst.Addr.RegCount > 1 {
			// 64 bit address
			regIdx := wf.inst.Addr.Register.RegIndex()
			regCount := wf.inst.Addr.RegCount
			if isRegOverlap {
				workItemAddrs[laneID] = insts.BytesToUint64(
					wf.ReadPrevReg(insts.VReg(regIdx), regCount, laneID))
			} else {
				workItemAddrs[laneID] = insts.BytesToUint64(
					wf.ReadReg(insts.VReg(regIdx), regCount, laneID))
			}

		} else if wf.inst.Addr.RegCount == 1 {
			// 32 bit address
			regIdx := wf.inst.Addr.Register.RegIndex()

			if isRegOverlap {
				workItemAddrs[laneID] = uint64(insts.BytesToUint32(
					wf.ReadPrevReg(insts.VReg(regIdx), 1, laneID)))
			} else {
				workItemAddrs[laneID] = uint64(insts.BytesToUint32(
					wf.ReadReg(insts.VReg(regIdx), 1, laneID)))
			}

		}

	}

	// Snapshot the regfile
	wf.SRegSaveSnapShot()
	wf.VRegSaveSnapShot()

	// Memwidth info is encoded in the disasm decode table
	memStr := fmt.Sprintf("%d", wf.inst.MemoryWidth)

	// Need array of reg values and exec mask
	mask := wf.Exec
	var base_stride_success bool
	var base_addr uint64
	var stride int
	var deltas []int64

	base_stride_success, base_addr, stride = base_stride_compress(&workItemAddrs, mask)

	if base_stride_success {
		return fmt.Sprintf("%s 1 0x%x %d", memStr, base_addr, stride)
	} else {
		base_addr, deltas = base_delta_compress(&workItemAddrs, mask)
		baseStr := fmt.Sprintf("2 0x%x", base_addr)

		for i := 0; i < len(deltas); i++ {
			baseStr += fmt.Sprintf(" %d", deltas[i])
		}
		return fmt.Sprintf("%s %s", memStr, baseStr)
	}
}

func base_stride_compress(workItemAddrs *[64]uint64, mask uint64) (const_stride bool, base_addr uint64, stride int) {
	const_stride = true
	var first_bit1_found bool = false
	var last_bit1_found bool = false

	for s := 0; s < 64; s++ {
		if (((mask >> s) & 1) == 1) && !first_bit1_found { // Find first bit that is 1
			first_bit1_found = true
			base_addr = workItemAddrs[s]            // Load base address into it
			if s < 31 && (((mask >> s) & 1) == 1) { // Attempt to find an initial constant stride?
				stride = int(workItemAddrs[s+1] - workItemAddrs[s])

			} else { // If no constant stride found, exit loop
				const_stride = false
				break
			}
		} else if first_bit1_found && !last_bit1_found {
			if ((mask >> s) & 1) == 1 {
				if stride != int(workItemAddrs[s]-workItemAddrs[s-1]) {
					const_stride = false
					break
				}
			} else {
				last_bit1_found = true
			}
		} else if last_bit1_found {
			if ((mask >> s) & 1) == 1 {
				const_stride = false
				break
			}
		}
	}

	return const_stride, base_addr, stride
}

func base_delta_compress(workItemAddrs *[64]uint64, mask uint64) (base_addr uint64, deltas []int64) {
	var first_bit1_found bool = false
	var last_address uint64 = 0
	for s := 0; s < 64; s++ {
		if (((mask >> s) & 1) == 1) && !first_bit1_found {
			base_addr = workItemAddrs[s]
			first_bit1_found = true
			last_address = workItemAddrs[s]
		} else if (((mask >> s) & 1) == 1) && first_bit1_found {
			deltas = append(deltas, int64(workItemAddrs[s]-last_address))
			last_address = workItemAddrs[s]
		}
	}

	return base_addr, deltas
}
