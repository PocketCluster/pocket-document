# Golang Snippet

**Hyphen `-` becomes UTF-8 string if used in path**

```Go
var path string = "/usr/local/share/ca­-certificates/"
```

The string above is supposed to be `"\x2f\x75\x73\x72\x2f\x6c\x6f\x63\x61\x6c\x2f\x73\x68\x61\x72\x65\x2f\x63\x61\x2d\x63\x65\x72\x74\x69\x66\x69\x63\x61\x74\x65\x73"` in hex value but due to native golang UTF-8 encoding, hyphen `-` becomes something else linux path cannot handle. Whenever there is a special character, we should be looking into different char encoding. for now (2018/02/22), we'll hardcode hex value into code. (you can check the string with `pwd |  od -A n -t x1` command in linux)

> Reference

- [linux - Convert string to hexadecimal on command line](linux - Convert string to hexadecimal on command line - Stack Overflow.pdf)