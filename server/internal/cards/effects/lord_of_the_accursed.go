package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord of the Accursed — Creature — Zombie {2}{B}, 2/3:
//
//	"Other Zombies you control get +1/+1.
//	 {1}{B}, {T}: All Zombies gain menace until end of turn."
//
// The anthem is TribalAnthem over Zombies you control (the lord itself
// excluded). The activated ability says "all Zombies", not "Zombies you
// control", so it reaches every Zombie creature on the table, the lord and
// the opponents' included; the set is fixed as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f372107-6c52-4f1f-8bf0-7b3fa21b2cae",
		Name:         "Lord of the Accursed",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}: All Zombies gain menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(Creature(), HasSubtype("Zombie")),
				Keywords: []string{"menace"},
				Label:    "Lord of the Accursed — all Zombies gain menace",
			}),
		}},
	})
}
