# XCode Settings

**When CocoaPod goes whacky, we might want to reset everything and start over**

1. Deintegrated cocoa pods, using `pod deintegrate`. Check this link <https://github.com/kylef/cocoapods-deintegrate>  
2. Search on the target settings and project file for `pod`. Remove anything that looked like it belonged to cocoa pods.  
3. `pod install` once again.
  
> Reference  

- <http://stackoverflow.com/questions/32768516/cocoapods-ld-library-not-found-for-lpods-objectivesugar>

**Need to build/copy extra static library before compiling the final binary**

1. Add a new script file in XCode.  

  ![](img/xcode-script1.png)
2. Setup a new `Run Script` phase in `Build Phases`  

  ![](img/xcode-script2.png)
3. Drag the script into `Run Script` phase.  

  ![](img/xcode-script3.png)
4. Edit the script for what it should do.

  ```sh
  #!/bin/bash
  
  # Exit if any command fails
  set -e
  
  # Figure out where things are coming from and going to
  PRJPATH=${SRCROOT}/GOTEST
  GO_SRC=${SRCROOT}/GOSRC
  GO=/usr/local/Cellar/go/1.6.2/libexec/bin/go
  GG_OBJ=${BUILT_PRODUCTS_DIR}/GO-OBJ
  GG_CGO=${BUILT_PRODUCTS_DIR}/CGO-OBJ
  ARCHIVE=${PRJPATH}/pc-core.a
  
  # Make the temp folders for go objects
  mkdir -p ${GG_OBJ} ${GG_CGO}
  
  # Remove the previous archive
  if [[ -f ${ARCHIVE} ]]; then
      rm ${ARCHIVE}
  fi
  
  # Generate _cgo_export.h and copy into source folder
  ${GO} tool cgo -objdir ${GG_CGO} ${GO_SRC}/main.go
  cp ${GG_CGO}/_cgo_export.h ${PRJPATH}
  
  # Compile and produce object files
  pushd ${GO_SRC}
  CGO_ENABLED=1 CC=clang ${GO} build -ldflags '-tmpdir '${GG_OBJ}' -linkmode external' ./...
  popd
  
  # Combine the object files into a static library
  ar rcs ${ARCHIVE} ${GG_OBJ}/*o
  echo "${ARCHIVE} generated!"
  ```
  
  Here's link for all the XCode environmental variables.
  - [Xcode Build Settings Reference](https://pewpewthespells.com/blog/buildsettings.html)
  - [How do I print a list of “Build Settings” in Xcode project?](http://stackoverflow.com/questions/6910901/how-do-i-print-a-list-of-build-settings-in-xcode-project)
