package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Machine God's Effigy — Artifact {4}:
//
//	"You may have this artifact enter as a copy of any creature on
//	 the battlefield, except it's an artifact and it has '{T}: Add
//	 {U}.' (It's not a creature.)
//	 {T}: Add {U}."
//
// The copy clause is EntersAsCopyOf over any creature on the battlefield
// with two except edits. "Except it's an artifact" REPLACES the copied
// types rather than adding to them (PrintedValues.SetCardTypes — the
// parenthetical says so), so the copy is an Artifact with the copied
// creature's name, abilities and printed power and toughness, and no
// creature subtypes: it can't attack, block or be hit by anything that
// cares about creatures, and its copied mana ability, triggers and
// statics all work through the copied oracle ID. "It has '{T}: Add
// {U}'" is the registered grant below (CR 707.9a), because a copy takes
// the copied card's mana abilities and not the Effigy's own printed
// one. Declining the copy leaves a plain Effigy with its printed mana
// ability.
const machineGodsEffigyGrant = "machine-gods-effigy/tap-for-blue"

var machineGodsEffigyMana = ManaAbility{
	Cost:     ManaAbilityCost{Tap: true},
	Produced: "{U}",
	Label:    "Add {U}",
}

func init() {
	Register(Spec{
		OracleID:     "64ebdd6f-acde-4aab-a86b-2798bad5f70c",
		Name:         "Machine God's Effigy",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			machineGodsEffigyMana,
		},
		Grants: []AbilityGrant{{
			Key:  machineGodsEffigyGrant,
			Text: "{T}: Add {U}.",
			Mana: []ManaAbility{machineGodsEffigyMana},
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Machine God's Effigy",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsCreature() })
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetCardTypes("Artifact")
					v.GrantAbility(machineGodsEffigyGrant)
				},
			),
		},
	})
}
