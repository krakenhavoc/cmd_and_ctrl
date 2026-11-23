package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heron-Blessed Geist — Creature — Spirit {4}{W}, 3/3 (Innistrad: Crimson
// Vow):
//
//	"Flying
//	 {3}{W}, Exile this card from your graveyard: Create two 1/1 white
//	 Spirit creature tokens with flying. Activate only if you control an
//	 enchantment and only as a sorcery."
//
// Flying is printed data. The ability works from the graveyard (Zones),
// exiles the card as a cost (ExileThis), is sorcery-speed, and is gated on
// controlling an enchantment at activation (CR 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c5897cd-f209-4e51-adc2-d791e5671029",
		Name:            "Heron-Blessed Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{W}, Exile this card from your graveyard: Create two 1/1 white Spirit creature tokens with flying. Activate only if you control an enchantment and only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Condition:    effects.VowB12ControlsAnEnchantment(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Spirit with flying"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
