package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghave, Guru of Spores — Legendary Creature — Fungus Shaman {2}{W}{B}{G},
// 0/0:
//
//	"Ghave enters with five +1/+1 counters on it.
//	 {1}, Remove a +1/+1 counter from a creature you control: Create a 1/1
//	 green Saproling creature token.
//	 {1}, Sacrifice a creature: Put a +1/+1 counter on target creature."
//
// The entry counters are a CR 614.1c clause (Bloatfly Swarm's). The first
// cost removes the counter from a creature chosen as the cost is paid
// (Sage of Fables), Ghave itself or another. The second sacrifices any
// creature you control as a cost (the one the counter goes on may not be
// the one sacrificed: it is chosen as the ability is put on the stack, and
// a sacrificed target leaves the ability with nothing to do, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e579a72f-4933-40fe-9e57-96f8d65370bc",
		Name:         "Ghave, Guru of Spores",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 5, "Ghave, Guru of Spores: enters with five +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, Remove a +1/+1 counter from a creature you control: Create a 1/1 green Saproling creature token.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
			},
			{
				Label:   "{1}, Sacrifice a creature: Put a +1/+1 counter on target creature.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "a creature", effects.Creature())),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
		},
	})
}
