package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// galadrielLightOfValinorLabel is the trigger's stack label, and with
// it the key its "hasn't been chosen this turn" memory is kept under.
const galadrielLightOfValinorLabel = "Galadriel, Light of Valinor — alliance"

// Galadriel, Light of Valinor — Legendary Creature — Elf Noble
// {2}{G}{W}{U}, 3/3:
//
//	"Alliance — Whenever another creature you control enters, choose
//	 one that hasn't been chosen this turn —
//	 • Add {G}{G}{G}.
//	 • Put a +1/+1 counter on each creature you control.
//	 • Scry 2, then draw a card."
//
// Gala Greeters' shape at five mana: the alliance condition and
// ChooseOneNotChosenThisTurn (ADR 0097), so the first three creatures
// that enter each turn pay three different bullets and a fourth
// triggers and is removed with no effect.
//
// "Add {G}{G}{G}" is mana from a triggered ability that is NOT a mana
// ability — it triggers on an entry, not on mana being produced (CR
// 605.1b) — so it uses the stack and the mana lands as the trigger
// resolves, emptying with the pool at the end of the step (CR 106.4).
// The counters are one placement per creature, the set read as the
// bullet resolves. "Scry 2, then draw" draws in the scry's
// continuation, so the card drawn is the one the scry left on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "faf44683-339d-4fc0-8029-1e867c69de0e",
		Name:         "Galadriel, Light of Valinor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			galadrielLightOfValinorTrigger(),
		},
	})
}

func galadrielLightOfValinorTrigger() game.TriggeredAbility {
	t := WheneverAnotherCreatureEntersUnderYourControl(galadrielLightOfValinorLabel,
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("Add {G}{G}{G}.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddMana{Player: item.Controller, Produced: "{G}{G}{G}"}.Apply(ctx)
			}),
		ModeDoing("Put a +1/+1 counter on each creature you control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return b11PutCounterOnEachCreatureYouControl(ctx.Game, item)
			}),
		ModeDoing("Scry 2, then draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				you := item.Controller
				return Scry{Player: you, N: 2, Then: func(g *game.Game) error {
					return g.DrawNForEffect(you, 1)
				}}.Apply(ctx)
			}),
	)
	return t
}
