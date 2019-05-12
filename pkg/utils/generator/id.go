package generator

import (
	"errors"
	"math/rand"
	"time"

	"github.com/ustackq/indagate/pkg/service"
)

func init() {
	rand.Seed(time.Now().UnixNano())

}

// ID is a unique identifier
var ErrInvalidID = errors.New("invalid ID")
var ErrInvalidIDLength = errors.New("id must have be [16]byte")

func NewIDGenerator(opts ...IDGeneratorOp) service.IDGenerator {
	g := &idGenerator{}
	for _, f := range opts {
		f(g)
	}
	if g.Generator == nil {
		// TODO
	}
	return g
}

func (idG *idGenerator) ID() service.ID {
	var id service.ID
	for !id.Valid() {
		id = service.ID(idG.Generator.Next())
	}
	return id
}
