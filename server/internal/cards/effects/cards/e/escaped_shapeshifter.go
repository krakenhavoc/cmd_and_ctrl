package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escaped Shapeshifter — Creature — Shapeshifter {3}{U}{U}, 3/4 (Tempest, S82
// OLD, #5628):
//
//	"As long as an opponent controls a creature with flying not named Escaped
//	 Shapeshifter, this creature has flying. The same is true for first strike,
//	 trample, and protection from any color."
//
// One layer-6 static on the Shapeshifter itself. For each of flying, first
// strike, trample and protection from each colour, it has the ability while an
// opponent controls a creature not named Escaped Shapeshifter that has it, read
// through the post-layer keywords. Protection is per colour, as the 2004 ruling
// has it: an opponent's creature with protection from red gives it protection
// from red only. Two Shapeshifters never feed each other, by the name clause.
//
// No simplification.
func init() {
	const name = "Escaped Shapeshifter"
	effects.Register(effects.Spec{
		OracleID:     "17f2342b-4dd8-445a-a7c1-4a15a5688294",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "As long as an opponent controls a creature with flying not named " + name + ", this creature has flying. The same is true for first strike, trample, and protection from any color.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target != nil && source != nil && target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				for _, kw := range effects.S82OldB762ShapeshifterKeywords(g, source, name) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
				}
			},
		}},
	})
}
