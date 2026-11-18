package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bess, Soul Nourisher — Legendary Creature — Human Citizen {1}{G}{W}, 1/1
// (Streets of New Capenna Commander):
//
//	"Whenever one or more other creatures you control with base power and
//	 toughness 1/1 enter, put a +1/+1 counter on Bess.
//	 Whenever Bess attacks, each other creature you control with base power
//	 and toughness 1/1 gets +X/+X until end of turn, where X is the number
//	 of +1/+1 counters on Bess."
//
// Base power and toughness is the value after layer 7b (CR 208.4b), read as
// the creature enters. Creatures entering together are one trigger
// (OncePerBatch), so Bess grows once however many arrive. The attack pump
// fixes who is affected as the trigger resolves (CR 611.2c) and reads X
// then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df72b5d6-6f9e-4d6b-acf6-7dec4ff35468",
		Name:         "Bess, Soul Nourisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && baseOneByOne(g, source.Controller, c)
			}, "Bess, Soul Nourisher — put a +1/+1 counter on Bess",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
			WheneverThisAttacks("Bess, Soul Nourisher — each other creature you control with base power and toughness 1/1 gets +X/+X until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					src, ok := g.LookupCardForEffect(ctx.Source())
					if !ok {
						return nil
					}
					x := src.Counters[game.CounterPlusOne]
					if x <= 0 {
						return nil
					}
					return BoostUntilEOT{
						Match:     And(YouControl(), NotSelf(src.InstanceID), baseOneByOne),
						Power:     x,
						Toughness: x,
						Label:     "Bess, Soul Nourisher — +X/+X until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
