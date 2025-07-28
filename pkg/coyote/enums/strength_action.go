package enums

type StrengthAction int

const (
	StrengthActionDecrease StrengthAction = iota
	StrengthActionIncrease
	StrengthActionSetTo
)
