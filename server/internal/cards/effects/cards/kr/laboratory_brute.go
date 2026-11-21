package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Laboratory Brute — Creature — Zombie Horror {3}{U}, 3/3 (Eldritch Moon,
// #5541):
//
//	"When this creature enters, mill four cards."
//
// Wailing Ghoul's enters trigger with four cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc92bbee-a445-4927-932b-15985bb93f19",
		Name:         "Laboratory Brute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Laboratory Brute — mill four cards", effects.Do(effects.MillCards{N: 4})),
		},
	})
}
