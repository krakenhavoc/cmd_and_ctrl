package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fumigate — Sorcery {3}{W}{W}:
//
//	"Destroy all creatures. You gain 1 life for each creature
//	 destroyed this way."
//
// A Wrath of God that pays you for the board you were losing to,
// which at a four-player table is routinely eight or ten life —
// enough that the wipe leaves you ahead rather than merely even.
//
// "Destroyed THIS WAY" is the clause the primitive's count exists
// for. It is not "creatures that died this turn" and not "creatures
// on the battlefield": a creature that was already gone, or one
// whose destruction was replaced away entirely, is not counted. The
// Then callback receives exactly the number that left, so the life
// gain cannot drift from the sweep.
//
// A commander hit by this IS counted, and that is right: CR 903.9
// replaces the zone change, not the destruction — the commander was
// destroyed, it just went somewhere else. Its owner is asked first,
// though, so the life gain happens when they answer rather than on
// this line (#815).
//
// A destruction a replacement rewrote into an EXILE is the case that
// looks the same and is counted the other way: CR 701.7a defines
// destroying a permanent as moving it to its owner's GRAVEYARD, so a
// permanent that was exiled instead was never destroyed, however
// thoroughly it left. See ADR 0013 §5i.
func init() {
	Register(Spec{
		OracleID:     "b17ea905-0696-4e58-b564-557e87236e27",
		Name:         "Fumigate",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, _ []game.Card, destroyed int) error {
					if destroyed <= 0 {
						return nil
					}
					return GainLife{Player: ctx.Controller(), Amount: destroyed}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
