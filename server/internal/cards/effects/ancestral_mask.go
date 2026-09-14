package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestral Mask — Enchantment — Aura for {2}{G} (EDHREC rank 2247):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 for each other enchantment on the
//	 battlefield."
//
// The enchantress deck's finisher, and the counted pump that is NOT
// scoped to you: "on the battlefield" means every player's
// enchantments, so a table full of them makes the Mask enormous and a
// Wrath-style enchantment sweep makes it a vanilla Aura again.
//
// TWO THINGS THE WORDING DECIDES, both of them in
// otherEnchantmentsOnTheBattlefield:
//
//   - "OTHER" excludes the Mask itself, so the floor is +0/+0 rather
//     than +2/+2. That is the difference between this card and All
//     That Glitters, which counts itself.
//   - "EACH OTHER ENCHANTMENT" includes opponents' — there is no
//     controller filter at all, which is why this is the one counter
//     in the batch that does not go through permanentsYouControl.
//
// The count reads post-layer types, so an enchantment creature counts
// once and an animated enchantment still counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db5380ed-ba28-4ea2-abc3-4998e2022903",
		Name:         "Ancestral Mask",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttachedPer(2, 2, otherEnchantmentsOnTheBattlefield),
		},
	})
}
