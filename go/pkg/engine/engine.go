package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Engine defines the contract for simulation orchestrators.
type Engine interface {
	Run(ctx context.Context, gm types.Entity, players []types.Entity, premise string, maxSteps int) error
}

// New constructs an Engine implementation matching settings.Engine.
func New(settings types.EngineSettings, bus *types.EventBus, checkpointPath string) Engine {
	switch settings.Engine {
	case types.EngineSimultaneous:
		return &Simultaneous{
			Bus:            bus,
			MinSteps:       settings.MinSteps,
			CheckpointPath: checkpointPath,
		}
	case types.EngineAsynchronous:
		return &Asynchronous{
			Bus:      bus,
			MinSteps: settings.MinSteps,
		}
	default:
		return &Sequential{
			Bus:            bus,
			MinSteps:       settings.MinSteps,
			Strategy:       settings.Strategy,
			CheckpointPath: checkpointPath,
		}
	}
}

// SelectNextGM queries currentGM for which Game Master should govern the next step
// when multiple GMs are configured.
func SelectNextGM(ctx context.Context, currentGM types.Entity, gms []types.Entity) (types.Entity, error) {
	if len(gms) <= 1 {
		return currentGM, nil
	}

	gmNames := make([]string, len(gms))
	gmMap := make(map[string]types.Entity, len(gms))
	for i, g := range gms {
		gmNames[i] = g.Name()
		gmMap[g.Name()] = g
	}

	spec := types.ActionSpec{
		CallToAction: fmt.Sprintf(
			"System: Available Game Masters (rule sets): %s. Current Game Master: %s. Which Game Master should govern the next step? Reply ONLY with the name.",
			strings.Join(gmNames, ", "),
			currentGM.Name(),
		),
		OutputType: types.OutputNextGameMaster,
		Options:    gmNames,
	}
	choice, err := currentGM.Act(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("selecting next game master: %w", err)
	}
	choice = strings.TrimSpace(choice)
	if next, ok := gmMap[choice]; ok {
		return next, nil
	}
	for _, name := range gmNames {
		if strings.Contains(choice, name) {
			return gmMap[name], nil
		}
	}
	return currentGM, nil
}

// IsSkipStep reports whether an actor selection response requests skipping player actions for the step.
func IsSkipStep(actorResponse string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(actorResponse))
	return cleaned == string(types.OutputSkipThisStep) || strings.Contains(cleaned, string(types.OutputSkipThisStep))
}
