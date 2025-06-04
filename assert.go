package main

import "github.com/illbjorn/echo"

func assert(cond bool, msgAndArgs ...any) {
	if !cond {
		if len(msgAndArgs) == 0 {
			echo.Fatalf("Assertion failed with no provided error!")
		} else if msg, ok := msgAndArgs[0].(string); !ok {
			echo.Fatalf("%v", msgAndArgs...)
		} else {
			echo.Fatalf(msg, msgAndArgs[1:]...)
		}
		panic(0)
	}
}
