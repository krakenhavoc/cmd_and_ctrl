package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selesnya Guildmage — Creature — Elf Wizard {G/W}{G/W}, 2/2 (Ravnica: Clue
// Edition):
//
//	"{3}{G}: Create a 1/1 green Saproling creature token.
//	 {3}{W}: Creatures you control get +1/+1 until end of turn."
//
// Vitu-Ghazi Guildmage's token row and a mass pump until end of turn
// (CR 611.2c: the set of creatures is fixed on resolution).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fbdca3a-7aa3-4b5a-98d7-9168856045ac",
		Name:         "Selesnya Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{G}: Create a 1/1 green Saproling creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    ManaCost("{3}{G}"),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}),
			},
			{
				Label:   "{3}{W}: Creatures you control get +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Power: 1, Toughness: 1}},
				Cost:    ManaCost("{3}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
						Label: "Selesnya Guildmage — creatures you control get +1/+1",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
