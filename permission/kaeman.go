package permission

type Permission uint64

const (
	Developer Permission = 1 << iota
	Default

	Splasher
)
