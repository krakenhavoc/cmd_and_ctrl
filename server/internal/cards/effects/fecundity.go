package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fecundity — Enchantment {2}{G}:
//
//	"Whenever a creature dies, that creature's controller may draw a
//	 card."
//
// Any creature, anyone's: the trigger sits on Fecundity's controller's
// side of the stack, but the "you may" and the draw both belong to the
// dead creature's CONTROLLER — the controller it last had on the
// battlefield (CR 603.10a), which for a stolen creature is the thief,
// not the owner. Both the prompt (OptionalPrompt.Chooser) and the draw
// read it off the event. "Dies" is graveyard-only, so a bounced or
// exiled creature draws nothing, and a token does draw (it died).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffa64ac6-fe55-48b8-b015-849982cd7ad8",
		Name:         "Fecundity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			},
			Key: "Fecundity — that creature's controller may draw a card",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Fecundity — that creature's controller may draw a card")
				if dead, ok := g.LookupCardForEffect(ev.CardID); ok {
					item.Params.Player = leftUnderControlOf(ev, dead)
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Params.Player, N: 1}.Apply(NewContext(g, item))
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Fecundity — draw a card?",
				Chooser: func(ev game.Event, _ *game.Card, g *game.Game) uuid.UUID {
					dead, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return uuid.Nil
					}
					return leftUnderControlOf(ev, dead)
				},
			},
		}},
	})
}
