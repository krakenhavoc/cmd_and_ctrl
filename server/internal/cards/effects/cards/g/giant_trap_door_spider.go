package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giant Trap Door Spider — Creature — Spider {1}{R}{G}, 2/3 (Ice Age, S82 OLD #5636):
//
//	"{1}{R}{G}, {T}: Exile this creature and target creature without flying
//	 that's attacking you."
//
// Both are exiled as the ability resolves; the target must be attacking you
// (the player, CR 506.3) and lack flying, as chosen and again on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c87565b-3902-4b11-b894-6922f66008ba",
		Name:         "Giant Trap Door Spider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{G}, {T}: Exile this creature and target creature without flying that's attacking you.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature without flying that's attacking you", effects.AttackingYou(), effects.WithoutKeyword("flying")),
			Effect:  effects.SpiderExileB831,
		}},
	})
}
