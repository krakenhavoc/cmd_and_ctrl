package cardsp

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phantasmal Sphere — Creature — Illusion {1}{U} 0/1 (Alliances, S82 OLD
// batch 819):
//
//	"Flying
//	 At the beginning of your upkeep, put a +1/+1 counter on this creature,
//	 then sacrifice this creature unless you pay {1} for each +1/+1 counter on
//	 it.
//	 When this creature leaves the battlefield, target opponent creates an X/X
//	 blue Orb creature token with flying, where X is the number of +1/+1
//	 counters on this creature."
//
// The counter lands first and the payment is counted from the counters that
// are there then (all of them, whoever put them on). X is read from the
// creature as it left (CR 608.2h) and the token enters under the opponent's
// control (CR 111.2).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4230c7a3-4591-411e-afab-d2bfb58446de",
		Name:            "Phantasmal Sphere",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Phantasmal Sphere — put a +1/+1 counter on it, then sacrifice it unless you pay {1} for each +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					if !effects.OnBattlefield(g, source) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return g.AddCounterByKeyedForEffect(uuid.Nil, source, game.CounterPlusOne, 1, effects.NewContext(g, item).ItemTail(phantasmalSpherePay))
				}),
			{
				Watches:   []game.EventKind{game.EventLTB},
				AppliesTo: effects.Self,
				Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
				Key:       "Phantasmal Sphere — target opponent creates an X/X blue Orb creature token with flying",
				Effect:    phantasmalSphereOrb,
			},
		},
	})
}

var phantasmalSpherePay = effects.CardTail("phantasmal-sphere/pay-upkeep", func(ctx *effects.Context, _ game.TailOutcome) error {
	source := ctx.Item.SourceCardID
	n := effects.CounterCountOn(ctx.Game, source, game.CounterPlusOne)
	if n <= 0 {
		return nil
	}
	return effects.UpkeepPayUnless{
		Chooser:  ctx.Controller(),
		Cost:     strings.Repeat("{1}", n),
		Question: "Phantasmal Sphere — pay {1} for each +1/+1 counter on it, or sacrifice it",
		Else:     phantasmalSphereSacrifice,
	}.Apply(ctx)
})

var phantasmalSphereSacrifice = effects.PayUnlessBranches("pay-unless/phantasmal-sphere-sacrifice", nil,
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.SacrificePermanent{Target: r.Source}.Apply(ctx)
	})

func phantasmalSphereOrb(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := 0
	if item.Trigger != nil && item.Trigger.Object != nil {
		x = item.Trigger.Object.Counters[game.CounterPlusOne]
	}
	opp := effects.TargetedPlayerOf(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return effects.CreateToken{
		Controller: opp,
		Template: game.Card{
			Name:      "Orb",
			TypeLine:  "Token Creature — Orb",
			Colors:    []string{"U"},
			Keywords:  []string{"flying"},
			Power:     x,
			Toughness: x,
		},
		N: 1,
	}.Apply(ctx)
}
