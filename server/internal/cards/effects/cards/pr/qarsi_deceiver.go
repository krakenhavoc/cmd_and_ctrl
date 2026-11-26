package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qarsi Deceiver — Creature — Snake Wizard {1}{U}, 0/4 (Dragons of Tarkir, S82 old
// batch 332):
//
//	"{T}: Add {C}. Spend this mana only to cast a face-down creature spell,
//	 pay a mana cost to turn a manifested creature face up, or pay a morph
//	 cost. (A megamorph cost is a morph cost.)"
//
// Tin Street Gossip's face-down cast clause and Creeping Peeper's turn-face-up
// payment, joined by ManaRestrictAnyOf. The turn-face-up half is narrower than
// Peeper's: it passes only the morph route (CR 702.37b, megamorph included, and a
// manifested card's morph cost, CR 701.40c) and the mana cost of a MANIFESTED
// creature (CR 701.40b). A disguise cost and a cloaked creature's mana cost are
// other routes, so this mana can't pay them (a small seam: the spend context now
// names the route and the face-down kind, ManaRestrictTurnFaceUpMorphOrManifest).
// A face-down spell is always a creature spell (CR 708.4). The ability taps the
// Deceiver, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e12681cc-c66e-4dac-882c-bd8638a2e12c",
		Name:         "Qarsi Deceiver",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictFaceDownSpell},
				[]string{game.ManaRestrictTurnFaceUpMorphOrManifest},
			)},
			Label: "Add {C}. Spend this mana only to cast a face-down creature spell, pay a mana cost to turn a manifested creature face up, or pay a morph cost",
		}},
	})
}
