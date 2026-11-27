package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyes of the Wisent — Kindred Enchantment — Elemental {1}{G} (Lorwyn, S82 OLD,
// #5593):
//
//	"Whenever an opponent casts a blue spell during your turn, you may create a
//	 4/4 green Elemental creature token."
//
// Witchstalker's cast trigger (an opponent's spell of a colour, on its
// controller's turn), optional at resolution (CR 603.5).
//
// No simplification.
func init() {
	opponentBlue := effects.AnOpponentCast(effects.OfColor("U"))
	effects.Register(effects.Spec{
		OracleID:     "3c59ab03-8744-4148-be06-eb1b0147f2cc",
		Name:         "Eyes of the Wisent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.IsActivePlayer(g, source.Controller) && opponentBlue(ev, source, lki, g)
			}, "Eyes of the Wisent — you may create a 4/4 green Elemental",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Elemental"), N: 1})),
				"Eyes of the Wisent — create a 4/4 green Elemental?"),
		},
	})
}
