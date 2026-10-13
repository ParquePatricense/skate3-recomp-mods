#!/bin/sh
set -e
cd "$(dirname "$0")/.."
mkdir -p out/test
CFLAGS="--target=x86_64-pc-windows-msvc -O2 -ffreestanding -fno-builtin -fno-stack-protector -fms-extensions -Wall -Wno-unused-function -mno-stack-arg-probe -nostdlib -Isrc -DS3D_TEST"
for f in src/*.c; do clang $CFLAGS -c "$f" -o "out/test/$(basename "$f" .c).o"; done
lld-link /nologo /dll /entry:DllMain /nodefaultlib /machine:x64 /out:out/test/xinput9_1_0.dll out/test/*.o lib/kernel32.lib
printf 'LIBRARY KERNEL32.dll\nEXPORTS\nLoadLibraryExW\nGetStdHandle\nExitProcess\nWriteFile\nLoadLibraryW\nGetProcAddress\nVirtualAlloc\n' > out/test/k32h.def
llvm-dlltool -m i386:x86-64 -d out/test/k32h.def -l out/test/k32h.lib
clang $CFLAGS -c test/harness.c -o out/test/harness.o
lld-link /nologo /entry:mainCRTStartup /nodefaultlib /subsystem:console /machine:x64 /out:out/test/harness.exe out/test/harness.o out/test/core.o lib/kernel32.lib
