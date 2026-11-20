package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Ogre — Creature — Ogre Warrior {2}{R}, 2/2 (Magic 2012, #5573):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); first
// strike is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8bbe3f0d-41d0-4c58-ad65-6675a234cad3",
		Name:            "Blood Ogre",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(1, "Blood Ogre")},
	})
}
