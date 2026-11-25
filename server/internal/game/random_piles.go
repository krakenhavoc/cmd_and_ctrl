package game

import (
	"strconv"

	"github.com/google/uuid"
)

// random_piles.go — Whims of the Fates (#6248): "Starting with you, each
// player separates all permanents they control into three piles. Then
// each player chooses one of their piles at random and sacrifices those
// permanents."
//
//	700.3. Some effects cause objects to be temporarily grouped into
//	       piles.
//	700.3a Each of the affected objects must be put into exactly one of
//	       those piles, unless the effect specifies otherwise.
//	700.3d A pile can contain zero or more objects.
//
// PileSplitPrompt (option_pick.go) is two piles, split by one player and
// taken by another. This is N piles, each player separating their own
// permanents, and the pile picked at random. Each player, in turn order
// from the effect's controller, is asked N-1 keyed choose_cards legs over
// the permanents they control not yet in a pile ("pile 1 of 3", "pile 2
// of 3"); whatever is left is the last pile (CR 700.3a: every permanent
// is in exactly one). A player who controls nothing has N empty piles and
// is not asked. Once every player has separated, one of each player's
// piles is picked at random with the seeded pick stream (ADR 0054), and
// the card's tail is told the permanents in the picked piles, every
// player's together. Every leg carries the players and the piles so far
// as data, so a table waiting on a pile is a restore point.

// RandomPiles is what SeparateIntoPilesPickOneAtRandomForEffect is handed.
type RandomPiles struct {
	// Source is the card asking.
	Source uuid.UUID
	// Players separate, in this order (the card's "starting with you").
	Players []uuid.UUID
	// Piles is how many piles each player makes: at least two.
	Piles int
	// Label names the card in each prompt.
	Label string
	// Then is told the permanents in the piles picked at random, as IDs.
	Then KeyedTail
}

var randomPilesPile ChooseCardsThen

func init() {
	// An on-disk key: never renamed, never reused.
	randomPilesPile = RegisterChooseCardsThen("choose-cards/random-piles", randomPilesPileAnswered)
}

// SeparateIntoPilesPickOneAtRandomForEffect runs the separation. Caller
// must hold g.mu (write).
func (g *Game) SeparateIntoPilesPickOneAtRandomForEffect(p RandomPiles) error {
	n := max(p.Piles, 2)
	label := p.Label
	if label == "" {
		label = "Piles"
	}
	return g.askNextRandomPileLocked(p.Source, label, copyUUIDs(p.Players), n, 0, nil, nil, p.Then)
}

// askNextRandomPileLocked asks player `who` (an index into players) for
// their next pile. `members` are every pile made so far, flattened, and
// `sizes` their lengths: player 0's N piles, then player 1's, …. With
// every pile made, the piles are picked at random and the tail runs.
//
// Caller must hold g.mu (write).
func (g *Game) askNextRandomPileLocked(source uuid.UUID, label string, players []uuid.UUID, n, who int,
	members []uuid.UUID, sizes []int, then KeyedTail) error {
	for who < len(players) {
		made := len(sizes) - who*n // piles this player has made
		seat := players[who]
		rest := g.unpiledPermanentsLocked(seat, members)
		if !g.playerInGameLocked(seat) || len(rest) == 0 || made == n-1 {
			// The last pile is whatever is left; a player with nothing
			// (or who has left) has empty piles.
			for made < n {
				if made == n-1 && g.playerInGameLocked(seat) {
					members = append(members, rest...)
					sizes = append(sizes, len(rest))
				} else {
					sizes = append(sizes, 0)
				}
				made++
			}
			who++
			continue
		}
		carry := append(copyUUIDs(players), members...)
		nums := append([]int{n, who, len(players)}, sizes...)
		if g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     source,
			Question: label + ": separate all permanents you control into " + strconv.Itoa(n) +
				" piles. Choose pile " + strconv.Itoa(made+1) + " of " + strconv.Itoa(n) +
				". One of your piles will be chosen at random.",
			Cards:     rest,
			Min:       0,
			Max:       len(rest),
			Zone:      ZoneBattlefield,
			ThenKey:   randomPilesPile,
			Carry:     carry,
			CarryNums: nums,
			CarryWord: label,
			Tail:      copyKeyedTail(then),
		}) != uuid.Nil {
			return nil
		}
		// Nobody could be asked: an empty pile.
		sizes = append(sizes, 0)
	}
	return then.RunTold(g, g.pickRandomPilesLocked(source, players, n, members, sizes), 0)
}

// unpiledPermanentsLocked is the permanents `seat` controls that are in
// no pile yet, in battlefield order.
//
// Caller must hold g.mu.
func (g *Game) unpiledPermanentsLocked(seat uuid.UUID, members []uuid.UUID) []uuid.UUID {
	piled := make(map[uuid.UUID]bool, len(members))
	for _, id := range members {
		piled[id] = true
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && !piled[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// pickRandomPilesLocked picks one of each player's piles at random and
// returns the permanents in the picked piles that are still on the
// battlefield under that player's control.
//
// Caller must hold g.mu (write).
func (g *Game) pickRandomPilesLocked(source uuid.UUID, players []uuid.UUID, n int, members []uuid.UUID, sizes []int) []uuid.UUID {
	var out []uuid.UUID
	at := 0
	for who, seat := range players {
		piles := make([][]uuid.UUID, n)
		for k := 0; k < n; k++ {
			i := who*n + k
			size := 0
			if i < len(sizes) {
				size = sizes[i]
			}
			end := min(at+size, len(members))
			piles[k] = members[at:end]
			at = end
		}
		if !g.playerInGameLocked(seat) {
			continue
		}
		// One draw per player, from the pick stream: a pile index.
		chosen := g.randForLocked(rngStream{kind: rngStreamPick, player: seat, source: source}).IntN(n)
		for _, id := range piles[chosen] {
			if c := findBattlefieldCard(g, id); c != nil && c.Controller == seat {
				out = append(out, id)
			}
		}
	}
	return out
}

// randomPilesPileAnswered records one pile and asks for the next. Carry
// is the players, then the piles so far; Nums is [N, the player asked,
// how many players], then the pile sizes.
func randomPilesPileAnswered(g *Game, r CardsChosen) error {
	if len(r.Nums) < 3 {
		return nil
	}
	n, who, np := r.Nums[0], r.Nums[1], r.Nums[2]
	if np > len(r.Carry) {
		return nil
	}
	players := copyUUIDs(r.Carry[:np])
	members := copyUUIDs(r.Carry[np:])
	sizes := copyInts(r.Nums[3:])
	if r.Answered {
		members = append(members, r.Picked...)
		sizes = append(sizes, len(r.Picked))
	} else {
		sizes = append(sizes, 0)
	}
	return g.askNextRandomPileLocked(r.Source, r.Word, players, n, who, members, sizes, r.Tail)
}
