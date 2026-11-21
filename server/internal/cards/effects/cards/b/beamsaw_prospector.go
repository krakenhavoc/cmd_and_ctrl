package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beamsaw Prospector — Creature — Human Artificer {1}{B}, 2/1 (Edge of
// Eternities, #4461):
//
//	"When this creature dies, create a Lander token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: Search your library for a
//	 basic land card, put it onto the battlefield tapped, then
//	 shuffle.")"
//
// A dies trigger (graveyard only) that makes Horizon Explorer's Lander
// (b16LanderToken), whose search ability rides the token template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbe740f8-6e7c-4cfd-9134-9bf8d5f48cf5",
		Name:         "Beamsaw Prospector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Beamsaw Prospector — create a Lander token",
				effects.Do(effects.CreateToken{Template: effects.B16LanderToken(), N: 1})),
		},
	})
}
