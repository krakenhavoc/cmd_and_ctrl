package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heirloom Epic — Artifact {1} (Bloomburrow, #4467):
//
//	"{4}, {T}: Draw a card. For each mana in this ability's activation cost,
//	 you may tap an untapped creature you control rather than pay that mana.
//	 Activate only as a sorcery."
//
// Waterbend's tap-for-mana machinery over creatures alone (TapCreaturesForManaCost,
// std_b27_helpers.go): each creature tapped pays for one of the {4}. Tapping
// them is not the {T} symbol, so a creature that came in this turn may pay
// (CR 302.6), and the Epic itself can't pay for its own {T}. It is not waterbending.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43037ef5-09e8-4881-beff-ed38cf7574c6",
		Name:         "Heirloom Epic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}: Draw a card. For each mana in this ability's activation cost, you may tap an untapped creature you control rather than pay that mana. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.TapCreaturesForManaCost("{4}")),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
