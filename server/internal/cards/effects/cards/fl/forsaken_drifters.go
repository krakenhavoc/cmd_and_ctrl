package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forsaken Drifters — Creature — Zombie {3}{B}, 4/2 (Born of the Gods,
// #5559):
//
//	"When this creature dies, mill four cards."
//
// Aven Heartstabber's dies trigger with the mill alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e94f685-f392-40a8-8ad4-683351877fd9",
		Name:         "Forsaken Drifters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Forsaken Drifters — mill four cards", effects.Do(effects.MillCards{N: 4})),
		},
	})
}
