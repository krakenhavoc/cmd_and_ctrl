package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiendish Duo — Creature — Devil {4}{R}{R}, 5/5 (EDHREC rank 3559):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)
//	 If a source would deal damage to an opponent, it deals double
//	 that damage to that player instead."
//
// The one-sided damage doubler. First strike rides PrintedKeywords.
// The replacement doubles damage from ANY source — combat and
// noncombat, the controller's own creatures and spells, an
// opponent's Bolt at another opponent — as long as the recipient is
// a PLAYER who is an opponent of the Duo's controller. Damage to the
// controller, and to any permanent, is untouched. CR 616: with a
// second doubler the affected player orders them, and doubling
// twice is x4 either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab0dfae5-b9d4-417b-8a0d-2525ae3a73b9",
		Name:            "Fiendish Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{
			b34DoubleDamageToOpponents("Fiendish Duo: double damage to opponents"),
		},
	})
}
