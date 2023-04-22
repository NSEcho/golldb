package golldb

import (
	"encoding/hex"
)

func encodeToHexByteSlice(value string) []byte {
	return []byte(hex.EncodeToString([]byte(value)))
}
