package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Threefold Signal — Artifact {3} (Streets of New Capenna Commander,
// #5491):
//
//	"When this artifact enters, scry 3.
//	 Each spell you cast that's exactly three colors has replicate {3}.
//	 (When you cast it, copy it for each time you paid its replicate cost.
//	 You may choose new targets for the copies. A copy of a permanent spell
//	 becomes a token.)"
//
// Hugh, First of Many's grant and trigger for a different set of spells:
// the replicate cost is a granted optional cost (CR 702.56a) offered on
// every spell of exactly three colours, read off the card as it is
// announced, and the trigger copies the spell once per payment. A
// colourless or mono-, two-, four- or five-colour spell is offered nothing.
// The scry is the Scry primitive on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc48da27-77e0-440f-b47b-4bdea3c3cc5e",
		Name:         "Threefold Signal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.SpellsYouCastHave(
			"Each spell you cast that's exactly three colors has replicate {3}.", effects.S82SncExactlyThreeColors, effects.ReplicateMana("{3}", 20))},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Threefold Signal — scry 3", effects.Do(effects.Scry{N: 3})),
			effects.S82SncReplicateTrigger("Threefold Signal — replicate: copy the three-color spell",
				effects.S82SncExactlyThreeColors),
		},
	})
}
