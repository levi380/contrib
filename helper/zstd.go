package helper

import (
	"github.com/klauspost/compress/zstd"
)

var (
	encoder, _ = zstd.NewWriter(nil)
	decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
)

// Compress a buffer.
// If you have a destination buffer, the allocation in the call can also be eliminated.
func ZStdCompress(src []byte) []byte {
	return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}

// Decompress a buffer. We don't supply a destination buffer,
// so it will be allocated by the decoder.
func ZStdDecompress(src []byte) ([]byte, error) {
	return decoder.DecodeAll(src, nil)
}
