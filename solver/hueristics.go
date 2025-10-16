package solver

import "errors"

// VSIDS эвристика для выбора литерала
type VSIDSHeuristic struct {
	posActivity      []float64
	negActivity      []float64
	bumpInc          float64
	bumpFactor       float64
	RescaleThreshold float64
	RescaleFactor    float64
}

type VSIDSConfig struct {
	Name             string
	InitialBumpInc   float64
	DecayFactor      float64
	RescaleThreshold float64
	RescaleFactor    float64
}

func NewVSIDSHeuristic(nvars int, config VSIDSConfig) *VSIDSHeuristic {
	return &VSIDSHeuristic{
		posActivity:      make([]float64, nvars+1),
		negActivity:      make([]float64, nvars+1),
		bumpInc:          config.InitialBumpInc,
		bumpFactor:       1.0 / config.DecayFactor,
		RescaleThreshold: config.RescaleThreshold,
		RescaleFactor:    config.RescaleFactor,
	}
}

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

func (h *VSIDSHeuristic) decay() {
	h.bumpInc *= h.bumpFactor
	if h.bumpInc > 1e100 {
		h.rescale()
	}
}

func (h *VSIDSHeuristic) rescale() {
	for i := 1; i < len(h.posActivity); i++ {
		h.posActivity[i] /= h.RescaleFactor
		h.negActivity[i] /= h.RescaleFactor
	}
	h.bumpInc /= h.RescaleFactor
}

func (h *VSIDSHeuristic) bump(lit Literal) {
	if lit > 0 {
		h.posActivity[int(lit)] += h.bumpInc
	} else {
		h.negActivity[int(-lit)] += h.bumpInc
	}
}

func (h *VSIDSHeuristic) selectLiteral(s *SolverState) (Literal, error) {
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
