package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Disk — Creature — Human Wizard {2}{W}{W}, 2/4 (Time
// Spiral, #5596):
//
//	"This creature enters tapped.
//	 {1}, {T}: Destroy all artifacts, creatures, and enchantments."
//
// Nevinyrral's Disk on a creature: the tapped entry is the self-
// replacement, and the {T} is a creature's, so the Magus waits out
// summoning sickness as well (CR 302.6). The Magus is a creature, so it
// is destroyed with the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3996aa0c-3322-43be-bbaf-7b66836a8ff4",
		Name:         "Magus of the Disk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Destroy all artifacts, creatures, and enchantments.",
			// ADR 0126 §6: artifacts, creatures and enchantments, the smallest listed class that holds all three.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  tspB2DestroyArtifactsCreaturesEnchantments,
		}},
	})
}
