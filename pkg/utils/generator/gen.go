package generator

type Generator struct {
	state   uint64
	machine uint64
}

// IDGenerator implement
type idGenerator struct {
	Generator *Generator
}

type IDGeneratorOp func(*idGenerator)

func (g *Generator) Next() ID {
	// TODO: complete actual Next
	return 0
}
