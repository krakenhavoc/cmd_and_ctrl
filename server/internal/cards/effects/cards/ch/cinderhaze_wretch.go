package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinderhaze Wretch — Creature — Elemental Shaman {4}{B}, 3/2 (Shadowmoor, S82
// OLD b532, #5591):
//
//	"{T}: Target player discards a card. Activate only during your turn.
//	 Put a -1/-1 counter on this creature: Untap this creature."
//
// The discard is Davriel's: the player chooses the card (CR 701.9a). It is
// activated only during the Wretch's controller's turn. The untapper's whole
// cost is the counter, so it can be paid while the Wretch is tapped and uses
// the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df6d9787-b423-4067-adf4-c09dcee1329e",
		Name:         "Cinderhaze Wretch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:     "{T}: Target player discards a card. Activate only during your turn.",
				Purpose:   effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
				Cost:      effects.TapCost(),
				Condition: effects.DuringYourTurn(),
				Targets:   effects.TargetPlayer("target player"),
				Effect:    effects.TargetPlayerDiscards(1),
			},
			effects.OldB532UntapThisForAMinusOneCounter(),
		},
	})
}
