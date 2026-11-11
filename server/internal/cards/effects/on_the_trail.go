package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5a060788-a601-4c01-8341-73398b497a85",
		Name:         "On the Trail",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "On the Trail — you may put a land card from your hand onto the battlefield tapped",
				Do(MayPutALandFromHandTapped("On the Trail"))),
		},
	})
}
