package main

import (
	"errors"
	"testing"

	"github.com/illbjorn/echo"
	"github.com/illbjorn/zest"
)

func TestSome(t *testing.T) {
	echo.Infof("%*s", 20, "Hello")
}

func init() {
	echo.SetLevel(echo.LevelDebug)
	echo.SetFlags(echo.WithCallerFile, echo.WithCallerLine, echo.WithLevel, echo.WithColor)
}

func TestParseExtensionInput(t *testing.T) {
	z := zest.New(t)

	// well-formed inputs
	testParseExtensionInput(z, "modular-mojotools.vscode-mojo", "modular-mojotools", "vscode-mojo", "", nil)
	testParseExtensionInput(z, "modular-mojotools.vscode-mojo@3.25", "modular-mojotools", "vscode-mojo", "3.25", nil)

	// ill-formed version
	testParseExtensionInput(z, "modular-mojotools.vscode-mojo@", "modular-mojotools", "vscode-mojo", "3.25", ErrAtOOB)
	testParseExtensionInput(z, "@modular-mojotools.vscode-mojo", "modular-mojotools", "vscode-mojo", "3.25", ErrAtOOB)

	// ill-formed publisher
	testParseExtensionInput(z, ".vscode-mojo", "modular-mojotools", "vscode-mojo", "3.25", ErrNoPub)
	testParseExtensionInput(z, ".@modular-mojotools.vscode-mojo", "modular-mojotools", "vscode-mojo", "3.25", ErrNoPub)

	// ill-formed ID
	testParseExtensionInput(z, "modular-mojotools.", "modular-mojotools", "vscode-mojo", "3.25", ErrNoID)

	// missing 'publisher.ID' dot separator
	testParseExtensionInput(z, "modular-mojotools vscode-mojo", "", "", "", ErrNoDot)
}

func testParseExtensionInput(z zest.Zester, input, wantPub, wantID, wantVer string, wantErr error) {
	// Attempt the parse
	gotPub, gotID, gotVer, gotErr := ParseExtension(input)

	// Assert expectations
	if wantErr != nil {
		z.Assert(errors.Is(gotErr, wantErr), "expected error [%v], got [%v]", wantErr, gotErr)
	} else {
		z.Assert(gotErr == nil, "expected no error, got [%s]", gotErr)
		z.Assert(gotPub == wantPub, "expected publisher [%s] got [%s]", wantPub, gotPub)
		z.Assert(gotID == wantID, "expected ID [%s] got [%s]", wantID, gotID)
		z.Assert(gotVer == wantVer, "expected version [%s] got [%s]", wantVer, gotVer)
	}
}
