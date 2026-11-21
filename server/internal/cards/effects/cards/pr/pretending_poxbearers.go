package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pretending Poxbearers — Creature — Human Citizen Ally {1}{W/B}, 2/1
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature dies, create a 1/1 white Ally creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fefef024-3a05-4221-bb74-419ee08716e5",
		Name:         "Pretending Poxbearers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Pretending Poxbearers — create a 1/1 Ally", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
