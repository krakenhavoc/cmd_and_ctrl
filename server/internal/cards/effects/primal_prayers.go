package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Prayers — Enchantment {2}{G}{G}:
//
//	"When this enchantment enters, you get {E}{E} (two energy counters).
//	 You may cast creature spells with mana value 3 or less by paying
//	 {E} rather than paying their mana costs. If you cast a spell this
//	 way, you may cast it as though it had flash."
//
// The second line is ADR 0118 §3's granted offer, narrowed to creature
// spells with mana value 3 or less and priced at one energy (ADR 0129 §5,
// PayEnergyForSmallCreatureSpellsWithFlash). The flash belongs to the
// claim: CR 601.3c lets a spell that may be cast as though it had flash
// only if an alternative cost is paid be begun at instant speed, so the
// same creature cast for its printed cost keeps its own timing. A
// creature with {X} in its cost is cast with X = 0 this way (CR 107.3b),
// and its mana value is judged so.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d087fe0-d554-4d7c-ba22-32db2cf71887",
		Name:         "Primal Prayers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Primal Prayers", 2),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayEnergyForSmallCreatureSpellsWithFlash()},
	})
}
