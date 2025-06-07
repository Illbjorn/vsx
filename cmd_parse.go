package main

import (
	"strings"

	"github.com/illbjorn/echo"
)

func tokenizeCMD(input string) []string {
	i := -1
	from := i + 1
	tokens := make([]string, 0, 4)

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

func parseCMD(tokens []string) (cmd string, flags map[string][]string, args []string) {
	flags = map[string][]string{}
	i := -1
	for {
		i += 1
		if i >= len(tokens) {
			return
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
			echo.Errorf("Found unexpected token [%s][%s].", next, next2)
			return

		case !nextIsFlag:
			// Positional arg
			if cmd == "" {
				cmd = next
			} else {
				args = append(args, next)
			}

		case next2 != "" && !next2IsFlag:
			// int/string flag
			name := strings.TrimLeft(next, "-")
			flags[name] = append(flags[name], next2)
			// Offset our position to account for consuming the next2 value
			i++

		case next2 == "" || next2IsFlag:
			// bool flag
			name := strings.TrimLeft(next, "-")
			flags[name] = append(flags[name], "true")
		}
	}
}
