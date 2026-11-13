package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague Dogs — Creature — Phyrexian Zombie Dog {4}{B}, 3/3 (Urza's
// Destiny, #5621):
//
//	"When this creature dies, all creatures get -1/-1 until end of turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// The dies trigger shrinks every creature on the battlefield as it
// resolves, every player's (CR 611.2c); one that enters later this turn
// is not affected. Declared a -1/-1 sweep of creatures. The second row
// is the shared udsB2SacrificeThisDrawRow, and sacrificing the Dogs
// that way fires the dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7cc69fd-0c67-44fc-be9d-d8f5b1bd54f5",
		Name:         "Plague Dogs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisDies("Plague Dogs — all creatures get -1/-1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: Creature(), Power: -1, Toughness: -1, Label: "Plague Dogs — -1/-1"}.Apply(NewContext(g, item))
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}}),
		},
		Activated: []ActivatedAbility{udsB2SacrificeThisDrawRow("creature")},
	})
}
