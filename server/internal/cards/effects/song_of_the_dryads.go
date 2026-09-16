package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Song of the Dryads — Enchantment — Aura for {2}{G}:
//
//	"Enchant permanent
//	 Enchanted permanent is a colorless Forest land."
//
// The widest of the three, and the one that answers things no
// removal spell in green can touch: a planeswalker, an Equipment, an
// opposing Aura, a land. "Enchant permanent" is the whole card.
//
// The ability loss is not printed on it. CR 305.7: an effect that
// sets a land's subtype to a basic land type makes it lose all
// abilities generated from its rules text and GAIN the mana ability
// of that land type, and "this doesn't remove any abilities that were
// granted to the land by other effects". Here is how each half is
// built, and where that differs from the rules:
//
//   - layer 4 — "is a ... Forest land", replacing card types and
//     subtypes. The single most consequential clause in the engine,
//     because the host stops being a creature: every Equipment on it
//     unattaches under CR 704.5n and every "enchant creature" Aura
//     on it goes to the graveyard under CR 704.5m, both as
//     state-based actions, in the same settling. An enchanted
//     Aura or Equipment should itself become unattached under
//     CR 704.5p; that rule isn't implemented (#675).
//   - layer 5 — colourless. Not cosmetic in Commander: it takes a
//     commander out of range of a colour-restricted answer, and it
//     changes what the permanent contributes to devotion.
//   - layer 6 — the ability loss, built as a full LoseAllAbilities.
//     That is what stops a Song'd Sol Ring making mana and a Song'd
//     Saga advancing. The rules put this loss in layer 4, as part of
//     the type change, and limit it to the permanent's own rules
//     text, so abilities other effects granted earlier should
//     survive; this wipes them (#669).
//
// The {G} it taps for is deliberately NOT declared here. It comes
// from the Forest subtype, through game.ManaAbilitiesForCard's
// intrinsic half, which reads the EFFECTIVE subtypes — the same
// derivation Urborg, Tomb of Yawgmoth uses. That is the right place
// for it and not an accident of convenience: CR 305.6's intrinsic
// ability is granted by the land type, so it survives the same
// effect's ability removal, and declaring it on the spec would have
// put it in the half that gets removed.
//
// A note on what does NOT happen, because it looks like a bug and is
// not: the enchanted permanent keeps its counters. A planeswalker
// turned into a Forest still has its loyalty counters sitting on it,
// and gets them back if the Song leaves. CR 613.1f removes
// abilities; CR 122 counters are objects on the permanent and are
// not abilities. The loyalty state-based action does not kill it
// either, because it is not a planeswalker while the Song is on it.
//
// DECLARED GAPS. Neither is weaker than printed: both leave the
// enchanted permanent with less than the rules give it.
//
//   - CR 305.7, STRONGER than printed. The Song removes only the
//     abilities from the permanent's own rules text; Song's
//     2014-11-07 ruling says the permanent "will still have any
//     abilities it gained from other effects." LoseAllAbilities is a
//     full layer-6 wipe, so an ability another effect granted BEFORE
//     the Song attached (Boros Charm's indestructible) is removed
//     too, and a wrath then destroys a Forest the printed card would
//     leave alive. A grant newer than the Song survives, by timestamp.
//   - CR 613.8, and the enchanted permanent again ends up with less:
//     Urborg, Tomb of Yawgmoth's "each land" depends on the land type
//     this writes. The layer engine orders layer 4 by timestamp only,
//     so with an Urborg that entered BEFORE the Song attached, the
//     enchanted permanent is a Forest and not also a Swamp, and it
//     loses the {B} mana ability Urborg should give it.
//
// Both are pinned, skipped, in layer_dependency_pairs_test.go and go
// with the CR 613.8 dependency work, which also moves the CR 305.7
// removal into layer 4.
func init() {
	Register(Spec{
		OracleID:     "7c3944fa-7c86-4979-85a9-86196aa94594",
		Name:         "Song of the Dryads",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The enchanted permanent also loses abilities other effects gave it before Song of the Dryads was attached (for example indestructible from Boros Charm); the printed card only removes its own abilities.",
			"If Urborg, Tomb of Yawgmoth was on the battlefield before Song of the Dryads was attached, the enchanted permanent isn't also a Swamp and doesn't tap for {B}.",
		},
		Targets: EnchantPermanent(),
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Land"}, []string{"Forest"}),
			SetAttachedColors(),
			LoseAllAbilities(),
		},
	})
}
