package nex

func NewGroup() *NexGroup {
	return &NexGroup{}
}

func (g *NexGroup) Before(before ...BeforeFunc) *NexGroup {
	for _, b := range before {
		if b != nil {
			g.before = append(g.before, b)
		}
	}
	return g
}

func (g *NexGroup) After(after ...AfterFunc) *NexGroup {
	for _, a := range after {
		if a != nil {
			g.after = append(g.after, a)
		}
	}
	return g
}

func (g *NexGroup) Handler(f interface{}) *Nex {
	n := Handler(f)

	// copy before middleware
	if length := len(g.before); length > 0 {
		n.before = make([]BeforeFunc, length)
		copy(n.before, g.before)
	}

	// copy after middleware
	if length := len(g.after); length > 0 {
		n.after = make([]AfterFunc, length)
		copy(n.after, g.after)
	}

	return n
}
