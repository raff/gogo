// Example repl is a simple REPL (read-eval-print loop) for GO using
// http://github.com/0xfaded/eval to the heavy lifting to implement
// the eval() part.
//
// The intent here is to show how more to use the library, rather than
// be a full-featured REPL.
//
// A more complete REPL including command history, tab completion and
// readline editing is available as a separate package:
// http://github.com/rocky/go-fish
//
// (rocky) My intent here is also to have something that I can debug in
// the ssa-debugger tortoise/gub.sh. Right now that can't handle the
// unsafe package, pointers, and calls to C code. So that let's out
// go-gnureadline and lineedit.
package main

import (
	"fmt"
	"go/parser"
	"os"
	"reflect"

	"github.com/raff/eval"
	"github.com/gobs/cmd"
)

func intro_text() {
	fmt.Printf(`=== A simple Go eval REPL ===

The environment is stored in global variable "env".
The last result is stored in a global variable "_".

Enter expressions to be evaluated at the "go>" prompt.

To quit, enter: "quit" or Ctrl-D (EOF).
`)

}

var (
	commander = &cmd.Cmd{
		Prompt:      "go> ",
		HistoryFile: ".replhistory",
		PreLoop:     intro_text,
		Default:     evalCmd,
		EnableShell: true,
	}

	env = makeEnv()
)

func makeEnv() *eval.Env {

	// A couple of things from the fmt package.
	var fmt_funcs map[string]reflect.Value = make(map[string]reflect.Value)
	fmt_funcs["Println"] = reflect.ValueOf(fmt.Println)
	fmt_funcs["Printf"] = reflect.ValueOf(fmt.Printf)

	// A stripped down package environment.  See
	// http://github.com/rocky/go-fish and repl_imports.go for a more
	// complete environment.
	pkgs := map[string]eval.Pkg{
		"fmt": &eval.Env{
			Name:   "fmt",
			Path:   "fmt",
			Vars:   make(map[string]reflect.Value),
			Consts: make(map[string]reflect.Value),
			Funcs:  fmt_funcs,
			Types:  make(map[string]reflect.Type),
			Pkgs:   make(map[string]eval.Pkg),
		}, "os": &eval.Env{
			Name: "os",
			Path: "os",
			Vars: map[string]reflect.Value{
				"Stdout": reflect.ValueOf(&os.Stdout),
				"Args":   reflect.ValueOf(&os.Args)},
			Consts: make(map[string]reflect.Value),
			Funcs:  make(map[string]reflect.Value),
			Types: make(map[string]reflect.Type),
			Pkgs: make(map[string]eval.Pkg),
		},
	}

	mainEnv := eval.Env{
		Name:   ".",
		Path:   "",
		Consts: make(map[string]reflect.Value),
		Funcs:  make(map[string]reflect.Value),
		Types:  make(map[string]reflect.Type),
		Vars:   make(map[string]reflect.Value),
		Pkgs:   pkgs,
	}

	return &mainEnv
}


func evalCmd(line string) {
	ctx := &eval.Ctx{line}
	if expr, err := parser.ParseExpr(line); err != nil {
        //
        // error parsing
        //
		if pair := eval.FormatErrorPos(line, err.Error()); len(pair) == 2 {
			fmt.Println(pair[0])
			fmt.Println(pair[1])
		}
		fmt.Printf("parse error: %s\n", err)
	} else if cexpr, errs := eval.CheckExpr(ctx, expr, env); len(errs) != 0 {
        //
        // error ...
        //
		for _, cerr := range errs {
			fmt.Printf("%v\n", cerr)
		}
	} else if vals, _, err := eval.EvalExpr(ctx, cexpr, env); err != nil {
        //
        // error evaluating expression
        //
		fmt.Printf("eval error: %s\n", err)
    } else if vals == nil {
        fmt.Println(vals)
    } else {
        for i, v := range *vals {
            if i > 0 {
                    fmt.Printf(", ")
            }

            fmt.Printf("%s", eval.Inspect(v))
        }

        fmt.Printf("\n")

        if len(*vals) == 1 {
	        env.Vars["_"] = (*vals)[0]
        } else {
	        env.Vars["_"] = reflect.ValueOf(*vals)
        }
    }
}

func main() {
    env.Vars["env"] = reflect.ValueOf(env)
	env.Vars["_"] = reflect.ValueOf(nil)
	commander.Init()

	commander.Add(cmd.Command{
		"quit",
		`terminate application`,
		func(string) (stop bool) {
			return true
		}})

	commander.CmdLoop()
}
