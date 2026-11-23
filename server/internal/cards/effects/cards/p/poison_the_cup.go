package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Poison the Cup — Instant {1}{B}{B} (Kaldheim, #5505):
//
//	"Destroy target creature. If this spell was foretold, scry 2.
//	 Foretell {1}{B}"
//
// Two steps in printed order (CR 608.2c): the destroy, then the scry,
// which happens only for a spell cast from a foretold card
// (SpellWasForetold, CR 702.143c) however it was paid for. The scry is
// its own step so a prompt the destroy raised is answered first.
// Foretell is the shared special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "58020778-554c-4af6-8484-0b90f64bc5d9",
		Name:           "Poison the Cup",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:        effects.TargetCreature("target creature"),
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{B}")},
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.SpellWasForetold() {
					return nil
				}
				return effects.Scry{N: 2}.Apply(ctx)
			}),
		),
	})
}
