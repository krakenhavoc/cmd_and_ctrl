package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hazel's Nocturne — Instant {3}{B} (Bloomburrow, #4467):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand. Each opponent loses 2 life and you gain 2 life."
//
// Pull from the Grave's returns (up to two, so a spell with no target
// still resolves), then Acolyte of Aclazotz's drain at 2. When every
// chosen target has left, the spell does nothing at all (CR 608.2b), the
// drain included; with none chosen it still drains.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a74a055-ff7f-434b-8dfb-d9fb72579cee",
		Name:         "Hazel's Nocturne",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.EachOpponentLosesLife{N: 2},
			effects.GainLife{Amount: 2},
		),
	})
}
