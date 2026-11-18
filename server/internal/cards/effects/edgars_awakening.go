package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Edgar's Awakening — Sorcery {3}{B}{B} (Innistrad: Crimson Vow):
//
//	"Return target creature card from your graveyard to the battlefield.
//	 When you discard this card, you may pay {B}. When you do, return
//	 target creature card from your graveyard to your hand."
//
// The discard trigger watches from the graveyard (CR 113.6, the cycling
// shape): "you" is the card's owner there (CR 108.4). Paying {B} makes a
// CR 603.12 reflexive trigger that picks its target only then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c04371a-af69-4e42-8cda-096cf8312b5e",
		Name:         "Edgar's Awakening",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return returnTargetCardToTheBattlefield(ctx.Item, ctx) }),
		),
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDiscardCard, Self,
				"Edgar's Awakening — you may pay {B}", func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{B}",
						Question: "Edgar's Awakening — pay {B} to return target creature card from your graveyard to your hand?",
						Then:     edgarsAwakeningPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

var edgarsAwakeningPaid = PayUnlessBranches("pay-unless/edgars-awakening-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Edgar's Awakening — return target creature card from your graveyard to your hand",
			edgarsAwakeningReturnBody).Apply(ctx)
	}, nil)

var edgarsAwakeningReturnBody = game.ReflexiveBody("edgars-awakening/return-to-hand",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())
	}))
