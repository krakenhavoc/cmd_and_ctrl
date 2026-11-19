package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "4c9b8b82-7f61-4dc0-aeac-fba5dc115e91",
		Name:            "Eidolon of Astral Winds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, constellationTriggers, "Eidolon of Astral Winds — target creature you control has base power and toughness 4/4 and gains flying until end of turn",
				eidolonOfAstralWindsPump),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
