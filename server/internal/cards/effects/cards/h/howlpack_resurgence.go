package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howlpack Resurgence — Enchantment {2}{G}:
//
//	"Flash
//	 Each creature you control that's a Wolf or a Werewolf gets +1/+1
//	 and has trample."
//
// A tribal anthem over Wolves and Werewolves you control, read through
// the layered view so a changeling counts, and the same filter grants
// trample. Flash is PrintedKeywords.
//
// No simplification.
func init() {
	f := effects.TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "614c3f71-e231-4c12-9e6b-cb8ea1d100bf",
		Name:            "Howlpack Resurgence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(f, 1, 1),
			effects.TribalKeywordGrant(f, "trample"),
		},
	})
}
