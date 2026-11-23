package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravening Warg — Creature — Wolf {1}{B}, 2/2 (The Hobbit):
//
//	"Deathtouch
//	 Ferocious — Whenever this creature attacks while you control a
//	 creature with power 4 or greater, you gain 2 life."
//
// See hobbitFerociousAttack for the "while" window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8060d231-7477-441d-b249-8e75aa1d7791",
		Name:            "Ravening Warg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.HobbitFerociousAttack, "Ravening Warg — ferocious: you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
