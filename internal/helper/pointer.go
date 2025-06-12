package helper

// func PointerOf[A any](a A) *A { returns a pointer to a.
func PointerOf[A any](a A) *A {
	return &a
}
