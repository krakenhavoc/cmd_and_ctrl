package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minister of Pain — Creature — Human Shaman {2}{B}, 2/3 (Dragons of Tarkir, #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, creatures your opponents control get -1/-1 until end of turn."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// The creatures shrunk are the ones the opponents control as the payoff resolves (CR 611.2c).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d5b63a4-c9d9-4971-9d4e-2eab86f251e2",
		Name:         "Minister of Pain",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: only the opponents' creatures shrink.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Minister of Pain", ministerOfPainShrinkBody),
		},
	})
}

// ministerOfPainShrinkBody is "When this creature exploits a creature, creatures your opponents control get -1/-1 until end of turn." An on-disk key:
// never renamed, never reused.
var ministerOfPainShrinkBody = game.ReflexiveBody("minister-of-pain/shrink", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.BoostUntilEOT{
		Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -1, Toughness: -1,
		Label: "Minister of Pain — creatures your opponents control get -1/-1",
	}.Apply(effects.NewContext(g, item))
}), nil)
