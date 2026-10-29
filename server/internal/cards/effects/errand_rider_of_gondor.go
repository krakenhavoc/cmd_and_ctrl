package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "00749bad-dded-4fec-bf34-ff14cf561c26",
		Name:         "Errand-Rider of Gondor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Errand-Rider of Gondor — draw a card, then if you don't control a legendary creature, put a card from your hand on the bottom of your library",
				Do(DrawCards{N: 1}, Instruction(errandRiderBottomUnlessLegendary))),
		},
	})
}

// errandRiderBottomUnlessLegendary is "Then if you don't control a
// legendary creature, put a card from your hand on the bottom of your
// library."
func errandRiderBottomUnlessLegendary(ctx *Context) error {
	if len(permanentsControlledByMatching(ctx.Game, ctx.Controller(), And(Creature(), Legendary()))) > 0 {
		return nil
	}
	return rfPutCardFromHandOnBottom(ctx, "Errand-Rider of Gondor — put a card from your hand on the bottom of your library")
}
