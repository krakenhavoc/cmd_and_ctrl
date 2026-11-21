package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainsaw — Artifact — Equipment {1}{R} (Duskmourn: House of Horror,
// #4466):
//
//	"When this Equipment enters, it deals 3 damage to up to one target
//	 creature.
//	 Whenever one or more creatures die, put a rev counter on this
//	 Equipment.
//	 Equipped creature gets +X/+0, where X is the number of rev counters
//	 on this Equipment.
//	 Equip {3}"
//
// The enters trigger is Blood Spatter Analysis's, with "up to one" target.
// The death trigger fires once for a batch of deaths, anyone's creatures
// (OncePerBatch over EventLTB). The pump counts the rev counters on the
// Equipment itself, re-read on every layer pass (Excalibur II's shape).
//
// No simplification.
func init() {
	const rev = "rev"
	effects.Register(effects.Spec{
		OracleID:     "3db40361-5f55-417e-a7cd-7e360cc91b4d",
		Name:         "Chainsaw",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, func(_ *game.Game, source *game.Card) int { return source.Counters[rev] }),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Chainsaw — 3 damage to up to one target creature", effects.DealDamageToTheTarget(3)),
				effects.UpToOneTargetCreature("up to one target creature")),
			effects.OncePerBatch(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Chainsaw — put a rev counter on it", effects.Do(effects.CounterOnThis{Kind: rev, N: 1}))),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
