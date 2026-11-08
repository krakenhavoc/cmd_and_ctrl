package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valley Flamecaller — Creature — Lizard Warlock {2}{R}, 3/3
// (Bloomburrow, #4467):
//
//	"If a Lizard, Mouse, Otter, or Raccoon you control would deal damage
//	 to a permanent or player, it deals that much damage plus 1
//	 instead."
//
// A damage replacement in Mechanized Warfare's shape (CR 614), with
// Trance Kuja's source test: the source's characteristics as the damage
// event snapshotted them, its controller you, and one of the four
// subtypes (or every creature type) among its subtypes. The subtypes are
// words of the Flamecaller's text (CR 612.2). "A Lizard you control"
// names a permanent, so a Kindred instant or sorcery is never boosted.
// The Flamecaller is a
// Lizard itself, so its own damage gets the extra 1. Two Flamecallers add
// 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60654004-7d9a-47ac-a70d-b246dd1728d3",
		Name:         "Valley Flamecaller",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: valleyFlamecallerApplies,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Valley Flamecaller: +1 damage",
		}},
	})
}

// valleyFlamecallerApplies is "a Lizard, Mouse, Otter, or Raccoon you
// control would deal damage".
func valleyFlamecallerApplies(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil || src == nil {
		return false
	}
	ch, ok := damageSourceCharacteristics(ev, g)
	// A Kindred instant or sorcery is a spell, never "a Lizard you
	// control"; every other source with one of the four subtypes is a
	// permanent (or was, read as it last existed).
	if !ok || ch.Controller != src.Controller || hasFold(ch.Types, "Instant") || hasFold(ch.Types, "Sorcery") {
		return false
	}
	if ch.AllCreatureTypes {
		return true
	}
	for _, sub := range []string{src.TextSubtype("Lizard"), src.TextSubtype("Mouse"), src.TextSubtype("Otter"), src.TextSubtype("Raccoon")} {
		if hasFold(ch.Subtypes, sub) {
			return true
		}
	}
	return false
}
