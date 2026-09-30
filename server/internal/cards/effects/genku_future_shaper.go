package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// genkuFutureShaperLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen this turn" memory is kept under.
const genkuFutureShaperLabel = "Genku, Future Shaper — a nontoken permanent left"

// Genku, Future Shaper — Legendary Creature — Moonfolk Wizard
// {2}{W}{U}, 2/5:
//
//	"Whenever another nontoken permanent you control leaves the
//	 battlefield, choose one that hasn't been chosen this turn. Create a
//	 creature token with those characteristics.
//	 • 2/2 white Fox with vigilance.
//	 • 1/2 blue Moonfolk with flying.
//	 • 1/1 black Rat with lifelink.
//	 {3}{W}{U}: Put a +1/+1 counter on each creature you control."
//
// Every exit counts — dying, a bounce, an exile, a sacrifice — and
// "you control" is the controller the permanent LEFT under (CR
// 603.10a), so a creature an opponent stole and sacrificed does not
// trigger its owner's Genku. A leaves-the-battlefield trigger looks
// back in time, so Genku sees the other permanents that leave with it
// in a wrath. ChooseOneNotChosenThisTurn (ADR 0097) gives the first
// three exits each turn three different tokens, and a fourth exit
// triggers and is removed with no effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca16b735-26c7-4d0b-959b-b031f6cb376e",
		Name:         "Genku, Future Shaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			genkuFutureShaperTrigger(),
		},
		Activated: []ActivatedAbility{{
			Label:  "{3}{W}{U}: Put a +1/+1 counter on each creature you control.",
			Cost:   ManaCost("{3}{W}{U}"),
			Effect: b11PutCounterOnEachCreatureYouControl,
		}},
	})
}

func genkuFutureShaperTrigger() game.TriggeredAbility {
	t := On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.CardID == source.InstanceID {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !IsToken(c) && leftUnderControlOf(ev, c) == source.Controller
	}, genkuFutureShaperLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("2/2 white Fox with vigilance.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 white Fox with vigilance"), N: 1}.Apply(ctx)
		}),
		ModeDoing("1/2 blue Moonfolk with flying.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/2 blue Moonfolk with flying"), N: 1}.Apply(ctx)
		}),
		ModeDoing("1/1 black Rat with lifelink.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 black Rat with lifelink"), N: 1}.Apply(ctx)
		}),
	)
	return t
}
