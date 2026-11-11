package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aminatou, Veil Piercer — Legendary Creature — Human Wizard {1}{W}{U}{B},
// 2/4 (Duskmourn Commander):
//
//	"At the beginning of your upkeep, surveil 2.
//	 Each enchantment card in your hand has miracle. Its miracle cost is
//	 equal to its mana cost reduced by {4}."
//
// The static is the granted miracle of granted_miracle.go priced at each
// card's own mana cost less {4} generic (MiracleForEnchantmentCardsInYour
// HandReducedBy): coloured symbols stay, and the price floors at {0}. The
// reveal and the cast permission are the engine's miracle machinery
// (game/miracle.go), so only the first card drawn each turn is offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "bf45f5ee-4d5f-4920-b612-b5492c2f035d",
		Name:                    "Aminatou, Veil Piercer",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{MiracleForEnchantmentCardsInYourHandReducedBy(4)},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Aminatou, Veil Piercer — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
