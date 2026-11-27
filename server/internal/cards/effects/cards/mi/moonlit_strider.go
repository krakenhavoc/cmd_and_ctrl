package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonlit Strider — Creature — Spirit {3}{W}, 1/4 (Betrayers of Kamigawa, S82 OLD b596):
//
//	"Sacrifice this creature: Target creature you control gains protection from the color
//	 of your choice until end of turn.
//	 Soulshift 3"
//
// Midvast Protector's effect behind a sacrifice cost: the target is chosen as the ability is
// activated, the colour as it resolves (the colour the creature is protected FROM). Soulshift
// 3 is the shared keyword trigger (CR 702.46a), so a Strider sacrificed as the cost still
// dies and may return a Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8524a6a9-4d4c-4796-aad3-0f08f825a239",
		Name:         "Moonlit Strider",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Soulshift(3, "Moonlit Strider")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature you control gains protection from the color of your choice until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")), game.AnswerProtect),
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ProtectionFromAChosenColorForTheTarget("Moonlit Strider")(item, effects.NewContext(g, item))
			},
		}},
	})
}
