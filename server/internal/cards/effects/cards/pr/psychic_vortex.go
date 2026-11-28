package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Vortex — {2}{U}{U} Enchantment (Weatherlight, S82 OLD batch 772, #5629):
//
//	"Cumulative upkeep—Draw a card.
//	 At the beginning of your end step, sacrifice a land and discard your hand."
//
// The cumulative upkeep's payment is drawing (#6188, DrawPayment): a card for each
// age counter, or the Vortex is sacrificed. At the beginning of its controller's
// end step they sacrifice a land of their choice, then discard their whole hand,
// the discard happening whether or not a land could be sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "451e35dd-e0b5-4157-8511-bacb315e19bd",
		Name:         "Psychic Vortex",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Psychic Vortex — cumulative upkeep: draw a card", effects.DrawPayment(1), "", ""),
			effects.AtYourEndStep("Psychic Vortex — sacrifice a land and discard your hand", effects.S82OldB772SacrificeALandAndDiscardYourHand),
		},
	})
}
