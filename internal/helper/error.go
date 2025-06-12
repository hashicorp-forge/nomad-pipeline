package helper

func IgnoreError(fn func() error) { _ = fn() }
