package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Infiltrator — Creature — Vedalken Ninja {2}{U}{B}, 2/3:
//
//	"Ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever a Ninja you control deals combat damage to a player, draw
//	 a card."
//
// The ninjutsu proof with the TRIBAL payoff, and the one that shows
// the keyword's own arrival paying itself off: the Infiltrator is a
// Ninja, so the body ninjutsu puts onto the battlefield attacking
// draws its own card when it connects.
//
// One trigger per damage EVENT rather than per batch — "whenever A
// Ninja" is per creature (CR 603.2c) — so two Ninja connecting draw
// two cards, and a double striker draws in both damage steps. The
// subtype is read post-layer off the damage source, so a changeling
// counts; a source already gone when the event is harvested does not
// fire, which is weaker than printed rather than stronger. Shared
// condition with Seafloor Oracle, one subtype over.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "354defd2-f63f-4a48-9fd4-2526b3878a84",
		Name:         "Ingenious Infiltrator",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{U}{B}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Ninja")
			}, "Ingenious Infiltrator — draw a card", func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
