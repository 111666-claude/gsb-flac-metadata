// Package flacmeta 解析 FLAC 元数据。
package flacmeta

import (
	"errors"
	"strings"
)

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Block 是一个元数据块。
type Block struct {
	Type     int
	Last     bool
	Declared int
	Data     []byte
}

// Meta 是解析结果。
type Meta struct {
	Blocks       []Block
	SampleRate   int
	Channels     int
	BitsPerSmpl  int
	TotalSamples int
	Vendor       string
	Comments     map[string]string
}

func readBE24(b []byte) int {
	return int(b[0])<<16 | int(b[1])<<8 | int(b[2])
}

// Parse 解析整条元数据链。
func Parse(data []byte, c *Counter) (*Meta, error) {
	if len(data) < 4 || string(data[:4]) != "fLaC" {
		return nil, errors.New("bad-magic")
	}
	meta := &Meta{Comments: map[string]string{}}
	pos := 4
	for pos+4 <= len(data) {
		c.Scanned++
		head := data[pos]
		declared := readBE24(data[pos+1 : pos+4])
		start := pos + 4
		length := declared
		if start+length > len(data) {
			length = len(data) - start
		}
		block := Block{Type: int(head & 0x7f), Last: head&0x80 != 0, Declared: declared,
			Data: append([]byte(nil), data[start:start+length]...)}
		meta.Blocks = append(meta.Blocks, block)
		switch block.Type {
		case 0:
			parseStreamInfo(meta, block.Data)
		case 4:
			parseComment(meta, block.Data)
		}
		pos = start + length
		if block.Last {
			break
		}
	}
	return meta, nil
}

func parseStreamInfo(meta *Meta, data []byte) {
	if len(data) < 18 {
		return
	}
	x := uint64(data[10])<<56 | uint64(data[11])<<48 | uint64(data[12])<<40 | uint64(data[13])<<32 |
		uint64(data[14])<<24 | uint64(data[15])<<16 | uint64(data[16])<<8 | uint64(data[17])
	meta.SampleRate = int(x >> 44)
	meta.Channels = int(x>>41)&0x7 + 1
	meta.BitsPerSmpl = int(x>>36)&0x1f + 1
	meta.TotalSamples = int(x & 0xfffffffff)
}

func parseComment(meta *Meta, data []byte) {
	if len(data) < 4 {
		return
	}
	vlen := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24
	if 4+vlen > len(data) {
		return
	}
	data = data[4+vlen:]
	if len(data) < 4 {
		return
	}
	count := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24
	data = data[4:]
	for i := 0; i < count; i++ {
		if len(data) < 4 {
			return
		}
		clen := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24
		data = data[4:]
		if clen > len(data) {
			return
		}
		entry := string(data[:clen])
		data = data[clen:]
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			meta.Comments[strings.ToUpper(key)] = value
		}
	}
}

// Comment 按注释键取值。
func (m *Meta) Comment(key string, c *Counter) (string, bool) {
	c.Scanned++
	v, ok := m.Comments[strings.ToUpper(key)]
	return v, ok
}
