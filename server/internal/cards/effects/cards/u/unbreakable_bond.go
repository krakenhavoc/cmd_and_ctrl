package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unbreakable Bond — Sorcery {4}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Return target creature card from your graveyard to the battlefield with
//	 a lifelink counter on it."
//
// The counter is part of the entry (CR 122.1b), so the creature never exists
// without lifelink. A target that has left the graveyard fizzles the spell
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36cfdbd6-f971-4ee5-aa1b-fa3ddad5b4b1",
		Name:         "Unbreakable Bond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, false, map[string]int{game.CounterLifelink: 1})
			return effects.NothingIfGone(err)
		})),
	})
}
