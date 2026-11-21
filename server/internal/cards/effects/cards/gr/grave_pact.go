package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grave Pact — Enchantment {1}{B}{B}{B}:
//
//	"Whenever a creature you control dies, each other player
//	sacrifices a creature of their choice."
//
// The archetype's board-control engine. One sacrifice outlet plus
// Grave Pact turns every creature you were going to lose anyway into a
// one-sided edict, and with a recurring body (Reassembling Skeleton,
// any token producer) it dismantles three boards at once.
//
// "Of their choice" is the rules content, and the reason this needs a
// per-player prompt rather than a loop the controller drives: each
// opponent picks their own worst creature, not the one you would pick.
//
// It is also not targeted. A hexproof or protected creature is a
// perfectly legal choice, "can't be the target of spells or abilities"
// does nothing, and Grave Pact resolves happily when nobody has a
// creature to give up.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f4ac4a4-53ec-4bc9-8f5c-d4b801d867b2",
		Name:         "Grave Pact",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Grave Pact — each other player sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Creature(),
				Label:            "a creature",
			})), game.Purpose{DeathPayoff: true, Answers: game.AnswerRemove}),
		},
	})
}
