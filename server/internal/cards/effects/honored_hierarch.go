package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honored Hierarch — Creature — Human Druid {G}, 1/1 (Magic Origins
// Clash Pack, #5550):
//
//	"Renown 1 (When this creature deals combat damage to a player, if it
//	 isn't renowned, put a +1/+1 counter on it and it becomes renowned.)
//	 As long as this creature is renowned, it has vigilance and '{T}: Add
//	 one mana of any color.'"
//
// #2049: renown is the engine's keyword trigger (game/renown.go). The
// vigilance is Goblin Glory Chaser's layer-6 self-grant behind the
// Renowned gate; the mana ability is an ADR 0093 bundle the Hierarch
// grants itself behind the same gate (Alchemist's Talent's gated grant),
// so until it is renowned it has no mana ability at all, and {T} is
// subject to summoning sickness as the Hierarch's own (CR 302.6).
//
// No simplification.
func init() {
	const grant = "honored-hierarch/any-color"
	mana := GrantAbilities(selfOnly, grant)
	mana.ActiveWhen = Renowned()
	Register(Spec{
		OracleID:        "ae42fa8a-b0f6-43dc-a4c6-e2531890ca3a",
		Name:            "Honored Hierarch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Grants:          []AbilityGrant{AnyColorManaGrant(grant)},
		Static:          []game.StaticAbility{RenownedKeywords("vigilance"), mana},
	})
}
