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
// changes a permanent's subtype to a basic land type makes it lose
// all abilities generated from its rules text and GAIN the mana
// ability of that land type. Both halves are here:
//
//   - layer 4 — "is a ... Forest land", replacing card types and
//     subtypes. The single most consequential clause in the engine,
//     because the host stops being a creature: every Equipment on it
//     unattaches under CR 704.5n and every "enchant creature" Aura
//     on it goes to the graveyard under CR 704.5m, both as
//     state-based actions, in the same settling.
//   - layer 5 — colourless. Not cosmetic in Commander: it takes a
//     commander out of range of a colour-restricted answer, and it
//     changes what the permanent contributes to devotion.
//   - layer 6 — CR 305.7's ability loss, which is what stops a
//     Song'd Sol Ring making mana and a Song'd Saga advancing.
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
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c3944fa-7c86-4979-85a9-86196aa94594",
		Name:         "Song of the Dryads",
		Completeness: CompletenessFull,
		Targets:      EnchantPermanent(),
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Land"}, []string{"Forest"}),
			SetAttachedColors(),
			LoseAllAbilities(),
		},
	})
}
