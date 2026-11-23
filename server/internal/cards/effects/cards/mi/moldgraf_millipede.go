package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moldgraf Millipede — Creature — Insect Horror {4}{G}, 2/2 (Duskmourn
// Commander):
//
//	"When this creature enters, mill three cards, then put a +1/+1 counter
//	 on this creature for each creature card in your graveyard."
//
// The mill is MillCards and the count is read once it is done (the count
// step waits for any prompt the mill asks, CR 608.2c), so the three cards
// just milled are counted. The counters go on this creature only while it is
// still on the battlefield (CounterOnThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48edac92-0dd8-4274-bf1b-122243770540",
		Name:         "Moldgraf Millipede",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Moldgraf Millipede — mill three cards, then put a +1/+1 counter on it for each creature card in your graveyard",
				effects.Do(effects.MillCards{N: 3}, effects.Instruction(func(ctx *effects.Context) error {
					n, _ := effects.Cmdr2S30CreatureCardsIn(ctx.Game, ctx.Controller())
					if n == 0 {
						return nil
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}))),
		},
	})
}
