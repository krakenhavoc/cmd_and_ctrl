package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mental Modulation — Instant {1}{U} (Edge of Eternities, #4461):
//
//	"This spell costs {1} less to cast during your turn.
//	 Tap target artifact or creature.
//	 Draw a card."
//
// The reduction is a self cost modifier (CR 601.2f) whose predicate asks
// whether the caster is the active player as the cost is determined, so
// on your turn the spell costs {U} and on anyone else's it costs {1}{U}.
// It spends against generic mana only, as every CostsLess does.
//
// The body is Hithlain Knots's: tap the target if it is still legal,
// then draw. A target that has become illegal fizzles the whole spell
// (CR 608.2b), so the card is not drawn either; a target that is already
// tapped is still legal, and the spell just draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "448d9391-9d15-4aa1-87b3-a1d755666c03",
		Name:         "Mental Modulation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "Mental Modulation — costs {1} less during your turn", func(q game.CostQuery) bool {
				return q.Game != nil && IsYourTurn(q.Game, q.Controller)
			}),
		},
		Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return TapTarget{Target: id}.Apply(ctx) }),
			DrawCards{N: 1},
		),
	})
}
