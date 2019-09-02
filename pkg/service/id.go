package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

// ID is a unique identifier
type ID uint64

// IDLeng is the exact length a string (or a byte slice representing it) must have in order to be decoded into a valid ID.
const IDLen = 16

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

func (id *ID) Decode(b []byte) error {
	if len(b) != IDLen {
		return ErrInvalidIDLength
	}

	res, err := strconv.ParseUint(unsafeBytesToString(b), 16, 64)
	if err != nil {
		return nil
	}

	if *id = ID(res); !id.Valid() {
		return ErrInvalidID
	}
	return nil
}

// When you used this func, must be understood string and slice underlying struct
func unsafeBytesToString(b []byte) string {
	src := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	dst := reflect.StringHeader{
		Data: src.Data,
		Len:  src.Len,
	}
	return *(*string)(unsafe.Pointer(&dst))
}

// IDFromString creates an ID from a given string.
//
// It errors if the input string does not match a valid ID.
func IDFromString(str string) (*ID, error) {
	var id ID
	err := id.DecodeFromString(str)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// DecodeFromString parses s as a hex-encoded string.
func (i *ID) DecodeFromString(s string) error {
	return i.Decode([]byte(s))
}
