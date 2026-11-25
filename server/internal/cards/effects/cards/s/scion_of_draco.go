package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Draco — Artifact Creature — Dragon {12}, 4/4 (Secret Lair Drop):
//
//	"Domain — This spell costs {2} less to cast for each basic land type among
//	 lands you control.
//	 Flying
//	 Each creature you control has vigilance if it's white, hexproof if it's
//	 blue, lifelink if it's black, first strike if it's red, and trample if
//	 it's green."
//
// The discount is Yavimaya Sojourner's domain count (DomainCostAmount) at {2}
// for each type, on generic mana only (CR 601.2f). The last sentence is five
// layer 6 keyword grants, one per colour, to every creature its controller
// controls (the Scion included) that is that colour as the layers have it, so a
// multicoloured creature gets each of its colours' abilities. The colour words
// are read through the Scion's text (CR 612.2).
//
// No simplification.
func init() {
	grant := func(colour, keyword string) game.StaticAbility {
		ab := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.HasColor(source.TextColor(colour))
		}, keyword)
		ab.Label = "Each creature you control has " + keyword + " if it's " + map[string]string{"W": "white", "U": "blue", "B": "black", "R": "red", "G": "green"}[colour] + "."
		return ab
	}
	effects.Register(effects.Spec{
		OracleID:        "fd547632-a29a-4dad-a9da-155882391815",
		Name:            "Scion of Draco",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return 2 * effects.DomainCostAmount(q) },
				"Domain — This spell costs {2} less to cast for each basic land type among lands you control."),
		},
		Static: []game.StaticAbility{
			grant("W", "vigilance"), grant("U", "hexproof"), grant("B", "lifelink"), grant("R", "first strike"), grant("G", "trample"),
		},
	})
}
