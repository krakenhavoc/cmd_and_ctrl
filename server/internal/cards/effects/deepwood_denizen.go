package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepwood Denizen — Creature — Elf Warrior {2}{G}, 3/2:
//
//	"Vigilance
//	 {5}{G}, {T}: Draw a card. This ability costs {1} less to activate
//	 for each +1/+1 counter on creatures you control."
//
// The reduction counts counters, not creatures (s82BnsB28CountersOnYourCreatures).
func init() {
	Register(Spec{
		OracleID:        "4a6887e3-7a65-4f24-be71-183b00a0ad56",
		Name:            "Deepwood Denizen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}, {T}: Draw a card. This ability costs {1} less to activate for each +1/+1 counter on creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}{G}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(s82BnsB28CountersOnYourCreatures,
					"This ability costs {1} less to activate for each +1/+1 counter on creatures you control."),
			},
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
