package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tragic Arrogance — Sorcery {3}{W}{W}:
//
//	"For each player, you choose from among the permanents that player
//	 controls an artifact, a creature, an enchantment, and a
//	 planeswalker. Then each player sacrifices all other nonland
//	 permanents they control."
//
// The card on the "a non-owner choosing among another player's
// permanents" seam row (#1214), and the one that needed the whole
// shape rather than a piece of it:
//
//   - the chooser is NOT the permanents' controller, which
//     PendingChoiceSacrifice cannot express (it is queued
//     {Chooser: p, FromPlayer: p} at its one queue site);
//   - it is ONE printed instruction asked as one prompt per player, so
//     "then each player sacrifices" may not run until the last of them
//     is answered — the RUN shape from #1019 / #1027;
//   - "an artifact, a creature, an enchantment, and a planeswalker" is
//     a rule about the SET that no bound on the count and no per-card
//     predicate can state, so it rides #1017's set-level Validate.
//
// It walks EVERY player including the controller, so the controller's
// own leg is an own_permanents prompt and everyone else's is a
// their_permanents one. The card says nothing about either: the engine
// decides per leg, from whose board is on offer.
//
// # Choosing one of each, when a permanent is more than one type
//
// The official ruling is that each choice is a distinct permanent and
// an artifact creature fills one slot, not two. So the legality of a
// pick is a MATCHING question — can these permanents be assigned
// one-to-one to distinct type slots? — and the floor is the largest
// number of slots that player's board can fill. Both are computed by
// the same four-slot matching (tragicArroganceCover), which is small
// enough to brute force and is the only honest reading of the clause:
// a player with one artifact creature and nothing else keeps exactly
// one permanent, not two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a29bd35-33ef-4317-9fe5-8aaff5d7d64d",
		Name:         "Tragic Arrogance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question:   "Tragic Arrogance — choose an artifact, a creature, an enchantment and a planeswalker that player keeps",
				Of:         seatsFromActive(ctx.Game),
				Candidates: tragicArroganceCandidates,
				Validate:   tragicArroganceAssignable,
				Then:       tragicArroganceSacrificeTheRest,
			}.Apply(ctx)
		},
	})
}

// tragicArroganceTypes are the four slots, in the card's printed order.
var tragicArroganceTypes = []func(game.Card) bool{
	func(c game.Card) bool { return c.IsArtifact() },
	func(c game.Card) bool { return c.IsCreature() },
	func(c game.Card) bool { return c.IsEnchantment() },
	func(c game.Card) bool { return c.IsPlaneswalker() },
}

// seatsFromActive is every seated, living player, APNAP from the
// active player — the order "for each player" walks the table in.
//
// Caller holds g.mu.
func seatsFromActive(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	n := len(g.Seats)
	if n == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < n; i++ {
		p := g.Seats[(start+i)%n]
		if p == nil || p.Eliminated {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// tragicArroganceCandidates is one player's artifacts, creatures,
// enchantments and planeswalkers, with the floor AND ceiling set to
// how many of the four slots that board can actually fill.
//
// Floor equals ceiling because the clause is not optional: you choose
// one of each type the player has, and a player with two creatures and
// no artifact keeps exactly one permanent.
//
// Caller holds g.mu.
func tragicArroganceCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var (
		ids   []uuid.UUID
		cards []game.Card
	)
	for _, c := range g.Battlefield.Cards {
		if c.Controller != of || c.IsLand() {
			continue
		}
		if !tragicArroganceAnyType(c) {
			continue
		}
		ids = append(ids, c.InstanceID)
		cards = append(cards, c)
	}
	n := tragicArroganceCover(cards)
	return ids, n, n
}

// tragicArroganceAnyType reports whether a permanent could fill any of
// the four slots.
func tragicArroganceAnyType(c game.Card) bool {
	for _, is := range tragicArroganceTypes {
		if is(c) {
			return true
		}
	}
	return false
}

// tragicArroganceAssignable is the set-level hook: can these
// permanents be assigned one-to-one to DISTINCT type slots?
//
// With the prompt's floor and ceiling both set to the board's cover,
// "assignable" plus "exactly `cover` of them" is the whole clause — a
// pick of the right size that is assignable covers exactly the slots
// the board can fill.
func tragicArroganceAssignable(picked []game.Card) bool {
	return tragicArroganceMatch(picked, make([]bool, len(tragicArroganceTypes))) == len(picked)
}

// tragicArroganceCover is the largest number of slots `cards` can
// fill — the prompt's floor.
func tragicArroganceCover(cards []game.Card) int {
	best := 0
	for k := len(cards); k >= 0; k-- {
		if k > len(tragicArroganceTypes) {
			continue
		}
		if tragicArroganceCoverOf(cards, k) {
			best = k
			break
		}
	}
	return best
}

// tragicArroganceCoverOf reports whether some k of `cards` can be
// assigned to k distinct slots.
func tragicArroganceCoverOf(cards []game.Card, k int) bool {
	if k == 0 {
		return true
	}
	var walk func(start int, chosen []game.Card) bool
	walk = func(start int, chosen []game.Card) bool {
		if len(chosen) == k {
			return tragicArroganceMatch(chosen, make([]bool, len(tragicArroganceTypes))) == k
		}
		for i := start; i < len(cards); i++ {
			if walk(i+1, append(chosen, cards[i])) {
				return true
			}
		}
		return false
	}
	return walk(0, nil)
}

// tragicArroganceMatch is the matching itself: assign as many of
// `cards` as possible to distinct unused slots, greedily with
// backtracking. Four slots and at most four cards, so the search is
// trivially bounded.
func tragicArroganceMatch(cards []game.Card, used []bool) int {
	if len(cards) == 0 {
		return 0
	}
	best := 0
	for i, is := range tragicArroganceTypes {
		if used[i] || !is(cards[0]) {
			continue
		}
		used[i] = true
		if n := 1 + tragicArroganceMatch(cards[1:], used); n > best {
			best = n
		}
		used[i] = false
	}
	// Leaving this card unassigned is also an option — it is what
	// makes a pick of five cards over four slots come back as four
	// rather than as an error.
	if n := tragicArroganceMatch(cards[1:], used); n > best {
		best = n
	}
	return best
}

// tragicArroganceSacrificeTheRest is the second sentence: "then each
// player sacrifices all other nonland permanents they control."
//
// It sacrifices what was NOT chosen, which is why ChoosePermanents'
// own Sacrifice flag is left off — that one sacrifices the choice.
//
// Caller holds g.mu.
func tragicArroganceSacrificeTheRest(ctx *Context, picked game.PromptedPicks) error {
	g := ctx.Game
	var doomed []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() || picked.Contains(c.InstanceID) {
			continue
		}
		doomed = append(doomed, c.InstanceID)
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.SacrificeAllThenForEffect(ctx.Source(), doomed, nil)
}
