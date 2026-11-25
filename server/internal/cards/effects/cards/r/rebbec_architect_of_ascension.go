package cardsr

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rebbec, Architect of Ascension — {3}{W} Legendary Creature — Human
// Artificer 3/4 (Commander Legends, S82 OLD b077, #6149):
//
//	"Artifacts you control have protection from each mana value among
//	 artifacts you control.
//	 Partner (You can have two commanders if both have partner.)"
//
// CR 702.16i: "protection from each mana value among artifacts you
// control" is a separate "protection from mana value N" (#6149's exact
// quality) for each N, so the layer-6 static grants one per mana value
// among the artifacts Rebbec's controller controls, recomputed whenever
// the layers are: an artifact with mana value 3 can't be targeted by a
// mana value 3 source while one of yours has that value, and a 2 can.
// Partner is a deck-construction rule (CR 702.124a) and needs nothing
// on the card. The body and cost are printed card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccad9b56-3c31-445f-8b78-b5bbdb3d077a",
		Name:         "Rebbec, Architect of Ascension",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Artifacts you control have protection from each mana value among artifacts you control.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsArtifact() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				for _, mv := range rebbecManaValues(g, source) {
					if tok := game.ProtectionFromManaValue(mv); tok != "" {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
					}
				}
			},
		}},
	})
}

// rebbecManaValues is each mana value among the artifacts the source's
// controller controls, ascending, once each.
func rebbecManaValues(g *game.Game, source *game.Card) []int {
	if g == nil || g.Battlefield == nil {
		return nil
	}
	seen := map[int]bool{}
	var out []int
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsArtifact() || c.Controller != source.Controller {
			continue
		}
		mv := c.ManaValue()
		if !seen[mv] {
			seen[mv] = true
			out = append(out, mv)
		}
	}
	sort.Ints(out)
	return out
}
