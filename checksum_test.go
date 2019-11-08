package libbitcoin

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func Test_hello(t *testing.T) {
	const mask uint64 = 0xffffffffffff8000
	//const mask uint64 = 0x0080ffffffffffff

	const index uint32 = 0

	const invmask uint32 = 0x00007FFF
	//invmask

	//hash, _ := hex.DecodeString("36291c89d289f8f5e5eaf1b6744d86e24a43d271291f33693b6aab2adb954733")

	hashpart, _ := hex.DecodeString("36291c89d289f8f5e5eaf1b6744d86e24a43d271")
	//hashpart, _ := hex.DecodeString("71d2434ae2864d74b6f1eae5f5f889d2891c2936")

	upper := binary.LittleEndian.Uint64(reverse(hashpart)) & mask
	//upper := binary.BigEndian.U.Uint64(hashpart) & mask

	//lower := index & invmask

	answer := upper

	print(answer)

	// Use an arbitrary offset to the middle of the hash.
	//const auto tx = from_little_endian_unsafe < uint64_t > (hash_.begin() + 12)

	//     const auto index = static_cast<uint64_t>(index_);

	//     const auto tx_upper_49_bits = tx & mask;
	//     const auto index_lower_15_bits = index & ~mask;
	//     return tx_upper_49_bits | index_lower_15_bits;
	// }
}

func reverse(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}
