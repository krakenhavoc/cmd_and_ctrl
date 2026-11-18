package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belligerent Guest — Creature — Vampire {2}{R}, 3/2 (Innistrad: Crimson Vow):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create a Blood token."
func init() {
	Register(Spec{
		OracleID:        "417e798d-9ca6-4cd0-adb4-f61d3becb201",
		Name:            "Belligerent Guest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Belligerent Guest — create a Blood token", Do(s82Vow03CreateBlood(1))),
		},
	})
}
