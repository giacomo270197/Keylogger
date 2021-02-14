package main

import "C"

import (
	"log"
	"unsafe"

	"github.com/nanitefactory/winmb"
	"golang.org/x/sys/windows"
)

//export Test
func Test() {
	winmb.MessageBoxPlain("export Test", "export Test")
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

	//winmb.MessageBoxPlain("Message", "Message")

	err = windows.CreateProcess(nil, cmd, nil, nil, false, 0, nil, nil, startupinfo, outprocinfo)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// nothing really. xD
}
