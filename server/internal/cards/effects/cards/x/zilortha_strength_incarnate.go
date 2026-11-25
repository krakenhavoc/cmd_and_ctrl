package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zilortha, Strength Incarnate — Legendary Creature — Dinosaur {3}{R}{G},
// 7/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 Lethal damage dealt to creatures you control is determined by their
//	 power rather than their toughness."
//
// Characteristic.LethalByPower on every creature its controller controls
// (Zilortha too), a layer 6 static read through the layers. Card.
// LethalDamageThreshold is the one place the lethal-damage state-based
// action (CR 704.5g), the combat damage assignment (CR 702.19b), excess
// damage (CR 120.4a) and the legal-move enumerator ask, so a creature of
// power 2 dies to 2 damage whatever its toughness. Damage is only lethal
// once some is marked, so a creature of power 0 or less is destroyed by 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2dba8d8d-891e-482e-8014-879f701d0619",
		Name:            "Zilortha, Strength Incarnate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Lethal damage dealt to creatures you control is determined by their power rather than their toughness.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.LethalByPower = true
			},
		}},
	})
}
