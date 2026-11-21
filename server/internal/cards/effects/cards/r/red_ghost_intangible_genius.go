package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Red Ghost, Intangible Genius — Legendary Creature — Human Scientist
// Villain {2}{U}{R}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Ward {2}
//	 Red Ghost can't be blocked.
//	 Whenever you draw your second card each turn, create a 3/3 red Ape
//	 Villain creature token with haste."
//
// Ward is the shared trigger, unblockability a self restriction (Blighted
// Agent's), and the draw trigger reads the turn's draw tally exactly as
// Roxxon Brutes does, so it fires on the second draw and no later one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9984d950-29dc-4d6b-8750-8d1bf86ea98f",
		Name:         "Red Ghost, Intangible Genius",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Red Ghost — ward {2}"),
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Red Ghost — create a 3/3 red Ape Villain creature token with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 red Ape Villain with haste"), N: 1})),
		},
	})
}
