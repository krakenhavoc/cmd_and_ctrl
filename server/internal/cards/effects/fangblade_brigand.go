package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	firebreathing := func(name string) ActivatedAbility {
		return ActivatedAbility{
			Label:   "{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisCreatureUntilEOT(name+" — +1/+0 and first strike", 1, 0, "first strike"),
		}
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Fangblade Brigand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated:       []ActivatedAbility{firebreathing("Fangblade Brigand")},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Fangblade Eviscerator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Activated: []ActivatedAbility{
			firebreathing("Fangblade Eviscerator"),
			{
				Label:   "{4}{R}: Creatures you control get +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{4}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()),
						Power: 2,
						Label: "Fangblade Eviscerator — creatures you control get +2/+0",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
