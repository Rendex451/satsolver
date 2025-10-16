package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Clause представляет клаузу как срез литералов
type Clause []int

// Assignment представляет назначение переменным (map[int]bool)
type Assignment map[int]bool

func parseDIMACS(path string) (int, []Clause, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	var clauses []Clause
	var nvars int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем комментарии, пустые строки и строки с '0'
		if len(line) == 0 || strings.HasPrefix(line, "c") ||
			strings.HasPrefix(line, "%") || strings.HasPrefix(line, "0") {
			continue
		}

		// Читаем параметры p cnf nvars nclauses
		if strings.HasPrefix(line, "p") {
			parts := strings.Fields(line)
			if len(parts) >= 4 && parts[1] == "cnf" {
				nvars, err = strconv.Atoi(parts[2])
				if err != nil {
					return 0, nil, err
				}
			}
			continue
		}

		// Парсим клаузы
		numsStr := strings.Fields(line)
		var nums []int
		for _, numStr := range numsStr {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			nums = append(nums, num)
		}

		if len(nums) == 0 {
			continue
		}

		// Разбиваем на клаузы (каждая заканчивается 0)
		for i := 0; i < len(nums); {
			if nums[i] == 0 {
				i++
				continue
			}
			zeroIdx := -1
			for j := i; j < len(nums); j++ {
				if nums[j] == 0 {
					zeroIdx = j
					break
				}
			}
			if zeroIdx != -1 {
				clause := nums[i:zeroIdx]
				if len(clause) > 0 {
					clauses = append(clauses, clause)
				}
				i = zeroIdx + 1
			} else {
				// Последняя клауза без 0 в конце
				clause := nums[i:]
				if len(clause) > 0 {
					clauses = append(clauses, clause)
				}
				break
			}
		}
	}

	return nvars, clauses, scanner.Err()
}

func evaluateLiteral(assignment Assignment, literal int) bool {
	varValue, exists := assignment[abs(literal)]
	if !exists {
		return false // Не назначена
	}
	if literal > 0 {
		return varValue
	}
	return !varValue
}

func isClauseSatisfied(clause Clause, assignment Assignment) (bool, []int) {
	satisfied := false
	var unassigned []int

	for _, literal := range clause {
		if evaluateLiteral(assignment, literal) {
			satisfied = true
			break
		}
		varIdx := abs(literal)
		if _, exists := assignment[varIdx]; !exists {
			unassigned = append(unassigned, literal)
		}
	}

	return satisfied, unassigned
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func fullUnitPropagation(clauses []Clause, assignment *Assignment) bool {
	changed := true
	for changed {
		changed = false

		for i := range clauses {
			satisfied, unassigned := isClauseSatisfied(clauses[i], *assignment)
			if satisfied {
				continue
			}

			if len(unassigned) == 0 {
				// Конфликт
				return false
			}

			if len(unassigned) == 1 {
				// Unit clause
				literal := unassigned[0]
				varIdx := abs(literal)
				(*assignment)[varIdx] = literal > 0
				changed = true
				break // Перезапускаем проверку
			}
		}
	}
	return true
}

func allClausesSatisfied(clauses []Clause, assignment Assignment) bool {
	for _, clause := range clauses {
		satisfied, _ := isClauseSatisfied(clause, assignment)
		if !satisfied {
			return false
		}
	}
	return true
}

func copyAssignment(original Assignment) Assignment {
	copy := make(Assignment)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func dpll(clauses []Clause, assignment Assignment, nvars int) Assignment {
	// Создаем копию для модификации
	currentAssignment := copyAssignment(assignment)

	// Выполняем полную unit propagation
	if !fullUnitPropagation(clauses, &currentAssignment) {
		return nil
	}

	// Проверяем, все ли клаузы удовлетворены
	if allClausesSatisfied(clauses, currentAssignment) {
		return currentAssignment
	}

	// Выбираем непроназначенную переменную для ветвления
	for v := 1; v <= nvars; v++ {
		if _, exists := currentAssignment[v]; !exists {
			// Пробуем True
			trueAssignment := copyAssignment(currentAssignment)
			trueAssignment[v] = true
			result := dpll(clauses, trueAssignment, nvars)
			if result != nil {
				return result // Возвращаем успешное назначение
			}

			// Пробуем False
			falseAssignment := copyAssignment(currentAssignment)
			falseAssignment[v] = false
			result = dpll(clauses, falseAssignment, nvars)
			if result != nil {
				return result // Возвращаем успешное назначение
			}

			// Оба значения не работают
			return nil
		}
	}

	return nil
}

func printAssignment(assignment Assignment) {
	fmt.Print("Assignment: {")
	first := true
	for varIdx, value := range assignment {
		if !first {
			fmt.Print(", ")
		}
		fmt.Printf("%d=%t", varIdx, value)
		first = false
	}
	fmt.Println("}")
}

func main() {
	nvars, clauses, err := parseDIMACS("cnfs/sat/big/CBS_k3_n100_m449_b90_2.cnf")
	if err != nil {
		fmt.Printf("DIMACS parsing error: %v\n", err)
		return
	}

	start := time.Now()

	fmt.Printf("Vars: %d\n", nvars)
	fmt.Printf("Clauses: %d\n", len(clauses))

	res := dpll(clauses, make(Assignment), nvars)
	execTime := time.Since(start)

	if res != nil {
		fmt.Println("SAT")
		printAssignment(res)
	} else {
		fmt.Println("UNSAT")
	}

	fmt.Printf("Exec time: %v\n", execTime)
}
