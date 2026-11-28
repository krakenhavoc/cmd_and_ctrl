package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tolarian Emissary (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Human Wizard {2}{U}, 1/2:
//	"Kicker {1}{W}
//	Flying
//	When this creature enters, if it was kicked, destroy target enchantment."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8042ade2-f589-4c3e-a461-a4c284a67ee4",
		Name:            "Tolarian Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Tolarian Emissary — kicked, destroy target enchantment",
					effects.Do(effects.Instruction(effects.S82OldB648DestroyTheTarget))),
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
		},
	})
}
