package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halcyon Glaze — Enchantment {1}{U}{U} (Ravnica: City of Guilds):
//
//	"Whenever you cast a creature spell, this enchantment becomes a 4/4 Illusion
//	 creature with flying in addition to its other types until end of turn."
//
// Riddleform's animation without the "may": the Glaze adds the Creature type and
// the Illusion subtype, is 4/4 (layer 7b) and has flying, all until end of turn
// and pinned to this object (CR 400.7). It stays an enchantment. A Glaze that
// has been a creature since an earlier cast this turn simply is one again, with
// a newer timestamp.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36a247d6-0d40-4f61-935c-60f42aa98562",
		Name:         "Halcyon Glaze",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Creature()),
				"Halcyon Glaze — becomes a 4/4 Illusion creature with flying in addition to its other types until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BecomeCreatureUntilEOT{Types: []string{"Creature"}, Subtypes: []string{"Illusion"}, SetPower: 4, SetToughness: 4,
						Label: "Halcyon Glaze — becomes a 4/4 Illusion creature"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Halcyon Glaze — flying"}.Apply(ctx)
				}),
		},
	})
}
