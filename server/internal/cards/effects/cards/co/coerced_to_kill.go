package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coerced to Kill — Enchantment — Aura {3}{U}{B} (Murders at Karlov
// Manor, #4469):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Enchanted creature has base power and toughness 1/1, has
//	 deathtouch, and is an Assassin in addition to its other types."
//
// Control Magic's layer-2 static, with Honest Work's layer-7b base 1/1
// (counters and pumps still apply on top), a layer-6 deathtouch grant
// and a layer-4 Assassin type read through the Aura's text (CR 612.2).
// Control reverts when the Aura leaves, and CR 302.6 gives the creature
// summoning sickness under its new controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a007dbd-18fd-44d1-a305-7cc8dd6231eb",
		Name:         "Coerced to Kill",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.SetAttachedBasePT(1, 1),
			effects.GrantToAttached("deathtouch"),
			effects.AttachedIsAlso("Assassin"),
		},
	})
}
