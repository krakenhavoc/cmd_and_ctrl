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
// The conditional flash is FlashIfTargets: the card's own timing
// permission (Spec.FlashIf, #5207), which narrows the target clause to a
// commander for a cast that is open only because of it (CR 601.3b: the
// target the caster will choose can make the permission apply, so the
// cast may begin at instant speed and the choice must then satisfy it).
// The plain cast stays sorcery-speed with any creature. It is not an
// alternative cost, so a flash cast may still claim one (CR 118.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e589dca-d31e-4d2f-81b1-18e5f9198795",
		Name:         "Timely Ward",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		FlashIf:      FlashIfTargets("a commander", EnchantCreature(ACommander())),
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
	})
}
