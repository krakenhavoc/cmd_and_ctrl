package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baral and Kari Zev — Legendary Creature — Human {1}{U}{R}, 2/4:
//
//	"First strike, menace
//	 Whenever you cast your first instant or sorcery spell each turn, you
//	 may cast a spell with lesser mana value that shares a card type with
//	 it from your hand without paying its mana cost. If you don't, create
//	 First Mate Ragavan, a legendary 2/1 red Monkey Pirate creature token.
//	 It gains haste until end of turn."
//
// The trigger is Arcane Bombardment's (the first-instant-or-sorcery tally).
// The free cast is the resolution_cast offer over the hand cards with a
// spell of lesser mana value (the triggering spell's, X counted, CR 202.3e)
// that has at least one card type in common with it (PermissionFilter.
// CardTypesAny, CR 205.2a), each judged on the face being cast. "If you
// don't" is the offer's continuation: it runs when the prompt is answered
// with nothing cast (declined, or no card was offered at all) and makes
// the Monkey Pirate, whose haste lasts until end of turn only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "386a6fa7-db5b-4e82-bd25-fd86d910ed0c",
		Name:            "Baral and Kari Zev",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				AllOf(YouCast(Or(Instant(), Sorcery())), youCastYourFirstInstantOrSorceryEachTurn),
				"Baral and Kari Zev — you may cast a spell with lesser mana value that shares a card type from your hand free; if you don't, create First Mate Ragavan",
				func(g *game.Game, item *game.StackItem) error {
					return cmdrB21BaralOffer(g, item)
				}),
		},
	})
}
