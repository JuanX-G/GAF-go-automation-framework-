package main

import (
	comps "GAF/internal/components"
	"GAF/internal/interpreter"
	"GAF/internal/parser"
	"GAF/internal/runer"
	"GAF/pkg/scheduler"
	"context"
	"fmt"
	"os"
	"strings"
)

// provide a / slash separated path
func PathLocalize(filePath string) (string, error) {
	if filePathStyle == "unix" {
		return filePath, nil 		
	} else if filePathStyle == "windows" {
		return strings.ReplaceAll(filePath, "/", "\\"), nil
	}
	
	return "", fmt.Errorf("Config error")
}
func parseConf() map[string]string {
	err := os.Chdir("./configs")
	if err != nil {
		err := os.Chdir(".\\configs")
		if err != nil {
			fmt.Println("Error parsing main conf", err)
			os.Exit(1)
		}
	}
	defer os.Chdir("..")
	f, err := os.ReadFile("main.conf")
	if err != nil {
		fmt.Println("Error parsing main conf", err)
		os.Exit(1)
	}
	vkMap := make (map[string]string)
	lines := strings.Split(string(f), "\n")
	for i, v := range lines {
		valKey := strings.Split(v, "=")
		if len(valKey) < 2 {
			fmt.Println("config error at line", i + 1)
		} else {
			vkMap[valKey[0]] = valKey[1]
		}
	}
	return vkMap
}

func runWrap(imap parser.IMaps) {
	filePath, err := PathLocalize("./configs/automations")

	if err != nil {
		panic(err)
	}
	err = os.Chdir(filePath)
	if err != nil {
		fmt.Println("error occured reading the automations!", err)
		os.Exit(1)
	}
	aDir, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("error occured reading the automations!", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())

	for _, v := range aDir {
		name := v.Name()
		nameS := strings.Split(name, "_")
		if len(nameS) != 2 {
			continue
		}
		if nameS[0] != "gaf" {
			continue
		}
		afc, err := os.ReadFile(v.Name())
		if err != nil {
			fmt.Println("error occured reading the automations!", err)
			os.Exit(1)
		}

		r := &runer.Runner{
			Automations: make(map[string]*comps.Automation),
			S: &scheduler.Scheduler{
				Jobs: make(map[string]*scheduler.JobEntry),
			},
		}

		err = r.Adder(string(afc), v.Name(), imap)
		if err != nil {
			fmt.Println("error adding automation:", err)
			continue
		}
		err = r.Run()
		if err != nil {
			fmt.Println("error running automation:", err)
			continue
		}

		go func() {
			err := r.Starter(ctx)
			if err != nil {
				fmt.Println("error starting automation:", err)
			}
		}()
	}
	defer cancel()
	select {
	case <-ctx.Done():
	}
}

func loadImports() parser.IMaps {
	rMap := parser.IMaps{
		Map1s_0:  make(map[string]*interpreter.FWrapper_1s_0),
		Map1s_1s: make(map[string]*interpreter.FWrapper_1s_1s),
	}
	declMap, err := interpreter.ReadDecl()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	aMap, err := interpreter.LoadAction()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	for k, v := range declMap {
		for e, t := range aMap {
			if e == k {
				switch v {
				case "string":
					action := interpreter.IRunAction_1s_0(t, e, "")
					rMap.Map1s_0[e] = action
				case "string string":
					action := interpreter.IRunAction_1s_1s(t, e, "")
					rMap.Map1s_1s[e] = action
				}
			}
		}
	}
	return rMap
}

var filePathStyle string

func main() {
	confMap := parseConf()
	for k, v := range confMap {
	switch k {
	case "PathStyle":
		if v == "Unix" {
			filePathStyle = "unix"
		} else if v == "windows" {
			filePathStyle = "windows"
		}
		// TODO: add systax error handling
	}
	}

	iMap := loadImports()
	// in := bufio.NewReader(os.Stdin)
	args := os.Args
	switch len(args) {
	case 1:
		fmt.Println("specify an action")
		os.Exit(1)
	case 2:
		action := args[1]
		switch action {
		case "run":
			var userAName string
			fmt.Println("what automatinon do you want to run")
			fmt.Scan(&userAName)
			filePath, err := PathLocalize("./configs/automations")

			if err != nil {
				panic(err)
			}
			err = os.Chdir(filePath)
			aDir, err := os.ReadDir(".")
			if err != nil {
				fmt.Println("error occured reading the automations!", err)
				os.Exit(1)
			}
			for _, v := range aDir {
				name := v.Name()
				nameS := strings.Split(name, "_")
				if len(nameS) != 2 {
					continue
				}
				if nameS[0] != "gaf" {
					continue
				}
				aName := nameS[1]
				aName = strings.Split(aName, ".")[0]
				if aName == userAName {
					fullName := fmt.Sprint("gaf_", aName, ".txt")
					afc, err := os.ReadFile(fullName)
					if err != nil {
						fmt.Println("error reading the automation file!", err)
						os.Exit(1)
					}
					r := &runer.Runner{}
					err = r.Adder(string(afc), "printer", iMap)
					if err != nil {
						fmt.Println("error adding the automation")
					}
					err = r.Run()
					if err != nil {
						fmt.Println("error running the automation", err)
					}
					r.Starter(context.Background())
				}

			}
		case "run_all":
			fmt.Println("Running all actions")
			runWrap(iMap)
		case "interp_all":

		}

	}
}
