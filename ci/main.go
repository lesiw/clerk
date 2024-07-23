package main

import (
	"os"

	"labs.lesiw.io/ci/golib"
	"lesiw.io/ci"
)

type actions struct {
	golib.Actions
}

func main() {
	golib.Name = "clerk"
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	ci.Handle(actions{})
}
