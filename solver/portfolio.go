package solver

import (
	"context"
)

type Result struct {
	sat        bool
	state      *SolverState
	configName string
}

func RunPortfolioSolver(ctx context.Context, nvars int, formula Formula, configs []VSIDSConfig) (bool, *SolverState, string) {
	resultChan := make(chan Result, len(configs))

	for _, config := range configs {
		go func(cfg VSIDSConfig) {
			h := NewVSIDSHeuristic(nvars, cfg)
			h.Init(formula)
			s := NewSolverState(nvars)
			sat, finalState := Dpll(formula, s, h)
			resultChan <- Result{sat: sat, state: finalState, configName: cfg.Name}
		}(config)
	}

	select {
	case res := <-resultChan:
		return res.sat, res.state, res.configName
	case <-ctx.Done():
		return false, nil, ""
	}
}
