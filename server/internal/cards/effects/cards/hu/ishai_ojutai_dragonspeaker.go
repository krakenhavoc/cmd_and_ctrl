package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ishai, Ojutai Dragonspeaker — Legendary Creature — Bird Monk {2}{W}{U},
// 1/1: "Flying. Whenever an opponent casts a spell, put a +1/+1 counter on
// Ishai. Partner" (partner is a deck-construction rule, ADR 0144).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4aae7102-301a-429d-b169-4f462384d55d",
		Name:            "Ishai, Ojutai Dragonspeaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(nil),
				"Ishai, Ojutai Dragonspeaker — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
