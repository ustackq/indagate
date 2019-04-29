package service

type TokenGenerator interface {
	Token() (string, error)
}
