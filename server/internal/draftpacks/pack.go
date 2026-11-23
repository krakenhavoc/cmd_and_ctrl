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
	// SlotFoil is a play booster's traditional foil, drawn as a plain
	// card.
	SlotFoil Slot = "foil"
)

// packOrder is the order of the slots in a pack. A bonus card sits after
// the cards of the slot it took the place of.
var packOrder = []Slot{SlotCommon, SlotUncommon, SlotRare, SlotLand, SlotWildcard, SlotFoil}

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

// bySplit draws a card of any rarity: the rarity by the split's weights
// over the rarities the pool has, then a card of it. A rare or mythic
// falls back to the other rather than repeat a card, as the rare slot
// does.
func (p Pool) bySplit(d *drawer, split RaritySplit) (Card, bool) {
	weights := []struct {
		w    int
		pool []Card
	}{
		{split.Common, p.Commons},
		{split.Uncommon, p.Uncommons},
		{split.Rare, p.Rares},
		{split.Mythic, p.Mythics},
	}
	total := 0
	for i := range weights {
		if weights[i].w < 0 || len(weights[i].pool) == 0 {
			weights[i].w = 0
		}
		total += weights[i].w
	}
	if total <= 0 {
		// Every weight is off, or names an empty rarity: a common.
		return d.draw(p.Commons)
	}
	n := d.rng.IntN(total)
	pick := 0
	for i, w := range weights {
		if n < w.w {
			pick = i
			break
		}
		n -= w.w
	}
	switch pick {
	case 2, 3:
		first, second := p.Rares, p.Mythics
		if pick == 3 {
			first, second = p.Mythics, p.Rares
		}
		if d.free(first) == 0 && d.free(second) > 0 {
			first = second
		}
		return d.draw(first)
	default:
		return d.draw(weights[pick].pool)
	}
}

// Pack draws one pack from the pool with rng. The same pool and rng
// state give the same pack. No oracle ID repeats within a pack unless
// the pool is too small to avoid it.
//
// The bonus slots are rolled first, in the recipe's order. Each one that
// opens takes the place of one slot of the kind it replaces, while one
// is left; a slot with no cards in the data never opens, and draws
// nothing from rng to decide so. Then the slots are drawn in packOrder.
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

	left := map[Slot]int{}
	for _, s := range packOrder {
		left[s] = r.slotCount(s)
	}
	opened := map[Slot][]int{}
	for i, b := range r.Bonus {
		if i >= len(p.Bonus) || len(p.Bonus[i]) == 0 || !r.bonusOn(b) || left[b.Slot()] <= 0 {
			continue
		}
		if b.OneIn > 1 && !oneIn(rng, b.OneIn) {
			continue
		}
		left[b.Slot()]--
		opened[b.Slot()] = append(opened[b.Slot()], i)
	}

	for _, slot := range packOrder {
		if slot == SlotRare {
			c, ok := p.rareOrMythic(d)
			add(SlotRare, c, ok)
			continue
		}
		for range left[slot] {
			switch slot {
			case SlotCommon:
				c, ok := d.draw(p.Commons)
				add(SlotCommon, c, ok)
			case SlotUncommon:
				c, ok := d.draw(p.Uncommons)
				add(SlotUncommon, c, ok)
			case SlotLand:
				if len(p.Lands) > 0 {
					c, ok := d.draw(p.Lands)
					add(SlotLand, c, ok)
				} else {
					// A set with no basic lands in its boosters.
					c, ok := d.draw(p.Commons)
					add(SlotLand, c, ok)
				}
			case SlotWildcard:
				c, ok := p.bySplit(d, r.Wildcard)
				add(SlotWildcard, c, ok)
			case SlotFoil:
				c, ok := p.bySplit(d, r.Foil)
				add(SlotFoil, c, ok)
			}
		}
		for _, i := range opened[slot] {
			c, ok := d.draw(p.Bonus[i])
			add(SlotBonus, c, ok)
		}
	}
	return out, nil
}
