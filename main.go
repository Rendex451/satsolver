package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const TIMEOUT = 30 * time.Second

var configs []VSIDSConfig = []VSIDSConfig{
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

func printAssignments(nvars int, finalS SolverState) {
	fmt.Print("Assignments: ")
	for v := 1; v <= nvars; v++ {
		if finalS.assigned[v] {
			val := finalS.assignment[v] == 1
			fmt.Printf("%d=%t ", v, val)
		}
	}
	fmt.Println()
}

func main() {
	cnfFile := flag.String("f", "", "DIMACS file")
	verboseMode := flag.Bool("v", false, "enable verbose mode")
	parallelMode := flag.Bool("p", false, "enable parallel portfolio mode")
	chosenConfig := flag.Int("c", 0, "choose config index (0 to 3) when not in parallel mode")
	flag.Parse()

	if *cnfFile == "" || (!*parallelMode && (*chosenConfig < 0 || *chosenConfig >= len(configs))) {
		log.Fatalf("Usage: %s -f <cnf_file> [-verbose -p | -c <config_index>]", os.Args[0])
	}

	nvars, formula, err := parseDIMACS(*cnfFile)
	if err != nil {
		log.Fatalf("CNF parsing error: %v\n", err)
		return
	}

	start := time.Now()

	var sat bool
	var finalState *SolverState
	var configName string

	if *parallelMode {
		ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
		defer cancel()
		sat, finalState, configName = runPortfolioSolver(ctx, nvars, formula, configs)

		if configName == "" {
			fmt.Printf("Filename: %s\tResult: TIMEOUT\n", *cnfFile)
			return
		}
	} else {
		h := NewVSIDSHeuristic(nvars, configs[*chosenConfig])
		h.Init(formula)

		s := NewSolverState(nvars)
		sat, finalState = dpll(formula, s, h)
		configName = configs[*chosenConfig].Name
	}

	elapsed := time.Since(start)

	var ans string
	if sat {
		ans = "SAT"
	} else {
		ans = "UNSAT"
	}

	fmt.Printf("Filename: %s\tResult: %s\tElapsed time: %s\tFastest Config: %s\n",
		*cnfFile, ans, elapsed, configName)

	if *verboseMode && sat {
		printAssignments(nvars, *finalState)
	}
}
