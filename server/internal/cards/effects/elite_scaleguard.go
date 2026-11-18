package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elite Scaleguard — Creature — Human Soldier {4}{W}, 2/3:
//
//	"When this creature enters, bolster 2. (Choose a creature with the
//	 least toughness among creatures you control and put two +1/+1 counters
//	 on it.)
//	 Whenever a creature you control with a +1/+1 counter on it attacks,
//	 tap target creature defending player controls."
//
// Bolster is the shared CR 701.39 effect. The attack trigger fires for each
// of your attackers that has a +1/+1 counter as it is declared (the Scaleguard
// included), once each. "Defending player" is the one that attacker attacks
// (CR 508.1b), read off the declared attack (Distracting Geist's reading, for
// an attacker that need not be the source). With no legal creature the
// trigger is not put on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b90af42d-edb6-4654-b7dc-10f0f272c883",
		Name:         "Elite Scaleguard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elite Scaleguard — bolster 2", Do(Bolster{N: 2})),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackerWithPlusCounterYouControl(ev, source, g)
				},
				TargetsFrom: targetCreatureTheAttackedPlayerControls,
				Key:         "Elite Scaleguard — tap target creature defending player controls",
				Effect:      b36TapChosenCreature,
			},
		},
	})
}
