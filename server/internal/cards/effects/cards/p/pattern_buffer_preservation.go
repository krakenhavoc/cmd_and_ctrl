package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pattern Buffer Preservation — Sorcery {1}{B} (Star Trek, #4973):
//
//	"Look at the top two cards of your library. Put one of them into your
//	 hand and the other into your graveyard.
//	 Beam me up {1}{B} (You may cast this card from your graveyard for
//	 {1}{B} if you also return a creature you control to its owner's hand.
//	 Then exile this spell.)"
//
// The spell is Visionary's Dance's look-and-take: a private look at the top
// two, a mandatory pick of one for the hand, and the other put into the
// graveyard (takeRestGraveyard), not discarded or milled. A library of one
// card puts that card into the hand. Beam me up is BeamMeUp
// (trek_b8_helpers.go): flashback's cast from the graveyard and exile, with
// the return of a creature you control as the rest of the price.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "690d0d8f-48b9-40a6-b881-6f17ac19bf3b",
		Name:             "Pattern Buffer Preservation",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.BeamMeUp("{1}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 2),
				Max:    1,
				Label:  "Pattern Buffer Preservation — put one of them into your hand and the other into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
