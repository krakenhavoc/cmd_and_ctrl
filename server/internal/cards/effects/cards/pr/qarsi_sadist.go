package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qarsi Sadist — Creature — Human Cleric {1}{B}, 1/3 (Dragons of Tarkir,
// #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, target opponent loses 2 life and you gain 2 life."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff
// is a reflexive trigger raised only when a creature was really
// sacrificed, and a targeted payoff chooses its target then.
// The payoff is Dakmor Ghoul's drain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69e15fd8-9531-4392-8b85-b539c138ea91",
		Name:         "Qarsi Sadist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Qarsi Sadist", qarsiSadistDrainBody),
		},
	})
}

// qarsiSadistDrainBody is "When this creature exploits a creature, target opponent loses 2 life and you gain 2 life." An on-disk key:
// never renamed, never reused.
var qarsiSadistDrainBody = game.ReflexiveBody("qarsi-sadist/drain", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.TargetOpponentLosesAndYouGain(g, item, 2)
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetPlayer("target opponent", effects.Opponent()) }))
