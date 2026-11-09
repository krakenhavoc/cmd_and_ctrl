package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Magic — Instant {1}{U} (Final Fantasy, #4462):
//
//	"Tiered (Choose one additional cost.)
//	 • Blizzard — {0} — Return target creature to its owner's hand.
//	 • Blizzara — {2} — Target creature's owner puts it on their choice
//	   of the top or bottom of their library.
//	 • Blizzaga — {5}{U} — Target creature's owner shuffles it into
//	   their library."
//
// Thunder Magic's tiered shape. Blizzara is Happy Hogan's owner's
// choice at depth one (the top); Blizzaga is Chaos Warp's tuck followed
// by the owner's shuffle (thenShuffle). The mover is the card's owner in
// both, so a stolen creature goes back to its true owner's library.
//
// No simplification.
func init() {
	bullet := func(label, cost string, apply func(ctx *Context, id uuid.UUID) error) game.ModeOption {
		return ModeWithPurpose(SpreeModeDoing(label, cost, TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, picked := ModeTarget(ctx, occ)
				if !picked {
					return nil
				}
				return apply(ctx, t.ID)
			}), game.Purpose{Answers: game.AnswerRemove})
	}
	Register(Spec{
		OracleID:     "88b0bd3b-d832-4059-b0d9-04630917ae94",
		Name:         "Ice Magic",
		Completeness: CompletenessFull,
		Modes: Tiered(
			bullet("Blizzard — Return target creature to its owner's hand.", "{0}",
				func(ctx *Context, id uuid.UUID) error { return BounceToHand{Target: id}.Apply(ctx) }),
			bullet("Blizzara — Target creature's owner puts it on their choice of the top or bottom of their library.", "{2}",
				func(ctx *Context, id uuid.UUID) error {
					return PutIntoLibraryAtDepthOrBottom{Card: id, Depth: 1, Label: "Ice Magic — Blizzara: put it on your choice of the top or bottom of your library"}.Apply(ctx)
				}),
			bullet("Blizzaga — Target creature's owner shuffles it into their library.", "{5}{U}",
				func(ctx *Context, id uuid.UUID) error {
					c, _ := ctx.Game.LookupCardForEffect(id)
					return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: c.Owner})
				}),
		),
	})
}
