package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citanul Stalwart — Creature — Elf Druid Soldier {G}, 1/1 (The Brothers'
// War):
//
//	"{T}, Tap an untapped artifact or creature you control: Add one mana of
//	 any color."
//
// Gene Pollinator's mana ability with an artifact-or-creature filter. The
// {T} taps the Stalwart, so the other permanent is another one
// (TapAnotherUntapped); it is not the {T} symbol (CR 302.6), so one that
// arrived this turn pays it, and it does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce291948-125d-493d-8770-8b3aac515f9b",
		Name:         "Citanul Stalwart",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, TapOthers: effects.TapAnotherUntapped("an untapped artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature())).TapOthers},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
