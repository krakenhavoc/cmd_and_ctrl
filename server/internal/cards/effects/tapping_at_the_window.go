package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tapping at the Window — Sorcery {1}{G} (Midnight Hunt, #5497):
//
//	"Look at the top three cards of your library. You may reveal a
//	 creature card from among them and put it into your hand. Put the rest
//	 into your graveyard.
//	 Flashback {2}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The look-and-take with the pick optional, at most one creature card,
// revealed, and the rest into the graveyard (a take with both the reveal
// and the graveyard rest, registered here).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d5b80d04-1650-4b5d-9f9b-091ecd17b6e8",
		Name:             "Tapping at the Window",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Match:    Creature(),
				Max:      1,
				Optional: true,
				Label:    "Tapping at the Window — you may reveal a creature card and put it into your hand",
				Key:      midB03TakeRevealRestGraveyard,
			}.Apply(ctx)
		},
	})
}
