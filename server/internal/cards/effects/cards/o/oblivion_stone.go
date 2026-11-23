package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oblivion Stone — Artifact {3}:
//
//	"{4}, {T}: Put a fate counter on target permanent.
//	 {5}, {T}, Sacrifice this artifact: Destroy each nonland permanent
//	 without a fate counter on it, then remove all fate counters from all
//	 permanents."
//
// The sacrifice is a cost, so the Stone is in the graveyard before the
// wipe and is not in it. Destruction respects indestructible; the
// counters come off every permanent that still has any, in a second step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fd61c9f-44d2-48d7-a271-8e4208407cca",
		Name:         "Oblivion Stone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}, {T}: Put a fate counter on target permanent.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				Targets: effects.TargetPermanent("target permanent"),
				Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.AddCounter{Target: id, Kind: effects.RprB15FateCounter, N: 1}.Apply(ctx)
				})),
			},
			{
				Label:   "{5}, {T}, Sacrifice this artifact: Destroy each nonland permanent without a fate counter on it, then remove all fate counters from all permanents.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.RprB15DestroyUnfatedThenClear,
			},
		},
	})
}
