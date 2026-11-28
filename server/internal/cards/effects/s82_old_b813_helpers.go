package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b813_helpers.go — shared pieces of the S82 OLD batch 813 cards
// (Mirage, #5633).

// OldB813DoesntUntapWhenThisAttacks is "Whenever this creature attacks, it
// doesn't untap during its controller's next untap step" (Lead Golem): the
// skip is recorded on the attacker as the object it is (CR 502.3, CR 400.7).
func OldB813DoesntUntapWhenThisAttacks(name string) game.TriggeredAbility {
	return WheneverThisAttacks(name+" — it doesn't untap during its controller's next untap step", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if ctx.SourceLeftBattlefield(item.SourceCardID) {
			return nil
		}
		return DoesntUntapNextUntapStep{Targets: []uuid.UUID{item.SourceCardID}, Label: name}.Apply(ctx)
	})
}

// OldB813TokenNamedWood matches a token named Wood (Jungle Patrol's cost).
func OldB813TokenNamedWood() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return IsToken(c) && c.Name == "Wood"
	}
}
