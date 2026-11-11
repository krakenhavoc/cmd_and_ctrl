package draft

import (
	"sort"
	"strings"
)

// botdeck.go builds a bot's deck from the pool it drafted (ADR 0222 §7,
// #5439): its two best colours, the best 22 or 23 nonland cards under a
// curve check, and 17 or 18 basic lands split by the colour weight of
// the mana costs.
//
// The deck it returns meets the limited deck rule:
//
//	CR 100.2b: "In limited play (a way of playing in which each player
//	gets the same quantity of unopened Magic product such as booster
//	packs and creates their own deck using only this product and basic
//	land cards), each deck has a minimum deck size of 40 cards. A
//	limited deck may contain as many duplicates of a card as are
//	included with the product."
//
// Every card of the main deck is a card of the pool, taken at most once
// (each Card of a draft is one physical card), and the rest is basic
// lands, which CR 905.1d allows in any number. The cards of the pool
// left out are the sideboard (CR 100.4b).

// BasicLands names the basic land a deck may add for each colour, by
// colour letter: {"W": "Plains", ...}. Nil, or an empty map, is
// StandardBasics.
type BasicLands map[string]string

// StandardBasics is the five basic lands.
var StandardBasics = BasicLands{"W": "Plains", "U": "Island", "B": "Swamp", "R": "Mountain", "G": "Forest"}

// The shape of a bot's deck.
const (
	// DeckSize is the deck BuildDeck returns: the limited minimum (CR
	// 100.2b). A bot never plays more.
	DeckSize = 40
	// DeckSpells and DeckLands are the usual split, 23 and 17. A deck
	// that is heavy plays one spell fewer and one land more.
	DeckSpells = 23
	DeckLands  = 17
	// HeavyAverage is the average mana value of the nonland cards above
	// which a deck is heavy.
	HeavyAverage = 3.3
	// SplashCards is the most cards of a third colour a deck splashes,
	// and SplashLands the lands it plays for them. A deck splashes only
	// when its two colours do not fill it.
	SplashCards = 3
	SplashLands = 3
	// MinColourLands is the fewest lands a deck plays of a colour it
	// has a card of.
	MinColourLands = 3
	// MinDeckCreatures is the fewest creatures a deck plays when its
	// colours have that many. A spell that makes creature tokens counts
	// as one.
	MinDeckCreatures = 13
	// MaxTopEnd is the most cards of TopEndManaValue or more a deck
	// plays when its colours have anything cheaper.
	MaxTopEnd       = 6
	TopEndManaValue = 5
	// ShortCreature and ShortSpell are what a pair of colours loses, in
	// rating points, for each creature its deck is short of
	// MinDeckCreatures and each card it is short of DeckSpells, when
	// the deck's colours are chosen.
	ShortCreature = 3.5
	ShortSpell    = 3.0
)

// BuildDeck builds a 40-card deck from a drafted pool: the bot's deck,
// and what slice 6 (#5440) offers a person as a suggestion.
//
// main is the IDs of the pool's cards in the main deck, in pool order;
// lands is how many of each basic land to add, by the name `basics`
// gives its colour. len(main) plus the lands is always DeckSize. Every
// other card of the pool is the sideboard.
//
// The seed breaks ties between equally rated cards, so two bots with
// like pools do not build alike; the same pool and seed always give the
// same deck. Cards are rated with the server's card index (UseCards).
func BuildDeck(pool []Card, basics BasicLands, seed uint64) (main []string, lands map[string]int) {
	return defaultRater().BuildDeck(pool, basics, seed)
}

// BuildDeck is the package's BuildDeck with this rater.
func (r Rater) BuildDeck(pool []Card, basics BasicLands, seed uint64) (main []string, lands map[string]int) {
	known := false
	for i := range len(colourLetters) {
		known = known || basics[colourLetters[i:i+1]] != ""
	}
	if !known {
		basics = StandardBasics
	}
	var spells []rated
	for _, c := range rateAll(r, pool) {
		// Lands of the pool stay in the sideboard: the deck's lands are
		// basics, split to its colours.
		if c.c.Basic || c.f.Land {
			continue
		}
		// A tie-break far below any real difference in rating.
		c.r += 0.001 * pickNoise(seed, PickRequest{}, c.c.ID)
		spells = append(spells, c)
	}
	sort.SliceStable(spells, func(i, j int) bool {
		if spells[i].r != spells[j].r {
			return spells[i].r > spells[j].r
		}
		return spells[i].c.ID < spells[j].c.ID
	})

	x, y := bestColours(spells)
	in := func(c rated, allowed ...int) bool {
		for _, col := range c.colours {
			ok := false
			for _, a := range allowed {
				ok = ok || col == a
			}
			if !ok {
				return false
			}
		}
		return true
	}
	var cands, rest []rated
	for _, c := range spells {
		if in(c, x, y) {
			cands = append(cands, c)
		} else {
			rest = append(rest, c)
		}
	}
	chosen := chooseSpells(cands, DeckSpells)

	// A splash, only when the pool forces it: the two colours do not
	// fill the deck.
	splash := -1
	if len(chosen) < DeckSpells-1 {
		splash = bestSplash(rest, x, y)
		taken := 0
		for _, c := range rest {
			if splash < 0 || taken >= SplashCards || len(chosen) >= DeckSpells {
				break
			}
			if in(c, x, y, splash) && pips(c.c.ManaCost)[splash] <= 1 && c.r >= RatingPlayable {
				chosen = append(chosen, c)
				taken++
			}
		}
		if taken == 0 {
			splash = -1
		}
	}

	// 23 and 17, or 22 and 18 for a heavy deck.
	if len(chosen) == DeckSpells && averageManaValue(chosen) > HeavyAverage {
		worst := 0
		for i, c := range chosen {
			if c.r < chosen[worst].r {
				worst = i
			}
		}
		chosen = append(chosen[:worst], chosen[worst+1:]...)
	}

	inDeck := map[string]bool{}
	for _, c := range chosen {
		inDeck[c.c.ID] = true
	}
	main = make([]string, 0, len(chosen))
	for _, c := range pool {
		if inDeck[c.ID] {
			main = append(main, c.ID)
		}
	}
	lands = splitLands(chosen, DeckSize-len(chosen), basics, x, y, splash)
	return main, lands
}

// bestColours is the two colours whose cards make the best deck: for
// each pair, the total rating of the DeckSpells cards the pair would
// play (chooseSpells over its playable cards), less ShortCreature for
// each creature it is short of MinDeckCreatures and ShortSpell for each
// card it is short of a full deck. -1 stands for a colour when the pool
// has no coloured card at all.
func bestColours(sorted []rated) (int, int) {
	bx, by, best := -1, -1, 0.0
	for x := range 5 {
		for y := x + 1; y < 5; y++ {
			var cands []rated
			coloured := false
			for _, c := range sorted {
				ok := true
				for _, col := range c.colours {
					ok = ok && (col == x || col == y)
				}
				if ok && c.r >= RatingPlayable {
					cands = append(cands, c)
					coloured = coloured || len(c.colours) > 0
				}
			}
			if !coloured {
				continue
			}
			deck := chooseSpells(cands, DeckSpells)
			total, creatures := 0.0, 0
			for _, c := range deck {
				total += c.r
				if c.body() {
					creatures++
				}
			}
			total -= ShortCreature * float64(max(MinDeckCreatures-creatures, 0))
			total -= ShortSpell * float64(DeckSpells-len(deck))
			if bx < 0 || total > best {
				bx, by, best = x, y, total
			}
		}
	}
	return bx, by
}

// bestSplash is the third colour whose light cards add most.
func bestSplash(rest []rated, x, y int) int {
	best, bestTotal := -1, 0.0
	for z := range 5 {
		if z == x || z == y {
			continue
		}
		total, n := 0.0, 0
		for _, c := range rest {
			if n == SplashCards {
				break
			}
			only := len(c.colours) > 0
			for _, col := range c.colours {
				only = only && (col == x || col == y || col == z)
			}
			if only && pips(c.c.ManaCost)[z] <= 1 && c.r >= RatingPlayable {
				total += c.r
				n++
			}
		}
		if total > bestTotal {
			best, bestTotal = z, total
		}
	}
	return best
}

// chooseSpells takes up to n cards from the candidates, best first under
// the curve check: at each step the card worth most to the deck as it
// stands, which is its rating adjusted for what the deck still needs
// (deckShape.need). A seventh five-drop loses to a slightly worse
// two-drop.
func chooseSpells(cands []rated, n int) []rated {
	var chosen []rated
	var shape deckShape
	used := make([]bool, len(cands))
	creaturesLeft := 0
	for _, c := range cands {
		if c.body() {
			creaturesLeft++
		}
	}
	for len(chosen) < n {
		progress := float64(len(chosen)+1) / float64(n)
		// The two hard parts of the check. The last slots are kept for
		// creatures until the deck has MinDeckCreatures, and the deck
		// takes no more than MaxTopEnd cards of TopEndManaValue or more.
		// Either gives way when the pool has nothing else.
		reserve := creaturesLeft > 0 && MinDeckCreatures-shape.creatures >= n-len(chosen)
		topFull := shape.topEnd() >= MaxTopEnd
		best := -1
		for _, strict := range []struct{ reserve, top bool }{{reserve, topFull}, {reserve, false}, {false, topFull}, {false, false}} {
			bestScore := 0.0
			for i, c := range cands {
				if used[i] || (strict.reserve && !c.body()) || (strict.top && c.f.ManaValue >= TopEndManaValue) {
					continue
				}
				if s := c.r + shape.need(c, progress); best < 0 || s > bestScore {
					best, bestScore = i, s
				}
			}
			if best >= 0 {
				break
			}
		}
		if best < 0 {
			break
		}
		used[best] = true
		chosen = append(chosen, cands[best])
		shape.add(cands[best])
		if cands[best].body() {
			creaturesLeft--
		}
	}
	return chosen
}

func averageManaValue(cs []rated) float64 {
	if len(cs) == 0 {
		return 0
	}
	total := 0
	for _, c := range cs {
		total += c.f.ManaValue
	}
	return float64(total) / float64(len(cs))
}

// pips counts the coloured mana symbols of a printed cost, by colour. A
// hybrid symbol counts half for each of its colours.
func pips(cost string) [5]float64 {
	var out [5]float64
	for {
		open := strings.IndexByte(cost, '{')
		if open < 0 {
			return out
		}
		end := strings.IndexByte(cost[open:], '}')
		if end < 0 {
			return out
		}
		var cols []int
		for _, part := range strings.Split(cost[open+1:open+end], "/") {
			if i := colourIndex(part); i >= 0 {
				cols = append(cols, i)
			}
		}
		for _, i := range cols {
			out[i] += 1 / float64(len(cols))
		}
		cost = cost[open+end+1:]
	}
}

// splitLands divides n basic lands between the deck's colours by the
// coloured mana symbols of its cards, by largest remainder. A colour the
// deck has a card of gets at least MinColourLands, and a splash exactly
// SplashLands.
func splitLands(chosen []rated, n int, basics BasicLands, x, y, splash int) map[string]int {
	var weight [5]float64
	for _, c := range chosen {
		p := pips(c.c.ManaCost)
		for i := range 5 {
			weight[i] += p[i]
		}
	}
	name := func(i int) string {
		if i < 0 {
			return ""
		}
		return basics[colourLetters[i:i+1]]
	}
	lands := map[string]int{}
	left := n
	if splash >= 0 && name(splash) != "" && weight[splash] > 0 {
		k := min(SplashLands, left)
		lands[name(splash)] = k
		left -= k
	}
	var cols []int
	total := 0.0
	for _, i := range []int{x, y} {
		if i >= 0 && name(i) != "" && weight[i] > 0 {
			cols = append(cols, i)
			total += weight[i]
		}
	}
	if len(cols) == 0 {
		// No coloured card to cast: any basic will do. The first the
		// set has, in colour order.
		for _, i := range []int{x, y, 0, 1, 2, 3, 4} {
			if name(i) != "" {
				lands[name(i)] += left
				return lands
			}
		}
		return lands
	}
	if len(cols) == 1 {
		lands[name(cols[0])] += left
		return lands
	}
	a, b := cols[0], cols[1]
	na := int(float64(left)*weight[a]/total + 0.5)
	floor := min(MinColourLands, left/2)
	na = min(max(na, floor), left-floor)
	lands[name(a)] += na
	lands[name(b)] += left - na
	return lands
}
