package ctxlog

// Export unexported fields and methods for testing

func (s *Scope) Parent() *Scope {
	return s.parent
}
