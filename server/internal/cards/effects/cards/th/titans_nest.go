package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Titans' Nest — Enchantment {1}{B}{G}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"At the beginning of your upkeep, surveil 1.
//	 Exile a card from your graveyard: Add {C}. Spend this mana only to cast
//	 a spell that's one or more colors without {X} in its mana cost."
//
// The upkeep surveil is Mindwhisker's. The mana ability is Molt Tender's exile
// cost with a CR 106.12 spend restriction: only to cast a spell, one that is
// mono- or multicolored (the colours the spell being cast has, a hybrid
// spell is every colour of its symbols, CR 202.2d) and has no {X} in its
// printed mana cost (ManaRestrictSpellNoXCost, the mirror of Troyan's
// has-{X} tag). An activation, a colorless spell and a payment that does not
// say what it casts are all refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04505544-8341-4bd9-afa7-b31d6b415c47",
		Name:         "Titans' Nest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Titans' Nest — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{ExileCards: effects.ExileCardsFromGraveyard(1, "a card", nil)},
			Produced: "{C}",
			Restrictions: []string{
				effects.ManaRestrictCast,
				game.ManaRestrictAnyOf([]string{game.ManaRestrictMonocolored}, []string{game.ManaRestrictMulticolored}),
				game.ManaRestrictSpellNoXCost,
			},
			Label: "Exile a card from your graveyard: Add {C}. Spend this mana only to cast a spell that's one or more colors without {X} in its mana cost.",
		}},
	})
}
