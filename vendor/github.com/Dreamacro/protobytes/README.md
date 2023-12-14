# Protobytes

Protobytes is a Go library inspired by Rust's crate bytes. It provides a series of methods for big-endian and little-endian number operations, as well as a helper for `io.Reader`.

## Purpose

The goal of this library is to provide an easy-to-use `bytes.Buffer`. However, it has been split into `BytesReader` and `BytesWriter`, which perform similarly to using `bytes.Buffer` directly.

When using `bytes.Buffer`, `binary.Write` has poor performance and unnecessary allocation. Instead, you can use the methods provided by protobytes.

## Usage

`BytesReader` and `BytesWriter` are similar to `bytes.Buffer`, but they are not thread-safe. `bytes.Buffer`, `binary.Write` and `[]byte` conversion is very easy and cheap.

```go
buf := make([]byte, 0, 1024)
w := BytesWriter(buf)
w.ReadFull(rand.Reader, 64)
w.PutUint8(0x01)
w.PutUint16be(0x0203)

r := BytesReader(w.Bytes())
randomBytes, r := r.SplitAt(64) // split to two BytesReader
r.ReadUint8() // auto step forward
r.ReadUint16be()
```

example for parse proxy protocol v2 using `BytesReader`:

```go
hexStr := "0d0a0d0a000d0a515549540a20120c000c22384eac10000104d21f90"
buf, _ := hex.DecodeString(hexStr)

r := BytesReader(buf)

if r.Len() < 16 {
    panic("short buffer")
}

sign, r := r.SplitAt(12)
if !bytes.Equal(signature, sign) {
    panic("invalid signature")
}

header := &Header{}

switch command := r.ReadUint8(); command {
case LOCAL, PROXY:
    header.Command = command
default:
    panic(fmt.Errorf("invalid command %x", command))
}

switch protocol := r.ReadUint8(); protocol {
case UNSPEC, TCPOverIPv4, UDPOverIPv4, TCPOverIPv6, UDPOverIPv6, UNIXStream, UNIXDatagram:
    header.TransportProtocol = protocol
default:
    panic(fmt.Errorf("invalid protocol %x", protocol))
}

length := r.ReadUint16le()
switch length {
case lengthIPv4, lengthIPv6, lengthUnix:
default:
    panic(fmt.Errorf("invalid length %x", length))
}

if r.Len() < int(length) {
    panic("short buffer")
}

switch length {
case lengthIPv4:
    srcAddr := r.ReadIPv4()
    dstAddr := r.ReadIPv4()
    srcPort := r.ReadUint16be()
    dstPort := r.ReadUint16be()

    header.SourceAddr = netip.AddrPortFrom(srcAddr, srcPort)
    header.DestinationAddr = netip.AddrPortFrom(dstAddr, dstPort)
case lengthIPv6:
    srcAddr := r.ReadIPv6()
    dstAddr := r.ReadIPv6()
    srcPort := r.ReadUint16be()
    dstPort := r.ReadUint16be()

    header.SourceAddr = netip.AddrPortFrom(srcAddr, srcPort)
    header.DestinationAddr = netip.AddrPortFrom(dstAddr, dstPort)
default:
    panic(fmt.Errorf("unsupported protocol %x", length))
}

fmt.Printf("%+v\n", header)
```
