package draftpacks

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// Slot names the pack slot a card was drawn for.
type Slot string

// The slots, in pack order.
const (
	SlotCommon   Slot = "common"
	SlotBonus    Slot = "bonus"
	SlotUncommon Slot = "uncommon"
	SlotRare     Slot = "rare"
	SlotLand     Slot = "land"
	SlotWildcard Slot = "wildcard"
)

// Pick is one card of a pack and the slot that drew it.
type Pick struct {
	Card
	Slot Slot `json:"slot"`
}

// Pack is one booster pack: Recipe.Size() cards, in slot order.
type Pack struct {
	Set   string `json:"set"`
	Cards []Pick `json:"cards"`
}

// ErrNoPack is returned for a set the data cannot make a pack for: it is
// not in the dump, or its boosters lack a common, an uncommon, or a rare
// or mythic.
var ErrNoPack = errors.New("draftpacks: the set has no booster cards to make a pack from")

// Generate builds one pack for a set from the index. The same index,
// set and rng state give the same pack. It builds the set's pool on
// every call; use PoolFor and Pool.Pack to make many.
func Generate(idx *cards.Index, set string, rng *rand.Rand) (Pack, error) {
	return PoolFor(idx, set).Pack(rng)
}

// drawer draws cards for one pack without repeating an oracle ID.
type drawer struct {
	rng   *rand.Rand
	taken map[uuid.UUID]bool
}

// draw picks one card of pool uniformly from those not yet in the pack.
// When every card of the pool is already in the pack, the data forces a
// repeat and it picks uniformly from the whole pool. ok is false only
// for an empty pool.
func (d *drawer) draw(pool []Card) (Card, bool) {
	if len(pool) == 0 {
		return Card{}, false
	}
	free := 0
	for _, c := range pool {
		if !d.taken[c.OracleID] {
			free++
		}
	}
	if free == 0 {
		return pool[d.rng.IntN(len(pool))], true
	}
	n := d.rng.IntN(free)
	for _, c := range pool {
		if d.taken[c.OracleID] {
			continue
		}
		if n == 0 {
			d.taken[c.OracleID] = true
			return c, true
		}
		n--
	}
	panic("draftpacks: draw ran past the pool") // unreachable: n < free
}

// oneIn reports a 1-in-n event. n <= 0 never happens; it draws nothing
// from the rng then, so turning a rate off does not shift later draws
// for sets that keep it on.
func oneIn(rng *rand.Rand, n int) bool {
	if n <= 0 {
		return false
	}
	return rng.IntN(n) == 0
}

// free counts the cards of pool not yet in the pack.
func (d *drawer) free(pool []Card) int {
	n := 0
	for _, c := range pool {
		if !d.taken[c.OracleID] {
			n++
		}
	}
	return n
}

// rareOrMythic draws the rare slot: a mythic 1 time in MythicOneIn when
// the set has mythics, a rare otherwise. A recipe with no mythic rate
// has no mythics in its pool (buildPool), so it always draws a rare. A set with only one of the two
// draws from the one it has, and when every card of the rolled rarity
// is already in the pack it takes the other rather than repeat a card.
func (p Pool) rareOrMythic(d *drawer) (Card, bool) {
	first, second := p.Rares, p.Mythics
	if len(p.Mythics) > 0 && (len(p.Rares) == 0 || oneIn(d.rng, p.Recipe.MythicOneIn)) {
		first, second = p.Mythics, p.Rares
	}
	if len(first) == 0 || (d.free(first) == 0 && d.free(second) > 0) {
		first = second
	}
	return d.draw(first)
}

// wildcard draws a card of any rarity: the rarity by the recipe's
// weights over the rarities the set has, then a card of it.
func (p Pool) wildcard(d *drawer) (Card, bool) {
	wc, wu, wr := p.Recipe.WildcardCommon, p.Recipe.WildcardUncommon, p.Recipe.WildcardRareOrMythic
	if len(p.Commons) == 0 {
		wc = 0
	}
	if len(p.Uncommons) == 0 {
		wu = 0
	}
	if len(p.Rares)+len(p.Mythics) == 0 {
		wr = 0
	}
	total := wc + wu + wr
	if total <= 0 {
		// Every weight is off, or names an empty rarity: a common.
		return d.draw(p.Commons)
	}
	switch n := d.rng.IntN(total); {
	case n < wc:
		return d.draw(p.Commons)
	case n < wc+wu:
		return d.draw(p.Uncommons)
	default:
		return p.rareOrMythic(d)
	}
}

// Pack draws one pack from the pool with rng. The same pool and rng
// state give the same pack. No oracle ID repeats within a pack unless
// the pool is too small to avoid it.
func (p Pool) Pack(rng *rand.Rand) (Pack, error) {
	if !p.CanFill() {
		return Pack{}, fmt.Errorf("%w: %q", ErrNoPack, p.Set)
	}
	r := p.Recipe
	d := &drawer{rng: rng, taken: map[uuid.UUID]bool{}}
	out := Pack{Set: p.Set, Cards: make([]Pick, 0, r.Size())}
	add := func(slot Slot, c Card, ok bool) {
		if ok {
			out.Cards = append(out.Cards, Pick{Card: c, Slot: slot})
		}
	}
	common := func(slot Slot) {
		c, ok := d.draw(p.Commons)
		add(slot, c, ok)
	}

	for range r.Commons {
		common(SlotCommon)
	}
	// The bonus slots (a play booster's seventh common slot). The roll
	// is skipped, not just ignored, when the set has no bonus sheet in
	// the data.
	for range r.BonusSlots {
		if len(p.Bonus) > 0 && oneIn(rng, r.BonusSheetOneIn) {
			c, ok := d.draw(p.Bonus)
			add(SlotBonus, c, ok)
		} else {
			common(SlotCommon)
		}
	}
	for range r.Uncommons {
		c, ok := d.draw(p.Uncommons)
		add(SlotUncommon, c, ok)
	}
	c, ok := p.rareOrMythic(d)
	add(SlotRare, c, ok)
	for range r.Lands {
		if len(p.Lands) > 0 {
			c, ok := d.draw(p.Lands)
			add(SlotLand, c, ok)
		} else {
			// A set with no basic lands in its boosters.
			common(SlotLand)
		}
	}
	for range r.Wildcards {
		c, ok := p.wildcard(d)
		add(SlotWildcard, c, ok)
	}
	return out, nil
}
