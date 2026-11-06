package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tectonic Split — Enchantment {4}{G}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"As an additional cost to cast this spell, sacrifice half the lands
//	 you control, rounded up.
//	 Hexproof
//	 Lands you control have "{T}: Add three mana of any one color.""
//
// The cost is the sacrifice-half seam (avB30cSacrificeHalfRoundedUpCost,
// game.AdditionalCost.SacrificeHalfRoundedUp): the caster picks exactly
// half the lands they control, rounded up, counted as the cost is
// determined (CR 601.2f), and they leave with the spell on the stack.
// The lands may be tapped for mana first, as any permanent sacrificed to
// pay a cost may (CR 601.2g). The grant is Chromatic Lantern's shape, a
// layer 6 ability grant (CR 613.1f) to every land you control, with the
// "N mana of any one color" pipe Command Chair's mana ability uses.
//
// No simplification.
func init() {
	const grant = "tectonic-split/three-of-one-color"
	Register(Spec{
		OracleID:        "b71c3c1e-f927-4192-ac3a-cf6a442ffd00",
		Name:            "Tectonic Split",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		AdditionalCost:  avB30cSacrificeHalfRoundedUpCost("lands you control", Land(), YouControl()),
		Grants: []AbilityGrant{TapForManaGrant(grant, OneColorOfAmount(3),
			"Add three mana of any one color", "{T}: Add three mana of any one color.")},
		Static: []game.StaticAbility{Labelled(GrantAbilities(landsYouControl, grant),
			"Lands you control have \"{T}: Add three mana of any one color.\"")},
	})
}
