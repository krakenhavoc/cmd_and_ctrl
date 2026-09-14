package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All That Glitters — Enchantment — Aura for {1}{W} (EDHREC rank
// 625):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each artifact and/or
//	 enchantment you control."
//
// Two mana for what is routinely a +5/+5 or better in the deck that
// wants it, and the Aura counterpart of Blackblade Reforged: the
// bonus is COUNTED on every layer recompute rather than fixed at
// resolution, so playing a Signet after the Aura grows the creature
// immediately.
//
// The count includes the Aura ITSELF, which is correct — All That
// Glitters is an enchantment you control — so the floor is +1/+1.
//
// "AND/OR" IS ONE COUNT, NOT TWO. A permanent that is both an
// artifact and an enchantment (the Shrine cycle, a Bequeathal-style
// artifact Aura) counts once, which is what the wording means and
// what the `||` in artifactOrEnchantmentsControlledBy gives.
//
// "You control" is the AURA's controller, not the creature's (CR
// 109.5). Put All That Glitters on an opponent's creature to shrink
// nothing and grow nothing of theirs; the count still reads your
// board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4d751e0-41c1-4e90-853d-512f385acd81",
		Name:         "All That Glitters",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, artifactOrEnchantmentsControlledBy),
		},
	})
}
