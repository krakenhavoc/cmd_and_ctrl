package cardse

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evereth, Viceroy of Plunder — Legendary Creature — Vampire Soldier
// {2}{B}, 2/2 (Foundations Jumpstart):
//
//		"Flying
//		 Sacrifice another creature or artifact: Put a +1/+1 counter on
//		 Evereth. If the sacrificed permanent was a Treasure, Evereth gains
//		 lifelink until end of turn. Activate only as a sorcery.
//		 When Evereth dies, you may pay {1}{B/R}. When you do, Evereth deals
//		 damage equal to its power to each opponent."
//
//	  - The sacrifice is a cost, so "was a Treasure" is read off the
//	    sacrificed permanent's last-known subtypes (CR 608.2h).
//	  - The death trigger's "may pay" is MayPay; the paid branch is the
//	    CR 603.12 reflexive trigger, which carries Evereth's last-known
//	    power (CR 608.2h) and goes on the stack above the parent, so it can
//	    be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55b8c81b-f26b-4ae1-94ba-30b19e6a944e",
		Name:            "Evereth, Viceroy of Plunder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice another creature or artifact: Put a +1/+1 counter on Evereth. If the sacrificed permanent was a Treasure, Evereth gains lifelink until end of turn. Activate only as a sorcery.",
			Cost:         effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				treasure := false
				if info, ok := ctx.SacrificedPermanent(); ok {
					treasure = slices.Contains(info.Characteristic.Subtypes, "Treasure")
				}
				if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if !treasure || !effects.OnBattlefield(g, ctx.Source()) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"lifelink"},
					Label: "Evereth, Viceroy of Plunder — lifelink until end of turn"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Evereth, Viceroy of Plunder — you may pay {1}{B/R}; when you do, it deals damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{B/R}",
						Question: "Evereth, Viceroy of Plunder — pay {1}{B/R}? When you do, it deals damage equal to its power to each opponent.",
						Then:     everethPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// everethPaid is the paid branch: "when you do", a reflexive trigger
// carrying Evereth's last-known power as its amount.
var everethPaid = effects.PayUnlessBranches("evereth-viceroy-of-plunder/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	power := 0
	if info, ok := ctx.TriggeringPermanent(); ok {
		power = max(info.Power, 0)
	}
	return effects.ReflexiveTrigger{
		Label:  "Evereth, Viceroy of Plunder — it deals damage equal to its power to each opponent",
		Body:   everethDamageBody,
		Params: game.EffectParams{Amount: power},
	}.Apply(ctx)
}, nil)

var everethDamageBody = game.ReflexiveBody("evereth-viceroy-of-plunder/damage",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Amount <= 0 {
			return nil
		}
		return g.DealDamageEachKeyedForEffect(item.SourceCardID, effects.NewContext(g, item).Opponents(), p.Amount, game.KeyedTail{})
	}, nil)
