package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ponyback Brigade — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Goblin Warrior {3}{R}{W}{B}, 2/2: When this creature enters or is
//	turned face up, create three 1/1 red Goblin creature tokens.
//	Morph {2}{R}{W}{B}

func init() {
	effects.Register(effects.Spec{
		OracleID:         "8cfb5181-8ab7-484b-8821-a62c561f762f",
		Name:             "Ponyback Brigade",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{R}{W}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, effects.Self,
				"Ponyback Brigade — create three 1/1 red Goblin creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 3})),
		},
	})
}
