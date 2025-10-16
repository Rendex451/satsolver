package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

// Конфигурации для портфеля
var portfolioConfigs = []SolverConfig{
	{"VSIDS-Fast", 0.95, 1.05, "aggressive", true},
	{"VSIDS-Slow", 0.99, 1.01, "conservative", false},
	{"Random", 0.0, 1.0, "none", false},
	{"DLIS", 0.0, 1.0, "none", true},
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
	cnfFile := flag.String("f", "example.cnf", "DIMACS file")
	verboseMode := flag.Bool("v", false, "enable verbose mode")
	flag.Parse()

	//if flag.NArg() < 1 {
	//	log.Fatalf("Usage: %s -f <cnf_file> [-verbose]", os.Args[0])
	//}

	nvars, formula, err := parseCNF(*cnfFile)
	if err != nil {
		log.Fatalf("CNF parsing error: %v\n", err)
		return
	}
	if *verboseMode {
		fmt.Printf("Parsed %d variables and %d clauses\n", nvars, len(formula))
	}

	h := NewVSIDSHeuristic(nvars)
	h.Init(formula)

	start := time.Now()
	s := NewSolverState(nvars)
	sat, finalS := dpll(formula, s, h)

	elapsed := time.Since(start)

	var ans string
	if sat {
		ans = "SAT"
	} else {
		ans = "UNSAT"
	}

	fmt.Printf("Filename: %s\tResult: %s\tElapsed time: %s\n",
		*cnfFile, ans, elapsed)

	if *verboseMode && sat {
		printAssignments(nvars, *finalS)
	}
}
