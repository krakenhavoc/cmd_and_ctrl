package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lys Alana Dignitary — Creature — Elf Advisor {1}{G}, 2/3:
//
//	"As an additional cost to cast this spell, behold an Elf or pay {2}. (To behold an Elf, choose an Elf you control or reveal an Elf card from your hand.)
//	 {T}: Add {G}{G}. Activate only if there is an Elf card in your graveyard."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "ce7582a2-e014-41fb-8753-00d29cffc81a",
		Name:           "Lys Alana Dignitary",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("an", "Elf", "{2}"),
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{G}{G}",
			Label:     "Add {G}{G}",
			Condition: GraveyardAtLeast(1, func(c game.Card) bool { return c.HasSubtype("Elf") }),
		}},
	})
}
