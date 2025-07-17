package services

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unsafe"
)

// PE文件结构常量
const (
	IMAGE_DOS_SIGNATURE      = 0x5A4D     // MZ
	IMAGE_NT_SIGNATURE       = 0x00004550 // PE00
	IMAGE_FILE_MACHINE_I386  = 0x014c
	IMAGE_FILE_MACHINE_AMD64 = 0x8664
)

// PE文件头结构
type DOSHeader struct {
	Signature          uint16
	BytesOnLastPage    uint16
	PagesInFile        uint16
	Relocations        uint16
	SizeOfHeader       uint16
	MinExtraParagraphs uint16
	MaxExtraParagraphs uint16
	InitialSS          uint16
	InitialSP          uint16
	Checksum           uint16
	InitialIP          uint16
	InitialCS          uint16
	RelocTableOffset   uint16
	OverlayNumber      uint16
	Reserved1          [4]uint16
	OEMIdentifier      uint16
	OEMInformation     uint16
	Reserved2          [10]uint16
	NewExeHeaderOffset uint32
}

type NTHeaders struct {
	Signature        uint32
	FileHeader       FileHeader
	OptionalHeader32 *OptionalHeader32
	OptionalHeader64 *OptionalHeader64
}

type FileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

type OptionalHeader32 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	BaseOfData                  uint32
	ImageBase                   uint32
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint32
	SizeOfStackCommit           uint32
	SizeOfHeapReserve           uint32
	SizeOfHeapCommit            uint32
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
}

type OptionalHeader64 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	ImageBase                   uint64
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
}

type SectionHeader struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

type ImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

type ExportDirectory struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}

// PEAnalyzer PE文件分析器
type PEAnalyzer struct {
}

// NewPEAnalyzer 创建PE分析器
func NewPEAnalyzer() *PEAnalyzer {
	return &PEAnalyzer{}
}

// AnalyzeFile 分析PE文件
func (pe *PEAnalyzer) AnalyzeFile(filePath string) (*PEAnalysisResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return pe.AnalyzeReader(file)
}

// AnalyzeReader 分析PE文件流
func (pe *PEAnalyzer) AnalyzeReader(reader io.ReadSeeker) (*PEAnalysisResult, error) {
	result := &PEAnalysisResult{
		Imports:        []ImportFunction{},
		Exports:        []ExportFunction{},
		Sections:       []SectionInfo{},
		SuspiciousAPIs: []string{},
	}

	// 读取DOS头
	dosHeader, err := pe.readDOSHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("读取DOS头失败: %v", err)
	}

	// 验证DOS签名
	if dosHeader.Signature != IMAGE_DOS_SIGNATURE {
		return nil, fmt.Errorf("无效的DOS签名: 0x%x", dosHeader.Signature)
	}

	// 读取NT头
	ntHeaders, err := pe.readNTHeaders(reader, int64(dosHeader.NewExeHeaderOffset))
	if err != nil {
		return nil, fmt.Errorf("读取NT头失败: %v", err)
	}

	// 验证PE签名
	if ntHeaders.Signature != IMAGE_NT_SIGNATURE {
		return nil, fmt.Errorf("无效的PE签名: 0x%x", ntHeaders.Signature)
	}

	// 确定架构
	switch ntHeaders.FileHeader.Machine {
	case IMAGE_FILE_MACHINE_I386:
		result.Architecture = "x86"
	case IMAGE_FILE_MACHINE_AMD64:
		result.Architecture = "x64"
	default:
		result.Architecture = fmt.Sprintf("unknown(0x%x)", ntHeaders.FileHeader.Machine)
	}

	// 读取节表
	sections, err := pe.readSectionHeaders(reader, ntHeaders)
	if err != nil {
		return nil, fmt.Errorf("读取节表失败: %v", err)
	}

	// 分析节信息
	for _, section := range sections {
		sectionInfo := pe.analyzeSectionInfo(reader, &section)
		result.Sections = append(result.Sections, sectionInfo)
	}

	// 分析导入表
	imports, err := pe.analyzeImports(reader, ntHeaders, sections)
	if err == nil {
		result.Imports = imports
	}

	// 分析导出表
	exports, err := pe.analyzeExports(reader, ntHeaders, sections)
	if err == nil {
		result.Exports = exports
	}

	// 检测加壳
	result.PackerDetected = pe.detectPacker(result.Sections)

	return result, nil
}

// readDOSHeader 读取DOS头
func (pe *PEAnalyzer) readDOSHeader(reader io.ReadSeeker) (*DOSHeader, error) {
	reader.Seek(0, 0)

	var dosHeader DOSHeader
	err := binary.Read(reader, binary.LittleEndian, &dosHeader)
	if err != nil {
		return nil, err
	}

	return &dosHeader, nil
}

// readNTHeaders 读取NT头
func (pe *PEAnalyzer) readNTHeaders(reader io.ReadSeeker, offset int64) (*NTHeaders, error) {
	reader.Seek(offset, 0)

	var ntHeaders NTHeaders

	// 读取PE签名
	err := binary.Read(reader, binary.LittleEndian, &ntHeaders.Signature)
	if err != nil {
		return nil, err
	}

	// 读取文件头
	err = binary.Read(reader, binary.LittleEndian, &ntHeaders.FileHeader)
	if err != nil {
		return nil, err
	}

	// 根据架构读取可选头
	if ntHeaders.FileHeader.Machine == IMAGE_FILE_MACHINE_I386 {
		var optHeader OptionalHeader32
		err = binary.Read(reader, binary.LittleEndian, &optHeader)
		if err != nil {
			return nil, err
		}
		ntHeaders.OptionalHeader32 = &optHeader
	} else if ntHeaders.FileHeader.Machine == IMAGE_FILE_MACHINE_AMD64 {
		var optHeader OptionalHeader64
		err = binary.Read(reader, binary.LittleEndian, &optHeader)
		if err != nil {
			return nil, err
		}
		ntHeaders.OptionalHeader64 = &optHeader
	}

	return &ntHeaders, nil
}

// readSectionHeaders 读取节表
func (pe *PEAnalyzer) readSectionHeaders(reader io.ReadSeeker, ntHeaders *NTHeaders) ([]SectionHeader, error) {
	sections := make([]SectionHeader, ntHeaders.FileHeader.NumberOfSections)

	for i := 0; i < int(ntHeaders.FileHeader.NumberOfSections); i++ {
		err := binary.Read(reader, binary.LittleEndian, &sections[i])
		if err != nil {
			return nil, err
		}
	}

	return sections, nil
}

// analyzeSectionInfo 分析节信息
func (pe *PEAnalyzer) analyzeSectionInfo(reader io.ReadSeeker, section *SectionHeader) SectionInfo {
	name := strings.TrimRight(string(section.Name[:]), "\x00")

	// 计算熵值
	entropy := pe.calculateSectionEntropy(reader, section)

	// 检查节特性
	isExecutable := (section.Characteristics & 0x20000000) != 0 // IMAGE_SCN_MEM_EXECUTE
	isWritable := (section.Characteristics & 0x80000000) != 0   // IMAGE_SCN_MEM_WRITE

	return SectionInfo{
		Name:           name,
		VirtualAddress: section.VirtualAddress,
		VirtualSize:    section.VirtualSize,
		RawSize:        section.SizeOfRawData,
		Entropy:        entropy,
		IsExecutable:   isExecutable,
		IsWritable:     isWritable,
	}
}

// calculateSectionEntropy 计算节的熵值
func (pe *PEAnalyzer) calculateSectionEntropy(reader io.ReadSeeker, section *SectionHeader) float64 {
	if section.SizeOfRawData == 0 {
		return 0.0
	}

	// 读取节数据
	reader.Seek(int64(section.PointerToRawData), 0)
	data := make([]byte, section.SizeOfRawData)
	n, err := reader.Read(data)
	if err != nil || n == 0 {
		return 0.0
	}

	// 计算字节频率
	freq := make(map[byte]int)
	for _, b := range data[:n] {
		freq[b]++
	}

	// 计算熵值
	entropy := 0.0
	length := float64(n)

	for _, count := range freq {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// analyzeImports 分析导入表
func (pe *PEAnalyzer) analyzeImports(reader io.ReadSeeker, ntHeaders *NTHeaders, sections []SectionHeader) ([]ImportFunction, error) {
	var imports []ImportFunction

	// 获取导入表RVA
	var importRVA uint32
	if ntHeaders.OptionalHeader32 != nil {
		// 32位PE文件的导入表在数据目录的第2个条目
		reader.Seek(int64(unsafe.Offsetof(ntHeaders.OptionalHeader32.NumberOfRvaAndSizes))+4+8, 1) // 跳到导入表条目
		binary.Read(reader, binary.LittleEndian, &importRVA)
	} else if ntHeaders.OptionalHeader64 != nil {
		reader.Seek(int64(unsafe.Offsetof(ntHeaders.OptionalHeader64.NumberOfRvaAndSizes))+4+8, 1)
		binary.Read(reader, binary.LittleEndian, &importRVA)
	}

	if importRVA == 0 {
		return imports, nil
	}

	// 将RVA转换为文件偏移
	importOffset := pe.rvaToFileOffset(importRVA, sections)
	if importOffset == 0 {
		return imports, fmt.Errorf("无法找到导入表")
	}

	reader.Seek(int64(importOffset), 0)

	// 读取导入描述符
	for {
		var importDesc ImportDescriptor
		err := binary.Read(reader, binary.LittleEndian, &importDesc)
		if err != nil {
			break
		}

		// 空描述符表示结束
		if importDesc.Name == 0 {
			break
		}

		// 读取DLL名称
		dllNameOffset := pe.rvaToFileOffset(importDesc.Name, sections)
		if dllNameOffset == 0 {
			continue
		}

		dllName := pe.readStringAtOffset(reader, int64(dllNameOffset))

		// 读取导入函数
		thunkRVA := importDesc.OriginalFirstThunk
		if thunkRVA == 0 {
			thunkRVA = importDesc.FirstThunk
		}

		thunkOffset := pe.rvaToFileOffset(thunkRVA, sections)
		if thunkOffset == 0 {
			continue
		}

		funcImports := pe.readImportFunctions(reader, int64(thunkOffset), dllName, sections, ntHeaders.FileHeader.Machine == IMAGE_FILE_MACHINE_AMD64)
		imports = append(imports, funcImports...)
	}

	return imports, nil
}

// analyzeExports 分析导出表
func (pe *PEAnalyzer) analyzeExports(reader io.ReadSeeker, ntHeaders *NTHeaders, sections []SectionHeader) ([]ExportFunction, error) {
	var exports []ExportFunction

	// 获取导出表RVA
	var exportRVA uint32
	if ntHeaders.OptionalHeader32 != nil {
		reader.Seek(int64(unsafe.Offsetof(ntHeaders.OptionalHeader32.NumberOfRvaAndSizes))+4, 1) // 跳到导出表条目
		binary.Read(reader, binary.LittleEndian, &exportRVA)
	} else if ntHeaders.OptionalHeader64 != nil {
		reader.Seek(int64(unsafe.Offsetof(ntHeaders.OptionalHeader64.NumberOfRvaAndSizes))+4, 1)
		binary.Read(reader, binary.LittleEndian, &exportRVA)
	}

	if exportRVA == 0 {
		return exports, nil
	}

	// 将RVA转换为文件偏移
	exportOffset := pe.rvaToFileOffset(exportRVA, sections)
	if exportOffset == 0 {
		return exports, fmt.Errorf("无法找到导出表")
	}

	reader.Seek(int64(exportOffset), 0)

	// 读取导出目录
	var exportDir ExportDirectory
	err := binary.Read(reader, binary.LittleEndian, &exportDir)
	if err != nil {
		return exports, err
	}

	// 读取函数名称数组
	namesOffset := pe.rvaToFileOffset(exportDir.AddressOfNames, sections)
	if namesOffset == 0 {
		return exports, nil
	}

	reader.Seek(int64(namesOffset), 0)
	nameRVAs := make([]uint32, exportDir.NumberOfNames)
	for i := uint32(0); i < exportDir.NumberOfNames; i++ {
		binary.Read(reader, binary.LittleEndian, &nameRVAs[i])
	}

	// 读取序号数组
	ordinalsOffset := pe.rvaToFileOffset(exportDir.AddressOfNameOrdinals, sections)
	if ordinalsOffset == 0 {
		return exports, nil
	}

	reader.Seek(int64(ordinalsOffset), 0)
	ordinals := make([]uint16, exportDir.NumberOfNames)
	for i := uint32(0); i < exportDir.NumberOfNames; i++ {
		binary.Read(reader, binary.LittleEndian, &ordinals[i])
	}

	// 读取函数地址数组
	functionsOffset := pe.rvaToFileOffset(exportDir.AddressOfFunctions, sections)
	if functionsOffset == 0 {
		return exports, nil
	}

	reader.Seek(int64(functionsOffset), 0)
	functionRVAs := make([]uint32, exportDir.NumberOfFunctions)
	for i := uint32(0); i < exportDir.NumberOfFunctions; i++ {
		binary.Read(reader, binary.LittleEndian, &functionRVAs[i])
	}

	// 构建导出函数列表
	for i := uint32(0); i < exportDir.NumberOfNames; i++ {
		nameOffset := pe.rvaToFileOffset(nameRVAs[i], sections)
		if nameOffset == 0 {
			continue
		}

		functionName := pe.readStringAtOffset(reader, int64(nameOffset))
		ordinal := ordinals[i]

		if int(ordinal) < len(functionRVAs) {
			exports = append(exports, ExportFunction{
				FunctionName: functionName,
				Offset:       functionRVAs[ordinal],
				Ordinal:      ordinal + uint16(exportDir.Base),
			})
		}
	}

	return exports, nil
}

// rvaToFileOffset 将RVA转换为文件偏移
func (pe *PEAnalyzer) rvaToFileOffset(rva uint32, sections []SectionHeader) uint32 {
	for _, section := range sections {
		if rva >= section.VirtualAddress && rva < section.VirtualAddress+section.VirtualSize {
			return rva - section.VirtualAddress + section.PointerToRawData
		}
	}
	return 0
}

// readStringAtOffset 在指定偏移读取字符串
func (pe *PEAnalyzer) readStringAtOffset(reader io.ReadSeeker, offset int64) string {
	reader.Seek(offset, 0)

	var result []byte
	buf := make([]byte, 1)

	for {
		n, err := reader.Read(buf)
		if err != nil || n == 0 || buf[0] == 0 {
			break
		}
		result = append(result, buf[0])

		// 防止无限循环
		if len(result) > 256 {
			break
		}
	}

	return string(result)
}

// readImportFunctions 读取导入函数
func (pe *PEAnalyzer) readImportFunctions(reader io.ReadSeeker, thunkOffset int64, dllName string, sections []SectionHeader, is64bit bool) []ImportFunction {
	var imports []ImportFunction

	reader.Seek(thunkOffset, 0)

	for {
		var thunk uint64
		var err error

		if is64bit {
			err = binary.Read(reader, binary.LittleEndian, &thunk)
		} else {
			var thunk32 uint32
			err = binary.Read(reader, binary.LittleEndian, &thunk32)
			thunk = uint64(thunk32)
		}

		if err != nil || thunk == 0 {
			break
		}

		// 检查是否为序号导入
		var functionName string
		if (is64bit && (thunk&0x8000000000000000) != 0) || (!is64bit && (thunk&0x80000000) != 0) {
			// 序号导入
			ordinal := thunk & 0xFFFF
			functionName = fmt.Sprintf("Ordinal_%d", ordinal)
		} else {
			// 名称导入
			nameRVA := uint32(thunk & 0xFFFFFFFF)
			nameOffset := pe.rvaToFileOffset(nameRVA, sections)
			if nameOffset != 0 {
				reader.Seek(int64(nameOffset)+2, 0) // 跳过hint
				functionName = pe.readStringAtOffset(reader, int64(nameOffset)+2)
				reader.Seek(thunkOffset+int64(len(imports)+1)*8, 0) // 返回thunk位置
			}
		}

		if functionName != "" {
			imports = append(imports, ImportFunction{
				DLLName:      dllName,
				FunctionName: functionName,
				Offset:       uint32(thunk),
			})
		}
	}

	return imports
}

// detectPacker 检测加壳
func (pe *PEAnalyzer) detectPacker(sections []SectionInfo) bool {
	// 检查高熵值节
	for _, section := range sections {
		if section.Entropy > 7.5 && section.IsExecutable {
			return true
		}
	}

	// 检查可疑节名
	suspiciousSectionNames := []string{
		"UPX", "upx", ".packed", ".compress", ".aspack", ".rlpack",
		".petite", ".mew", ".nsp", ".yoda", ".wwpack", ".svkp",
	}

	for _, section := range sections {
		for _, suspicious := range suspiciousSectionNames {
			if strings.Contains(strings.ToLower(section.Name), strings.ToLower(suspicious)) {
				return true
			}
		}
	}

	return false
}
