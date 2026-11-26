package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tuktuk the Explorer — Legendary Creature — Goblin {2}{R} 1/1 (Commander 2014):
//
//	"Haste
//	 When Tuktuk the Explorer dies, create Tuktuk the Returned, a
//	 legendary 5/5 colorless Goblin Golem artifact creature token."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "754e227b-ff31-4b5c-b612-2221c2b44246",
		Name:            "Tuktuk the Explorer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tuktuk the Explorer — create Tuktuk the Returned, a legendary 5/5 colorless Goblin Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 colorless Legendary Goblin Golem artifact named Tuktuk the Returned"), N: 1})),
		},
	})
}
