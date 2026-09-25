package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thousand-Year Storm — Enchantment {4}{U}{R} (EDHREC rank 1407):
//
//	"Whenever you cast an instant or sorcery spell, copy it for each
//	 other instant and sorcery spell you've cast before it this turn.
//	 You may choose new targets for the copies."
//
// Storm as an enchantment: the third spell of the turn is cast three
// times. The count is fixed when the trigger goes on the stack —
// "cast BEFORE it this turn" — and read off the event log
// (b12InstantsAndSorceriesCastBeforeThisTurn), where a copy is not a
// cast and does not count, as printed. The copies are the CR 707.10
// primitive: each is created separately, each with its own "you may
// choose new targets" prompt, and a spell countered in response to
// the trigger is simply not there to copy.
//
// No simplification.
const thousandYearStormLabel = "Thousand-Year Storm — copy the spell for each instant and sorcery cast before it this turn"

func init() {
	Register(Spec{
		OracleID:     "dd4cf149-2fae-40e5-b50b-639f6bcec65e",
		Name:         "Thousand-Year Storm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			Key:     thousandYearStormLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			},
			// A fill-in Build (ADR 0041 P9): the count is fixed when
			// the trigger goes on the stack, so it is a fact of the
			// moment the spell was cast, not something the Effect can
			// recompute later.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, thousandYearStormLabel)
				item.Params.Amount = b12InstantsAndSorceriesCastBeforeThisTurn(g, source.Controller, ev.CardID)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				count := item.Params.Amount
				if count <= 0 {
					return nil
				}
				return CopySpell{
					StackID:          item.Trigger.Event.CardID,
					Controller:       item.Controller,
					Count:            count,
					ChooseNewTargets: true,
					// CR 608.2h, #1255: "copy it" names the
					// spell without targeting it, so a spell
					// countered before this resolves is still
					// copied.
					FromLastKnown: true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
