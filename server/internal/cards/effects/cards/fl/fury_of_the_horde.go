package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c90e2db7-6973-4243-9645-265fd3bc8546",
		Name:         "Fury of the Horde",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch("Exile two red cards from your hand", 0,
				effects.CardInYourHand("two red cards from your hand", effects.OfColor("R")).WithCount(2, 2),
				"two red cards"),
		},
		OnResolve: effects.UntapAttackersThenCombatAndMain,
	})
}
