package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burrowing Razormaw — Creature — Beast {2}{G}, 4/2 (The Brothers' War):
//
//	"When this creature dies, mill four cards."
//
// One dies trigger (WhenThisDies) whose effect is the controller milling
// four, from the library as it is when the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8551e1df-47eb-419e-8dd2-e96b9a450912",
		Name:         "Burrowing Razormaw",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Burrowing Razormaw — mill four cards", effects.Do(effects.MillCards{N: 4})),
		},
	})
}
