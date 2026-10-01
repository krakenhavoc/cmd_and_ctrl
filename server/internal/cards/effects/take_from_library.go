package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// take_from_library.go — "put it into your hand", where "it" is a card
// NAMED out of a look at or a reveal from the top of a library (#952).
//
// The sibling of put_from_library.go, clause for clause, and split the
// same way. The MOVE is engine:
// game.TakeFromLibraryToHandThenForEffect routes library -> hand like
// every other zone change, so the CR 614 window opens, a commander is
// offered CR 903.9, and the continuation is told what ARRIVED. The
// PICK is card text — which of the looked-at cards qualify, "a" or
// "any number", "you may", whether the taken cards are revealed, and
// where the rest go — and it lives here, on a choose_cards prompt with
// Zone: ZoneLibrary.
//
// Hidden information is the caller's first line, not this file's. A
// card that LOOKS calls game.LookAtTopOfLibraryForEffect (only the
// looker becomes a knower); a card that REVEALS the whole top calls
// game.RevealTopOfLibraryForEffect (every seat does). Horn of the Mark
// does both in one sentence — it LOOKS at five and REVEALS only the
// card it takes — which is what TakeFromLibraryToHand.Reveal is for.
//
// Why this exists rather than one more BounceToHand: the issue is in
// game/library_to_hand.go's header. In short, "return it to its
// owner's hand" happening to move a card out of a library is an
// accident of the zone router, and Goblin Ringleader shipped on it.

// TakeFromLibraryToHand offers `Player` the cards in Cards that pass
// Match and puts the chosen ones into their owner's hand.
//
// # The shapes a card prints
//
//   - "Put all Goblin cards revealed this way into your hand"
//     (Goblin Ringleader): All. No prompt — there is no choice.
//   - "You may reveal a creature card from among them and put it into
//     your hand" (Horn of the Mark): Optional, Max 1, Reveal.
//   - "Put any number of them into your hand" : Optional, Max 0,
//     meaning no ceiling.
//   - "You may reveal a land card and/or an instant or sorcery card
//     from among them" (Explore the Vastlands): Optional, Reveal, and
//     two Slots of one card each.
//
// "EACH player looks at the top N cards of their library and may …"
// is this question asked of every player at once, each over their own
// look: EachPlayerTakesFromLibrary.
//
// A mandatory clause whose only legal answer is every candidate is not
// asked either: a prompt with one possible answer is a click, not a
// choice.
//
// # It queues; it does not finish
//
// Apply returns once the prompt is queued. Anything sequenced after it
// in the same Do(…) runs BEFORE the player answers, so "put the rest
// on the bottom of your library in a random order" — which must know
// which cards were NOT taken — goes in Then. Then runs in every case:
// after the answer, immediately when there was nothing to take,
// immediately when All or a forced answer made the choice, and even
// when the move itself returned an error, so the rest are never left
// on top because one card failed. (On that last path `Taken` is empty
// rather than partial — a route that fails a leg never reports what
// landed — while `Rest` is read off the live board and is right
// regardless, which is what "the rest" needs.)
type TakeFromLibraryToHand struct {
	// Player chooses, and is the player taking the cards. uuid.Nil
	// means the resolving item's controller. The cards go to their
	// OWNER's hand, which is the same player for every printed case
	// and is the rule when it is not.
	Player uuid.UUID

	// Cards are the looked-at or revealed cards, top card first — the
	// "them" of "from among them". A card that has since left its
	// library is ignored.
	Cards []uuid.UUID

	// Match is the clause's filter ("a creature card", "Goblin
	// cards"). Nil means any card. Tokens are never candidates (CR
	// 108.2) whatever Match says.
	Match CardPredicate

	// Max is "a" (1) or "any number" (0). Ignored with All, and with
	// Slots, whose capacities are the ceiling.
	Max int

	// Slots is a clause that names more than one kind of card from the
	// same look — "may reveal a land card and/or an instant or sorcery
	// card from among them" (Explore the Vastlands) is two slots of
	// one card each. A card is a candidate when it fits any slot (and
	// Match, when Match is also set); the picked set must fit the
	// slots at once — each card in one slot it matches, no slot over
	// its Max — so a card matching both slots fills only one. See
	// TakeSlot.
	//
	// Still ONE choose_cards prompt with a set-level rule, so the wire
	// is choose_cards' own and internal/legal answers it through
	// ChooseCardsPickLegalLocked like any other Validate. The prompt's
	// ceiling is the largest set the slots can hold out of the actual
	// candidates (three lands and no spell is "up to one"), and for a
	// mandatory clause the floor is that same number.
	Slots []TakeSlot

	// Optional is the printed "you MAY". It sets the prompt's floor to
	// zero.
	Optional bool

	// All is "put ALL [matching] cards … into your hand": no prompt.
	All bool

	// Reveal is the printed "you may REVEAL a creature card from among
	// them and put it into your hand": the cards that are TAKEN become
	// public, while the rest of the look stays private to the looker.
	// It is revealed before the move, while the cards are still in the
	// library, which is where the reveal happens and where the rest of
	// the engine's reveals of library cards happen.
	Reveal bool

	// Validate is the clause's rule about the picked SET, as opposed
	// to Match's rule about each card — "any number of cards with
	// different names", "up to two cards with total mana value 4 or
	// less". Nil means the bounds and Match are the whole rule,
	// which is every card in the catalog that prints this sentence
	// today.
	//
	// It is game.ChooseCardsPrompt.Validate forwarded verbatim, with
	// that field's whole contract, and it is the same field
	// PutFromLibraryOntoBattlefield carries — see there for the long
	// version, including why it switches off the forced-answer
	// shortcut below and why All ignores it.
	Validate func([]game.Card) bool

	// Label is the prompt header, "<card> — <the printed clause>".
	Label string

	// Then runs after the move with what happened. It is the only
	// correct place for "the rest" and for any rider on the result. It
	// runs with g.mu held, may queue further prompts, and must capture
	// only scalars.
	Then func(g *game.Game, res TakeFromLibraryResult) error
}

// TakeFromLibraryResult is what Then is told about the take.
type TakeFromLibraryResult struct {
	// Source is the card whose clause this was.
	Source uuid.UUID
	// Player is who was asked.
	Player uuid.UUID
	// Taken are the cards that reached a hand, in the order given.
	// Empty when none did: a declined "you may", no candidate, or a
	// commander that took CR 903.9's offer instead.
	Taken []uuid.UUID
	// Rest are the cards of Cards that are still in a library — "the
	// rest", "the cards revealed this way that weren't put into your
	// hand" — in the order Cards gave them.
	Rest []uuid.UUID
}

func (p TakeFromLibraryToHand) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source := ctx.Source()
	finish := p.finisher(source, player)
	prompt, answer := p.ask(ctx.Game, player, source)
	if prompt == nil {
		return finish(ctx.Game, answer)
	}
	prompt.Then = finish
	ctx.Game.QueueChooseCardsForEffect(*prompt)
	return nil
}

// ask is the question this take puts to `player`: the choose_cards
// prompt to queue, with its Then left for the caller to set — or, when
// there is no choice to make, nil and the answer itself (nil when
// there is nothing to take).
//
// Split out of Apply so EachPlayerTakesFromLibrary asks every player
// exactly the question a single take would, as one leg of a run.
func (p TakeFromLibraryToHand) ask(g *game.Game, player, source uuid.UUID) (*game.ChooseCardsPrompt, []uuid.UUID) {
	candidates := libraryCardsTakeable(g, player, p.Cards, p.candidateMatch())
	if len(candidates) == 0 {
		return nil, nil
	}
	if p.All {
		return nil, candidates
	}
	hi := p.Max
	var slots *takeSlotRule
	if len(p.Slots) > 0 {
		// The ceiling is how many of the candidates the slots can
		// hold AT ONCE, not the sum of the slots: three lands and no
		// instant fill one slot of "a land card and/or an instant or
		// sorcery card", so the prompt asks for at most one.
		slots = newTakeSlotRule(g, player, p.Slots, candidates)
		hi = slots.fill(candidates)
	}
	if hi <= 0 || hi > len(candidates) {
		hi = len(candidates)
	}
	lo := hi
	if p.Optional {
		lo = 0
	}
	if lo == len(candidates) && p.Validate == nil {
		// The only legal answer is every candidate. Not so with a
		// set rule: it is the rule, not the count, that decides
		// which subsets are answers, and taking the shortcut would
		// perform a set the prompt would have refused. The slots'
		// own rule does not switch it off: a floor equal to the
		// candidate count is only reached when the slots can hold
		// all of them at once, so "all of them" is an answer.
		return nil, candidates
	}
	question := p.Label
	if question == "" {
		question = p.defaultQuestion()
	}
	validate := p.Validate
	if slots != nil {
		validate = slots.validate(p.Validate)
	}
	return &game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: question,
		Cards:    candidates,
		Min:      lo,
		Max:      hi,
		// Re-checked on submit: every pick must still be in a library
		// when the answer arrives.
		Zone:     game.ZoneLibrary,
		Validate: validate,
	}, nil
}

// finisher is what happens once `player`'s picks are known: reveal
// them if the clause reveals, move them, and hand Then the result.
func (p TakeFromLibraryToHand) finisher(source, player uuid.UUID) func(g *game.Game, picked []uuid.UUID) error {
	cards := append([]uuid.UUID(nil), p.Cards...)
	then := p.Then
	reveal, label := p.Reveal, p.Label

	return func(g *game.Game, picked []uuid.UUID) error {
		reported := false
		report := func(g *game.Game, taken []uuid.UUID) error {
			reported = true
			if then == nil {
				return nil
			}
			return then(g, TakeFromLibraryResult{
				Source: source,
				Player: player,
				Taken:  taken,
				Rest:   cardsStillInALibrary(g, cards),
			})
		}
		if len(picked) == 0 {
			return report(g, nil)
		}
		if reveal {
			revealTaken(g, player, source, label, picked)
		}
		moveErr := g.TakeFromLibraryToHandThenForEffect(player, picked, report)
		if moveErr != nil && !reported {
			// A leg failed and the route returned before reaching its
			// continuation, so "the rest" has not been dealt with.
			// Run it anyway: the looked-at cards must not be left on
			// TOP of the library — with the looker still a knower of
			// them — because one card of the batch failed. This is the
			// rule put_from_library.go states and keeps, and the doc
			// above promises.
			//
			// Taken is empty rather than partial: the route reports
			// what landed only through the continuation it did not
			// reach. Rest is read off the live board, so it is right
			// either way, which is what the rest leg actually needs.
			thenErr := report(g, nil)
			_ = thenErr // the move's error is the one worth returning
		}
		return moveErr
	}
}

// revealTaken makes the TAKEN cards public while they are still in the
// library — a reveal is not a move, and doing it after the move would
// be a reveal of cards in a hand. The rest of a look stays private.
func revealTaken(g *game.Game, player, source uuid.UUID, label string, picked []uuid.UUID) {
	if len(picked) == 0 {
		return
	}
	g.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: source,
		Reason: label,
		Cards:  cardsStillInALibrary(g, picked),
	})
}

// candidateMatch is the per-card filter: Match alone, or — with
// Slots — Match and at least one slot's predicate.
func (p TakeFromLibraryToHand) candidateMatch() CardPredicate {
	if len(p.Slots) == 0 {
		return p.Match
	}
	fitsASlot := make([]CardPredicate, len(p.Slots))
	for i, s := range p.Slots {
		fitsASlot[i] = s.matcher()
	}
	if p.Match == nil {
		return Or(fitsASlot...)
	}
	return And(p.Match, Or(fitsASlot...))
}

// defaultQuestion is the prompt header when Label is empty.
func (p TakeFromLibraryToHand) defaultQuestion() string {
	if len(p.Slots) == 0 {
		return "Put a card from among them into your hand"
	}
	labels := make([]string, 0, len(p.Slots))
	for _, s := range p.Slots {
		if s.Label != "" {
			labels = append(labels, s.Label)
		}
	}
	return "Put " + strings.Join(labels, " and/or ") + " from among them into your hand"
}

// TakeSlot is one "a [kind] card" of a take whose sentence names more
// than one kind of card from the same look — Explore the Vastlands'
// "a land card and/or an instant or sorcery card" is two slots.
//
// A slot is not a second prompt. The whole pick is one choose_cards
// question over every candidate that fits ANY slot, with a rule about
// the picked set: each card fills one slot it matches, no slot holds
// more than its Max, and no card fills two. A card that matches both
// slots may fill either, but only one of them.
type TakeSlot struct {
	// Label is the printed clause, "a land card". It names the slot
	// in the default prompt header; nothing parses it.
	Label string

	// Match is which cards may fill the slot. Nil means any card.
	Match CardPredicate

	// Max is how many cards the slot holds: "a" is 1, and 0 means no
	// ceiling — TakeFromLibraryToHand.Max's own convention.
	Max int
}

func (s TakeSlot) matcher() CardPredicate {
	if s.Match != nil {
		return s.Match
	}
	return func(*game.Game, uuid.UUID, game.Card) bool { return true }
}

// takeSlotRule is a slotted take's rule about the picked SET.
//
// Which slots each candidate fits is read ONCE, when the prompt is
// asked, and frozen here with the slot capacities. That is
// game.ChooseCardsPrompt.Validate's contract — it is handed no *Game,
// so anything the rule depends on is a scalar the queuing effect
// captures when it asks, exactly as Cards, Min and Max are — and it is
// sound for this rule: a card's characteristics in a library are its
// printed ones and do not move under an open prompt.
type takeSlotRule struct {
	caps []int
	fits map[uuid.UUID][]int
}

func newTakeSlotRule(g *game.Game, player uuid.UUID, slots []TakeSlot, candidates []uuid.UUID) *takeSlotRule {
	r := &takeSlotRule{
		caps: make([]int, len(slots)),
		fits: make(map[uuid.UUID][]int, len(candidates)),
	}
	for i, s := range slots {
		r.caps[i] = s.Max
		if s.Max <= 0 {
			r.caps[i] = len(candidates)
		}
	}
	for _, id := range candidates {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for i, s := range slots {
			if s.matcher()(g, player, c) {
				r.fits[id] = append(r.fits[id], i)
			}
		}
	}
	return r
}

// fill is how many of `cards` the slots can hold at once — each card
// in a slot it fits, no slot over its capacity, no card in two. A
// maximum bipartite matching by augmenting paths, which for two slots
// and a five-card look is a handful of steps.
func (r *takeSlotRule) fill(cards []uuid.UUID) int {
	holders := make([][]int, len(r.caps))
	var place func(card int, tried []bool) bool
	place = func(card int, tried []bool) bool {
		for _, s := range r.fits[cards[card]] {
			if tried[s] {
				continue
			}
			tried[s] = true
			if len(holders[s]) < r.caps[s] {
				holders[s] = append(holders[s], card)
				return true
			}
			for k, other := range holders[s] {
				if place(other, tried) {
					holders[s][k] = card
					return true
				}
			}
		}
		return false
	}
	n := 0
	for i := range cards {
		if place(i, make([]bool, len(r.caps))) {
			n++
		}
	}
	return n
}

// validate is the prompt's set rule: the slots hold every picked card
// at once, and `also` (the take's own Validate) agrees.
func (r *takeSlotRule) validate(also func([]game.Card) bool) func([]game.Card) bool {
	return func(picked []game.Card) bool {
		ids := make([]uuid.UUID, len(picked))
		for i, c := range picked {
			ids[i] = c.InstanceID
		}
		if r.fill(ids) != len(ids) {
			return false
		}
		return also == nil || also(picked)
	}
}

// libraryCardsTakeable filters `ids` to the cards still in a library
// that pass `pred`, in the order given. Characteristics are read off
// the card in the library, where they are printed.
//
// A token is never a candidate: it is not a "card" (CR 108.2), and a
// token that has left the battlefield cannot be put into a hand (CR
// 111.8). One only reaches a library because the engine has no CR
// 704.5d sweep.
//
// libraryCardsMatching's twin, and separate from it on purpose: that
// one drops every NONPERMANENT card, because "put it onto the
// battlefield" can only mean a permanent. A hand takes anything.
//
// Caller holds g.mu.
func libraryCardsTakeable(g *game.Game, player uuid.UUID, ids []uuid.UUID, pred CardPredicate) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsToken() {
			continue
		}
		if pred != nil && !pred(g, player, c) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// TakeRestOnBottomInRandomOrder is the Then for "put the rest on the
// bottom of your library in a random order" — Horn of the Mark.
//
// Only for cards that PRINT "a random order". A card that prints "in
// any order" uses TakeRestOnBottomInAnyOrder (library_order.go), which
// asks the player (#996, ADR 0088).
func TakeRestOnBottomInRandomOrder(g *game.Game, res TakeFromLibraryResult) error {
	return g.PutOnBottomInRandomOrderForEffect(res.Player, game.ZoneLibrary, res.Rest)
}

// TakeRestIntoGraveyard is the Then for "put the rest into your
// graveyard". Not a mill (CR 701.17a is a count off the TOP), so it
// emits an ordinary zone move — and it goes through
// game.PutIntoGraveyardForEffect, which routes, so Rest in Peace,
// Leyline of the Void and CR 903.9 all see the arrival.
//
// A token among the rest stays where it is (CR 111.8). An error on one
// card does not keep the others out of the graveyard: the loop carries
// on and the first error is returned.
func TakeRestIntoGraveyard(g *game.Game, res TakeFromLibraryResult) error {
	return restIntoGraveyard(g, res.Rest)
}

// LookAtTopThenMayTakeToHand is the whole sentence: "look at the top N
// cards of your library. You may reveal a [Match] card from among them
// and put it into your hand. Put the rest on the bottom of your library
// in a random order" — Horn of the Mark.
//
// A LOOK, so only the looker sees the five; the card taken is revealed,
// which is the half the table is entitled to. Max is 1 for "a" and 0
// for "any number".
func LookAtTopThenMayTakeToHand(n int, match CardPredicate, max int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, n),
			Match:    match,
			Max:      max,
			Optional: true,
			Reveal:   true,
			Label:    label,
			Then:     TakeRestOnBottomInRandomOrder,
		}.Apply(ctx)
	}
}

// RevealTopThenTakeToHand is the whole sentence for the public
// version: "reveal the top N cards of your library. Put all [Match]
// cards revealed this way into your hand and the rest on the bottom of
// your library in any order" — Goblin Ringleader, Sylvan Messenger,
// Garruk, Caller of Beasts. The rest are ordered by the player (#996).
//
// Mandatory and unbounded, which is why it raises no prompt at all:
// "all of them" is the only answer.
func RevealTopThenTakeToHand(ctx *Context, player uuid.UUID, n int, match CardPredicate, label string) error {
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), n, label),
		Match:  match,
		All:    true,
		Label:  label,
		Then:   TakeRestOnBottomInAnyOrder,
	}.Apply(ctx)
}

// EachPlayerTakesFromLibrary is TakeFromLibraryToHand asked of every
// player at once, each over a look at the top of their OWN library —
// "each player looks at the top five cards of their library and may
// reveal a land card and/or an instant or sorcery card from among
// them. Each player puts the cards they revealed this way into their
// hand and the rest on the bottom of their library in a random order"
// (Explore the Vastlands, #1743).
//
// # One question per player, one instruction
//
// Every player is asked exactly the question Take would ask them alone
// — the same candidates, bounds and set rule, from
// TakeFromLibraryToHand's own ask — as one leg of a choose_cards run
// (game.ChooseCardsRunThenForEffect). The legs go up together in
// Players order and may be answered in any order, which is CR 101.4's
// APNAP order for a choice made face down (CR 101.4a): nothing is
// revealed and nothing moves until the last player has answered.
//
// # Hidden information
//
// It is a LOOK. Each player becomes a knower of their own N cards and
// of nobody else's, and no seat sees another seat's prompt
// (choose_cards' non-chooser redaction). With Take.Reveal the cards
// each player TOOK become public — every player's, before any card
// moves, because the choices are made in turn and the actions happen
// together — and the rest of every look stays private. "The rest on
// the bottom of their library in a random order" is Take.Then =
// TakeRestOnBottomInRandomOrder, run for each player with their own
// result, on that player's own random-order stream.
//
// # After the answers
//
// In Players order, each player's picks move to their hand and that
// player's Take.Then runs; the next player's cards move from inside
// that continuation, so a commander's CR 903.9 pause holds the rest of
// the table's moves behind it rather than being overtaken. Then runs
// last, once, with every result. A short library is looked at as far
// as it goes and an empty one is a look at nothing; a player with no
// candidate is not asked, and their whole look is "the rest".
type EachPlayerTakesFromLibrary struct {
	// Players are asked, and their cards move, in this order. Nil
	// means every player still in the game, APNAP from the active
	// player (CR 101.4). A player must not appear twice.
	Players []uuid.UUID

	// N is how many cards each player looks at off the top of their
	// own library.
	N int

	// Take is one player's clause — Match or Slots, Max, Optional,
	// Reveal, Validate, Label — and Take.Then is that player's own
	// "the rest". Take.Player and Take.Cards are filled in per player;
	// whatever they hold here is ignored.
	Take TakeFromLibraryToHand

	// Then runs once, after every player's take has finished, with one
	// result per player in Players order — "each player gains 3 life".
	// It runs with g.mu held, may queue further prompts, and must
	// capture only scalars.
	Then func(g *game.Game, res []TakeFromLibraryResult) error
}

func (e EachPlayerTakesFromLibrary) Apply(ctx *Context) error {
	g := ctx.Game
	source := ctx.Source()
	players := e.Players
	if players == nil {
		players = apnapPlayers(g)
	}
	// Frozen before the run starts and never written after: the run's
	// continuation shares them with every undo snapshot.
	takes := make([]TakeFromLibraryToHand, len(players))
	answers := make([][]uuid.UUID, len(players))
	asked := make([]bool, len(players))
	var legs []game.ChooseCardsPrompt
	for i, player := range players {
		t := e.Take
		t.Player = player
		t.Cards = g.LookAtTopOfLibraryForEffect(player, e.N)
		takes[i] = t
		prompt, answer := t.ask(g, player, source)
		if prompt == nil {
			answers[i] = answer
			continue
		}
		asked[i] = true
		legs = append(legs, *prompt)
	}
	reveal, label, then := e.Take.Reveal, e.Take.Label, e.Then
	_, err := g.ChooseCardsRunThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
		chosen := make([][]uuid.UUID, len(takes))
		for i, t := range takes {
			chosen[i] = answers[i]
			if asked[i] {
				chosen[i] = picks.By(t.Player)
			}
		}
		if reveal {
			for i, t := range takes {
				revealTaken(g, t.Player, source, label, chosen[i])
			}
		}
		return eachPlayerTakeMoves(g, source, takes, chosen, nil, then)
	})
	return err
}

// eachPlayerTakeMoves finishes the take of player len(done) — moves
// their picks and runs their own Then — and goes on to the next player
// from inside that continuation, then runs `then` with every result.
//
// `done` is copied at every step rather than appended in place: the
// continuation can outlive this frame (a paused CR 903.9 leg), and an
// undone-then-replayed answer must not see the first run's entries.
func eachPlayerTakeMoves(g *game.Game, source uuid.UUID, takes []TakeFromLibraryToHand, chosen [][]uuid.UUID,
	done []TakeFromLibraryResult, then func(*game.Game, []TakeFromLibraryResult) error,
) error {
	i := len(done)
	if i == len(takes) {
		if then == nil {
			return nil
		}
		return then(g, done)
	}
	t := takes[i]
	// Every player's take was revealed together, before any card moved.
	t.Reveal = false
	own := t.Then
	t.Then = func(g *game.Game, res TakeFromLibraryResult) error {
		var ownErr error
		if own != nil {
			ownErr = own(g, res)
		}
		next := append(append(make([]TakeFromLibraryResult, 0, i+1), done...), res)
		if err := eachPlayerTakeMoves(g, source, takes, chosen, next, then); err != nil {
			return err
		}
		return ownErr
	}
	return t.finisher(source, t.Player)(g, chosen[i])
}
