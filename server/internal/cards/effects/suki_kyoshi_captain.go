package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suki, Kyoshi Captain — Legendary Creature — Human Warrior Ally {2}{W},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Other Warriors you control get +1/+1.
//	 {3}{W}: Attacking Warriors you control gain double strike until end
//	 of turn."
//
// The anthem is Goblin Trashmaster's lord (TribalAnthem, others, yours
// only), read through the layered view so a changeling counts. The
// activated ability is War Machine's grant: the attacking Warriors you
// control are chosen once, as it resolves (CR 611.2c), so a Warrior that
// starts attacking later this turn does not gain double strike. Suki is
// a Warrior, so she gains it too if she is attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "558513bb-a76d-434a-98fc-3707a5d8886a",
		Name:         "Suki, Kyoshi Captain",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: Attacking Warriors you control gain double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}"),
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(OfCreatureType("Warrior"), YouControl(), AttackingCreature()),
				Keywords: []string{"double strike"},
				Label:    "Suki, Kyoshi Captain — double strike",
			}),
		}},
	})
}
