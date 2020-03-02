package rules

// enumerated type with reap, spare, and ignore
type result int

const (
	reap result = iota
	spare
	ignore
)
