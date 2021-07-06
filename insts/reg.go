package insts

// Reg is the representation of a register
type Reg struct {
	RegType  RegType
	Name     string
	ByteSize int
	IsBool   bool
}

// VReg returns a vector register object given a certain index
func VReg(index int) *Reg {
	return Regs[V0+RegType(index)]
}

// SReg returns a scalar register object given a certain index
func SReg(index int) *Reg {
	return Regs[S0+RegType(index)]
}

// IsVReg checks if a register is a vector register
func (r *Reg) IsVReg() bool {
	return r.RegType >= V0 && r.RegType <= V255
}

// IsSReg checks if a register is a scalar register
func (r *Reg) IsSReg() bool {
	return r.RegType >= S0 && r.RegType <= S101
}

// RegIndex returns the index of the index in the s-series or the v-series.
// If the register is not s or v register, -1 is returned.
func (r *Reg) RegIndex() int {
	if r.IsSReg() {
		return int(r.RegType - S0)
	} else if r.IsVReg() {
		return int(r.RegType - V0)
	}
	return -1
}

// RegType is the register type
type RegType int

// All the registers
const (
	InvalidRegType = iota
	PC
	V0
	V1
	V2
	V3
	V4
	V5
	V6
	V7
	V8
	V9
	V10
	V11
	V12
	V13
	V14
	V15
	V16
	V17
	V18
	V19
	V20
	V21
	V22
	V23
	V24
	V25
	V26
	V27
	V28
	V29
	V30
	V31
	V32
	V33
	V34
	V35
	V36
	V37
	V38
	V39
	V40
	V41
	V42
	V43
	V44
	V45
	V46
	V47
	V48
	V49
	V50
	V51
	V52
	V53
	V54
	V55
	V56
	V57
	V58
	V59
	V60
	V61
	V62
	V63
	V64
	V65
	V66
	V67
	V68
	V69
	V70
	V71
	V72
	V73
	V74
	V75
	V76
	V77
	V78
	V79
	V80
	V81
	V82
	V83
	V84
	V85
	V86
	V87
	V88
	V89
	V90
	V91
	V92
	V93
	V94
	V95
	V96
	V97
	V98
	V99
	V100
	V101
	V102
	V103
	V104
	V105
	V106
	V107
	V108
	V109
	V110
	V111
	V112
	V113
	V114
	V115
	V116
	V117
	V118
	V119
	V120
	V121
	V122
	V123
	V124
	V125
	V126
	V127
	V128
	V129
	V130
	V131
	V132
	V133
	V134
	V135
	V136
	V137
	V138
	V139
	V140
	V141
	V142
	V143
	V144
	V145
	V146
	V147
	V148
	V149
	V150
	V151
	V152
	V153
	V154
	V155
	V156
	V157
	V158
	V159
	V160
	V161
	V162
	V163
	V164
	V165
	V166
	V167
	V168
	V169
	V170
	V171
	V172
	V173
	V174
	V175
	V176
	V177
	V178
	V179
	V180
	V181
	V182
	V183
	V184
	V185
	V186
	V187
	V188
	V189
	V190
	V191
	V192
	V193
	V194
	V195
	V196
	V197
	V198
	V199
	V200
	V201
	V202
	V203
	V204
	V205
	V206
	V207
	V208
	V209
	V210
	V211
	V212
	V213
	V214
	V215
	V216
	V217
	V218
	V219
	V220
	V221
	V222
	V223
	V224
	V225
	V226
	V227
	V228
	V229
	V230
	V231
	V232
	V233
	V234
	V235
	V236
	V237
	V238
	V239
	V240
	V241
	V242
	V243
	V244
	V245
	V246
	V247
	V248
	V249
	V250
	V251
	V252
	V253
	V254
	V255
	S0
	S1
	S2
	S3
	S4
	S5
	S6
	S7
	S8
	S9
	S10
	S11
	S12
	S13
	S14
	S15
	S16
	S17
	S18
	S19
	S20
	S21
	S22
	S23
	S24
	S25
	S26
	S27
	S28
	S29
	S30
	S31
	S32
	S33
	S34
	S35
	S36
	S37
	S38
	S39
	S40
	S41
	S42
	S43
	S44
	S45
	S46
	S47
	S48
	S49
	S50
	S51
	S52
	S53
	S54
	S55
	S56
	S57
	S58
	S59
	S60
	S61
	S62
	S63
	S64
	S65
	S66
	S67
	S68
	S69
	S70
	S71
	S72
	S73
	S74
	S75
	S76
	S77
	S78
	S79
	S80
	S81
	S82
	S83
	S84
	S85
	S86
	S87
	S88
	S89
	S90
	S91
	S92
	S93
	S94
	S95
	S96
	S97
	S98
	S99
	S100
	S101
	EXEC
	EXECLO
	EXECHI
	EXECZ
	VCC
	VCCLO
	VCCHI
	VCCZ
	SCC
	FlatSratch
	FlatSratchLo
	FlatSratchHi
	XnackMask
	XnackMaskLo
	XnackMaskHi
	Status
	Mode
	M0
	Trapsts
	Tba
	TbaLo
	TbaHi
	Tma
	TmaLo
	TmaHi
	Timp0
	Timp1
	Timp2
	Timp3
	Timp4
	Timp5
	Timp6
	Timp7
	Timp8
	Timp9
	Timp10
	Timp11
	VMCNT
	EXPCNT
	LGKMCNT
)

// Regs are a list of all registers
// Accel-Sim change reg names to match that of NVBit
var Regs = map[RegType]*Reg{
	InvalidRegType: {InvalidRegType, "invalidregtype", 0, false},
	PC:             {PC, "pc", 8, false},
	V0:             {V0, "R0", 4, false},
	V1:             {V1, "R1", 4, false},
	V2:             {V2, "R2", 4, false},
	V3:             {V3, "R3", 4, false},
	V4:             {V4, "R4", 4, false},
	V5:             {V5, "R5", 4, false},
	V6:             {V6, "R6", 4, false},
	V7:             {V7, "R7", 4, false},
	V8:             {V8, "R8", 4, false},
	V9:             {V9, "R9", 4, false},
	V10:            {V10, "R10", 4, false},
	V11:            {V11, "R11", 4, false},
	V12:            {V12, "R12", 4, false},
	V13:            {V13, "R13", 4, false},
	V14:            {V14, "R14", 4, false},
	V15:            {V15, "R15", 4, false},
	V16:            {V16, "R16", 4, false},
	V17:            {V17, "R17", 4, false},
	V18:            {V18, "R18", 4, false},
	V19:            {V19, "R19", 4, false},
	V20:            {V20, "R20", 4, false},
	V21:            {V21, "R21", 4, false},
	V22:            {V22, "R22", 4, false},
	V23:            {V23, "R23", 4, false},
	V24:            {V24, "R24", 4, false},
	V25:            {V25, "R25", 4, false},
	V26:            {V26, "R26", 4, false},
	V27:            {V27, "R27", 4, false},
	V28:            {V28, "R28", 4, false},
	V29:            {V29, "R29", 4, false},
	V30:            {V30, "R30", 4, false},
	V31:            {V31, "R31", 4, false},
	V32:            {V32, "R32", 4, false},
	V33:            {V33, "R33", 4, false},
	V34:            {V34, "R34", 4, false},
	V35:            {V35, "R35", 4, false},
	V36:            {V36, "R36", 4, false},
	V37:            {V37, "R37", 4, false},
	V38:            {V38, "R38", 4, false},
	V39:            {V39, "R39", 4, false},
	V40:            {V40, "R40", 4, false},
	V41:            {V41, "R41", 4, false},
	V42:            {V42, "R42", 4, false},
	V43:            {V43, "R43", 4, false},
	V44:            {V44, "R44", 4, false},
	V45:            {V45, "R45", 4, false},
	V46:            {V46, "R46", 4, false},
	V47:            {V47, "R47", 4, false},
	V48:            {V48, "R48", 4, false},
	V49:            {V49, "R49", 4, false},
	V50:            {V50, "R50", 4, false},
	V51:            {V51, "R51", 4, false},
	V52:            {V52, "R52", 4, false},
	V53:            {V53, "R53", 4, false},
	V54:            {V54, "R54", 4, false},
	V55:            {V55, "R55", 4, false},
	V56:            {V56, "R56", 4, false},
	V57:            {V57, "R57", 4, false},
	V58:            {V58, "R58", 4, false},
	V59:            {V59, "R59", 4, false},
	V60:            {V60, "R60", 4, false},
	V61:            {V61, "R61", 4, false},
	V62:            {V62, "R62", 4, false},
	V63:            {V63, "R63", 4, false},
	V64:            {V64, "R64", 4, false},
	V65:            {V65, "R65", 4, false},
	V66:            {V66, "R66", 4, false},
	V67:            {V67, "R67", 4, false},
	V68:            {V68, "R68", 4, false},
	V69:            {V69, "R69", 4, false},
	V70:            {V70, "R70", 4, false},
	V71:            {V71, "R71", 4, false},
	V72:            {V72, "R72", 4, false},
	V73:            {V73, "R73", 4, false},
	V74:            {V74, "R74", 4, false},
	V75:            {V75, "R75", 4, false},
	V76:            {V76, "R76", 4, false},
	V77:            {V77, "R77", 4, false},
	V78:            {V78, "R78", 4, false},
	V79:            {V79, "R79", 4, false},
	V80:            {V80, "R80", 4, false},
	V81:            {V81, "R81", 4, false},
	V82:            {V82, "R82", 4, false},
	V83:            {V83, "R83", 4, false},
	V84:            {V84, "R84", 4, false},
	V85:            {V85, "R85", 4, false},
	V86:            {V86, "R86", 4, false},
	V87:            {V87, "R87", 4, false},
	V88:            {V88, "R88", 4, false},
	V89:            {V89, "R89", 4, false},
	V90:            {V90, "R90", 4, false},
	V91:            {V91, "R91", 4, false},
	V92:            {V92, "R92", 4, false},
	V93:            {V93, "R93", 4, false},
	V94:            {V94, "R94", 4, false},
	V95:            {V95, "R95", 4, false},
	V96:            {V96, "R96", 4, false},
	V97:            {V97, "R97", 4, false},
	V98:            {V98, "R98", 4, false},
	V99:            {V99, "R99", 4, false},
	V100:           {V100, "R100", 4, false},
	V101:           {V101, "R101", 4, false},
	V102:           {V102, "R102", 4, false},
	V103:           {V103, "R103", 4, false},
	V104:           {V104, "R104", 4, false},
	V105:           {V105, "R105", 4, false},
	V106:           {V106, "R106", 4, false},
	V107:           {V107, "R107", 4, false},
	V108:           {V108, "R108", 4, false},
	V109:           {V109, "R109", 4, false},
	V110:           {V110, "R110", 4, false},
	V111:           {V111, "R111", 4, false},
	V112:           {V112, "R112", 4, false},
	V113:           {V113, "R113", 4, false},
	V114:           {V114, "R114", 4, false},
	V115:           {V115, "R115", 4, false},
	V116:           {V116, "R116", 4, false},
	V117:           {V117, "R117", 4, false},
	V118:           {V118, "R118", 4, false},
	V119:           {V119, "R119", 4, false},
	V120:           {V120, "R120", 4, false},
	V121:           {V121, "R121", 4, false},
	V122:           {V122, "R122", 4, false},
	V123:           {V123, "R123", 4, false},
	V124:           {V124, "R124", 4, false},
	V125:           {V125, "R125", 4, false},
	V126:           {V126, "R126", 4, false},
	V127:           {V127, "R127", 4, false},
	V128:           {V128, "R128", 4, false},
	V129:           {V129, "R129", 4, false},
	V130:           {V130, "R130", 4, false},
	V131:           {V131, "R131", 4, false},
	V132:           {V132, "R132", 4, false},
	V133:           {V133, "R133", 4, false},
	V134:           {V134, "R134", 4, false},
	V135:           {V135, "R135", 4, false},
	V136:           {V136, "R136", 4, false},
	V137:           {V137, "R137", 4, false},
	V138:           {V138, "R138", 4, false},
	V139:           {V139, "R139", 4, false},
	V140:           {V140, "R140", 4, false},
	V141:           {V141, "R141", 4, false},
	V142:           {V142, "R142", 4, false},
	V143:           {V143, "R143", 4, false},
	V144:           {V144, "R144", 4, false},
	V145:           {V145, "R145", 4, false},
	V146:           {V146, "R146", 4, false},
	V147:           {V147, "R147", 4, false},
	V148:           {V148, "R148", 4, false},
	V149:           {V149, "R149", 4, false},
	V150:           {V150, "R150", 4, false},
	V151:           {V151, "R151", 4, false},
	V152:           {V152, "R152", 4, false},
	V153:           {V153, "R153", 4, false},
	V154:           {V154, "R154", 4, false},
	V155:           {V155, "R155", 4, false},
	V156:           {V156, "R156", 4, false},
	V157:           {V157, "R157", 4, false},
	V158:           {V158, "R158", 4, false},
	V159:           {V159, "R159", 4, false},
	V160:           {V160, "R160", 4, false},
	V161:           {V161, "R161", 4, false},
	V162:           {V162, "R162", 4, false},
	V163:           {V163, "R163", 4, false},
	V164:           {V164, "R164", 4, false},
	V165:           {V165, "R165", 4, false},
	V166:           {V166, "R166", 4, false},
	V167:           {V167, "R167", 4, false},
	V168:           {V168, "R168", 4, false},
	V169:           {V169, "R169", 4, false},
	V170:           {V170, "R170", 4, false},
	V171:           {V171, "R171", 4, false},
	V172:           {V172, "R172", 4, false},
	V173:           {V173, "R173", 4, false},
	V174:           {V174, "R174", 4, false},
	V175:           {V175, "R175", 4, false},
	V176:           {V176, "R176", 4, false},
	V177:           {V177, "R177", 4, false},
	V178:           {V178, "R178", 4, false},
	V179:           {V179, "R179", 4, false},
	V180:           {V180, "R180", 4, false},
	V181:           {V181, "R181", 4, false},
	V182:           {V182, "R182", 4, false},
	V183:           {V183, "R183", 4, false},
	V184:           {V184, "R184", 4, false},
	V185:           {V185, "R185", 4, false},
	V186:           {V186, "R186", 4, false},
	V187:           {V187, "R187", 4, false},
	V188:           {V188, "R188", 4, false},
	V189:           {V189, "R189", 4, false},
	V190:           {V190, "R190", 4, false},
	V191:           {V191, "R191", 4, false},
	V192:           {V192, "R192", 4, false},
	V193:           {V193, "R193", 4, false},
	V194:           {V194, "R194", 4, false},
	V195:           {V195, "R195", 4, false},
	V196:           {V196, "R196", 4, false},
	V197:           {V197, "R197", 4, false},
	V198:           {V198, "R198", 4, false},
	V199:           {V199, "R199", 4, false},
	V200:           {V200, "R200", 4, false},
	V201:           {V201, "R201", 4, false},
	V202:           {V202, "R202", 4, false},
	V203:           {V203, "R203", 4, false},
	V204:           {V204, "R204", 4, false},
	V205:           {V205, "R205", 4, false},
	V206:           {V206, "R206", 4, false},
	V207:           {V207, "R207", 4, false},
	V208:           {V208, "R208", 4, false},
	V209:           {V209, "R209", 4, false},
	V210:           {V210, "R210", 4, false},
	V211:           {V211, "R211", 4, false},
	V212:           {V212, "R212", 4, false},
	V213:           {V213, "R213", 4, false},
	V214:           {V214, "R214", 4, false},
	V215:           {V215, "R215", 4, false},
	V216:           {V216, "R216", 4, false},
	V217:           {V217, "R217", 4, false},
	V218:           {V218, "R218", 4, false},
	V219:           {V219, "R219", 4, false},
	V220:           {V220, "R220", 4, false},
	V221:           {V221, "R221", 4, false},
	V222:           {V222, "R222", 4, false},
	V223:           {V223, "R223", 4, false},
	V224:           {V224, "R224", 4, false},
	V225:           {V225, "R225", 4, false},
	V226:           {V226, "R226", 4, false},
	V227:           {V227, "R227", 4, false},
	V228:           {V228, "R228", 4, false},
	V229:           {V229, "R229", 4, false},
	V230:           {V230, "R230", 4, false},
	V231:           {V231, "R231", 4, false},
	V232:           {V232, "R232", 4, false},
	V233:           {V233, "R233", 4, false},
	V234:           {V234, "R234", 4, false},
	V235:           {V235, "R235", 4, false},
	V236:           {V236, "R236", 4, false},
	V237:           {V237, "R237", 4, false},
	V238:           {V238, "R238", 4, false},
	V239:           {V239, "R239", 4, false},
	V240:           {V240, "R240", 4, false},
	V241:           {V241, "R241", 4, false},
	V242:           {V242, "R242", 4, false},
	V243:           {V243, "R243", 4, false},
	V244:           {V244, "R244", 4, false},
	V245:           {V245, "R245", 4, false},
	V246:           {V246, "R246", 4, false},
	V247:           {V247, "R247", 4, false},
	V248:           {V248, "R248", 4, false},
	V249:           {V249, "R249", 4, false},
	V250:           {V250, "R250", 4, false},
	V251:           {V251, "R251", 4, false},
	V252:           {V252, "R252", 4, false},
	V253:           {V253, "R253", 4, false},
	V254:           {V254, "R254", 4, false},
	V255:           {V255, "R255", 4, false},
	// TODO How to deal with scalar regs?
	S0:           {S0, "S0", 4, false},
	S1:           {S1, "S1", 4, false},
	S2:           {S2, "S2", 4, false},
	S3:           {S3, "S3", 4, false},
	S4:           {S4, "S4", 4, false},
	S5:           {S5, "S5", 4, false},
	S6:           {S6, "S6", 4, false},
	S7:           {S7, "S7", 4, false},
	S8:           {S8, "S8", 4, false},
	S9:           {S9, "S9", 4, false},
	S10:          {S10, "S10", 4, false},
	S11:          {S11, "S11", 4, false},
	S12:          {S12, "S12", 4, false},
	S13:          {S13, "S13", 4, false},
	S14:          {S14, "S14", 4, false},
	S15:          {S15, "S15", 4, false},
	S16:          {S16, "S16", 4, false},
	S17:          {S17, "S17", 4, false},
	S18:          {S18, "S18", 4, false},
	S19:          {S19, "S19", 4, false},
	S20:          {S20, "S20", 4, false},
	S21:          {S21, "S21", 4, false},
	S22:          {S22, "S22", 4, false},
	S23:          {S23, "S23", 4, false},
	S24:          {S24, "S24", 4, false},
	S25:          {S25, "S25", 4, false},
	S26:          {S26, "S26", 4, false},
	S27:          {S27, "S27", 4, false},
	S28:          {S28, "S28", 4, false},
	S29:          {S29, "S29", 4, false},
	S30:          {S30, "S30", 4, false},
	S31:          {S31, "S31", 4, false},
	S32:          {S32, "S32", 4, false},
	S33:          {S33, "S33", 4, false},
	S34:          {S34, "S34", 4, false},
	S35:          {S35, "S35", 4, false},
	S36:          {S36, "S36", 4, false},
	S37:          {S37, "S37", 4, false},
	S38:          {S38, "S38", 4, false},
	S39:          {S39, "S39", 4, false},
	S40:          {S40, "S40", 4, false},
	S41:          {S41, "S41", 4, false},
	S42:          {S42, "S42", 4, false},
	S43:          {S43, "S43", 4, false},
	S44:          {S44, "S44", 4, false},
	S45:          {S45, "S45", 4, false},
	S46:          {S46, "S46", 4, false},
	S47:          {S47, "S47", 4, false},
	S48:          {S48, "S48", 4, false},
	S49:          {S49, "S49", 4, false},
	S50:          {S50, "S50", 4, false},
	S51:          {S51, "S51", 4, false},
	S52:          {S52, "S52", 4, false},
	S53:          {S53, "S53", 4, false},
	S54:          {S54, "S54", 4, false},
	S55:          {S55, "S55", 4, false},
	S56:          {S56, "S56", 4, false},
	S57:          {S57, "S57", 4, false},
	S58:          {S58, "S58", 4, false},
	S59:          {S59, "S59", 4, false},
	S60:          {S60, "S60", 4, false},
	S61:          {S61, "S61", 4, false},
	S62:          {S62, "S62", 4, false},
	S63:          {S63, "S63", 4, false},
	S64:          {S64, "S64", 4, false},
	S65:          {S65, "S65", 4, false},
	S66:          {S66, "S66", 4, false},
	S67:          {S67, "S67", 4, false},
	S68:          {S68, "S68", 4, false},
	S69:          {S69, "S69", 4, false},
	S70:          {S70, "S70", 4, false},
	S71:          {S71, "S71", 4, false},
	S72:          {S72, "S72", 4, false},
	S73:          {S73, "S73", 4, false},
	S74:          {S74, "S74", 4, false},
	S75:          {S75, "S75", 4, false},
	S76:          {S76, "S76", 4, false},
	S77:          {S77, "S77", 4, false},
	S78:          {S78, "S78", 4, false},
	S79:          {S79, "S79", 4, false},
	S80:          {S80, "S80", 4, false},
	S81:          {S81, "S81", 4, false},
	S82:          {S82, "S82", 4, false},
	S83:          {S83, "S83", 4, false},
	S84:          {S84, "S84", 4, false},
	S85:          {S85, "S85", 4, false},
	S86:          {S86, "S86", 4, false},
	S87:          {S87, "S87", 4, false},
	S88:          {S88, "S88", 4, false},
	S89:          {S89, "S89", 4, false},
	S90:          {S90, "S90", 4, false},
	S91:          {S91, "S91", 4, false},
	S92:          {S92, "S92", 4, false},
	S93:          {S93, "S93", 4, false},
	S94:          {S94, "S94", 4, false},
	S95:          {S95, "S95", 4, false},
	S96:          {S96, "S96", 4, false},
	S97:          {S97, "S97", 4, false},
	S98:          {S98, "S98", 4, false},
	S99:          {S99, "S99", 4, false},
	S100:         {S100, "S100", 4, false},
	S101:         {S101, "S101", 4, false},
	EXEC:         {EXEC, "EXEC", 8, false},
	EXECLO:       {EXECLO, "EXECLO", 4, false},
	EXECHI:       {EXECHI, "EXECHI", 4, false},
	EXECZ:        {EXECZ, "EXECZ", 1, true},
	VCC:          {VCC, "VCC", 8, false},
	VCCLO:        {VCCLO, "VCCLO", 4, false},
	VCCHI:        {VCCHI, "VCCHI", 4, false},
	VCCZ:         {VCCZ, "VCCZ", 1, true},
	SCC:          {SCC, "SCC", 1, true},
	FlatSratch:   {FlatSratch, "FLATSRATCH", 8, false},
	FlatSratchLo: {FlatSratchLo, "FLATSRATCHLO", 4, false},
	FlatSratchHi: {FlatSratchHi, "FLATSRATCHHI", 4, false},
	XnackMask:    {XnackMask, "XNACKMASK", 8, false},
	XnackMaskLo:  {XnackMaskLo, "XNACKMASKLO", 4, false},
	XnackMaskHi:  {XnackMaskHi, "XNACKMASKHI", 4, false},
	Status:       {Status, "STATUS", 4, false},
	Mode:         {Mode, "MODE", 4, false},
	M0:           {M0, "M0", 4, false},
	Trapsts:      {Trapsts, "TRAPSTS", 4, false},
	Tba:          {Tba, "TBA", 8, false},
	TbaLo:        {TbaLo, "TBALO", 4, false},
	TbaHi:        {TbaHi, "TBAHI", 4, false},
	Tma:          {Tma, "TMA", 8, false},
	TmaLo:        {TmaLo, "TMALO", 4, false},
	TmaHi:        {TmaHi, "TMAHI", 4, false},
	Timp0:        {Timp0, "TIMP0", 4, false},
	Timp1:        {Timp1, "TIMP1", 4, false},
	Timp2:        {Timp2, "TIMP2", 4, false},
	Timp3:        {Timp3, "TIMP3", 4, false},
	Timp4:        {Timp4, "TIMP4", 4, false},
	Timp5:        {Timp5, "TIMP5", 4, false},
	Timp6:        {Timp6, "TIMP6", 4, false},
	Timp7:        {Timp7, "TIMP7", 4, false},
	Timp8:        {Timp8, "TIMP8", 4, false},
	Timp9:        {Timp9, "TIMP9", 4, false},
	Timp10:       {Timp10, "TIMP10", 4, false},
	Timp11:       {Timp11, "TIMP11", 4, false},
	VMCNT:        {VMCNT, "VMCNT", 1, false},
	EXPCNT:       {EXPCNT, "EXPCNT", 1, false},
	LGKMCNT:      {LGKMCNT, "LOGKMCNT", 1, false},
}
