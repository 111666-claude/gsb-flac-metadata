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
	return int(b[0]) | int(b[1])<<8 | int(b[2])<<16
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
		if block.Type == 4 {
			parseComment(meta, block.Data)
		}
		pos = start + length
	}
	return meta, nil
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
			meta.Comments[key] = value
		}
	}
}

// Comment 按注释键取值。
func (m *Meta) Comment(key string, c *Counter) (string, bool) {
	for k, v := range m.Comments {
		c.Scanned++
		if k == key {
			return v, true
		}
	}
	return "", false
}
