package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// X-23, Deadly Weapon — Legendary Creature — Mutant Berserker Hero {2}{G},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever another Mutant you control enters, put a +1/+1 counter on
//	 that creature and a +1/+1 counter on X-23."
//
// Captain America, Team Leader's enters trigger (enteredUnderYourControl,
// another) narrowed to Mutant ("Mutant" a word of the card's text, CR 612.2).
// The entering creature takes the first counter and X-23 the second; a
// Mutant that left before the trigger resolves gets nothing (CR 608.2b)
// while X-23 still grows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b4dd121-b477-42ff-9dd1-5d0f957a509d",
		Name:         "X-23, Deadly Weapon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && effects.CardHasSubtype(c, source.TextSubtype("Mutant"))
			}, "X-23 — a +1/+1 counter on that Mutant and a +1/+1 counter on X-23",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.PlusOneCountersOnThis(1)(g, item)
				}),
		},
	})
}
