package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Araña, Heart of the Spider — {1}{R}{W} Legendary Creature — Spider
// Human Hero 3/3 (Marvel's Spider-Man promos, #4460):
//
//	"Whenever you attack, put a +1/+1 counter on target attacking
//	 creature.
//	 Whenever a modified creature you control deals combat damage to a
//	 player, exile the top card of your library. You may play that card
//	 this turn."
//
// "Whenever you attack" is Bast's shape: once per declaration
// (OncePerBatch), with the attacking creature chosen as the trigger goes
// on the stack and re-checked as it resolves.
//
// The second trigger is one per creature and per player it hits. A
// creature is modified (CR 700.9) when it has a counter, is equipped, or
// is enchanted by an Aura its controller controls, read as the damage is
// dealt (b07IsModified). The exile is the ordinary impulse grant, good
// for this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a16037f2-a0f6-465d-8d6e-cec34b4a6320",
		Name:         "Araña, Heart of the Spider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Araña, Heart of the Spider — put a +1/+1 counter on target attacking creature",
					effects.Do(effects.Instruction(effects.CounterOnFirstCreatureTarget))),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
			effects.On(game.EventDealDamage, effects.ModifiedCreatureDealtCombatDamageToPlayer, "Araña, Heart of the Spider — exile the top card of your library; you may play it this turn",
				effects.OutpostSiegeImpulse),
		},
	})
}
