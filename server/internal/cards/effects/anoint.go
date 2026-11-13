package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anoint — Instant {W} (Tempest, #5628):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Prevent the next 3 damage that would be dealt to target creature
//	 this turn."
//
// Mending Hands' charged shield (PreventNextDamage, CR 615.7) at 3 on a
// target creature, with Capsize's buyback (CR 702.27a). The shield is
// spent as damage is prevented: 2 damage leaves 1 point of it for the
// next event, and 5 damage gets 3 prevented and 2 through. It goes away
// in the cleanup step (CR 514.2), and it stays with the creature it was
// put on, not a new object the creature becomes (CR 400.7). A target
// gone in response means the spell doesn't resolve and goes to the
// graveyard, buyback paid or not (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "fcb03c54-5b9a-428d-bbf6-ef8977770c8d",
		Name:          "Anoint",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPrevent},
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return PreventNextDamage{Target: id, Amount: 3, Label: "Anoint: prevent the next 3 damage"}.Apply(ctx)
		})),
	})
}
