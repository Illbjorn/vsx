package argv

import "strconv"

type (
	Flag interface {
		GetName() string
		SetValue(value string)
		GetKind() FlagKind
	}

	flagBase[T any] struct {
		Name    string
		Desc    string
		Value   T
		Default T
	}

	StringFlag flagBase[string]
	BoolFlag   flagBase[bool]
	IntFlag    flagBase[int]
)

/*                             GetName() string                              */

func (self *StringFlag) GetName() string {
	return self.Name
}

func (self *BoolFlag) GetName() string {
	return self.Name
}

func (self *IntFlag) GetName() string {
	return self.Name
}

/*                           SetValue(value string)                          */

func (self *StringFlag) SetValue(value string) {
	self.Value = value
}

func (self *BoolFlag) SetValue(value string) {
	self.Value = true
}

func (self *IntFlag) SetValue(value string) {
	self.Value, _ = strconv.Atoi(value)
}

/*                             GetKind() FlagKind                            */

func (self *StringFlag) GetKind() FlagKind {
	return KindString
}

func (self *BoolFlag) GetKind() FlagKind {
	return KindBool
}

func (self *IntFlag) GetKind() FlagKind {
	return KindInt
}

type FlagKind uint8

const (
	KindString FlagKind = 1 + iota
	KindInt
	KindBool
)
