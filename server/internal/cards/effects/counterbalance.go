package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Counterbalance — Enchantment {U}{U}:
//
//	"Whenever an opponent casts a spell, you may reveal the top card of your
//	 library. If you do, counter that spell if it has the same mana value as
//	 the revealed card."
//
// An optional cast trigger on each opponent's spell (CR 603.2). The reveal is
// the controller's choice (the trigger's own prompt); the card stays on top
// of the library, and an empty library reveals nothing, which counters
// nothing. Mana values are the engine's (ManaValueForEffect): the spell on
// the stack counts the X it was cast with (CR 202.3e), a land or a card with
// no mana cost is 0. A spell that has already left the stack by resolution is
// not countered (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "088dfd02-c0e0-41d1-9923-efe023fb2ad1",
		Name:         "Counterbalance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.CardID != uuid.Nil && ev.Actor != uuid.Nil && ev.Actor != source.Controller
			}, "Counterbalance — reveal the top card of your library; counter that spell if it has the same mana value", counterbalanceResolve),
				"Reveal the top card of your library? If you do, that spell is countered if it has the same mana value."),
		},
	})
}

func counterbalanceResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spellID := ctx.Trigger().Event.CardID
	spell, ok := g.LookupCardForEffect(spellID)
	if !ok || g.StackItemForEffect(spellID) == nil {
		return nil
	}
	revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 1, "Counterbalance")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	spellMV, ok1 := g.ManaValueForEffect(spell)
	topMV, ok2 := g.ManaValueForEffect(top)
	if !ok1 || !ok2 || spellMV != topMV {
		return nil
	}
	return CounterTarget{StackID: spellID}.Apply(ctx)
}
