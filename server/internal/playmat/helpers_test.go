package playmat

import "hash/crc32"

// crc32Of is the PNG chunk checksum: CRC-32 over the type and data.
func crc32Of(kind string, data []byte) uint32 {
	h := crc32.NewIEEE()
	_, _ = h.Write([]byte(kind))
	_, _ = h.Write(data)
	return h.Sum32()
}
