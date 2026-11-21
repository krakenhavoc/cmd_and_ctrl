package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eidolon of Astral Winds — Enchantment Creature — Spirit 2/4 {2}{W}
// (Foundations Jumpstart):
//
//	"Vigilance
//	 Constellation — Whenever this creature or another enchantment you
//	 control enters, choose target creature you control. Until end of turn,
//	 that creature has base power and toughness 4/4 and gains flying."
//
// The target is chosen as the trigger goes on the stack; the layer 7b base
// P/T sits under counters and pumps (CR 613.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c9b8b82-7f61-4dc0-aeac-fba5dc115e91",
		Name:            "Eidolon of Astral Winds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Eidolon of Astral Winds — target creature you control has base power and toughness 4/4 and gains flying until end of turn",
				effects.EidolonOfAstralWindsPump),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
