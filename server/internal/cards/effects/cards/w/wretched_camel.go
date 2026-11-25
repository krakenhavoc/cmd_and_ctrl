package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wretched Camel — {1}{B} Creature — Zombie Camel 2/1 (Hour of Devastation):
//
//	"When this creature dies, if you control a Desert or there is a Desert card
//	 in your graveyard, target player discards a card."
//
// The "if" is an intervening condition (CR 603.4): it is checked as the
// creature dies, with the creature already in the graveyard, and again as the
// trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "525d1562-4fb6-4df3-ab6d-ab6a94c866e0",
		Name:         "Wretched Camel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIfYou("you control a Desert or there is a Desert card in your graveyard", effects.OldB240YouHaveADesert,
					effects.WhenThisDies("Wretched Camel — target player discards a card",
						func(g *game.Game, item *game.StackItem) error {
							p, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
							if !ok {
								return nil
							}
							return effects.DiscardOne(g, item, p)
						})),
				effects.TargetPlayer("target player")),
		},
	})
}
