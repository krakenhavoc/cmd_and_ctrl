package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siren's Call — Instant {U} (#4062):
//
//	"Cast this spell only during an opponent's turn, before attackers
//	 are declared.
//	 Creatures the active player controls attack this turn if able.
//	 At the beginning of the next end step, destroy all non-Wall
//	 creatures that player controls that didn't attack this turn. Ignore
//	 this effect for each creature the player didn't control
//	 continuously since the beginning of the turn."
//
// The cast window is a CastCondition: another player's turn, until the
// first declare attackers step begins. On resolution, every creature the
// active player controlled since the turn began is required to attack
// (CR 508.1d), and each non-Wall one gets a delayed destroy at the next
// end step unless it attacked (CR 603.7). A creature that came under
// that player's control this turn is ignored for both halves, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "269fc857-a052-4f0a-9759-467ccf42bebb",
		Name:         "Siren's Call",
		Completeness: CompletenessFull,
		// A wipe in the audit's eyes: it destroys the active player's
		// creatures, a class it spares by condition (Walls, creatures
		// that attacked, creatures that arrived this turn).
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true, Partial: true}},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return !IsYourTurn(g, controller) && !g.DeclareAttackersHasBegunThisTurnForEffect()
		},
		CastConditionLabel: "Cast this spell only during an opponent's turn, before attackers are declared.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			match := And(Creature(), ControlledByTheActivePlayer(), ControlledSinceTheBeginningOfTheTurn())
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !match(ctx.Game, ctx.Controller(), c) {
					continue
				}
				if err := mustAttackOrBeDestroyed(ctx, c.InstanceID, "Siren's Call"); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
