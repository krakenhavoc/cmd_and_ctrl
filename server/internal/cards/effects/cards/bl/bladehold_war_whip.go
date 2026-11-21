package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladehold War-Whip — Artifact — Equipment {1}{R}{W} (Phyrexia: All Will
// Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equip abilities you activate of other Equipment cost {1} less to
//	 activate.
//	 Equipped creature has double strike.
//	 Equip {3}{R}{W}"
//
// #5479, S82. Bureau Headmaster's equip discount (the equip bit on the
// activation's cost query) restricted to other permanents you control
// (Boom Scholar's "other"), so the War-Whip's own {3}{R}{W} is not
// discounted. A second War-Whip does discount the first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3142c6b-86e9-42b7-916c-ce7c34e7bd53",
		Name:         "Bladehold War-Whip",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ForMirrodin("Bladehold War-Whip")},
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsLess(1, "Equip abilities you activate of other Equipment cost {1} less to activate.",
				effects.AnEquipAbilityCost(), effects.YouActivate(), effects.OfAnotherPermanentYouControl()),
		},
		Static:    []game.StaticAbility{effects.GrantToAttached("double strike")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}{R}{W}")},
	})
}
