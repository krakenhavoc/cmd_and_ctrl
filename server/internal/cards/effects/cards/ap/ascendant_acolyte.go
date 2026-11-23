package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ascendant Acolyte — Creature — Human Monk {4}{G}, 1/1 (Kamigawa: Neon
// Dynasty):
//
//	"This creature enters with a +1/+1 counter on it for each +1/+1 counter
//	 among other creatures you control.
//	 At the beginning of your upkeep, double the number of +1/+1 counters on
//	 this creature."
//
// The first ability is a CR 614.1c self-replacement (Callous Sell-Sword's
// b19EntersWithCountersCounted) that adds up the +1/+1 counters on every
// other creature its controller controls as it enters, so it counts what is
// there when it arrives, not when it was cast. The upkeep trigger puts as
// many more +1/+1 counters on it as it has as the ability resolves; a
// counter doubler on the board applies to that placement as it does to any
// other (CR 614.1). Only +1/+1 counters are doubled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b9918c3-1135-4c63-991f-f0a38340a33b",
		Name:         "Ascendant Acolyte",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID != src.InstanceID && c.Controller == src.Controller && c.IsCreature() {
						n += c.Counters[game.CounterPlusOne]
					}
				}
				return n
			}, "Ascendant Acolyte enters with a +1/+1 counter on it for each +1/+1 counter among other creatures you control"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ascendant Acolyte — double the number of +1/+1 counters on this creature",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					ctx := effects.NewContext(g, item)
					if !ok || ctx.IsNewSourceObject(item.SourceCardID) || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					n := c.Counters[game.CounterPlusOne]
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}
