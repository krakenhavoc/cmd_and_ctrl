package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necromancy — Enchantment {2}{B}:
//
//	"You may cast this spell as though it had flash. If you cast it any
//	 time a sorcery couldn't have been cast, the controller of the
//	 permanent it becomes sacrifices it at the beginning of the next
//	 cleanup step.
//	 When this enchantment enters, if it's on the battlefield, it becomes
//	 an Aura with "enchant creature put onto the battlefield with
//	 Necromancy." Put target creature card from a graveyard onto the
//	 battlefield under your control and attach this enchantment to it.
//	 When this enchantment leaves the battlefield, that creature's
//	 controller sacrifices it."
//
// #3156, on ADR 0154's gained clause. The cast is not targeted: it is an
// enchantment spell, not an Aura spell. The flash rider is
// Spec.FlashWithCleanupSacrifice (game/flash_cleanup_sacrifice.go): cast
// at instant speed with no other permission that would have allowed it,
// the enchantment is sacrificed at the beginning of the next cleanup
// step, and the reanimated creature goes with it through the "leaves the
// battlefield" delayed trigger. The enters trigger targets a creature
// card in any graveyard and becomes an Aura as it resolves
// (BecomeAnAuraAndReanimateTarget, graveyard_auras.go). Put onto the
// battlefield without being cast (Sun Titan) it is still an enchantment
// that enters and triggers.
func init() {
	t := Targeting(
		WhenThisEnters("Necromancy — put target creature card from a graveyard onto the battlefield under your control",
			Do(BecomeAnAuraAndReanimateTarget{})),
		targetCreatureInAnyGraveyard(),
	)
	t.Purpose = ForTargets(Reanimates(0))
	Register(Spec{
		OracleID:                  "699b192c-f3f2-4ac5-9ca9-e35bcf08ad8a",
		Name:                      "Necromancy",
		Completeness:              CompletenessFull,
		FlashWithCleanupSacrifice: true,
		Triggered:                 []game.TriggeredAbility{t},
	})
}
