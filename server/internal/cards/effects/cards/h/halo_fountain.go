package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halo Fountain — Artifact {2}{W} (Streets of New Capenna, #5669):
//
//	"{W}, {T}, Untap a tapped creature you control: Create a 1/1 green
//	 and white Citizen creature token.
//	 {W}{W}, {T}, Untap two tapped creatures you control: Draw a card.
//	 {W}{W}{W}{W}{W}, {T}, Untap fifteen tapped creatures you control:
//	 You win the game."
//
// Each cost untaps tapped creatures (UntapTapped, #5669): only a tapped
// creature pays (CR 701.26b), it is not the {Q} symbol so a creature
// that arrived this turn may be untapped (CR 302.6), and untapping an
// attacking creature leaves it attacking. Every row declares
// AnswerMakesBlocker: the creatures its cost untaps can block. The win is
// CR 104.2b, during the resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36a9de4c-8073-475d-8679-e50844eef09e",
		Name:         "Halo Fountain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}, Untap a tapped creature you control: Create a 1/1 green and white Citizen creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}",
					TapOthers: effects.UntapTapped(1, "a tapped creature you control", effects.Creature())},
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1}),
			},
			{
				Label:   "{W}{W}, {T}, Untap two tapped creatures you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Draws: 1},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}{W}",
					TapOthers: effects.UntapTapped(2, "two tapped creatures you control", effects.Creature())},
				Effect: effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{W}{W}{W}{W}{W}, {T}, Untap fifteen tapped creatures you control: You win the game.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}{W}{W}{W}{W}",
					TapOthers: effects.UntapTapped(15, "fifteen tapped creatures you control", effects.Creature())},
				Effect: effects.Do(effects.WinTheGame{}),
			},
		},
	})
}
