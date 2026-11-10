package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// For the Common Good — Sorcery {X}{X}{G} (Bloomburrow, #4467):
//
//	"Create X tokens that are copies of target token you control. Then
//	 tokens you control gain indestructible until your next turn. You gain
//	 1 life for each token you control."
//
// Three steps in order (CR 608.2c), each waiting for any prompt the one
// before asks (the copies can stop on a CR 616 ordering prompt):
//
//   - X copies of the target, one creation (CreateTokenCopy): the copiable
//     values only, never the original's tapped state, counters or Auras
//     (CR 707.2), and an "as enters" or "enters with" ability of the copy
//     works. {X}{X} is announced once (XMatters); X = 0 makes none.
//   - The grant is a scoped effect whose set is pinned when it is
//     registered (CR 611.2c): the tokens on the battlefield under the
//     caster's control at that moment, the new copies included, keep
//     indestructible until the start of the caster's next turn. A token
//     created later does not get it.
//   - The life is counted after both: every token the caster controls,
//     creature or not, the copies included.
//
// The target is the spell's one target, so a target that is gone by
// resolution fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de6e4aa2-23f1-46e5-a061-05522944e0c6",
		Name:         "For the Common Good",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPermanent("target token you control", IsTokenPredicate(), YouControl()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id := FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return ScopedEffectFor{
					Match:    And(IsTokenPredicate(), YouControl()),
					Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
					Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
					Label:    "For the Common Good — tokens you control gain indestructible until your next turn",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && IsToken(c) {
						n++
					}
				}
				return GainLife{Amount: n}.Apply(ctx)
			}),
		),
	})
}
