package protobytes

import (
	"bufio"
	"encoding/binary"
	"io"
)

type Reader struct {
	r   *bufio.Reader
	err error
}

func (p *Reader) TryReadUint8() (i uint8) {
	i = p.TryPeekUint8()
	if p.err == nil {
		p.r.Discard(1)
	}
	return
}

func (p *Reader) TryReadUint16le() (i uint16) {
	i = p.TryPeekUint16le()
	if p.err == nil {
		p.r.Discard(2)
	}
	return
}

func (p *Reader) TryReadUint32le() (i uint32) {
	i = p.TryPeekUint32le()
	if p.err == nil {
		p.r.Discard(4)
	}
	return
}

func (p *Reader) TryReadUint64le() (i uint64) {
	i = p.TryPeekUint64le()
	if p.err == nil {
		p.r.Discard(8)
	}
	return
}

func (p *Reader) TryReadUint16be() (i uint16) {
	i = p.TryPeekUint16be()
	if p.err == nil {
		p.r.Discard(2)
	}
	return
}

func (p *Reader) TryReadUint32be() (i uint32) {
	i = p.TryPeekUint32be()
	if p.err == nil {
		p.r.Discard(4)
	}
	return
}

func (p *Reader) TryReadUint64be() (i uint64) {
	i = p.TryPeekUint64be()
	if p.err == nil {
		p.r.Discard(8)
	}
	return
}

// hack for no alloc
func (p *Reader) next(n int) ([]byte, error) {
	buf, err := p.tryPeek(n)
	if err == nil {
		p.r.Discard(n)
	}
	return buf, err
}

func (p *Reader) TryPeekUint8() (i uint8) {
	buf, err := p.tryPeek(1)
	if err != nil {
		return
	}
	return buf[0]
}

func (p *Reader) TryPeekUint16le() (i uint16) {
	buf, err := p.tryPeek(2)
	if err != nil {
		return
	}

	return binary.LittleEndian.Uint16(buf)
}

func (p *Reader) TryPeekUint32le() (i uint32) {
	buf, err := p.tryPeek(4)
	if err != nil {
		return
	}

	return binary.LittleEndian.Uint32(buf)
}

func (p *Reader) TryPeekUint64le() (i uint64) {
	buf, err := p.tryPeek(8)
	if err != nil {
		return
	}

	return binary.LittleEndian.Uint64(buf)
}

func (p *Reader) TryPeekUint16be() (i uint16) {
	buf, err := p.tryPeek(2)
	if err != nil {
		return
	}

	return binary.BigEndian.Uint16(buf)
}

func (p *Reader) TryPeekUint32be() (i uint32) {
	buf, err := p.tryPeek(4)
	if err != nil {
		return
	}

	return binary.BigEndian.Uint32(buf)
}

func (p *Reader) TryPeekUint64be() (i uint64) {
	buf, err := p.tryPeek(8)
	if err != nil {
		return
	}

	return binary.BigEndian.Uint64(buf)
}

func (p *Reader) tryPeek(size int) ([]byte, error) {
	if p.err != nil {
		return nil, p.err
	}

	buf, err := p.r.Peek(size)
	if err != nil {
		p.err = err
	}
	return buf, err
}

func (p *Reader) TryByte() (b byte) {
	if p.err != nil {
		return
	}

	b, p.err = p.r.ReadByte()
	return
}

func (p *Reader) TryNext(n int) BytesReader {
	buf, err := p.Next(n)
	if err != nil {
		return nil
	}

	return buf
}

func (p *Reader) Next(n int) (BytesReader, error) {
	buf, err := p.next(n)
	if err != nil {
		return nil, err
	}

	return BytesReader(buf), nil
}

func (p *Reader) Read(buf []byte) (n int, err error) {
	if p.err != nil {
		return 0, p.err
	}

	n, err = p.r.Read(buf)
	if err != nil {
		p.err = err
	}
	return
}

func (p *Reader) ReadFull(buf []byte) (err error) {
	_, err = io.ReadFull(p.r, buf)
	return
}

func (p *Reader) TryReadFull(buf []byte) {
	if err := p.ReadFull(buf); err != nil {
		p.err = err
	}
}

func (p *Reader) Error() error {
	return p.err
}

func (p *Reader) Reset(reader io.Reader) {
	p.r.Reset(reader)
	p.err = nil
}

func (p *Reader) Buffered() int {
	return p.r.Buffered()
}

func New(reader io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(reader),
	}
}
