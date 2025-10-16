package solver

import (
	"slices"
)

type Literal int
type Clause []Literal
type Formula []Clause

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

func (s *SolverState) Assign(varIdx int, value bool) {
	s.assignment[varIdx] = 1
	if !value {
		s.assignment[varIdx] = -1
	}
	s.assigned[varIdx] = true
}

func checkClauseValidity(formula Formula) bool {
	for _, clause := range formula {
		if len(clause) == 0 {
			return false
		}
	}
	return true
}

// evaluateLiteral проверяет корректность литерала с текущим присвоением
func evaluateLiteral(literal Literal, s *SolverState) bool {
	varIdx := abs(int(literal))
	if !s.assigned[varIdx] {
		return false
	}
	val := s.assignment[varIdx]
	return (literal > 0 && val == 1) || (literal < 0 && val == -1)
}

// isSatisfied проверяет удовлетворена ли формула текущим присвоением
// Для каждой клаузы хотя бы один литерал true.
func isSatisfied(formula Formula, s *SolverState) bool {
	for _, clause := range formula {
		satisfied := false
		for _, literal := range clause {
			if evaluateLiteral(literal, s) {
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

// unitPropagate:
// Копирует формулу.
// В цикле находит unit-клаузы (длина 1).
// Для каждой: Присваивает значение литералу.
// Фильтрует клаузы, содержащие этот литерал (они удовлетворены).
// Удаляет противоположный литерал (-literal) из оставшихся клауз.
// Повторяет, пока есть unit-клаузы.
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
			varIdx := abs(int(literal))
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

// pureLiteralAssignment:
// Копирует формулу.
// Собирает все уникальные литералы в set.
// Находит "чистые" литералы (те, чьи противоположности не встречаются).
// Присваивает им значение (true для положительных, false для отрицательных).
// Фильтрует клаузы, содержащие эти литералы (они удовлетворены).
func pureLiteralAssignment(formula Formula, s *SolverState) (Formula, *SolverState) {
	updatedFormula := slices.Clone(formula)

	allLiteralsSet := NewSet()
	for _, clause := range formula {
		for _, literal := range clause {
			allLiteralsSet.Add(literal)
		}
	}

	allLiterals := allLiteralsSet.Values()
	pureLiterals := NewSet()
	for _, literal := range allLiterals {
		if !slices.Contains(allLiterals, -literal) {
			pureLiterals.Add(literal)
		}
	}

	for _, literal := range pureLiterals.Values() {
		varIdx := abs(int(literal))
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

// Dpll:
// Базовые случаи: Пустая формула — SAT; Пустая клауза — UNSAT; Если удовлетворена — SAT.
// Вызывает unitPropagate и pureLiteralAssignment для упрощения.
// Вызывает h.Decay() перед выбором.
// Выбирает литерал с помощью VSIDS.
// Создаёт копию состояния s1, присваивает значение (true для selectedLiteral > 0).
// Упрощает формулу: Фильтрует клаузы с selectedLiteral, удаляет -selectedLiteral из остальных.
// Рекурсивно вызывает Dpll на s1.
// Если провал — bump противоположного литерала (увеличивает его приоритет для будущих выборов).
// Аналогично для противоположной ветви (s2, false).
// Если и вторая ветвь провал — bump оригинального литерала.
// Возвращает результат и финальное состояние.
func Dpll(formula Formula, s *SolverState, h *VSIDSHeuristic) (bool, *SolverState) {
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

	h.decay()

	selectedLiteral, err := h.selectLiteral(s)
	if err != nil {
		return false, s
	}

	var simplifiedFormula Formula
	s1 := NewSolverState(s.nvars)
	copy(s1.assignment, s.assignment)
	copy(s1.assigned, s.assigned)
	s1.Assign(abs(int(selectedLiteral)), selectedLiteral > 0)
	for _, clause := range newFormula {
		if !slices.Contains(clause, selectedLiteral) {
			updatedClause := slices.Clone(clause)
			if index := slices.Index(updatedClause, -selectedLiteral); index >= 0 {
				updatedClause = slices.Delete(updatedClause, index, index+1)
			}
			simplifiedFormula = append(simplifiedFormula, updatedClause)
		}
	}

	result, finalS := Dpll(simplifiedFormula, s1, h)
	if result {
		return result, finalS
	}

	h.bump(-selectedLiteral)

	simplifiedFormula = make(Formula, 0)
	s2 := NewSolverState(s.nvars)
	copy(s2.assignment, s.assignment)
	copy(s2.assigned, s.assigned)
	s2.Assign(abs(int(selectedLiteral)), selectedLiteral < 0)
	for _, clause := range newFormula {
		if !slices.Contains(clause, -selectedLiteral) {
			updatedClause := slices.Clone(clause)
			if index := slices.Index(updatedClause, selectedLiteral); index >= 0 {
				updatedClause = slices.Delete(updatedClause, index, index+1)
			}
			simplifiedFormula = append(simplifiedFormula, updatedClause)
		}
	}

	result, finalS = Dpll(simplifiedFormula, s2, h)
	if !result {
		h.bump(selectedLiteral)
	}
	return result, finalS
}
