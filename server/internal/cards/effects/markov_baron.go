package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Markov Baron — Creature — Vampire Noble {2}{B}, 2/2 (March of the
// Machine: The Aftermath):
//
//	"Convoke
//	 Lifelink
//	 Other Vampires you control get +1/+1.
//	 Madness {2}{B}"
//
// Convoke is the spell's tap cost (Autochthon Wurm's Spec.TapCost, CR
// 702.51), lifelink the printed keyword, the lord is Legion Lieutenant's
// TribalAnthem over effective subtypes, and madness is the engine's
// (Arrogant Wurm's Madness string, CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bafc2e11-5f1c-46f1-a852-2330e099b8e2",
		Name:            "Markov Baron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		TapCost:         Convoke(),
		Madness:         "{2}{B}",
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
