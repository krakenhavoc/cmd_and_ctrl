package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Adversary — Creature — Zombie {1}{B}, 2/3:
//
//	"Deathtouch
//	 When this creature enters, you may pay {2}{B} any number of times.
//	 When you pay this cost one or more times, put that many +1/+1
//	 counters on this creature, then create twice that many 2/2 black
//	 Zombie creature tokens with decayed."
//
// The Adversary cycle's shape (Bloodthirsty Adversary): MayPayX chooses N,
// the payment makes a CR 603.12 reflexive trigger carrying N, and its
// resolution puts the counters on, then makes 2N decayed Zombies.
func init() {
	Register(Spec{
		OracleID:        "8bed92cc-b3f8-4201-bcbc-44535ff1bd96",
		Name:            "Tainted Adversary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tainted Adversary — pay {2}{B} any number of times; counters, then twice that many decayed Zombies", taintedAdversaryEnters),
		},
	})
}

var taintedAdversaryBody = game.ReflexiveBody("tainted-adversary/counters-and-zombies",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		if err := (CounterOnThis{Kind: game.CounterPlusOne, N: p.Amount}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 2 * p.Amount}.Apply(ctx)
	}, nil)

var taintedAdversaryPaid = PayUnlessBranches("pay-unless/tainted-adversary-times",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Tainted Adversary — " + strconv.Itoa(r.Number) + " +1/+1 counters, then " + strconv.Itoa(2*r.Number) + " decayed Zombies",
			Body:   taintedAdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func taintedAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{2}{B}",
		Label:  "Tainted Adversary",
		Buys:   "put X +1/+1 counters on it and create twice X 2/2 Zombies with decayed",
		Goal:   XAsHighAsYouCan,
		Then:   taintedAdversaryPaid,
	}.Apply(NewContext(g, item))
}
