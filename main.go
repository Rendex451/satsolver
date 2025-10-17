package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	slr "satsolver/solver"
)

const TIMEOUT = 30 * time.Second

var configs []slr.VSIDSConfig = []slr.VSIDSConfig{
	{
		Name:             "minisat-classic",
		InitialBumpInc:   1.0,   // Стандартный bump
		DecayFactor:      0.95,  // Классическое затухание
		RescaleThreshold: 1e100, // Стандартный порог
		RescaleFactor:    1e100,
	},
	{
		Name:             "glucose-aggressive",
		InitialBumpInc:   1.5,  // Более сильные bumps
		DecayFactor:      0.92, // Быстрее забывает старые конфликты
		RescaleThreshold: 1e50, // Частая нормализация
		RescaleFactor:    1e50,
	},
	{
		Name:             "stable-longterm",
		InitialBumpInc:   0.5,   // Слабые bumps для стабильности
		DecayFactor:      0.98,  // Очень медленное затухание
		RescaleThreshold: 1e200, // Редкая нормализация
		RescaleFactor:    1e100,
	},
	{
		Name:             "init-heavy",
		InitialBumpInc:   0.1,   // Минимальные bumps
		DecayFactor:      0.99,  // Почти нет затухания
		RescaleThreshold: 1e300, // Практически без rescale
		RescaleFactor:    1e100,
	},
}

var flagsConfig = map[string]slr.VSIDSConfig{
	"mc": configs[0],
	"ga": configs[1],
	"sl": configs[2],
	"ih": configs[3],
}

func main() {
	cnfFile := flag.String("f", "", "DIMACS file")
	verboseMode := flag.Bool("v", false, "enable verbose mode")
	parallelMode := flag.Bool("p", false, "enable parallel portfolio mode")
	chosenConfig := flag.String("c", "mc", "choose config to run (\"mc\", \"ga\", \"sl\", \"ih\") when not in parallel mode")
	flag.Parse()

	if *cnfFile == "" || (!*parallelMode && (flagsConfig[*chosenConfig] == slr.VSIDSConfig{})) {
		log.Fatalf("Usage: %s -f <cnf_file> [-verbose -p | -c <config_name>]", os.Args[0])
	}

	nvars, formula, err := slr.ParseDIMACS(*cnfFile)
	if err != nil {
		log.Fatalf("CNF parsing error: %v\n", err)
		return
	}

	start := time.Now()

	var sat bool
	var finalState *slr.SolverState
	var configName string

	if *parallelMode {
		ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
		defer cancel()
		sat, finalState, configName = slr.RunPortfolioSolver(ctx, nvars, formula, configs)

		if configName == "" {
			fmt.Printf("Filename: %s\tResult: TIMEOUT\n", *cnfFile)
			return
		}
	} else {
		h := slr.NewVSIDSHeuristic(nvars, flagsConfig[*chosenConfig])
		h.Init(formula)

		s := slr.NewSolverState(nvars)
		sat, finalState = slr.Dpll(formula, s, h)
		configName = flagsConfig[*chosenConfig].Name
	}

	elapsed := time.Since(start)

	var ans string
	if sat {
		ans = "SAT"
	} else {
		ans = "UNSAT"
	}

	fmt.Printf("Filename: %s\tResult: %s\tElapsed time: %s\tConfig: %s\n",
		*cnfFile, ans, elapsed, configName)

	if *verboseMode && sat {
		slr.PrintAssignments(nvars, *finalState)
	}
}
