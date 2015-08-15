#!/bin/bash

set -e

echo Clone libvterm from neovim repo
git clone https://github.com/neovim/libvterm.git

echo Build libvterm
cp libvterm-makefile libvterm/Makefile
pushd libvterm
make
popd

echo Copy libvterm.a
cp libvterm/libvterm.a .

echo Copy includes
cp libvterm/include/*.h .

echo Cleanup
rm -rf libvterm

echo Done
