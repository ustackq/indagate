package generator

import (
	"github.com/ustackq/indagate/pkg/service"
)

type Generator struct {
	state   uint64
	machine uint64
}

// IDGenerator implement
type idGenerator struct {
	Generator *Generator
}

type IDGeneratorOp func(*idGenerator)

func (g *Generator) Next() service.ID {
	// TODO: complete actual Next
	return 0
}
