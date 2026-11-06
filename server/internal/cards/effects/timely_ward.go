package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timely Ward — Enchantment — Aura for {2}{W} (EDHREC rank 1682):
//
//	"You may cast this spell as though it had flash if it targets a
//	 commander.
//	 Enchant creature
//	 Enchanted creature has indestructible."
//
// A permanent Swiftfoot-style shield for a commander: unlike a
// protection spell it does not wear off at end of turn, so the
// creature stays through every board wipe until somebody answers the
// Aura.
//
// Indestructible is honoured — S25 (#77) put it in the keyword table
// with a consumer in the destruction path — so the grant is real and
// not a badge. What it does NOT stop is the same list Darksteel Plate's
// file carries: sacrifice, exile, -X/-X to zero toughness, and the
// legend rule.
//
// The conditional flash is FlashIfTargets: an offer at the printed cost
// that opens the instant-speed window for its own claim and narrows the
// target clause to a commander (CR 601.3b: the target the caster will
// choose can make the permission apply, so the cast may begin at instant
// speed and the choice must then satisfy it). The plain cast stays
// sorcery-speed with any creature.
//
// Simplification (#5207): the offer is an alternative cost, so the flash
// cast can't also take another alternative cost (CR 118.9a). On the
// printed card the flash is a timing permission, so it could. Declared
// as the caveat.
func init() {
	Register(Spec{
		OracleID:     "8e589dca-d31e-4d2f-81b1-18e5f9198795",
		Name:         "Timely Ward",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Casting this at instant speed on a commander can't be combined with another alternative cost.",
		},
		Targets: EnchantCreature(),
		Purpose: game.Purpose{Aura: game.AuraBenefit},
		AlternativeCosts: []game.AlternativeCost{
			FlashIfTargets("{2}{W}", "a commander", EnchantCreature(ACommander())),
		},
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
	})
}
