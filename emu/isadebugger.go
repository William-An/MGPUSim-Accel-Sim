package emu

import (
	"encoding/base64"
	"fmt"
	"log"

	"os"
	"sync"

	// "github.com/tebeka/atexit"
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mgpusim/v2/insts"
)

// ISADebugger is a hook that hooks to a emulator computeunit for each intruction
type ISADebugger struct {
	sim.LogHookBase

	isFirstEntry bool
	// prevWf *Wavefront

	// Accel-Sim
	kernelID   string
	kernelName string
	cuName     string // Compute unit name
	mutex      *sync.Mutex

	// Saved current workgroup indices
	IDX int
	IDY int
	IDZ int
}

// NewISADebugger returns a new ISADebugger that keeps instruction log in logger
func NewISADebugger(cuName string, mutex *sync.Mutex) *ISADebugger {
	h := new(ISADebugger)
	h.Logger = nil
	h.isFirstEntry = true

	// Accel-Sim
	h.kernelID = ""
	h.cuName = cuName
	h.mutex = mutex

	// Initially invalid indices
	h.IDX, h.IDY, h.IDZ = -1, -1, -1

	return h
}

// Func defines the behavior of the tracer when the tracer is invoked.
func (h *ISADebugger) Func(ctx sim.HookCtx) {
	wf, ok := ctx.Item.(*Wavefront)
	if !ok {
		return
	}

	// For debugging
	// if wf.FirstWiFlatID != 0 {
	// 	return
	// }

	// For each compute unit
	// Switch logger if kernelID not matched
	h.mutex.Lock()
	if wf.CodeObject.ID != h.kernelID {
		// Finish up previous kernel trace
		// if !h.isFirstEntry {
		// 	h.Logger.Println("#END_TB")
		// }

		// Create a new logger file entry
		firstLogger := true // is this logger the first to write a kernel trace? In order to ignore header info
		h.kernelID = wf.CodeObject.ID
		h.kernelName = wf.CodeObject.KernalName

		// Check for log file existence
		kernelTraceFileName := fmt.Sprintf("kernel-%s.trace", h.kernelID)
		kernelFile, _ := os.Stat(kernelTraceFileName)
		if kernelFile == nil {
			// First logger to this kernel
			kernelTraceFile, _ := os.Create(kernelTraceFileName)
			h.Logger = log.New(kernelTraceFile, "", 0)

			// TODO Log basic header info from hsakerneldispatchpacket
			// for header info as all will be concatenating together
			if firstLogger {
				// Used for mataching hardware results
				h.Logger.Printf("-kernel name = %s\n", h.kernelName)
				h.Logger.Printf("-kernel id = %s\n", h.kernelID)

				// Warp size
				h.Logger.Printf("-warp size = 64\n")

				// ISA Type
				h.Logger.Printf("-isa type = GCN3\n")

				// static shmem bytes + dynamic shared mem bytes
				// Use GroupSegmentSize, which is used to initialize LDS storage size in the simulator
				h.Logger.Printf("-shmem = %d\n", wf.Packet.GroupSegmentSize)

				// The number of registers used by each thread of this kernel function.
				// Get from HSACO WFSgprCount (wavefront scalar reg count) and WIVgprCount (work item vector reg count)
				regCount := wf.CodeObject.WFSgprCount + wf.CodeObject.WIVgprCount
				h.Logger.Printf("-nregs = %d\n", regCount)

				// Used to get the opcode mapping for the GPU
				// Set to 100 for now for this AMD GPU
				h.Logger.Printf("-binary version = 100\n")

				// Ignored
				h.Logger.Printf("-cuda stream id = 0\n")

				// TODO Unknown where to find the base addr for these two
				h.Logger.Printf("-shmem base_addr = 0x%x\n", 0)
				h.Logger.Printf("-local mem base_addr = 0x%x\n", 0)

				h.Logger.Printf("-nvbit version = -1\n")
				h.Logger.Printf("-accelsim tracer version = 3\n")

			}

			// Dims
			wgSizeX := uint32(wf.Packet.WorkgroupSizeX)
			wgSizeY := uint32(wf.Packet.WorkgroupSizeY)
			wgSizeZ := uint32(wf.Packet.WorkgroupSizeZ)

			// As grid size in MGPUSim is the total size rather than grid size
			h.Logger.Printf("-grid dim = (%d,%d,%d)\n",
				wf.Packet.GridSizeX/wgSizeX,
				wf.Packet.GridSizeY/wgSizeY,
				wf.Packet.GridSizeZ/wgSizeZ,
			)
			h.Logger.Printf("-block dim = (%d,%d,%d)\n", wgSizeX, wgSizeY, wgSizeZ)

			// Print the trace format
			h.Logger.Printf("#traces format = threadblock_x threadblock_y threadblock_z warpid_tb PC mask dest_num [reg_dests] opcode src_num [reg_srcs] mem_width [adrrescompress?] [mem_addresses]")

		} else {
			// Not first one
			kernelTraceFile, _ := os.Create(fmt.Sprintf("%s-kernel-%s.trace", h.cuName, h.kernelID))
			h.Logger = log.New(kernelTraceFile, "", 0)
		}
	}
	h.mutex.Unlock()

	// No need for this, use the post-processing tool to handle raw trace
	// Check if to start a new wavegroup
	// if wf.WG.IDX != h.IDX ||
	// 	wf.WG.IDY != h.IDY ||
	// 	wf.WG.IDZ != h.IDZ {
	// 	h.IDX, h.IDY, h.IDZ = wf.WG.IDX, wf.WG.IDY, wf.WG.IDZ

	// 	// Don't print at first
	// 	if h.isFirstEntry {
	// 		h.isFirstEntry = false
	// 	} else {
	// 		h.Logger.Println("#END_TB")
	// 	}

	// 	h.Logger.Println()
	// 	h.Logger.Println("#BEGIN_TB")
	// 	h.Logger.Println()
	// 	h.Logger.Printf("thread block = %d,%d,%d\n", h.IDX, h.IDY, h.IDZ)
	// 	h.Logger.Println()
	// }

	h.logWholeWfAccelSim(wf)

	// h.logWholeWf(wf)
	// if h.prevWf == nil || h.prevWf.FirstWiFlatID != wf.FirstWiFlatID {
	// 	h.logWholeWf(wf)
	// } else {
	// 	h.logDiffWf(wf)
	// }

	// h.stubWf(wf)
}

func (h *ISADebugger) logWholeWfAccelSim(wf *Wavefront) {
	// Format is as the following:
	// #traces format = threadblock_x threadblock_y threadblock_z warpid_tb PC mask dest_num [reg_dests] opcode src_num [reg_srcs] mem_width [adrrescompress?] [mem_addresses]
	// warpid_tb should be just wf.wlflatid/wavefront.size
	output := ""
	output += fmt.Sprintf(`%d %d %d %d `, wf.WG.IDX, wf.WG.IDY, wf.WG.IDZ, wf.FirstWiFlatID/64)
	output += fmt.Sprintf(`%08x `, wf.PC)

	// Make all the scalar insts to has Exec mask of 0xFFFF FFFF FFFF FFFF
	if wf.Inst().IsScalarInst() {
		output += "ffffffffffffffff "
	} else {
		output += fmt.Sprintf(`%016x `, wf.Exec)
	}

	output += fmt.Sprintf(`%s %s`, wf.Inst().String(nil), wf.compressedMemoryAddr())

	h.Logger.Print(output)
}

func (h *ISADebugger) logWholeWf(wf *Wavefront) {
	output := ""
	if h.isFirstEntry {
		h.isFirstEntry = false
	} else {
		output += ","
	}

	output += fmt.Sprintf("{")
	output += fmt.Sprintf(`"wg":[%d,%d,%d],"wf":%d,`,
		wf.WG.IDX, wf.WG.IDY, wf.WG.IDZ, wf.FirstWiFlatID)
	output += fmt.Sprintf(`"Inst":"%s",`, wf.Inst().String(nil))
	output += fmt.Sprintf(`"PCLo":%d,`, wf.PC&0xffffffff)
	output += fmt.Sprintf(`"PCHi":%d,`, wf.PC>>32)
	output += fmt.Sprintf(`"EXECLo":%d,`, wf.Exec&0xffffffff)
	output += fmt.Sprintf(`"EXECHi":%d,`, wf.Exec>>32)
	output += fmt.Sprintf(`"VCCLo":%d,`, wf.VCC&0xffffffff)
	output += fmt.Sprintf(`"VCCHi":%d,`, wf.VCC>>32)
	output += fmt.Sprintf(`"SCC":%d,`, wf.SCC)

	output += fmt.Sprintf(`"SGPRs":[`)
	for i := 0; i < int(wf.CodeObject.WFSgprCount); i++ {
		if i > 0 {
			output += ","
		}
		regValue := insts.BytesToUint32(wf.ReadReg(insts.SReg(i), 1, 0))
		output += fmt.Sprintf("%d", regValue)
	}
	output += "]"

	output += `,"VGPRs":[`
	for i := 0; i < int(wf.CodeObject.WIVgprCount); i++ {
		if i > 0 {
			output += ","
		}
		output += "["

		for laneID := 0; laneID < 64; laneID++ {
			if laneID > 0 {
				output += ","
			}

			regValue := insts.BytesToUint32(
				wf.ReadReg(insts.VReg(i), 1, laneID))
			output += fmt.Sprintf("%d", regValue)
		}

		output += "]"
	}
	output += "]"

	output += `,"LDS":`
	output += fmt.Sprintf(`"%s"`, base64.StdEncoding.EncodeToString(wf.LDS))

	output += fmt.Sprintf("}")

	h.Logger.Print(output)
}
