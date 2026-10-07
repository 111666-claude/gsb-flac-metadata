package flacmeta

import "testing"

func TestParseMagic(t *testing.T) {
	data := append([]byte("fLaC"), block(1, true, []byte{1, 2, 3})...)
	meta, err := Parse(data, &Counter{})
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if len(meta.Blocks) != 1 {
		t.Fatalf("应有一个块：%d", len(meta.Blocks))
	}
}

func TestBadMagic(t *testing.T) {
	if _, err := Parse([]byte("xxxx"), &Counter{}); err == nil {
		t.Fatal("魔数不对应报错")
	}
}

func TestBlockType(t *testing.T) {
	data := append([]byte("fLaC"), block(1, true, []byte{1})...)
	meta, _ := Parse(data, &Counter{})
	if meta.Blocks[0].Type != 1 {
		t.Fatalf("类型应是 1：%d", meta.Blocks[0].Type)
	}
}

func TestBlockData(t *testing.T) {
	data := append([]byte("fLaC"), block(1, true, []byte{1, 2, 3})...)
	meta, _ := Parse(data, &Counter{})
	if len(meta.Blocks[0].Data) != 3 {
		t.Fatalf("数据长度应是 3：%d", len(meta.Blocks[0].Data))
	}
}

func TestCommentMissing(t *testing.T) {
	data := append([]byte("fLaC"), block(1, true, []byte{1})...)
	meta, _ := Parse(data, &Counter{})
	if _, ok := meta.Comment("title", &Counter{}); ok {
		t.Fatal("缺失注释不该命中")
	}
}

func TestCounterCounts(t *testing.T) {
	c := &Counter{}
	Parse(append([]byte("fLaC"), block(1, true, []byte{1})...), c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}
