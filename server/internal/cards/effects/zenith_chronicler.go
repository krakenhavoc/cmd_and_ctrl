package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zenith Chronicler — Artifact Creature — Phyrexian Construct {2},
// 3/1 (EDHREC rank 2683):
//
//	"Whenever a player casts their first multicolored spell each turn,
//	 each other player draws a card."
//
// The symmetrical gold-hoser: a two-mana 3/1 that hands the rest of
// the table a card whenever anyone leads with a gold spell. The
// trigger is any player's cast (b25FirstMulticoloredSpellThisTurn)
// — the spell on the stack has two or more colours, and no earlier
// cast by that player this turn did. The engine's per-turn tally
// counts creature and noncreature casts only, so the "first
// multicolored" is read off the event log back to the start of the
// turn, the same walk Dionus makes for a first tap. "Each other
// player" is every seated player but the caster — the Chronicler's
// controller draws when an opponent casts, and the opponents draw
// when the controller does, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fed6494-3bdd-4bc6-901f-da5fb623b152",
		Name:         "Zenith Chronicler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			Key:     "Zenith Chronicler — each other player draws a card",
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b25FirstMulticoloredSpellThisTurn(ev, g)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b25EachPlayerExceptDraws(NewContext(g, item), item.Trigger.Event.Actor, 1)
			},
		}},
	})
}
