#!/bin/sh
# Cross-compile skate3-direct (xinput9_1_0.dll) for Windows x64, no CRT.
set -e
cd "$(dirname "$0")"
mkdir -p out
[ -f lib/kernel32.lib ] || llvm-dlltool -m i386:x86-64 -d lib/kernel32.def -l lib/kernel32.lib
CFLAGS="--target=x86_64-pc-windows-msvc -O2 -ffreestanding -fno-builtin -fno-stack-protector -fno-exceptions -fms-extensions -Wall -Wno-unused-function -mno-stack-arg-probe -nostdlib -Isrc"
for f in src/*.c; do
  clang $CFLAGS -c "$f" -o "out/$(basename "$f" .c).o"
done
lld-link /nologo /dll /entry:DllMain /nodefaultlib /subsystem:windows /machine:x64 \
  /out:out/xinput9_1_0.dll out/*.o lib/kernel32.lib /opt:ref /opt:icf
cp out/xinput9_1_0.dll ../go/payload/xinput9_1_0.dll
ls -la out/xinput9_1_0.dll
