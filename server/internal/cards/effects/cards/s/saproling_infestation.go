package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saproling Infestation — Enchantment {1}{G} (Invasion):
//
//	"Whenever a player kicks a spell, you create a 1/1 green Saproling
//	 creature token."
//
// Rumbling Aftershocks's kicked-cast reading (the cast record carries the
// times kicked, CR 702.33d), over ANY player's spell instead of its
// controller's: the trigger is built once per kicked spell, however many times
// it was kicked, and its controller, not the caster, creates the token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd257eb3-4112-41e5-9c7d-00d1f683dbdb",
		Name:         "Saproling Infestation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && effects.S82OldB489KickedTimesOfCast(g, ev.CardID) > 0
			}, "Saproling Infestation — you create a 1/1 green Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
		},
	})
}
