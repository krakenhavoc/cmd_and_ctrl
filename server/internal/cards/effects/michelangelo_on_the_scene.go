package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Michelangelo, On the Scene — {4}{G}{G} Legendary Creature — Mutant
// Ninja Turtle 2/2 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Trample
//	 Michelangelo enters with a +1/+1 counter on him for each land you
//	 control.
//	 When Michelangelo dies, return this card to your hand."
//
// The counters are a CR 614.1c self-replacement, counted as he enters
// (Golgari Grave-Troll's shape): the lands his controller controls at
// that moment. The dies trigger is Endless Cockroaches': "this card" is
// the card he became in the graveyard, so it does nothing if that card has
// left the graveyard before the trigger resolves (CR 400.7). It returns
// to its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0e3feb8-2add-41c9-839c-587af6ca8f93",
		Name:            "Michelangelo, On the Scene",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return countControlled(g, src.Controller, func(c game.Card) bool { return c.IsLand() })
			}, "Michelangelo, On the Scene: enters with a +1/+1 counter per land you control"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Michelangelo, On the Scene — return this card to your hand", returnSelfFromGraveyardToHand),
		},
	})
}
