package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Season of the Bold — Sorcery {3}{R}{R} (Bloomburrow, #4467):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Create a tapped Treasure token.
//	 {P}{P} — Exile the top two cards of your library. Until the end of your
//	 next turn, you may play them.
//	 {P}{P}{P} — Until the end of your next turn, whenever you cast a spell,
//	 Season of the Bold deals 2 damage to up to one target creature."
//
// A pawprint spell (CR 700.2i, Season of the Burrow's shape); nothing targets
// as the spell is cast. The Treasure is created tapped (CreateTokenAdvanced).
// The exile is the impulse exile with the window closing at the end of your
// next turn (ExileTopNUntilYourNextTurn, UntilEndOfYourNextTurnDuration).
// The third bullet registers a repeating delayed trigger (CR 603.7) held by
// the spell's controller until the end of their next turn: it fires on each
// spell they cast, even one that is then countered, and resolves before the
// spell that set it off. Its target clause, "up to one target creature", is
// announced as each instance goes on the stack (TargetedDelayedBody), so
// each may pick a different creature or none, and the damage is dealt by the
// spell (as it last existed) at that time (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5fe29c25-a7ad-4c79-a5a5-da1bbc832141",
		Name:         "Season of the Bold",
		Completeness: CompletenessFull,
		Modes: ChooseUpToPawprints(5, true,
			Pawprint(1, ModeDoing("Create a tapped Treasure token.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 1}.Apply(ctx)
				})),
			Pawprint(2, ModeDoing("Exile the top two cards of your library. Until the end of your next turn, you may play them.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return ExileTopNUntilYourNextTurn(ctx, 2)
				})),
			Pawprint(3, ModeDoing("Until the end of your next turn, whenever you cast a spell, Season of the Bold deals 2 damage to up to one target creature.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					until := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
					ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
						Controller:   ctx.Controller(),
						SourceCardID: ctx.Source(),
						Label:        "Season of the Bold — 2 damage to up to one target creature",
						On:           []game.EventKind{game.EventCast},
						Condition:    youNextCastCondition,
						Body:         seasonOfTheBoldDamageBody,
						Repeats:      true,
						Duration:     &until,
					})
					return nil
				})),
		),
	})
}

// seasonOfTheBoldDamageBody is the delayed trigger's resolution: 2 damage to
// the still-legal target, from the spell that made it. An on-disk key.
var seasonOfTheBoldDamageBody = game.TargetedDelayedBody("season-of-the-bold/damage", simpleBody(
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to one target creature").WithCount(0, 1)
	}))
