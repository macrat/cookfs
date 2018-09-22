package cookfs

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
)

type Hash [sha512.Size]byte

func CalcHash(data ...[]byte) Hash {
	h := sha512.New()
	for _, d := range data {
		h.Write(d)
	}
	var result Hash
	h.Sum(result[:0])
	return result
}

func ParseHash(raw string) (Hash, error) {
	x, err := hex.DecodeString(raw)
	if err != nil {
		return Hash{}, err
	}

	var h Hash
	copy(h[:], x)

	return h, nil
}

func (h Hash) ShortHash() string {
	return hex.EncodeToString([]byte(h[:4]))
}

func (h Hash) String() string {
	return hex.EncodeToString([]byte(h[:]))
}

func (h Hash) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, sha512.Size*2+2))

	if _, err := buf.WriteRune('"'); err != nil {
		return nil, err
	}

	if _, err := hex.NewEncoder(buf).Write(h[:]); err != nil {
		return nil, err
	}

	if _, err := buf.WriteRune('"'); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *Hash) UnmarshalJSON(raw []byte) error {
	var err error
	_, err = hex.Decode(h[:], raw[1:len(raw)-1])
	if err != nil {
		return err
	}
	return err
}
