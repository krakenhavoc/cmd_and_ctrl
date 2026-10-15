package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rashmi, Eternities Crafter — Legendary Creature — Elf Druid {2}{G}{U},
// 2/3:
//
//	"Whenever you cast your first spell each turn, reveal the top card
//	 of your library. You may cast it without paying its mana cost if
//	 it's a spell with lesser mana value. If you don't cast it, put it
//	 into your hand."
//
// Ziatora's Envoy's offer (#3801, Game.OfferFreePlayFromTopOrHandForEffect)
// with a REVEAL and no land branch: every seat sees the card, a land is
// not a spell so it goes straight to the hand, and the cap is one below
// the cast spell's mana value. The cast spell's mana value is read as the
// trigger goes on the stack (CR 202.3e counts X while it is on the
// stack) and carried on the item, so a counter in response does not
// change it. A spell of mana value 0 leaves no lesser value, so the
// revealed card goes to the hand.
//
// The cast is an ordinary cast_spell after the resolution rather than a
// frame inside it, the posture every resolution-time "you may cast" takes
// (ADR 0066, ADR 0091).
func init() {
	const label = "Rashmi, Eternities Crafter — reveal the top card, cast it free if lesser or put it into your hand"
	t := On(game.EventCast, rashmiFirstSpell, label,
		func(g *game.Game, item *game.StackItem) error {
			return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
				Player:       item.Controller,
				Source:       item.SourceCardID,
				SourceName:   "Rashmi, Eternities Crafter",
				MaxManaValue: item.Params.Amount - 1,
				Reveal:       true,
			})
		})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = spellManaValueForEffect(g, ev.CardID)
		return item
	}
	Register(Spec{
		OracleID:     "01a79429-7d57-4a51-9815-a20905ed4707",
		Name:         "Rashmi, Eternities Crafter",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// rashmiFirstSpell is "whenever you cast your first spell each turn": the
// engine's per-turn cast tally is bumped before EventCast fires.
func rashmiFirstSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		g.CastTallyFor(source.Controller).Total == 1
}
