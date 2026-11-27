package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Common Cause — Enchantment {2}{W} (Mercadian Masques, S82 OLD, #5619):
//
//	"Nonartifact creatures get +2/+2 as long as they all share a color."
//
// A layer-7c modification over every nonartifact creature on the battlefield,
// whoever controls it, applying only while a single color is shared by all of
// them (a creature of two colors counts for each, and one with no color shares
// none, so a colorless nonartifact creature switches it off). Artifact creatures
// neither count nor get the bonus. With no nonartifact creature there is nothing
// to give it to. It is re-read on every pass, so a creature that enters or dies
// turns the bonus on or off at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1dac139-de12-46ba-94b1-42e107c2c95f",
		Name:         "Common Cause",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Nonartifact creatures get +2/+2 as long as they all share a color.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return commonCauseNonartifactCreature(target) && commonCauseAllShareAColor(g)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}

func commonCauseNonartifactCreature(c *game.Card) bool { return c.IsCreature() && !c.IsArtifact() }

// commonCauseAllShareAColor reports whether one color is shared by every
// nonartifact creature on the battlefield.
func commonCauseAllShareAColor(g *game.Game) bool {
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		all, any := true, false
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if !commonCauseNonartifactCreature(c) {
				continue
			}
			any = true
			if !c.HasColor(color) {
				all = false
				break
			}
		}
		if any && all {
			return true
		}
	}
	return false
}
