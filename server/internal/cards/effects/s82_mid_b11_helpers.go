package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b11_helpers.go — shared pieces of Midnight Hunt batch 11 (#5497):
// the five Visions sorceries and the cards that count flashback.

// visionsFlashbackAlternative is "Flashback {8}{C}{C}" of a Visions sorcery
// (CR 702.34a). The "costs {X} less" half is visionsFlashbackDiscount.
func VisionsFlashbackAlternative(cost string) game.AlternativeCost { return Flashback(cost) }

// visionsFlashbackDiscount is "This spell costs {X} less to cast this way,
// where X is the greatest mana value of a commander you own on the
// battlefield or in the command zone." It is a self modifier read at CR
// 601.2f, and it applies only when the cast claims the flashback cost
// (CostQuery.AltFlashback): a graveyard cast under another permission pays
// the printed cost, and so does a cast from hand. The reduction spends
// against generic mana only (CR 118.7).
func VisionsFlashbackDiscount() game.CostModifier {
	return CostsLessEach(func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		return greatestOwnedCommanderManaValue(q.Game, q.Controller)
	}, "This spell costs {X} less to cast this way, where X is the greatest mana value of a commander you own on the battlefield or in the command zone",
		func(q game.CostQuery) bool { return q.AltFlashback })
}

// s82MidB11HasPrintedFlashback is "a card with flashback": the card's own
// printed flashback cost, read from the catalog, so a card that was only
// granted flashback (Snapcaster Mage) does not count.
func s82MidB11HasPrintedFlashback(c game.Card) bool {
	return game.AlternativeCostByKey(game.CatalogKey(c), "flashback") != nil
}

// s82MidB11FlashbackYouOwn is the CardPredicate for "card with flashback you own".
func S82MidB11FlashbackYouOwn() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.Owner == caster && s82MidB11HasPrintedFlashback(c)
	}
}

// s82MidB11TypeLineWithCreatureType is a type line with its creature types
// replaced by one — "except it's a 1/1 green Frog": supertypes, card types
// and non-creature subtypes (a land type) stay (CR 707.9b).
func S82MidB11TypeLineWithCreatureType(printed, creatureType string) string {
	super, types, subs := game.ParseTypeLine(printed)
	head := make([]string, 0, len(super)+len(types))
	head = append(head, super...)
	head = append(head, types...)
	var kept []string
	for _, s := range subs {
		if !game.IsCreatureType(s) {
			kept = append(kept, s)
		}
	}
	kept = append(kept, creatureType)
	return strings.Join(head, " ") + " — " + strings.Join(kept, " ")
}
