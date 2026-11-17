package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baru, Wurmspeaker — Legendary Creature — Human Druid {2}{G}{G}, 3/3
// (Dominaria United Commander, S82):
//
//	"Wurms you control get +2/+2 and have trample.
//	 {7}{G}, {T}: Create a 4/4 green Wurm creature token. This ability costs
//	 {X} less to activate, where X is the greatest power among Wurms you
//	 control."
//
// The lord is the shared tribal anthem and keyword grant. The discount is
// the ability's own cost clause, counted as the total cost is determined
// (CR 601.2f): it takes X off the generic {7} and never reduces the {G}.
//
// No simplification.
func init() {
	wurms := TribeFilter{Tribes: []string{"Wurm"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "1cf27d78-d6ae-4fa2-9c8d-c277913f821c",
		Name:         "Baru, Wurmspeaker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(wurms, 2, 2),
			TribalKeywordGrant(wurms, "trample"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{7}{G}, {T}: Create a 4/4 green Wurm creature token. This ability costs {X} less to activate, where X is the greatest power among Wurms you control.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{7}{G}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(greatestPowerAmongYourWurms,
					"This ability costs {X} less to activate, where X is the greatest power among Wurms you control."),
			},
			Effect: Do(CreateToken{Template: TokenCard("4/4 green Wurm"), N: 1}),
		}},
	})
}
