package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Umara Mystic — Creature — Merfolk Wizard {1}{U}{R}, 1/3 (Zendikar Rising,
// #5508):
//
//	"Flying
//	 Whenever you cast an instant, sorcery, or Wizard spell, this creature
//	 gets +2/+0 until end of turn."
//
// One cast trigger over instant, sorcery and Wizard spells (a Wizard
// creature spell included), each cast its own trigger, and the Mystic's own
// cast does not trigger it (it is not on the battlefield yet).
//
// No simplification.
func init() {
	const label = "Umara Mystic — +2/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "e1c33cfe-1d48-463d-83bb-111374505194",
		Name:            "Umara Mystic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.Subtype("Wizard")), label,
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: label}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
