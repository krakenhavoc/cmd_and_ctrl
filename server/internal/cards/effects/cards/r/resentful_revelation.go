package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resentful Revelation — Sorcery {1}{B} (Final Fantasy, #4462):
//
//	"Look at the top three cards of your library. Put one of them into
//	 your hand and the rest into your graveyard.
//	 Flashback {6}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Commune with Evil's look-and-take over three cards, with nothing after
// the pick (the rest go to the graveyard as the take's continuation), and
// the shared Flashback cast path.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "71d6aa12-4888-458d-9dc8-9954b247a246",
		Name:             "Resentful Revelation",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{6}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Resentful Revelation — put one of them into your hand and the rest into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
