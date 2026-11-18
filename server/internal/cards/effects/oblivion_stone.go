package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "2fd61c9f-44d2-48d7-a271-8e4208407cca",
		Name:         "Oblivion Stone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{4}, {T}: Put a fate counter on target permanent.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Targets: TargetPermanent("target permanent"),
				Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return AddCounter{Target: id, Kind: rprB15FateCounter, N: 1}.Apply(ctx)
				})),
			},
			{
				Label:   "{5}, {T}, Sacrifice this artifact: Destroy each nonland permanent without a fate counter on it, then remove all fate counters from all permanents.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
				Effect:  rprB15DestroyUnfatedThenClear,
			},
		},
	})
}
