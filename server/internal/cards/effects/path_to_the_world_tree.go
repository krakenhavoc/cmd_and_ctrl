package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a933f443-dd53-4805-8446-72693a27cb56",
		Name:         "Path to the World Tree",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Path to the World Tree — search your library for a basic land card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Path to the World Tree — a basic land card",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}{U}{B}{R}{G}, Sacrifice this enchantment: You gain 2 life and draw two cards. Target opponent loses 2 life. This enchantment deals 2 damage to up to one target creature. You create a 2/2 green Bear creature token.",
			Cost:  Plus(ManaCost("{2}{W}{U}{B}{R}{G}"), SacrificeThis()),
			Targets: Clauses(
				TargetPlayer("target opponent", Opponent()),
				UpToOneTargetCreature("up to one target creature"),
			),
			Effect: Do(
				GainLife{Amount: 2},
				DrawCards{N: 2},
				Instruction(func(ctx *Context) error {
					if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
					}
					return nil
				}),
				Instruction(func(ctx *Context) error {
					if t, ok := ctx.ClauseTarget(1); ok {
						return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
					}
					return nil
				}),
				CreateToken{Template: TokenCard("2/2 green Bear"), N: 1},
			),
		}},
	})
}
