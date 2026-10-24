package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ebony Charm — Instant {B}:
//
//	"Choose one —
//	 • Target opponent loses 1 life and you gain 1 life.
//	 • Exile up to three target cards from a single graveyard.
//	 • Target creature gains fear until end of turn. (It can't be
//	   blocked except by artifact creatures and/or black creatures.)"
//
// #1807, ADR 0106 §5. The second bullet is Decompose's clause on a
// mode. Each bullet reads its own target group (ModeTarget), never
// another bullet's: only one is chosen, but the read is the right one
// either way. Fear is a canonical keyword the block check honours.
// The targets are chosen as the spell is cast (the 2004-10-04 ruling),
// which is when every mode's clause is announced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f3237fe-9e07-4aad-8bc8-7ef4b510dd36",
		Name:         "Ebony Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			// The gain waits for a prompt the life loss asks (a CR 616 order; CR 608.2c, #4155).
			ModeSteps("Target opponent loses 1 life and you gain 1 life.", TargetPlayer("target opponent", Opponent()),
				Instruction(func(ctx *Context) error {
					t, ok := ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -1)
				}),
				Instruction(func(ctx *Context) error {
					t, ok := ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				})),
			ModeDoing("Exile up to three target cards from a single graveyard.", upToNCardsFromASingleGraveyard(3),
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return exileTargetCardsThen(ctx, nil)
				}),
			ModeWithPurpose(ModeDoing("Target creature gains fear until end of turn.", TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"fear"}}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 0, 0, "fear"))),
		),
	})
}
