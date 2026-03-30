package server

type KeyChecker interface {
	Check(name string, key string) bool
}
