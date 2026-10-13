// Minimal Win32 declarations for a freestanding (no CRT) DLL.
#pragma once

typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;
typedef signed int i32;
typedef signed long long i64;
typedef unsigned long long usize;
typedef unsigned short wchar;
typedef int BOOL;
typedef unsigned long DWORD;
typedef void* HANDLE;
typedef void* HMODULE;

#define WINAPI __stdcall
#define DLLIMPORT __declspec(dllimport)
#define INVALID_HANDLE_VALUE ((HANDLE)(i64)-1)
#define TRUE 1
#define FALSE 0
#define NULL ((void*)0)

#define PAGE_EXECUTE_READWRITE 0x40
#define PAGE_READWRITE 0x04
#define MEM_COMMIT 0x1000
#define MEM_RESERVE 0x2000
#define MEM_RELEASE 0x8000
#define GENERIC_READ 0x80000000
#define GENERIC_WRITE 0x40000000
#define FILE_SHARE_READ 1
#define FILE_SHARE_WRITE 2
#define CREATE_ALWAYS 2
#define OPEN_EXISTING 3
#define OPEN_ALWAYS 4
#define FILE_ATTRIBUTE_NORMAL 0x80
#define FILE_END 2
#define INVALID_FILE_ATTRIBUTES 0xFFFFFFFF

typedef struct {
  u16 wYear, wMonth, wDayOfWeek, wDay, wHour, wMinute, wSecond, wMilliseconds;
} SYSTEMTIME;

typedef struct {
  void* BaseAddress;
  void* AllocationBase;
  DWORD AllocationProtect;
  u16 PartitionId;
  usize RegionSize;
  DWORD State;
  DWORD Protect;
  DWORD Type;
} MEMORY_BASIC_INFORMATION;

typedef struct {
  long LowPart;
  long HighPart;
} LARGE_INTEGER_PARTS;
typedef union {
  LARGE_INTEGER_PARTS u;
  i64 QuadPart;
} LARGE_INTEGER;

typedef struct {
  DWORD DebugInfo_lo, DebugInfo_hi;
  i32 LockCount, RecursionCount;
  HANDLE OwningThread, LockSemaphore;
  usize SpinCount;
} CRITICAL_SECTION;

DLLIMPORT HMODULE WINAPI GetModuleHandleW(const wchar* name);
DLLIMPORT DWORD WINAPI GetModuleFileNameW(HMODULE m, wchar* buf, DWORD n);
DLLIMPORT void* WINAPI GetProcAddress(HMODULE m, const char* name);
DLLIMPORT HMODULE WINAPI LoadLibraryW(const wchar* name);
DLLIMPORT u32 WINAPI GetSystemDirectoryW(wchar* buf, u32 n);
DLLIMPORT BOOL WINAPI VirtualProtect(void* a, usize n, DWORD prot, DWORD* old);
DLLIMPORT void* WINAPI VirtualAlloc(void* a, usize n, DWORD type, DWORD prot);
DLLIMPORT BOOL WINAPI VirtualFree(void* a, usize n, DWORD type);
DLLIMPORT usize WINAPI VirtualQuery(const void* a, MEMORY_BASIC_INFORMATION* mbi, usize n);
DLLIMPORT BOOL WINAPI FlushInstructionCache(HANDLE p, const void* a, usize n);
DLLIMPORT HANDLE WINAPI GetCurrentProcess(void);
DLLIMPORT HANDLE WINAPI CreateFileW(const wchar* name, DWORD access, DWORD share, void* sa,
                                    DWORD disp, DWORD flags, HANDLE tmpl);
DLLIMPORT BOOL WINAPI WriteFile(HANDLE h, const void* buf, DWORD n, DWORD* written, void* ov);
DLLIMPORT BOOL WINAPI ReadFile(HANDLE h, void* buf, DWORD n, DWORD* read, void* ov);
DLLIMPORT BOOL WINAPI CloseHandle(HANDLE h);
DLLIMPORT DWORD WINAPI GetFileSize(HANDLE h, DWORD* hi);
DLLIMPORT DWORD WINAPI SetFilePointer(HANDLE h, long lo, long* hi, DWORD method);
DLLIMPORT BOOL WINAPI QueryPerformanceCounter(LARGE_INTEGER* c);
DLLIMPORT BOOL WINAPI QueryPerformanceFrequency(LARGE_INTEGER* f);
DLLIMPORT void WINAPI GetLocalTime(SYSTEMTIME* t);
DLLIMPORT HANDLE WINAPI CreateThread(void* sa, usize stack, DWORD(WINAPI* fn)(void*), void* arg,
                                     DWORD flags, DWORD* tid);
DLLIMPORT void WINAPI Sleep(DWORD ms);
DLLIMPORT u64 WINAPI GetTickCount64(void);
DLLIMPORT DWORD WINAPI GetEnvironmentVariableW(const wchar* name, wchar* buf, DWORD n);
DLLIMPORT BOOL WINAPI DisableThreadLibraryCalls(HMODULE m);
DLLIMPORT DWORD WINAPI GetCurrentThreadId(void);
DLLIMPORT void WINAPI InitializeCriticalSection(CRITICAL_SECTION* cs);
DLLIMPORT void WINAPI EnterCriticalSection(CRITICAL_SECTION* cs);
DLLIMPORT void WINAPI LeaveCriticalSection(CRITICAL_SECTION* cs);
DLLIMPORT DWORD WINAPI GetLastError(void);
DLLIMPORT DWORD WINAPI GetFileAttributesW(const wchar* name);

// IMAGE headers (only what we read)
typedef struct {
  u16 e_magic;
  u16 pad[29];
  i32 e_lfanew;
} IMAGE_DOS_HEADER;
