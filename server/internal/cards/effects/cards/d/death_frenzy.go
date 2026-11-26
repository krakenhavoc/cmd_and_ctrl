package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death Frenzy — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Sorcery {3}{B}{G}: All creatures get -2/-2 until end of turn. Whenever a creature
//	dies this turn, you gain 1 life.
//
// The life-gain is a delayed trigger that lasts the turn and repeats, set up
// as the spell resolves; the creatures the -2/-2 kills trigger it too (their
// deaths happen after the spell has finished resolving).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0842cc1d-de1e-4fb7-a54b-78a4c8f40d3e",
		Name:         "Death Frenzy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Death Frenzy — -2/-2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DelayedOnEvent{
				Label:     "Death Frenzy — a creature died; you gain 1 life",
				On:        []game.EventKind{game.EventLTB},
				Condition: effects.ACreatureDiedCondition,
				Body:      deathFrenzyGainOne,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}

var deathFrenzyGainOne = game.SimpleDelayedBody("death-frenzy/gain-one-life", func(g *game.Game, item *game.StackItem) error {
	return effects.GainLife{Amount: 1}.Apply(effects.NewContext(g, item))
})
