package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Outcasts — Creature — Vampire {2}{B}{B}, 2/2 (Magic 2012, #5573):
//
//	"Bloodthirst 2 (If an opponent was dealt damage this turn, this
//	 creature enters with two +1/+1 counters on it.)
//	 Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); lifelink
// is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25ad4e88-b5ee-4853-a4fc-0a0420c746f6",
		Name:            "Vampire Outcasts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(2, "Vampire Outcasts")},
	})
}
