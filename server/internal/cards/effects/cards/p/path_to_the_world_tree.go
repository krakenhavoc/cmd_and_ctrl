package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Path to the World Tree — Enchantment {1}{G}:
//
//	"When this enchantment enters, search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 {2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw
//	 two cards. Target opponent loses 2 life. This enchantment deals 2 damage
//	 to up to one target creature. You create a 2/2 green Bear creature
//	 token."
//
// The ability has two target clauses, in printed order: the opponent (slot 0)
// and up to one creature (slot 1). The enchantment is sacrificed as a cost, so
// the damage is dealt by it as it last existed (CR 608.2h). The instructions
// run in printed order, each after the one before has finished (CR 608.2c): a
// target that became illegal is skipped (CR 608.2b) and the rest still happen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a933f443-dd53-4805-8446-72693a27cb56",
		Name:         "Path to the World Tree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Path to the World Tree — search your library for a basic land card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Path to the World Tree — a basic land card",
						Source:    item.SourceCardID,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw two cards. Target opponent loses 2 life. This enchantment deals 2 damage to up to one target creature. You create a 2/2 green Bear creature token.",
			Cost:  effects.Plus(effects.ManaCost("{2}{W}{U}{B}{R}{G}"), effects.SacrificeThis()),
			Targets: effects.Clauses(
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.UpToOneTargetCreature("up to one target creature"),
			),
			Effect: effects.Do(
				effects.GainLife{Amount: 2},
				effects.DrawCards{N: 2},
				effects.Instruction(func(ctx *effects.Context) error {
					if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
					}
					return nil
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if t, ok := ctx.ClauseTarget(1); ok {
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
					}
					return nil
				}),
				effects.CreateToken{Template: effects.TokenCard("2/2 green Bear"), N: 1},
			),
		}},
	})
}
