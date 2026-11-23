package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kianne, Corrupted Memory — Legendary Creature — Illusion {2}{G}{U},
// 2/2 (Duskmourn Commander, #5368):
//
//	"As long as Kianne's power is even, you may cast noncreature spells
//	 as though they had flash.
//	 As long as Kianne's power is odd, you may cast creature spells as
//	 though they had flash.
//	 Whenever you draw a card, put a +1/+1 counter on Kianne."
//
// Both flash grants are cast-timing declarations (Quicksilver,
// Speedster's shape) gated on the live parity of Kianne's power, read
// through the layers and counters on every query, so each draw flips
// which kind of spell may be cast at instant speed. Exactly one is open
// at a time. The counter trigger is one per card drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fc9e800-da66-4dda-a56d-79daec9e943e",
		Name:         "Kianne, Corrupted Memory",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{
			effects.Cmdr2S22CastWhilePowerParity(game.PermissionFilter{NoncreatureOnly: true}, true,
				"As long as Kianne's power is even, you may cast noncreature spells as though they had flash."),
			effects.Cmdr2S22CastWhilePowerParity(game.PermissionFilter{CreatureOnly: true}, false,
				"As long as Kianne's power is odd, you may cast creature spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Kianne — put a +1/+1 counter on Kianne", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
