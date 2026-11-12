package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skull Fracture — Sorcery {B} (Odyssey, #5613):
//
//	"Target player discards a card.
//	 Flashback {3}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The discarding player picks the card (CR 701.9a): the shared
// targetPlayerDiscards body, which does nothing when the player is no
// longer a legal target. Flashback is the shared constructor (graveyard
// cast, price, and CR 702.34a's exile as the spell leaves the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3384606b-7724-4584-89fa-fbe006d5de1f",
		Name:             "Skull Fracture",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{B}")},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			return targetPlayerDiscards(1)(ctx.Game, ctx.Item)
		})),
	})
}
