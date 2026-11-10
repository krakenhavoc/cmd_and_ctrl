package draftpacks

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// Rarity is one of the four rarities a pack slot draws by.
type Rarity string

// The rarities, spelled as Scryfall spells them.
const (
	Common   Rarity = "common"
	Uncommon Rarity = "uncommon"
	Rare     Rarity = "rare"
	Mythic   Rarity = "mythic"
)

// Card is one card a pack can contain: the one printing that stands for
// its oracle ID in this set.
type Card struct {
	// ID is the Scryfall printing ID, what the image routes take.
	ID uuid.UUID `json:"id"`
	// OracleID is the card's identity, the catalog's key.
	OracleID uuid.UUID `json:"oracle_id"`
	Name     string    `json:"name"`
	// Set and CollectorNo name the printing ("blb", "214").
	Set         string `json:"set"`
	CollectorNo string `json:"collector_number"`
	Rarity      Rarity `json:"rarity"`
	// Basic marks a basic land. The engine plays basic lands without a
	// catalog entry, so they never block a set (see Blockers).
	Basic bool `json:"basic,omitempty"`
}

// Pool is every card one set's packs can contain, split by the slot that
// draws it. Each slice holds one Card per oracle ID, sorted by name, so
// a seeded draw is the same on every run.
type Pool struct {
	// Set is the lowercase set code the pool was built for.
	Set    string
	Recipe Recipe

	Commons   []Card
	Uncommons []Card
	Rares     []Card
	Mythics   []Card
	// Lands are the set's basic lands, one printing each.
	Lands []Card
	// Bonus is the set's bonus sheet, empty when Recipe.BonusSheet is
	// empty or names a set the dump does not hold.
	Bonus []Card
}

// inBoosters reports whether a printing is a card a booster of its set
// can hold: marked as found in boosters, a real paper card, not a token
// or other placeholder.
func inBoosters(c cards.Card) bool {
	if !c.Booster || c.Digital || c.Oversized || c.OracleID == uuid.Nil || c.SetCode == "" {
		return false
	}
	if c.Lang != "" && c.Lang != "en" {
		return false
	}
	if !c.Playable() {
		return false
	}
	switch c.Layout {
	case "emblem", "planar", "scheme", "vanguard":
		return false
	}
	return true
}

// isBasicLand reports whether a printing is a basic land ("Basic Land —
// Forest", "Basic Snow Land — Snow-Covered Forest").
func isBasicLand(c cards.Card) bool {
	return strings.HasPrefix(c.TypeLine, "Basic ") && strings.Contains(c.TypeLine, "Land")
}

// treatmentFrameEffects are the frame_effects values that mark a
// collector treatment rather than the card's plain booster frame.
var treatmentFrameEffects = map[string]bool{
	"showcase":       true,
	"extendedart":    true,
	"inverted":       true,
	"etched":         true,
	"fullart":        true,
	"shatteredglass": true,
	"textless":       true,
}

// treatment is 0 for a printing in its plain booster frame and grows
// with each sign of a collector treatment. The printing that stands for
// a card is the one with the lowest.
func treatment(c cards.Card) int {
	n := 0
	if c.Promo {
		n++
	}
	if len(c.PromoTypes) > 0 {
		n++
	}
	if c.FullArt {
		n++
	}
	if c.BorderColor == "borderless" {
		n++
	}
	for _, e := range c.FrameEffects {
		if treatmentFrameEffects[e] {
			n++
			break
		}
	}
	return n
}

// collectorOrdinal is the leading number of a collector number ("214",
// "214a", "214★" all give 214), or a large value when there is none, so
// a lettered or starred variant sorts after the main run.
func collectorOrdinal(s string) int {
	n, digits := 0, 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
		digits++
		if digits > 8 {
			break
		}
	}
	if digits == 0 {
		return 1 << 30
	}
	return n
}

// plainer reports whether printing a should stand for a card over b:
// the plainer frame first, then the lower collector number (a set's
// main run comes before its showcase and extended-art reprints), then
// the ID, so the choice does not depend on the order the dump is read.
func plainer(a, b cards.Card) bool {
	if ta, tb := treatment(a), treatment(b); ta != tb {
		return ta < tb
	}
	if oa, ob := collectorOrdinal(a.CollectorNo), collectorOrdinal(b.CollectorNo); oa != ob {
		return oa < ob
	}
	if a.CollectorNo != b.CollectorNo {
		return a.CollectorNo < b.CollectorNo
	}
	return a.ID.String() < b.ID.String()
}

// setPrintings collects, for each set code, the one printing per oracle
// ID that stands for the card in that set's boosters.
type setPrintings map[string]map[uuid.UUID]cards.Card

func (s setPrintings) add(c cards.Card) {
	code := strings.ToLower(c.SetCode)
	m := s[code]
	if m == nil {
		m = map[uuid.UUID]cards.Card{}
		s[code] = m
	}
	if cur, ok := m[c.OracleID]; !ok || plainer(c, cur) {
		m[c.OracleID] = c
	}
}

// collect walks the index once and returns the booster printings of
// every set, or only of the sets in want when want is not nil.
func collect(idx *cards.Index, want map[string]bool) setPrintings {
	out := setPrintings{}
	if idx == nil {
		return out
	}
	idx.EachPrinting(func(c cards.Card) {
		if want != nil && !want[strings.ToLower(c.SetCode)] {
			return
		}
		if inBoosters(c) {
			out.add(c)
		}
	})
	return out
}

func toCard(c cards.Card) Card {
	return Card{
		ID:          c.ID,
		OracleID:    c.OracleID,
		Name:        c.Name,
		Set:         strings.ToLower(c.SetCode),
		CollectorNo: c.CollectorNo,
		Rarity:      Rarity(c.Rarity),
		Basic:       isBasicLand(c),
	}
}

func sortCards(cs []Card) {
	sort.Slice(cs, func(a, b int) bool {
		if cs[a].Name != cs[b].Name {
			return cs[a].Name < cs[b].Name
		}
		return cs[a].OracleID.String() < cs[b].OracleID.String()
	})
}

// buildPool sorts one set's booster printings into the slots that draw
// them. Basic lands go to the land slot whatever rarity the dump gives
// them. A card of any other rarity ("special", "bonus") belongs to no
// slot of this model, so no pack can contain it and it is left out.
//
// bonus is the bonus sheet's printings, nil for none. A card that is in
// both the set and its bonus sheet stays in the set's own slot.
func buildPool(set string, r Recipe, main, bonus map[uuid.UUID]cards.Card) Pool {
	p := Pool{Set: set, Recipe: r}
	for _, c := range main {
		pc := toCard(c)
		switch {
		case pc.Basic:
			p.Lands = append(p.Lands, pc)
		case pc.Rarity == Common:
			p.Commons = append(p.Commons, pc)
		case pc.Rarity == Uncommon:
			p.Uncommons = append(p.Uncommons, pc)
		case pc.Rarity == Rare:
			p.Rares = append(p.Rares, pc)
		case pc.Rarity == Mythic:
			p.Mythics = append(p.Mythics, pc)
		}
	}
	for id, c := range bonus {
		if _, dup := main[id]; dup {
			continue
		}
		if pc := toCard(c); !pc.Basic {
			p.Bonus = append(p.Bonus, pc)
		}
	}
	for _, s := range [][]Card{p.Commons, p.Uncommons, p.Rares, p.Mythics, p.Lands, p.Bonus} {
		sortCards(s)
	}
	return p
}

// PoolFor builds the pool of one set from the index, to the set's
// recipe (RecipeFor). It walks every printing once, so a caller making
// many packs builds the pool once and calls Pool.Pack on it.
func PoolFor(idx *cards.Index, set string) Pool {
	return poolFor(idx, set, RecipeFor(set))
}

func poolFor(idx *cards.Index, set string, r Recipe) Pool {
	set = strings.ToLower(strings.TrimSpace(set))
	want := map[string]bool{set: true}
	if r.BonusSheet != "" {
		want[r.BonusSheet] = true
	}
	found := collect(idx, want)
	var bonus map[uuid.UUID]cards.Card
	if r.BonusSheet != "" && r.BonusSheet != set {
		bonus = found[r.BonusSheet]
	}
	return buildPool(set, r, found[set], bonus)
}

// CanFill reports whether the pool can make a pack at all: it has a
// common, an uncommon, and a rare or mythic. A set with no basic lands
// still can (the land slot falls back to a common).
func (p Pool) CanFill() bool {
	return len(p.Commons) > 0 && len(p.Uncommons) > 0 && len(p.Rares)+len(p.Mythics) > 0
}

// Contents is every card a pack from this pool can contain, one per
// oracle ID, sorted by name. It is what the draftable check reads.
func (p Pool) Contents() []Card {
	var out []Card
	seen := map[uuid.UUID]bool{}
	for _, s := range [][]Card{p.Commons, p.Uncommons, p.Rares, p.Mythics, p.Lands, p.Bonus} {
		for _, c := range s {
			if !seen[c.OracleID] {
				seen[c.OracleID] = true
				out = append(out, c)
			}
		}
	}
	sortCards(out)
	return out
}
