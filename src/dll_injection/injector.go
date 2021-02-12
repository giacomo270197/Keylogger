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
		target   = "notepad.exe"
		targetID uint32
	)

	// Grant the process SeDebugPrivilege
	setSeDebugPrivilege()

	// Check to find the process we are interested in
	processesSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
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
	fmt.Println(targetID)

}
