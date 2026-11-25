package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Depala, Pilot Exemplar — Legendary Creature — Dwarf Pilot {1}{R}{W}, 3/3
// (Secret Lair Drop):
//
//	"Other Dwarves you control get +1/+1.
//	 Each Vehicle you control gets +1/+1 as long as it's a creature.
//	 Whenever Depala becomes tapped, you may pay {X}. If you do, reveal the
//	 top X cards of your library, put all Dwarf and Vehicle cards from among
//	 them into your hand, then put the rest on the bottom of your library in
//	 a random order."
//
// The first two are layer-7c bonuses (the second only for Vehicles that are
// creatures). The trigger fires when Depala becomes tapped (tapping for a cost,
// by an effect, or attacking without vigilance); X is chosen as it resolves and
// paid as a may-pay, and the paid half reveals, takes every Dwarf and Vehicle
// without a prompt, and puts the rest under the library in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc913f28-18ef-485e-8d2b-ae0115523020",
		Name:         "Depala, Pilot Exemplar",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Dwarf"}, Others: true, YoursOnly: true}, 1, 1),
			effects.S82OldB024VehiclesYouControlPlusOne(),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Depala, Pilot Exemplar — you may pay {X}; if you do, reveal the top X cards, put all Dwarf and Vehicle cards among them into your hand, the rest on the bottom in a random order",
				effects.S82OldB024DepalaAsk),
		},
	})
}
