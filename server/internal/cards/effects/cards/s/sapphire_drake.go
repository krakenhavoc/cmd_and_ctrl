package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Drake — Creature — Drake {5}{U}, 4/4 (Gatecrash, S82 OLD #5565):
//
//	"Flying
//	 Each creature you control with a +1/+1 counter on it has flying."
//
// Beast, Erudite Aerialist's layer 6 flying grant, over every creature its
// controller controls that has a +1/+1 counter on it, read live (a counter
// added or removed changes it at once). The Drake itself qualifies once it holds
// one, which it already has printed flying for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1987bb61-a31b-4d5a-910f-88e9cbc6c8b3",
		Name:            "Sapphire Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Counters[game.CounterPlusOne] > 0
			}, "flying"),
		},
	})
}
