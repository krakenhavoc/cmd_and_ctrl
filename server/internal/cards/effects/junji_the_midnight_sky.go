package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Junji, the Midnight Sky — Legendary Creature — Dragon Spirit {3}{B}{B},
// 5/5 (Tarkir: Dragonstorm Commander):
//
//	"Flying, menace
//	 When Junji dies, choose one —
//	 • Each opponent discards two cards and loses 2 life.
//	 • Put target non-Dragon creature card from a graveyard onto the
//	   battlefield under your control. You lose 2 life."
//
// Atsushi's modal dies trigger: the mode is chosen as the trigger goes on
// the stack (CR 603.3c), and the second bullet's target is asked only when
// it is chosen (a bullet with no legal target isn't offered). The first
// bullet makes every opponent lose 2 life, then asks each in turn to
// discard two cards (CR 701.9b). The second returns the card under your
// control and then you lose 2 life; a target gone by resolution makes the
// bullet do nothing (CR 608.2b), and the life loss goes with it.
//
// No simplification.
func init() {
	dies := WhenThisDies("Junji, the Midnight Sky — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = ChooseOne(
		ModeDoing("Each opponent discards two cards and loses 2 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := eachOpponentLosesLife(ctx.Game, item, 2); err != nil {
					return err
				}
				ctx.Game.EachPlayerDiscardsForEffect(ctx.Controller(), game.DiscardPrompt{
					Source: item.SourceCardID,
					N:      2,
				})
				return nil
			}),
		ModeDoing("Put target non-Dragon creature card from a graveyard onto the battlefield under your control. You lose 2 life.",
			TargetCardInGraveyard("target non-Dragon creature card from a graveyard", Creature(), Not(HasSubtype("Dragon"))),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
	)
	Register(Spec{
		OracleID:        "78a48f4e-ba4c-4d8d-938a-9ddb342a885b",
		Name:            "Junji, the Midnight Sky",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Triggered:       []game.TriggeredAbility{dies},
	})
}
