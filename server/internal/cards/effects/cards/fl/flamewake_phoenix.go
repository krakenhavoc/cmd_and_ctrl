package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flamewake Phoenix — Creature — Phoenix {1}{R}{R}, 2/2 (Foundations,
// #4465):
//
//	"Flying, haste
//	 This creature attacks each combat if able.
//	 Ferocious — At the beginning of combat on your turn, if you control
//	 a creature with power 4 or greater, you may pay {R}. If you do,
//	 return this card from your graveyard to the battlefield."
//
// Flying and haste are engine keywords; "attacks each combat if able" is
// Zurgo Helmsmasher's requirement (AttacksEachCombat, CR 508.1d), which
// an attack cost or a tapped Phoenix does not force. The ferocious
// ability watches from the GRAVEYARD only (Silversmote Ghoul's
// shape), at the beginning of its owner's combat. Its intervening "if"
// (CR 603.4) is checked as the combat begins and again as the ability
// resolves, off the current power of the controller's creatures, so
// losing the creature in response removes it before the payment is asked.
// "You may pay {R}" is MayPay: the return runs only when the mana was
// really paid, and a card that has left the graveyard by then (CR 400.7)
// stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e43805d-3576-47d4-9331-524869ab99d0",
		Name:            "Flamewake Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.InterveningIf(effects.CondPowerAtLeast{N: 4},
				effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, true), effects.StdYouControlAPower4Creature),
					"Flamewake Phoenix — you may pay {R} to return it from your graveyard to the battlefield",
					flamewakePhoenixAsk))),
		},
	})
}

// flamewakePhoenixAsk asks whether to pay {R}, carrying the Phoenix as
// the object it is in the graveyard now.
func flamewakePhoenixAsk(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.MayPay{
		Chooser:      item.Controller,
		Cost:         "{R}",
		Question:     "Pay {R} to return Flamewake Phoenix from your graveyard to the battlefield?",
		Then:         flamewakePhoenixPaid,
		CarryObjects: []game.ObjectRef{{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(effects.NewContext(g, item))
}

// flamewakePhoenixPaid is the "If you do": the Phoenix, if it is still the
// card that was asked about and still in a graveyard, returns under its
// owner's control. An on-disk key: never renamed, never reused.
var flamewakePhoenixPaid = effects.PayUnlessBranches("pay-unless/flamewake-phoenix-return",
	func(ctx *effects.Context, ans game.PayUnlessAnswer) error {
		if len(ans.Objects) == 0 {
			return nil
		}
		o := ans.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(o.ID)
		if !ok || c.ObjectEpoch != o.Epoch {
			return nil
		}
		if z := ctx.Game.FindCardZoneForEffect(o.ID); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: o.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
	}, nil)
