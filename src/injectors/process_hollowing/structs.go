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
	DataDirs                [31]uint32 // The extra 4 bytes are undocumented (afaik) and doesn't show in CFF, other than as a gap
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
