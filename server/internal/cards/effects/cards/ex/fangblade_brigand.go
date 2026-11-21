package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fangblade Brigand // Fangblade Eviscerator — {3}{R} Creature — Human
// Werewolf 3/4 // Creature — Werewolf 4/5 (#2586, ADR 0132):
//
//	Front: "{1}{R}: This creature gets +1/+0 and gains first strike until
//	        end of turn.
//	        Daybound"
//	Back:  "{1}{R}: This creature gets +1/+0 and gains first strike until
//	        end of turn.
//	        {4}{R}: Creatures you control get +2/+0 until end of turn.
//	        Nightbound"
//
// No simplification.
func init() {
	const oracle = "dbd22a65-4ccb-4435-ae27-03a47a86d630"
	firebreathing := func(name string) effects.ActivatedAbility {
		return effects.ActivatedAbility{
			Label:   "{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisCreatureUntilEOT(name+" — +1/+0 and first strike", 1, 0, "first strike"),
		}
	}
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Fangblade Brigand",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated:       []effects.ActivatedAbility{firebreathing("Fangblade Brigand")},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Fangblade Eviscerator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Activated: []effects.ActivatedAbility{
			firebreathing("Fangblade Eviscerator"),
			{
				Label:   "{4}{R}: Creatures you control get +2/+0 until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2}, Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{4}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Power: 2,
						Label: "Fangblade Eviscerator — creatures you control get +2/+0",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
