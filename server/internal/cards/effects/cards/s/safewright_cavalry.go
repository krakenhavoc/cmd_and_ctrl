package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safewright Cavalry — Creature — Elf Warrior {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"This creature can't be blocked by more than one creature.
//	 {5}: Target Elf you control gets +2/+2 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e04a375-303e-43ba-bca1-4e959711a7f6",
		Name:         "Safewright Cavalry",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: Target Elf you control gets +2/+2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    effects.ManaCost("{5}"),
			Targets: effects.TargetCreature("target Elf you control", effects.YouControl(), effects.HasSubtype("Elf")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.VsPumpTheTarget(effects.NewContext(g, item), 2, 2, "Safewright Cavalry — +2/+2 until end of turn")
			},
		}},
	})
}
