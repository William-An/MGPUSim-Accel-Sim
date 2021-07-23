package insts

import (
	"debug/elf"
	"fmt"
	"log"
	"strings"
)

// ExeUnit defines which execution unit should execute the instruction
type ExeUnit int

// Defines all possible execution units
const (
	ExeUnitVALU ExeUnit = iota
	ExeUnitScalar
	ExeUnitVMem
	ExeUnitBranch
	ExeUnitLDS
	ExeUnitGDS
	ExeUnitSpecial
)

// A InstType represents an instruction type. For example s_barrier instruction
// is a instruction type
type InstType struct {
	InstName    string
	Opcode      Opcode
	Format      *Format
	ID          int
	ExeUnit     ExeUnit
	DSTWidth    int
	SRC0Width   int
	SRC1Width   int
	SRC2Width   int
	SDSTWidth   int
	MemoryWidth int
}

// An Inst is a GCN3 instruction
type Inst struct {
	*Format
	*InstType
	ByteSize int
	PC       uint64

	Src0 *Operand
	Src1 *Operand
	Src2 *Operand
	Dst  *Operand
	SDst *Operand // For VOP3b

	Addr   *Operand
	Data   *Operand
	Data1  *Operand
	Base   *Operand
	Offset *Operand
	SImm16 *Operand

	Abs                 int
	Omod                int
	Neg                 int
	Offset0             uint32
	Offset1             uint32
	SystemLevelCoherent bool
	GlobalLevelCoherent bool
	TextureFailEnable   bool
	Imm                 bool
	Clamp               bool
	GDS                 bool
	VMCNT               int
	LKGMCNT             int

	//Fields for SDWA extensions
	IsSdwa    bool
	DstSel    SDWASelect
	DstUnused SDWAUnused
	Src0Sel   SDWASelect
	Src0Sext  bool
	Src0Neg   bool
	Src0Abs   bool
	Src1Sel   SDWASelect
	Src1Sext  bool
	Src1Neg   bool
	Src1Abs   bool
	Src2Neg   bool
	Src2Abs   bool
}

// NewInst creates a zero-filled instruction
func NewInst() *Inst {
	i := new(Inst)
	i.Format = new(Format)
	i.InstType = new(InstType)
	return i
}

// TODO Might want to use function to calculate the reg count and reg string
func (i Inst) sop2String() string {
	srcString := ""
	regCount := 0

	// Make sure the right reg count is added
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}
	dstCount := 1
	if i.Dst.RegCount != 0 {
		dstCount = i.Dst.RegCount
	}
	instString := fmt.Sprintf("%d %s %s %d %s",
		dstCount, i.Dst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) vop1String() string {
	srcString := ""
	regCount := 0
	regDstCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}

	// VOP1 Inst might use special indexing? like V_CVT_F64_I32 will use two vreg instead of one
	if i.Dst.OperandType == RegOperand {
		count := 1
		if i.Dst.RegCount != 0 {
			count = i.Dst.RegCount
		}
		regDstCount += count
	}
	// VOP1 might assume to have 1 reg, thus the i.Dst.RegCount might be zero
	instString := fmt.Sprintf("%d %s %s %d %s",
		regDstCount, i.Dst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) flatString() string {
	/*
		Flat Memory instructions read, or write, one piece of data into, or out of, VGPRs;
		they do this separately for each work-item in a wavefront.
	*/

	// var s string
	// if i.Opcode >= 16 && i.Opcode <= 23 {
	// 	s = i.InstName + " " + i.Dst.String() + ", " +
	// 		i.Addr.String()
	// } else if i.Opcode >= 24 && i.Opcode <= 31 {
	// 	s = i.InstName + " " + i.Addr.String() + ", " +
	// 		i.Data.String()
	// }

	var instString string
	if i.Opcode >= 16 && i.Opcode <= 23 { // Load
		regDstCount := 0
		if i.Dst.OperandType == RegOperand {
			count := 1
			if i.Dst.RegCount != 0 {
				count = i.Dst.RegCount
			}
			regDstCount += count
		}
		instString = fmt.Sprintf("%d %s %s %d %s",
			regDstCount, i.Dst.String(), i.InstName, i.Addr.RegCount, i.Addr.String())
	} else if i.Opcode >= 24 && i.Opcode <= 31 { // Store
		// How to deal with store reg? Treat all as src regs for now
		regAddrCount := 0
		regDataCount := 0
		if i.Addr.OperandType == RegOperand {
			count := 1
			if i.Addr.RegCount != 0 {
				count = i.Addr.RegCount
			}
			regAddrCount += count
		}
		if i.Data.OperandType == RegOperand {
			count := 1
			if i.Data.RegCount != 0 {
				count = i.Data.RegCount
			}
			regDataCount += count
		}
		instString = fmt.Sprintf("0 %s %d %s %s",
			i.InstName, regAddrCount+regDataCount, i.Addr.String(), i.Data.String())
	}

	return instString
}

func (i Inst) smemString() string {
	// TODO: Consider store instructions, and the case if imm = 0
	//
	// s := fmt.Sprintf("%s %s, %s, %#x",
	// 	i.InstName, i.Data.String(), i.Base.String(), uint16(i.Offset.IntValue))
	// return s

	// If load op, have dst regs, else be store op, treat addr and data reg as src
	var instString string
	if strings.Contains(i.InstName, "load") && i.Imm { // load with immediate val offset
		instString = fmt.Sprintf("%d %s %s %d %s",
			i.Data.RegCount, i.Data.String(), i.InstName, i.Base.RegCount, i.Base.String())
	} else if strings.Contains(i.InstName, "load") && !i.Imm { // load with SREG offset
		instString = fmt.Sprintf("%d %s %s %d %s %s",
			i.Data.RegCount, i.Data.String(), i.InstName, i.Base.RegCount+1, i.Base.String(), i.Offset.regOperandToString())
	} else if !i.Imm { // Store with SREG offset
		instString = fmt.Sprintf("0 %s %d %s %s %s",
			i.InstName, i.Base.RegCount+i.Data.RegCount+1, i.Base.String(), i.Data.String(), i.Offset.String())
	} else { // Store with immediate value offset and other smem ops
		instString = fmt.Sprintf("0 %s %d %s %s",
			i.InstName, i.Base.RegCount+i.Data.RegCount, i.Base.String(), i.Data.String())

	}

	return instString
}

func (i Inst) soppString(file *elf.File) string {
	// operandStr := ""
	// if i.Opcode == 12 { // S_WAITCNT
	// 	operandStr = i.waitcntOperandString()
	// } else if i.Opcode >= 2 && i.Opcode <= 9 { // Branch
	// 	symbolFound := false
	// 	if file != nil {
	// 		imm := int16(uint16(i.SImm16.IntValue))
	// 		target := i.PC + uint64(imm*4) + 4
	// 		symbols, _ := file.Symbols()
	// 		for _, symbol := range symbols {
	// 			if symbol.Value == target {
	// 				operandStr = " " + symbol.Name
	// 				symbolFound = true
	// 			}
	// 		}
	// 	}
	// 	if !symbolFound {
	// 		operandStr = " " + i.SImm16.String()
	// 	}
	// } else if i.Opcode == 1 || i.Opcode == 10 {
	// 	// Does not print anything
	// } else {
	// 	operandStr = " " + i.SImm16.String()
	// }

	// TODO How to handle this?
	instString := fmt.Sprintf("0 %s 0",
		i.InstName)
	return instString
}

func (i Inst) waitcntOperandString() string {
	operandStr := ""
	if i.VMCNT != 15 {
		operandStr += fmt.Sprintf(" vmcnt(%d)", i.VMCNT)
	}

	if i.LKGMCNT != 15 {
		operandStr += fmt.Sprintf(" lgkmcnt(%d)", i.LKGMCNT)
	}
	return operandStr
}

func (i Inst) vop2String() string {
	// VOP2 is for instructions with two inputs and a single vector destination.
	// Instructions that have a carry-out implicitly write the carry-out to the VCC register.

	// Original printout
	// s := fmt.Sprintf("%s %s", i.InstName, i.Dst.String())

	// switch i.Opcode {
	// case 25, 26, 27, 28, 29, 30:
	// 	s += ", vcc"
	// }

	// s += fmt.Sprintf(", %s, %s", i.Src0.String(), i.Src1.String())

	// switch i.Opcode {
	// case 0, 28, 29:
	// 	s += ", vcc"
	// case 24, 37: // madak
	// 	s += ", " + i.Src2.String()
	// }

	// if i.IsSdwa {
	// 	s = strings.ReplaceAll(s, "_e32", "_sdwa")
	// 	s += i.sdwaVOP2String()
	// }

	// return s

	// Accel-Sim printout
	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}

	// VOP2 inst assume dst is 1 reg
	instString := fmt.Sprintf("1 %s %s %d %s",
		i.Dst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) sdwaVOP2String() string {
	s := ""

	s += " dst_sel:"
	s += sdwaSelectString(i.DstSel)
	s += " dst_unused:"
	s += sdwaUnusedString(i.DstUnused)
	s += " src0_sel:"
	s += sdwaSelectString(i.Src0Sel)
	s += " src1_sel:"
	s += sdwaSelectString(i.Src1Sel)

	return s
}

func (i Inst) vopcString() string {
	// Write to either VCC or EXEC, ignore currently in Accel-Sim
	// TODO Could assign special registers numbering in Accel-Sim? for writing to these special regs
	// dst := "vcc"
	// if strings.Contains(i.InstName, "cmpx") {
	// 	dst = "exec"
	// }

	// return fmt.Sprintf("%s %s, %s, %s",
	// 	i.InstName, dst, i.Src0.String(), i.Src1.String())

	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}

	instString := fmt.Sprintf("0 %s %d %s",
		i.InstName, regCount, srcString)
	return instString
}

func (i Inst) sopcString() string {
	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}
	instString := fmt.Sprintf("0 %s %d %s",
		i.InstName, regCount, srcString)
	return instString
}

func (i Inst) vop3aString() string {
	/**
	3 in, 1 out
	VOP3 is for instructions with up to three inputs, input modifiers (negate and
	absolute value), and output modifiers. There are two forms of VOP3: one which
	uses a scalar destination field (used only for div_scale, integer add and subtract);
	this is designated VOP3b. All other instructions use the common form,
	designated VOP3a.
	*/
	// s := fmt.Sprintf("%s %s",
	// 	i.InstName, i.Dst.String())

	// s += ", " + i.vop3aInputOperandString(*i.Src0,
	// 	i.Src0Neg,
	// 	i.Src0Abs)

	// s += ", " + i.vop3aInputOperandString(*i.Src1,
	// 	i.Src1Neg,
	// 	i.Src1Abs)

	// if i.Src2 == nil {
	// 	return s
	// }

	// s += ", " + i.vop3aInputOperandString(*i.Src2,
	// 	i.Src2Neg,
	// 	i.Src2Abs)

	// return s

	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}
	if i.Src2 != nil && i.Src2.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src2.String()
		} else {
			srcString = srcString + " " + i.Src2.String()
		}
		count := 1
		if i.Src2.RegCount != 0 {
			count = i.Src2.RegCount
		}
		regCount += count
	}

	var instString string

	dstCount := 1
	if i.Dst.RegCount != 0 {
		dstCount = i.Dst.RegCount
	}
	instString = fmt.Sprintf("%d %s %s %d %s",
		dstCount, i.Dst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) vop3aInputOperandString(operand Operand, neg, abs bool) string {
	s := ""

	if neg {
		s += "-"
	}

	if abs {
		s += "|"
	}

	s += operand.String()

	if abs {
		s += "|"
	}

	return s
}

func (i Inst) vop3bString() string {
	// 3 in, 2 out

	// s := i.InstName + " "

	// if i.Dst != nil {
	// 	s += i.Dst.String() + ", "
	// }

	// s += fmt.Sprintf("%s, %s, %s",
	// 	i.SDst.String(),
	// 	i.Src0.String(),
	// 	i.Src1.String(),
	// )

	// if i.Opcode != 281 && i.Src2 != nil {
	// 	s += ", " + i.Src2.String()
	// }

	// return s

	// Accel-Sim
	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	if i.Src1.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src1.String()
		} else {
			srcString = srcString + " " + i.Src1.String()
		}
		count := 1
		if i.Src1.RegCount != 0 {
			count = i.Src1.RegCount
		}
		regCount += count
	}
	if i.Src2 != nil && i.Src2.OperandType == RegOperand {
		if regCount == 0 {
			srcString = i.Src2.String()
		} else {
			srcString = srcString + " " + i.Src2.String()
		}
		count := 1
		if i.Src2.RegCount != 0 {
			count = i.Src2.RegCount
		}
		regCount += count
	}

	instString := fmt.Sprintf("%d %s %s %s %d %s",
		i.Dst.RegCount+i.SDst.RegCount, i.Dst.String(), i.SDst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) sop1String() string {
	// TODO Vcc is 32bit with lo and high
	srcString := ""
	regCount := 0
	if i.Src0.OperandType == RegOperand {
		srcString = i.Src0.String()
		count := 1
		if i.Src0.RegCount != 0 {
			count = i.Src0.RegCount
		}
		regCount += count
	}
	dstCount := 1
	if i.Dst.RegCount != 0 {
		dstCount = i.Dst.RegCount
	}
	instString := fmt.Sprintf("%d %s %s %d %s",
		dstCount, i.Dst.String(), i.InstName, regCount, srcString)
	return instString
}

func (i Inst) sopkString() string {
	// No src regs for immediate values
	instString := fmt.Sprintf("1 %s %s 0",
		i.Dst.String(), i.InstName)
	return instString
}

func (i Inst) dsString() string {
	// Data share operations
	// LDS scratchpad
	// Accel-Sim
	srcString := i.Addr.String()
	regCount := i.Addr.RegCount

	var instString string
	switch i.Opcode {
	case 54, 55, 56, 57, 58, 59, 60, 118, 119, 120, 254, 255: // Read ops
		instString = fmt.Sprintf("%d %s %s %d %s",
			i.Dst.RegCount, i.Dst.String(), i.InstName, regCount, srcString)
		break
	default: // Write ops
		instString = fmt.Sprintf("0 %s %d %s",
			i.InstName, regCount, srcString)
	}

	return instString
}

//nolint:gocyclo
// String returns the disassembly of an instruction
func (i Inst) String(file *elf.File) string {
	switch i.FormatType {
	case SOP2:
		return i.sop2String()
	case SMEM:
		return i.smemString()
	case VOP1:
		return i.vop1String()
	case VOP2:
		return i.vop2String()
	case FLAT:
		return i.flatString()
	case SOPP:
		return i.soppString(file)
	case VOPC:
		return i.vopcString()
	case SOPC:
		return i.sopcString()
	case VOP3a:
		return i.vop3aString()
	case VOP3b:
		return i.vop3bString()
	case SOP1:
		return i.sop1String()
	case SOPK:
		return i.sopkString()
	case DS:
		return i.dsString()
	default:
		log.Panic("Unknown instruction format type.")
		return i.InstName
	}
}

func (i Inst) IsMemInst() bool {
	switch i.FormatType {
	case SMEM, FLAT, DS:
		return true
	default:
		return false
	}
}

func (i Inst) IsLoadInst() bool {
	switch i.FormatType {
	case SMEM:
		return strings.Contains(i.InstName, "load")
	case FLAT:
		return i.Opcode >= 16 && i.Opcode <= 23
	case DS:
		switch i.Opcode {
		case 54, 55, 56, 57, 58, 59, 60, 118, 119, 120, 254, 255: // Read ops
			return true
		default:
			return false
		}
	default:
		return false
	}
}
