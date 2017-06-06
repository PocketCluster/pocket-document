# Cocoa <-> GO

### Static GO object file symbol checker

In Linux there are nm, readelf, objdump. you can install 

>  brew install binutils

On OSX, however, we have `otool`, `dwarftool` as well as `nm`.

> References

- <https://superuser.com/questions/206547/how-can-i-install-objdump-on-mac-os-x>
- <http://stackoverflow.com/questions/3286675/readelf-like-tool-for-mac-os-x>

### Static GO archive -> Xcode Binary

1. Create go main like following

  ```go
  package main
  
  /*
  #cgo CFLAGS: -x objective-c
  #cgo LDFLAGS: -Wl,-U,_osxmain -framework Cocoa
  
  extern int osxmain(int argc, const char * argv[]);
  */
  import "C"
  
  func main() {
      // Perhaps the first thing main() function needs to do is initiate OSX main
      C.osxmain(0, nil)
  }
  
  ```
  
  **Remark**  
  a. `LDFLAGS: -Wl,-U,_osxmain` is important to tell `clang` to compile only go sources.  
  b. `#cgo` has following directives: `CFLAGS`, `CPPFLAGS`, `CXXFLAGS`, `FFLAGS` and `LDFLAGS`.  
  
  > Reference

  - [hyperlink - Xcode clang link_ Build Dynamic Framework (or dylib) not embed dependencies](hyperlink - Xcode clang link_ Build Dynamic Framework (or dylib) not embed dependencies - Stack Overflow.pdf)
  - [glandium.org Fun with weak dynamic linking](glandium.org Fun with weak dynamic linking.pdf)

2. Build GO library
  - With `-s` linker option, debug symbol is stripped. Do not use `strip` as it would break go to run
  - With `-w` linker option, `dwarf` symbol is refrained from being generated
  - Default mode seems to be the same as `c-archive` mode as file, and ar commands confirms final output 

  ```sh
  # Compile and produce object files
  CGO_ENABLED=1 CC=clang go build -v -x -ldflags '-v -w -s -tmpdir .tmp -linkmode external' ./...
  ```

  - In Go 1.7.5, there are more build mode to build static library. You can read more from how `gomobile` handles binding.
  - With buildmode `c-archive` or `c-shared`, `_main()` reference in symbol is missing and linker fails
  - With buildmode `c-archive`, `strip -S go.o` fails with following message, which causes `strip -S` in Xcode to quit with 1 in Archving (release). This seems to relate the ability of golang to write an object file.

    > strip: string table not at the end of the file (can't be processed) in file: go.o

  - With buildmode `c-shared` and `-o target_object` option, go can write really nice `.dylib` binary which doesn't cause an issue with `strip` (You should not strip symbols with it though). but `_main()` is missing.<br/>Further, this creates dynamic library which is separated from main executable.<br/>Plus, when you have external function call in your go library, it doesn't seem to be very clear how to link other dynamic library to perform all the linking. (maybe we should build framework like Go mobile did)<br/>**This is postponed**

  - (03/28/2017) Therefore, we'll strip debug symbol from golang, and statically link to executable for now.

    ```
    > go help buildmode
    
    # **This is the mode we go with**
    -buildmode=default
        Listed main packages are built into executables and listed
        non-main packages are built into .a files (the default
        behavior).
  
    # essentially the same as default. It lacks of main().
    -buildmode=c-archive
        Build the listed main package, plus all packages it imports,
        into a C archive file. The only callable symbols will be those
        functions exported using a cgo //export comment. Requires
        exactly one main package to be listed.
    
    # generate shared library (.dylib). It lacks of main()
    -buildmode=c-shared
        Build the listed main packages, plus all packages that they
        import, into C shared libraries. The only callable symbols will
        be those functions exported using a cgo //export comment.
        Non-main packages are ignored.


    # (03/29/2017) There are more to buildmode, but not applicable 

    -buildmode=pie [Not supported in amd64]
        Build the listed main packages and everything they import into
        position independent executables (PIE). Packages not named
        main are ignored.
    
    -buildmode=exe 
        Build the listed main packages and everything they import into
        executables. Packages not named main are ignored.
    
    -buildmode=default
        Listed main packages are built into executables and listed
        non-main packages are built into .a files (the default
        behavior).
    
    -buildmode=archive
        Build the listed non-main packages into .a files. Packages named
        main are ignored.
    
    -buildmode=shared
        Combine all the listed non-main packages into a single shared
        library that will be used when building with the -linkshared
        option. Packages named main are ignored.
    ```


3. Generate header file

  ```sh
  # Generate _cgo_export.h and copy into source folder
  go tool cgo -objdir ./tmp main.go
  ```
  - You cannot seem to generate `_main()` symbol from `go tool cgo`

4. Combine `.o` files to an `.a` archive file.

  ```sh
  # Combine the object files into a static library
  ar rcs output.a ./tmp/*o  
  ```

5. Check if all the symbols are included

  ```sh
  nm object_file.o
  ```

### Build Script

(03/29/2017) Build Script

```sh
#!/bin/bash

# Exit if any command fails
set -e

# Figure out where things are coming from and going to
GOREPO=${GOREPO:-"${HOME}/Workspace/POCKETPKG"}
GOPATH=${GOPATH:-"${GOREPO}:$GOWORKPLACE"}
GO=${GOROOT}/bin/go
GG_BUILD="${PWD}/../../.build"
ARCHIVE="${GG_BUILD}/pc-core.a"
PATH=${PATH:-"$GEM_HOME/ruby/2.0.0/bin:$HOME/.util:$GOROOT/bin:$GOREPO/bin:$GOWORKPLACE/bin:$HOME/.util:$NATIVE_PATH"}

# Clean old directory
if [ -d ${GG_BUILD} ]; then
    rm -rf ${GG_BUILD} && mkdir -p ${GG_BUILD}
fi

echo "Make the temp folders for go objects"
mkdir -p ${GG_BUILD}

echo "Generate _cgo_export.h and copy into source folder"
${GO} tool cgo -objdir ${GG_BUILD} *.go

echo "Compile and produce object files"
# [Default mode] First trial
#CGO_ENABLED=1 CC=clang ${GO} build -ldflags '-tmpdir '${GG_BUILD}' -linkmode external' ./...

# [Default mode] External clang linker
#CGO_ENABLED=1 CC=clang ${GO} build -v -x -ldflags '-v -tmpdir '${GG_BUILD}' -linkmode external -extld clang' ./...

# [Archive mode]
#CGO_ENABLED=1 CC=clang ${GO} build -v -x -buildmode=c-archive -ldflags '-v -tmpdir '${GG_BUILD}' -linkmode external' ./...

# [Shared mode] go.dwarf file
#CGO_ENABLED=1 CC=clang ${GO} build -v -x -buildmode=c-shared -ldflags '-v -tmpdir '${GG_BUILD}' -linkmode external' ./...

# [Archive mode] prevents go.dwarf generated (-w), strip symbol (-s)
#CGO_ENABLED=1 CC=clang ${GO} build -v -x -buildmode=c-archive -ldflags '-v -w -s -tmpdir '${GG_BUILD}' -linkmode external' ./...

# [Default mode] default mode (we need main() function), disable go.dwarf generation (-w), strip symbol (-s)
CGO_ENABLED=1 CC=clang ${GO} build -v -x -ldflags '-v -s -w -tmpdir '${GG_BUILD}' -linkmode external' ./...

echo "Combine the object files into a static library"
ar rcs ${ARCHIVE} ${GG_BUILD}/*.o
mv ${GG_BUILD}/_cgo_export.h ${GG_BUILD}/pc-core.h
rm static*
echo "${ARCHIVE} generated!"
```

> References

- [Mobile Go, part 1/ Calling Go functions from C](Mobile Go, part 1 Calling Go functions from C — Mobile Go — Medium.pdf)
- [Mobile Go, part 2/ Building an iOS app with `go build`](Mobile Go, part 2 Building an iOS app with `go build` — Mobile Go — Medium.pdf)
- [Mobile Go, part 3/ Calling Go from Objective C](Mobile Go, part 3 Calling Go from Objective C — Mobile Go — Medium.pdf)
- [Mobile Go, Part 4/ Calling Objective C from Go](Mobile Go, Part 4 Calling Objective C from Go — Mobile Go — Medium.pdf)


### Dynamic GO archive -> Xcode Binary

In PocketCluster OSX, we will not use Dynamic archive as it opens up an attack vector and could be reverse engineered. Nonetheless, here comes how to build one.

- [Using C Dynamic Libraries In Go Programs](Using C Dynamic Libraries In Go Programs.pdf)
- [Using CGO with Pkg-Config And Custom Dynamic Library Locations](Using CGO with Pkg-Config And Custom Dynamic Library Locations.pdf)
- [`gomobile` ios binding](https://github.com/golang/mobile/blob/226c1c8284dc3ee080c91f7ff62e6431b0b2a74b/cmd/gomobile/bind_iosapp.go#L197-L206)
- How to build a dylib from several .o in Mac OS X using gcc

  ```
  g++ -dynamiclib -undefined suppress -flat_namespace *.o -o something.dylib
  ```
