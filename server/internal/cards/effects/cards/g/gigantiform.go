package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gigantiform — Enchantment — Aura {3}{G}{G} (Zendikar, S82 OLD batch 491,
// #5583):
//
//	"Kicker {4}
//	 Enchant creature
//	 Enchanted creature has base power and toughness 8/8 and has trample.
//	 When this Aura enters, if it was kicked, you may search your library for a
//	 card named Gigantiform, put it onto the battlefield, then shuffle."
//
// Base 8/8 in layer 7b and trample in layer 6 are the attached statics.
// The trigger's intervening "if" reads the Aura's own kicked record; the
// search is optional, finds a card named Gigantiform and puts it onto the
// battlefield (the engine asks what it enchants, CR 303.4f), and the library
// is shuffled. The Aura it fetches was not cast, so it is not kicked and
// fetches nothing further.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d9e1bae6-2d56-45d1-bdac-2431a18aab3e",
		Name:          "Gigantiform",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		Targets:       effects.EnchantCreature(),
		Static:        []game.StaticAbility{effects.SetAttachedBasePT(8, 8), effects.GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Gigantiform — kicked, you may search your library for a card named Gigantiform, put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.Name == "Gigantiform" },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Optional:  true,
						Shuffle:   true,
						Source:    ctx.Source(),
						Reason:    "Gigantiform — search your library for a card named Gigantiform",
					}.Apply(ctx)
				}),
		},
	})
}
