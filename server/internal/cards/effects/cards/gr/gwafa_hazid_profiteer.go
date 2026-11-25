package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gwafa Hazid, Profiteer — Legendary Creature — Human Rogue {1}{W}{U}, 2/2 (Battlebond, #5529):
//
//	{W}{U}, {T}: Put a bribery counter on target creature you don't control. Its controller draws a card.
//	Creatures with bribery counters on them can't attack or block.
//
// The counter is a plain named counter (like a quest counter); the static restricts every creature carrying one, whoever controls it, for as long as Gwafa is on the battlefield. "Its controller" is read as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ba31f45-c178-4bac-ab7b-5ba4d050133c",
		Name:         "Gwafa Hazid, Profiteer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Creatures with bribery counters on them can't attack or block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature() && target.Counters["bribery"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}, {T}: Put a bribery counter on target creature you don't control. Its controller draws a card.",
			Cost:    effects.Plus(effects.ManaCost("{W}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, found := g.LookupCardForEffect(id)
				if !found {
					return nil
				}
				owner := c.Controller
				if err := (effects.AddCounter{Target: id, Kind: "bribery", N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.DrawCards{Player: owner, N: 1}.Apply(ctx)
			},
		}},
	})
}
