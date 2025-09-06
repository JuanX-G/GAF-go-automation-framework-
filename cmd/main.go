package main

import (
	comps "GAF/internal/components"
	"GAF/internal/interpreter"
	"GAF/internal/parser"
	"GAF/internal/runer"
	"GAF/pkg/scheduler"
	"bufio"
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
		if len(valKey) < 2 && v != "\n"{
			fmt.Println("config error at line", i + 1)
		} else {
			vkMap[valKey[0]] = valKey[1]
		}
	}
	return vkMap
}

func runWrap() {
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

		err = r.Adder(string(afc), v.Name(), iMap)
		if err != nil {
			fmt.Println("error adding automation:", err)
			continue
		}
		err = r.Run()
		if err != nil {
			fmt.Println("error running automation:", err)
			continue
		}

		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cChan := make(chan string)
		fmt.Println("Press q AND then enter to stop")
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				text := scanner.Text()
				if text == "q" {
					cChan <- "q"
					return
				}
			}
		}()

		go func() {
			if <-cChan == "q" {
				cancel()       
				os.Exit(0)    
			}
		}()

		err = r.Starter(ctx)
		if err != nil {
			fmt.Println("error running starter:", err)
		}

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

func runByName() {
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cChan := make(chan string)
			fmt.Println("Press q AND then enter to stop")
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					text := scanner.Text()
					if text == "q" {
						cChan <- "q"
						return
					}
				}
			}()

			go func() {
				if <-cChan == "q" {
					cancel()       
					os.Exit(0)    
				}
			}()

			err = r.Starter(ctx)
			if err != nil {
				fmt.Println("error running starter:", err)
			}
		}
	}
}

var filePathStyle string
var iMap parser.IMaps

func main() {
	confMap := parseConf()
	for k, v := range confMap {
	switch k {
	case "PathStyle":
		if v == "Unix" {
			filePathStyle = "unix"
		} else if v == "windows" {
			filePathStyle = "windows"
		} else {
			fmt.Println("Syntax error, filpath-style must be either 'unix' or 'windows'")
		}
		
	}
	}

	iMap = loadImports()
	args := os.Args
	switch len(args) {
	case 1:
		fmt.Println("specify an action")
		os.Exit(1)
	case 2:
		action := args[1]
		if action == "run" || action == "r" {
			runByName()
		} else if action == "runAll" || action == "ra" {
			fmt.Println("Running all actions")
			runWrap()
		}
	}
}
