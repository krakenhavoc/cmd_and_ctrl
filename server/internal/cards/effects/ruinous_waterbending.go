package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruinous Waterbending — Sorcery — Lesson {1}{B}{B}:
//
//	"As an additional cost to cast this spell, you may waterbend {4}.
//	 (While paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}.)
//	 All creatures get -2/-2 until end of turn. If this spell's
//	 additional cost was paid, whenever a creature dies this turn, you
//	 gain 1 life."
//
// The waterbend is Secret of Bloodbending's optional additional cost
// (OptionalWaterbend): chosen, it charges {4} that the caster's untapped
// artifacts and creatures may each pay {1} of. The -2/-2 is a sweep over
// the creatures on the battlefield as it resolves (CR 611.2c), and when
// the cost was paid the spell also sets up a "whenever … this turn"
// delayed trigger (CR 603.7b) — Massacre Girl's: every creature that
// dies for the rest of the turn, anyone's, gains its caster 1 life. The
// creatures the sweep kills die after the spell has resolved, so each
// of them counts.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:      "7d3a0216-871b-4c1b-adb1-de99b832f577",
		Name:          "Ruinous Waterbending",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalWaterbend("{4}")},
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2, Label: "Ruinous Waterbending — -2/-2"}).Apply(ctx); err != nil {
				return err
			}
			if ctx.OptionalCostTimes(game.WaterbendKey) == 0 {
				return nil
			}
			return DelayedOnEvent{
				Label:     "Ruinous Waterbending — a creature died; you gain 1 life",
				On:        []game.EventKind{game.EventLTB},
				Condition: aCreatureDiedCondition,
				Body:      ruinousWaterbendingGainOne,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}

// ruinousWaterbendingGainOne is the delayed trigger's "you gain 1 life".
var ruinousWaterbendingGainOne = game.SimpleDelayedBody("ruinous-waterbending/gain-one-life", func(g *game.Game, item *game.StackItem) error {
	return GainLife{Amount: 1}.Apply(NewContext(g, item))
})
