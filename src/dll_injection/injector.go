package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

func PopCmd() {
	cmd, err := windows.UTF16PtrFromString("C:\\Windows\\System32\\calc.exe")
	if err != nil {
		log.Fatal(err)
	}
	startupinfo := new(windows.StartupInfo)
	startupinfo.ShowWindow = 1
	outprocinfo := new(windows.ProcessInformation)
	outprocinfo.ProcessId = windows.GetCurrentProcessId()
	outprocinfo.ThreadId = windows.GetCurrentThreadId()
	outprocinfo.Process, err = windows.GetCurrentProcess()
	if err != nil {
		log.Fatal(err)
	}
	outprocinfo.Thread, err = windows.GetCurrentThread()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	err = windows.CreateProcess(nil, cmd, nil, nil, false, 0, nil, nil, startupinfo, outprocinfo)
	if err != nil {
		log.Fatal(err)
	}
}

func getProcessName(processID uint32) string {
	var (
		psapiDLL           = windows.NewLazyDLL("psapi.dll")
		GetModuleBaseNameA = psapiDLL.NewProc("GetModuleBaseNameA")
		returnedLength     uint32
	)

	buffer := make([]uint16, 256)
	//fmt.Println("Test1")
	handle, err := windows.OpenProcess(windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION, false, processID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Test2")
	GetModuleBaseNameA.Call(uintptr(handle), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(returnedLength))
	return windows.UTF16PtrToString(&buffer[0])
}

func setSeDebugPrivilege() {
	handle, err := windows.GetCurrentProcess()
	if err != nil {
		return
	}
	oldToken := new(windows.Token)
	err = windows.OpenProcessToken(handle, windows.TOKEN_ADJUST_PRIVILEGES, oldToken)
	if err != nil {
		log.Fatal(err)
	}
	luid := new(windows.LUID)
	seDebugName, _ := windows.UTF16FromString("SeDebugPrivilege")
	err = windows.LookupPrivilegeValue(nil, &seDebugName[0], luid)
	if err != nil {
		log.Fatal(err)
	}
	tokenPriviledges := new(windows.Tokenprivileges)
	tokenPriviledges.PrivilegeCount = 1
	tokenPriviledges.Privileges[0].Luid = *luid
	tokenPriviledges.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
	err = windows.AdjustTokenPrivileges(*oldToken, false, tokenPriviledges, uint32(unsafe.Sizeof(tokenPriviledges)), nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Debug Priviledge granted")
}

func main() {
	var (
		target      = "notepad.exe"
		targetID    uint32
		bytesneeded uint32 = 0
	)
	processes := make([]uint32, 512)

	// Grant the process SeDebugPrivilege

	setSeDebugPrivilege()

	PopCmd()

	return

	// Check to find the process we are interested in

	err := windows.EnumProcesses(processes, &bytesneeded)
	if err != nil {
		log.Fatal(err)
	}

	for _, id := range processes {
		if id > 100 {
			name := getProcessName(id)
			fmt.Println(name)
			if name == target {
				targetID = id
			}
		}
	}

	fmt.Println(targetID)

}
