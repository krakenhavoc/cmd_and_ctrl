package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blanchwood Armor — Enchantment — Aura for {2}{G} (EDHREC rank
// 3079):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each Forest you control."
//
// The original counted Aura, and in mono-green Commander it is
// routinely bigger than All That Glitters for a mana more.
//
// IT COUNTS THE LAND TYPE, NOT THE NAME. "Each Forest you control"
// means any permanent with the Forest subtype, so a Stomping Ground,
// a Bayou and a Dryad Arbor all count, a basic Forest that something
// turned into an Island does not, and an Urborg-style layer-4 land
// type grant lands correctly — Card.HasSubtype reads the post-layer
// subtypes on the battlefield, which is exactly the accessor that
// question needs.
//
// Like every counted pump in this batch the total is re-read on every
// recompute, so a fetchland cracking mid-combat changes the creature's
// size before damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80ea56ad-e741-4a85-b4e8-ce62e7d593d5",
		Name:         "Blanchwood Armor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, forestsControlledBy),
		},
	})
}
