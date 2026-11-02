package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fury of the Horde — Sorcery {5}{R}{R} (Marvel Universe, #4460):
//
//	"You may exile two red cards from your hand rather than pay this
//	 spell's mana cost.
//	 Untap all creatures that attacked this turn. After this main phase,
//	 there is an additional combat phase followed by an additional main
//	 phase."
//
// Commandeer's two-card pitch over Relentless Assault's resolution
// (untapAttackersThenCombatAndMain, CR 500.8). The pitch is an alternative
// cost (CR 118.9) with no life: both red cards leave the hand as the cost
// is paid, and the spell itself, already on the stack, can't be one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c90e2db7-6973-4243-9645-265fd3bc8546",
		Name:         "Fury of the Horde",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile two red cards from your hand", 0,
				CardInYourHand("two red cards from your hand", OfColor("R")).WithCount(2, 2),
				"two red cards"),
		},
		OnResolve: untapAttackersThenCombatAndMain,
	})
}
