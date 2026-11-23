package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// District Mascot — Creature — Dog Mount {G}:
//
//	"This creature enters with a +1/+1 counter on it.
//	 {1}{G}, Remove two +1/+1 counters from this creature: Destroy
//	 target artifact.
//	 Whenever this creature attacks while saddled, put a +1/+1 counter
//	 on it.
//	 Saddle 1"
//
// The entry counter is a self-replacement, so it is on the creature
// before any state check sees the printed body. The counter cost is the
// ADR 0020 counter-removal component, paid at announce, so the ability
// cannot be activated with fewer than two counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fa7cf3a-2a4d-4dd7-890b-49c497570b16",
		Name:         "District Mascot",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "District Mascot: enters with a +1/+1 counter"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{G}, Remove two +1/+1 counters from this creature: Destroy target artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 2)),
				Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
				Effect:  effects.DestroyFirstLegalTarget,
			},
			effects.Saddle(1),
		},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("District Mascot — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
