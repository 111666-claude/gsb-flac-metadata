package flacmeta

import "fmt"

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

func block(typ int, last bool, data []byte) []byte {
	head := byte(typ & 0x7f)
	if last {
		head |= 0x80
	}
	out := []byte{head, byte(len(data) >> 16), byte(len(data) >> 8), byte(len(data))}
	return append(out, data...)
}

func le32(n int) []byte {
	return []byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24)}
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "len", Run: func() map[string]any {
			data := append([]byte("fLaC"), block(1, true, make([]byte, 300))...)
			meta, _ := Parse(data, &Counter{})
			if len(meta.Blocks) == 0 {
				return map[string]any{"declared": 0}
			}
			return map[string]any{"declared": meta.Blocks[0].Declared}
		}},
		{Name: "last", Run: func() map[string]any {
			data := append([]byte("fLaC"), block(1, false, make([]byte, 4))...)
			data = append(data, block(1, true, make([]byte, 4))...)
			data = append(data, make([]byte, 20)...)
			meta, _ := Parse(data, &Counter{})
			return map[string]any{"blocks": len(meta.Blocks)}
		}},
		{Name: "rate", Run: func() map[string]any {
			stream := make([]byte, 34)
			stream[10] = 0x0a
			stream[11] = 0xc4
			stream[12] = 0x42
			data := append([]byte("fLaC"), block(0, true, stream)...)
			meta, _ := Parse(data, &Counter{})
			return map[string]any{"rate": meta.SampleRate}
		}},
		{Name: "comment", Run: func() map[string]any {
			vendor := []byte("vendor")
			body := le32(len(vendor))
			body = append(body, vendor...)
			body = append(body, le32(1)...)
			entry := []byte("TITLE=Song")
			body = append(body, le32(len(entry))...)
			body = append(body, entry...)
			data := append([]byte("fLaC"), block(4, true, body)...)
			meta, _ := Parse(data, &Counter{})
			value, _ := meta.Comment("title", &Counter{})
			return map[string]any{"value": value}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	body := le32(0)
	body = append(body, le32(n)...)
	for i := 0; i < n; i++ {
		entry := []byte(fmt.Sprintf("K%04d=v", i))
		body = append(body, le32(len(entry))...)
		body = append(body, entry...)
	}
	data := append([]byte("fLaC"), block(4, true, body)...)
	meta, _ := Parse(data, &Counter{})
	c := &Counter{}
	for i := 0; i < n; i++ {
		meta.Comment(fmt.Sprintf("K%04d", i), c)
	}
	return map[string]any{"comments": n, "scanned": c.Scanned}
}
