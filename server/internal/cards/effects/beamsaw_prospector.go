package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bbe740f8-6e7c-4cfd-9134-9bf8d5f48cf5",
		Name:         "Beamsaw Prospector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Beamsaw Prospector — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
		},
	})
}
