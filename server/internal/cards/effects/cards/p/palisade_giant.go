package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Palisade Giant — Creature — Giant Soldier {4}{W}{W}, 2/7:
//
//	"All damage that would be dealt to you and other permanents you
//	 control is dealt to this creature instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) to the
// Giant itself, of damage to its controller and to every other permanent
// they control that can be dealt damage. The ruling: two Giants are two
// effects, you choose which applies, and the damage is never split.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f38fe1e9-8997-4b63-9109-7513034bac88",
		Name:         "Palisade Giant",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.StaticRedirection("Palisade Giant — damage to you and your other permanents is dealt to it instead",
				effects.RedirectWhere{Applies: effects.DamageToYouOrYourOtherPermanents, To: effects.ToThisPermanent}),
		},
	})
}
