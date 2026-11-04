package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// nullGroupLootLabel is the attack trigger's stack label and its discard
// prompt.
const nullGroupLootLabel = "Null Group Biological Assets — you may discard a card; if you do, draw a card"

// Null Group Biological Assets — {2}{R} Creature — Mutant Mercenary 3/1
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"During your turn, this creature has first strike.
//	 Whenever this creature attacks, you may discard a card. If you do,
//	 draw a card."
//
// The first strike is Radha, Heart of Keld's grant to itself, gone on the
// opponents' turns. The attack trigger is Burning-Tree Vandal's rummage:
// the discard comes first and the draw happens only if a card was really
// discarded, so declining draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49598fd5-2aad-4974-b499-38b60fc371b6",
		Name:         "Null Group Biological Assets",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{firstStrikeDuringYourTurn()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(nullGroupLootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, nullGroupLootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
