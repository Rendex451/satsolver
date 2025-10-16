package solver

import "fmt"

func PrintAssignments(nvars int, finalS SolverState) {
	fmt.Print("Assignments: ")
	for v := 1; v <= nvars; v++ {
		if finalS.assigned[v] {
			val := finalS.assignment[v] == 1
			fmt.Printf("%d=%t ", v, val)
		}
	}
	fmt.Println()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
