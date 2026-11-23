package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firecannon Blast — Sorcery {1}{R}{R} (Jumpstart 2022, #5948):
//
//	"Firecannon Blast deals 3 damage to target creature.
//	 Raid — Firecannon Blast deals 6 damage instead if you attacked this
//	 turn."
//
// Raid is read as the spell resolves (SpellRaid, Arrow Storm's reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97bff952-3bc8-41d0-b03f-5c76f3c7e0a3",
		Name:         "Firecannon Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := 3
			if effects.SpellRaid()(ctx.Game, item) {
				amount = 6
			}
			return effects.DamageToFirstTargetStep(amount)(item, ctx)
		},
	})
}
