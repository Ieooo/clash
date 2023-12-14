package protobytes

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"net/netip"
	"unicode/utf8"
)

const (
	smallBufferSize = 64
	maxInt          = math.MaxInt
)

// BytesReader is a wrapper for a byte slice that provides helper methods for
// reading various types of data from the slice.
type BytesReader []byte

// Len returns the length of the byte slice.
func (b *BytesReader) Len() int {
	return len(*b)
}

// Cap returns the capacity of the byte slice.
func (b *BytesReader) Cap() int {
	return cap(*b)
}

// IsEmpty checks if the byte slice is empty.
func (b *BytesReader) IsEmpty() bool {
	return b.Len() == 0
}

// SplitAt splits the byte slice at the given index and returns two new
// BytesReader.
func (b *BytesReader) SplitAt(n int) (BytesReader, BytesReader) {
	if n > b.Len() {
		n = b.Len()
	}

	buf := *b
	return buf[:n], buf[n:]
}

// SplitBy splits the byte slice by a given function and returns two new BytesReader.
func (b *BytesReader) SplitBy(f func(byte) bool) (BytesReader, BytesReader) {
	for i, c := range *b {
		if f(c) {
			return b.SplitAt(i)
		}
	}

	return *b, nil
}

// ReadUint8 reads a uint8 value from the byte slice and skips 1 byte.
func (b *BytesReader) ReadUint8() uint8 {
	r := (*b)[0]
	*b = (*b)[1:]
	return r
}

// ReadUint16be reads a uint16 value in big endian from the byte slice and skips 2 bytes.
func (b *BytesReader) ReadUint16be() uint16 {
	r := binary.BigEndian.Uint16((*b)[:2])
	*b = (*b)[2:]
	return r
}

// ReadUint32be reads a uint32 value in big endian from the byte slice and skips 4 bytes.
func (b *BytesReader) ReadUint32be() uint32 {
	r := binary.BigEndian.Uint32((*b)[:4])
	*b = (*b)[4:]
	return r
}

// ReadUint64be reads a uint64 value in big endian from the byte slice and skips 8 bytes.
func (b *BytesReader) ReadUint64be() uint64 {
	r := binary.BigEndian.Uint64((*b)[:8])
	*b = (*b)[8:]
	return r
}

// ReadUint16le reads a uint16 value in little endian from the byte slice and skips 2 bytes.
func (b *BytesReader) ReadUint16le() uint16 {
	r := binary.LittleEndian.Uint16((*b)[:2])
	*b = (*b)[2:]
	return r
}

// ReadUint32le reads a uint32 value in little endian from the byte slice and skips 4 bytes.
func (b *BytesReader) ReadUint32le() uint32 {
	r := binary.LittleEndian.Uint32((*b)[:4])
	*b = (*b)[4:]
	return r
}

// ReadUint64le reads a uint64 value in little endian from the byte slice and skips 8 bytes.
func (b *BytesReader) ReadUint64le() uint64 {
	r := binary.LittleEndian.Uint64((*b)[:8])
	*b = (*b)[8:]
	return r
}

// ReadUvarint read Uvarint from the byte slice.
// it return error because of the length of the byte slice can't be sure.
func (b *BytesReader) ReadUvarint() (uint64, error) {
	return binary.ReadUvarint(b)
}

// ReadVarint read Varint from the byte slice.
// it return error because of the length of the byte slice can't be sure.
func (b *BytesReader) ReadVarint() (int64, error) {
	return binary.ReadVarint(b)
}

// Skip skips the given number of bytes.
func (b *BytesReader) Skip(n int) {
	*b = (*b)[n:]
}

// Read reads up to len(p) bytes from the byte slice and skips len(p) bytes.
// implements io.Reader. If the buffer has no data to return, err is
// io.EOF (unless len(p) is zero); otherwise it is nil.
func (b *BytesReader) Read(p []byte) (n int, err error) {
	if b.IsEmpty() {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}

	n = copy(p, *b)
	*b = (*b)[n:]
	return
}

// ReadByte implements io.ByteReader.
func (b *BytesReader) ReadByte() (byte, error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	return b.ReadUint8(), nil
}

// ReadIPv4 reads a net.IPAddr with an IPv4 address.
func (b *BytesReader) ReadIPv4() netip.Addr {
	ip := netip.AddrFrom4([4]byte((*b)[:4]))
	*b = (*b)[4:]
	return ip
}

// ReadIPv6 reads a net.IPAddr with an IPv6 address.
func (b *BytesReader) ReadIPv6() netip.Addr {
	ip := netip.AddrFrom16([16]byte((*b)[:16]))
	*b = (*b)[16:]
	return ip
}

type BytesWriter []byte

func (b *BytesWriter) Len() int {
	return len(*b)
}

// Cap returns the capacity of the byte slice.
func (b *BytesWriter) Cap() int {
	return cap(*b)
}

// modify from bytes.Buffer
// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *BytesWriter) tryGrowByReslice(n int) (int, bool) {
	if l := len(*b); n <= cap(*b)-l {
		*b = (*b)[:l+n]
		return l, true
	}
	return 0, false
}

// modify from bytes.Buffer
// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

// modify from bytes.Buffer
// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *BytesWriter) grow(n int) int {
	m := b.Len()
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if *b == nil && n <= smallBufferSize {
		*b = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(*b)
	if c > maxInt-c-n {
		panic(bytes.ErrTooLarge)
	} else if n > c/2-m {
		// Add b.off to account for *b[:b.off] being sliced off the front.
		*b = growSlice((*b)[:], n)
	}
	// Restore b.off and len(b.buf).
	*b = (*b)[:m+n]
	return m
}

// Grow grows the buffer's capacity. It returns the index where bytes
// should be written.
func (b *BytesWriter) Grow(n int) int {
	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}
	return m
}

func (b *BytesWriter) next(n int) []byte {
	m := b.Grow(n)
	return (*b)[m : m+n]
}

func (b *BytesWriter) PutUint8(v uint8) {
	m := b.Grow(1)
	(*b)[m] = v
}

func (b *BytesWriter) PutUint16be(v uint16) {
	binary.BigEndian.PutUint16(b.next(2), v)
}

func (b *BytesWriter) PutUint32be(v uint32) {
	binary.BigEndian.PutUint32(b.next(4), v)
}

func (b *BytesWriter) PutUint64be(v uint64) {
	binary.BigEndian.PutUint64(b.next(8), v)
}

func (b *BytesWriter) PutUint16le(v uint16) {
	binary.LittleEndian.PutUint16(b.next(2), v)
}

func (b *BytesWriter) PutUint32le(v uint32) {
	binary.LittleEndian.PutUint32(b.next(4), v)
}

func (b *BytesWriter) PutUint64le(v uint64) {
	binary.LittleEndian.PutUint64(b.next(8), v)
}

func (b *BytesWriter) PutUvarint(v uint64) {
	n := binary.MaxVarintLen64
	m := b.Grow(n)

	n = binary.PutUvarint((*b)[m:], v)
	*b = (*b)[:m+n]
}

func (b *BytesWriter) PutVarint(v int64) {
	n := binary.MaxVarintLen64
	m := b.Grow(n)

	n = binary.PutVarint((*b)[m:], v)
	*b = (*b)[:m+n]
}

func (b *BytesWriter) PutSlice(p []byte) {
	copy(b.next(len(p)), p)
}

func (b *BytesWriter) PutString(s string) {
	copy(b.next(len(s)), s)
}

func (b *BytesWriter) PutRune(r rune) {
	// Compare as uint32 to correctly handle negative runes.
	if uint32(r) < utf8.RuneSelf {
		b.PutUint8(byte(r))
		return
	}
	m := b.Grow(utf8.UTFMax)
	*b = utf8.AppendRune((*b)[:m], r)
}

func (b *BytesWriter) ReadFull(reader io.Reader, n int) error {
	length := b.Len()
	n, err := io.ReadFull(reader, b.next(n))
	if err != nil {
		*b = (*b)[:length]
	}
	return err
}

func (b *BytesWriter) Slice(start, end int) BytesWriter {
	w := (*b)[start:end]
	w.Reset()
	return w
}

func (b *BytesWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	b.PutSlice(p)
	return
}

func (b *BytesWriter) Bytes() []byte {
	return (*b)[:]
}

func (b *BytesWriter) Reset() {
	*b = (*b)[:0]
}
