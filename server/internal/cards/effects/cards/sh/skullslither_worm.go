package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullslither Worm — {3}{B} Creature — Worm 3/3 (Jumpstart 2022, #5948):
//
//	"When this creature enters, each opponent discards a card. For each
//	 opponent who can't, put two +1/+1 counters on this creature."
//
// An opponent with a card in hand picks it and discards it; each
// opponent whose hand is empty (CR 701.9b) is one who can't, and the Worm
// gets two counters for each of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a36dda2e-4cb8-4821-9c93-bd637ae7b333",
		Name:         "Skullslither Worm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skullslither Worm — each opponent discards a card; two +1/+1 counters for each who can't",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					cant := 0
					for _, opp := range ctx.Opponents() {
						p := g.PlayerByIDForEffect(opp)
						if p == nil || p.Eliminated {
							continue
						}
						if p.Hand == nil || p.Hand.Size() == 0 {
							cant++
							continue
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
					}
					if cant == 0 || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 2 * cant}.Apply(ctx)
				}),
		},
	})
}
