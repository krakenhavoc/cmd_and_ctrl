package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikey & Mona, Mutant Sitters — {2}{G} Legendary Creature — Mutant
// Ninja Turtle Lizard 2/2 (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"When Mikey & Mona enter, choose one or both. Each mode must target a
//	 different player.
//	 • Target player chooses a creature they control and puts two +1/+1
//	   counters on it.
//	 • Target player returns a creature or land card from their graveyard
//	   to their hand."
//
// Casey & Raph's frame: a modal enters trigger whose two player clauses
// are Distinct, so the bullets name two different players. Each bullet's
// choice is the TARGET player's, made at resolution and not targeting: the
// creature to grow (the target player is the one who puts the counters),
// and the creature or land card to return. A player with nothing to choose
// gets nothing.
//
// No simplification.
func init() {
	etb := WhenThisEnters("Mikey & Mona, Mutant Sitters — choose one or both",
		func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseN("Choose one or both", 1, 2,
		ModeDoing("Target player chooses a creature they control and puts two +1/+1 counters on it.",
			Distinct(TargetPlayer("target player")),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return tmcB8TargetPlayerPutsCountersOnACreature(ctx, t.ID, 2,
					"Mikey & Mona — choose a creature you control to put two +1/+1 counters on")
			}),
		ModeDoing("Target player returns a creature or land card from their graveyard to their hand.",
			Distinct(TargetPlayer("target player")),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return tmcB8TargetPlayerReturnsFromGraveyard(ctx, t.ID,
					func(c game.Card) bool { return c.IsCreature() || c.IsLand() },
					"Mikey & Mona — return a creature or land card from your graveyard to your hand")
			}),
	)
	Register(Spec{
		OracleID:     "15e14a91-a596-465f-becc-cef2e7fbbd30",
		Name:         "Mikey & Mona, Mutant Sitters",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
