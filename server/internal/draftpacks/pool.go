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
// draws it. It holds only what the set's recipe has a slot for: a recipe
// with no mythic rate has no Mythics, one with no land slot has no
// Lands, and one with no bonus slot has no Bonus. Each slice holds one Card per oracle ID, sorted by name, so
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
// or other placeholder. Scryfall does not flag every set (see collect),
// so this is the rule for a set that has flags, and mainRunNumber is the
// one for a set that has none.
func inBoosters(c cards.Card) bool {
	return c.Booster && paperCard(c)
}

// paperCard is inBoosters without the booster flag: a real paper card of
// the English printing, not a token or other placeholder.
func paperCard(c cards.Card) bool {
	if c.Digital || c.Oversized || c.OracleID == uuid.Nil || c.SetCode == "" {
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

// universesBeyondPromoType is the promo_types value Scryfall puts on
// every printing of a Universes Beyond set, plain ones included. It marks
// the franchise, not a treatment, so plainFrame ignores it; treatment
// does not, and counts it equally on every printing of the set.
const universesBeyondPromoType = "universesbeyond"

// treatment is 0 for a printing in its plain booster frame and grows
// with each sign of a collector treatment. The printing that stands for
// a card is the one with the lowest.
func treatment(c cards.Card) int {
	return treatmentSigns(c, false)
}

// plainFrame reports whether a printing is in its plain frame: treatment
// 0, but with the Universes Beyond marker not counted as a treatment.
func plainFrame(c cards.Card) bool {
	return treatmentSigns(c, true) == 0
}

func treatmentSigns(c cards.Card, ignoreUniversesBeyond bool) int {
	n := 0
	if c.Promo {
		n++
	}
	for _, t := range c.PromoTypes {
		if !ignoreUniversesBeyond || t != universesBeyondPromoType {
			n++
			break
		}
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

// plainNumber is a collector number that is only digits ("214", not
// "214a", "214★" or "A-12"), as a number.
func plainNumber(s string) (int, bool) {
	if s == "" || len(s) > 8 {
		return 0, false
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

// mainRunNumber is the collector number of a printing that can be in the
// main run of a set Scryfall has not flagged for boosters: a paper card
// in its plain frame, not a promo, with a plain collector number.
func mainRunNumber(c cards.Card) (int, bool) {
	if !paperCard(c) || !plainFrame(c) {
		return 0, false
	}
	n, ok := plainNumber(c.CollectorNo)
	return n, ok && n > 0
}

// mainRunLength is how far a set's main run goes: the longest run of
// collector numbers 1, 2, 3, ... that each have a plain-frame, non-promo
// printing. A set prints its main run first and its showcase, extended-art
// and borderless reprints after it, and its basic lands apart from both,
// so the first number with no plain printing is where the main run ends.
// A set with no printing numbered 1 has none (0).
func mainRunLength(plain map[int]bool) int {
	n := 0
	for plain[n+1] {
		n++
	}
	return n
}

// collect walks the index and returns the printings that stand for a
// card in the boosters of every set, or only of the sets in want when
// want is not nil. Every route to a pool (the list, PoolFor, a draft's
// pools and the image scope) goes through it, so they cannot disagree.
//
// A set's cards are the ones Scryfall flags as found in boosters
// (inBoosters). Scryfall has flagged none of the printings of some sets,
// the newest ones among them (Star Trek, Secrets of Strixhaven, TMNT, on
// 2026-11-11, #5689), and such a set has no pool by that rule. For a
// set none of whose printings carries the flag, the cards are its main
// run instead: the printings numbered 1 to mainRunLength that are in the
// plain frame and are not promos. A set with any flagged printing is
// never read this way.
func collect(idx *cards.Index, want map[string]bool) setPrintings {
	out := setPrintings{}
	if idx == nil {
		return out
	}
	wanted := func(c cards.Card) bool {
		return want == nil || want[strings.ToLower(c.SetCode)]
	}
	// Pass one: which sets carry the flag at all, and the plain numbers
	// of the ones that might not.
	flagged := map[string]bool{}
	plain := map[string]map[int]bool{}
	idx.EachPrinting(func(c cards.Card) {
		if !wanted(c) {
			return
		}
		code := strings.ToLower(c.SetCode)
		if c.Booster {
			flagged[code] = true
			return
		}
		if n, ok := mainRunNumber(c); ok {
			if plain[code] == nil {
				plain[code] = map[int]bool{}
			}
			plain[code][n] = true
		}
	})
	runLen := map[string]int{}
	for code, nums := range plain {
		if !flagged[code] {
			runLen[code] = mainRunLength(nums)
		}
	}
	// Pass two: the printings.
	idx.EachPrinting(func(c cards.Card) {
		if !wanted(c) {
			return
		}
		code := strings.ToLower(c.SetCode)
		if flagged[code] {
			if inBoosters(c) {
				out.add(c)
			}
			return
		}
		if n, ok := mainRunNumber(c); ok && n <= runLen[code] {
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
// The recipe decides which slots exist. With no land slot the basic
// lands are left out, with no mythic rate the mythic rares are, and with
// no bonus slot the bonus sheet is: a pack cannot contain them, so the
// draftable check never sees them.
//
// bonus is the bonus sheet's printings, nil for none. A card that is in
// both the set and its bonus sheet stays in the set's own slot.
func buildPool(set string, r Recipe, main, bonus map[uuid.UUID]cards.Card) Pool {
	p := Pool{Set: set, Recipe: r}
	for _, c := range main {
		pc := toCard(c)
		switch {
		case pc.Basic:
			if r.HasLands() {
				p.Lands = append(p.Lands, pc)
			}
		case pc.Rarity == Common:
			p.Commons = append(p.Commons, pc)
		case pc.Rarity == Uncommon:
			p.Uncommons = append(p.Uncommons, pc)
		case pc.Rarity == Rare:
			p.Rares = append(p.Rares, pc)
		case pc.Rarity == Mythic:
			if r.HasMythics() {
				p.Mythics = append(p.Mythics, pc)
			}
		}
	}
	if !r.UsesBonusSheet() {
		bonus = nil
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

// setMeta is what the data says about a set as a whole.
type setMeta struct {
	Name     string
	SetType  string
	Released string
}

// metaFor reads a set's name, type and release date: from the set table
// when the index has one, else from the set's own printings (the most
// frequent release date among them).
func metaFor(idx *cards.Index, code string, main map[uuid.UUID]cards.Card) setMeta {
	var m setMeta
	dates := map[string]int{}
	for _, c := range main {
		if m.Name == "" {
			m.Name = c.SetName
		}
		if m.SetType == "" {
			m.SetType = c.SetType
		}
		if c.ReleasedAt != "" {
			dates[c.ReleasedAt]++
		}
	}
	m.Released = commonDate(dates)
	if idx == nil {
		return m
	}
	if info, ok := idx.Set(code); ok {
		if info.Name != "" {
			m.Name = info.Name
		}
		if info.SetType != "" {
			m.SetType = info.SetType
		}
		if info.ReleasedAt != "" {
			m.Released = info.ReleasedAt
		}
	}
	return m
}

// PoolFor builds the pool of one set from the index, to the set's
// recipe: the recipe of the era the set was released in, with its
// override (RecipeFor). It walks every printing once (twice for a set
// with a bonus sheet), so a caller making many packs builds the pool
// once and calls Pool.Pack on it.
func PoolFor(idx *cards.Index, set string) Pool {
	set = strings.ToLower(strings.TrimSpace(set))
	main := collect(idx, map[string]bool{set: true})[set]
	r := RecipeFor(set, metaFor(idx, set, main).Released)
	if !r.UsesBonusSheet() || r.BonusSheet == set {
		return buildPool(set, r, main, nil)
	}
	return buildPool(set, r, main, collect(idx, map[string]bool{r.BonusSheet: true})[r.BonusSheet])
}

// poolFor is PoolFor to a recipe the caller gives.
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
// still can (the land slot, where the recipe has one, falls back to a
// common).
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
