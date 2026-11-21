package cardssy

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "8bed92cc-b3f8-4201-bcbc-44535ff1bd96",
		Name:            "Tainted Adversary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tainted Adversary — pay {2}{B} any number of times; counters, then twice that many decayed Zombies", taintedAdversaryEnters),
		},
	})
}

var taintedAdversaryBody = game.ReflexiveBody("tainted-adversary/counters-and-zombies",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: p.Amount}).Apply(ctx); err != nil {
			return err
		}
		return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 2 * p.Amount}.Apply(ctx)
	}, nil)

var taintedAdversaryPaid = effects.PayUnlessBranches("pay-unless/tainted-adversary-times",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Tainted Adversary — " + strconv.Itoa(r.Number) + " +1/+1 counters, then " + strconv.Itoa(2*r.Number) + " decayed Zombies",
			Body:   taintedAdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func taintedAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return effects.MayPayX{
		Symbol: "{2}{B}",
		Label:  "Tainted Adversary",
		Buys:   "put X +1/+1 counters on it and create twice X 2/2 Zombies with decayed",
		Goal:   effects.XAsHighAsYouCan,
		Then:   taintedAdversaryPaid,
	}.Apply(effects.NewContext(g, item))
}
