package interpreter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type FWrapper_1s_0 struct {
	Nname    string
	A1       string
	Function func(string)
}

func (w *FWrapper_1s_0) Run(ctx context.Context) {
	w.Function(w.A1)
}

func (w *FWrapper_1s_0) ReadRet_1s() string {
	return "_NORET"
}

func (w *FWrapper_1s_0) Name() string {
	return w.Nname
}

func (w *FWrapper_1s_0) PassArgs_1s(arg string) {
	w.A1 = arg
}

func (w *FWrapper_1s_0) PassArgs_2s(arg, arg2 string) {
	w.A1 = arg
}

type FWrapper_1s_1s struct {
	Nname    string
	A1       string
	R1       string
	Function func(string) string
}

func (w *FWrapper_1s_1s) Run(ctx context.Context) {
	w.R1 = w.Function(w.A1)
}

func (w *FWrapper_1s_1s) PassArgs_1s(arg string) {
	w.A1 = arg
}
func (w *FWrapper_1s_1s) PassArgs_2s(arg, arg2 string) {
	w.A1 = arg
}
func (w *FWrapper_1s_1s) Name() string {
	return w.Nname
}

func (w *FWrapper_1s_1s) ReadRet_1s() string {
	return w.R1
}

/*
This function reads the users coded
functions declarations so that it know
what wrapper to use.
The declarations are stored under "configs/declarations/gadf.txt"
(gadf stands for Go Automation Declaration File)
*/
func ReadDecl() (map[string]string, error) {
	rMap := make(map[string]string)
	os.Chdir(".\\configs\\declarations")
	defer os.Chdir("..\\..")
	dir, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, v := range dir {
		nameS := strings.Split(v.Name(), ".")
		if len(nameS) != 2 {
			continue
		} else if nameS[1] != "txt" {
			continue
		} else if nameS[0] != "gadf" {
			continue
		}
		fc, err := os.ReadFile(v.Name())
		if err != nil {
			continue
		}
		lines := strings.Split(string(fc), "\n")
		for _, e := range lines {
			lnS := strings.Split(e, ":")
			if len(lnS) != 2 {
				continue
			} else if lnS[0] != "gaa" {
				continue
			}
			lnS2 := strings.Split(lnS[1], "(")
			if len(lnS2) != 2 {
				continue
			}
			fName := lnS2[0]
			fName = strings.ReplaceAll(fName, " ", "")
			fName = strings.ReplaceAll(fName, "\n", "")

			fSig := strings.Split(lnS2[1], ")")
			fSiga := strings.Join(fSig, " ")
			fSiga = strings.ReplaceAll(fSiga, "\n", "")
			fSiga = strings.TrimSpace(fSiga)
			fSiga = strings.ReplaceAll(fSiga, "\r", "")
			rMap[fName] = fSiga
		}
	}
	return rMap, nil
}

/*
	loadAction loads user defined code for
	interpretation returning a map of user
	function names and their respective code
*/

func LoadAction() (map[string]string, error) {
	rMap := make(map[string]string)
	os.Chdir(".\\configs\\actions\\")
	defer os.Chdir("..\\..")
	dir, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, v := range dir {
		nameS := strings.Split(v.Name(), ".")
		if len(nameS) != 2 {
			continue
		}
		if nameS[1] != "go" {
			continue
		}

		nameS2 := strings.Split(nameS[0], "_")
		if len(nameS2) != 2 {
			continue
		} else if nameS2[0] != "gaa" {
			continue
		}
		fc, err := os.ReadFile(v.Name())
		if err != nil {
			continue
		}
		nameF := strings.ReplaceAll(nameS2[1], " ", "")
		rMap[nameF] = string(fc)
	}
	return rMap, nil
}

/*
	This creates a new interpreter and
	returns a wrapper containing the
	interpreted function its name and args.
	Because the function are interpreted
	the context disallows the of a general
	wrapper. That is why we use this noation:
	_1s_0 stands for argument is 1 string
	and no returns.
*/

func IRunAction_1s_0(data, funcName string, a1 string) *FWrapper_1s_0 {
	i := interp.New(interp.Options{})

	i.Use(stdlib.Symbols)

	_, err := i.Eval(data)
	if err != nil {
		panic(err)
	}
	fullName := fmt.Sprint("main.", funcName)
	v, err := i.Eval(fullName)
	if err != nil {
		panic(err)
	}
	aFunc := v.Interface().(func(string))
	wrapper := &FWrapper_1s_0{
		Nname:    funcName,
		Function: aFunc,
		A1:       a1,
	}
	return wrapper
}

func IRunAction_1s_1s(data, funcName string, a1 string) *FWrapper_1s_1s {
	i := interp.New(interp.Options{})

	i.Use(stdlib.Symbols)

	_, err := i.Eval(data)
	if err != nil {
		panic(err)
	}
	fullName := fmt.Sprint("main.", funcName)
	v, err := i.Eval(fullName)
	if err != nil {
		panic(err)
	}
	aFunc := v.Interface().(func(string) string)
	wrapper := &FWrapper_1s_1s{
		Nname:    funcName,
		Function: aFunc,
		A1:       a1,
	}
	return wrapper
}

/*
func IRunAction_2s_0(data, funcName string, a1, a2 string) fWrapper {
	i := interp.New(interp.Options{})

	i.Use(stdlib.Symbols)

	_, err := i.Eval(data)
	if err != nil {
		panic(err)
	}
	fullName := fmt.Sprint("main.", funcName)
	v, err := i.Eval(fullName)
	if err != nil {
		panic(err)
	}
	aFunc := v.Interface().(func(string, string))
	return nil
}
*/
