package concatkdf

import (
	"crypto"
	"encoding/binary"

	"github.com/lestrrat-go/jwx/buffer"
	"github.com/lestrrat-go/pdebug/v3"
	"github.com/pkg/errors"
)

type KDF struct {
	buf       []byte
	hash      crypto.Hash
	otherinfo []byte
	z         []byte
}

func New(hash crypto.Hash, alg, Z, apu, apv, pubinfo, privinfo []byte) *KDF {
	algbuf := buffer.Buffer(alg).NData()
	apubuf := buffer.Buffer(apu).NData()
	apvbuf := buffer.Buffer(apv).NData()

	if pdebug.Enabled {
		pdebug.Printf("alg          = %s", alg)
		pdebug.Printf("algID   (%d) = %x", len(algbuf), algbuf)
		pdebug.Printf("zBytes  (%d) = %x", len(Z), Z)
		pdebug.Printf("apu     (%d) = %x", len(apubuf), apubuf)
		pdebug.Printf("apv     (%d) = %x", len(apvbuf), apvbuf)
		pdebug.Printf("pubinfo (%d) = %x", len(pubinfo), pubinfo)
	}

	concat := make([]byte, len(algbuf)+len(apubuf)+len(apvbuf)+len(pubinfo)+len(privinfo))
	n := copy(concat, algbuf)
	n += copy(concat[n:], apubuf)
	n += copy(concat[n:], apvbuf)
	n += copy(concat[n:], pubinfo)
	copy(concat[n:], privinfo)

	return &KDF{
		hash:      hash,
		otherinfo: concat,
		z:         Z,
	}
}

func (k *KDF) Read(out []byte) (int, error) {
	var round uint32 = 1
	h := k.hash.New()

	for len(out) > len(k.buf) {
		h.Reset()

		if err := binary.Write(h, binary.BigEndian, round); err != nil {
			return 0, errors.Wrap(err, "failed to write round using kdf")
		}
		if _, err := h.Write(k.z); err != nil {
			return 0, errors.Wrap(err, "failed to write z using kdf")
		}
		if _, err := h.Write(k.otherinfo); err != nil {
			return 0, errors.Wrap(err, "failed to write other info using kdf")
		}

		k.buf = append(k.buf, h.Sum(nil)...)
		round++
	}

	n := copy(out, k.buf[:len(out)])
	k.buf = k.buf[len(out):]
	return n, nil
}
