package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of Urabrask — Creature — Phyrexian Human Cleric {2}{R}, 2/1 (New Phyrexia, #5575):
//
//	"When this creature enters, add {R}{R}{R}."
//
// Priest of Gix's enters trigger in red: the mana is added as the trigger
// resolves and empties with the step (CR 500.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4bd8910-770b-4220-8a26-2673491f4a3e",
		Name:         "Priest of Urabrask",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Priest of Urabrask — add {R}{R}{R}", effects.Do(effects.AddMana{Produced: "{R}{R}{R}"})),
		},
	})
}
