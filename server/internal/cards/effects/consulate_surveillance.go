package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consulate Surveillance — Enchantment {3}{W}:
//
//	"When this enchantment enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 Pay {E}{E}: Prevent all damage that would be dealt to you this turn
//	 by a source of your choice."
//
// ADR 0129 §2 (#1995): "Pay {E}{E}" is the energy cost component,
// removed from the activator before the ability goes on the stack
// (CR 107.14, CR 602.1a). The shield is ADR 0108 §7's (#1904): the
// source is chosen as the ability resolves (CR 609.7a), and every
// instance of its damage to the controller this turn is prevented.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "06d62e22-df9c-488e-8d2d-793eaf4cd75c",
		Name:         "Consulate Surveillance",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Consulate Surveillance", 4),
		},
		Activated: []ActivatedAbility{sourceShieldRow(
			"Pay {E}{E}: Prevent all damage that would be dealt to you this turn by a source of your choice.",
			PayEnergy(2), nil,
			PreventDamageFromChosenSource(ShieldYou))},
	})
}
