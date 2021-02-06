package tensor

import (
	"fmt"
	"math"

	"gitlab.com/akita/dnn/tensor"
	"gitlab.com/akita/mgpusim/driver"
	"gitlab.com/akita/mgpusim/insts"
	"gitlab.com/akita/mgpusim/kernels"
)

var sizeOfFloat32 int = 4
var sizeOfInt32 int = 4

// GPUOperator can perform operations on GPU tensors.
type GPUOperator struct {
	driver       *driver.Driver
	ctx          *driver.Context
	verification bool
	cpuOperator  *tensor.CPUOperator

	sumKernel                           *insts.HsaCo
	transposeKernel                     *insts.HsaCo
	rotateKernel                        *insts.HsaCo
	dilateKernel                        *insts.HsaCo
	im2ColKernel                        *insts.HsaCo
	softmaxExpKernel                    *insts.HsaCo
	reductionSumKernel                  *insts.HsaCo
	softmaxDivKernel                    *insts.HsaCo
	scaleAddKernel                      *insts.HsaCo
	elemWiseMulKernel                   *insts.HsaCo
	rmsPropKernel                       *insts.HsaCo
	adamKernel                          *insts.HsaCo
	reluForwardKernel                   *insts.HsaCo
	reluBackwardKernel                  *insts.HsaCo
	maxPoolingForwardKernel             *insts.HsaCo
	maxPoolingBackwardKernel            *insts.HsaCo
	gemmKernel                          *insts.HsaCo
	crossEntropyDerivativeKernel        *insts.HsaCo
	softmaxCrossEntropyDerivativeKernel *insts.HsaCo
}

// NewGPUOperator creates a new GPU Operator.
func NewGPUOperator(
	gpuDriver *driver.Driver,
	ctx *driver.Context,
) *GPUOperator {
	o := &GPUOperator{
		driver: gpuDriver,
		ctx:    ctx,
	}

	o.loadKernels()

	return o
}

// EnableVerification will run the same operations in a CPU operator and
// compare the results.
func (o *GPUOperator) EnableVerification() {
	o.verification = true
	o.cpuOperator = &tensor.CPUOperator{}
}

func (o *GPUOperator) gpuTensorToCPUTensor(
	t tensor.Tensor,
) *tensor.SimpleTensor {
	out := o.cpuOperator.CreateWithData(t.Vector(), t.Size(), t.Descriptor())
	return out.(*tensor.SimpleTensor)
}

func (o *GPUOperator) tensorMustMatch(a, b tensor.Tensor) {
	o.descriptorMustMatch(a, b)
	o.sizeMustMatch(a, b)
	o.valueMustMatch(a, b)
}

func (o *GPUOperator) descriptorMustMatch(a, b tensor.Tensor) {
	descriptorA := a.Descriptor()
	descriptorB := b.Descriptor()
	if descriptorA != descriptorB {
		panic("discriptor not match")
	}
}

func (o *GPUOperator) sizeMustMatch(a, b tensor.Tensor) {
	sizeA := a.Size()
	sizeB := b.Size()
	if len(sizeA) != len(sizeB) {
		panic("dimension mismatch")
	}
	for i := range sizeA {
		if sizeA[i] != sizeB[i] {
			panic("size mismatch")
		}
	}
}

func (o *GPUOperator) valueMustMatch(a, b tensor.Tensor) {
	vectorA := a.Vector()
	vectorB := b.Vector()
	for i := range vectorA {
		if math.Abs(vectorA[i]-vectorB[i]) > 1e-3 {
			panic("value mismatch")
		}
	}
}

func (o *GPUOperator) loadKernels() {
	kernelBytes := _escFSMustByte(false, "/operator.hsaco")
	loadKernel(&o.sumKernel, kernelBytes, "sum_one_axis")
	loadKernel(&o.transposeKernel, kernelBytes, "transpose_tensor")
	loadKernel(&o.rotateKernel, kernelBytes, "rotate_tensor")
	loadKernel(&o.dilateKernel, kernelBytes, "dilate_tensor")
	loadKernel(&o.im2ColKernel, kernelBytes, "im2col")
	loadKernel(&o.softmaxExpKernel, kernelBytes, "softmax_exp")
	loadKernel(&o.softmaxDivKernel, kernelBytes, "softmax_div")
	loadKernel(&o.scaleAddKernel, kernelBytes, "scaleAdd")
	loadKernel(&o.elemWiseMulKernel, kernelBytes, "mul")
	loadKernel(&o.rmsPropKernel, kernelBytes, "rmsProp")
	loadKernel(&o.adamKernel, kernelBytes, "adam")
	loadKernel(&o.reluForwardKernel, kernelBytes, "reluForward")
	loadKernel(&o.reluBackwardKernel, kernelBytes, "reluBackward")

	kernelBytes = _escFSMustByte(false, "/maxpooling.hsaco")
	loadKernel(&o.maxPoolingForwardKernel, kernelBytes, "MaxPoolForward")
	loadKernel(&o.maxPoolingBackwardKernel, kernelBytes, "MaxPoolBackward")

	kernelBytes = _escFSMustByte(false, "/gemm.hsaco")
	loadKernel(&o.gemmKernel, kernelBytes, "gemm")

	kernelBytes = _escFSMustByte(false, "/cross_entropy.hsaco")
	loadKernel(&o.crossEntropyDerivativeKernel, kernelBytes,
		"cross_entropy_derivative")
	loadKernel(&o.softmaxCrossEntropyDerivativeKernel, kernelBytes,
		"softmax_cross_entropy_derivative")
}

func loadKernel(hsaco **insts.HsaCo, kernelBytes []byte, name string) {
	*hsaco = kernels.LoadProgramFromMemory(kernelBytes, name)
	if *hsaco == nil {
		panic("Failed to load " + name + "kernel")
	}
}

// Create creates a new GPU tensor
func (o *GPUOperator) Create(size []int) tensor.Tensor {
	t := &Tensor{
		driver: o.driver,
		ctx:    o.ctx,
		size:   size,
	}

	t.ptr = o.driver.AllocateMemory(o.ctx, uint64(t.NumElement()*sizeOfFloat32))

	return t
}

// CreateWithData creates the tensor and copies the given data to the GPU
// memory.
func (o *GPUOperator) CreateWithData(
	data []float64,
	size []int,
	descriptor string,
) tensor.Tensor {
	t := o.Create(size).(*Tensor)
	t.descriptor = descriptor

	f32Data := f64SliceToF32Slice(data)

	o.driver.MemCopyH2D(o.ctx, t.ptr, f32Data)

	return t
}

func f64SliceToF32Slice(in []float64) []float32 {
	f32Data := make([]float32, len(in))

	for i := 0; i < len(in); i++ {
		f32Data[i] = float32(in[i])
	}

	return f32Data
}

// Free releases the allocated GPU memory.
func (o *GPUOperator) Free(t tensor.Tensor) {
	o.driver.FreeMemory(o.ctx, t.(*Tensor).ptr)
	t.(*Tensor).ptr = 0
}

// Copy copies data from one tensor to another tensor. The src and dst tensor
// must have the same number of elements.
func (o *GPUOperator) Copy(dst tensor.Tensor, src tensor.Tensor) {
	d := dst.(*Tensor)
	s := src.(*Tensor)

	if d.NumElement() != s.NumElement() {
		panic("mismatch in size")
	}

	o.driver.MemCopyD2D(o.ctx, d.ptr, s.ptr, dst.NumElement()*sizeOfFloat32)
}

// Clone duplicates the input tensor.
func (o *GPUOperator) Clone(t tensor.Tensor) tensor.Tensor {
	inT := t.(*Tensor)
	outT := o.Create(t.Size()).(*Tensor)

	outT.size = make([]int, len(inT.size))
	copy(outT.size, inT.size)

	o.Copy(outT, inT)

	return outT
}

// Dump writes the content of the tensor to a string.
func (o *GPUOperator) Dump(t tensor.Tensor) string {
	v := make([]float32, t.NumElement())
	o.driver.MemCopyD2H(o.ctx, v, t.(*Tensor).ptr)

	dimSize := make([]int, len(t.Size()))
	product := 1
	for i := len(t.Size()) - 1; i >= 0; i-- {
		product *= t.Size()[i]
		dimSize[i] = product
	}

	out := ""
	indent := 0
	for i := 0; i < t.NumElement(); i++ {
		for _, d := range dimSize {
			if i%d == 0 {
				out += "\n"
				for k := 0; k < indent; k++ {
					out += " "
				}
				out += "["
				indent++
			}
		}

		out += fmt.Sprintf("%4f, ", v[i])

		for _, d := range dimSize {
			if (i+1)%d == 0 {
				out += "],"
				indent--
			}
		}
	}
	out += "\n"

	return out
}

// Init sets the data of the tensor
func (o *GPUOperator) Init(t tensor.Tensor, data []float64) {
	if t.NumElement() != len(data) {
		panic("mismatch in buffer shape")
	}

	f32Data := f64SliceToF32Slice(data)

	o.driver.MemCopyH2D(o.ctx, t.(*Tensor).ptr, f32Data)
}

// Slice will create another tensor that shares part of the buffer with the
// input tensor.
func (o *GPUOperator) Slice(t tensor.Tensor, start int, end int) tensor.Tensor {
	out := &Tensor{
		driver: o.driver,
		ctx:    o.ctx,

		size: []int{end - start},
		ptr: driver.GPUPtr(uint64(t.(*Tensor).ptr) +
			uint64(start*sizeOfFloat32)),
	}

	return out
}

// Repeat will create another tensor that duplicates the input tensor by n
// times.
func (o *GPUOperator) Repeat(t tensor.Tensor, times int) tensor.Tensor {
	numElem := t.NumElement()
	out := o.Create([]int{numElem * times}).(*Tensor)

	inPtr := t.(*Tensor).ptr
	outPtr := uint64(out.ptr)

	for i := 0; i < times; i++ {
		o.driver.MemCopyD2D(o.ctx,
			driver.GPUPtr(outPtr+uint64(i*numElem*sizeOfFloat32)),
			inPtr,
			numElem*sizeOfFloat32,
		)
	}

	return out
}

// Clear sets the content of the tensor to 0.
func (o *GPUOperator) Clear(t tensor.Tensor) {
	data := make([]float64, t.NumElement())

	o.Init(t, data)
}

// Zeros creates a tensor prefilled with zeros.
func (o *GPUOperator) Zeros(size []int) tensor.Tensor {
	t := o.Create(size)

	o.Clear(t)

	return t
}

// Reshape creates another tensor with the same elements but a different size.
func (o *GPUOperator) Reshape(t tensor.Tensor, newSize []int) tensor.Tensor {
	out := o.Clone(t)

	out.SetSize(newSize)

	return out
}

type transposeKernelArgs struct {
	In, Out, InSize, OutSize, Order driver.GPUPtr
	InIndexBuf, OutIndexBuf         driver.GPUPtr
	Dim, Padding                    int32
	OffsetX, OffsetY, OffsetZ       int64
}

// Transpose reorders the axises of the tensor.
func (o *GPUOperator) Transpose(t tensor.Tensor, order []int) tensor.Tensor {
	if len(order) != len(t.Size()) {
		panic("order should include all axes")
	}

	dim := len(order)
	hOrder := make([]int32, dim)
	hInSize := make([]int32, dim)
	hOutSize := make([]int32, dim)
	outSize := make([]int, dim)
	for i := 0; i < dim; i++ {
		hOrder[i] = int32(order[i])
		hInSize[i] = int32(t.Size()[i])
		hOutSize[i] = int32(t.Size()[order[i]])
		outSize[i] = t.Size()[order[i]]
	}

	output := o.Create(outSize).(*Tensor)

	dOrder := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dOrder, hOrder)
	defer o.driver.FreeMemory(o.ctx, dOrder)
	dInSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dInSize, hInSize)
	defer o.driver.FreeMemory(o.ctx, dInSize)
	dOutSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dOutSize, hOutSize)
	defer o.driver.FreeMemory(o.ctx, dOutSize)
	dInIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(t.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dInIndexBuf)
	dOutIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(t.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dOutIndexBuf)

	args := transposeKernelArgs{
		In:          t.(*Tensor).ptr,
		Out:         output.ptr,
		InSize:      dInSize,
		OutSize:     dOutSize,
		Order:       dOrder,
		InIndexBuf:  dInIndexBuf,
		OutIndexBuf: dOutIndexBuf,
		Dim:         int32(len(order)),
	}
	o.driver.LaunchKernel(o.ctx,
		o.transposeKernel,
		[3]uint32{uint32(t.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	output.descriptor = ""
	for i := 0; i < len(t.Descriptor()); i++ {
		output.descriptor += string(t.Descriptor()[order[i]])
	}

	return output
}

type rotateKernelArgs struct {
	In, Out, InSize, OutSize  driver.GPUPtr
	InIndexBuf, OutIndexBuf   driver.GPUPtr
	Dim, Padding              int32
	OffsetX, OffsetY, OffsetZ int64
}

// Rotate180 rotates the lowest two dimensions of the tensor by 180 degree.
func (o *GPUOperator) Rotate180(t tensor.Tensor) tensor.Tensor {
	dim := len(t.Size())
	hInSize := make([]int32, dim)
	hOutSize := make([]int32, dim)
	outSize := make([]int, dim)
	for i := 0; i < dim; i++ {
		hInSize[i] = int32(t.Size()[i])
		hOutSize[i] = int32(t.Size()[i])
		outSize[i] = t.Size()[i]
	}

	output := o.Create(outSize).(*Tensor)

	dInSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dInSize, hInSize)
	defer o.driver.FreeMemory(o.ctx, dInSize)
	dOutSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dOutSize, hOutSize)
	defer o.driver.FreeMemory(o.ctx, dOutSize)
	dInIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(t.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dInIndexBuf)
	dOutIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(t.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dOutIndexBuf)

	args := rotateKernelArgs{
		In:          t.(*Tensor).ptr,
		Out:         output.ptr,
		InSize:      dInSize,
		OutSize:     dOutSize,
		InIndexBuf:  dInIndexBuf,
		OutIndexBuf: dOutIndexBuf,
		Dim:         int32(len(t.Size())),
	}
	o.driver.LaunchKernel(o.ctx,
		o.rotateKernel,
		[3]uint32{uint32(t.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	return output
}

type dilateKernelArgs struct {
	In, Out, InSize, OutSize  driver.GPUPtr
	Dilate                    driver.GPUPtr
	InIndexBuf, OutIndexBuf   driver.GPUPtr
	Dim, Padding              int32
	OffsetX, OffsetY, OffsetZ int64
}

// Dilate addes 0s between rows and columns.
func (o *GPUOperator) Dilate(t tensor.Tensor, dilate []int) tensor.Tensor {
	dim := len(t.Size())
	hDilate := []int32{int32(dilate[0]), int32(dilate[1])}

	outSize := make([]int, len(t.Size()))
	copy(outSize, t.Size())

	outSize[len(outSize)-1] = (outSize[len(outSize)-1]-1)*dilate[1] + 1
	outSize[len(outSize)-2] = (outSize[len(outSize)-2]-1)*dilate[0] + 1
	output := o.Create(outSize).(*Tensor)

	hInSize := make([]int32, dim)
	hOutSize := make([]int32, dim)
	for i := 0; i < dim; i++ {
		hInSize[i] = int32(t.Size()[i])
		hOutSize[i] = int32(outSize[i])
	}

	dInSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dInSize, hInSize)
	defer o.driver.FreeMemory(o.ctx, dInSize)
	dOutSize := o.driver.AllocateMemory(o.ctx, uint64(dim*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dOutSize, hOutSize)
	defer o.driver.FreeMemory(o.ctx, dOutSize)
	dDilate := o.driver.AllocateMemory(o.ctx, uint64(2*sizeOfInt32))
	o.driver.MemCopyH2D(o.ctx, dDilate, hDilate)
	defer o.driver.FreeMemory(o.ctx, dDilate)
	dInIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(output.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dInIndexBuf)
	dOutIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(output.NumElement()*dim*sizeOfInt32))
	defer o.driver.FreeMemory(o.ctx, dOutIndexBuf)

	args := dilateKernelArgs{
		In:          t.(*Tensor).ptr,
		Out:         output.ptr,
		InSize:      dInSize,
		OutSize:     dOutSize,
		Dilate:      dDilate,
		InIndexBuf:  dInIndexBuf,
		OutIndexBuf: dOutIndexBuf,
		Dim:         int32(len(t.Size())),
	}
	o.driver.LaunchKernel(o.ctx,
		o.dilateKernel,
		[3]uint32{uint32(output.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	return output
}

// Sum reduces the number of axes by summing the numbers on given axes.
func (o *GPUOperator) Sum(t tensor.Tensor, axis []int) tensor.Tensor {
	var in, out tensor.Tensor

	o.axisMustBeIncreasing(axis)

	in = t
	for i, a := range axis {
		out = o.sumOneAxis(in, a-i)

		if i > 0 {
			o.Free(in)
		}

		in = out
	}

	return out
}

func (o *GPUOperator) axisMustBeIncreasing(axis []int) {
	for i := 1; i < len(axis); i++ {
		if axis[i] < axis[i-1] {
			panic("axis not increasing")
		}
	}
}

type sumOneAxisKernelArgs struct {
	In, Out, InSize, OutSize  driver.GPUPtr
	InDim, Axis               int32
	InIndexBuf, OutIndexBuf   driver.GPUPtr
	OffsetX, OffsetY, OffsetZ int64
}

func (o *GPUOperator) sumOneAxis(t tensor.Tensor, axis int) tensor.Tensor {
	outSize := make([]int, 0)
	for i := range t.Size() {
		if i != axis {
			outSize = append(outSize, t.Size()[i])
		}
	}

	out := o.Create(outSize)

	hOutSize := make([]int32, len(outSize))
	for i := range outSize {
		hOutSize[i] = int32(outSize[i])
	}

	hInSize := make([]int32, len(t.Size()))
	for i := range t.Size() {
		hInSize[i] = int32(t.Size()[i])
	}

	localSize := 64
	globalSize := out.NumElement()

	dInSize := o.driver.AllocateMemory(o.ctx, uint64(t.Dim()*4))
	o.driver.MemCopyH2D(o.ctx, dInSize, hInSize)
	defer o.driver.FreeMemory(o.ctx, dInSize)

	dOutSize := o.driver.AllocateMemory(o.ctx, uint64(len(outSize)*4))
	o.driver.MemCopyH2D(o.ctx, dOutSize, hOutSize)
	defer o.driver.FreeMemory(o.ctx, dOutSize)

	dInIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(globalSize*t.Dim()*4))
	defer o.driver.FreeMemory(o.ctx, dInIndexBuf)

	dOutIndexBuf := o.driver.AllocateMemory(o.ctx,
		uint64(globalSize*out.Dim()*4))
	defer o.driver.FreeMemory(o.ctx, dOutIndexBuf)

	args := sumOneAxisKernelArgs{
		In:          t.(*Tensor).ptr,
		Out:         out.(*Tensor).ptr,
		InSize:      dInSize,
		OutSize:     dOutSize,
		InDim:       int32(t.Dim()),
		Axis:        int32(axis),
		InIndexBuf:  dInIndexBuf,
		OutIndexBuf: dOutIndexBuf,
	}

	o.driver.LaunchKernel(o.ctx, o.sumKernel,
		[3]uint32{uint32(globalSize), 1, 1},
		[3]uint16{uint16(localSize), 1, 1},
		&args,
	)

	return out
}

type gemmKernArgs struct {
	M, N, K                   int32
	Alpha, Beta               float32
	Padding                   int32
	A, B, C, D                driver.GPUPtr
	OffsetX, OffsetY, OffsetZ int32
}

// Gemm performs alpha * A * B + beta * C operation.
func (o *GPUOperator) Gemm(
	transA, transB bool,
	alpha, beta float64,
	a, b, c tensor.Tensor,
) tensor.Tensor {
	tempA := a
	if transA {
		tempA = o.Transpose(a, []int{1, 0})
	}

	tempB := b
	if transB {
		tempB = o.Transpose(b, []int{1, 0})
	}

	d := o.matrixMultiplication(alpha, beta, tempA, tempB, c)

	if transA {
		o.Free(tempA)
	}

	if transB {
		o.Free(tempB)
	}

	if o.verification {
		cpuA := o.gpuTensorToCPUTensor(a)
		cpuB := o.gpuTensorToCPUTensor(b)
		cpuC := o.gpuTensorToCPUTensor(c)
		cpuOut := o.cpuOperator.Gemm(
			transA, transB, alpha, beta, cpuA, cpuB, cpuC)
		o.tensorMustMatch(d, cpuOut)
		fmt.Println("Gemm verified.")
	}

	return d
}

func (o *GPUOperator) matrixMultiplication(
	alpha, beta float64,
	a, b, c tensor.Tensor,
) tensor.Tensor {
	o.gemmDimMustBeValid(a, b, c)

	m := a.Size()[0]
	n := b.Size()[1]
	k := b.Size()[0]

	blockSize := 16
	wiWidth := uint32(n)
	wiHeight := uint32(m)

	d := o.Create([]int{m, n})

	kernArg := gemmKernArgs{
		M:     int32(m),
		N:     int32(n),
		K:     int32(k),
		Alpha: float32(alpha),
		Beta:  float32(beta),
		A:     a.(*Tensor).ptr,
		B:     b.(*Tensor).ptr,
		C:     c.(*Tensor).ptr,
		D:     d.(*Tensor).ptr,
	}

	o.driver.LaunchKernel(
		o.ctx,
		o.gemmKernel,
		[3]uint32{wiWidth, wiHeight, 1},
		[3]uint16{uint16(blockSize), uint16(blockSize), 1},
		&kernArg,
	)

	return d
}

func (o *GPUOperator) gemmDimMustBeValid(a, b, c tensor.Tensor) {
	if a.Dim() != 2 {
		panic("not a matrix")
	}

	if b.Dim() != 2 {
		panic("not a matrix")
	}

	if c.Dim() != 2 {
		panic("not a matrix")
	}

	if a.Size()[1] != b.Size()[0] {
		panic("width of matrix A does not match height of matrix B")
	}

	if a.Size()[0] != c.Size()[0] || b.Size()[1] != c.Size()[1] {
		panic("matrix C size mismatch")
	}
}

type im2ColKernelArg struct {
	Input, Output             driver.GPUPtr
	InputDim, MaskDim         [2]uint32
	Stride, Pad, Dilation     [2]uint32
	Channel, Batch            uint32
	OffsetX, OffsetY, OffsetZ int64
}

// Im2Col converts images to colums so that convolutional operations can be
// completed with GEMM.
func (o *GPUOperator) Im2Col(
	t tensor.Tensor,
	kernelSize, padding, stride, dilation []int,
) tensor.Tensor {
	inputSize := t.Size()

	batch := inputSize[0]
	channel := inputSize[1]
	width := inputSize[2]
	height := inputSize[3]

	kernelHeight := kernelSize[0]
	kernelWidth := kernelSize[1]

	effKernelHeight := (kernelSize[0]-1)*dilation[0] + 1
	effKernelWidth := (kernelSize[1]-1)*dilation[1] + 1

	fieldHeight := (height-effKernelHeight+2*padding[0])/stride[0] + 1
	fieldWidth := (width-effKernelWidth+2*padding[1])/stride[1] + 1

	outWidth := fieldHeight * fieldWidth * batch
	outHeight := kernelHeight * kernelWidth * channel

	output := o.Create([]int{outHeight, outWidth})

	kernArg := im2ColKernelArg{
		Input:    t.(*Tensor).ptr,
		Output:   output.(*Tensor).ptr,
		InputDim: [2]uint32{uint32(inputSize[3]), uint32(inputSize[2])},
		MaskDim:  [2]uint32{uint32(kernelSize[1]), uint32(kernelSize[0])},
		Stride:   [2]uint32{uint32(stride[1]), uint32(stride[0])},
		Pad:      [2]uint32{uint32(padding[1]), uint32(padding[0])},
		Dilation: [2]uint32{uint32(dilation[1]), uint32(dilation[0])},
		Channel:  uint32(inputSize[1]),
		Batch:    uint32(inputSize[0]),
	}
	gridSize := fieldWidth * fieldHeight * inputSize[0]

	o.driver.LaunchKernel(
		o.ctx,
		o.im2ColKernel,
		[3]uint32{uint32(gridSize), 1, 1},
		[3]uint16{uint16(64), 1, 1},
		&kernArg,
	)

	return output
}

type maxPoolingForwardKernelArgs struct {
	NThreads, Padding            int32
	BottomData                   driver.GPUPtr
	Num, Channels, Height, Width int32
	PooledH, PooledW             int32
	KernelH, KernelW             int32
	StrideH, StrideW             int32
	PadH, PadW                   int32
	TopData, MaskData            driver.GPUPtr
	OffsetX, OffsetY, OffsetZ    int64
}

// MaxPoolingForward calculates the forward propagation of the max pooling
// layer.
func (o *GPUOperator) MaxPoolingForward(
	t tensor.Tensor,
	kernelSize, padding, stride []int,
) (out tensor.Tensor, mask tensor.Tensor) {
	input := t.(*Tensor)
	n := input.size[0]
	c := input.size[1]
	hIn := input.size[2]
	wIn := input.size[3]

	hOut := (hIn+2*padding[0]-kernelSize[0])/stride[0] + 1
	wOut := (wIn+2*padding[1]-kernelSize[1])/stride[1] + 1

	outT := o.Create([]int{n, c, hOut, wOut}).(*Tensor)
	maskT := o.Create([]int{n, c, hOut, wOut}).(*Tensor)

	kernArg := maxPoolingForwardKernelArgs{
		NThreads:   int32(n * c * hOut * wOut),
		BottomData: input.ptr,
		Num:        int32(n),
		Channels:   int32(c),
		Height:     int32(hIn),
		Width:      int32(wIn),
		PooledH:    int32(hOut),
		PooledW:    int32(wOut),
		KernelH:    int32(kernelSize[0]),
		KernelW:    int32(kernelSize[1]),
		StrideH:    int32(stride[0]),
		StrideW:    int32(stride[1]),
		PadH:       int32(padding[0]),
		PadW:       int32(padding[1]),
		TopData:    outT.ptr,
		MaskData:   maskT.ptr,
	}

	o.driver.LaunchKernel(
		o.ctx,
		o.maxPoolingForwardKernel,
		[3]uint32{uint32(n * c * hOut * wOut), 1, 1},
		[3]uint16{64, 1, 1},
		&kernArg,
	)

	return outT, maskT
}

type maxPoolingBackwardKernelArgs struct {
	NThreads, Padding            int32
	TopDiff, TopMask             driver.GPUPtr
	Num, Channels, Height, Width int32
	PooledHeight, PooledWidth    int32
	KernelH, KernelW             int32
	StrideH, StrideW             int32
	PadH, PadW                   int32
	BottomDiff                   driver.GPUPtr
	OffsetX, OffsetY, OffsetZ    int64
}

// MaxPoolingBackward calculates the backward propagation of the max pooling
// layer.
func (o *GPUOperator) MaxPoolingBackward(
	forwardIn, backwardIn tensor.Tensor,
	mask tensor.Tensor,
	kernelSize, padding, stride []int,
) tensor.Tensor {
	n := forwardIn.Size()[0]
	c := forwardIn.Size()[1]
	hIn := forwardIn.Size()[2]
	wIn := forwardIn.Size()[3]
	hOut := backwardIn.Size()[2]
	wOut := backwardIn.Size()[3]

	out := o.Create([]int{n, c, hIn, wIn})

	kernArg := maxPoolingBackwardKernelArgs{
		NThreads:     int32(n * c * hIn * hOut),
		TopDiff:      backwardIn.(*Tensor).ptr,
		TopMask:      mask.(*Tensor).ptr,
		Num:          int32(n),
		Channels:     int32(c),
		Height:       int32(hIn),
		Width:        int32(wIn),
		PooledHeight: int32(hOut),
		PooledWidth:  int32(wOut),
		KernelH:      int32(kernelSize[0]),
		KernelW:      int32(kernelSize[1]),
		StrideH:      int32(stride[0]),
		StrideW:      int32(stride[1]),
		PadH:         int32(padding[0]),
		PadW:         int32(padding[1]),
		BottomDiff:   out.(*Tensor).ptr,
	}

	o.driver.LaunchKernel(o.ctx,
		o.maxPoolingBackwardKernel,
		[3]uint32{uint32(n * c * hIn * wIn), 1, 1},
		[3]uint16{64, 1, 1},
		&kernArg)

	return out
}

// AvgPoolingForward calculates the forward propagation of the average pooling
// layer.
func (o *GPUOperator) AvgPoolingForward(t tensor.Tensor, kernelSize []int, padding []int, stride []int) tensor.Tensor {
	panic("not implemented") // TODO: Implement
}

// AvgPoolingBackward claculates the backward propagation of the average pooling
// layer.
func (o *GPUOperator) AvgPoolingBackward(
	forwardIn, backwardIn tensor.Tensor,
	kernelSize, padding, stride []int,
) tensor.Tensor {
	panic("not implemented") // TODO: Implement
}

type softmaxExpKernelArg struct {
	Input                     driver.GPUPtr
	Output                    driver.GPUPtr
	N, Padding                int32
	OffsetX, OffsetY, OffsetZ int64
}

type reductionSumKernelArg struct {
	Data                      driver.GPUPtr
	PartialSums               driver.LocalPtr
	Padding                   int32
	Output                    driver.GPUPtr
	InputN, Padding2          int32
	OffsetX, OffsetY, OffsetZ int64
}

type softmaxDivKernelArg struct {
	ExpInput                  driver.GPUPtr
	Output                    driver.GPUPtr
	Denominator               driver.GPUPtr
	NumElement, BatchSize     int32
	OffsetX, OffsetY, OffsetZ int64
}

// Softmax performs the softmax operation.
func (o *GPUOperator) Softmax(t tensor.Tensor) tensor.Tensor {
	o.mustBeTwoDimension(t)

	input := t.(*Tensor)
	output := o.Create(input.size).(*Tensor)
	expInput := o.Create(
		[]int{input.size[0], t.NumElement() / input.size[0]},
	).(*Tensor)
	defer o.Free(expInput)

	expArgs := softmaxExpKernelArg{
		Input:  input.ptr,
		Output: expInput.ptr,
		N:      int32(input.NumElement()),
	}
	o.driver.LaunchKernel(o.ctx, o.softmaxExpKernel,
		[3]uint32{uint32(input.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&expArgs,
	)

	denominator := o.Sum(expInput, []int{1})

	divArgs := softmaxDivKernelArg{
		ExpInput:    expInput.ptr,
		Output:      output.ptr,
		Denominator: denominator.(*Tensor).ptr,
		NumElement:  int32(expInput.NumElement()),
		BatchSize:   int32(t.Size()[0]),
	}
	o.driver.LaunchKernel(o.ctx, o.softmaxDivKernel,
		[3]uint32{uint32(expInput.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&divArgs,
	)

	return output
}

func (o *GPUOperator) mustBeTwoDimension(t tensor.Tensor) {
	if t.Dim() != 2 {
		panic("Tensor is not two dimension")
	}
}

// CrossEntropy calculates the cross entropy of the output.
func (o *GPUOperator) CrossEntropy(t tensor.Tensor, label []int) float64 {
	o.mustBeTwoDimension(t)

	loss := 0.0
	inputV := t.Vector()
	for i := 0; i < t.Size()[0]; i++ {
		start := i * t.Size()[1]
		end := start + t.Size()[1]

		inputSlice := inputV[start:end]

		loss += -math.Log(inputSlice[label[i]])
	}

	loss /= float64(t.Size()[0])

	return loss
}

type crossEntropyDerivativeArgs struct {
	Output, Input, Label      driver.GPUPtr
	BatchSize, NumPerImage    int32
	OffsetX, OffsetY, OffsetZ int64
}

// CrossEntropyDerivative calculates the derivative using cross entropies.
func (o *GPUOperator) CrossEntropyDerivative(
	t tensor.Tensor, label []int,
) tensor.Tensor {
	hLabel := make([]int32, len(label))
	for i := 0; i < len(label); i++ {
		hLabel[i] = int32(label[i])
	}
	dLabel := o.driver.AllocateMemory(o.ctx, uint64(len(label)*4))
	defer o.driver.FreeMemory(o.ctx, dLabel)
	o.driver.MemCopyH2D(o.ctx, dLabel, hLabel)

	output := o.Create(t.Size()).(*Tensor)

	args := crossEntropyDerivativeArgs{
		Output:      output.ptr,
		Input:       t.(*Tensor).ptr,
		Label:       dLabel,
		BatchSize:   int32(t.Size()[0]),
		NumPerImage: int32(t.Size()[1]),
	}

	o.driver.LaunchKernel(o.ctx, o.crossEntropyDerivativeKernel,
		[3]uint32{uint32(t.Size()[0] * t.Size()[1]), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	return output
}

// SoftmaxCrossEntropyDerivative calculates the derivatives using both softmax /// and cross entropy algorithms.
func (o *GPUOperator) SoftmaxCrossEntropyDerivative(
	t tensor.Tensor,
	label []int,
) tensor.Tensor {
	hLabel := make([]int32, len(label))
	for i := 0; i < len(label); i++ {
		hLabel[i] = int32(label[i])
	}
	dLabel := o.driver.AllocateMemory(o.ctx, uint64(len(label)*4))
	defer o.driver.FreeMemory(o.ctx, dLabel)
	o.driver.MemCopyH2D(o.ctx, dLabel, hLabel)

	output := o.Create(t.Size()).(*Tensor)

	args := crossEntropyDerivativeArgs{
		Output:      output.ptr,
		Input:       t.(*Tensor).ptr,
		Label:       dLabel,
		BatchSize:   int32(t.Size()[0]),
		NumPerImage: int32(t.Size()[1]),
	}

	o.driver.LaunchKernel(o.ctx, o.softmaxCrossEntropyDerivativeKernel,
		[3]uint32{uint32(t.Size()[0] * t.Size()[1]), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	if o.verification {
		cpuIn := o.gpuTensorToCPUTensor(t)
		cpuOut := o.cpuOperator.SoftmaxCrossEntropyDerivative(cpuIn, label)
		o.tensorMustMatch(output, cpuOut)
		fmt.Println("SoftmaxCrossEntropyDerivative verified.")
	}

	return output
}

type elemWiseMulKernArg struct {
	Out, In1, In2             driver.GPUPtr
	N, Padding                int32
	OffsetX, OffsetY, OffsetZ int64
}

// ElementWiseMul calculates the element multiplication of A and B.
func (o *GPUOperator) ElementWiseMul(
	a, b tensor.Tensor,
) tensor.Tensor {
	if a.NumElement() != b.NumElement() {
		panic("size not match")
	}

	out := o.Create(a.Size()).(*Tensor)
	args := elemWiseMulKernArg{
		Out: out.ptr,
		In1: a.(*Tensor).ptr,
		In2: b.(*Tensor).ptr,
		N:   int32(a.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.elemWiseMulKernel,
		[3]uint32{uint32(a.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	if o.verification {
		cpuA := o.gpuTensorToCPUTensor(a)
		cpuB := o.gpuTensorToCPUTensor(a)
		cpuOut := o.cpuOperator.ElementWiseMul(cpuA, cpuB)
		o.tensorMustMatch(out, cpuOut)
		fmt.Println("ElementWiseMul verified.")
	}

	return out
}

type scaleAddKernArg struct {
	Out, In1, In2             driver.GPUPtr
	Alpha, Beta               float32
	N, Padding                int32
	OffsetX, OffsetY, OffsetZ int64
}

// ScaleAdd performs element-wide alpha*A + beta*B operation.
func (o *GPUOperator) ScaleAdd(
	alpha, beta float64,
	a, b tensor.Tensor,
) tensor.Tensor {
	if a.NumElement() != b.NumElement() {
		panic("size not match")
	}

	out := o.Create(a.Size()).(*Tensor)
	args := scaleAddKernArg{
		Out:   out.ptr,
		In1:   a.(*Tensor).ptr,
		In2:   b.(*Tensor).ptr,
		Alpha: float32(alpha),
		Beta:  float32(beta),
		N:     int32(a.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.scaleAddKernel,
		[3]uint32{uint32(a.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args,
	)

	if o.verification {
		cpuA := o.gpuTensorToCPUTensor(a)
		cpuB := o.gpuTensorToCPUTensor(a)
		cpuOut := o.cpuOperator.ScaleAdd(alpha, beta, cpuA, cpuB)
		o.tensorMustMatch(out, cpuOut)
		fmt.Println("ScaleAdd verified.")
	}

	return out
}

type rmsPropKernArg struct {
	Params, Gradients, SHistory driver.GPUPtr
	SmoothFactor, LearningRate  float32
	N, Padding                  int32
	OffsetX, OffsetY, OffsetZ   int64
}

// RMSProp uses the RMSProp algorithm to update the parameters
func (o *GPUOperator) RMSProp(
	params, gradient, sHistory tensor.Tensor,
	smoothFactor, learningRate float64,
) {
	if params.NumElement() != gradient.NumElement() ||
		params.NumElement() != sHistory.NumElement() {
		panic("size mismatch")
	}

	args := rmsPropKernArg{
		Params:       params.(*Tensor).ptr,
		Gradients:    gradient.(*Tensor).ptr,
		SHistory:     sHistory.(*Tensor).ptr,
		SmoothFactor: float32(smoothFactor),
		LearningRate: float32(learningRate),
		N:            int32(params.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.rmsPropKernel,
		[3]uint32{uint32(params.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args)
}

type adamKernArg struct {
	Params, Gradients, SHistory, VHistory      driver.GPUPtr
	SmoothFactor1, SmoothFactor2, LearningRate float32
	N                                          int32
	OffsetX, OffsetY, OffsetZ                  int64
}

//Adam uses the Adam algorithm to update the parameters
func (o *GPUOperator) Adam(
	params, gradient, vHistory, sHistory tensor.Tensor,
	smoothFactor1, smoothFactor2, learningRate float64,
) {
	if params.NumElement() != gradient.NumElement() ||
		params.NumElement() != sHistory.NumElement() ||
		params.NumElement() != vHistory.NumElement() {
		panic("size mismatch")
	}

	var cpuParams, cpuGradient, cpuSHistory, cpuVHistory *tensor.SimpleTensor
	if o.verification {
		cpuParams = o.gpuTensorToCPUTensor(params)
		cpuGradient = o.gpuTensorToCPUTensor(gradient)
		cpuSHistory = o.gpuTensorToCPUTensor(sHistory)
		cpuVHistory = o.gpuTensorToCPUTensor(vHistory)
	}

	args := adamKernArg{
		Params:        params.(*Tensor).ptr,
		Gradients:     gradient.(*Tensor).ptr,
		SHistory:      sHistory.(*Tensor).ptr,
		VHistory:      vHistory.(*Tensor).ptr,
		SmoothFactor1: float32(smoothFactor1),
		SmoothFactor2: float32(smoothFactor2),
		LearningRate:  float32(learningRate),
		N:             int32(params.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.adamKernel,
		[3]uint32{uint32(params.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args)

	if o.verification {
		o.cpuOperator.Adam(cpuParams, cpuGradient, cpuVHistory, cpuSHistory, smoothFactor1, smoothFactor2, learningRate)

		o.tensorMustMatch(vHistory, cpuVHistory)
		o.tensorMustMatch(sHistory, cpuSHistory)
		o.tensorMustMatch(params, cpuParams)
	}
}

type reluForwardKernelArgs struct {
	In, Out                   driver.GPUPtr
	Count, Padding            int32
	OffsetX, OffsetY, OffsetZ int64
}

//ReluForward Implementation
func (o *GPUOperator) ReluForward(
	in tensor.Tensor) tensor.Tensor {
	out := o.Create(in.Size()).(*Tensor)
	args := reluForwardKernelArgs{
		In:    in.(*Tensor).ptr,
		Out:   out.ptr,
		Count: int32(in.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.reluForwardKernel,
		[3]uint32{uint32(in.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args)

	if o.verification {
		cpuIn := o.gpuTensorToCPUTensor(in)
		cpuOut := o.cpuOperator.ReluForward(cpuIn)
		o.tensorMustMatch(out, cpuOut)
		fmt.Println("ReluForward verified.")
	}

	return out
}

type reluBackwardKernelArgs struct {
	In, Backin, Out           driver.GPUPtr
	Count, Padding            int32
	OffsetX, OffsetY, OffsetZ int64
}

// ReluBackward Implementation
func (o *GPUOperator) ReluBackward(
	forwardIn, backIn tensor.Tensor,
) tensor.Tensor {
	out := o.Create(forwardIn.Size()).(*Tensor)
	args := reluBackwardKernelArgs{
		In:     forwardIn.(*Tensor).ptr,
		Backin: backIn.(*Tensor).ptr,
		Out:    out.ptr,
		Count:  int32(forwardIn.NumElement()),
	}

	o.driver.LaunchKernel(o.ctx, o.reluBackwardKernel,
		[3]uint32{uint32(forwardIn.NumElement()), 1, 1},
		[3]uint16{64, 1, 1},
		&args)

	if o.verification {
		cpuForwardIn := o.gpuTensorToCPUTensor(forwardIn)
		cpuBackIn := o.gpuTensorToCPUTensor(backIn)
		cpuOut := o.cpuOperator.ReluBackward(cpuForwardIn, cpuBackIn)
		o.tensorMustMatch(out, cpuOut)
		fmt.Println("ReluBackward verified.")
	}

	return out
}
