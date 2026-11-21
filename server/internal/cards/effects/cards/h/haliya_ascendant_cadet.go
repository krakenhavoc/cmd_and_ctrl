package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haliya, Ascendant Cadet — Legendary Creature — Human Soldier
// {2}{G}{W}{W}, 3/3 (Edge of Eternities, #4461):
//
//	"Whenever Haliya enters or attacks, put a +1/+1 counter on target
//	 creature you control.
//	 Whenever one or more creatures you control with +1/+1 counters on
//	 them deal combat damage to a player, draw a card."
//
// The first trigger is Picard, Steadfast Captain's enters-or-attacks
// shape with Builder's Talent's one counter on the chosen creature, which
// may be Haliya herself. The second is Mu Yanling's batched combat-damage
// trigger (one draw per combat damage step however many creatures
// connect, CR 603.2c), over creatures with a +1/+1 counter as the damage
// is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6846448b-d5a0-4ae0-bb11-ea7fa6ed2558",
		Name:         "Haliya, Ascendant Cadet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersOrAttacks("Haliya, Ascendant Cadet — put a +1/+1 counter on target creature you control",
				effects.Do(effects.Instruction(effects.CounterOnFirstCreatureTarget))),
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return effects.HasCounterOfKind(c, game.CounterPlusOne)
				},
				"Haliya, Ascendant Cadet — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
