package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Mauler — Creature — Kavu {4}{G}{G}, 4/4 (Apocalypse, S82 OLD, #5614):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+1 until end of turn for each
//	 other attacking Kavu."
//
// Rimehorn Aurochs's attack trigger over Kavu at +1/+1: the other attacking
// Kavu are counted as the trigger resolves, and the Mauler itself never counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ffbcaba-5437-4fb6-a2d6-e94b1e6dc1d2",
		Name:            "Kavu Mauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Kavu Mauler — +1/+1 until end of turn for each other attacking Kavu",
				effects.OldB630PumpPerOtherAttackingKavu("Kavu Mauler — +1/+1 for each other attacking Kavu")),
		},
	})
}
