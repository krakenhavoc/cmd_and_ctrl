package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var cyclopsSuperconductorPaid = effects.EnergyPaid("pay-unless/cyclops-superconductor-damage",
	func(g *game.Game, item *game.StackItem) error {
		t := effects.WhenYouDo("Cyclops Superconductor — damage equal to its power to any target", effects.CyclopsSuperconductorDamageBody)
		t.Params = game.EffectParams{Amount: effects.B13LastKnownPower(g, item.SourceCardID, game.Characteristic{})}
		return t.Apply(effects.NewContext(g, item))
	})

// Cyclops Superconductor — Creature — Cyclops Wizard {1}{U}{R}, 2/2:
//
//	"Prowess
//	 When this creature enters, you get {E}{E}{E} (three energy counters).
//	 When this creature dies, you may pay {E}{E}{E}. When you do, this
//	 creature deals damage equal to its power to any target."
//
// ADR 0129 §3 (#1995): the energy is paid as the dies trigger resolves
// (CR 118.12); "when you do" is a reflexive trigger (CR 603.12) whose
// target is chosen as it goes on the stack. "Its power" is the power
// it last had on the battlefield (CR 608.2h), fixed when the energy is
// paid, and the creature is the damage's source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10bca1b1-4f92-40e8-b9c0-d08c534903e1",
		Name:            "Cyclops Superconductor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Purpose:         game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Cyclops Superconductor", 3),
			effects.WhenThisDies("Cyclops Superconductor — you may pay {E}{E}{E}",
				effects.MayPayEnergyThen("Cyclops Superconductor", 3, "deal damage equal to its power to any target", cyclopsSuperconductorPaid)),
		},
	})
}
