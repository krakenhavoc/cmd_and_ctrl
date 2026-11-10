package game

import (
	"github.com/google/uuid"
)

// clash.go — the CR 701.30 keyword action "clash" (#5424).
//
//	701.30a To clash, a player reveals the top card of their library.
//	That player may then put that card on the bottom of their library.
//	701.30b "Clash with an opponent" means "Choose an opponent. You and
//	that opponent each clash."
//	701.30c Each clashing player reveals the top card of their library
//	at the same time. Then those players decide in APNAP order (see
//	rule 101.4) where to put those cards, then those cards move at the
//	same time.
//	701.30d A player wins a clash if that player revealed a card with a
//	higher mana value than all other cards revealed in that clash.
//
// One verb, built from parts that already exist: the reveal is
// RevealTopOfLibraryForEffect, and each "top or bottom" question is a
// keyed option pick, so a table waiting on one is a restore point and a
// clasher who leaves the game with the question open is read as "leave it
// on top" (the option pick's drop default) and the clash still finishes,
// unless the one who left controlled the effect, which ends with them
// (CR 800.4a).
// The whole clash rides the prompt as plain data: the two players, the two
// revealed cards, who won, what has been decided so far and the tail the
// card handed in.
//
// The winner is settled by the reveal (CR 701.30d reads the cards
// "revealed", not where they went), so it is worked out once, then, and
// carried. A player whose library was empty revealed no card and so cannot
// win; the other clasher's card is then higher than "all other cards
// revealed" because there are none, and that player wins. Equal mana
// values, or two empty libraries, have no winner.
//
// A card moved to the bottom of its own library has not changed zones, so
// it is lifted out and pushed under, as a scry's bottom lane is: no
// replacement window opens and no zone-move event fires. It was revealed
// and is one card, so everyone still knows it and knows where it is.
//
// Not built: a CR 614 window on the action. No printed card has one.

// EventClashed records that a player clashed (CR 701.30a), once the cards
// of the clash have moved (CR 701.30c). Actor is that player, Target the
// player they clashed with, Source the card whose effect told them to, and
// Label where their revealed card went: ClashPutOnBottom, ClashKeptOnTop,
// or "" when their library was empty and they revealed nothing. One per
// clashing player, the effect's controller first.
const EventClashed EventKind = "clashed"

// EventClashWon records that a player won a clash (CR 701.30d): Actor is
// the winner, Target the player they clashed with, Source the card whose
// effect told them to clash. It follows the clash's EventClashed pair, and
// is not emitted for a clash nobody won. "Whenever you win a clash"
// watches it (Marvo, Deep Operative).
const EventClashWon EventKind = "clash_won"

// The Label of an EventClashed.
const (
	ClashPutOnBottom = "bottom"
	ClashKeptOnTop   = "top"
)

// The two answers to a clash's question, in the order they are offered.
// The first is the one that can always be taken.
const (
	clashTopLabel    = "Leave it on top of your library"
	clashBottomLabel = "Put it on the bottom of your library"
)

// clashRun is one clash in flight. Index 0 is the player whose effect
// said to clash, index 1 the player they clash with.
type clashRun struct {
	source  uuid.UUID
	players [2]uuid.UUID
	// cards are the revealed cards, uuid.Nil for a player who revealed
	// none.
	cards [2]uuid.UUID
	// winner is the index of the player who won, or -1.
	winner int
	// asked and bottom are what has been asked and answered so far.
	asked, bottom [2]bool
	// then is the rest of the card.
	then KeyedTail
}

// carry and nums are the run as a prompt's plain data; clashRunFrom reads
// it back. nums is the winner, the two masks and the index being asked.
func (r clashRun) carry() []uuid.UUID {
	return []uuid.UUID{r.players[0], r.players[1], r.cards[0], r.cards[1]}
}

func (r clashRun) nums(asking int) []int {
	return []int{r.winner, clashMask(r.asked), clashMask(r.bottom), asking}
}

func clashMask(b [2]bool) int {
	m := 0
	for i, set := range b {
		if set {
			m |= 1 << i
		}
	}
	return m
}

func clashUnmask(m int) [2]bool { return [2]bool{m&1 != 0, m&2 != 0} }

// clashRunFrom rebuilds the run a clash question carried, and the index
// of the player it was asked of. False for a prompt that carries no run.
func clashRunFrom(r OptionPicked) (clashRun, int, bool) {
	if len(r.Carry) < 4 || len(r.Nums) < 4 || r.Nums[3] < 0 || r.Nums[3] > 1 {
		return clashRun{}, 0, false
	}
	return clashRun{
		source:  r.Source,
		players: [2]uuid.UUID{r.Carry[0], r.Carry[1]},
		cards:   [2]uuid.UUID{r.Carry[2], r.Carry[3]},
		winner:  r.Nums[0],
		asked:   clashUnmask(r.Nums[1]),
		bottom:  clashUnmask(r.Nums[2]),
		then:    r.Tail,
	}, r.Nums[3], true
}

// ClashForEffect has `player` clash with `opponent` on behalf of `source`
// (CR 701.30): each reveals the top card of their library at the same
// time, each then decides in APNAP order whether that card goes to the
// bottom, and the cards move together. Choosing the opponent ("clash with
// an opponent", CR 701.30b) is the caller's.
//
// `then` is the rest of the card, run once the clash is over: its
// TailOutcome.Amount is 1 when `player` won the clash and 0 otherwise.
// Pass the zero tail when nothing follows. The questions are prompts, so a
// statement after this call would run before they are answered.
//
// Caller must hold g.mu.
func (g *Game) ClashForEffect(source, player, opponent uuid.UUID, then KeyedTail) error {
	run := clashRun{source: source, players: [2]uuid.UUID{player, opponent}, winner: -1, then: then}
	// CR 701.30c: both reveal at the same time, before either decides.
	var values [2]int
	for i, id := range run.players {
		p := g.playerByIDLocked(id)
		if p == nil || p.Eliminated {
			continue
		}
		revealed := g.RevealTopOfLibraryForEffect(id, source, 1, "Clash")
		if len(revealed) == 0 {
			continue
		}
		top, ok := g.cardInZoneLocked(p.Library, revealed[0])
		if !ok {
			continue
		}
		run.cards[i] = top.InstanceID
		// A card with no mana cost has mana value 0 (CR 202.3a).
		values[i] = top.ManaValue()
	}
	// CR 701.30d.
	switch {
	case run.cards[0] != uuid.Nil && (run.cards[1] == uuid.Nil || values[0] > values[1]):
		run.winner = 0
	case run.cards[1] != uuid.Nil && (run.cards[0] == uuid.Nil || values[1] > values[0]):
		run.winner = 1
	}
	return g.clashAskNextLocked(run)
}

// ClashWithAnOpponentForEffect is "clash with an opponent" (CR 701.30b):
// `player` chooses an opponent, then the two clash (ClashForEffect, which
// documents `then`). The choice is part of the action, made as the effect
// resolves, so it is a question and not a target. When no opponent can be
// chosen there is no clash and `then` is told `player` did not win.
//
// Caller must hold g.mu.
func (g *Game) ClashWithAnOpponentForEffect(source, player uuid.UUID, then KeyedTail) error {
	opponents := g.eligibleChosenPlayersLocked(g.apnapSeatsExceptLocked(player))
	if len(opponents) > 0 && g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Clash — choose an opponent to clash with",
		Options:  g.seatChoiceOptionsLocked(opponents),
		ThenKey:  clashOpponentChosen,
		Tail:     then,
	}) != uuid.Nil {
		return nil
	}
	return then.RunTold(g, nil, 0)
}

// clashOpponentChosen is the answer to "clash with an opponent": the clash
// with the opponent chosen, or the rest of the card with no clash when the
// question ended unanswered. An on-disk key: never renamed, never reused.
var clashOpponentChosen = RegisterOptionPickThen("option-pick/engine-clash-opponent", func(g *Game, r OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil {
		return r.Tail.RunTold(g, nil, 0)
	}
	return clashWith(g, r.Source, r.Chooser, r.Option.Player, r.Tail)
})

// clashAskNextLocked asks the next clasher, in APNAP order, where their
// revealed card goes, or finishes the clash when nobody is left to ask.
//
// Caller must hold g.mu.
func (g *Game) clashAskNextLocked(run clashRun) error {
	for _, seat := range g.apnapSeatsExceptLocked(uuid.Nil) {
		for i, id := range run.players {
			if id != seat || run.asked[i] || run.cards[i] == uuid.Nil {
				continue
			}
			run.asked[i] = true
			if g.QueueOptionPickForEffect(OptionPickPrompt{
				Chooser:  id,
				Source:   run.source,
				Question: g.clashQuestionLocked(run, i),
				Options:  []ChoiceOption{{Label: clashTopLabel}, {Label: clashBottomLabel}},
				ThenKey:  clashDecided,
				Carry:    run.carry(),
				Nums:     run.nums(i),
				Tail:     run.then,
			}) != uuid.Nil {
				return nil
			}
		}
	}
	return g.clashFinishLocked(run)
}

// clashQuestionLocked is the question clasher i is asked: their own card,
// and what a clasher who decided before them did (CR 701.30c: the
// decisions are made in order, so the later one is made knowing the
// earlier).
//
// Caller must hold g.mu.
func (g *Game) clashQuestionLocked(run clashRun, i int) string {
	q := "Clash — you revealed " + g.clashCardNameLocked(run, i) + ". Where do you put it?"
	other := 1 - i
	if !run.asked[other] || run.cards[other] == uuid.Nil {
		return q
	}
	where := "left " + g.clashCardNameLocked(run, other) + " on top of their library"
	if run.bottom[other] {
		where = "is putting " + g.clashCardNameLocked(run, other) + " on the bottom of their library"
	}
	return q + " (" + g.seatLabelLocked(run.players[other]) + " " + where + ".)"
}

// clashCardNameLocked is the name of clasher i's revealed card.
//
// Caller must hold g.mu.
func (g *Game) clashCardNameLocked(run clashRun, i int) string {
	if p := g.playerByIDLocked(run.players[i]); p != nil {
		if c, ok := g.cardInZoneLocked(p.Library, run.cards[i]); ok {
			return nameOr(c.Name, "the revealed card")
		}
	}
	return "the revealed card"
}

// clashDecided is the answer to a clash's question: record it, then ask
// the next clasher or finish. A question that ended unanswered (its player
// left the game) leaves the card where it is. An on-disk key: never
// renamed, never reused.
var clashDecided = RegisterOptionPickThen("option-pick/engine-clash", func(g *Game, r OptionPicked) error {
	run, i, ok := clashRunFrom(r)
	if !ok {
		return nil
	}
	run.bottom[i] = r.Option != nil && r.Option.Label == clashBottomLabel
	return clashAskNext(g, run)
})

// clashAskNext is clashAskNextLocked, reached through a variable so the
// package's initialisation order has no cycle (a clash's answer asks the
// next question).
var clashAskNext func(g *Game, run clashRun) error

// clashWith is ClashForEffect, reached the same way.
var clashWith func(g *Game, source, player, opponent uuid.UUID, then KeyedTail) error

func init() {
	clashAskNext = func(g *Game, run clashRun) error { return g.clashAskNextLocked(run) }
	clashWith = func(g *Game, source, player, opponent uuid.UUID, then KeyedTail) error {
		return g.ClashForEffect(source, player, opponent, then)
	}
}

// clashFinishLocked moves the cards (CR 701.30c: at the same time, once
// every decision is in), says who clashed and who won, and runs the rest
// of the card. A card that is no longer on top of its library when the
// answers are in stays where it is.
//
// Caller must hold g.mu.
func (g *Game) clashFinishLocked(run clashRun) error {
	var where [2]string
	for i, id := range run.players {
		if run.cards[i] == uuid.Nil {
			continue
		}
		where[i] = ClashKeptOnTop
		p := g.playerByIDLocked(id)
		if !run.bottom[i] || p == nil || p.Eliminated || p.Library == nil {
			continue
		}
		if n := p.Library.Size(); n == 0 || p.Library.Cards[n-1].InstanceID != run.cards[i] {
			continue
		}
		c, err := p.Library.Remove(run.cards[i])
		if err != nil {
			continue
		}
		p.Library.PushBottom(c)
		where[i] = ClashPutOnBottom
	}
	for i, id := range run.players {
		if !g.playerInGameLocked(id) {
			continue
		}
		g.EmitEvent(Event{Kind: EventClashed, Actor: id, Target: run.players[1-i], Source: run.source, Label: where[i]})
	}
	if w := run.winner; w >= 0 && g.playerInGameLocked(run.players[w]) {
		g.EmitEvent(Event{Kind: EventClashWon, Actor: run.players[w], Target: run.players[1-w], Source: run.source})
	}
	won := 0
	if run.winner == 0 {
		won = 1
	}
	return run.then.RunTold(g, nil, won)
}
