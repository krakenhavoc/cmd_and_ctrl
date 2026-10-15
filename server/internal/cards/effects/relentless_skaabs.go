package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relentless Skaabs — Creature — Zombie {3}{U}{U}, 4/4 (#3537, ADR 0189
// slice 3):
//
//	"As an additional cost to cast this spell, exile a creature card from
//	 your graveyard.
//	 Undying (When this creature dies, if it had no +1/+1 counters on it,
//	 return it to the battlefield under its owner's control with a +1/+1
//	 counter on it.)"
//
// The cost is ExileFromGraveyardCost (#3536) matching a creature card, so a
// graveyard with none leaves the spell uncastable (CR 118.3). Undying is
// PrintedKeywords (#2075); the engine returns it with the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5089cd0f-1338-4ecd-bb1b-664d7e72f8fc",
		Name:            "Relentless Skaabs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		AdditionalCost: ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
	})
}
