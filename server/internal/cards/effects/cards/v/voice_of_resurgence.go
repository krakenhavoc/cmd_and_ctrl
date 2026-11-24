package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voice of Resurgence — Creature — Elemental {G}{W}, 2/2 (Secret Lair Drop):
//
//	"Whenever an opponent casts a spell during your turn and when this
//	 creature dies, create a green and white Elemental creature token with
//	 "This token's power and toughness are each equal to the number of
//	 creatures you control.""
//
// One printed ability with two conditions (OnAny, as WhenThisEntersOrDies is):
// an opponent's cast while it is the Voice controller's turn, or the Voice
// dying. The token is a printed template with a layer 7a CDA, so it counts
// its own controller's creatures, itself included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cbbb3f3-63a4-4983-81ea-8c405b10e63f",
		Name:         "Voice of Resurgence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventCast, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventLTB {
						return effects.ThisDied(ev, source, lki, g)
					}
					return ev.Kind == game.EventCast && effects.ByAnOpponent(ev, source, lki, g) && effects.IsYourTurn(g, source.Controller)
				},
				"Voice of Resurgence — create a green and white Elemental token with power and toughness equal to the number of creatures you control",
				effects.Do(effects.CreateToken{Template: effects.OldB10VoiceElementalToken(), N: 1})),
		},
	})
}
