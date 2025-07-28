package util

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MemoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	_                 uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
	_                 uint32
}

type ProcessEntry32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")

	CreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	Process32First           = kernel32.NewProc("Process32FirstW")
	Process32Next            = kernel32.NewProc("Process32NextW")
)

func GetProcPID(processName string) (uint32, error) {
	snapshot, _, _ := CreateToolhelp32Snapshot.Call(0x00000002, 0)
	if snapshot == 0 {
		return 0, fmt.Errorf("cant create snapshot")
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var pe ProcessEntry32
	pe.dwSize = uint32(unsafe.Sizeof(pe))

	ret, _, _ := Process32First.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	for ret != 0 {
		name := syscall.UTF16ToString(pe.szExeFile[:])
		if name == processName {
			return pe.th32ProcessID, nil
		}
		ret, _, _ = Process32Next.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	}
	return 0, fmt.Errorf("no process: %s", processName)
}

func GetProcPath(pid uint32) (string, error) {
	hProcess, err := syscall.OpenProcess(0x1000, false, pid) //0x1000: PROCESS_QUERY_LIMITED_INFORMATION
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(hProcess)

	bufferSize := syscall.MAX_PATH
	buffer := make([]uint16, bufferSize)
	ret, _, _ := procQueryFullProcessImageNameW.Call(
		uintptr(hProcess),
		0,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&bufferSize)))
	if ret == 0 {
		return "", fmt.Errorf("cant query process path")
	}
	return syscall.UTF16ToString(buffer), nil
}
