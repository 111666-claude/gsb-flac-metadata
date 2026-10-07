# flac-metadata

解析 FLAC 元数据块链：24 位大端块长、末块标志、STREAMINFO 位域、VORBIS_COMMENT 小端结构。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/flacmeta --sample=len
go run ./cmd/flacmeta --sample=last
go run ./cmd/flacmeta --sample=rate
go run ./cmd/flacmeta --sample=comment
go run ./cmd/flacmeta --sample=work
```

## 口径

- **块头**：4 字节，首字节最高位是"末块"标志、低 7 位是类型，后 3 字节是**大端**数据长度。
- **链**：从 `fLaC` 之后开始，遇到带末块标志的块就结束；之后是音频帧，不属于元数据。
- **STREAMINFO**（类型 0）：采样率 20 位、声道数 3 位、位深 5 位、总样本数 36 位，按位取。
- **VORBIS_COMMENT**（类型 4）：厂商长度与注释长度都是**小端 32 位**；注释形如 `KEY=value`，键大小写不敏感。
- **查找**：按注释键查走索引。

## 不变量

- 块链在末块标志处结束，不会把音频帧当成块。
- 采样率、声道数、位深、总样本数来自 STREAMINFO 的位域。
- `scanned` 不随注释数乘查询次数放大：八百条注释查八百次的 `scanned` 不超过 6000。

## 输出契约

`Meta` 含块列表、STREAMINFO 字段、厂商与注释映射。场景打印一行 JSON。
`--sample=work` 打印 `{"comments": N, "scanned": N}`。
