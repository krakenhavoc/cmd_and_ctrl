package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galactic Wayfarer — Creature — Human Scout {2}{G}, 3/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, create a Lander token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: Search your library for a
//	 basic land card, put it onto the battlefield tapped, then
//	 shuffle.")"
//
// Beamsaw Prospector's Lander (b16LanderToken) made by an entry trigger
// in place of a dies trigger. The search ability rides the token
// template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86746a84-3e8a-44b4-8494-60cbee9816b5",
		Name:         "Galactic Wayfarer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Galactic Wayfarer — create a Lander token",
				effects.Do(effects.CreateToken{Template: effects.B16LanderToken(), N: 1})),
		},
	})
}
