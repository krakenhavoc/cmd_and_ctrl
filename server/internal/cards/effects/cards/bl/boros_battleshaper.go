package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boros Battleshaper — Creature — Minotaur Soldier {5}{R}{W}, 5/5 (Dragon's
// Maze, S82 OLD b405, #5564):
//
//	"At the beginning of each combat, up to one target creature attacks or
//	 blocks this combat if able and up to one target creature can't attack or
//	 block this combat."
//
// Two optional target clauses (Monstrous Step's pair), chosen as the trigger
// goes on the stack (CR 603.3d). Each is read by its slot, so a trigger that
// names only the second creature still restricts it, and each skips a target
// that is illegal as it resolves (CR 608.2b). Both effects last for this
// combat phase (CR 511.3): an additional combat is a different combat. The
// first creature gets an attack requirement and a block requirement together
// (CR 508.1d, 509.1c); whichever applies in that combat is what binds it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf5bf1ef-e40b-4fb5-8148-d4ca7a307501",
		Name:         "Boros Battleshaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfEachCombat("Boros Battleshaper — up to one creature attacks or blocks, up to one can't attack or block", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					d, ok := effects.DurationUntilEndOfCombat(ctx)
					if !ok {
						return nil
					}
					if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
						if err := (effects.ScopedEffectFor{
							Target: t.ID,
							Mods: []game.Mod{
								game.AddAttackRequirementMod(uuid.Nil),
								game.AddBlockRequirementMod(game.BlockRequirementBlocks),
							},
							Duration: d,
							Label:    "Boros Battleshaper — attacks or blocks this combat if able",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
						return effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
							Duration: d,
							Label:    "Boros Battleshaper — can't attack or block this combat",
						}.Apply(ctx)
					}
					return nil
				}),
				effects.Clauses(
					effects.UpToOneTargetCreature("up to one target creature"),
					effects.UpToOneTargetCreature("up to one target creature"),
				)),
		},
	})
}
