package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slinza, the Spiked Stampede — Legendary Creature — Beast {4}{G}, 5/5
// (Foundations Jumpstart, #5929):
//
//	"Beast spells you cast cost {2} less to cast.
//	 Each other Beast creature you control enters with an additional +1/+1
//	 counter on it.
//	 Whenever Slinza or another creature with power 4 or greater enters, you
//	 may pay {1}{R/G}. When you do, Slinza fights target creature you don't
//	 control."
//
// The discount is generic mana only (Ballyrush Banneret's shape), priced as
// the spell is cast. The counter is Dragonstorm Globe's replacement on the
// entry of another creature its controller controls (CR 614.1c), so a
// changeling Beast, and a creature put onto the battlefield rather than cast,
// both get it. The trigger fires for Slinza and any creature you control
// whose power as it enters is 4 or greater (counters included, as the
// replacement above has already been applied); the payment is a real "you may
// pay" and the fight is a reflexive trigger (CR 603.12), its target chosen
// as it goes on the stack, so declining or being unable to pay makes none.
// If Slinza has left by then there is no fight.
//
// No simplification.
func init() {
	const label = "Beast spells you cast cost {2} less to cast."
	effects.Register(effects.Spec{
		OracleID:     "8bed095d-1538-4c80-b147-3ec308a8acb1",
		Name:         "Slinza, the Spiked Stampede",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, label, effects.YourSpell(), effects.SpellOfCreatureType("Beast")),
		},
		Replacements: []game.ReplacementEffect{
			effects.B28PermanentsYouControlEnterWithACounter("Slinza: another Beast enters with an additional +1/+1 counter",
				func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Beast") }),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, slinzaOrPowerFourEntered,
				"Slinza — you may pay {1}{R/G}; when you do, Slinza fights target creature you don't control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{R/G}",
						Question: "Slinza, the Spiked Stampede — pay {1}{R/G} to have Slinza fight target creature you don't control?",
						Then:     slinzaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// slinzaOrPowerFourEntered is "Slinza or another creature with power 4 or
// greater enters": the source itself always counts, any other creature must be
// one its controller controls, read as it enters.
func slinzaOrPowerFourEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.CardID == source.InstanceID || c.CurrentPower() >= 4
}

// slinzaPaid is the paid half, keyed: the reflexive fight trigger is queued
// against the asking item.
var slinzaPaid = effects.PayUnlessBranches("pay-unless/bns-b52-slinza-fight",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Slinza — fights target creature you don't control",
			Cards: []uuid.UUID{ctx.Item.SourceCardID},
			Body:  slinzaFightBody,
		}.Apply(ctx)
	}, nil)

var slinzaFightBody = game.ReflexiveBody("bns-b52/slinza-fight",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		fighters := ctx.PayloadCards()
		if len(fighters) == 0 {
			return nil
		}
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.B10Fight(ctx, fighters[0], id)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you don't control", effects.OpponentControls())
	}))
