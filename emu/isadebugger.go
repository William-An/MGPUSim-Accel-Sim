package emu

import (
	"encoding/base64"
	"fmt"
	"log"

	"os"

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
	kernelID string
	cuName   string // Compute unit name
}

// NewISADebugger returns a new ISADebugger that keeps instruction log in logger
func NewISADebugger(cuName string) *ISADebugger {
	h := new(ISADebugger)
	h.Logger = nil
	h.isFirstEntry = true

	// Accel-Sim: No need for json like printout
	// h.Logger.Print("[")
	// atexit.Register(func() { h.Logger.Print("\n]") })

	// Accel-Sim
	h.kernelID = ""
	h.cuName = cuName

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

	// Switch logger if kernelID not matched
	if wf.CodeObject.ID != h.kernelID {
		// Create a new logger file entry
		h.kernelID = wf.CodeObject.ID
		kernelTraceFile, err := os.Create(
			fmt.Sprintf("%s-kernel-%s.trace", h.cuName, h.kernelID))
		if err != nil {
			log.Fatal(err.Error())
		}
		h.Logger = log.New(kernelTraceFile, "", 0)
	}

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
	// TODO Switch logger file here if ID not match

	output := ""
	// TODO Format this as the Accel-Sim
	//  Like intermediate format?
	//  TODO How to group? Find the kernel id
	// TODO Log begin and end of TB/WG

	// output += fmt.Sprintf("{")
	output += fmt.Sprintf(`"wg":[%d,%d,%d],"wf":%d,`,
		wf.WG.IDX, wf.WG.IDY, wf.WG.IDZ, wf.FirstWiFlatID)
	output += fmt.Sprintf(`"Kernel ID":"%s",`, wf.CodeObject.ID)
	output += fmt.Sprintf(`"Inst":"%s",`, wf.Inst().String(nil))
	output += fmt.Sprintf(`"PC":%08x,`, wf.PC)
	output += fmt.Sprintf(`"EXEC":%016x,`, wf.Exec)
	output += fmt.Sprintf(`"VCC":%d,`, wf.VCC)
	output += fmt.Sprintf(`"SCC":%d,`, wf.SCC)

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
