package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mega Flare — Sorcery {2}{R} (Final Fantasy, #4462):
//
//	"Kicker {3}{R}{R}
//	 If this spell was kicked, create a 6/6 red Dragon creature token with
//	 flying.
//	 For each opponent, choose up to one target creature that player
//	 controls. Mega Flare deals damage equal to the greatest power among
//	 creatures you control to each of the chosen creatures."
//
// Shellshock's per-opponent target clauses, each pick bound to its opponent
// so one that changed hands in response is illegal (CR 608.2b). The Dragon
// is made first, in printed order (CR 608.2c), so a kicked Mega Flare counts
// it: the damage is the greatest layered power among the caster's creatures
// as the spell resolves (CR 608.2h), the Dragon's 6 at least. With every
// target gone the spell doesn't resolve, and the Dragon is not made (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "891c6ca5-d238-4ac2-8df4-c6dc9a807562",
		Name:          "Mega Flare",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{3}{R}{R}")},
		Targets:       perOpponentClauses(game.MaxPlayers-1, "creature", Creature()),
		Purpose:       game.Purpose{Answers: game.AnswerRemove, Kicked: &game.Purpose{Tokens: 1}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return CreateToken{Template: b14RedDragonToken(6), N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				return DealDamageToEachThen(ctx, ids, b42GreatestPowerControlledBy(ctx.Game, ctx.Controller()), RecipientThen{})
			}),
		),
	})
}
