package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Black Breath — Sorcery {2}{B}:
//
//	"Creatures your opponents control get -1/-1 until end of turn. The
//	 Ring tempts you."
//
// The creatures are fixed as the spell resolves (CR 611.2c): one that
// enters later this turn is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0afbb32-dc50-436a-bd24-2990c103a4f8",
		Name:         "The Black Breath",
		Purpose:      effects.WithRiders(game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}, game.Purpose{RingTempts: 1}),
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.OpponentControls()),
				Power:     -1,
				Toughness: -1,
				Label:     "The Black Breath — -1/-1 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
