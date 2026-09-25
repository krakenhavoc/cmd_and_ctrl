package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cream of the Crop — Enchantment {1}{G}:
//
//	"Whenever a creature you control enters, you may look at the top X
//	 cards of your library, where X is that creature's power. If you do,
//	 put one of those cards on top of your library and the rest on the
//	 bottom of your library in any order."
//
// #1298, on ADR 0088's put_in_library (2026-09-23 amendment): an EXACT
// count on the top lane. A `top_or_bottom` prompt with TopCount 1 —
// the answer must hold exactly one card on top, and the rest go under
// in the order given. X of 1 has no choice left (one card, on top) and
// raises no prompt; X of 0 or less looks at nothing.
//
// X is read at resolution (CR 608.2h) through the trigger's own event:
// the entered creature's power NOW while it is still on the battlefield,
// so a pump in response counts, and its last-known power — counters
// included — once it has left (#1379). A creature that left and came
// back is a new object, and X is still the power of the one that
// entered.
func init() {
	Register(Spec{
		OracleID:     "b61a87e3-dc98-4db9-abed-b47b677d81ab",
		Name:         "Cream of the Crop",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Cream of the Crop — look at the top X cards of your library",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Cream of the Crop — look at the top X cards of your library, where X is that creature's power?",
			},
			Effect: creamOfTheCropLook,
		}},
	})
}

// creamOfTheCropLook is the trigger's effect. It captures nothing: the
// entered creature is the trigger's event object, carried on the item.
func creamOfTheCropLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := 0
	if entered, ok := ctx.TriggeringPermanent(); ok {
		x = entered.Power
	}
	return LookAtLibraryThenPlace{
		N:         x,
		Placement: game.LibraryPlaceTopOrBottom,
		TopCount:  1,
		Label:     "Cream of the Crop — put one of those cards on top and the rest on the bottom in any order",
	}.Apply(ctx)
}
