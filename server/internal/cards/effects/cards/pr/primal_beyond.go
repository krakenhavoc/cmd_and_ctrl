package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Beyond — Land (Lorwyn Eclipsed Commander, #5368):
//
//	"As this land enters, you may reveal an Elemental card from your hand. If
//	 you don't, this land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast an Elemental
//	 spell or activate an ability of an Elemental."
//
// The reveal-land entry (reveal_lands.go) over a card with the Elemental
// subtype, and Smokebraider's spend restriction on the coloured mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "541744d9-449d-420a-a5a1-2fffba18450f",
		Name:         "Primal Beyond",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnlessYouRevealFromHand("Primal Beyond", "an Elemental card", effects.Cmdr2S12IsElementalCard),
		},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: effects.Cmdr2S12ElementalManaRestrictions(),
				Label:        "Add one mana of any color (Elemental spells and Elementals' abilities only)",
			},
		},
	})
}
