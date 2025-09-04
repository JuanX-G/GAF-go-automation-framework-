package parser

import (
	comps "GAF/internal/components"
	"GAF/internal/interpreter"
	"fmt"
	"strconv"
	"strings"
)

type IMaps struct {
	Map1s_0  map[string]*interpreter.FWrapper_1s_0
	Map1s_1s map[string]*interpreter.FWrapper_1s_1s
}

type parseAutomationError struct {
	ErrMsg string
}

func (pe parseAutomationError) Error() string {
	return pe.ErrMsg
}

/*
Wrappers for imported function so that
they conform to the action interface
*/
func CloneFWrapper1s1s(orig *interpreter.FWrapper_1s_1s) *interpreter.FWrapper_1s_1s {
	return &interpreter.FWrapper_1s_1s{
		Nname:    orig.Nname,
		A1:       orig.A1,
		R1:       "", // fresh
		Function: orig.Function,
	}
}
func CloneFWrapper1s0(orig *interpreter.FWrapper_1s_0) *interpreter.FWrapper_1s_0 {
	return &interpreter.FWrapper_1s_0{
		Nname:    orig.Nname,
		A1:       orig.A1,
		Function: orig.Function,
	}
}

// parses a single line of code, used mostly for look ahead when registering code block i.e. inside if statments
func parseLine(data string, owner *comps.Automation) comps.Action {
	lineElems := strings.Split(data, " ")
	var actionVar comps.Action
	if len(lineElems) < 3 {
	} else {

		switch lineElems[0] {
		case "job":
			switch lineElems[1] {
			case "print":
				if len(lineElems) < 4 {
				} else {
					actionVar := &comps.PrintAction{
						Text:  lineElems[3],
						NameV: lineElems[2],
					}
					if actionVar.Text[0] == '$' {
						for k := range owner.Svars {
							k = strings.Replace(k, "\r", "", 1)
							if k == actionVar.Text {
								actionVar.NameV = comps.SanitizeLFCRVar(actionVar.NameV)
								owner.VarCalls[actionVar.NameV] = k
							}
						}
					}
					return actionVar
				}

			}
		case "if":
			ifVar := &comps.IfBlock{}
			ifVar.Name1 = fmt.Sprint(lineElems[1], "1")
			ifVar.Name2 = fmt.Sprint(lineElems[1], "2")
			ifVar.Arg1 = lineElems[2]
			ifVar.Operator = lineElems[3]
			ifVar.Arg2 = lineElems[4]
			ifVar.Owner = owner
			actionVar = ifVar
		}
	}

	return actionVar
}

func MakeAutomation() comps.Automation {
	rAutomation := comps.Automation{
		Svars:    make(map[string]string),
		VarCalls: make(map[string]string),
	}

	return rAutomation
}

func ReadStringFromScript(fullLine string) string {
	var inString bool
	var stringLs []rune
	for _, v := range fullLine {
		if v == '"' && !inString {
			inString = true
			continue
		}
		if inString && v == '"' {
			break
		}
		if inString {
			stringLs = append(stringLs, v)
		}
	}
	return string(stringLs)
}

func ParseAutomation(data, name string, iFuncs IMaps) (comps.Automation, error) {
	lines := strings.Split(data, "\n")
	AutomationV := MakeAutomation()
	AutomationV.Name = name
	blockedMap := make(map[string]bool)
	for lni, v := range lines {
		lineElems := strings.Split(v, " ")
		if len(lineElems) > 0 {
			if lineElems[0] == "#" {
				continue
			}
		}
		if len(lineElems) == 1 {
			lineElems[0] = comps.SanitizeLFCR(lineElems[0])
			switch lineElems[0] {
			case "seq":
				AutomationV.AutomType = "seq"
			}
		} else if len(lineElems) < 3 {
			continue
		}
		switch lineElems[0] {
		case "if":
			if len(lineElems) != 6 {
				err := parseAutomationError{
					ErrMsg: fmt.Sprintf("error parsing line %d\n%s", lni, strings.Join(lineElems, "")),
				}
				fmt.Println(err.ErrMsg)

				continue
			}
			ifVar := &comps.IfBlock{
				Arg1:     lineElems[2],
				Operator: lineElems[3],
				Arg2:     lineElems[4],
				Name1:    fmt.Sprint(lineElems[1], "1"),
				Name2:    fmt.Sprint(lineElems[1], "2"),
				Owner:    &AutomationV,
			}
			var codeBlock []comps.Action
			var elseBlock []comps.Action
			block := false

			for i, e := range lines {
				eS := strings.Split(e, " ")
				eS[0] = comps.SanitizeLFCR(eS[0])
				if len(eS) < 1 {
					continue
				}
				if eS[0] == ">|" && block {
					if len(eS) > 1 {
						eS[1] = comps.SanitizeLFCR(eS[1])
						if eS[1] == ">>" {
							for _, p := range lines[i:] {
								parsed := parseLine(p, &AutomationV)
								if parsed != nil {
									blockedMap[parsed.Name()] = true
									if parsed != nil {
										elseBlock = append(elseBlock, parsed)
										ifVar.ElseBlock = elseBlock
									}
									if eS[0] == ">|" {
										break
									}
								}
							}
						}
					}
					block = false
					break
				} else if block {
					parsed := parseLine(e, &AutomationV)
					if parsed != nil {
						codeBlock = append(codeBlock, parsed)
						ifVar.Block = codeBlock
					}
				}
				lineElems2 := strings.Split(e, " ")
				if lineElems2[0] == "if" && lineElems2[1] == strings.Replace(ifVar.Name1, "1", "", 1) {
					block = true
				}
			}

			ifVar.Block = codeBlock
			if ifVar.Arg1[0] == '$' {
				for k := range AutomationV.Svars {
					if true {
						ifVar.Name1 = comps.SanitizeLFCRVar(ifVar.Name1)
						AutomationV.VarCalls[ifVar.Name1] = k
					}
				}
			}
			if ifVar.Arg2[0] == '$' {
				for k := range AutomationV.Svars {
					if true {
						ifVar.Name2 = comps.SanitizeLFCRVar(ifVar.Name2)
						AutomationV.VarCalls[ifVar.Name2] = k
					}
				}
			}
			AutomationV.Actions = append(AutomationV.Actions, ifVar)
		case "trig":
			switch lineElems[1] {
			case "time_cycle":
				AutomationV.AutomTrigger.TrigType = "time_cycle"
				AutomationV.AutomTrigger.TimeCycle = lineElems[2]

			case "on_call":
				AutomationV.AutomTrigger.TrigType = "on_call"
			}
		case "job":
			switch lineElems[1] {
			case "print":
				if len(lineElems) < 4 {
					err := parseAutomationError{
						ErrMsg: fmt.Sprintf("error parsing line %d\n%s", lni, strings.Join(lineElems, "")),
					}
					fmt.Println(err.ErrMsg)
					continue
				}
				action := &comps.PrintAction{
					NameV: lineElems[2],
					Text:  lineElems[3],
				}

				if action.Text[0] == '$' {
					for k := range AutomationV.Svars {
						k = comps.SanitizeLFCR(k)

						action.Text = comps.SanitizeLFCRVar(action.Text)

						if k == action.Text {
							action.NameV = comps.SanitizeLFCRVar(action.NameV)
							AutomationV.VarCalls[action.NameV] = k

						}
					}
				}
				codeBlock := false
				if len(lineElems) > 4 {
					codeMark := comps.SanitizeLFCR(lineElems[4])
					if codeMark == "<>" {
						codeBlock = true
					}
				}
				if !codeBlock {
					AutomationV.Actions = append(AutomationV.Actions, action)
				}
			case "runCommand":
				if len(lineElems) < 4 {
					continue
				}
				actionCmd := &comps.RunCommand{
					NameV:   lineElems[2],
					Command: ReadStringFromScript(v),
				}
				codeBlock := false
				if len(lineElems) > 5 {
					if lineElems[4] == "<>" {
						codeBlock = true
					}
				}
				if actionCmd.Command[0] == '$' {
					for k := range AutomationV.Svars {
						k = comps.SanitizeLFCR(k)

						if k == actionCmd.Command {
							actionCmd.NameV = strings.Replace(actionCmd.NameV, "$", "", 1)
							AutomationV.VarCalls[actionCmd.NameV] = k
						}
					}
				}
				if !codeBlock {
					AutomationV.Actions = append(AutomationV.Actions, actionCmd)
				}
				if len(lineElems) == 6 {
					varN := lineElems[5]
					varN = comps.SanitizeLFCRVar(varN)
					AutomationV.Svars[varN] = "_"

					AutomationV.SvarsName = append(AutomationV.SvarsName, varN)
				}
			case "rssFetchParse":
				if len(lineElems) < 5 {
					continue
				}
				actionRss := &comps.RssFetcherParser{
					NameV: lineElems[2],
					Url: lineElems[3],
					Format: lineElems[4],
				}
				AutomationV.Actions = append(AutomationV.Actions, actionRss)
			}
		case "delay":
			sec, err := ParseTimeCycle(lineElems[2])
			if err != nil {
				errMsgVal := fmt.Sprintf("the following error occured in parsing automation, %s", err.Error())
				err := parseAutomationError{
					ErrMsg: errMsgVal,
				}
				return AutomationV, err
			}
			var action comps.DelayAction
			action.NameV = lineElems[1]
			action.Len = sec
			AutomationV.Actions = append(AutomationV.Actions, action)
		case "ijob":
			action := lineElems[1]
			action = strings.ReplaceAll(action, " ", "")
			action = comps.SanitizeLFCR(action)
			action = strings.Trim(action, " ")
			todo, found := iFuncs.Map1s_0[action]
			if !found {
				orig, found := iFuncs.Map1s_1s[action]
				if found {
					clone := CloneFWrapper1s1s(orig)
					clone.Nname = lineElems[2]
					clone.A1 = lineElems[3]
					if clone.A1[0] == '$' {
						for k := range AutomationV.Svars {
							if k == clone.A1 {
								k = comps.SanitizeLFCR(k)
								clone.Nname = strings.Replace(clone.Nname, "$", "", 1)
								AutomationV.VarCalls[clone.Nname] = k
							}
						}
					}
					AutomationV.Actions = append(AutomationV.Actions, clone)
				}

				if len(lineElems) == 6 {
					varN := lineElems[5]
					varN = comps.SanitizeLFCRVar(varN)
					AutomationV.Svars[varN] = "_"
					AutomationV.SvarsName = append(AutomationV.SvarsName, varN)
				}
				continue
			}
			todo.A1 = lineElems[3]
			todo.Nname = lineElems[2]
			if todo.A1[0] == '$' {
				todo.Nname = comps.SanitizeLFCR(todo.Nname)
				AutomationV.VarCalls[todo.Nname] = todo.A1
			}
			AutomationV.Actions = append(AutomationV.Actions, todo)
			if len(lineElems) == 6 {
				varN := lineElems[5]
				varN = strings.Trim(varN, " ")
				varN = comps.SanitizeLFCR(varN)
				AutomationV.Svars[varN] = ""
				AutomationV.SvarsName = append(AutomationV.SvarsName, varN)
			}
		}
	}
	return AutomationV, nil
}

// used for cleaner error messages, containing which field is formated wrong
type timeSyntaxError struct {
	errMsg string
}

func (te timeSyntaxError) Error() string {
	return te.errMsg
}

func (te *timeSyntaxError) SetMsg(timeSize string) {
	te.errMsg = fmt.Sprintf("error handling the time format in timeframe %s", timeSize)
}

func ParseTimeCycle(data string) (seconds int, err error) {
	dataS := strings.Split(data, "_")
	if len(dataS) != 5 {
		err := timeSyntaxError{}
		err.SetMsg("All time frames, either to many or to few fields")
		return 0, err
	}
	dayI, err := strconv.Atoi(dataS[0])
	if err != nil {
		err := timeSyntaxError{}
		err.SetMsg("day")
		return 0, err
	}
	ds := dayI * 86400

	hourI, err := strconv.Atoi(dataS[1])
	if err != nil {
		err := timeSyntaxError{}
		err.SetMsg("hour")
		return 0, err
	}
	hs := hourI * 3600

	minI, err := strconv.Atoi(dataS[2])
	if err != nil {
		err := timeSyntaxError{}
		err.SetMsg("minute")
		return 0, err
	}
	mins := minI * 60

	s, err := strconv.Atoi(dataS[3])
	if err != nil {
		err := timeSyntaxError{}
		err.SetMsg("seconds")
		return 0, err
	}

	return ds + hs + mins + s, nil

}
