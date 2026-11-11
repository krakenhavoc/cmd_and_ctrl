package draft

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftpacks"
)

// ErrNotDraftable is returned for a set a draft may not hand out: its
// packs can contain a card the catalog does not hold as Full, or the
// data has no pack for it (ADR 0222 §4).
var ErrNotDraftable = errors.New("draft: that set cannot be drafted yet")

// SetSource is a Source for one set's retail-style packs, from the
// draftpacks generator.
type SetSource struct {
	pool draftpacks.Pool
	idx  *cards.Index
}

// NewSetSource returns the pack source of one set, or ErrNotDraftable.
// The check is draftpacks' own (Pool.Draftable), the one GET /draft/sets
// shows, so a draft never hands someone a card that does not work.
func NewSetSource(idx *cards.Index, set string, isFull draftpacks.IsFull) (*SetSource, error) {
	set = strings.ToLower(strings.TrimSpace(set))
	if idx == nil || set == "" || isFull == nil {
		return nil, fmt.Errorf("%w: %q", ErrNotDraftable, set)
	}
	pool := draftpacks.PoolFor(idx, set)
	if !pool.Draftable(isFull) {
		return nil, fmt.Errorf("%w: %q", ErrNotDraftable, set)
	}
	return &SetSource{pool: pool, idx: idx}, nil
}

// NewPoolSource is the pack source of a pool the caller has already
// found draftable (draftpacks.Lister.DraftablePool).
func NewPoolSource(idx *cards.Index, pool draftpacks.Pool) *SetSource {
	return &SetSource{pool: pool, idx: idx}
}

// Booster is the kind of booster the set's packs approximate.
func (s *SetSource) Booster() string { return string(s.pool.Recipe.Kind) }

// Pack draws one pack.
func (s *SetSource) Pack(rng *rand.Rand) ([]Card, error) {
	p, err := s.pool.Pack(rng)
	if err != nil {
		return nil, err
	}
	out := make([]Card, 0, len(p.Cards))
	for _, c := range p.Cards {
		out = append(out, s.card(c.Card))
	}
	return out, nil
}

// card fills in what a pick screen shows beyond the pack generator's
// own fields: colours, mana cost and type line, from the printing.
func (s *SetSource) card(c draftpacks.Card) Card {
	out := Card{
		ScryfallID:  c.ID,
		OracleID:    c.OracleID,
		Name:        c.Name,
		Set:         c.Set,
		CollectorNo: c.CollectorNo,
		Rarity:      string(c.Rarity),
		Colors:      []string{},
		Basic:       c.Basic,
	}
	full, ok := s.idx.Get(c.ID)
	if !ok {
		return out
	}
	fillPrinted(&out, full)
	return out
}

// fillPrinted fills a draft card's colours, mana cost, mana value and
// type line from its printing.
func fillPrinted(out *Card, full cards.Card) {
	out.TypeLine = full.TypeLine
	out.ManaCost = full.ManaCost
	if len(full.Colors) > 0 {
		out.Colors = append(out.Colors, full.Colors...)
	}
	// A two-faced card keeps its cost, and on some layouts its colours
	// and type line, on its front face.
	if len(full.CardFaces) > 0 {
		front := full.CardFaces[0]
		if out.ManaCost == "" {
			out.ManaCost = front.ManaCost
		}
		if out.TypeLine == "" {
			out.TypeLine = front.TypeLine
		}
		if len(out.Colors) == 0 && len(front.Colors) > 0 {
			out.Colors = append(out.Colors, front.Colors...)
		}
	}
	out.ManaValue = ManaValue(out.ManaCost)
}

// A cube's packs (ADR 0222 §8, #5441).
const (
	// DefaultCubePackSize is the pack a cube draft deals unless its
	// host chooses another; MinCubePackSize and MaxCubePackSize bound
	// the choice. Nine is the fewest that leaves a seat a deck's worth
	// of cards after three packs (27 picks, 23 played).
	DefaultCubePackSize = 15
	MinCubePackSize     = 9
	MaxCubePackSize     = 20
)

// ErrCubeTooSmall is a cube asked for more packs than its list fills.
var ErrCubeTooSmall = errors.New("draft: the cube has too few cards for this draft")

// CubeSource is a Source that deals a saved card list (ADR 0222 §8):
// the list is shuffled once, from the draft's seed, and cut into packs
// of one size. There are no rarity slots, and no card is dealt twice:
// a draft's packs are a partition of the front of the shuffle. Cards
// the draft does not need stay undealt.
//
// It implements Dealer, which is how Start draws a whole draft from it.
type CubeSource struct {
	list []Card
	size int
}

// NewCubeSource is the pack source of a card list, dealing packs of
// size cards. The caller has checked the list (cubes.Check): a draft
// never hands someone a card that does not work.
func NewCubeSource(list []cards.Card, size int) (*CubeSource, error) {
	if size < MinCubePackSize || size > MaxCubePackSize {
		return nil, fmt.Errorf("%w: a pack of %d cards (want %d to %d)", ErrBadConfig, size, MinCubePackSize, MaxCubePackSize)
	}
	out := make([]Card, 0, len(list))
	for _, c := range list {
		card := Card{
			ScryfallID:  c.ID,
			OracleID:    c.OracleID,
			Name:        c.Name,
			Set:         c.SetCode,
			CollectorNo: c.CollectorNo,
			Rarity:      c.Rarity,
			Colors:      []string{},
			Basic:       strings.HasPrefix(c.TypeLine, "Basic ") && strings.Contains(c.TypeLine, "Land"),
		}
		fillPrinted(&card, c)
		out = append(out, card)
	}
	return &CubeSource{list: out, size: size}, nil
}

// Size is the cards in one pack; Cards is the cards in the list.
func (s *CubeSource) Size() int  { return s.size }
func (s *CubeSource) Cards() int { return len(s.list) }

// Deal shuffles the list and cuts the first n packs from it.
func (s *CubeSource) Deal(rng *rand.Rand, n int) ([][]Card, error) {
	if n <= 0 || n*s.size > len(s.list) {
		return nil, fmt.Errorf("%w: %d packs of %d need %d cards and it has %d", ErrCubeTooSmall, n, s.size, n*s.size, len(s.list))
	}
	order := rng.Perm(len(s.list))
	packs := make([][]Card, n)
	for p := range packs {
		pack := make([]Card, s.size)
		for i := range pack {
			pack[i] = s.list[order[p*s.size+i]]
		}
		packs[p] = pack
	}
	return packs, nil
}

// Pack deals one pack from a fresh shuffle. A draft does not call it:
// it takes all its packs from one shuffle (Deal).
func (s *CubeSource) Pack(rng *rand.Rand) ([]Card, error) {
	packs, err := s.Deal(rng, 1)
	if err != nil {
		return nil, err
	}
	return packs[0], nil
}

// ManaValue is the mana value of a printed mana cost ("{2}{W}{W}" is
// 4), for sorting a pool into a curve: the total amount of mana in the
// cost (CR 202.3). X counts as 0 off the stack (CR 202.3e), a hybrid
// symbol as its largest component, so "{2/W}" is 2 (CR 202.3f), and a
// Phyrexian symbol as 1 (CR 202.3g). A split card's two costs
// ("{1}{R} // {2}{U}") are added (CR 202.3d).
func ManaValue(cost string) int {
	total := 0
	for {
		open := strings.IndexByte(cost, '{')
		if open < 0 {
			return total
		}
		end := strings.IndexByte(cost[open:], '}')
		if end < 0 {
			return total
		}
		total += symbolValue(cost[open+1 : open+end])
		cost = cost[open+end+1:]
	}
}

func symbolValue(sym string) int {
	best := 0
	for _, part := range strings.Split(sym, "/") {
		v := 1
		switch {
		case part == "":
			v = 0
		case part == "X" || part == "Y" || part == "Z":
			v = 0
		case part == "P": // the Phyrexian half adds nothing of its own
			v = 0
		case part == "HW" || part == "HR": // half mana (Un-sets)
			v = 0
		case part[0] >= '0' && part[0] <= '9':
			v = 0
			for _, r := range part {
				if r < '0' || r > '9' {
					v = 0
					break
				}
				v = v*10 + int(r-'0')
			}
		}
		best = max(best, v)
	}
	return best
}
