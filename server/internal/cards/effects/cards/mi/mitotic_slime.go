package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mitotic Slime — Creature — Ooze {4}{G}, 4/4:
//
//	"When this creature dies, create two 2/2 green Ooze creature
//	 tokens. They have "When this token dies, create two 1/1 green
//	 Ooze creature tokens.""
//
// The Slime's death makes two 2/2 Oozes, and each Ooze is the printed
// token card that carries its own dies trigger (a catalog token
// template), making two plain 1/1 Oozes in turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1162c7ff-c0cb-469a-ad3c-322b0981ae1d",
		Name:         "Mitotic Slime",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mitotic Slime — create two 2/2 green Ooze creature tokens that split when they die",
				effects.Do(effects.CreateToken{Template: effects.MitoticSlimeOozeToken(), N: 2})),
		},
	})
}
