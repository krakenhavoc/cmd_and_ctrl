package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blighted Agent — Creature — Phyrexian Human Rogue {1}{U}, 1/1:
//
//	"Infect
//	 This creature can't be blocked."
//
// Infect is the engine's damage tail (ADR 0056, #748). "Can't be
// blocked" is the S24 restriction bit on the creature itself, read at
// the block declaration (CR 509.1b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "e48ea9ea-64bc-4f1c-a424-592d48569244",
		Name:            "Blighted Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
