package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b63_helpers.go — shared bodies for the Lorwyn Eclipsed leftovers,
// batches 63 to 66 (#4458).

// stdB63GreatestElementalManaValue is Sunderflock's X: the greatest mana
// value among Elementals its caster controls, read as the cost is
// determined (CR 601.2f). An {X} in the mana cost of a permanent is 0 on
// the battlefield (CR 202.3e), which is what ManaValueForEffect reads.
// No Elemental is 0.
func stdB63GreatestElementalManaValue() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		best := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller != q.Controller || !c.HasSubtypeWord("Elemental") {
				continue
			}
			if mv, ok := q.Game.ManaValueForEffect(c); ok && mv > best {
				best = mv
			}
		}
		return best
	}
}

// stdB63ThirstForIdentityDiscard is "then discard two cards unless you
// discard a creature card" (Thirst for Discovery's shape over creature
// cards).
var stdB63ThirstForIdentityDiscard = discardTwoUnlessOne("discard-two-unless-creature-identity",
	"Thirst for Identity — discard two cards, or one creature card",
	func(c game.Card) bool { return c.IsCreature() })

// stdB63OwnerTucksTargetTopOrBottom is "The owner of target spell or
// creature puts it on their choice of the top or bottom of their library"
// (Swat Away): ownerTucksFirstTargetTopOrBottom for a target that may be a
// spell on the stack as well as a permanent. A spell is PUT into the
// library, not countered, so a spell that can't be countered still goes
// and CR 701.6 never applies.
func stdB63OwnerTucksTargetTopOrBottom(ctx *Context, label string) error {
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Owner == uuid.Nil {
		return nil
	}
	from := game.ZoneBattlefield
	if ctx.Game.StackItemForEffect(id) != nil {
		from = game.ZoneStack
	}
	return PutInLibraryInAnyOrder{
		Chooser:   c.Owner,
		Cards:     []uuid.UUID{id},
		From:      from,
		Placement: game.LibraryPlaceTopOrBottom,
		Label:     label,
	}.Apply(ctx)
}
