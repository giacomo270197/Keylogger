// This entrypoint implementation comes entirely (except for the calc part) from  https://gist.github.com/NaniteFactory/7a82b68e822b7d2de44037d6e7511734

package main

import "C"

import (
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

//export PopCalc
func PopCalc() {
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

// OnProcessAttach is an async callback (hook).
//export OnProcessAttach
func OnProcessAttach(
	hinstDLL unsafe.Pointer, // handle to DLL module
	fdwReason uint32, // reason for calling function
	lpReserved unsafe.Pointer, // reserved
) {
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

func main() {
	// nothing really. xD
}
