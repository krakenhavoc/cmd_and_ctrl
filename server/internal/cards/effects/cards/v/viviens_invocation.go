package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien's Invocation — Sorcery {5}{G}{G} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Look at the top seven cards of your library. You may put a creature card
//	 from among them onto the battlefield. Put the rest on the bottom of your
//	 library in a random order. When a creature is put onto the battlefield
//	 this way, it deals damage equal to its power to target creature an
//	 opponent controls."
//
// Armored Skyhunter's look-and-put (PutFromLibraryOntoBattlefield) with the
// bottoming as its keyed tail, then a CR 603.12 reflexive trigger created only
// when a creature really entered, as Back for More's fight is. The trigger's
// target is chosen as it goes on the stack; the creature rides the payload and
// deals damage equal to its power, last-known if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e2b30d2-81e9-40cd-86dd-3ec5dd06ced0",
		Name:         "Vivien's Invocation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			controller := ctx.Controller()
			return effects.PutFromLibraryOntoBattlefield{
				Player:   controller,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(controller, 7),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Vivien's Invocation — you may put a creature card from among them onto the battlefield",
				ThenTail: viviensInvocationRest,
			}.Apply(ctx)
		},
	})
}

// viviensInvocationRest is the rest to the bottom in a random order, then the
// reflexive damage when a creature entered. An on-disk key.
var viviensInvocationRest = effects.PutFromLibraryTail("viviens-invocation/rest-then-damage",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		if err := effects.PutRestOnBottomInRandomOrder(ctx.Game, res); err != nil {
			return err
		}
		if len(res.Entered) == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label: "Vivien's Invocation — the creature deals damage equal to its power to target creature an opponent controls",
			Cards: res.Entered[:1],
			Body:  viviensInvocationDamageBody,
		}.Apply(ctx)
	})

var viviensInvocationDamageBody = game.ReflexiveBody("s82-old-b198/viviens-invocation-damage",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		entered := ctx.PayloadCards()
		if len(entered) == 0 {
			return nil
		}
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		power := effects.B17LastKnownPowerOffBattlefield(g, entered[0])
		if effects.OnBattlefield(g, entered[0]) {
			if c, ok := g.LookupCardForEffect(entered[0]); ok {
				power = c.CurrentPower()
			}
		}
		if power <= 0 {
			return nil
		}
		return effects.DealDamage{Source: entered[0], Target: target, Amount: power}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
