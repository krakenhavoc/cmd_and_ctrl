package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "6846448b-d5a0-4ae0-bb11-ea7fa6ed2558",
		Name:         "Haliya, Ascendant Cadet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Haliya, Ascendant Cadet — put a +1/+1 counter on target creature you control",
				Do(Instruction(counterOnFirstCreatureTarget))),
				TargetCreature("target creature you control", YouControl())),
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hasCounterOfKind(c, game.CounterPlusOne) },
				"Haliya, Ascendant Cadet — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
