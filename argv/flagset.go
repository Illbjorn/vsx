package argv

import "strings"

var (
	CMD = FlagSet{}
)

type FlagSet struct {
	flags []Flag
}

func (self *FlagSet) Push(f Flag) {
	self.flags = append(self.flags, f)
}

func (self *FlagSet) Parse(args []string) error {
	// outer:
	for {
		if len(args) == 0 {
			return nil
		}

		next := args[0]
		args = args[1:]

		// Identify if we have a flag or positional arg
		if strings.HasPrefix(next, "-") {
			// Parse flag
			//
			// Trim the `-|--` prefix
			next = strings.TrimLeft(next, "-")

			// Attempt to locate the flag
			for i := range len(self.flags) {
				if self.flags[i].GetName() == next {
					switch self.flags[i].GetKind() {
					case KindString:
					case KindInt:
						// Look for a value at i+1
					case KindBool:
						// Set and forget
						self.flags[i].SetValue("true")
					}
				}
			}
		}
	}
}
