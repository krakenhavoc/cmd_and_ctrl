package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague Dogs — Creature — Phyrexian Zombie Dog {4}{B}, 3/3 (Urza's
// Destiny, #5621):
//
//	"When this creature dies, all creatures get -1/-1 until end of turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Poison the Waters' -1/-1 to every creature, every player's, on a dies
// trigger: the creatures are read as it resolves, so one that arrives
// later in the turn is not affected (CR 611.2c), and the -1/-1 ends in the cleanup
// step (CR 514.2). A 1-toughness creature goes to its owner's graveyard
// as a state-based action (CR 704.5f). The trigger declares its Sweep
// (ADR 0126), a minus of 1 to all creatures. The draw row is Brass
// Secretary's: the sacrifice is part of the cost, so sacrificing the
// Dogs for a card also triggers the -1/-1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7cc69fd-0c67-44fc-be9d-d8f5b1bd54f5",
		Name:         "Plague Dogs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WhenThisDies("Plague Dogs — all creatures get -1/-1 until end of turn", plagueDogsShrinkAll),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}

// plagueDogsShrinkAll is "all creatures get -1/-1 until end of turn".
func plagueDogsShrinkAll(g *game.Game, item *game.StackItem) error {
	return BoostUntilEOT{Match: Creature(), Power: -1, Toughness: -1, Label: "Plague Dogs — all creatures get -1/-1"}.Apply(NewContext(g, item))
}
