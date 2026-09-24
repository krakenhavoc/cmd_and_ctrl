package effects

import (
	"github.com/google/uuid"

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
// X is read at resolution — the entered creature's power NOW, so a
// pump in response counts. If the creature has left the battlefield by
// then, the engine keeps no last-known power for it past the trigger's
// own dispatch, so X falls back to its power when the ability
// triggered; that is the caveat.
func init() {
	Register(Spec{
		OracleID:     "b61a87e3-dc98-4db9-abed-b47b677d81ab",
		Name:         "Cream of the Crop",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If the creature leaves the battlefield before the ability resolves, X is its power from when the ability triggered.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Cream of the Crop — look at the top X cards",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Cream of the Crop — look at the top X cards of your library, where X is that creature's power?",
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				entered, power := ev.CardID, 0
				if c, ok := g.LookupCardForEffect(entered); ok {
					power = c.CurrentPower()
				}
				return game.NewTriggeredItem(source, "Cream of the Crop — look at the top X cards of your library",
					creamOfTheCropLook(entered, power))
			},
		}},
	})
}

// creamOfTheCropLook is the trigger's effect. It captures the entered
// creature's ID and its power when the ability triggered — scalars
// only, so an undo replays it.
func creamOfTheCropLook(entered uuid.UUID, powerAtTrigger int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		x := powerAtTrigger
		if c, ok := g.LookupCardForEffect(entered); ok {
			if z := g.FindCardZoneForEffect(entered); z != nil && z.Kind == game.ZoneBattlefield {
				x = c.CurrentPower()
			}
		}
		return LookAtLibraryThenPlace{
			N:         x,
			Placement: game.LibraryPlaceTopOrBottom,
			TopCount:  1,
			Label:     "Cream of the Crop — put one of those cards on top and the rest on the bottom in any order",
		}.Apply(NewContext(g, item))
	}
}
