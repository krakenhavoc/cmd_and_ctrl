package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fyndhorn Pollen — Enchantment {2}{G} (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep {1}
//	 All creatures get -1/-0.
//	 {1}{G}: All creatures get -1/-0 until end of turn."
//
// The static is a layer 7c modification over every creature, any controller's
// (B16Anthem over B29AllCreatures), so a creature that enters later is affected.
// The activated ability is a continuous effect from a resolving ability, so it
// affects the creatures on the battlefield as it resolves (CR 611.2c) and ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f18967d-c2cd-449d-a288-df6343672743",
		Name:         "Fyndhorn Pollen",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B29AllCreatures, -1, 0)},
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Fyndhorn Pollen — cumulative upkeep {1}", "{1}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: All creatures get -1/-0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Label: "Fyndhorn Pollen — -1/-0 until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
