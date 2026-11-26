package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tracker's Instincts — Sorcery {1}{G}:
//
//	"Reveal the top four cards of your library. Put a creature card from
//	 among them into your hand and the rest into your graveyard.
//	 Flashback {2}{U}"
//
// The creature is not optional: with one among the four, one comes to hand.
func init() {
	const label = "Tracker's Instincts — reveal the top four cards, put a creature card into your hand and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:         "1bc7c38b-9571-4395-be90-922f92816987",
		Name:             "Tracker's Instincts",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, label),
				Match:  effects.Creature(),
				Max:    1,
				Label:  label,
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
