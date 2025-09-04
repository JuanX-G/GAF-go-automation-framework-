package components

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/mmcdole/gofeed"
)


/*
action interface is a interface
wrapping some executable steps;
e.g.interpreted user code or
printing text to the stdout.

If a routine does not take
args PassArgs_n is left empty.

Read ret stores the return of
the routine. If no return exists
then the string "_NORET" as the
return.
*/
type Action interface {
	Run(ctx context.Context)
	Name() string
	ReadRet_1s() string
	PassArgs_1s(string)
	PassArgs_2s(string, string)
}

/*
trigger, contains a field for the
string representation of the trigger
cycle, x_y_z_a_; where:
x - days; y - hours
z - minutes; a - seconds
*/
type Trigger struct {
	TrigType     string // for now only time cycle. TODO: add diffrent triggers, on user call, on call from a diffrent automation
	TimeCycle    string
	TimeCycleSec int // integer representation, in seconds, of the timeCycle
}

/*
Automation keeps track of all the data
for a single automation, its trigger
and similar. Importanlty also the variable
calls.
*/
type Automation struct {
	Name         string
	AutomType    string
	AutomTrigger Trigger
	Actions      []Action
	Svars        map[string]string
	SvarsName    []string
	VarCalls     map[string]string
}

// used for conditionals in code, duh...
// |> is the delimiter for if statments, why? who knows...
type IfBlock struct {
	Operator  string
	Arg1      string
	Arg2      string
	Name1     string // appended with 1 when processed for the purpose of calling variables
	Name2     string // appened with 2 for the same purpose as above
	Ret       bool
	Block     []Action // the block of code to be ran if ret == true
	ElseBlock []Action
	Owner     *Automation // the automation used to acess variables from outisde the block in it
}

// copy of run logic to run code block in if statments, could be made into a separate function to not repeat myself
func (i *IfBlock) RunCodeBlock(block []Action) {
	for _, action := range block {
		name := action.Name()
		nameS := strings.Split(name, "|")
		if len(nameS) == 1 {
			calledV, found := i.Owner.VarCalls[name]
			if found {
				calledV = fmt.Sprint(calledV, "\r")
				argToPass := i.Owner.Svars[calledV]
				action.PassArgs_1s(argToPass)
			}
		} else if len(nameS) == 2 {
			calledV1, fnd1 := i.Owner.VarCalls[nameS[0]]
			calledV2, fnd2 := i.Owner.VarCalls[nameS[1]]
			var Arg1, Arg2 string
			if fnd1 {
				Arg1 = i.Owner.Svars[calledV1]
			}
			if fnd2 {
				Arg2 = i.Owner.Svars[calledV2]
			}
			action.PassArgs_2s(Arg1, fmt.Sprint(Arg2, "_LIT"))
		}
		action.Run(context.Background())
		ret1 := action.ReadRet_1s()
		// optionally handle return values:
		if ret1 != "_NORET" && ret1 != "true" && ret1 != "false" {
			for _, e := range i.Owner.SvarsName {
				i.Owner.Svars[e] = ret1
			}
		}
	}

}

// resolves arguments for the if statment and
// compares them, for now only using equals but oprators
// will be aded. If the comparasion is true runs the code block
func (i *IfBlock) Run(ctx context.Context) {
	var val1, val2 string

	// resolve Arg1
	if len(i.Arg1) > 0 && i.Arg1[0] == '$' {
		v, ok := i.Owner.Svars[i.Arg1]
		if ok {
			val1 = v
		} else {
			val1 = i.Arg1
		}
	} else {
		val1 = i.Arg1
	}

	// resolve Arg2
	if len(i.Arg2) > 0 && i.Arg2[0] == '$' {
		v, ok := i.Owner.Svars[i.Arg2]
		if ok {
			val2 = v
		} else {
			val2 = i.Arg2
		}
	} else {
		val2 = i.Arg2
	}

	switch i.Operator {
	case "==":
		if val1 == val2 {
			i.RunCodeBlock(i.Block)
		} else {
			i.RunCodeBlock(i.ElseBlock)
		}
	case "!=":
		if val1 != val2 {
			i.RunCodeBlock(i.Block)
		} else {
			i.RunCodeBlock(i.ElseBlock)
		}
	}
}

// reads the boolean return from the comparasion
// and outputs it as a string
func (i *IfBlock) ReadRet_1s() string {
	return strconv.FormatBool(i.Ret)
}

// doesn't do anything
func (i *IfBlock) PassArgs_1s(arg string) {
}

// passes arguments to the if block object, checks if
// the argument has NORET, if so does no asigning
func (i *IfBlock) PassArgs_2s(Arg1, Arg2 string) {
	Arg1S := strings.Split(Arg1, "_")
	if len(Arg1S) != 2 {
		i.Arg1 = Arg1
	} else if Arg1S[1] != "NORET" {
		i.Arg1 = Arg1
	} else {
		fmt.Printf("NORET used as a rgument value in ifBlock of name %s as the first argument", i.Name())
	}
	Arg2S := strings.Split(Arg2, "_")
	if len(Arg2S) <= 2 {
		goto skip
	}
	if len(Arg2S) != 2 {
		i.Arg2 = Arg2
	} else if Arg1S[1] != "NORET" {
		i.Arg2 = Arg2
	} else {
		fmt.Printf("NORET used as argument value in ifBlock of name %s as the second argument ", i.Name())
	}
skip:
}

// returns the name in the form, someName|someName
// so the return has to split to corectly reference
// the arguments. Names are, when parsed, appended
// with 1 for the first name and 2 for the second
func (i *IfBlock) Name() string {
	return fmt.Sprint(i.Name1, "|", i.Name2)
}

// precompiled print action to reduce latency
type PrintAction struct {
	NameV string
	Text  string
}

func (p *PrintAction) Run(ctx context.Context) {
	_, err := fmt.Print(p.Text)
	if err != nil {
		// TODO: add error handling!
	}
}

func (p *PrintAction) Name() string {
	return p.NameV
}

func (p *PrintAction) ReadRet_1s() string {
	return "_NORET"
}

func (p *PrintAction) PassArgs_1s(arg string) {
	p.Text = arg
}

func (p *PrintAction) PassArgs_2s(arg, arg2 string) {
	p.Text = arg
}

// precompiled delay
type DelayAction struct {
	NameV string
	Len   int
}

func (d DelayAction) Run(ctx context.Context) {
	time.Sleep(time.Duration(d.Len * 1000000000))
}

func (d DelayAction) ReadRet_1s() string {
	return "_NORET"
}

func (d DelayAction) PassArgs_1s(string) {
}
func (d DelayAction) PassArgs_2s(string, string) {
}
func (d DelayAction) Name() string {
	return d.NameV
}

type RssFetcherParser struct {
	NameV string
	Url string
	Format string
	feed *gofeed.Feed
	err error
}

func (R *RssFetcherParser) Run(ctx context.Context) {
	fp := gofeed.NewParser()
	feed, errV := fp.ParseURL(R.Url)
	R.err = errV
	if feed == nil {
		fmt.Printf("Error getting the rss feed under url '%s' with name '%s'\n", R.Url, R.Name())
		return 
	}
	R.feed = feed
}

func (R RssFetcherParser) Name() string {
	return R.NameV
}

func (R RssFetcherParser) PassArgs_1s(arg string) {
	R.Url = arg
}

func (R RssFetcherParser) PassArgs_2s(arg string, arg2 string) {
	R.Url = arg
}

func (R RssFetcherParser) ReadRet_1s() string {
	if R.err != nil {
		return "_NORET" // to be changed to _ERRRET when error handling is added
	}
	R.Format = SanitizeLFCR(R.Format)
	switch R.Format {
	case "title":
		return R.feed.Title
	}
	return "_NORET" // to be changed to _ERRRET when error handling is added, this would be an error in format syntax

}
type RunCommand struct {
	NameV   string
	Command string
	ret     string
}

func SanitizeLFCRVar(arg string) string {
	arg = strings.ReplaceAll(arg, "$", "")
	arg = strings.ReplaceAll(arg, "\n", "")
	arg = strings.ReplaceAll(arg, "\r", "")
	return arg
}

func SanitizeLFCR(arg string) string {
	arg = strings.ReplaceAll(arg, "\n", "")
	arg = strings.ReplaceAll(arg, "\r", "")
	return arg
}

// TODO: add defining shell file
func (rc *RunCommand) Run(ctx context.Context) {
	cmd := exec.Command("powershell.exe", SanitizeLFCR(rc.Command))
	for _, y := range rc.Command {
		fmt.Printf("%U", y)
	}
	out, err := cmd.Output()
	if err == nil {
		rc.ret = string(out)
		return
	} else {
		rc.ret = fmt.Sprint("ERR_", err)
	}
}

func (rc *RunCommand) Name() string {
	return rc.NameV
}

func (rc *RunCommand) ReadRet_1s() string {
	return rc.ret
}

func (rc *RunCommand) PassArgs_1s(arg string) {
	rc.Command = arg
}

func (rc *RunCommand) PassArgs_2s(arg, arg2 string) {
	rc.Command = arg
}
