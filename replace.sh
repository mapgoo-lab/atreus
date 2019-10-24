#!/bin/sh

sed -i "s/github.com\/bilibili\/atreus/github.com\/mapgoo-lab\/atreus/g" `grep github.com/bilibili/atreus -rl ./`

if [ -d tool/atreus ]; then
    mv tool/atreus tool/atreus
fi

if [ -d tool/atreus-gen-bts ]; then
    mv tool/atreus-gen-bts tool/atreus-gen-bts 
fi

if [ -d tool/atreus-gen-mc ]; then
    mv tool/atreus-gen-mc tool/atreus-gen-mc
fi

if [ -d tool/atreus-protoc ]; then
    mv tool/atreus-protoc tool/atreus-protoc
fi

sed -i "s/atreus/atreus/g" `grep atreus -rl ./`

sed -i "s/Atreus/Atreus/g" `grep Atreus -rl ./`