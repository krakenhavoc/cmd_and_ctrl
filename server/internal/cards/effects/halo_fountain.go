package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "36a9de4c-8073-475d-8679-e50844eef09e",
		Name:         "Halo Fountain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, {T}, Untap a tapped creature you control: Create a 1/1 green and white Citizen creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}",
					TapOthers: UntapTapped(1, "a tapped creature you control", Creature())},
				Effect: Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}),
			},
			{
				Label:   "{W}{W}, {T}, Untap two tapped creatures you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Draws: 1},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}{W}",
					TapOthers: UntapTapped(2, "two tapped creatures you control", Creature())},
				Effect: Do(DrawCards{N: 1}),
			},
			{
				Label:   "{W}{W}{W}{W}{W}, {T}, Untap fifteen tapped creatures you control: You win the game.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost: game.AbilityCost{Tap: true, Mana: "{W}{W}{W}{W}{W}",
					TapOthers: UntapTapped(15, "fifteen tapped creatures you control", Creature())},
				Effect: Do(WinTheGame{}),
			},
		},
	})
}
