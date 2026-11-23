package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// On the Trail — Enchantment {1}{G} (Murders at Karlov Manor Commander, #5368):
//
//	"Whenever you draw your second card each turn, you may put a land card from
//	 your hand onto the battlefield tapped."
//
// The second-draw condition is Wolfbat's and Bard the Bowman's (the tally is
// read after the draw, so it fires once a turn, on the second card and not on
// a third). The put is the shared MayPutALandFromHandTapped: a put, not a play,
// optional, the land arriving tapped. The card just drawn is in the hand, so it
// is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a060788-a601-4c01-8341-73398b497a85",
		Name:         "On the Trail",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "On the Trail — you may put a land card from your hand onto the battlefield tapped",
				effects.Do(effects.MayPutALandFromHandTapped("On the Trail"))),
		},
	})
}
