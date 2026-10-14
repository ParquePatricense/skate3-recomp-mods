//go:build windows

package main

// Warm-up: while the modder is open, the recomp's own files (exe, DLLs,
// shader cache) and then the game's data files are read once in
// the background, at the lowest disk priority, so Windows already has them
// in memory when you hit Drop in. On a cold PC (just turned on, or a hard
// disk) that is the slowest part of the load. If they are already cached it
// costs nothing.

import (
	"io"
	"os"
	"runtime"
	"sync"
	"unsafe"
)

var (
	warmOnce              sync.Once
	pSetThreadPriority          = kernel32.NewProc("SetThreadPriority")
	pGetCurrentThread           = kernel32.NewProc("GetCurrentThread")
	pGlobalMemoryStatusEx       = kernel32.NewProc("GlobalMemoryStatusEx")
	warmBudget            int64 = 1536 << 20
)

type memoryStatusEx struct {
	Length, MemoryLoad                       uint32
	TotalPhys, AvailPhys                     uint64
	TotalPageFile, AvailPageFile             uint64
	TotalVirtual, AvailVirtual, AvailExtVirt uint64
}

func warmGameFiles(root string) {
	warmOnce.Do(func() { go warmRun(root) })
}

func warmRun(root string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	th, _, _ := pGetCurrentThread.Call()
	pSetThreadPriority.Call(th, 0x00010000) // THREAD_MODE_BACKGROUND_BEGIN: low I/O priority
	ms := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	budget := warmBudget
	if r, _, _ := pGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms))); r != 0 {
		if q := int64(ms.AvailPhys / 4); q < budget {
			budget = q // never more than a quarter of the free memory
		}
	}
	list := warmList(root, os.Getenv("APPDATA"))
	buf := make([]byte, 1<<20)
	for _, x := range list {
		if x.n > budget {
			continue
		}
		budget -= x.n
		if h, err := os.Open(x.p); err == nil {
			io.CopyBuffer(io.Discard, struct{ io.Reader }{h}, buf)
			h.Close()
		}
	}
}
