package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodvine Elemental — Creature — Elemental {4}{G}{W}, 4/4 (Conspiracy, S82
// OLD batch 364, #5557):
//
//	"Trample
//	 Parley — Whenever this creature attacks, each player reveals the top
//	 card of their library. For each nonland card revealed this way,
//	 attacking creatures you control get +1/+1 until end of turn. Then each
//	 player draws a card."
//
// Cmdr2S49Parley with the boost as the reward: +N/+N for the N nonland cards,
// given once, as the ability resolves, to the creatures its controller controls
// that are attacking then (CR 611.2c); a creature that joins combat later gets
// nothing.
//
// No simplification.
func init() {
	const label = "Woodvine Elemental — each player reveals the top card of their library; attacking creatures you control get +1/+1 for each nonland card; then each player draws a card"
	effects.Register(effects.Spec{
		OracleID:        "b265ac18-3923-46ba-b7b6-4190855815e1",
		Name:            "Woodvine Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Cmdr2S49Parley(g, ctx.Source(), label, func(nonland int) error {
					if nonland == 0 {
						return nil
					}
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.AttackingCreature(), effects.YouControl()),
						Power:     nonland,
						Toughness: nonland,
						Label:     "Woodvine Elemental — attacking creatures get +1/+1 for each nonland card",
					}.Apply(ctx)
				})
			}),
		},
	})
}
