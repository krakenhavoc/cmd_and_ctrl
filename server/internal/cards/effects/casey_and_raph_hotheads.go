package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Casey & Raph, Hotheads — {4}{R} Legendary Creature — Mutant Ninja
// Human Turtle 4/4 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"When Casey & Raph enter, choose one or both. Each mode must target a
//	 different player.
//	 • Target player exiles the top card of their library. Until that
//	   player's next end step, they may play that card without paying its
//	   mana cost.
//	 • Target player creates two Treasure tokens."
//
// A modal enters trigger (CR 700.2b): the modes are chosen as it goes on
// the stack, then each chosen bullet's target. "Each mode must target a
// different player" is both clauses Distinct, so whichever bullet is
// targeted second can't name the player the first one did; either bullet
// may name the trigger's controller.
//
// The first bullet is an impulse exile held by the target player: free (an
// alternative cost, CR 118.9; additional costs still apply, CR 118.9d; X is
// 0, CR 107.3b), for a land a land play under the ordinary timing (the
// rulings), until that
// player's next end step begins. The second makes the Treasures under the
// target player's control.
//
// No simplification.
func init() {
	etb := WhenThisEnters("Casey & Raph, Hotheads — choose one or both",
		func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseN("Choose one or both", 1, 2,
		ModeDoing("Target player exiles the top card of their library. Until that player's next end step, they may play that card without paying its mana cost.",
			Distinct(TargetPlayer("target player")),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				window := ctx.Game.UntilYourNextEndStepDuration(t.ID)
				return ExileTopWithPermission{From: t.ID, GrantToOwner: true, N: 1, Free: true, Window: &window}.Apply(ctx)
			}),
		ModeDoing("Target player creates two Treasure tokens.",
			Distinct(TargetPlayer("target player")),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return CreateToken{Controller: t.ID, Template: TreasureToken(), N: 2}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "f20a5297-31ba-4bbd-810e-be23175a116f",
		Name:         "Casey & Raph, Hotheads",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
