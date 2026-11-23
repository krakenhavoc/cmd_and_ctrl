package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Errand-Rider of Gondor — Creature — Human Soldier {2}{W}, 3/2 (The Hobbit
// Eternal, HOC):
//
//	"When this creature enters, draw a card. Then if you don't control a
//	 legendary creature, put a card from your hand on the bottom of your
//	 library."
//
// Jace, Multiverse Architect's "draw, then put a card from your hand on the
// bottom" (rfPutCardFromHandOnBottom: the controller picks, the whole hand
// is the pick when it holds one card, an empty hand is not asked), behind
// the printed condition. The condition is read AFTER the draw, as the
// second sentence is (CR 608.2c), off the battlefield at that moment, so a
// legendary creature the controller controls as the trigger resolves
// counts, the Rider itself included if it is legendary (it is not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00749bad-dded-4fec-bf34-ff14cf561c26",
		Name:         "Errand-Rider of Gondor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Errand-Rider of Gondor — draw a card, then if you don't control a legendary creature, put a card from your hand on the bottom of your library",
				effects.Do(effects.DrawCards{N: 1}, effects.Instruction(errandRiderBottomUnlessLegendary))),
		},
	})
}

// errandRiderBottomUnlessLegendary is "Then if you don't control a
// legendary creature, put a card from your hand on the bottom of your
// library."
func errandRiderBottomUnlessLegendary(ctx *effects.Context) error {
	if len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.And(effects.Creature(), effects.Legendary()))) > 0 {
		return nil
	}
	return effects.RfPutCardFromHandOnBottom(ctx, "Errand-Rider of Gondor — put a card from your hand on the bottom of your library")
}
