package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrath of God — "Destroy all creatures. They can't be regenerated."
//
// S23: rewritten onto DestroyAllMatching. The behavioural change is
// not the destruction — it is that the creatures now die as ONE
// event, so a Blood Artist caught in the wrath drains for every
// creature that died with it instead of for however many happened to
// be processed after it. See mass.go.
//
// "They can't be regenerated" stays cosmetic: regeneration is not
// modelled (game/keywords.go's canonical set is closed and does not
// contain it), so nothing could have been regenerated anyway. The
// clause becomes load-bearing when regeneration lands, and
// DestroyAllMatching is where it will be honoured.
func init() {
	Register(Spec{
		OracleID:  "34515b16-c9a4-4f98-8c77-416a7a523407",
		Name:      "Wrath of God",
		OnResolve: wrathDestroyAllCreatures,
	})
}

// wrathDestroyAllCreatures is shared between Wrath of God, Damnation,
// Day of Judgment and Supreme Verdict — four cards whose whole text,
// once the uncounterable and can't-be-regenerated riders are
// accounted for, is the same effect. Extracted as a package-local
// func so the per-card files stay thin.
func wrathDestroyAllCreatures(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{Match: Creature()}.Apply(ctx)
}
