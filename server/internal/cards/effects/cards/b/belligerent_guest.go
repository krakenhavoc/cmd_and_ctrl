package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belligerent Guest — Creature — Vampire {2}{R}, 3/2 (Innistrad: Crimson Vow):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create a Blood token."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "417e798d-9ca6-4cd0-adb4-f61d3becb201",
		Name:            "Belligerent Guest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Belligerent Guest — create a Blood token", effects.Do(effects.S82Vow03CreateBlood(1))),
		},
	})
}
