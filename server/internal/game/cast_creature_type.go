package game

import "sort"

// cast_creature_type.go — a creature type chosen as a spell is cast
// (#6245): Selective Snare's "Return X target creatures of the creature
// type of your choice to their owner's hand."
//
//	601.2c The player announces their choice of an appropriate object or
//	       player for each target the spell requires.
//	205.3m Creatures and kindreds share their lists of subtypes; these
//	       subtypes are called creature types.
//
// The type is named with the targets, so it is announced on
// CastSpellParams.ChosenCreatureType and stamped on the spell itself as
// Card.NamedTribe, on the copy the announce gates judge and on the card
// on the stack. Its target clauses are OfSourcesChosenType, the same
// clause Dawn-Blessed Pennant's ability uses, so the announce check (CR
// 601.2c), the legal sets and the resolution re-check (CR 608.2b) all
// read the type off the spell. A copy of the spell copies the choice (CR
// 707.10). The type is the spell's and not the card's: it is cleared as
// the card leaves the stack (CR 400.7).

// CatalogChoosesCreatureTypeAsCast reports whether the catalog card with
// this key names a creature type as it is cast
// (CardDef.ChoosesCreatureTypeAsCast). Nil before the catalog is linked.
var CatalogChoosesCreatureTypeAsCast func(key string) bool

// ChoosesCreatureTypeAsCast reports whether `c` names a creature type as
// it is cast.
func ChoosesCreatureTypeAsCast(c Card) bool {
	return CatalogChoosesCreatureTypeAsCast != nil && CatalogChoosesCreatureTypeAsCast(CatalogKey(c))
}

// castCreatureTypeLocked validates the announcement's chosen creature
// type for `card` and returns it in its canonical spelling: required and
// a real creature type (CR 205.3m) on a spell that chooses one, absent on
// every other.
func castCreatureTypeLocked(card Card, named string) (string, error) {
	if !ChoosesCreatureTypeAsCast(card) {
		if named != "" {
			return "", ErrInvalidParam
		}
		return "", nil
	}
	t, ok := CanonicalCreatureType(named)
	if !ok {
		return "", ErrInvalidParam
	}
	return t, nil
}

// CastCreatureTypeCandidatesForEffect is every creature type a creature
// on the battlefield has, through the layers, in alphabetical order: the
// types worth offering for a spell that names one (the enumerator, the
// view's picker). A changeling has every type (CR 702.73a) and adds none
// of its own: a type it shares with nothing else is a choice no other
// target follows, and the full list of creature types is no menu.
//
// Caller must hold g.mu.
func (g *Game) CastCreatureTypeCandidatesForEffect() []string {
	if g.Battlefield == nil {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	seen := map[string]bool{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() || HasKeyword(c, "changeling") {
			continue
		}
		for _, st := range c.Effective().Subtypes {
			if t, ok := CanonicalCreatureType(st); ok {
				seen[t] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
