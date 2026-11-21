package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invisible Woman — {2}{W} Legendary Creature — Human Hero 3/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, create a 0/3 colorless Wall creature token with
//	 defender and reach.
//	 Whenever you attack, you may pay {R}{G}{W}{U}. When you do, target
//	 creature gets +1/+0 until end of turn for each creature you control and
//	 can't be blocked this turn."
//
// The Wall trigger has Franklin Richards' intervening if (CR 603.4), checked
// as combat begins and as the ability resolves. "Whenever you attack" fires
// once per declaration; paying the {R}{G}{W}{U} raises the CR 603.12
// reflexive trigger, which picks its target as it goes on the stack and
// counts your creatures as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "085016ee-4429-453f-b7f3-d61065f1214e",
		Name:         "Invisible Woman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, true), effects.FranklinCastNoncreature),
				"Invisible Woman — create a 0/3 colorless Wall creature token with defender and reach",
				func(g *game.Game, item *game.StackItem) error {
					if g.CastTallyFor(item.Controller).Noncreature == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("0/3 colorless Wall with defender and reach"), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Invisible Woman — you may pay {R}{G}{W}{U}", msc13InvisibleWomanOffer)),
		},
	})
}

func msc13InvisibleWomanOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{R}{G}{W}{U}",
		Question: "Invisible Woman — pay {R}{G}{W}{U}? When you do, target creature gets +1/+0 for each creature you control and can't be blocked",
		Then:     msc13InvisibleWomanPaid,
	}.Apply(effects.NewContext(g, item))
}

var msc13InvisibleWomanPaid = effects.PayUnlessBranches("pay-unless/invisible-woman-pump",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Invisible Woman — target creature gets +1/+0 for each creature you control and can't be blocked this turn",
			Body:  msc13InvisibleWomanBody,
		}.Apply(ctx)
	}, nil)

var msc13InvisibleWomanBody = game.ReflexiveBody("invisible-woman/pump", effects.SimpleBody(msc13InvisibleWomanPump),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))

func msc13InvisibleWomanPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BoostUntilEOT{
			Target: t.ID, Power: effects.B04CreaturesControlled(g, item.Controller),
			Label: "Invisible Woman — +1/+0 for each creature you control",
		}).Apply(ctx); err != nil {
			return err
		}
		return effects.RestrictUntilEOT{
			Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Invisible Woman — can't be blocked",
		}.Apply(ctx)
	}
	return nil
}
