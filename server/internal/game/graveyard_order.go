package game

import "github.com/google/uuid"

// graveyard_order.go — #6194: the clauses that read WHERE a card sits in
// a graveyard. CR 404.2: "Each graveyard is kept in a single face-up
// pile. A player can examine the cards in any graveyard at any time but
// normally can't change their order." Zone.Cards keeps that order, the
// last card being the top (the newest arrival, CR 404.3 aside), and a
// snapshot writes the slice as it is, so the order survives a restore.
//
// The readers here are the one place the "top", "bottom" and "above"
// spellings are worked out:
//
//	Exile the top creature card of your graveyard   Zombie Scavengers
//	Exile the bottom card of target player's graveyard   Phyrexian Furnace
//	with a creature card directly above it          Krovikan Horror
//	three or more creature cards above this card    Ashen Ghoul
//	Reorder your graveyard at random                Search for Survivors

// topOfPileOrder keeps the last n of ids (a pile walked bottom to top)
// and returns them topmost first. ids shorter than n are all kept.
func topOfPileOrder(ids []uuid.UUID, n int) []uuid.UUID {
	if n < 0 {
		n = 0
	}
	if n > len(ids) {
		n = len(ids)
	}
	out := make([]uuid.UUID, 0, n)
	for i := len(ids) - 1; i >= len(ids)-n; i-- {
		out = append(out, ids[i])
	}
	return out
}

// TopGraveyardCardsForEffect is the topmost n cards of `player`'s
// graveyard that `match` accepts (nil accepts any card), topmost first:
// "the top creature card of your graveyard" is n 1 with a creature
// predicate, "the top three black cards" n 3 with a colour one. Fewer
// than n when the pile holds fewer matches. `match` is READ-ONLY and
// runs under g.mu. Caller must hold g.mu.
func (g *Game) TopGraveyardCardsForEffect(player uuid.UUID, n int, match func(Card) bool) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Graveyard == nil || n <= 0 {
		return nil
	}
	var out []uuid.UUID
	for i := len(p.Graveyard.Cards) - 1; i >= 0 && len(out) < n; i-- {
		c := p.Graveyard.Cards[i]
		if match == nil || match(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// BottomGraveyardCardForEffect is the bottom card of `player`'s graveyard
// (the one that has been there longest), or uuid.Nil for an empty pile.
// Caller must hold g.mu.
func (g *Game) BottomGraveyardCardForEffect(player uuid.UUID) uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
		return uuid.Nil
	}
	return p.Graveyard.Cards[0].InstanceID
}

// GraveyardCardsAboveForEffect is the cards above `cardID` in the
// graveyard it is in, nearest first: the first is the card "directly
// above it". ok is false when the card is in no graveyard. Caller must
// hold g.mu.
func (g *Game) GraveyardCardsAboveForEffect(cardID uuid.UUID) (above []Card, ok bool) {
	z := g.findCardZoneLocked(cardID)
	if z == nil || z.Kind != ZoneGraveyard {
		return nil, false
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID != cardID {
			continue
		}
		for j := i + 1; j < len(z.Cards); j++ {
			above = append(above, z.Cards[j])
		}
		return above, true
	}
	return nil, false
}

// CardDirectlyAboveIsForEffect reports whether `cardID` is in a graveyard
// with a card directly above it that `match` accepts: Krovikan Horror's
// and Death Spark's "if this card is in your graveyard with a creature
// card directly above it". The top card has nothing above it. Caller
// must hold g.mu.
func (g *Game) CardDirectlyAboveIsForEffect(cardID uuid.UUID, match func(Card) bool) bool {
	above, ok := g.GraveyardCardsAboveForEffect(cardID)
	if !ok || len(above) == 0 {
		return false
	}
	return match == nil || match(above[0])
}

// CountGraveyardCardsAboveForEffect is how many cards above `cardID` in
// its graveyard `match` accepts: Ashen Ghoul's and Nether Shadow's "three
// or more creature cards above this card". Zero when the card is in no
// graveyard. Caller must hold g.mu.
func (g *Game) CountGraveyardCardsAboveForEffect(cardID uuid.UUID, match func(Card) bool) int {
	above, _ := g.GraveyardCardsAboveForEffect(cardID)
	n := 0
	for _, c := range above {
		if match == nil || match(c) {
			n++
		}
	}
	return n
}

// ReorderGraveyardAtRandomForEffect is "Reorder your graveyard at random"
// (Search for Survivors): the pile is shuffled in place from the
// player's random-order stream, so a rewind reproduces it. No card
// changes zone, so nothing triggers and no object becomes a new one
// (CR 400.7 is about zone changes). Caller must hold g.mu in write mode.
func (g *Game) ReorderGraveyardAtRandomForEffect(player uuid.UUID) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) < 2 {
		return
	}
	p.Graveyard.Shuffle(g.randForLocked(rngStream{kind: rngStreamRandomOrder, player: p.ID}))
}

// ReorderGraveyardForEffect puts `player`'s graveyard in `topFirst`'s
// order, the first id on top: "then reorder your graveyard as you
// choose" once the order is known. `topFirst` must name every card in
// the pile once and nothing else, or nothing moves. Caller must hold
// g.mu in write mode.
func (g *Game) ReorderGraveyardForEffect(player uuid.UUID, topFirst []uuid.UUID) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Graveyard == nil {
		return ErrPlayerNotFound
	}
	pile := p.Graveyard.Cards
	if len(topFirst) != len(pile) {
		return ErrInvalidParam
	}
	at := make(map[uuid.UUID]int, len(pile))
	for i := range pile {
		at[pile[i].InstanceID] = i
	}
	next := make([]Card, 0, len(pile))
	seen := make(map[uuid.UUID]bool, len(pile))
	for i := len(topFirst) - 1; i >= 0; i-- {
		id := topFirst[i]
		j, ok := at[id]
		if !ok || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		next = append(next, pile[j])
	}
	p.Graveyard.Cards = next
	return nil
}
