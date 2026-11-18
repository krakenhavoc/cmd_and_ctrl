package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Infusion — Instant {U}{R} (Midnight Hunt, #5497):
//
//	"Look at the top four cards of your library. You may reveal an instant
//	 or sorcery card from among them and put it into your hand. Put the
//	 rest on the bottom of your library in a random order.
//	 Flashback {3}{U}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Monumental Henge's dig (takeRevealRestBottomRandom) over instant and
// sorcery cards: the pick is optional, at most one, revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d5251d47-11de-41f7-9965-1f921e12d927",
		Name:             "Arcane Infusion",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Match:    Or(Instant(), Sorcery()),
				Max:      1,
				Optional: true,
				Label:    "Arcane Infusion — you may reveal an instant or sorcery card and put it into your hand",
				Key:      takeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
