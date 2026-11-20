package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joiner Adept — Creature — Elf Druid {1}{G}, 2/1 (Fifth Dawn, #5604):
//
//	"Lands you control have "{T}: Add one mana of any color.""
//
// Chromatic Lantern's layer-6 grant to every land you control (ADR 0093),
// on a creature.
//
// No simplification.
const joinerAdeptGrant = "joiner-adept/any-color"

func init() {
	Register(Spec{
		OracleID:     "139b098c-d4be-4dca-9bde-9eb8ea97c7bf",
		Name:         "Joiner Adept",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(joinerAdeptGrant)},
		Static:       []game.StaticAbility{Labelled(GrantAbilities(landsYouControl, joinerAdeptGrant), "Lands you control have \"{T}: Add one mana of any color.\"")},
	})
}
