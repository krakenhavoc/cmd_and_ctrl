package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murmuring Mystic — Creature — Human Wizard {3}{U}, 1/5 (EDHREC
// rank 1816):
//
//	"Whenever you cast an instant or sorcery spell, create a 1/1 blue
//	 Bird Illusion creature token with flying."
//
// Talrand on a wall. The condition is instantOrSorceryCastByYou
// and the token is b16BlueBirdIllusionToken, whose flying is real
// (carried on its Keywords). The trigger goes on the stack above the
// spell that caused it, so the Bird is on the battlefield before the
// spell resolves, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcd4da46-5438-4454-8b1b-43ca51bda1f9",
		Name:         "Murmuring Mystic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Murmuring Mystic — create a 1/1 Bird Illusion with flying", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Bird Illusion with flying"), N: 1})),
		},
	})
}
