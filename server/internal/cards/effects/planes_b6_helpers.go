package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b6_helpers.go holds the shapes planes batch 6 (ADR 0216 slice 4,
// #4870) share. Append-only.

// b6ToughnessDamagePair is "Each of those creatures deals damage equal to
// its toughness to the other" over clause slots 0 and 1 of the resolving
// ability (The Great Aerie): both amounts are read before either deals, as
// the damage is simultaneous (CR 120.3), and a missing or illegal creature
// leaves the other dealing nothing to it and receiving nothing.
func b6ToughnessDamagePair(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids [2]uuid.UUID
	var amt [2]int
	var ok [2]bool
	for i := range ids {
		t, legal := ctx.ClauseTarget(i)
		if !legal || t.Kind != game.TargetCard {
			continue
		}
		c, found := g.LookupCardForEffect(t.ID)
		if !found {
			continue
		}
		ids[i], amt[i], ok[i] = t.ID, g.ToughnessAnywhereForEffect(c), true
	}
	for i := range ids {
		j := 1 - i
		if ok[i] && ok[j] {
			if err := (DealDamage{Source: ids[i], Target: ids[j], Amount: amt[i]}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
