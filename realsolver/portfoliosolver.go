package main

import (
	"context"
	"sync"
)

// SolverConfig - конфигурация солвера
type SolverConfig struct {
	Name          string
	DecayFactor   float64
	BumpFactor    float64
	RestartPolicy string
	Polarity      bool
}

func runPortfolioSolver(ctx context.Context, nvars int, formula Formula, configs []SolverConfig) (bool, *SolverState, string) {
	var wg sync.WaitGroup
	resultCh := make(chan bool, len(configs))
	stateCh := make(chan *SolverState, len(configs))

	for _, config := range configs {
		wg.Add(1)
		go func(cfg SolverConfig) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}

			h := NewVSIDSHeuristic(nvars)
			h.Init(formula)
			h.bumpFactor = cfg.BumpFactor

			s := NewSolverState(nvars)
			sat, finalS := dpll(formula, s, h)

			select {
			case resultCh <- sat:
				stateCh <- finalS
			case <-ctx.Done():
			}
		}(config)
	}

	// Ждем первый успешный результат или таймаут
	go func() {
		wg.Wait()
		close(resultCh)
		close(stateCh)
	}()

	for {
		select {
		case sat := <-resultCh:
			if sat {
				// Получаем состояние и завершаем
				finalS := <-stateCh
				return true, finalS, "portfolio"
			}
		case <-ctx.Done():
			return false, nil, "timeout"
		}
	}
}
