package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warg Rider — Creature — Orc Warrior {4}{B}, 4/3:
//
//	"Menace
//	 Other Orcs and Goblins you control have menace.
//	 At the beginning of combat on your turn, amass Orcs 2."
//
// Menace rides PrintedKeywords. The lord half is a layer-6 keyword
// grant over a tribe filter (the Goblin Chieftain shape), so a creature
// that becomes an Orc, the amassed Army included, has menace in the
// same recompute. The beginning-of-combat amass is the Dreadhorde
// Invasion find-or-create shape on a different step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f4dd9e2-6460-413e-a3b8-96b20bf57f8b",
		Name:            "Warg Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Orc", "Goblin"}, Others: true, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Warg Rider — amass Orcs 2", Do(Amass{Subtype: "Orc", N: 2})),
		},
	})
}
