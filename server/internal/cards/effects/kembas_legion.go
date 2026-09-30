package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kemba's Legion — Creature — Cat Soldier {5}{W}{W}, 4/6:
//
//	"Vigilance
//	 This creature can block an additional creature each combat for
//	 each Equipment attached to this creature."
//
// Vigilance rides PrintedKeywords. The block count is NOT the fixed-N
// CanBlockAdditional every other #1706 card uses — it is counted live
// off the attachment relation on every recompute, the same read Uril,
// the Miststalker's Aura count uses (uril_the_miststalker.go), over
// Equipment instead of Auras: `game.AttachmentsOf`'s reverse direction
// of `Card.AttachedTo`, safe from inside Apply because the layer
// recompute already holds the lock this reads under.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "338ad8c4-c89f-4e78-be28-7224fac0fd0f",
		Name:            "Kemba's Legion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if g == nil {
					return
				}
				n := 0
				for _, att := range g.BattlefieldCardsForEffect() {
					if att.IsAttachedTo(source.InstanceID) && att.HasSubtype("Equipment") {
						n++
					}
				}
				c.AdditionalBlocks += n
			},
		}},
	})
}
