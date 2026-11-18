package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercity Scrounger — Artifact Creature — Human Rogue {2}{B}, 1/4
// (Kamigawa: Neon Dynasty, S82):
//
//	"{T}: Create a Treasure token. Activate only if a creature died this
//	 turn."
//
// Barad-dûr's gate (ACreatureDiedThisTurn) on an ordinary activated ability,
// since it makes a token and is not a mana ability. The {T} is the
// creature's own, so a summoning-sick Scrounger can't use it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2750451f-348e-452f-88d0-bc921c46e8cc",
		Name:         "Undercity Scrounger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Create a Treasure token. Activate only if a creature died this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      TapCost(),
			Condition: ACreatureDiedThisTurn(),
			Effect:    Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
