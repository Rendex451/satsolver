package solver

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

func ParseDIMACS(filename string) (int, Formula, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	var formula Formula
	var nvars int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") || strings.HasPrefix(line, "%") {
			continue
		}

		if strings.HasPrefix(line, "p cnf") {
			parts := strings.Fields(line)
			if len(parts) < 4 {
				return 0, nil, errors.New("invalid p cnf header")
			}
			nvars, err = strconv.Atoi(parts[2])
			if err != nil {
				return 0, nil, err
			}
			continue
		}

		numsStr := strings.Fields(line)
		var clause Clause
		for _, numStr := range numsStr {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			if num == 0 {
				break
			}
			clause = append(clause, Literal(num))
		}
		if len(clause) > 0 {
			formula = append(formula, clause)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}

	return nvars, formula, nil
}
