#include "../src/core.h"
DLLIMPORT void WINAPI ExitProcess(u32 code);
DLLIMPORT HANDLE WINAPI GetStdHandle(DWORD n);
static void out(const char* s) { DWORD w; WriteFile(GetStdHandle((DWORD)-11), s, (DWORD)s_len(s), &w, NULL); }
typedef int (*ft_t)(wchar*);
void mainCRTStartup(void) {
  char buf[128];
  HMODULE rt = LoadLibraryW(L"rexruntime.dll");
  fmt(buf, sizeof buf, "rexruntime %p\n", rt); out(buf);
  HMODULE d = LoadLibraryW(L"xinput9_1_0.dll");
  int r = ((ft_t)GetProcAddress(d, "s3d_flagtest"))(L".\\");
  fmt(buf, sizeof buf, "flagtest -> %d\n", r); out(buf);
  ExitProcess(0);
}
