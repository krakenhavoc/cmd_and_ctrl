package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tormented Hero — Creature — Human Warrior {B}, 2/1 (Theros, S82 OLD b388):
//
//	"This creature enters tapped.
//	 Heroic — Whenever you cast a spell that targets this creature, each
//	 opponent loses 1 life. You gain life equal to the life lost this way."
//
// The entry is Rotting Legion's. The heroic trigger takes 1 from each opponent
// (life loss, CR 119.3) and the caster gains the life that really went: an
// opponent whose loss was prevented or replaced adds nothing. With no loss
// nothing is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47fbb4f0-565b-4682-9971-0735210c76e5",
		Name:         "Tormented Hero",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Tormented Hero — each opponent loses 1 life and you gain life equal to the life lost this way (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.LoseLifeEachKeyedForEffect(ctx.Source(), ctx.Opponents(), 1, game.KeyedTail{
						Ref: tormentedHeroGain, Player: item.Controller, Source: item.SourceCardID,
					})
				}),
		},
	})
}

// tormentedHeroGain is "you gain life equal to the life lost this way", told the
// total that really went once every loss has settled. An on-disk key: never
// renamed.
var tormentedHeroGain = game.RegisterTail("tormented-hero/gain-life-lost", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 {
		return nil
	}
	return effects.GainLife{Player: o.Tail.Player, Amount: o.Amount}.Apply(
		effects.NewContext(g, &game.StackItem{Controller: o.Tail.Player, SourceCardID: o.Tail.Source}))
})
