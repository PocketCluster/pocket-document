# Build Virtualbox SDK

- [Download SDK](https://www.virtualbox.org/wiki/Downloads) and Unarchive it

- Compile object files

  ```
  cd sdk/bindings/c/samples
  make
  ```

- Take following files

  ```
  sdk/bindings/c/glue/VBoxCAPIGlue.h
  sdk/bindings/c/include/VBoxCAPI_v5_1.h
  sdk/bindings/c/samples/VBoxCAPIGlue.o
  sdk/bindings/c/samples/VirtualBox_i.o
  ```

# GPL Header License Issue

- <http://elinux.org/Legal_Issues#Use_of_kernel_header_files_in_user-space>


  > It is allowed to use kernel header files in user space, in order for user-space programs to interact with the kernel via ordinary system calls. This is allowed without the result that the user-space program becomes a derivative work of the kernel and therefore subject to GPL.
  >
  > In general, use of header files do not create derivative works, although there can be exceptions. There used to be a lot of attention paid to the amount of code (e.g. number of lines) included from a header file, but no one seems to care about that these days, and this is almost never a problem. Richard Stallman has stated that use of header files for data structures, constant definitions, and enumerations (and even small inlines) does not create a derivative work. See: http://lkml.indiana.edu/hypermail/linux/kernel/0301.1/0362.html
  >
  > The user-space use of the kernel header files is expected and ordinary. This explicitly encompasses non-GPL software using these files, and not being affected by the GPL. In order to calm fears about using the header files directly, and to prevent leakage of kernel internal information to user-space (where it might be abused), the mainline kernel developers added an option to the kernel build system to specifically create "sanitized" headers that are deemed safe for use by user-space programs, without incurring licensing issues.
  >
  > These are the "make headers_check" and "make headers_install" targets in the kernel build system.
  >
  > In general, it is legally safest to use such sanitized headers (that is, headers that have specifically been stripped of large inline macros or anything not required for user space.)
  >
  > This article explains how to create sanitized kernel headers using the kernel build system. http://darmawan-salihun.blogspot.jp/2008/03/sanitizing-linux-kernel-headers-strange.html
  >
  > Note that a different process was used by the developers of the Android operating system, to sanitize headers for bionic for their system. Their process was developed around the same time as the mainline header sanitization feature.