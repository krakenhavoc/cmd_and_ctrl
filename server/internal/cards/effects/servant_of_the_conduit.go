package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servant of the Conduit — Creature — Elf Druid {1}{G}, 2/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Add one mana of any color."
//
// ADR 0129 §5: a creature's own mana ability, so summoning sickness
// applies to its {T} (CR 302.6), and the energy is checked before
// anything is paid (CR 118.3). The auto-tapper plans it in its energy
// tier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df1a8846-b4da-4e9b-8cf3-ee687e4c606b",
		Name:         "Servant of the Conduit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Servant of the Conduit", 2),
		},
		ManaAbilities: []ManaAbility{anyColorForEnergyRow(1)},
	})
}
