# Go / Object-C Static Library

- [Go Static Library Building](Go-Static.md)
- [Xcode Static Library Integration](Xcode.md)

**`LIPTOOL`**

You can also use liptool to produce static `.a` archive file

```sh
libtool -static i386/libAwesome.a armv7/libAwesome.a -o fat/libAwesome.a
```

**`LIPO`**

Lipo is to combine object files for different platforms. This might be for `.dylib` as well.

<http://stackoverflow.com/questions/15939877/creating-fat-files-libtool-vs-lipo-should-i-prefer-lipo>

> If you run file on the output of the two commands, you'll see that both are “Mach­O universal binary with 2 architectures”, but the lipo ­created file encapsulates two Mach­O object files while the libtool ­ created file encapsulates two “current ar archive random library” files. I don't think the lipo ­created file will work as a static library file. Furthermore, if you try to put more armv7 and i386 object files into the library, the lipo command will fail (because it can't put multiple object files with the same architecture into its output), while the libtool command will succeed. 

```sh
lipo -create i386/libAwesome.a armv7/libAwesome.a -o fat/libAwesome.a
```

> Reference

- [Object File / Symbol Table Format Specification](OBJSPEC.PDF)
- [UNIX tools for exploring object files](au-unixtools-pdf.pdf)
- [C Archive Util](C_Archive_Files_Tips.pdf)
- [Static and Dynamic Libraries](Static_and_Dynamic_Libraries.pdf)
- [Binary Stripping in a Nutshell](Binary_Stripping_in_a_Nutshell.pdf)
- [Statically compiled Go programs, always, even with cgo, using musl](Statically_compiled_Go_with_musl)