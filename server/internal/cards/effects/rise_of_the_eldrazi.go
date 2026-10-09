package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Eldrazi — Sorcery {9}{C}{C}{C}:
//
//	"This spell can't be countered.
//	 Destroy target permanent. Target player draws four cards. Take an
//	 extra turn after this one.
//	 Exile Rise of the Eldrazi."
//
// Two target clauses, each judged on its own at resolution (CR 608.2b):
// a permanent that left in response costs the card only the destruction,
// and a player who left costs it only the draw. The spell fizzles only
// when both are gone, and the engine's all-illegal short-circuit has
// already handled that before OnResolve runs.
//
// The extra turn is CR 500.7 through TakeExtraTurn, and the spell
// exiles itself last, as the printed order has it: Temporal Mastery's
// ExileCardForEffect on the resolving card, so it is not left in the
// graveyard for a recursion deck to find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "982f70af-077f-40df-b2fe-7c80ebcc7712",
		Name:            "Rise of the Eldrazi",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		// Slot 0 is destroyed (removal, not declared here); slot 1 draws four.
		Purpose: ForTargets(game.TargetPurpose{Slot: 1, Draws: 4}),
		Targets: Clauses(
			TargetPermanent("target permanent"),
			TargetPlayer("target player"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if len(item.Targets) > 1 && ctx.IsTargetLegal(item.Targets[1]) {
				if err := (DrawCards{Player: item.Targets[1].ID, N: 4}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (TakeExtraTurn{}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
