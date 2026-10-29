package game

import "github.com/google/uuid"

// resolution_pick.go — three resolution-time questions about CARDS
// and PERMANENTS, asked of a named seat while an effect is resolving
// (CR 608.2), each with the rest of the card hanging off the answer.
//
// # The three rows this closes (#1214)
//
// docs/engine-seams.md carried three prompt rows with the same shape
// and no home:
//
//   - "Opponent picks from a revealed set" — Gifts Ungiven, Intuition.
//     You reveal cards (CR 701.20); somebody else says which of them
//     you get.
//   - "A non-owner choosing among another player's permanents" —
//     Tragic Arrogance. PendingChoiceSacrifice is queued
//     {Chooser: playerID, FromPlayer: playerID} at its one queue site
//     (sacrifice_run.go) and ResolveSacrificeChoice refuses anything
//     else, so nothing could point a permanent pick ACROSS the table.
//   - "Resolution-time 'choose N of your own permanents' with
//     continuation" — Scapeshift. The untargeted sibling of a target
//     clause: "sacrifice any number of lands" names no target, so no
//     targeting rule applies to it (hexproof is irrelevant, CR 115.6
//     never runs) and the choice is made on RESOLUTION.
//
// # One payload, three kinds
//
// All three carry the choose-cards payload — ChooseCards, ChooseMin and
// ChooseMax (#4361 S7b: no closure frame) — so isCardSetPickKind admits them
// and they inherit, with no second copy of any of it:
// checkChooseCardsPicksLocked (bounds, candidates, duplicates, the
// live-zone re-check and #1017's set-level Validate),
// ChooseCardsPickLegalLocked (the enumerator's window onto that),
// setCardSetCandidates and the two prunes that keep an open prompt
// honest as the board moves under it.
//
// They are still three KINDS rather than three uses of choose_cards,
// and each difference is observable rather than cosmetic — the test
// surveil had to pass to be a kind rather than a flag on scry:
//
//   - WHAT A SEAT MAY SEE. A reveal_pick's candidates are public by
//     construction: the card revealed them, and that reveal is the
//     only thing entitling a chooser to look at cards out of somebody
//     else's library at all. choose_cards is the opposite — its pool
//     is usually a hand, so filterPendingChoices withholds even the
//     BOUNDS from a non-chooser. Running a reveal through that would
//     hide a public set from the table.
//   - WHICH WAY THE ANSWER POINTS, for a bot. choose_cards is as often
//     the cards a seat gives up; a their_permanents pick ranks somebody
//     ELSE's board by what matters on it (#1014's Options.OrderTargets),
//     and an own_permanents pick ranks the seat's OWN by what it would
//     miss least (Options.OrderCostFuel). One kind cannot be ordered
//     both ways, and a policy that got the sign wrong would keep its
//     best land and hand an opponent their bomb.
//   - WHO INHERITS IT (CR 800.4g). A reveal_pick and a
//     their_permanents pick are about another player's material and
//     move to a survivor; an own_permanents pick is about the chooser's
//     own permanents, which CR 800.4a has taken out of the game with
//     them. See choiceDepartureDecisions (leave_game.go).
//
// # One continuation model: they are RUNS
//
// Every one of the three is queued through runPromptsLocked
// (prompt_run.go), the shape #1019 built for a prompted sacrifice and
// #1027 generalised for a prompted discard. That is not reuse for its
// own sake — it is the only thing that makes the second row work at
// all. Tragic Arrogance is ONE printed instruction ("for each player,
// you choose …") asked as one prompt per player, and its "then" may
// not run until the last of them has been answered.
//
// It also settles the drop path for free. A prompt of each kind is
// one LEG of a run, so choiceDepartureDecisions' dropDefault reaches
// settleRunLegLocked through PendingChoice.promptRun — the branch
// defaultDroppedChoiceLocked already takes on the RUN LINK rather
// than on the kind (#1027) — and a withdrawn prompt settles its leg
// with nothing chosen instead of stranding the card halfway (#544,
// #1006). No kind here needed a fourth case anywhere.

// PendingChoiceRevealPick is "somebody else chooses N of the cards you
// revealed", answered with {card_ids: []string}.
//
// Gifts Ungiven ("target opponent chooses two of those cards") and
// Intuition ("target opponent chooses one") are the cards; the first
// leg of a pile split — Fact or Fiction's "an opponent separates those
// cards into two piles" — is the same question with a floor of zero
// and a ceiling of all of them, and is built on this kind rather than
// beside it (QueuePileSplitForEffect, option_pick.go).
//
// THE CANDIDATES MUST ALREADY BE REVEALED. The chooser is being asked
// about cards they do not own, and protocol's redaction pass shows
// them exactly what they have been made a knower of and nothing else
// (redactChoiceCards): an unrevealed pool reaches them as an empty
// prompt. That is the right failure — the alternative is an engine
// that hands one seat a handle on another's library — and it is why
// this kind has no Zone re-check by default. The cards are revealed
// where they sit, and the continuation re-checks each one as it moves
// it, the way every other effect does.
const PendingChoiceRevealPick PendingChoiceKind = "reveal_pick"

// PendingChoiceTheirPermanents is "choose among the permanents
// somebody ELSE controls", answered with {card_ids: []string}.
//
// The direction the sacrifice run cannot face. "Target opponent
// sacrifices a creature of their choice" is PendingChoiceSacrifice and
// the chooser IS the controller; this is "you choose from among the
// permanents that player controls" (Tragic Arrogance), where the
// chooser is a different seat entirely.
//
// It is NOT targeting, and the difference is observable in exactly the
// way PendingChoiceSacrifice's own comment sets out: a permanent with
// hexproof, shroud or protection can still be chosen this way, and the
// effect needs no legal target to resolve. Reusing pick_target would
// inherit every one of those restrictions silently.
//
// The candidates are permanents on the battlefield, which every seat
// can already see, so nothing about this prompt is redacted.
const PendingChoiceTheirPermanents PendingChoiceKind = "their_permanents"

// PendingChoiceOwnPermanents is "choose N of your own permanents",
// asked at resolution, answered with {card_ids: []string}.
//
// Scapeshift's "sacrifice any number of lands" is the card: an
// untargeted self-choice with a continuation that reads how many were
// chosen ("search your library for up to THAT MANY land cards"). The
// existing sacrifice prompt is a fixed count asked one permanent at a
// time (PlayerSacrificesNForEffect), which cannot say "any number" and
// cannot be read for a total; this is one question over the whole set.
//
// Its sibling is PendingChoiceTheirPermanents above, and the ONLY
// difference is who is asked. Two kinds rather than one flag because
// the answer points the opposite way for a bot (see the file comment)
// and because CR 800.4g reassigns one of them and not the other.
const PendingChoiceOwnPermanents PendingChoiceKind = "own_permanents"

// isResolutionPickKind reports whether a kind is one of the three this
// file declares. Used where a rule is about "a resolution-time pick
// over cards or permanents" rather than about one of them — the log's
// narration gate, and the view's projection.
func isResolutionPickKind(kind PendingChoiceKind) bool {
	switch kind {
	case PendingChoiceRevealPick, PendingChoiceTheirPermanents, PendingChoiceOwnPermanents:
		return true
	}
	return false
}

// PromptedPicks is what a permanent-pick run's continuation is handed:
// one entry per seat whose board the run WALKED, in the order the
// instruction walked them.
//
// The seat on each entry is whose permanents were chosen FROM, not who
// did the choosing — for Tragic Arrogance the chooser is one player
// and the entries are four. That is the reading every card of this
// family wants ("then each player sacrifices all OTHER permanents they
// control"), and it is why the run's ask list is the subjects rather
// than the chooser.
//
// A seat that was skipped at queue time — it controlled nothing the
// effect admits, so CR 608.2's "as much as possible" excused it — has
// no entry at all. `Picked` reads the same for it as for a seat that
// was asked and chose nothing.
//
// ChooseCardsRunThenKeyedForEffect (#1743) hands its continuation the same
// type, and there the seat on each entry IS the chooser: every seat
// chooses from its own look, so "whose cards" and "who chose" are one
// player.
type PromptedPicks []SeatCards

// By returns the permanents chosen from `seat`, or nil.
func (p PromptedPicks) By(seat uuid.UUID) []uuid.UUID { return seatCardsBy(p, seat) }

// Picked reports whether anything was chosen from `seat`.
func (p PromptedPicks) Picked(seat uuid.UUID) bool { return len(seatCardsBy(p, seat)) > 0 }

// Cards flattens the run into every permanent chosen, in ask order.
func (p PromptedPicks) Cards() []uuid.UUID { return seatCardsFlat(p) }

// Count is how many permanents were chosen across every seat.
func (p PromptedPicks) Count() int { return seatCardsCount(p) }

// Contains reports whether `id` was chosen by any leg of the run —
// "all OTHER permanents they control", written as the card prints it.
func (p PromptedPicks) Contains(id uuid.UUID) bool {
	for _, e := range p {
		for _, c := range e.Cards {
			if c == id {
				return true
			}
		}
	}
	return false
}

// RevealPickPrompt is the queue-side description of a
// PendingChoiceRevealPick.
type RevealPickPrompt struct {
	// Chooser picks. An opponent on every printed card of this
	// family; required.
	Chooser uuid.UUID

	// Owner owns the revealed cards — the resolving effect's
	// controller on every printed card. Defaults to Chooser, which is
	// never what a real card means and is here only so a test harness
	// can leave it out. Read by the redaction pass to decide what the
	// chooser may see of a pool that is not theirs.
	Owner uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is the prompt's header — the card's own sentence.
	Question string

	// Cards are the revealed cards, in the order the table saw them
	// revealed. They must ALREADY be revealed; see the kind's doc.
	Cards []uuid.UUID

	// Min / Max bound the pick. Max <= 0 means all of them.
	Min, Max int

	// ThenKey is the continuation as a REGISTERED choose_cards
	// continuation (RegisterChooseCardsThen), and Carry the plain data it
	// is handed (#4264): a reveal pick queued with one holds no closure,
	// so a table waiting on the opponent's choice is a restore point. The
	// continuation gets the cards picked in CardsChosen.Picked, the reveal
	// as CardsChosen.Candidates and the owner as CardsChosen.From. A set
	// rule rides the key (RegisterChooseCardsThenRule), and the pick is one
	// prompt, not a run. (The closure set rule, Validate, is gone: #4361
	// S7b.)
	ThenKey ChooseCardsThen
	Carry   []uuid.UUID
	// Asked and Nums are more plain data for a keyed continuation (#4253
	// sweep, second pass): the asking item (NewAskedItem), handed back as
	// CardsChosen.Asked, and numbers it fixed as it resolved, handed back
	// as CardsChosen.Nums. Read only with ThenKey.
	Asked *AskedItem
	Nums  []int
}

// RevealPickThenForEffect asks `p.Chooser` to pick from the revealed
// set and runs p.ThenKey once they have, handed the cards they PICKED
// and the whole reveal (CardsChosen.Picked, Candidates) — the
// partition, because every card of this family does something with both
// halves ("put the chosen cards into your graveyard and the rest into
// your hand").
//
// Nothing runs when no prompt could go up — an empty set, or a chooser
// who has left the game (CR 800.4a): it reports 0 and the caller carries
// on with nothing picked. The closure form, which ran its continuation
// itself in that case, is gone (#4361). With no key the pick is asked and
// nothing waits on it.
//
// Reports how many prompts went up (0 or 1), so a caller that has to
// distinguish "nobody was asked" from "asked and picked nothing" can.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) RevealPickThenForEffect(p RevealPickPrompt) (int, error) {
	if p.ThenKey.key != "" {
		return g.queueRevealPickKeyedLocked(p), nil
	}
	// Frozen at the ask, like every other prompt's candidate list.
	all := copyUUIDs(p.Cards)
	return g.runPromptsLocked([]uuid.UUID{p.Chooser}, func(seat, run uuid.UUID) bool {
		return g.queueRevealPickLocked(p, all, seat, run) != uuid.Nil
	})
}

// queueRevealPickKeyedLocked queues a reveal pick whose continuation is a
// registered key (#4264) and reports how many prompts went up. The prompt
// carries no frame and no run: the pick is its chooser's alone, so a table
// waiting on it is a restore point. Nothing to pick from, or a chooser who
// has left, queues nothing (QueueChoiceForEffect's CR 800.4a guard).
//
// Caller must hold g.mu.
func (g *Game) queueRevealPickKeyedLocked(p RevealPickPrompt) int {
	all := copyUUIDs(p.Cards)
	if len(all) == 0 {
		return 0
	}
	owner := p.Owner
	if owner == uuid.Nil {
		owner = p.Chooser
	}
	hi := p.Max
	if hi <= 0 || hi > len(all) {
		hi = len(all)
	}
	lo := max(p.Min, 0)
	if lo > hi {
		lo = hi
	}
	if g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceRevealPick,
		Chooser:     p.Chooser,
		FromPlayer:  owner,
		Count:       hi,
		Source:      p.Source,
		Reason:      p.Question,
		ChooseCards: all,
		ChooseMin:   lo,
		ChooseMax:   hi,
		PickThen:    p.ThenKey.key,
		OptionCarry: copyUUIDs(p.Carry),
		Asked:       copyAskedItem(p.Asked),
		PickNums:    copyInts(p.Nums),
	}) == uuid.Nil {
		return 0
	}
	return 1
}

// queueRevealPickLocked queues ONE reveal pick and reports its ID, or
// uuid.Nil when nothing went up.
//
// THE one place a PendingChoiceRevealPick is created, so the run link
// cannot be forgotten by a new caller — queueSacrificePromptLocked's
// rule, and for the same reason.
//
// Caller must hold g.mu.
func (g *Game) queueRevealPickLocked(p RevealPickPrompt, all []uuid.UUID, chooser, run uuid.UUID) uuid.UUID {
	if len(all) == 0 {
		return uuid.Nil
	}
	owner := p.Owner
	if owner == uuid.Nil {
		owner = chooser
	}
	hi := p.Max
	if hi <= 0 || hi > len(all) {
		hi = len(all)
	}
	lo := p.Min
	if lo < 0 {
		lo = 0
	}
	if lo > hi {
		// CR 608.2's "as much as possible": "chooses two of those
		// cards" over a set of one is a choice of one, not a prompt
		// nobody can answer.
		lo = hi
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceRevealPick,
		Chooser:     chooser,
		FromPlayer:  owner,
		Count:       hi,
		Source:      p.Source,
		Reason:      p.Question,
		ChooseCards: copyUUIDs(all),
		ChooseMin:   lo,
		ChooseMax:   hi,
		// No zone re-check: the cards are revealed where they sit
		// (Intuition's three are still in a library) and the pick is a
		// partition of what was revealed, not a claim about where anything
		// is now. The leg settles from these fields (settlePickLegLocked,
		// #4361 S7b), so it holds no closure.
		promptRun: run,
	})
}

// PermanentPickPrompt is the queue-side description of a permanent-pick
// run: ONE seat answering, one prompt per player whose board the
// printed instruction walks.
//
// The seat template pattern PlayersDiscardThenKeyedForEffect uses: `Of` is
// the ask list and everything else is what each of those seats is
// asked with, so "for each player, you choose …" is one call.
type PermanentPickPrompt struct {
	// Chooser answers EVERY leg. Required.
	Chooser uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is each prompt's header.
	Question string

	// Of are the players whose permanents are walked, in ask order —
	// APNAP from the active player for a "for each player" fan-out,
	// which is the caller's business to build. A seat named twice gets
	// two prompts and one entry in the answer.
	Of []uuid.UUID

	// Candidates computes ONE seat's offer: the permanents the chooser
	// may pick from and the bounds on the pick. It is a function
	// rather than a frozen list because the bounds are frequently a
	// reading of that seat's own board — Tragic Arrogance's floor is
	// "one of each of the four types this player actually controls",
	// which differs per seat and cannot be written as a constant.
	//
	// Returning no candidate skips the seat: it is asked nothing, gets
	// no leg, and has no entry in the answer (CR 608.2's "as much as
	// possible"). Runs with g.mu held; it must READ and not write.
	Candidates func(g *Game, of uuid.UUID) (cards []uuid.UUID, min, max int)

	// A set rule over every leg's picks (Tragic Arrogance's "an artifact,
	// a creature, an enchantment and a planeswalker", a rule about the SET
	// that no bound on the count and no per-card predicate can state) is
	// registered with the run (RegisterPromptRunThenRule). The closure
	// form, Validate, is gone (#4361 S7b).
}

// PermanentsPickedThenKeyedForEffect queues the run and returns once every
// prompt is up. `then` runs ONCE, after the last leg has settled, and
// is handed one entry per seat that was asked.
//
// WHICH KIND each leg is queued as is decided here and nowhere else: a
// leg whose subject IS the chooser is a PendingChoiceOwnPermanents, and
// every other leg is a PendingChoiceTheirPermanents. That is the whole
// difference between the two kinds, and putting it in one line is what
// stops a card from having to know which one it is — Tragic Arrogance
// walks every player including its own controller, so one printed
// sentence is both kinds and the card says nothing about either.
//
// `then` runs even when nothing could be asked, for the #544 / #1006
// reason every other run gives.
//
// The rest of the card is a registered run continuation
// (RegisterPromptRunThen), handed `data` beside the answer: annihilator's
// "defending player sacrifices N permanents" (#4290). (The closure form,
// PermanentsPickedThenForEffect, was deleted in #4253's last part once
// only tests called it.)
//
// Each leg is data — the battlefield re-check its kind names and the leg
// settle from its own fields (settlePickLegLocked, #4361 S7b) — so a table
// waiting on any leg is a restore point. The legs name the run's key in
// PickThen, so a binary without it refuses the file. A set rule over the
// picks is registered with the run (RegisterPromptRunThenRule), and each
// leg checks its answer against it.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PermanentsPickedThenKeyedForEffect(p PermanentPickPrompt, then PromptRunThen, data RunData) (int, error) {
	return g.runPromptsKeyedLocked(p.Of, then, p.Source, data, func(seat, run uuid.UUID) bool {
		return g.queuePermanentPickLocked(p, seat, run, then.key) != uuid.Nil
	})
}

// isPermanentPickKind is the two kinds queuePermanentPickLocked queues.
func isPermanentPickKind(k PendingChoiceKind) bool {
	return k == PendingChoiceOwnPermanents || k == PendingChoiceTheirPermanents
}

// queuePermanentPickLocked queues ONE permanent pick and reports its
// ID, or uuid.Nil when the seat was skipped.
//
// THE one place either permanent-pick kind is created.
//
// Caller must hold g.mu.
func (g *Game) queuePermanentPickLocked(p PermanentPickPrompt, of, run uuid.UUID, key string) uuid.UUID {
	if p.Candidates == nil {
		return uuid.Nil
	}
	if who := g.playerByIDLocked(p.Chooser); who == nil || who.Eliminated {
		return uuid.Nil
	}
	cards, lo, hi := p.Candidates(g, of)
	if len(cards) == 0 {
		return uuid.Nil
	}
	if hi <= 0 || hi > len(cards) {
		hi = len(cards)
	}
	if lo < 0 {
		lo = 0
	}
	if lo > hi {
		lo = hi
	}
	kind := PendingChoiceTheirPermanents
	if of == p.Chooser {
		kind = PendingChoiceOwnPermanents
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:        kind,
		Chooser:     p.Chooser,
		FromPlayer:  of,
		Count:       hi,
		Source:      p.Source,
		Reason:      p.Question,
		ChooseCards: copyUUIDs(cards),
		ChooseMin:   lo,
		ChooseMax:   hi,
		promptRun:   run,
		// #4290: a keyed run's leg names the run's key, whose set rule
		// the answer is checked against (checkChooseCardsPicksLocked).
		//
		// #4361 S7b: no frame. The kind names the battlefield, re-checked
		// LIVE on submit and pruned off the open prompt as the board moves
		// (pruneCardSetChoicesLocked): a permanent can leave between the
		// question and the answer, and a prompt that still offered it
		// would be one the resolver refuses. The leg settles keyed by
		// WHOSE permanents were chosen (FromPlayer, see PromptedPicks).
		PickThen: key,
	})
}

// ResolveRevealPick answers a PendingChoiceRevealPick.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveRevealPick(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceRevealPick, choiceID, chooserID, picks)
}

// ResolveTheirPermanents answers a PendingChoiceTheirPermanents.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveTheirPermanents(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceTheirPermanents, choiceID, chooserID, picks)
}

// ResolveOwnPermanents answers a PendingChoiceOwnPermanents.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOwnPermanents(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceOwnPermanents, choiceID, chooserID, picks)
}

// EventCardsChosen records a resolution-time pick in the event log:
// WHO chose, and what they chose (CR 608.2). The public log renders it
// (protocol/log.go, the #1023 gate — a decision the table watched
// somebody make is a decision the history has to carry, or the only
// trace of Tragic Arrogance is four sacrifices with nobody's name on
// them).
//
// Actor is the chooser. CardID is the one card chosen, for the common
// single-card pick, and empty otherwise; Amount is how many were
// chosen either way, so the line can say "chose 2 cards" without
// naming a set the viewer may not be entitled to read. Source is the
// card that asked.
const EventCardsChosen EventKind = "cards_chosen"

// emitCardsChosenLocked narrates one answered leg. A pick of nothing
// is still narrated: "chose nothing" is an answer, and a silence there
// is indistinguishable from a prompt that was never asked.
//
// Caller must hold g.mu.
func (g *Game) emitCardsChosenLocked(chooser, source uuid.UUID, picked []uuid.UUID) {
	ev := Event{
		Kind:   EventCardsChosen,
		Actor:  chooser,
		Source: source,
		Amount: len(picked),
	}
	if len(picked) == 1 {
		ev.CardID = picked[0]
	}
	g.EmitEvent(ev)
}
