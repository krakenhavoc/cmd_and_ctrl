package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Genesis Hydra — Creature — Plant Hydra {X}{G}{G}, 0/0 (#5368):
//
//	"When you cast this spell, reveal the top X cards of your library. You
//	 may put a nonland permanent card with mana value X or less from among
//	 them onto the battlefield. Then shuffle the rest into your library.
//	 This creature enters with X +1/+1 counters on it."
//
// The cast trigger is Annihilator's WhenYouCastThisSpell (it resolves even
// if the Hydra is countered), with X read off the spell as the trigger goes
// on the stack and carried on the item. The reveal and the optional put are
// Lonis, Cryptozoologist's (RevealTopOfLibraryForEffect, then
// PutFromLibraryOntoBattlefield over the revealed cards): nonland permanent
// cards only, mana value X or less. The rest stay in the library, which is
// then shuffled. The entry counters are the CR 614.1c clause (XCounters).
//
// No simplification.
func init() {
	const label = "Genesis Hydra — reveal the top X cards; you may put a nonland permanent card with mana value X or less onto the battlefield; shuffle the rest into your library"
	cast := WhenYouCastThisSpell(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		x := item.Params.Amount
		if x <= 0 {
			return nil
		}
		revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), x,
			"Genesis Hydra — revealed from the top of the library")
		return PutFromLibraryOntoBattlefield{
			Player:   ctx.Controller(),
			Cards:    revealed,
			Match:    And(Nonland(), ManaValueLE(x)),
			Max:      1,
			Optional: true,
			Label:    "Genesis Hydra — you may put a nonland permanent card with mana value X or less onto the battlefield",
			ThenTail: cmdr2B23ShuffleTheRestTail,
		}.Apply(ctx)
	})
	build := cast.Build
	cast.Build = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
		item := build(ev, source, lki, g)
		if x, ok := g.SpellXValueForEffect(ev.CardID); ok {
			item.Params.Amount = x
		}
		return item
	}
	Register(Spec{
		OracleID:                   "4345c359-b9d6-43c8-997f-e00746e6e318",
		Name:                       "Genesis Hydra",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered:                  []game.TriggeredAbility{cast},
	})
}
