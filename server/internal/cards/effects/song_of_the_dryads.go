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
//     Aura or Equipment becomes unattached itself under CR 704.5p,
//     which is why a Song'd Control Magic hands the creature back.
//   - layer 5 — colourless. Not cosmetic in Commander: it takes a
//     commander out of range of a colour-restricted answer, and it
//     changes what the permanent contributes to devotion.
//   - layer 4 again — the ability loss, which is part of the SAME
//     clause and not a second sentence (SetsBasicLandType). That is
//     what stops a Song'd Sol Ring making mana and a Song'd Saga
//     advancing, and putting it in layer 4 rather than layer 6 is
//     what makes CR 305.7's last sentence true: a grant that lands
//     in layer 6 is applied after this and survives it, whenever it
//     was made. Boros Charm's indestructible on a Sol Ring the Song
//     later enchants is the case (#669).
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
// Urborg, Tomb of Yawgmoth's "each land" depends on the land type
// this writes (CR 613.8a), so the enchanted permanent is a Forest
// Swamp that taps for {B} as well as {G} whichever of the two arrived
// first. The layer engine resolves that dependency since ADR 0067;
// both entry orders are pinned in layer_dependency_pairs_test.go.
func init() {
	Register(Spec{
		OracleID:     "7c3944fa-7c86-4979-85a9-86196aa94594",
		Name:         "Song of the Dryads",
		Completeness: CompletenessFull,
		Targets:      EnchantPermanent(),
		Static: []game.StaticAbility{
			SetsBasicLandType(AttachedToSource, []string{"Land"}, []string{"Forest"}),
			SetAttachedColors(),
		},
	})
}
