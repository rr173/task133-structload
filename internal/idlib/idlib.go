package idlib

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// NewID returns a random hex id. Deterministic generation in selfcheck uses NewSeq.
func NewID(prefix string) string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback: sequential id. Should never happen in practice.
		return fmt.Sprintf("%s-%012d", prefix, nextSeq())
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

var seqCounter uint64

func nextSeq() uint64 { return atomic.AddUint64(&seqCounter, 1) }

// MustJSON marshals v or panics; used for event payloads that must always encode.
func MustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
