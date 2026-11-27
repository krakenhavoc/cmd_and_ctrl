package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repel Intruders — Instant {3}{W/U} (Shadowmoor, S82 older sets):
//
//	"Create two 1/1 white Kithkin Soldier creature tokens if {W} was spent to
//	 cast this spell. Counter up to one target creature spell if {U} was spent
//	 to cast this spell. (Do both if {W}{U} was spent.)"
//
// River's Grasp's colour-spent reads (ReadsOwnPayment, WantsColors), each half
// independent: white spent makes the tokens, blue spent counters the target
// creature spell if one was chosen and is still legal (CR 608.2b). The tokens
// come first, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "94a241bf-5709-4890-8f06-017833869d62",
		Name:           "Repel Intruders",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W", "U"},
		Purpose:        game.Purpose{Tokens: 2, CreatureTokens: 2, Answers: game.AnswerRemove},
		Targets:        effects.TargetSpell("up to one target creature spell", effects.Creature()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("W") > 0 {
				if err := (effects.CreateToken{Template: effects.TokenCard("1/1 white Kithkin Soldier"), N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("U") > 0 && len(item.Targets) > 0 && ctx.IsTargetLegal(item.Targets[0]) {
				return effects.CounterTheTargetSpell(item, ctx)
			}
			return nil
		},
	})
}
