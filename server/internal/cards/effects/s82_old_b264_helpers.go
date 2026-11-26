package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b264_helpers.go — shared bodies for Kaladesh batch 264 (#5539).

// OldB264DamageTargetPlayerOrPlaneswalkerAndTheirCreatures is "<source> deals
// N damage to target player or planeswalker and each creature that player or
// that planeswalker's controller controls" (Chandra, Pyrogenius): the player is
// the target, or the controller of the targeted planeswalker, read as the
// ability resolves and before any damage lands. Every point is one simultaneous
// damage instance (CR 120.2); with its one target illegal it does nothing
// (CR 608.2b).
func OldB264DamageTargetPlayerOrPlaneswalkerAndTheirCreatures(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok || n <= 0 {
		return nil
	}
	player := t.ID
	if t.Kind != game.TargetPlayer {
		c, found := g.LookupCardForEffect(t.ID)
		if !found {
			return nil
		}
		player = c.Controller
	}
	ids := []uuid.UUID{t.ID}
	for _, c := range MatchingBattlefield(ctx, And(Creature(), ControlledBy(player))) {
		if c.InstanceID != t.ID {
			ids = append(ids, c.InstanceID)
		}
	}
	return DealDamageToEachThen(ctx, ids, n, RecipientThen{})
}
