package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chained Throatseeker — Creature — Phyrexian Horror {5}{U}, 5/5:
//
//	"Infect
//	 This creature can't attack unless defending player is poisoned."
//
// The restriction is #1879's (ADR 0107 §2, CR 508.1c): a player is
// poisoned with one or more poison counters (CR 122.1f), and the
// defending player is worked out per target (CR 508.5, 508.5a), so in
// Commander it may attack any poisoned opponent, their planeswalkers and
// the battles they protect, and no one else. Infect is the engine's
// damage tail (ADR 0056).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f971a597-caa7-4289-ae8f-1efdc9abb593",
		Name:            "Chained Throatseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Static:          []game.StaticAbility{CantAttackUnlessDefendingPlayerIsPoisoned()},
	})
}
