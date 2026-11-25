package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Refuse // Cooperate — split card with aftermath (Commander 2019):
//
//	Refuse {3}{R}, Instant: "Refuse deals damage to target spell's controller
//	 equal to that spell's mana value."
//	Cooperate {2}{U}, Instant: "Aftermath (Cast this spell only from your
//	 graveyard. Then exile it.) Copy target instant or sorcery spell. You may
//	 choose new targets for the copy."
//
// Refuse reads the spell's controller and mana value (X counted, CR 202.3e)
// as it resolves, from last-known information if it has since left the
// stack (CR 608.2h); a target no longer legal means no damage (CR 608.2b).
// Cooperate is Reverberate's copy (effects.CmdrB21CopyTheTargetSpell): any instant or sorcery spell, the copy
// made under its caster's control with the CR 707.10c retarget choice.
// Aftermath is the engine's (CR 702.127a), as for Road // Ruin.
//
// No simplification.
func init() {
	const oracleID = "7db22241-fe43-4071-9af9-de1cf394f7f5"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Refuse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			controller, ok := ctx.Game.SpellControllerForEffect(targets[0].ID)
			if !ok {
				return nil
			}
			n := effects.SpellManaValueForEffect(ctx.Game, targets[0].ID)
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: n}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Cooperate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve:    effects.CmdrB21CopyTheTargetSpell,
	})
}
