package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB477Borderpost is the Archenemy Borderpost cycle's Spec (S82 OLD, #5580):
//
//	"You may pay {1} and return a basic land you control to its owner's hand
//	 rather than pay this spell's mana cost.
//	 This artifact enters tapped.
//	 {T}: Add <colour A> or <colour B>."
//
// The alternative cost is a mana-and-bounce cost (Daze's ReturnInstead with
// the {1} added); the bounce is paid at announce, with the spell already on
// the stack (CR 601.2h). The two colours are two mana abilities, as the dual
// lands print them.
func OldB477Borderpost(oracleID, name, colorA, colorB string) Spec {
	alt := ReturnInstead(
		"Pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost",
		PermanentYouControl("a basic land you control", IsBasicLandCard()),
		"a basic land you control")
	alt.ManaCost = "{1}"
	return Spec{
		OracleID:         oracleID,
		Name:             name,
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{alt},
		Replacements:     []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:    []ManaAbility{Cmdr2B01TapForColor(colorA), Cmdr2B01TapForColor(colorB)},
	}
}

// IsBasicLandCard is the CardPredicate of "a basic land" (CR 205.4a): a land
// with the basic supertype, read through the layers.
func IsBasicLandCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return IsBasicLand(c) }
}
