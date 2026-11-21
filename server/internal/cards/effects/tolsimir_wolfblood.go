package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolsimir Wolfblood — Legendary Creature — Elf Warrior {4}{G}{W}, 3/4
// (Ravnica Remastered, #5948):
//
//	"Other green creatures you control get +1/+1.
//	 Other white creatures you control get +1/+1.
//	 {T}: Create Voja, a legendary 2/2 green and white Wolf creature token."
//
// Wilt-Leaf Liege's two anthems (a green and white creature gets +2/+2, each
// ability applying on its own). The token is the printed Voja: legendary, so
// a second one puts the legend rule to work (CR 704.5j); see
// s82_gap_b73_helpers.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30ee2790-1f5d-4073-aae7-5e7ae30421d2",
		Name:         "Tolsimir Wolfblood",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("G"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create Voja, a legendary 2/2 green and white Wolf creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    TapCost(),
			Effect:  Do(CreateToken{Template: TokenCard("2/2 green and white Legendary Wolf named Voja"), N: 1}),
		}},
	})
}
