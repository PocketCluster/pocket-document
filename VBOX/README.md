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

<http://elinux.org/Legal_Issues#Use_of_kernel_header_files_in_user-space>

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
  
# GPL Header Issue From Oracle

[API Licensing](https://forums.virtualbox.org/viewtopic.php?f=34&t=65063)

> The general VirtualBox sources are covered by GPLv2, and as you observed this generally includes the sample code. We have no intentions going after code which uses the samples as inspiration (copying significant portions of the code is as usual a different story), because that's what the intention behind having the samples.
>
> The actual headers/jar files etc. in the SDK which you need to use for creating yout own API client are covered by different, more liberal licenses - generally LGPLv2.1, and we consider the SDK header files to be free of code (this avoids certain difficulties with LGPL, as code from header files would otherwise "pollute" your binaries). The aim is to leave you free choice for licensing your application. We encourage going open source, but you can also elect any other license.
>
> If you find any files in the SDK (besides sample code) which mention a too strict license, please name them and we'll look into resolving this as soon as possible.

- - -

> Can you say what actual problem you have with the VBoxCAPI_*.h files covered by LGPL? They don't contain any traces of what anyone could consider code which would trigger the annoying re-linking clause which is to our knowledge the biggest issue (and thus out of the way). As the comment mentions, we clarified the licensing of the file (which is derived from code developed by the Mozilla project), and we can't slap a totally different license on it without getting this nodded off by the Mozilla project. That's a nightmare for us and them, because such fundamental licensing changes strictly speaking needs explicit approval of each and every contributor who touched the relevant code in the past.
> 
> The VBoxCAPIGlue.* files are different. They're fully developed by us. For the header file LGPL would be again no problem (no code in it), but of course we considered the consequences of the licensing of the C part, which is meant to be statically linked with 3rd party code, and there it's obvious that LGPL would be a problem.

# Full Package Download Path

<http://download.virtualbox.org/virtualbox/>
