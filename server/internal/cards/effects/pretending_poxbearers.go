package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pretending Poxbearers — Creature — Human Citizen Ally {1}{W/B}, 2/1
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature dies, create a 1/1 white Ally creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fefef024-3a05-4221-bb74-419ee08716e5",
		Name:         "Pretending Poxbearers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Pretending Poxbearers — create a 1/1 Ally", Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
