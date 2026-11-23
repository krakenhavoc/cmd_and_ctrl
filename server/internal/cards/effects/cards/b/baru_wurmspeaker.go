package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	wurms := effects.TribeFilter{Tribes: []string{"Wurm"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "1cf27d78-d6ae-4fa2-9c8d-c277913f821c",
		Name:         "Baru, Wurmspeaker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(wurms, 2, 2),
			effects.TribalKeywordGrant(wurms, "trample"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{G}, {T}: Create a 4/4 green Wurm creature token. This ability costs {X} less to activate, where X is the greatest power among Wurms you control.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{7}{G}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.GreatestPowerAmongYourWurms,
					"This ability costs {X} less to activate, where X is the greatest power among Wurms you control."),
			},
			Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Wurm"), N: 1}),
		}},
	})
}
