package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Lotus Tile — Artifact {4}:
//
//	"This artifact enters tapped.
//	 {T}: Add X mana of any one color, where X is the greatest number
//	 of creatures you control that have a creature type in common."
//
// #742's one pick of N tokens. X is the size of the largest group of
// creatures you control sharing one creature type: for each type, the
// creatures that have it, plus every changeling (CR 702.73a — a
// changeling has every type, so it joins every group). A creature with
// no creature type joins none. Types are effective types, so a Maskwood
// Nexus or a type-granting lord counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8a5e009-45b7-4e62-a494-1579f5fc0ba6",
		Name:         "White Lotus Tile",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(largestSharedCreatureTypeGroup),
			Label:        "Add X mana of any one color (X = most creatures you control sharing a type)",
		}},
	})
}

// largestSharedCreatureTypeGroup is the greatest number of creatures
// `controller` controls that have a creature type in common.
func largestSharedCreatureTypeGroup(g *game.Game, controller, _ uuid.UUID) int {
	counts := map[string]int{}
	changelings, best := 0, 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if game.HasAllCreatureTypes(&c) {
			changelings++
			continue
		}
		for _, t := range game.CreatureTypesOf(&c) {
			counts[t]++
			if counts[t] > best {
				best = counts[t]
			}
		}
	}
	if best == 0 && changelings == 0 {
		return 0
	}
	return best + changelings
}
