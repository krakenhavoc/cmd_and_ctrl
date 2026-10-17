package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	f := TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, YoursOnly: true}
	Register(Spec{
		OracleID:        "614c3f71-e231-4c12-9e6b-cb8ea1d100bf",
		Name:            "Howlpack Resurgence",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			TribalAnthem(f, 1, 1),
			TribalKeywordGrant(f, "trample"),
		},
	})
}
