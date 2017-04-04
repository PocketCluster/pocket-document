# Crash Report

**As of now, Crash Report mix from Golang and NSApplication doesn't mix**

## Issues (04/01/2017)

- It's due to the fact that Golang's `panic()` doesn't propagate to Cocoa side, and when it happens, the whole application crashes like `os.Exit(1)` being called.
- Likewise, when a crash happens on Cocoa side, Golang is not aware of and it keeps running. This is mitiated with **KSCrash** `onCrash()` handler. Nonetheless very cumbersome.
- Here are ideas to fix the situation
  1. When Cocoa runtime crashes, return from `NSApplication()`, terminate Golang, then report it to crash reporter.
  2. Or, graciously terminate Golang with `onCrash`, stop Golang runtime, then terminate Cocoa runtime with reporting the crash.
  3. Meanwhile, when crash happens on Golang, report the crash to Sentry, notify OSX, then terminate application.

## TODO
- [ ] Work with Golang mobile framework to have `NSApplication()` return.
- [ ] Take a look at electron and find out how it handles crash.
