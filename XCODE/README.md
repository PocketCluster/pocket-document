# Xcode + Golang Integration tips/logs

## Linker

![](img/linker-flag.png)

- No PIE (Position Independent Executable) Linker error

  > ld: warning: PIE disabled. Absolute addressing (perhaps -mdynamic-no-pic) not allowed in code signed PIE, but used in runtime.rodata fromlibstatic-core.a(go.o). To fix this warning, don't compile with -mdynamic-no-pic or link with -Wl,-no_pie

  We'll provide `-Wl,-no_pie` flag in linker to pass for now (03/20/2017)

- Unable to find responder error

  This happens when you're to load static library that has categories, which is not fully loaded.  
  In order to prevent this to happen, apply `-ObjC` and `-all_load` flags to linker

  - `-ObjC` <https://developer.apple.com/library/content/qa/qa1490/_index.html>

    > This flag causes the linker to load every object file in the library that defines an Objective-C class or category. While this option will typically result in a larger executable (due to additional object code loaded into the application), it will allow the successful creation of effective Objective-C static libraries that contain categories on existing classes.  

  - `-all_load` <http://stackoverflow.com/questions/2906147/what-does-the-all-load-linker-flag-do>

    > IMPORTANT: For 64-bit and iPhone OS applications, there is a linker bug that prevents -ObjC from loading objects files from static libraries that contain only categories and no classes. The workaround is to use the -all_load or -force_load flags. -all_load forces the linker to load all object files from every archive it sees, even those without Objective-C code. -force_load is available in Xcode 3.2 and later. It allows finer grain control of archive loading. Each -force_load option must be followed by a path to an archive, and every object file in that archive will be loaded.

## Static Library Project Integration in Workspace

![](img/static-search-path.png)

- Drag a static lib proj into workspace
- Add the resultant lib from the project into `linked binary` section of build phase
- Specify header search in user header
  - Don't forget to exclude unnecessary path
  - Make the search recursive if necessary
