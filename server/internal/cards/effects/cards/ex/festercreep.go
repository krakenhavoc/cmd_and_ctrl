package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festercreep — Creature — Elemental {1}{B}, 0/0 (Morningtide, #5592):
//
//	"This creature enters with a +1/+1 counter on it.
//	 {1}{B}, Remove a +1/+1 counter from this creature: All other
//	 creatures get -1/-1 until end of turn."
//
// District Mascot's entry counter and counter-removal cost, with D'Ghor
// Bloodletter's "all other creatures get -1/-1 until end of turn": the
// creatures are locked in as the ability resolves (CR 611.2c), and
// "other" is this object, so a Festercreep that left and came back is a
// new object the shrink reaches (CR 400.7). Spending its only counter
// leaves a 0/0 that dies to the state check before the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f7f841a-1913-4396-9c94-bbd49dfdb9d1",
		Name:         "Festercreep",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Festercreep: enters with a +1/+1 counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{B}, Remove a +1/+1 counter from this creature: All other creatures get -1/-1 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerRemove,
				Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true},
			},
			Cost:   effects.Plus(effects.ManaCost("{1}{B}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect: festercreepShrink,
		}},
	})
}

// festercreepShrink is "all other creatures get -1/-1 until end of turn".
func festercreepShrink(g *game.Game, item *game.StackItem) error {
	return effects.BoostUntilEOT{
		Match:     effects.And(effects.Creature(), effects.TrekB11OtherThanSourceObject(item)),
		Power:     -1,
		Toughness: -1,
		Label:     "Festercreep — -1/-1",
	}.Apply(effects.NewContext(g, item))
}
