package argv

import (
	"fmt"
	"strings"

	"github.com/illbjorn/echo"
)

type Command struct {
	Name  string
	Flags map[string][]string
	Args  []string
}

var empty = []string{}

func (c Command) Flag(names ...string) []string {
	for _, name := range names {
		if v, ok := c.Flags[name]; ok {
			return v
		}
	}
	return empty
}

func Tokenize(input string) []string {
	i := -1
	from := i + 1
	tokens := make([]string, 0, 8)

	capture := func() {
		if i-from <= 0 {
			return
		}
		tokens = append(tokens, input[from:i])
		from = i + 1
	}

	for {
		i += 1
		if i >= len(input) {
			capture()
			return tokens
		}

		// If we see an unquoted string, slice it as the next arg
		if input[i] == ' ' {
			capture()
		}

		// If we find a quote (single or double) - parse it as a string literal
		if input[i] == '\'' || input[i] == '"' {
			// Drop any dangling tokens
			capture()

			// Set the terminator we'll look for
			//
			// TODO: Handle manual escaping
			term := input[i]

			// Relocate the `from` position to inside the quote
			from = i + 1

			// Consume until we hit a terminator
			for {
				i += 1
				if i >= len(input) {
					// TODO: Produce better errors (annotate the bad string position)
					echo.Errorf(
						"Reached EOL while looking for matching [%c] in string [%s].",
						term, input,
					)
					return nil
				}

				// We've reached our terminator!
				if input[i] == term {
					capture()
					// We offset `from` here since we want to hop the trailing terminator
					from++
					break
				}
			}
		}
	}
}

func Parse(tokens []string) (Command, error) {
	cmd := Command{
		Flags: map[string][]string{},
	}

	i := -1
	for {
		i += 1
		if i >= len(tokens) {
			return cmd, nil
		}

		next := tokens[i]
		next2 := ""
		if i+1 < len(tokens) {
			next2 = tokens[i+1]
		}
		nextIsFlag := strings.HasPrefix(next, "-")
		next2IsFlag := strings.HasPrefix(next2, "-")

		switch {
		default:
			return cmd, fmt.Errorf("found unexpected token [%s][%s]", next, next2)

		case !nextIsFlag:
			// Positional arg
			if cmd.Name == "" {
				cmd.Name = next
			} else {
				cmd.Args = append(cmd.Args, next)
			}

		case next2 != "" && !next2IsFlag:
			// int/string flag
			name := strings.TrimLeft(next, "-")
			cmd.Flags[name] = append(cmd.Flags[name], next2)
			// Offset our position to account for consuming the next2 value
			i++

		case next2 == "" || next2IsFlag:
			// bool flag
			name := strings.TrimLeft(next, "-")
			cmd.Flags[name] = append(cmd.Flags[name], "true")
		}
	}
}
