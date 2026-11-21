package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "30ee2790-1f5d-4073-aae7-5e7ae30421d2",
		Name:         "Tolsimir Wolfblood",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create Voja, a legendary 2/2 green and white Wolf creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green and white Legendary Wolf named Voja"), N: 1}),
		}},
	})
}
