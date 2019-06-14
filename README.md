# About escpos #

This is a simple [Golang](http://www.golang.org/project) package that provides
[ESC-POS](https://en.wikipedia.org/wiki/ESC/P) library functions to help with
sending control codes to a ESC-POS capable printer such as an Epson TM-T82 or
similar.

These printers are often used in retail environments in conjunction with a
point-of-sale (POS) system.

## Installation ##

Install the package via the following:

    go get -u github.com/peter-mount/escpos

## Example epos-server ##

An example EPOS server implementation is available in the
[epos-server](epos-server) subdirectory of this project. This example
server is more or less compatible with [Epson TM-Intelligent](https://c4b.epson-biz.com)
printers and print server implementations.

## Usage ##

The escpos package can be used similarly to the following:

```go
package main

import (
    "bufio"
    "os"

    "github.com/knq/escpos"
)

func main() {
    f, err := os.Open("/dev/usb/lp3")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    w := bufio.NewWriter(f)
    p := escpos.New(w)

    p.Init()
    p.SetSmooth(1)
    p.SetFontSize(2, 3)
    p.SetFont("A")
    p.Write("test ")
    p.SetFont("B")
    p.Write("test2 ")
    p.SetFont("C")
    p.Write("test3 ")
    p.Formfeed()

    p.SetFont("B")
    p.SetFontSize(1, 1)

    p.SetEmphasize(1)
    p.Write("halle")
    p.Formfeed()

    p.SetUnderline(1)
    p.SetFontSize(4, 4)
    p.Write("halle")

    p.SetReverse(1)
    p.SetFontSize(2, 4)
    p.Write("halle")
    p.Formfeed()

    p.SetFont("C")
    p.SetFontSize(8, 8)
    p.Write("halle")
    p.FormfeedN(5)

    p.Cut()
    p.End()

    w.Flush()
}
```

## TODO
- Fix barcode/image support
- Update code to be idiomatic Go
- Fix example server implementation
