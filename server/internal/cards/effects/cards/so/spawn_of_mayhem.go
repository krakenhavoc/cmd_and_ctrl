package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawn of Mayhem — Creature — Demon {2}{B}{B}, 4/4 (Ravnica: Clue Edition):
//
//	"Spectacle {1}{B}{B} (You may cast this spell for its spectacle cost
//	 rather than its mana cost if an opponent lost life this turn.)
//	 Flying, trample
//	 At the beginning of your upkeep, this creature deals 1 damage to each
//	 player. Then if you have 10 or less life, put a +1/+1 counter on this
//	 creature."
//
// Spectacle is the shared alternative cost (Drill Bit). The upkeep
// trigger is one damage event to every living player (Lim-Dûl's Hex's
// DealDamageEachKeyedForEffect), then a keyed tail that reads the
// controller's life after the damage and puts the counter on this
// creature only while it is still on the battlefield (CounterOnThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6a290adf-7b75-451e-bb79-9c13e10b38aa",
		Name:             "Spawn of Mayhem",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{1}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Spawn of Mayhem — 1 damage to each player, then a +1/+1 counter if you have 10 or less life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.DealDamageEachKeyedForEffect(ctx.Source(), effects.ApnapLivingSeats(g), 1, ctx.Tail(spawnOfMayhemAfterDamage))
				}),
		},
	})
}

// spawnOfMayhemAfterDamage is "then if you have 10 or less life, put a
// +1/+1 counter on this creature". An on-disk key: never renamed.
var spawnOfMayhemAfterDamage = effects.CardTail("then/spawn-of-mayhem-counter", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Life > 10 {
		return nil
	}
	return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
