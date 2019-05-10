package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
)

// ID is a unique identifier
type ID uint64

var ErrInvalidID = errors.New("invalid ID")
var ErrInvalidIDLength = errors.New("id must have be [16]byte")

// IDGenerator represents a generator for IDs.
type IDGenerator interface {
	ID() ID
}

func (id ID) Valid() bool {
	return id != 0
}

func (id ID) Encode() ([]byte, error) {
	if !id.Valid() {
		return nil, ErrInvalidID
	}

	b := make([]byte, hex.DecodedLen(16))
	binary.BigEndian.PutUint64(b, uint64(id))
	dst := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(dst, b)
	return dst, nil
}

func (id ID) String() string {
	encode, _ := id.Encode()
	return string(encode)
}
