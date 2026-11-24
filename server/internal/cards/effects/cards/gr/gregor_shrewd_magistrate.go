package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gregor, Shrewd Magistrate — Legendary Creature — Human Advisor {1}{W}{U},
// 1/3 (Secret Lair Drop):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 Whenever Gregor, Shrewd Magistrate deals combat damage to a player, draw
//	 cards equal to its power."
//
// The count is Gregor's power as the trigger resolves, or its last-known power
// if it has left the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b64da77-ca1c-427b-b48b-b919f0fe08e9",
		Name:            "Gregor, Shrewd Magistrate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Gregor, Shrewd Magistrate — draw cards equal to its power",
				effects.Do(effects.DrawForEach{Name: "its power", Count: effects.SourcePowerNow})),
		},
	})
}
