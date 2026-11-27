package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulcatchers' Aerie — Enchantment {1}{W} (Judgment):
//
//	"Whenever a Bird is put into your graveyard from the battlefield, put a
//	 feather counter on this enchantment.
//	 Bird creatures get +1/+1 for each feather counter on this enchantment."
//
// Mazzy's put-into-your-graveyard reading (the card as it last existed was a
// Bird, and its OWNER is the Aerie's controller, so a Bird of yours that an
// opponent stole counts, and one of theirs that you stole does not) feeding a
// feather counter. The anthem is a layer-7c modification of every Bird
// creature, whoever controls it, read live from the counters on the Aerie.
//
// No simplification.
func init() {
	const (
		name    = "Soulcatchers' Aerie"
		feather = "feather"
	)
	effects.Register(effects.Spec{
		OracleID:     "69565f6e-2e23-4cd8-bd21-1283e06729a3",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Bird creatures get +1/+1 for each feather counter on this enchantment.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasSubtype(source.TextSubtype("Bird"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters[feather]
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsSubtype(ev, c, source.TextSubtype("Bird")) && c.Owner == source.Controller
			}, name+" — put a feather counter on this enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: feather, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
