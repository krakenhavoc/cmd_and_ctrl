package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stocking the Pantry — Enchantment {G} (Bloomburrow, #4467):
//
//	"Whenever you put one or more +1/+1 counters on a creature you
//	 control, put a supply counter on this enchantment.
//	 {2}, Remove a supply counter from this enchantment: Draw a card."
//
// Casey Jones's trigger: WheneverYouPutCountersOfKindOnACreature over a
// creature you control, once per creature per batch (CR 603.2c), so
// counters put on two creatures at once are two supply counters, and
// counters you put on an opponent's creature or an opponent puts on yours
// are none. The draw is Staff of the Storyteller's counter-removal row;
// the removal is a cost, paid as the ability is activated (CR 602.2b,
// 601.2h), and the draw answers nothing.
//
// No simplification.
func init() {
	const supply = "supply"
	Register(Spec{
		OracleID:     "068b370f-a358-4244-8050-6c81680c55c7",
		Name:         "Stocking the Pantry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne, creatureYouControl,
				"Stocking the Pantry — put a supply counter on it", putACounterOnThis(supply)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Remove a supply counter from this enchantment: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), RemoveCountersFromThis(supply, 1)),
			Effect:  b27DrawOne,
		}},
	})
}
