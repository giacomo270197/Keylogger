package main

import (
	"C"
	"log"

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

func main() {
	//PopCalc() // For compiling to exe and debugging
}
