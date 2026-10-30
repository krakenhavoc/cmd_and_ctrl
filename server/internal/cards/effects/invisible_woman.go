package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "085016ee-4429-453f-b7f3-d61065f1214e",
		Name:         "Invisible Woman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), franklinCastNoncreature),
				"Invisible Woman — create a 0/3 colorless Wall creature token with defender and reach",
				func(g *game.Game, item *game.StackItem) error {
					if g.CastTallyFor(item.Controller).Noncreature == 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("0/3 colorless Wall with defender and reach"), N: 1}.Apply(NewContext(g, item))
				}),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Invisible Woman — you may pay {R}{G}{W}{U}", msc13InvisibleWomanOffer)),
		},
	})
}

func msc13InvisibleWomanOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{R}{G}{W}{U}",
		Question: "Invisible Woman — pay {R}{G}{W}{U}? When you do, target creature gets +1/+0 for each creature you control and can't be blocked",
		Then:     msc13InvisibleWomanPaid,
	}.Apply(NewContext(g, item))
}

var msc13InvisibleWomanPaid = PayUnlessBranches("pay-unless/invisible-woman-pump",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Invisible Woman — target creature gets +1/+0 for each creature you control and can't be blocked this turn",
			Body:  msc13InvisibleWomanBody,
		}.Apply(ctx)
	}, nil)

var msc13InvisibleWomanBody = game.ReflexiveBody("invisible-woman/pump", simpleBody(msc13InvisibleWomanPump),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

func msc13InvisibleWomanPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{
			Target: t.ID, Power: b04CreaturesControlled(g, item.Controller),
			Label: "Invisible Woman — +1/+0 for each creature you control",
		}).Apply(ctx); err != nil {
			return err
		}
		return RestrictUntilEOT{
			Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Invisible Woman — can't be blocked",
		}.Apply(ctx)
	}
	return nil
}
