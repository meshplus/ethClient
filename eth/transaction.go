package eth

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func NewTransaction(from string) *Transaction {
	return &Transaction{
		To:    "0x0",
		From:  chPrefix(from),
		Nonce: getRandNonce(),
	}
}

// getRandNonce get a random nonce
func getRandNonce() int {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	buf[0] &= 0x7f
	r := binary.BigEndian.Uint64(buf[:])
	return int(r)
}

// chPrefix return a string start with '0x'
func chPrefix(origin string) string {
	if strings.HasPrefix(origin, "0x") {
		return origin
	}
	return "0x" + origin
}

// Invoke add transaction isInvoke
func (t *Transaction) Invoke(to string, payload []byte) *Transaction {
	if string(payload[0:8]) == "fefffbce" {
		t.payload = chPrefix("fefffbce" + common.Bytes2Hex(payload[8:]))
	} else {
		t.payload = chPrefix(common.Bytes2Hex(payload))
	}
	t.To = chPrefix(to)
	return t
}
