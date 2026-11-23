package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Vigilantes — Creature — Dwarf {2}{R}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may have it deal damage equal to its
//	 power to target creature. If you do, this creature assigns no combat damage this turn."
//
// #3531: Farrel's Zealot's shape with the amount read from the Vigilantes' power as the ability
// resolves, before the assigns-no-combat-damage gate is applied.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b49baf80-237c-41ee-8a22-2b6a68a313e7",
		Name:         "Dwarven Vigilantes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.UnblockedAttackerMay("Dwarven Vigilantes", "deal damage equal to its power to target creature", "dwarven-vigilantes-damage",
				func(ctx *effects.Context, target uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: effects.SourcePowerNow(ctx)}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
	})
}
