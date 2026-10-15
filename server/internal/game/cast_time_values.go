package game

// cast_time_values.go — a value a spell fixes AS IT IS CAST (#3781).
//
//	"Draw X cards, where X is the greatest power among creatures you
//	 controlled as you cast this spell." (Lifestream's Blessing)
//
// CR 601.2 makes the cast a sequence in which the spell is put on the
// stack and its costs are paid; anything the spell's text says is
// determined "as you cast" is fixed by the time the spell is cast
// (CR 601.2i) and does not change afterwards. A creature that dies, or
// grows, in response leaves the answer where it was, which is the whole
// card.
//
// THREE PIECES.
//
//   - The DECLARATION is CardDef.CastTimeValues, a keyed list of
//     CastTimeValue. Keyed rather than a single number so a card with
//     two such statements, or one that also wants a second fact (a
//     count of something at cast time), needs no new field.
//   - The STAMP is stampCastTimeValuesLocked, called by CastSpell as the
//     last act before the cast is announced: every cost is paid, the
//     spell is on the stack, nobody has had priority. A creature
//     sacrificed to pay the cost is therefore not counted, which is the
//     reading "as you cast" gives it (the cost is paid during the cast).
//   - The MEMORY is StackItem.CastValues, a plain map of ints. It is
//     carried by Clone and by the snapshot (additive within the current
//     schema version), and a spell copy keeps it: CR 707.10 copies the
//     values the original fixed as it was cast, and a copy is never cast
//     itself, so there is no second moment at which to recompute.
//
// The reader is effects.Context.CastValue.

// CastTimeValue is one value a spell records as it is cast.
type CastTimeValue struct {
	// Key names the value. Unique per card.
	Key string
	// Value computes it from the game as the cast completes. `item` is
	// the spell's stack item, already carrying its controller, X, modes,
	// targets and the payment record. Called under the game lock, so it
	// reads through the *ForEffect accessors like any other hook.
	Value func(g *Game, item *StackItem) int
}

// CatalogCastTimeValues returns the values the catalog card with the
// given key records as it is cast, or nil. carddef.go sets it from
// CardDef.CastTimeValues.
var CatalogCastTimeValues func(key string) []CastTimeValue

// stampCastTimeValuesLocked records the card's cast-time values on its
// stack item. A card with none leaves the map nil, so every other spell
// serialises exactly as before.
//
// Caller must hold g.mu.
func (g *Game) stampCastTimeValuesLocked(card Card, item *StackItem) {
	if CatalogCastTimeValues == nil || item == nil {
		return
	}
	key := CatalogKey(card)
	if key == "" {
		return
	}
	vals := CatalogCastTimeValues(key)
	if len(vals) == 0 {
		return
	}
	item.CastValues = make(map[string]int, len(vals))
	for _, v := range vals {
		if v.Value != nil {
			item.CastValues[v.Key] = v.Value(g, item)
		}
	}
}

// CastValue is the number the spell recorded under `key` as it was
// cast, and whether it recorded one. A copy reports its original's.
// False for an ability item, for a spell that records nothing under the
// key, and for a game restored from a snapshot that predates the field.
func (s *StackItem) CastValue(key string) (int, bool) {
	if s == nil {
		return 0, false
	}
	v, ok := s.CastValues[key]
	return v, ok
}
