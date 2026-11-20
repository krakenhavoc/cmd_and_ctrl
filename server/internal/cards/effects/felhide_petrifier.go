package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felhide Petrifier — Creature — Minotaur Warrior {2}{B}, 2/3 (Journey into Nyx, #5558):
//
//	"Deathtouch
//	 Other Minotaur creatures you control have deathtouch."
//
// A lord's keyword half: other Minotaur creatures you control have
// deathtouch, read through the layers so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1ad604ed-c645-4f94-8c5f-8e0ce66a85ff",
		Name:            "Felhide Petrifier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Minotaur"}, Others: true, YoursOnly: true}, "deathtouch"),
		},
	})
}
