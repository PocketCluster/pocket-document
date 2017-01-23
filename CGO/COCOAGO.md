# Cocoa <-> GO

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

  ```sh
# Compile and produce object files
CGO_ENABLED=1 CC=clang go build -ldflags '-tmpdir ./tmp -linkmode external' ./...
```

3. Generate header file

  ```sh
# Generate _cgo_export.h and copy into source folder
go tool cgo -objdir ./tmp main.go
```

4. Combine `.o` files to an `.a` archive file.

  ```sh
# Combine the object files into a static library
ar rcs output.a ./tmp/*o  
```

5. Check if all the symbols are included

  ```sh
nm object_file.o
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
- How to build a dylib from several .o in Mac OS X using gcc

  ```
  g++ -dynamiclib -undefined suppress -flat_namespace *.o -o something.dylib
  ```

