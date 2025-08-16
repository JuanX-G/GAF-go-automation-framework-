package runer

import (
	comps "GAF/internal/components"
	"GAF/internal/parser"
	"GAF/pkg/scheduler"
	"context"
	"fmt"
	"strings"
)

/*
A struct containnign the automation mapped
to their names and the scheduler object.
*/
type Runner struct {
	Automations map[string]*comps.Automation
	S           *scheduler.Scheduler
}

/*
After initializing the Runner struct the Adder() method is used
to register new automations. Data arguments is a string of the
automation file, name is the file name excluding the prefix
that is gaf_ and the extension.
Argument iMap is a map conforming to the struct specified in
parser.go. Thus iMap should be created before registering the
automations. It is a map of imported (interpreted) functions.
*/
func (r *Runner) Adder(data, name string, iMap parser.IMaps) error {
	r.S = &scheduler.Scheduler{
		Jobs: make(map[string]*scheduler.JobEntry),
	}
	AutomationD, err := parser.ParseAutomation(data, name, iMap)
	if err != nil {
		return err
	}
	if r.Automations == nil {
		r.Automations = make(map[string]*comps.Automation)
	}
	r.Automations[name] = &AutomationD
	return nil
}

/*
The runner() method is the runtime for the automations.
It parses the time cycle, is resposible for scheduling
and running the specified functions by calling the run()
method of the action struct. It also resolves variables
and allows assigning of return values to variables.
*/
func (r *Runner) Run() error {
	for _, v := range r.Automations {
		if v.AutomTrigger.TrigType == "time_cycle" {
			timeCycleS, err := parser.ParseTimeCycle(v.AutomTrigger.TimeCycle)
			if err != nil {
				return fmt.Errorf("error parsing time cycle data, %w", err)
			}

			v.AutomTrigger.TimeCycleSec = timeCycleS
			runnerFunc := func(ctx context.Context) {
				for _, action := range v.Actions {
					name := action.Name()
					nameS := strings.Split(name, "|")
					if len(nameS) == 1 {
						calledV, found := v.VarCalls[name]
						if found {
							argToPass := v.Svars[calledV]
							action.PassArgs_1s(argToPass)

						}
					} else if len(nameS) == 2 {
						calledV1, fnd := v.VarCalls[nameS[0]]
						var argToPass1 string
						if fnd {
							argToPass1 = v.Svars[calledV1]

						}
						calledV2, fnd := v.VarCalls[nameS[1]]
						var argToPass2 string
						if fnd {
							argToPass2 = v.Svars[calledV2]
						}
						action.PassArgs_2s(argToPass1, fmt.Sprint(argToPass2, "_LIT"))
					}
					action.Run(ctx)
					ret1 := action.ReadRet_1s()
					if ret1 == "_NORET" {
						continue
					}
					if ret1 == "true" || ret1 == "false" {
						continue
					}
					for _, e := range v.SvarsName {
						v.Svars[e] = ret1
					}
				}
			}
			v.Name = "autom.txt"
			r.S.Add(v.AutomTrigger.TimeCycleSec, v.Name+"Automation", runnerFunc)
		}
	}
	return nil
}

/*
The starter is reposnsible for startig the scheduler.
It starts squential mode if the automation is marked
that way and conc if it is marked with conc or not marked
*/
func (r *Runner) Starter(ctx context.Context) error {
	for _, v := range r.Automations {
		if v.AutomType == "seq" {
			r.S.Sequential = true
		} else {
			r.S.Sequential = false
		}
	}
	err := r.S.Start(ctx)
	if err != nil {
		return err
	}
	return nil
}
