package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forbidden Alchemy — Instant {2}{U} (Tarkir: Dragonstorm Commander):
//
//	"Look at the top four cards of your library. Put one of them into
//	 your hand and the rest into your graveyard.
//	 Flashback {6}{B}"
//
// Commune with Evil's look-and-take over four cards without the life:
// a private look, a mandatory pick of one for the hand, the rest put
// into the graveyard (not milled or discarded). With fewer than four
// cards in the library it looks at all of them (the printed ruling).
// Flashback is the shared alternative cost, exiling the spell as it
// leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "66d65cda-e3d1-4a00-8ff3-c6d5ef0ab0c1",
		Name:             "Forbidden Alchemy",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    1,
				Label:  "Forbidden Alchemy — put one of them into your hand and the rest into your graveyard",
				Key:    forbiddenAlchemyTake,
			}.Apply(ctx)
		},
	})
}

// forbiddenAlchemyTake is the look's continuation: the rest into the
// graveyard.
var forbiddenAlchemyTake = RegisterTake("choose-cards/forbidden-alchemy-take",
	TakeDef{Then: TakeRestIntoGraveyard})
