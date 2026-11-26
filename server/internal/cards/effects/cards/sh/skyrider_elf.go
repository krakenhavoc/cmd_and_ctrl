package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyrider Elf — Creature — Elf Warrior Ally {X}{G}{U}, 0/0 (Battle for
// Zendikar):
//
//	"Flying
//	 Converge — This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "c71f19be-aabc-44ef-a62b-57522b3a2c98",
		Name:                "Skyrider Elf",
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"flying"},
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerColorSpent(game.CounterPlusOne),
		},
	})
}
