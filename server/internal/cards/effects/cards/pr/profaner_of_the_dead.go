package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profaner of the Dead — Creature — Snake Wizard {3}{U}, 3/3 (Dragons of Tarkir, #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, return to their owners' hands all creatures your opponents control with toughness less than the exploited creature's toughness."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// The exploited creature's toughness is read as it last existed (CR 608.2h,
// CR 603.10a), and the creatures bounced are the opponents' with strictly less
// toughness as the payoff resolves.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6bfac15-ec3c-4a55-946a-0b4514eabafd",
		Name:         "Profaner of the Dead",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: only the opponents' creatures with toughness below the exploited creature's.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Profaner of the Dead", profanerOfTheDeadBounceBody),
		},
	})
}

// profanerOfTheDeadBounceBody is "When this creature exploits a creature, return to their owners' hands all creatures your opponents control with toughness less than the exploited creature's toughness." An on-disk key:
// never renamed, never reused.
var profanerOfTheDeadBounceBody = game.ReflexiveBody("profaner-of-the-dead/bounce", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	exploited := ctx.PayloadCards()
	if len(exploited) == 0 {
		return nil
	}
	toughness := effects.ToughnessNowOrLastKnown(g, exploited[0])
	return effects.BounceAllMatching{
		Match: effects.And(effects.Creature(), effects.OpponentControls(), effects.ToughnessLE(toughness-1)),
	}.Apply(ctx)
}), nil)
