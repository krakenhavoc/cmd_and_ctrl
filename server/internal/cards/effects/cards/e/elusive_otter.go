package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elusive Otter // Grove's Bounty — an adventurer card, Secrets of
// Strixhaven Commander reprint (#4834):
//
//	Elusive Otter {U}, Creature — Otter 1/1:
//	"Prowess
//	 Creatures with power less than this creature's power can't block
//	 it."
//	Grove's Bounty {X}{G}, Sorcery — Adventure:
//	"Distribute X +1/+1 counters among any number of target creatures
//	 you control."
//
// The block rule is Howlgeist's: compared as blockers are declared,
// against the Otter's power then, so a prowess pump widens it. The
// counters are divided as the spell is cast (CR 601.2d), each target
// getting at least one; a target gone by resolution loses its share
// rather than passing it on (PutDividedCounters). The engine owns the
// adventure: the spell resolves into exile and the Otter may be cast
// from there.
//
// No simplification.
const ElusiveOtterOracleID = "54175132-2c44-4749-8dfd-d08dcc63e4b3"

func init() {
	effects.Register(effects.Spec{
		OracleID:        ElusiveOtterOracleID,
		Name:            "Elusive Otter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Elusive Otter's can't block it"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     ElusiveOtterOracleID + "#1",
		Name:         "Grove's Bounty",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
