#!/bin/bash


function package() {
  local os=$1
  local arch=$2
  local format=$3

  echo "building for ${os} on ${arch}, packaging as ${format}"

  local dir="mp-compiler-${os}-${arch}"
  mkdir ${dir}

  GOOS=${os} go build cmd/compiler/mp-compiler.go
  if [[ "${os}" == "windows" ]]; then
    binary=mp-compiler.exe
  else
    binary=mp-compiler
  fi
  mv ${binary} ${dir} 
  cp LICENSE ${dir}/LICENSE
  if [[ "${format}" == "tar" ]]; then
    tar -czf mp-compiler-${os}-${arch}.tgz ${dir}
  fi
  if [[ "${format}" == "zip" ]]; then
    zip mp-compiler-${os}-${arch}.zip ${dir}/*
  fi

  rm -rf ${dir}
} 

package linux amd64 tar
package darwin amd64 tar
package windows amd64 zip
