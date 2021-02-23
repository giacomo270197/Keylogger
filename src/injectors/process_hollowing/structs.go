package main

// Structs definiton for undocumented data structures taken from https://github.com/x64dbg/x64dbg/blob/development/src/dbg/ntdll/ntdll.h
// and https://github.com/winlabs/gowin32/blob/54ddf04f16e612e71c3ebf6cd35da69b652fbbc7/wrappers/winternl.go

// PROCESS_BASIC_INFORMATION (winternl.h)
type PROCESS_BASIC_INFORMATION struct {
	Reserved1       uintptr
	PebBaseAddress  uintptr
	Reserved2       [2]uintptr
	UniqueProcessID uintptr
	Reserved3       uintptr
}

// PEB structure (winternl.h)
type PEB struct {
	InheritedAddressSpace    byte
	ReadImageFileExecOptions byte
	BeingDebugged            byte
	BitField                 byte
	padding                  [4]byte // WinDBG says there is a 4-bytes padding here when running in 64 bits mode
	Mutant                   uintptr
	ImageBaseAddress         uintptr
	Ldr                      uintptr
	ProcessParameters        uintptr
	Reserved4                [104]byte
	Reserved5                [52]uintptr
	PostProcessInitRoutine   uintptr
	Reserved6                [128]byte
	Reserved7                [1]uintptr
	SessionID                uint32
}

// This struct are custom built for this purpose, only mapping needed fields.
// Built checking CFF Explorer
// IMAGE_DOS_HEADERS struct
type IMAGE_DOS_HEADERS struct {
	NotNeeded [60]byte
	E_lfanew  uint32
}

// IMAGE_NT_HEADERS struct (winnt.h)
type IMAGE_NT_HEADERS struct {
	Signature      uint32
	FileHeader     IMAGE_FILE_HEADER
	OptionalHeader IMAGE_OPTIONAL_HEADER
}

// IMAGE_FILE_HEADER struct (winnt.h)
type IMAGE_FILE_HEADER struct {
	Machine               uint16
	NumberOfSections      uint16
	TimeDateStamp         uint32
	PointerToSymbol       uint32
	NumberOfSymbols       uint32
	SizeofOptionalHeaders uint16
	Characteristics       uint16 // No way this is spelled right
}

// Mapping everything was actually quicker than doing the math of how many bytes I need to get to my target.
// Built checking CFF Explorer
// IMAGE_OPTIONAL_HEADER struct (winnt.h)
type IMAGE_OPTIONAL_HEADER struct {
	Magic                   uint16
	MajorLinkerVersion      byte
	MinorLinkerVersion      byte
	SizeOfCode              uint32
	SideOfInitializedData   uint32
	SideOfUninitializedData uint32
	AddressOfEntryPoint     uint32
	BaseOfCode              uint32
	ImageBase               uint64
	SectionAlignment        uint32
	FileAlignment           uint32
	MajorOS                 uint16
	MinorOS                 uint16
	MajorImageVersion       uint16
	MinorImageVersion       uint16
	MajorSubVersion         uint16
	MinorSubVersion         uint16
	Win32Version            uint32
	SizeOfImage             uint32
	SizeOfHeaders           uint32
	Checksun                uint32
	Subsystem               uint16
	DllChars                uint16
	SizeOfStackReserve      uint64
	SizeOfStackCommit       uint64
	SizeOfHeapReserve       uint64
	SizeOfHeapCommit        uint64
	LoaderFlags             uint32
	NumberOfRva             uint32
	// The data direcotries are a different struct, but it can work here
	DataDirsLow             [10]uint32
	RelocationDirectoryRVA  uint32
	RelocationDirectorySize uint32
	DataDirsHigh            [18]uint32
	Padding                 [8]byte // The extra 8 bytes are undocumented (afaik) and doesn't show in CFF, other than as a gap
}

// IMAGE_SECTION_HEADER struct (winnt.h)
// PhysicalAddress is in the docs, but does not seem to appear in CFF
type IMAGE_SECTION_HEADER struct {
	Name [8]byte
	//PhysicalAddress      uint32
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

type CONTEXT struct {
	P1Home       uint64
	P2Home       uint64
	P3Home       uint64
	P4Home       uint64
	P5Home       uint64
	P6Home       uint64
	ContextFlags uint32
	MxCsr        uint32
	SegCs        uint16
	SegDs        uint16
	SegEs        uint16
	SegFs        uint16
	SegGs        uint16
	SegSs        uint16
	EFlags       uint32
	Dr0          uint64
	Dr1          uint64
	Dr2          uint64
	Dr3          uint64
	Dr6          uint64
	Dr7          uint64
	Rax          uint64
	Rcx          uint64
	Rdx          uint64
	Rbx          uint64
	Rsp          uint64
	Rbp          uint64
	Rsi          uint64
	Rdi          uint64
	R8           uint64
	R9           uint64
	R10          uint64
	R11          uint64
	R12          uint64
	R13          uint64
	R14          uint64
	R15          uint64
	Rip          uint64
	// I do need the struct to contain the entire context, but I don't care that Go understands what's in it
	Union          [512]byte // Got to this by subtraction known fileds size from total CONTEXT size
	VectorRegister [52]uint64
	OtherQuads     [6]uint64
}

// IMAGE_DATA_DIRECTORY structure (winnt.h)
type IMAGE_DATA_DIRECTORY struct {
	VirtualAddress uint32
	Size           uint32
}

type BASE_RELOCATION_BLOCK struct {
	PageRVA   uint32
	BlockSize uint32
}

// The actual entries are 4-bits and 12-bits. Very inconveninet since Go can address 1 byte at minimum
type BASE_RELOCATION_ENTRY struct {
	Type   uint8
	Offset uint16
}

// Need to convert a uint16 to a relocation entry
func GetBaseRelocationEntry(input uint16) BASE_RELOCATION_ENTRY {
	var entry BASE_RELOCATION_ENTRY
	var offset int
	var entryType int
	for i := 0; i < 16; i++ {
		remainder := input % 2
		input /= 2
		if i < 12 {
			offset += (i * int(remainder))
		} else {
			entryType += (i * int(remainder))
		}

	}
	entry.Type = uint8(entryType)
	entry.Offset = uint16(offset)
	return entry
}
