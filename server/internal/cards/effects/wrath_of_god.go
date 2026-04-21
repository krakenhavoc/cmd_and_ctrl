package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrath of God — "Destroy all creatures. They can't be regenerated."
// Iterated DestroyTarget. S14 sandbox does not model regeneration
// (it's a replacement-effect concern that lands in S17 / S18); the
// "can't be regenerated" clause is cosmetic until then.
func init() {
	Register(Spec{
		OracleID:  "34515b16-c9a4-4f98-8c77-416a7a523407",
		Name:      "Wrath of God",
		OnResolve: wrathDestroyAllCreatures,
	})
}

// wrathDestroyAllCreatures is shared between Wrath of God, Damnation,
// and Day of Judgment — three cards with functionally identical
// effects. Extracted as a package-local func so the per-card files
// stay thin.
func wrathDestroyAllCreatures(_ *game.StackItem, ctx *Context) error {
	for _, id := range ctx.CreatureIDs() {
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
