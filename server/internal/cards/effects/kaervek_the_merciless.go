package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaervek the Merciless — Legendary Creature — Human Shaman {5}{B}{R},
// 5/4 (EDHREC rank 2577):
//
//	"Whenever an opponent casts a spell, Kaervek deals damage equal to
//	 that spell's mana value to any target."
//
// The table's tax collector. b15OpponentCastSpell is the condition;
// the mana value is read off the stack as the trigger fires (CR
// 202.3e — game.(*Game).ManaValueForEffect, so a spell cast for X = 5 hits for its
// full cost) and captured in Build, because the spell may have
// resolved or been countered by the time the trigger does. The
// controller picks any target as the trigger goes on the stack;
// Kaervek is the damage source, so its colour is what a prevention
// check sees. A zero-mana-value spell (a land is not a spell; a
// Mox is) triggers and deals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7b72d38-0aa0-4e17-9dd5-9276d7cb21ec",
		Name:         "Kaervek the Merciless",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15OpponentCastSpell(ev, source)
			},
			Targets: TargetAny(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount := 0
				if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
					amount, _ = g.ManaValueForEffect(spell)
				}
				return game.NewTriggeredItem(source, "Kaervek the Merciless — deal damage equal to that spell's mana value to any target",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
