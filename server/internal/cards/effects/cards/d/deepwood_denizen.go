package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepwood Denizen — Creature — Elf Warrior {2}{G}, 3/2:
//
//	"Vigilance
//	 {5}{G}, {T}: Draw a card. This ability costs {1} less to activate
//	 for each +1/+1 counter on creatures you control."
//
// The reduction counts counters, not creatures (s82BnsB28CountersOnYourCreatures).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a6887e3-7a65-4f24-be71-183b00a0ad56",
		Name:            "Deepwood Denizen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}, {T}: Draw a card. This ability costs {1} less to activate for each +1/+1 counter on creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}{G}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.S82BnsB28CountersOnYourCreatures,
					"This ability costs {1} less to activate for each +1/+1 counter on creatures you control."),
			},
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
