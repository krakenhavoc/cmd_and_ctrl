package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawnsire of Ulamog — Creature — Eldrazi {10}, 7/11:
//
//	"Annihilator 1
//	 {4}: Create two 0/1 colorless Eldrazi Spawn creature tokens. They
//	 have "Sacrifice this token: Add {C}."
//	 {20}: Cast any number of Eldrazi spells from among cards you own
//	 outside the game without paying their mana costs."
//
// Annihilator 1 is the canonical keyword token (ADR 0113 §2), and the
// Spawn are the shared Eldrazi Spawn token.
//
// Sandbox simplification, declared — one whole ability omitted, the
// Kozilek, the Great Distortion posture: the {20} ability is not
// implemented. The engine has no zone for cards a player owns outside
// the game, so there is nothing for it to cast. Leaving it off is
// weaker than printed, and the caveat says so.
func init() {
	Register(Spec{
		OracleID:        "b90d2f4d-b4ea-40af-aea0-1ab4234ab80f",
		Name:            "Spawnsire of Ulamog",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The {20} ability isn't implemented — you can't cast Eldrazi spells from outside the game."},
		PrintedKeywords: []string{"annihilator 1"},
		Activated: []ActivatedAbility{{
			Label:   "{4}: Create two 0/1 colorless Eldrazi Spawn creature tokens. They have \"Sacrifice this token: Add {C}.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{4}"),
			Effect:  Do(CreateToken{Template: EldraziSpawnToken(), N: 2}),
		}},
	})
}
