package main

import (
	"bufio"
	"errors"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Aki0x137/concurrent-sat-solver-go/set"
)

type Literal int
type Clause []Literal
type Formula []Clause

// SolverState для assignment с 3 состояниями
type SolverState struct {
	assignment []int  // 0=undefined, 1=true, -1=false (index=abs(var))
	assigned   []bool // true если назначено (index=abs(var))
	nvars      int
}

func NewSolverState(nvars int) *SolverState {
	return &SolverState{
		assignment: make([]int, nvars+1),
		assigned:   make([]bool, nvars+1),
		nvars:      nvars,
	}
}

// Assign присваивает значение
func (s *SolverState) Assign(varIdx int, value bool) {
	s.assignment[varIdx] = 1
	if !value {
		s.assignment[varIdx] = -1
	}
	s.assigned[varIdx] = true
}

// parseCNF парсит DIMACS CNF файл, возвращает nvars и формулу
func parseCNF(filename string) (int, Formula, error) {
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

func checkClauseValidity(formula Formula) bool {
	for _, clause := range formula {
		if len(clause) == 0 {
			return false
		}
	}
	return true
}

// VSIDSHeuristic для выбора литерала
type VSIDSHeuristic struct {
	posActivity []float64 // активность положительных литералов (index=varIdx)
	negActivity []float64 // активность отрицательных литералов (index=varIdx)
	bumpInc     float64   // текущее значение прироста
	bumpFactor  float64   // множитель для bumpInc (1 / decay)
}

func NewVSIDSHeuristic(nvars int) *VSIDSHeuristic {
	return &VSIDSHeuristic{
		posActivity: make([]float64, nvars+1),
		negActivity: make([]float64, nvars+1),
		bumpInc:     1.0,
		bumpFactor:  1.0 / 0.95, // примерно 1.0526
	}
}

// Init инициализирует scores на основе числа вхождений литералов
func (h *VSIDSHeuristic) Init(formula Formula) {
	for _, clause := range formula {
		for _, lit := range clause {
			if lit > 0 {
				h.posActivity[int(lit)] += 1.0
			} else {
				h.negActivity[int(-lit)] += 1.0
			}
		}
	}
}

// Decay имитирует decay путем увеличения bumpInc
func (h *VSIDSHeuristic) Decay() {
	h.bumpInc *= h.bumpFactor
	if h.bumpInc > 1e100 {
		h.rescale()
	}
}

// rescale для предотвращения переполнения
func (h *VSIDSHeuristic) rescale() {
	for i := 1; i < len(h.posActivity); i++ {
		h.posActivity[i] /= 1e100
		h.negActivity[i] /= 1e100
	}
	h.bumpInc /= 1e100
}

// Bump увеличивает score литерала
func (h *VSIDSHeuristic) Bump(lit Literal) {
	if lit > 0 {
		h.posActivity[int(lit)] += h.bumpInc
	} else {
		h.negActivity[int(-lit)] += h.bumpInc
	}
}

// SelectLiteral выбирает литерал с наивысшим score
func (h *VSIDSHeuristic) SelectLiteral(s *SolverState) (Literal, error) {
	maxScore := -1.0
	var bestLit Literal = 0
	for v := 1; v <= s.nvars; v++ {
		if s.assigned[v] {
			continue
		}
		posScore := h.posActivity[v]
		negScore := h.negActivity[v]
		if posScore > maxScore {
			maxScore = posScore
			bestLit = Literal(v)
		}
		if negScore > maxScore {
			maxScore = negScore
			bestLit = Literal(-v)
		}
	}
	if bestLit == 0 {
		return 0, errors.New("no literal found")
	}
	return bestLit, nil
}

// evaluateLiteral оценивает литерал
func evaluateLiteral(s *SolverState, literal Literal) bool {
	varIdx := int(math.Abs(float64(literal)))
	if !s.assigned[varIdx] {
		return false // Undefined считается false для satisfied check
	}
	val := s.assignment[varIdx]
	return (literal > 0 && val == 1) || (literal < 0 && val == -1)
}

// isSatisfied checks if all clauses are satisfied with the current assignment
func isSatisfied(formula Formula, s *SolverState) bool {
	for _, clause := range formula {
		satisfied := false
		for _, literal := range clause {
			if evaluateLiteral(s, literal) {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

// unitPropagate performs unit propagation on formula, based on current assignments
func unitPropagate(formula Formula, s *SolverState) (Formula, *SolverState) {
	updatedFormula := slices.Clone(formula)
	for {
		var unitClauses []Clause
		for _, clause := range updatedFormula {
			if len(clause) == 1 {
				unitClauses = append(unitClauses, clause)
			}
		}

		if len(unitClauses) == 0 {
			break
		}

		for _, clause := range unitClauses {
			literal := clause[0]
			varIdx := int(math.Abs(float64(literal)))
			s.Assign(varIdx, literal > 0)

			var filteredFormula Formula
			for _, c := range updatedFormula {
				if !slices.Contains(c, literal) {
					filteredFormula = append(filteredFormula, c)
				}
			}

			var simplifiedFormula Formula
			for _, c := range filteredFormula {
				updatedClause := slices.Clone(c)
				if index := slices.Index(updatedClause, -literal); index >= 0 {
					updatedClause = slices.Delete(updatedClause, index, index+1)
				}
				simplifiedFormula = append(simplifiedFormula, updatedClause)
			}
			updatedFormula = simplifiedFormula
		}
	}

	return updatedFormula, s
}

// pureLiteralAssignment checks for pure literals and updates formula.
func pureLiteralAssignment(formula Formula, s *SolverState) (Formula, *SolverState) {
	updatedFormula := slices.Clone(formula)

	allLiteralsSet := set.NewSet[Literal]()
	for _, clause := range formula {
		for _, literal := range clause {
			allLiteralsSet.Add(literal)
		}
	}

	allLiterals := allLiteralsSet.Values()
	pureLiterals := set.NewSet[Literal]()
	for _, literal := range allLiterals {
		if !slices.Contains(allLiterals, -literal) {
			pureLiterals.Add(literal)
		}
	}

	for _, literal := range pureLiterals.Values() {
		varIdx := int(math.Abs(float64(literal)))
		s.Assign(varIdx, literal > 0)

		var filteredFormula Formula
		for _, clause := range updatedFormula {
			if index := slices.Index(clause, literal); index == -1 {
				filteredFormula = append(filteredFormula, clause)
			}
		}
		updatedFormula = filteredFormula
	}

	return updatedFormula, s
}

func dpll(formula Formula, s *SolverState, h *VSIDSHeuristic) (bool, *SolverState) {
	if len(formula) == 0 {
		return true, s
	}

	if !checkClauseValidity(formula) {
		return false, s
	}

	if isSatisfied(formula, s) {
		return true, s
	}

	newFormula, s := unitPropagate(formula, s)

	newFormula, s = pureLiteralAssignment(newFormula, s)

	if isSatisfied(newFormula, s) {
		return true, s
	}

	if !checkClauseValidity(newFormula) {
		return false, s
	}

	// Decay перед выбором
	h.Decay()

	selectedLiteral, err := h.SelectLiteral(s)
	if err != nil {
		return false, s
	}

	var simplifiedFormula Formula
	s1 := NewSolverState(s.nvars) // Копируем state (для true)
	copy(s1.assignment, s.assignment)
	copy(s1.assigned, s.assigned)
	s1.Assign(int(math.Abs(float64(selectedLiteral))), selectedLiteral > 0)
	for _, clause := range newFormula {
		if !slices.Contains(clause, selectedLiteral) {
			updatedClause := slices.Clone(clause)
			if index := slices.Index(updatedClause, -selectedLiteral); index >= 0 {
				updatedClause = slices.Delete(updatedClause, index, index+1)
			}
			simplifiedFormula = append(simplifiedFormula, updatedClause)
		}
	}

	result, finalS := dpll(simplifiedFormula, s1, h)
	if result {
		return result, finalS
	}

	// Если true-ветвь провалилась, bump отрицательного литерала (т.е. предполагаем, что конфликт связан с этим)
	h.Bump(-selectedLiteral)

	simplifiedFormula = make(Formula, 0)
	s2 := NewSolverState(s.nvars) // Копируем для false
	copy(s2.assignment, s.assignment)
	copy(s2.assigned, s.assigned)
	s2.Assign(int(math.Abs(float64(selectedLiteral))), selectedLiteral < 0)
	for _, clause := range newFormula {
		if !slices.Contains(clause, -selectedLiteral) {
			updatedClause := slices.Clone(clause)
			if index := slices.Index(updatedClause, selectedLiteral); index >= 0 {
				updatedClause = slices.Delete(updatedClause, index, index+1)
			}
			simplifiedFormula = append(simplifiedFormula, updatedClause)
		}
	}
	result, finalS = dpll(simplifiedFormula, s2, h)
	if !result {
		// Если и false-ветвь провалилась, bump положительного литерала
		h.Bump(selectedLiteral)
	}
	return result, finalS
}
