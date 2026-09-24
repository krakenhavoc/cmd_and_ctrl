package game

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
)

// effect_api.go is the exported "locked-context" API that the
// server/internal/cards/effects package calls from inside an
// already-locked mutation frame (the resolution path holds g.mu
// write). Every method here assumes the caller holds g.mu — calling
// a public locking mutator from within a resolution would deadlock.
//
// The naming convention is `XxxForEffectLocked`. Each method is a
// thin wrapper over existing internal helpers that also emits the
// right Event so consumers of Game.Events see primitive-grained
// mutations without having to pattern-match on paired ZoneMove +
// counter events.
//
// Added in S14 sub-PR 2 as the surface the declarative card catalog
// reaches through. Not part of the external HTTP/WS API — the
// resolution path and unit tests are the only callers.

// PlayerByIDForEffect looks up a player by ID without touching the
// lock. Returns nil if no player with that ID is seated. Used by
// primitives to verify target legality and route graveyard moves.
func (g *Game) PlayerByIDForEffect(id uuid.UUID) *Player {
	return g.playerByIDLocked(id)
}

// TurnsBegunFor reports the per-seat turn count from an already-locked
// effect callback. It is the card-facing meaning of "your Nth turn".
func (g *Game) TurnsBegunFor(player uuid.UUID) int {
	return g.turnsBegunForLocked(player)
}

// IsExtraTurn reports whether the current turn came from an effect.
// Like the rest of this file, it is for already-locked card callbacks.
func (g *Game) IsExtraTurn() bool { return g.Turn.Extra }

// StackItemForEffect looks up a stack item by its ID. Returns nil
// if no such item is on the stack. Used by effects that need to
// peek at a countered spell's controller / owner before
// CounterTarget deletes the StackMeta entry (Swan Song).
func (g *Game) StackItemForEffect(id uuid.UUID) *StackItem {
	if g.StackMeta == nil {
		return nil
	}
	return g.StackMeta[id]
}

// StackItemPaidForEffect is what the announcement that put `id` on
// the stack actually paid (#761), and whether the engine charged at
// all. The zero PaidCost for an item that is not on the stack — which
// reads as "nothing was spent, and that is known", the right answer
// for an object that was never cast.
//
// The read a CAST TRIGGER needs: for a spell, the stack item's ID IS
// the card's instance ID, so EventCast.CardID is already the key.
// Vexing Bauble's "whenever a player casts a spell, if no mana was
// spent to cast it" is this accessor and PaidCost.NoManaSpent, and
// nothing else.
//
// A value, not the item, deliberately: a trigger has no business
// mutating the spell it is watching, and the record is small.
func (g *Game) StackItemPaidForEffect(id uuid.UUID) PaidCost {
	if item := g.StackItemForEffect(id); item != nil {
		return item.Paid
	}
	return PaidCost{}
}

// LookupCardForEffect returns a value copy of the card with the
// given instance ID from whichever zone holds it, plus ok=true.
// Empty Card and ok=false when the card isn't in any tracked zone.
// Used by effects that need to read a target's printed
// characteristics BEFORE moving it (Swords to Plowshares → read
// power before exile; Path to Exile → read controller before
// exile to drive the search clause).
func (g *Game) LookupCardForEffect(cardID uuid.UUID) (Card, bool) {
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		// #762: a token whose entry window is open is in no zone at
		// all — it is minted and not yet pushed. An entry replacement
		// reads the entering permanent through this function
		// (Urabrask the Hidden, Kismet, Thalia), so it has to be able
		// to see one. See Game.enteringTokens.
		return g.enteringTokenLocked(cardID)
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return Card{}, false
}

// LastKnownCountersForEffect is the CR 603.10 LKI reader for a
// departing card's COUNTERS (#1218) — The Ozolith's "if it had
// counters on it": by the time a "creature you control leaves the
// battlefield" trigger is judged (whether it is the leaving
// permanent's OWN trigger, read off `sourceLKI`, or a BYSTANDER's,
// like Ozolith watching a permanent that is not itself), MoveCard has
// already zeroed Card.Counters on the departing card, so
// LookupCardForEffect(cardID).Counters answers "none" unconditionally
// (see zone.go). This answers what the card actually had, snapshotted
// by snapshotLKILocked in the same beat lastKnownBattlefield is.
//
// Returns nil for a card that had no counters, was never snapshotted,
// or whose snapshot has already been consumed and cleared — a card
// still on the battlefield with counters right now is not what this
// answers; read Card.Counters directly for that.
//
// Caller must hold g.mu.
func (g *Game) LastKnownCountersForEffect(cardID uuid.UUID) map[string]int {
	return g.lastKnownCounters[cardID]
}

// RevealHandForEffect marks every card in the named player's hand
// as known to all seated players. Used by Thoughtseize / Duress
// style "reveals hand" effects. No-op if the player isn't seated
// or is eliminated.
func (g *Game) RevealHandForEffect(playerID uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return
	}
	for i := range p.Hand.Cards {
		for _, viewer := range g.Seats {
			p.Hand.Cards[i].AddKnower(viewer.ID)
		}
	}
}

// DiscardPrompt is the queue-side description of an effect discard —
// "target player discards two cards", "draw a card, then discard a
// card", "each opponent discards a card." A struct rather than
// positional arguments because everything except Player and N is
// optional in different combinations.
type DiscardPrompt struct {
	// Player is the discarding player, who is also the CHOOSER: per
	// CR 701.8a the player discarding picks the cards, unless the
	// effect says "at random" (DiscardRandomForEffect) or names them.
	// Thoughtseize, where a DIFFERENT player picks, is the other
	// system — QueueDiscardFromRevealedHand, see ADR 0010 §10.
	Player uuid.UUID
	// Source is the card asking. Empty is legal (test harnesses); it
	// is what names the prompt when Question is empty.
	Source uuid.UUID
	// N is the printed count.
	N int
	// UpTo makes N a ceiling rather than an exact count ("discard up
	// to two cards"), which is the difference between a floor of N
	// and a floor of zero on the underlying pick.
	UpTo bool
	// Min is the floor on the pick for the one shape where the
	// COUNT is not the whole rule: "discard two cards UNLESS you
	// discard a creature card" (Teferi Akosa of Zhalfir) accepts a
	// single card when that card is a creature, so its floor is one
	// and Validate refuses the one-card sets that are not.
	//
	// Zero — the usual case — leaves the floor at N, or at zero with
	// UpTo. A Min above N is clamped to N, because the hand may be
	// smaller than the printed count (CR 701.8a). Min and UpTo do
	// not compose; Min wins, and no printed clause wants both.
	//
	// It exists only alongside Validate. A floor below N with
	// nothing judging the SET would simply be a weaker discard, and
	// that is what UpTo is for.
	Min int
	// Question is the prompt's header. Empty composes one from the
	// source card's name, so an ordinary "discard a card" caller
	// stays a one-liner.
	Question string
	// Validate is the set-level legality hook, the same one
	// ChooseCardsPrompt documents: "discard two cards unless you
	// discard a creature card" is a rule about the SET that no count
	// can express. It runs before the prompt is dequeued AND inside
	// the bot enumerator, so an answer `legal` offers is an answer
	// the resolver accepts (#544, #624).
	Validate func(picked []Card) bool
	// Then is everything the card prints after "then" — the rest of
	// the effect, which must not run until the cards are actually in
	// the graveyard. A rummage ("discard a card, then draw a card")
	// is the case that forces it; a loot ("draw a card, then discard
	// a card") draws before it asks and leaves this nil.
	//
	// It also runs, immediately and in the resolving effect's own
	// frame, when the discard asks for nothing because the hand is
	// empty: CR 701.8a discards as many as you can, and the
	// instruction after "then" is not conditional on there having
	// been cards to pitch. Same contract Scry's Then has for an
	// empty library.
	//
	// IT IS THIS PROMPT'S OWN "then", one leg of the instruction —
	// Vicious Rumors' "…discards a card, THEN mills a card" per
	// opponent. The rest of the printed INSTRUCTION, the clause that
	// runs once after every seat has answered and is told what was
	// really discarded, is the RUN's continuation instead
	// (PlayerDiscardsThenForEffect and its two siblings, #1027).
	// Reaching for this one to write a fan-out's payoff gives a card
	// that pays out per answer, which is what Syphon Mind did.
	//
	// `seat` is the player who was asked and `discarded` is what they
	// really discarded (discardedThisWayLocked), both because a
	// fan-out copies ONE template per seat: a closure that captured
	// the player would mill the wrong one, and a leg that read the
	// hand back would count a card a replacement left there. `seat`
	// is p.Player for a single-seat prompt, where the caller knew it
	// already; `discarded` is nil for an empty hand, which is the
	// case this still fires for.
	Then func(g *Game, seat uuid.UUID, discarded []uuid.UUID) error
}

// QueueDiscardChoiceForEffect queues a discard the discarding player
// chooses, and returns the prompt's ID (uuid.Nil when it queued
// nothing).
//
// #651: this used to bump Game.DiscardPending, the cleanup step's
// hand-size map, and that was wrong twice over. Nothing waited for
// the discard — DiscardPending is not a PendingChoice, so priority
// and the step cursor walked straight past a Mind Rot that was still
// owed — and entering cleanup calls populateDiscardPendingLocked,
// which RESETS that map, so an owed effect discard was erased. An
// effect's discard is part of the resolving effect (CR 608.2c), so it
// is a real prompt now: #791's gate refuses advance_step /
// pass_priority / pass_turn while it is open, for free, and
// DiscardPending is back to being only what CR 514.1 uses it for.
//
// The prompt is a PendingChoiceChooseCards over the player's own hand
// with Zone: ZoneHand — the same pick Sylvan Library and
// PutFromHandOntoBattlefield already use — rather than a third
// discard system. The live-zone re-check, the set-level Validate
// hook, the bot enumerator case and the client's ChoicePromptModal
// all come with it; what makes it a DISCARD is the continuation,
// which hands the picks to discardCardsLocked — the one discard path
// (discard.go) — and lets it run Then once they have landed.
//
// Two things it deliberately does not do. It does not prompt for a
// random discard (CR 701.8b — DiscardRandomForEffect stays a
// synchronous move) and it does not prompt for "discard your hand,"
// where there is nothing to choose. And it queues NOTHING for an
// empty hand: CR 701.8a discards as many as you can, and a prompt
// with no candidates and a floor of one is a prompt nobody can
// answer, holding the whole table (#544).
//
// THE FIRE-AND-FORGET FORM since #1027, and what it returns is the
// prompt's ID rather than an outcome: nothing has left the hand when
// it returns, and a clause on the next line pays out for a discard
// nobody has chosen yet. A card with anything hanging off the answer —
// "you draw a card for each card discarded this way", or simply a
// clause printed AFTER the discard, which is Archon of Cruelty's bug —
// uses PlayerDiscardsThenForEffect and its two siblings
// (discard_run.go, ADR 0013 §5y).
//
// Caller must hold g.mu.
func (g *Game) QueueDiscardChoiceForEffect(p DiscardPrompt) uuid.UUID {
	return g.queueDiscardPromptLocked(p, uuid.Nil)
}

// queueDiscardPromptLocked queues ONE discard prompt and returns its
// ID (uuid.Nil when nothing went up). `run` is the prompted-discard
// run the prompt is one leg of (prompt_run.go), or uuid.Nil for a
// prompt nothing is waiting on.
//
// THE one place a discard prompt is created, so the run link cannot be
// forgotten by a new caller — queueSacrificePromptLocked's shape, and
// for the same reason.
//
// Caller must hold g.mu.
func (g *Game) queueDiscardPromptLocked(p DiscardPrompt, run uuid.UUID) uuid.UUID {
	// A prompt addressed to a seat that has left the game can never
	// be answered. Unlike the empty-hand case, Then does NOT run:
	// "each other player discards a card, then you draw a card for
	// each card discarded this way" draws nothing for a player who is
	// no longer there. The RUN is not told about the seat either — it
	// was never asked — which reads the same as a seat that was asked
	// and discarded nothing, and is the right answer for both.
	if p.Player == uuid.Nil || g.chooserGoneLocked(p.Player) {
		return uuid.Nil
	}
	player := g.playerByIDLocked(p.Player)
	n := p.N
	if n > player.Hand.Size() {
		// CR 701.8a — you discard as many as you can.
		n = player.Hand.Size()
	}
	if n <= 0 {
		// Nothing to pitch, but "then draw three" still happens.
		//
		// The RUN is not told, and the seat gets no entry in the
		// answer: no prompt went up, so there is no leg to settle and
		// nothing was owed. That reads the same to a continuation as a
		// seat that was asked and discarded nothing — both discarded
		// nothing — which is the rule PromptedSacrifices states for a
		// seat CR 701.21a excused.
		if p.Then != nil {
			if err := p.Then(g, p.Player, nil); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    p.Player,
					Source:   p.Source,
					ErrorMsg: err.Error(),
				})
			}
		}
		return uuid.Nil
	}
	hand := make([]uuid.UUID, 0, player.Hand.Size())
	for _, c := range player.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	lo := n
	if p.UpTo {
		lo = 0
	}
	if p.Min > 0 {
		lo = min(p.Min, n)
	}
	question := p.Question
	if question == "" {
		question = g.discardQuestionLocked(p.Source, n, p.UpTo)
	}
	discarder, source := p.Player, p.Source
	legThen := p.Then
	opts := discardOptions{
		cause:  DiscardCauseEffect,
		source: p.Source,
		// Reached once the whole batch has landed. This prompt's own
		// "then" is the rest of ITS sentence and goes first; the RUN's
		// leg settle follows, because the run's continuation is the
		// rest of the printed instruction and must be the last thing
		// on the far side of the last leg (#1027).
		then: func(g *Game, landed []uuid.UUID) error {
			if legThen != nil {
				if err := legThen(g, discarder, landed); err != nil {
					g.emitChoiceEffectErrorLocked(discarder, source, err)
				}
			}
			return g.settleRunLegLocked(run, discarder, landed)
		},
	}
	return g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:    p.Player,
		FromPlayer: p.Player,
		Source:     p.Source,
		Question:   question,
		Cards:      hand,
		Min:        lo,
		Max:        n,
		Zone:       ZoneHand,
		Validate:   p.Validate,
		promptRun:  run,
		Then: func(g *Game, picked []uuid.UUID) error {
			return g.discardCardsLocked(discarder, picked, opts)
		},
	})
}

// discardQuestionLocked composes a prompt header for a discard that
// did not supply one: the source card's name plus the printed count,
// so the client modal and the bot's move label both read the way the
// card does. Caller must hold g.mu.
func (g *Game) discardQuestionLocked(source uuid.UUID, n int, upTo bool) string {
	var what string
	switch {
	case upTo && n == 1:
		what = "up to one card"
	case upTo:
		what = "up to " + strconv.Itoa(n) + " cards"
	case n == 1:
		what = "a card"
	default:
		what = strconv.Itoa(n) + " cards"
	}
	if source != uuid.Nil {
		if card, ok := g.LookupCardForEffect(source); ok && card.Name != "" {
			return card.Name + " — discard " + what
		}
	}
	return "Discard " + what
}

// DiscardChoiceForEffect is the plain "this player discards n cards"
// form of QueueDiscardChoiceForEffect, kept because it is the call
// every card in the catalog made before #651 and the one most of them
// still want. A card that needs the source's name on the prompt, an
// "up to", a set-level rule or an instruction after "then" uses the
// struct form.
//
// Caller must hold g.mu.
func (g *Game) DiscardChoiceForEffect(playerID uuid.UUID, n int) {
	g.QueueDiscardChoiceForEffect(DiscardPrompt{Player: playerID, N: n})
}

// FindCardZoneForEffect returns the zone a card currently lives in,
// or nil if the card is in none of the tracked zones. Lock-free —
// caller must hold g.mu. Used by primitives for the CR 608.2b
// per-slot existence re-check.
func (g *Game) FindCardZoneForEffect(cardID uuid.UUID) *Zone {
	return g.findCardZoneLocked(cardID)
}

// BattlefieldCardsForEffect returns the current battlefield card
// slice (by value — mutating it has no effect on the game). Used
// by iterated primitives such as Wrath-of-God's "for each creature
// on battlefield: destroy." Safe because Card carries maps
// (Counters, KnownBy) by reference, but no primitive in S14
// mutates those via this slice.
func (g *Game) BattlefieldCardsForEffect() []Card {
	if g.Battlefield == nil {
		return nil
	}
	out := make([]Card, len(g.Battlefield.Cards))
	copy(out, g.Battlefield.Cards)
	return out
}

// ChangePlayerLifeForEffect adjusts a player's life by delta and
// emits EventChangeLife. Lock-free.
//
// #482: routed through the CR 614 replacement window, the way
// DealDamageToPlayerForEffect has been since S22. Before this, every
// catalog GainLife, every drain and every "pay N life" cost wrote the
// total directly, so a life-change replacement — Rhox Faithmender's
// "you gain twice that much life instead" — only ever saw a life total
// typed in by hand through the public sandbox verb, which is the one
// path a game never takes. See life_tail.go.
//
// A CR 616 ordering prompt (two life replacements on one event)
// returns nil with the change not yet applied; it lands from the
// resume when the affected player answers. This entry point is the
// FIRE-AND-FORGET form and that is fine for it: a caller that only
// says "gain 3" has nothing left to do. A caller that needs to know how
// much actually moved — "you gain life equal to the life lost this
// way" — must use ChangePlayerLifeThenForEffect or
// LoseLifeEachThenForEffect instead, because reading the life total
// back on the next line reads it before the prompt is answered (#793).
func (g *Game) ChangePlayerLifeForEffect(source, playerID uuid.UUID, delta int) error {
	return g.ChangePlayerLifeThenForEffect(source, playerID, delta, nil)
}

// ChangePlayerLifeThenForEffect is a life change with a CONTINUATION:
// `then` runs with the amount that actually moved, once the CR 614
// window has settled it. Lock-free.
//
// This is the life sibling of the search, scry and confirm prompts'
// `Then`, and card authors should learn it as one idiom: the engine
// never hands a catalog effect a half-finished answer, it hands the
// rest of the effect back to the engine. The shape is
//
//	g.ChangePlayerLifeThenForEffect(src, opp, -x, func(g *game.Game, applied int) error {
//	        return g.ChangePlayerLifeForEffect(src, me, -applied)
//	})
//
// for "target opponent loses X life; you gain life equal to the life
// lost this way", and LoseLifeEachThenForEffect for the several-players
// version of the same sentence.
//
// WHAT `then` RECEIVES. The post-replacement delta, signed the way the
// event is: -3 for three life lost, +6 for a gain of three under Rhox
// Faithmender. Zero means nothing moved — the change was replaced away
// (CR 614.10, "your life total can't change") or the player has left
// the game. `then` is told either way, because a caller adding up what
// several players lost would otherwise wait forever on the one that
// lost nothing.
//
// WHEN IT RUNS. Immediately, on the ordinary path, before this function
// returns. After the affected player answers the CR 616 ordering prompt
// when two different life replacements apply to one event — the case
// #793 exists for, and the reason `then` is a continuation rather than
// a return value. Either way it runs with g.mu held and with the
// change's own EventChangeLife already emitted, so it may start the
// next life change or queue the next prompt itself.
//
// `then` takes the live *Game rather than capturing one, on the same
// undo-safety contract every other continuation frame follows.
func (g *Game) ChangePlayerLifeThenForEffect(source, playerID uuid.UUID, delta int, then func(g *Game, applied int) error) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		// #808, CR 800.4a: a player who has left the game neither gains
		// nor loses life, and there is no event for a replacement to
		// see. Not an error — an effect reaching for a seat that
		// conceded mid-resolution has done nothing wrong — but still a
		// terminal outcome, so the continuation is told zero.
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	// The event is built before the pipeline runs and carries
	// everything the tail needs — the continuation included — because a
	// CR 616 ordering prompt returns below without changing anything
	// and the resume has nothing else to go on.
	ev := &ReplacementEvent{
		Kind:       RepEventLife,
		Source:     source,
		LifePlayer: playerID,
		LifeDelta:  delta,
	}
	if then != nil {
		ev.lifeTail = &lifeTail{then: then}
	}
	_, err := g.changeLifeThroughReplacementsLocked(ev)
	return err
}

// LoseLifeEachThenForEffect is the batch form: each of `players` loses
// `amount` life, and then `then` runs with the TOTAL actually lost —
// "each opponent loses X life. You gain life equal to the life lost
// this way" (Exsanguinate, Debt to the Deathless, Gray Merchant of
// Asphodel, Kokusho). Lock-free.
//
// The total is the sum of what each player really lost after their own
// replacements, not amount × len(players): a player under a life-loss
// doubler loses more, a player whose life total can't change loses
// nothing, and the gain follows. It is reported POSITIVE, so the
// continuation reads the way the card does.
//
// Built on ChangePlayerLifeThenForEffect rather than beside it, so
// there is one place that knows how to wait rather than one per card.
// The players are drained IN SEQUENCE, each from the previous one's
// continuation: that is what makes the running total a plain value
// carried forward instead of a shared accumulator, which is what makes
// an undo across the prompt land where a clean run would. A player who
// has left the game — never seated, or eliminated (CR 800.4a), which
// is the case that actually happens because an eliminated seat stays
// in g.Seats — is skipped and adds nothing to the total (#808).
//
// The sequencing is observable only when a loss pauses: with two
// different life replacements on the first opponent, the second
// opponent's loss happens when the CR 616 prompt is answered rather
// than before it. The losses are simultaneous in the rules (CR 101.4)
// and sequential in this engine either way; doing them all on the far
// side of the prompt is the closer of the two, and it is the only one
// that can report a true total.
func (g *Game) LoseLifeEachThenForEffect(source uuid.UUID, players []uuid.UUID, amount int, then func(g *Game, totalLost int) error) error {
	return g.loseLifeEachStepLocked(source, players, amount, 0, then)
}

// loseLifeEachStepLocked drains the head of `players` and continues
// with the tail, carrying the running total forward by value. The empty
// list is the base case: the batch is done and `then` gets the total.
//
// Caller must hold g.mu.
func (g *Game) loseLifeEachStepLocked(source uuid.UUID, players []uuid.UUID, amount, lostSoFar int, then func(g *Game, totalLost int) error) error {
	for len(players) > 0 {
		next, rest := players[0], players[1:]
		if p := g.playerByIDLocked(next); p == nil || p.Eliminated {
			players = rest
			continue
		}
		return g.ChangePlayerLifeThenForEffect(source, next, -amount, func(g *Game, applied int) error {
			lost := lostSoFar
			if applied < 0 {
				// Only a LOSS counts as life lost this way. A
				// replacement that turned the loss into a gain lost
				// nobody anything, and must not be subtracted from
				// what the others lost either.
				lost -= applied
			}
			return g.loseLifeEachStepLocked(source, rest, amount, lost, then)
		})
	}
	if then == nil {
		return nil
	}
	return then(g, lostSoFar)
}

// PayLifeForEffect pays `amount` life as a COST (CR 118.3) — a spell's
// additional or alternative cost, an activated ability's {T}, Pay 2
// life, a ward's "unless that player pays N life", a shockland's "as
// this enters, you may pay 2 life". Lock-free.
//
// CR 119.4 makes paying life the same as losing that much life, so this
// runs the same CR 614 window every other life change runs and a
// life-loss replacement sees it. What it will not do is PAUSE: a cost
// is paid as one indivisible step of casting or activating (CR 601.2h,
// CR 602.2b), so the window settles without a prompt. The full argument
// and what the affected player gives up for it are in
// payLifeAsCostLocked (life_tail.go) and ADR 0013 §5b.
//
// Returns ErrInvalidParam when the payment cannot be made: from the
// payer's life total (CR 119.4), or at all because the window cancelled
// it — a player who can't lose life can't pay life (CR 119.8,
// CR 614.17b; #808). Callers validate the first before they start
// paying; this is the backstop.
func (g *Game) PayLifeForEffect(source, playerID uuid.UUID, amount int) error {
	return g.payLifeAsCostLocked(source, playerID, amount)
}

// DealDamageToPlayerForEffect writes amount damage to a player's
// life (via ChangeLife -amount) and emits an EventDealDamage. The
// damage is recorded before the life change so the event log reads
// "damage dealt → life changed." SBA check fires via the caller
// (effects run inside resolveTopOfStackLocked, which pairs with
// runStateChecks on the surrounding priority boundary).
//
// #711: CR 702.15b lifelink applies here. "Damage dealt by a source
// with lifelink also causes that source's controller to gain that much
// life" says nothing about combat, so a lifelinker's ping or Chandra's
// Ignition pays its controller exactly as a swing does. A source that
// is not a battlefield permanent (a spell, an emblem, uuid.Nil) has no
// lifelink to read, so Lightning Bolt still just deals 3.
//
// This is the FIRE-AND-FORGET form, and that is fine for it: a card
// that only says "deal 3 damage to target player" has nothing left to
// do. A caller that needs to know how much actually landed — "you gain
// life equal to the damage dealt this way" — must use
// DealDamageToPlayerThenForEffect or DealDamageEachThenForEffect
// instead, because a damage event can pause on a CR 616 ordering prompt
// and reading the life total back on the next line reads it before the
// prompt is answered (#807).
func (g *Game) DealDamageToPlayerForEffect(source, playerID uuid.UUID, amount int) error {
	return g.DealDamageToPlayerThenForEffect(source, playerID, amount, nil)
}

// DealDamageToPlayerThenForEffect is damage to a player with a
// CONTINUATION: `then` runs with the amount that actually landed, once
// the CR 614 window has settled it. Lock-free.
//
// The damage sibling of ChangePlayerLifeThenForEffect, and deliberately
// the same idiom — card authors learn `...ThenForEffect` once:
//
//	g.DealDamageToPlayerThenForEffect(src, opp, n, func(g *game.Game, dealt int) error {
//	        return g.ChangePlayerLifeForEffect(src, me, dealt)
//	})
//
// for "~ deals N damage to target opponent. You gain life equal to the
// damage dealt this way", and DealDamageEachThenForEffect for the
// each-opponent version of the same sentence.
//
// WHAT `then` RECEIVES. The post-replacement amount, non-negative: 6
// for a Lightning Bolt under Angrath's Marauders, 2 for one under a
// "prevent 1" shield. Zero means nothing landed — fully prevented,
// replaced away (CR 614.10, a Fog), or the player has left the game.
// `then` is told either way, because a caller adding up what several
// opponents took would otherwise wait forever on the one that took
// nothing.
//
// WHEN IT RUNS. Immediately, on the ordinary path, before this function
// returns. After the affected player answers the CR 616 ordering prompt
// when two DIFFERENT damage replacements apply to one event — the case
// #807 exists for, and the reason `then` is a continuation rather than
// a return value. Either way it runs with g.mu held and with the
// event's own EventDealDamage already emitted, so it may start the next
// damage event or queue the next prompt itself.
//
// Returns ErrPlayerNotFound without running `then` when the target is
// not seated at all, exactly as ChangePlayerLifeThenForEffect does. A
// seated player who has been eliminated is not an error: nothing is
// dealt and `then` runs with zero (CR 800.4a, #808). The batch form
// skips both rather than failing on them.
func (g *Game) DealDamageToPlayerThenForEffect(source, playerID uuid.UUID, amount int, then func(g *Game, dealt int) error) error {
	return g.dealDamageToPlayerLocked(source, nil, playerID, amount, then)
}

// dealDamageToPlayerLocked is DealDamageToPlayerThenForEffect's body,
// with the source OBJECT when the caller knows it (#1396,
// DealDamageFromObjectForEffect). obj only changes where the tail reads
// the source's lifelink from; see effectDamageTailLocked.
func (g *Game) dealDamageToPlayerLocked(source uuid.UUID, obj *ObjectRef, playerID uuid.UUID, amount int, then func(g *Game, dealt int) error) error {
	if amount <= 0 {
		// Not an event at all — "deals 0 damage" deals no damage
		// (CR 120.8) and fires no window. The continuation is still an
		// outcome, and it is zero. Checked before the seat, so a
		// zero-damage deal at a player who has left stays the no-op it
		// has always been rather than becoming an error.
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		// #808, CR 800.4a: no damage reaches a player who has left the
		// game. A terminal outcome with nothing dealt, like the
		// zero-damage case above.
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	// S22: route through the CR 614 replacement pipeline, the way
	// combat damage to a player already did. Before this, damage
	// dealt to a PLAYER by a spell or ability skipped replacements
	// entirely — so a damage doubler (Angrath's Marauders) or a
	// prevention shield could never see a Lightning Bolt, only
	// combat damage and damage marked on creatures. The three other
	// damage entry points were already routed; this was the hole.
	ev := &ReplacementEvent{
		Kind:         RepEventDamage,
		Source:       source,
		DamageSource: source,
		DamageTarget: playerID,
		DamageAmount: amount,
		// #694: the tail goes on BEFORE the pipeline runs, because a
		// CR 616 ordering prompt returns below without landing the
		// damage and the resume has nothing else to go on. #807: the
		// caller's continuation rides the same tail for the same
		// reason.
		//
		// #711: it carries the source's CR 702.15b lifelink, read
		// here rather than when the damage lands, so a prompt
		// answered after the source has left still credits the life
		// it dealt. Still no actor and no CR 903.10a commander tally:
		// those are combat-damage business.
		damageTail: g.effectDamageTailLocked(damageTailPlayer, source, obj),
	}
	ev.damageTail.then = then
	_, err := g.damageThroughReplacementsLocked(ev)
	return err
}

// DealDamageToCreatureForEffect marks amount damage on a
// battlefield creature. Emits EventDealDamage. The SBA pass fires
// on the surrounding priority boundary (resolution path already
// bookends with runStateChecks), so lethal damage routes the card
// via the normal SBA loop rather than a bespoke kill-now path.
//
// S30: routed through the CR 614 replacement pipeline. This was the
// last unrouted damage entry point — the sibling comment on
// DealDamageToPlayerForEffect claims the other three were already
// covered, and it was right about three of them. A Lightning Bolt
// aimed at a CREATURE reached DamageMarked directly, so neither a
// prevention shield nor a damage doubler could see it, while the
// same Bolt aimed at a player went through the pipeline. That
// asymmetry is invisible until a card exists that cares, and S30's
// prevention shields are that card.
// DealDamageToCreatureForEffect deals amount damage to a battlefield
// PERMANENT. Emits EventDealDamage. The SBA pass fires on the
// surrounding priority boundary (the resolution path already bookends
// with runStateChecks), so lethal damage routes the card via the
// normal SBA loop rather than a bespoke kill-now path.
//
// The name says "creature" for history's sake and is now a
// misnomer — every damage-dealing card in the catalog calls it, and
// since S27 (#406) the target may equally be a planeswalker or a
// battle. What the damage does is decided by CR 120.3 in
// applyDamageToPermanentLocked: marked on a creature, loyalty off a
// planeswalker, defense off a battle, and all of those at once for a
// permanent that is more than one of them. Before that split, a
// Lightning Bolt aimed at a planeswalker incremented a number nothing
// read — a two-mana no-op that looked like it had worked.
//
// #711: deathtouch and lifelink ARE applied here. CR 702.2b and
// CR 702.15b are about the source, not about combat — a fight between
// a Wurmcoil Engine and anything is lethal and gains eight life, and a
// Basilisk Collar makes a one-damage ping lethal. The keywords ride
// the damageTail, so the paused CR 616 path applies them identically;
// a source that is not a battlefield permanent (a spell, an emblem,
// uuid.Nil) has neither.
//
// The FIRE-AND-FORGET form. Use DealDamageToCreatureThenForEffect when
// the card goes on to say something about how much was dealt (#807).
func (g *Game) DealDamageToCreatureForEffect(source, cardID uuid.UUID, amount int) error {
	return g.DealDamageToCreatureThenForEffect(source, cardID, amount, nil)
}

// DealDamageToCreatureThenForEffect is damage to a battlefield
// permanent with a CONTINUATION — the permanent-target sibling of
// DealDamageToPlayerThenForEffect, with the identical contract: `then`
// runs once, with the post-replacement amount, and with 0 when the
// damage was fully prevented, replaced away, or the permanent left
// between a CR 616 prompt and its answer.
//
// "~ deals damage equal to its power to target creature. You gain that
// much life" is the sentence this exists for; a card that only deals
// the damage keeps using DealDamageToCreatureForEffect.
func (g *Game) DealDamageToCreatureThenForEffect(source, cardID uuid.UUID, amount int, then func(g *Game, dealt int) error) error {
	return g.dealDamageToPermanentLocked(source, nil, cardID, amount, then)
}

// dealDamageToPermanentLocked is DealDamageToCreatureThenForEffect's
// body, with the source OBJECT when the caller knows it (#1396). The
// sibling of dealDamageToPlayerLocked.
func (g *Game) dealDamageToPermanentLocked(source uuid.UUID, obj *ObjectRef, cardID uuid.UUID, amount int, then func(g *Game, dealt int) error) error {
	if amount <= 0 {
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	ev := &ReplacementEvent{
		Kind:         RepEventDamage,
		Source:       source,
		DamageSource: source,
		DamageTarget: cardID,
		DamageAmount: amount,
		// #694: the tail goes on BEFORE the pipeline runs, because a
		// CR 616 ordering prompt returns below without landing the
		// damage and the resume has nothing else to go on. #807: the
		// caller's continuation rides the same tail for the same
		// reason.
		//
		// #711: it carries the source's CR 702.2b deathtouch and
		// CR 702.15b lifelink, snapshotted here so a prompt answered
		// after the source has died still applies what it dealt with.
		damageTail: g.effectDamageTailLocked(damageTailPermanent, source, obj),
	}
	ev.damageTail.then = then
	// Through the permanent-aware tail: a creature marks damage, a
	// planeswalker loses loyalty (CR 120.3c, the #406 fix) and a
	// battle loses defence. The old inline loop here only ever
	// incremented DamageMarked, which is why damage could not kill a
	// planeswalker.
	_, err := g.damageThroughReplacementsLocked(ev)
	return err
}

// DealDamageFromObjectForEffect deals damage from a named permanent
// OBJECT (#1396, CR 400.7 / 608.2h) to a player or to a battlefield
// permanent — the target is dispatched exactly as the effects package's
// DealDamage primitive does, and a target that is neither deals nothing.
//
// The difference from the instance-ID entry points is only where the
// source's lifelink and deathtouch come from. While `source` is still
// on the battlefield as that object, they are read live. Once it has
// left, they are read off that object's last-known record — even if
// the card has since come back, because the permanent on the
// battlefield then is a new object that dealt nothing. Warstorm Surge
// is the case: the creature that entered is blinked in response, and
// the damage is the departed creature's, with the departed creature's
// keywords.
//
// Use it whenever the damage source is an object the effect already
// names by ref (effects.Context.Trigger().Object.Ref()). An instance ID
// alone keeps using DealDamageToPlayerForEffect /
// DealDamageToCreatureForEffect, which read the most recent departed
// object instead (departedDamageSourceLocked).
//
// Caller must hold g.mu in write mode.
func (g *Game) DealDamageFromObjectForEffect(source ObjectRef, target uuid.UUID, amount int) error {
	if g.playerByIDLocked(target) != nil {
		return g.dealDamageToPlayerLocked(source.ID, &source, target, amount, nil)
	}
	if findBattlefieldCard(g, target) != nil {
		return g.dealDamageToPermanentLocked(source.ID, &source, target, amount, nil)
	}
	return nil
}

// DealDamageEachThenForEffect is the batch form: `source` deals
// `amount` damage to each of `targets`, and then `then` runs with the
// TOTAL actually dealt — "this creature deals 1 damage to each
// opponent. You gain life equal to the damage dealt this way" (Creeping
// Bloodsucker). Lock-free.
//
// The total is the sum of what each target really took after its own
// replacements, not amount × len(targets): a target under a damage
// doubler takes more, one behind a prevention shield takes less or
// nothing, and the gain follows.
//
// Each target is damaged as the kind of thing it is, the way the
// DealDamage primitive already routes: a seated player takes life loss
// through the CR 120.3 player path, a battlefield permanent takes the
// CR 120.3 permanent split, and anything that is neither any more is
// skipped exactly as a fire-and-forget loop would skip it.
//
// Built on the single forms rather than beside them, so there is one
// place that knows how to wait rather than one per card. The targets
// are damaged IN SEQUENCE, each from the previous one's continuation:
// that is what makes the running total a plain value carried forward
// instead of a shared accumulator, which is what makes an undo across
// the prompt land where a clean run would.
//
// The sequencing is observable only when a leg pauses: with two
// different damage replacements on the first opponent, the second
// opponent's damage happens when the CR 616 prompt is answered rather
// than before it. The damage is simultaneous in the rules (CR 101.4)
// and sequential in this engine either way; doing the rest on the far
// side of the prompt is the closer of the two, and it is the only one
// that can report a true total. Same trade LoseLifeEachThenForEffect
// makes, for the same reason.
func (g *Game) DealDamageEachThenForEffect(source uuid.UUID, targets []uuid.UUID, amount int, then func(g *Game, totalDealt int) error) error {
	return g.dealDamageEachStepLocked(source, targets, amount, 0, then)
}

// dealDamageEachStepLocked damages the head of `targets` and continues
// with the tail, carrying the running total forward by value. The empty
// list is the base case: the batch is done and `then` gets the total.
//
// Caller must hold g.mu.
func (g *Game) dealDamageEachStepLocked(source uuid.UUID, targets []uuid.UUID, amount, dealtSoFar int, then func(g *Game, totalDealt int) error) error {
	for len(targets) > 0 {
		next, rest := targets[0], targets[1:]
		step := func(g *Game, dealt int) error {
			return g.dealDamageEachStepLocked(source, rest, amount, dealtSoFar+dealt, then)
		}
		if p := g.playerByIDLocked(next); p != nil {
			if p.Eliminated {
				// #808, CR 800.4a: a player who has left the game is
				// skipped, not dealt zero through a pipeline run.
				targets = rest
				continue
			}
			return g.DealDamageToPlayerThenForEffect(source, next, amount, step)
		}
		if findBattlefieldCard(g, next) != nil {
			return g.DealDamageToCreatureThenForEffect(source, next, amount, step)
		}
		// Neither a seated player nor a battlefield permanent: the
		// target is gone. Skip it and keep the total.
		targets = rest
	}
	if then == nil {
		return nil
	}
	return then(g, dealtSoFar)
}

// DrawNForEffect draws n cards for the given player, emitting one
// EventDrawCard per card (drawCardLocked already emits). Returns
// an error only on the first failure — partial draws are allowed
// (ErrZoneEmpty on the Nth card sets AttemptedEmptyDraw for the loss
// on the next SBA pass, which is already the drawCardLocked
// behaviour).
func (g *Game) DrawNForEffect(playerID uuid.UUID, n int) error {
	for i := 0; i < n; i++ {
		if err := g.drawCardLocked(playerID); err != nil {
			if err == ErrZoneEmpty {
				// Flag set, stop drawing. SBA loop will handle the loss.
				return nil
			}
			return err
		}
	}
	return nil
}

// DiscardRandomForEffect discards n cards from playerID's hand,
// chosen at random (CR 701.8b). Emits one EventDiscardCard per card
// and no EventZoneMove, through the one discard path (discard.go).
// If the hand has fewer than n cards, discards all of them.
//
// The cards are one pick on the discarding player's "pick" stream
// (ADR 0054 Decision 5), so an undone random discard redoes with the
// same cards. There is no "first card when the game has no RNG"
// branch any more: every game draws from a key (rng.go).
//
// THE FIRE-AND-FORGET FORM. A discard is an exit (ADR 0013 §5g) and a
// discarded commander's CR 903.9 prompt pauses the batch, so nil means
// "no error", never "the cards are in the graveyard". A card with a
// clause hanging off the discard uses DiscardRandomThenForEffect.
func (g *Game) DiscardRandomForEffect(playerID uuid.UUID, n int) error {
	return g.DiscardRandomThenForEffect(playerID, n, nil)
}

// DiscardRandomThenForEffect is DiscardRandomForEffect with the rest
// of the card attached — "discard a card at random. If it was a land
// card, …" — run once the whole batch has reached a terminal outcome
// and handed the cards that were really discarded
// (discardedThisWayLocked).
//
// It is the random discard's answer to the payout lint's `discardVerb`
// row (#1027, ADR 0013 §5y). `then` runs even when nothing was
// discarded, including for an empty hand: a continuation is the rest
// of a card, and one that is silently never called is a card that
// stops halfway (#544, #1006).
//
// Caller must hold g.mu.
func (g *Game) DiscardRandomThenForEffect(playerID uuid.UUID, n int, then func(g *Game, discarded []uuid.UUID) error) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if n <= 0 || p.Hand.Size() == 0 {
		if then == nil {
			return nil
		}
		return then(g, nil)
	}
	ids := make([]uuid.UUID, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		ids[i] = c.InstanceID
	}
	picked := g.ChooseAtRandomForEffect(RandomDraw{Player: playerID}, ids, n)
	return g.discardCardsLocked(playerID, picked, discardOptions{cause: DiscardCauseEffect, then: then})
}

// MillNForEffect moves n cards from the top of playerID's library
// to their graveyard. Emits EventMill per card. A library holding
// fewer than n mills what it has (CR 701.17b) and nobody loses for
// it — see MillToZoneForEffect.
func (g *Game) MillNForEffect(playerID uuid.UUID, n int) error {
	_, err := g.MillToZoneForEffect(playerID, n, ZoneGraveyard)
	return err
}

// MillToZoneForEffect is the general form: move cards off the top of
// playerID's library into `dest`, which is ZoneGraveyard (an ordinary
// mill) or ZoneExile ("exile the top N cards of your library").
//
// Returns the cards that LANDED in `dest`, in the order they came off
// the top. Callers that need to act on them — "exile all cards milled
// this way", "you may cast one of them this turn", a payoff that
// counts them — read the slice rather than diffing zones, which is the
// difference between this and MillNForEffect.
//
// The FIRE-AND-FORGET form, and since #893 its slice means what
// ExileCardsForEffect's count has meant since #866: the CR 400.7
// arrived-object reading (landedInZoneLocked). A card a replacement
// sent somewhere else is not in it — a commander whose owner took
// CR 903.9's offer went to the command zone rather than to a
// graveyard, and "if a card would be put into a graveyard from
// anywhere, exile it instead" moved the card to exile — and neither is
// a leg the CR 614 window cancelled. A leg that merely PAUSED on the
// CR 903.9 prompt cannot be in it either, because nothing has moved
// yet; that is what MillToZoneThenForEffect is for, and a caller that
// reads the slice should use it.
//
// There is NO `until` clause on this form, and #1161 took the
// parameter away rather than documenting a rule about it. A run that
// ends on what LANDED has to wait for each leg to land, and this form
// is the one that does not wait: a leg paused on the CR 903.9 prompt
// has not arrived when the loop asks, so the run walks past it. With
// a "stop after the first creature card" that costs one card; with
// Helm of Obedience's "or X cards", where the bound itself counts
// arrivals (#1161), it costs the whole library. So the clause lives
// on MillToZoneThenForEffect alone, which sequences, and the bad
// combination is not a rule the mill enforces — it is a function
// signature that cannot spell it. #1176 kept that property and
// strengthened it: a run is a loop over repetitions now, so no routing
// loop can be handed an early stop either.
//
// Running the library out stops the run, with no error and NO loss.
// CR 701.17b: a player instructed to mill more cards than their
// library holds "mill[s] as many as possible", and only an attempt to
// DRAW from an empty library loses the game (CR 704.5b, CR 121.4).
// The same holds for "exile the top N cards" and for the Then form's
// `until` run that never finds its card — both simply end when the
// library does, exactly as ExileTopFaceDownForEffect stops on an
// empty library.
// (#767: this used to set the empty-draw flag, then named
// LosesAtNextSBA, so Glimpse the Unthinkable on a nine-card library
// eliminated its target.)
//
// Mill COSTS are the other half of CR 701.17b — "can't pay a cost
// that includes milling a number of cards greater than the number of
// cards in their library" — and they are not this function's
// business: the engine has no mill cost component. The one catalog
// card with a mill-a-card cost, Millikin, gates its activation on a
// non-empty library itself; The Warring Triad declares the gap.
//
// The error is the up-front one — an unseated player, a destination
// that is neither a graveyard nor exile. A leg that cannot move is
// skipped by the batch body rather than failing the mill.
//
// #569: the AMOUNT is replaceable. A mill into a graveyard opens a
// RepEventMill window on n before any card moves, so Bruvac the
// Grandiloquent doubles the instruction; an exile of the top N is not a
// mill (CR 701.17a) and opens none. That window can PAUSE, on a CR 616
// ordering prompt between two amount replacements, and then this
// returns an EMPTY slice with the mill still owed — the contract
// CreateTokensForEffect's empty ID slice already carries, and the
// reason a caller that reads the list should use the Then form.
//
// Caller must hold g.mu.
func (g *Game) MillToZoneForEffect(playerID uuid.UUID, n int, dest ZoneKind) ([]uuid.UUID, error) {
	return g.millThroughReplacementsLocked(playerID, n, dest, nil, nil)
}

// MillToZoneThenForEffect is the CONTINUATION form: mill exactly what
// MillToZoneForEffect would and hand `then` the cards that reached
// `dest` — "for each creature card put into your graveyard this way",
// "for each card of the chosen color exiled this way" (Oona).
//
// #893, and the same pair ExileCardsThenForEffect / ExileCardsForEffect
// are two halves of (ADR 0013 §5k). Any leg can pause on the CR 903.9
// prompt, so what was milled is not knowable on the line after the
// mill: a commander on top of the library queues its owner's question
// and moves nowhere until they answer it. The legs therefore go IN
// SEQUENCE, each from the previous one's continuation, with the landed
// list carried forward BY VALUE — the property that makes an undo
// across the prompt replay identically.
//
// What "this way" means is CR 400.7's, landedInZoneLocked's: the
// object that ARRIVED in `dest`. A commander that took the command
// zone was not milled, and neither was a card an "exile it instead"
// replacement rewrote on the way to a graveyard.
//
// The cost of waiting, which the fire-and-forget form does not pay:
// the rest of the mill happens when the prompt is answered rather than
// immediately. That is deliberate and it is the only version that can
// report a true list — the same trade DestroyPermanentsThenForEffect
// and the discard batch already make.
//
// #569 adds a second thing that can pause, and it pauses EARLIER: the
// CR 614 window on the amount, before any card is chosen. `then` runs
// from wherever the mill ends either way — inline, from the CR 903.9
// resume of one leg, or from the amount window's resume — and it runs
// with an empty list when the mill was replaced away entirely, because
// a caller sequencing work behind it has to be told even then.
//
// # The `until` clause lives HERE, and only here (#1161)
//
// `until`, when non-nil, makes this a RUN rather than an instruction
// (#1176, millUntilRunLocked): "mills a card, then repeats this
// process until …" is a loop over one-card mill instructions, each
// with its own CR 614 window on its own amount, and the clause is
// asked with everything that has LANDED in `dest` so far after each
// REPETITION. The run ends on the first list it accepts, and the cards
// of the repetition that ended it all move.
//
// That is what makes a mill-amount replacement double the right thing.
// Bruvac the Grandiloquent doubles a MILL (CR 701.17b) and each
// repetition is one, so Helm of Obedience at X=3 with Bruvac out mills
// 2 + 2 = four cards and overshoots its bound by one, exactly as it
// does in paper. It does NOT double the bound: "until a creature card
// or X cards have been put into their graveyard this way" is a clause
// about arrivals, not a number the instruction names, and #1161's
// TestBruvacDoublesAMillAmountAndNotHelmsBound still pins that.
//
// `n` with `until` caps the number of REPETITIONS; 0 — every card in
// the catalog — is "no limit but the library". A bound a card prints
// is a clause (effects.UntilCount), not this.
//
// Why the clause is on the sequencing form and nowhere else: it is
// answered about cards that have ARRIVED, and only this form waits for
// them. The fire-and-forget loop walks past a leg paused on CR 903.9,
// which with a landed-count bound would mill the whole library while
// the prompts piled up. MillToZoneForEffect therefore takes no `until`
// at all — the combination is unspellable rather than guarded against.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) MillToZoneThenForEffect(
	playerID uuid.UUID,
	n int,
	dest ZoneKind,
	until func([]Card) bool,
	then func(g *Game, milled []uuid.UUID) error,
) error {
	if then == nil {
		// The tail's nil-ness is what picks the form inside, so a caller
		// that passes nil here would silently get the fire-and-forget
		// one. Refusing is not an option — a mill of nothing is still a
		// legal mill — so give it a continuation that does nothing.
		then = func(*Game, []uuid.UUID) error { return nil }
	}
	_, err := g.millThroughReplacementsLocked(playerID, n, dest, until, then)
	return err
}

// millPlanLocked validates a mill and chooses the cards it MAY move,
// top of the library first. Shared by both forms above, so they cannot
// disagree about what a mill of n is.
//
// #529: the cards that will move are chosen UP FRONT, top-down, and
// addressed by ID from there — rather than by repeatedly popping
// whatever is on top.
//
// A milled commander gets the CR 903.9 prompt, and a queued prompt
// leaves that card exactly where it was: still on top of the library.
// Re-reading the top each iteration would hand back the same commander
// every time and mill nothing else. Choosing the set first lets the
// fire-and-forget mill proceed AROUND the paused card, which is both
// what CR 701.17a's simultaneous mill wants and the only version that
// does not silently shorten the mill when a commander is in the way.
//
// It takes no `until` and never did anything useful with one. #1159
// moved the clause out of here because it was being answered against
// the cards that came OFF THE LIBRARY rather than the ones that
// ARRIVED (CR 400.7), and #1176 moved it out of the routing loop too:
// a run is a sequence of one-card mill instructions now, so the only
// number this function ever sees is the amount of ONE instruction,
// after its CR 614 window has settled.
//
// The returned cards are the pre-move copies, in library order (top
// first).
//
// Caller must hold g.mu.
func (g *Game) millPlanLocked(playerID uuid.UUID, n int, dest ZoneKind) ([]Card, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	switch dest {
	case ZoneGraveyard, ZoneExile:
	default:
		return nil, ErrInvalidParam
	}
	avail := len(p.Library.Cards)
	want := n
	if want > avail {
		want = avail
	}
	plan := make([]Card, 0, want)
	for i := 0; i < want; i++ {
		plan = append(plan, p.Library.Cards[avail-1-i])
	}
	return plan, nil
}

// DestroyPermanentForEffect destroys a battlefield permanent
// (CR 701.8), routing it to its owner's graveyard — or exile if the
// owner is no longer seated. Exposed so effect primitives can call
// it from an already-locked context.
//
// S25 (#77): this is the catalog's destruction verb, so it is where
// indestructible is honoured. A permanent with indestructible is
// left exactly where it is and nil is returned — see
// indestructible.go for why the check cannot live one level down in
// routeBattlefieldCardToOwnerGraveyardLocked, which sacrifice and
// the zero-counter SBAs share.
//
// #667: it is also where "it can't be regenerated" (CR 701.19c) goes.
// Pass DestroyOptions{CantBeRegenerated: true} — Damnation, Day of
// Judgment, Mortify, Putrefy, Pongify and the rest of the family — and
// the flag rides the destroy route onto the CR 614 event, where the
// regeneration built-in reads it and declines (regeneration.go). The
// rider is variadic so the hundred-odd plain "destroy this" calls stay
// as they were; at most one is meaningful.
func (g *Game) DestroyPermanentForEffect(cardID uuid.UUID, opts ...DestroyOptions) error {
	return g.destroyBattlefieldPermanentLocked(cardID, firstDestroyOptions(opts))
}

// SacrificePermanentForEffect sacrifices a battlefield permanent on
// behalf of its controller (CR 701.21): EventSacrifice fires while
// the card is still on the battlefield, then it takes the ordinary
// route to its owner's graveyard (emitting ZoneMove + LTB, so
// dies-triggers see it and the CR 903.9 commander-zone replacement
// still gets its say).
//
// Sacrifice is not destruction — no indestructible / regeneration
// check applies, which is why this doesn't reuse the destroy path's
// naming. A card that isn't on the battlefield returns
// ErrCardNotFound and emits nothing.
//
// Caller must hold g.mu. Added in S21 sub-PR 1.
func (g *Game) SacrificePermanentForEffect(cardID uuid.UUID) error {
	return g.sacrificePermanentLocked(cardID)
}

// ExileCardForEffect moves a card from whatever zone it's in to
// the shared exile zone. The source zone is found by scanning; if
// the card is already in exile, the call is a no-op.
//
// #529: routed through the shared exit primitive, so exile — a CR
// 903.9 destination — offers a commander's owner the command zone.
// That is #372 (Airbend exiles a commander with no prompt). When the
// prompt is queued nothing has moved yet and this returns nil; the
// move completes when the owner answers.
//
// Which is why this is the FIRE-AND-FORGET form: nil means "no error",
// never "it is in exile". A caller with an "if you do" or a "for each
// card exiled this way" hanging off the move uses
// ExileCardThenForEffect (one card) or ExileCardsThenForEffect
// (several), which wait for the answer and report what landed (#870).
func (g *Game) ExileCardForEffect(cardID uuid.UUID) error {
	_, err := g.routeCardToZoneLocked(zoneRoute{CardID: cardID, Dst: ZoneExile})
	return err
}

// ExileTopFaceDownForEffect exiles the top n cards of playerID's
// library FACE DOWN (CR 406.3) and returns them in the order they
// left the library. Necropotence's "exile the top card of your
// library face down" is the card this exists for.
//
// The difference from every other exile in the engine is the one
// line that is missing: there is no markCardKnownInZoneLocked call.
// Exile is a public zone, so the ordinary path marks every seat a
// knower and the wire ships the card's name to the whole table. A
// card exiled face down is one no player may look at — including
// the player who exiled it — so its knowledge set is set to the
// FaceDownExiled row of ADR 0069's viewers table, which is nobody,
// and Card.FaceDown is set. (FaceDownForetold is the row foretell
// uses, and its answer is the owner; the route takes the kind so the
// two cannot be told apart by anything but the rule.) An empty
// knowledge set is what
// the wire keys on: protocol.redactCardForViewer strips every
// identifying field (name, costs, abilities, catalog flags) for a
// non-knower. Card.FaceDown is what the client keys on to draw a
// card back for a face-down card the viewer does not know (#95).
//
// Clearing rather than leaving the set alone matters: a library card
// is not always unknown. A player who has just scryed or used
// Sensei's Divining Top is a knower of their own top card, and
// carrying that marking into exile would let exactly one seat read a
// card the rules say nobody can.
//
// Exiling off an empty library is not an error and is not a draw —
// it moves nothing and does NOT set AttemptedEmptyDraw. Necropotence
// with an empty library charges the life and exiles nothing, which
// is why its controller does not lose on the spot.
//
// Caller must hold g.mu. Added in S22.
func (g *Game) ExileTopFaceDownForEffect(playerID uuid.UUID, n int) ([]uuid.UUID, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	if g.Exile == nil || n <= 0 {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		if p.Library.Size() == 0 {
			return out, nil
		}
		// The library's top is the LAST element — the same read
		// PopTop and ExileTopWithPermissionForEffect use.
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		// #529: through the shared exit primitive so a commander
		// exiled off the top of its owner's library still gets the CR
		// 903.9 choice. A queued prompt leaves the card ON the
		// library, so the loop has to stop rather than read the same
		// top card again — see MillToZoneForEffect for the same
		// hazard handled without losing the rest of the batch. Here
		// the batch is Necropotence-shaped (n is 1 in every printed
		// case), so stopping costs nothing worth the machinery.
		paused, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:   top,
			Dst:      ZoneExile,
			Actor:    playerID,
			FaceDown: FaceDownExiled,
		})
		if err != nil {
			return out, err
		}
		if paused {
			return out, nil
		}
		out = append(out, top)
	}
	return out, nil
}

// BounceToHandForEffect moves a card from wherever it is to its
// owner's hand. Used by Unsummon and similar. If the owner is no
// longer seated, the call returns ErrPlayerNotFound without moving
// the card.
//
// #529: routed through the shared exit primitive, so the hand — a CR
// 903.9 destination — offers a commander's owner the command zone.
// The CR 400.7 face-down clear the hand owes happens down there, for
// every destination rather than just this one.
func (g *Game) BounceToHandForEffect(cardID uuid.UUID) error {
	_, err := g.routeCardToZoneLocked(zoneRoute{CardID: cardID, Dst: ZoneHand})
	return err
}

// TuckOptions is WHERE on the library a tuck lands. The zero value is
// the top, which is what "put it on top of its owner's library" and the
// bare "shuffles it into their library" (tuck, then shuffle) both want.
//
// It is one struct rather than two positional arguments because the
// position is a property of the ROUTE — it rides across a CR 903.9
// pause and is applied against the SETTLED destination — and because a
// third position (Depth) already existed on its own entry point. The
// `Then` forms take it once and the fire-and-forget forms build it from
// their older positional arguments, so there is one description of a
// tuck and two ways to ask for one.
//
// ToBottom wins if both are set, matching zoneRoute.
type TuckOptions struct {
	// ToBottom sends the card to the BOTTOM of the library
	// (Condemn, Mistveil Plains).
	ToBottom bool

	// Depth places the card N cards down from the top — the
	// God-Eternals' and Teferi's "third from the top" is 3. Zero and
	// 1 both mean the top. A library shorter than the depth takes the
	// card on the bottom.
	Depth int
}

// TuckToLibraryForEffect moves a card from wherever it is onto its
// owner's library — the top (`toBottom` false) or the bottom.
//
// BounceToHandForEffect with a different destination, and it exists
// for the same reason: "put it on top of its owner's library" is a
// printed instruction (Sensei's Divining Top putting itself back,
// Condemn, Hinder) that no other helper can express. MoveCard always
// pushes to the top of its destination, so the bottom case reorders
// afterwards rather than having its own path.
//
// A card leaving the battlefield emits LKI + EventLTB, so dies- and
// leaves-triggers see it; a card coming from anywhere else emits the
// zone move alone. Library cards have no knowers, so this CLEARS the
// card's knowledge set on the way in — a permanent everyone could
// read becomes a face-down card in a hidden zone, and leaving the
// owner marked would let them see their own top card forever.
//
// Caller must hold g.mu. Added in S22.
//
// #529: routed through the shared exit primitive, so the library — a
// CR 903.9 destination — offers a commander's owner the command zone.
// `toBottom` rides along on the route rather than being applied here,
// which is what lets it survive a queued prompt: a commander tucked
// to the bottom whose owner declines still lands on the bottom.
//
// #783: this is the FIRE-AND-FORGET half of the pair, on the same
// route template (tuckRoute) as TuckToLibraryThenForEffect. It returns
// nil whether the card moved or a prompt was queued, which is fine for
// a caller with nothing left to do — Sensei's Divining Top putting
// itself back is the last instruction on its ability — and wrong for
// any caller that reads the card's zone, counts what arrived or asks
// the next question. Those use the `Then` form. The same split the
// mill has, for the same reason.
func (g *Game) TuckToLibraryForEffect(cardID uuid.UUID, toBottom bool) error {
	r := tuckRoute(TuckOptions{ToBottom: toBottom})
	r.CardID = cardID
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// TuckToLibraryAtDepthForEffect is TuckToLibraryForEffect with the
// printed position spelled out: depth 1 is the top, depth 3 is
// "into its owner's library third from the top" (Teferi, Hero of
// Dominaria's −3), and a library shorter than the depth takes the
// card on the bottom.
//
// Everything TuckToLibraryForEffect's comment says applies here —
// the LKI + EventLTB on the way off the battlefield, the cleared
// knowledge set on the way into a hidden zone, and the CR 903.9
// prompt that lets a commander's owner choose the command zone
// instead. The depth rides the route, so a commander whose owner
// declines still lands third from the top.
//
// Caller must hold g.mu.
func (g *Game) TuckToLibraryAtDepthForEffect(cardID uuid.UUID, depth int) error {
	r := tuckRoute(TuckOptions{Depth: depth})
	r.CardID = cardID
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// TapTargetForEffect taps a battlefield card. Wrapper around the
// internal path; emits EventTapCard.
func (g *Game) TapTargetForEffect(cardID uuid.UUID) error {
	return g.setTapStateLocked(cardID, true)
}

// UntapTargetForEffect untaps a battlefield card.
func (g *Game) UntapTargetForEffect(cardID uuid.UUID) error {
	return g.setTapStateLocked(cardID, false)
}

// setTapStateLocked is the internal helper behind TapCard and the
// TapTarget / UntapTarget effect primitives. Caller must hold g.mu.
//
// Both directions are a CHANGE of state (CR 701.26a / 701.26b): a
// permanent that is already tapped does not become tapped, and one
// that is already untapped does not become untapped. Neither emits,
// and neither is an error — "untap target permanent" pointed at an
// upright permanent is a legal instruction that does nothing. The
// guard is what keeps Mesmeric Orb from milling on a Voltaic Key
// pointed at an untapped rock, and Quest for Renewal from banking a
// counter for a tap that never happened.
//
// The untap leg routes through untap.go's primitive so that every
// untap in the engine — this one, the untap step's, a sandbox click
// — announces through the same line of code.
func (g *Game) setTapStateLocked(cardID uuid.UUID, tapped bool) error {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		if !tapped {
			g.untapPermanentLocked(c)
			return nil
		}
		if c.Tapped {
			return nil
		}
		c.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, CardID: cardID})
		return nil
	}
	return ErrCardNotFound
}

// CounterTargetForEffect counters a stack item. Discriminates
// between spell and ability via the stored StackItem.Kind and
// delegates to the appropriate internal helper. Spell counters
// route to the item's owner's graveyard by default.
func (g *Game) CounterTargetForEffect(stackID uuid.UUID) error {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	switch item.Kind {
	case StackItemSpell:
		// CR 701.6a + the "this spell can't be countered" rider
		// (Supreme Verdict). The spell is still a LEGAL TARGET — a
		// Counterspell aimed at it resolves, and then does nothing.
		// Modelling it as an illegal target would be the easy
		// mistake and the wrong one: the counterspell would fizzle
		// instead of resolving, which is observable to anything
		// watching it resolve. Added in S23.
		if g.spellCantBeCounteredLocked(stackID) {
			slog.Info("counter had no effect: spell can't be countered",
				"spell_id", stackID)
			return nil
		}
		return g.counterSpellLocked(stackID, nil)
	case StackItemActivated, StackItemTriggered:
		return g.counterAbilityLocked(stackID)
	default:
		return ErrCardNotOnStack
	}
}

// CounterTargetToZoneForEffect is CounterTargetForEffect with the
// destination exposed — #1230's "counter target spell; if that spell
// is countered this way, put it into its owner's hand / on top of its
// owner's library / exile it instead of into the graveyard" (Remand,
// Memory Lapse, Dissipate; Devious Cover-Up's exile clause in this
// catalog). `dst.Owner` left uuid.Nil means "its owner", which is what
// every one of those destinations prints.
//
// Same "can't be countered" gate as CounterTargetForEffect — a
// destination override does not change that this IS a counter
// (CR 701.6a), so a spell printed uncounterable is untouched by
// either. A countered ABILITY has no destination to redirect
// (CR 701.6b: it ceases to exist), so this behaves exactly like
// CounterTargetForEffect for one.
func (g *Game) CounterTargetToZoneForEffect(stackID uuid.UUID, dst ZoneRef) error {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	switch item.Kind {
	case StackItemSpell:
		if g.spellCantBeCounteredLocked(stackID) {
			slog.Info("counter had no effect: spell can't be countered",
				"spell_id", stackID)
			return nil
		}
		return g.counterSpellLocked(stackID, &dst)
	case StackItemActivated, StackItemTriggered:
		return g.counterAbilityLocked(stackID)
	default:
		return ErrCardNotOnStack
	}
}

// CounterSpellToLibraryThenForEffect is "counter target spell. If that
// spell is countered this way, put it on the bottom of / on top of its
// owner's library instead of into that player's graveyard" — Spell
// Crumple's bottom, and each leg of Hinder's "your choice of the top or
// bottom" once the choice is made (#1298, ADR 0088's 2026-09-23
// amendment). `at` is the position: ToBottom, or Depth from the top.
//
// It is a COUNTER (CR 701.6a) with a destination, so it has
// CounterTargetToZoneForEffect's gates: a spell that can't be countered
// is untouched, and a countered ability simply ceases to exist
// (CR 701.6b) — neither reaches a library. The move is the shared stack
// exit, so CR 903.9 offers a commander the command zone and
// flashback's CR 702.34a exile still wins.
//
// `then` hears whether the card is in its owner's library once the
// move settles (CR 400.7) — false for every outcome above, and for a
// commander whose owner took the command zone. Nil is allowed.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) CounterSpellToLibraryThenForEffect(stackID uuid.UUID, at TuckOptions, then func(g *Game, placed bool) error) error {
	tell := func(g *Game, placed bool) error {
		if then == nil {
			return nil
		}
		return then(g, placed)
	}
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil {
		return tell(g, false)
	}
	if item.Kind == StackItemActivated || item.Kind == StackItemTriggered {
		if err := g.counterAbilityLocked(stackID); err != nil {
			return err
		}
		return tell(g, false)
	}
	if !g.counterableSpellOnStackLocked(stackID) {
		if item.Kind == StackItemSpell {
			slog.Info("counter had no effect: spell can't be countered", "spell_id", stackID)
		}
		return tell(g, false)
	}
	return g.exitSpellFromStackAtLocked(stackID, &ZoneRef{Kind: ZoneLibrary, Owner: item.Owner}, ZoneGraveyard, true, at,
		func(g *Game) error {
			return tell(g, g.landedInZoneLocked(stackID, ZoneLibrary, uuid.Nil))
		})
}

// counterableSpellOnStackLocked reports whether `stackID` is a spell on
// the stack that a counter would actually counter — what "if that
// spell is countered this way" can be true of. The put_in_library
// prompt asks it before offering a Hinder'd spell's position, so a
// spell that can't be countered is never asked about.
//
// Caller must hold g.mu.
func (g *Game) counterableSpellOnStackLocked(stackID uuid.UUID) bool {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return false
	}
	if g.Stack == nil || !g.Stack.Contains(stackID) {
		return false
	}
	return !g.spellCantBeCounteredLocked(stackID)
}

// ReturnSpellToHandForEffect returns a spell on the stack to its
// owner's hand WITHOUT countering it: CR 701.6 never applies, so a
// spell printed "can't be countered" is unaffected and nothing
// watching "whenever a spell is countered" fires. Reprieve; the
// second half of Narset's Reversal, once CopySpell has made the copy.
//
// Shares counterSpellLocked's stack-exit body — exitSpellFromStackLocked
// — with `countered: false` and the owner's hand as the fallback
// destination instead of the graveyard, so S29's flashback override
// and the CR 614 / CR 903.9 window through routeCardToZoneLocked apply
// exactly as they do to a counter: a commander Reprieve'd back to hand
// still gets the CR 903.9 offer, and a flashed-back spell bounced this
// way is still exiled instead (CR 702.34a is unconditional about how
// the spell would otherwise leave the stack).
//
// Caller must hold g.mu.
func (g *Game) ReturnSpellToHandForEffect(stackID uuid.UUID) error {
	return g.exitSpellFromStackLocked(stackID, nil, ZoneHand, false, nil)
}

// ExileSpellForEffect exiles a spell from the stack WITHOUT countering
// it — "exile target spell" (#1318). CR 701.6 never applies, so a
// spell printed "can't be countered" is exiled all the same and
// nothing watching "whenever a spell is countered" fires. The third
// verb over exitSpellFromStackLocked, beside the counter and the
// return to hand, and it exists for the reason those two share one
// body: a spell leaving the stack owes the CR 903.9 window, the
// CR 608.2h last-known record and the retirement of its stack record,
// and three spellings of that drift.
//
// Before this there was no exported way to say it, so the one card
// that needed it — airbend, which targets "a creature or spell" —
// went through ExileCardForEffect, and that route did not retire the
// StackMeta entry. The spell's card went to exile, its record stayed,
// and the table could never resolve past it. The route now retires the
// record by source zone (zone_route.go), so both spellings are
// correct; this one is the one that SAYS spell, and refuses anything
// else.
//
// Returns ErrCardNotOnStack for an ID that is not a spell on the
// stack, like its siblings.
//
// Caller must hold g.mu.
func (g *Game) ExileSpellForEffect(stackID uuid.UUID) error {
	return g.exitSpellFromStackLocked(stackID, &ZoneRef{Kind: ZoneExile}, ZoneExile, false, nil)
}

// ExileSpellThenForEffect is ExileSpellForEffect with the rest of the
// instruction handed over rather than written on the next line: `then`
// is told whether the spell actually reached exile. Aven Interrupter's
// "exile target spell. It becomes plotted." is the caller — the plot
// is a property of the card in exile, and a commander spell whose
// owner takes CR 903.9's offer went to the command zone instead, so
// there is nothing to plot.
//
// `exiled` is CR 400.7's reading (landedInZoneLocked): true only when
// the card is in exile once the move settles. A spell that has already
// left the stack — countered or resolved in response — is not an error
// here: the move is "nothing happened" and `then` hears false, the
// same terminal outcome a cancelled route reports.
//
// Caller must hold g.mu.
func (g *Game) ExileSpellThenForEffect(stackID uuid.UUID, then func(g *Game, exiled bool) error) error {
	tell := func(g *Game, exiled bool) error {
		if then == nil {
			return nil
		}
		return then(g, exiled)
	}
	if item, ok := g.StackMeta[stackID]; !ok || item == nil || item.Kind != StackItemSpell ||
		g.Stack == nil || !g.Stack.Contains(stackID) {
		return tell(g, false)
	}
	return g.exitSpellFromStackLocked(stackID, &ZoneRef{Kind: ZoneExile}, ZoneExile, false, func(g *Game) error {
		return tell(g, g.landedInZoneLocked(stackID, ZoneExile, uuid.Nil))
	})
}

// counterSpellLocked is the lock-free body of CounterSpell. Caller
// must hold g.mu.
//
// #529 folded the two into one body. They had drifted: the public
// CounterSpell honoured flashback's CR 702.34a "exile this card
// instead of putting it anywhere else any time it would leave the
// stack" and this one did not, so a flashed-back spell answered by a
// CATALOG counterspell went to the graveyard while the same spell
// answered from the admin action was exiled. One body, one answer.
//
// The move goes through the shared exit primitive, so a countered
// commander gets the CR 903.9 command-zone choice — #364. When the
// prompt is queued nothing has moved, the stack item is still
// registered, and both complete when the owner answers.
func (g *Game) counterSpellLocked(spellID uuid.UUID, dst *ZoneRef) error {
	return g.exitSpellFromStackLocked(spellID, dst, ZoneGraveyard, true, nil)
}

// exitSpellFromStackLocked is the shared body behind counterSpellLocked
// and ReturnSpellToHandForEffect (#1230): CR 701.6a's counter and a
// plain "return target spell to its owner's hand" both remove a spell
// from the stack without letting it resolve, through the same
// destination resolution and the same routeCardToZoneLocked call —
// they differ only in the destination and in whether the move counts
// as a COUNTER for CR 701.6-keyed triggers and for "can't be
// countered" (checked by the caller, not here).
//
// `fallback` is the destination when the caller names none (dst ==
// nil) and no S29 override applies — the owner's graveyard for a
// counter, the owner's hand for a bounce. `countered` sets
// zoneRoute.Countered, which is what actually decides EventCounterSpell
// vs. an ordinary zone move; it is the one bit the two callers
// disagree about.
//
// `then` is the caller's continuation, run once the move reaches a
// terminal outcome (zoneRoute.then) — nil for the fire-and-forget
// verbs. ExileSpellThenForEffect is the one caller that needs it.
//
// Caller must hold g.mu.
func (g *Game) exitSpellFromStackLocked(spellID uuid.UUID, dst *ZoneRef, fallback ZoneKind, countered bool, then func(*Game) error) error {
	return g.exitSpellFromStackAtLocked(spellID, dst, fallback, countered, TuckOptions{}, then)
}

// exitSpellFromStackAtLocked is exitSpellFromStackLocked with a library
// POSITION: `at` rides the route exactly as it does on a tuck, so a
// library destination can be the bottom or N from the top. It exists
// for #1298's counters — Spell Crumple's "put it on the bottom of its
// owner's library" and Hinder's "your choice of the top or bottom" —
// and is ignored for every other destination. The position rides the
// route rather than being applied afterwards for the reason
// zoneRoute.Depth gives: a commander whose owner declines CR 903.9's
// offer still lands where the counter said.
//
// Caller must hold g.mu.
func (g *Game) exitSpellFromStackAtLocked(spellID uuid.UUID, dst *ZoneRef, fallback ZoneKind, countered bool, at TuckOptions, then func(*Game) error) error {
	item, ok := g.StackMeta[spellID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return ErrCardNotOnStack
	}
	if g.Stack == nil || !g.Stack.Contains(spellID) {
		return ErrCardNotOnStack
	}
	// Resolve the destination. nil → the fallback at the spell's
	// owner (a countered spell's graveyard, a returned spell's hand),
	// or exile when that player has left the game (the primitive's
	// own fallback). Battlefield / stack are illegal — an effect that
	// "puts the spell onto the battlefield" would be a different
	// effect, and either verb MUST move the spell off the stack.
	destKind, destOwner := fallback, item.Owner
	if dst != nil {
		if dst.Kind == ZoneBattlefield || dst.Kind == ZoneStack {
			return ErrInvalidStackDestination
		}
		if g.zoneFromRefLocked(*dst) == nil {
			return ErrZoneNotFound
		}
		destKind, destOwner = dst.Kind, dst.Owner
	}
	// S29 flashback: "exile this card instead of putting it anywhere
	// else ANY TIME it would leave the stack" (CR 702.34a). It beats
	// the chosen destination whichever verb is asking, so a
	// flashed-back spell answered by Hinder OR by Reprieve is exiled
	// rather than shuffled away — which is the whole difference
	// between a replacement effect and an exile bolted onto the
	// resolution, and the only place in the engine where it is
	// observable.
	for _, c := range g.Stack.Cards {
		if c.InstanceID == spellID && altCostExilesFromStack(c, item) {
			destKind, destOwner = ZoneExile, uuid.Nil
			break
		}
	}
	r := zoneRoute{
		CardID:    spellID,
		Dst:       destKind,
		DstOwner:  destOwner,
		Countered: countered,
		then:      then,
	}
	if destKind == ZoneLibrary {
		r.ToBottom, r.Depth = at.ToBottom, at.Depth
	}
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// counterAbilityLocked is the lock-free body of CounterAbility, and
// the ONE place an ability leaves the stack without resolving.
//
// CR 701.6a: "an ability that's countered doesn't go anywhere" — it is
// a DELETION, not a zone change. There is no routeCardToZoneLocked
// call here and there must not be one: the ability's source permanent
// is standing on the battlefield and stays there, the item itself has
// no card, and so none of the exits a countered SPELL contends with
// (the CR 903.9 commander window, flashback's CR 702.34a exile, a
// destination override) has anything to act on. That is also why
// CounterTargetToZoneForEffect's `dst` is meaningless for an ability
// and routes here unchanged.
//
// Caller must hold g.mu.
func (g *Game) counterAbilityLocked(abilityID uuid.UUID) error {
	item, ok := g.StackMeta[abilityID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	if item.Kind != StackItemActivated && item.Kind != StackItemTriggered {
		return ErrCardNotOnStack
	}
	source, controller, label := item.SourceCardID, item.Controller, item.Label
	delete(g.StackMeta, abilityID)
	g.recomputeSplitSecondLocked()
	// #1211: the event shape events.go documents — Source is THE
	// COUNTER, Target is the countered item. This used to put the
	// countered ability's own source card in Source, which the game
	// log reads as "who countered it": the line came out as "Llanowar
	// Elves countered <uuid>", the victim in the attacker's place and
	// an unresolvable stack-item id in the victim's. The emit site
	// does not know the counter (CounterTargetForEffect is called from
	// a resolving effect that has its own item), so Source is left
	// unset exactly as the spell path leaves it, and what the ability
	// CAN say travels instead: the permanent whose ability it was, and
	// the item's label, which is the only name a countered ability has
	// (protocol/log.go renders it).
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Actor:  controller,
		CardID: source,
		Target: abilityID,
		Label:  label,
	})
	return nil
}

// AddCounterForEffect adds (or removes, via negative delta) the
// named counter on a card. Zero deltas are no-ops. Emits
// EventCounterPlaced with the post-change count.
//
// S17 sub-PR 2: routes through the replacement pipeline so a card-
// effect-driven counter placement (Hangarback Walker's ETB,
// Tamiyo's +1 stamping loyalty, …) picks up Doubling Season /
// Hardened Scales like the public AddCounter path. Caller must
// already hold g.mu.
func (g *Game) AddCounterForEffect(cardID uuid.UUID, name string, delta int) error {
	return g.AddCounterByForEffect(uuid.Nil, cardID, name, delta)
}

// AddCounterByForEffect is AddCounterForEffect with the CR 120.3d
// placer named: `placer` is the player who PUTS the counters, which is
// the proliferating player for CR 701.34 and the damage source's
// controller for an ADR 0056 damage result.
//
// It rides the window on ReplacementEvent.CounterPlacer and lands on
// the emitted EventCounterPlaced as Actor, so "if YOU would put" and
// "whenever YOU put" read a fact instead of guessing from the last
// resolution. uuid.Nil means unknown and is what the plain spelling
// above passes; a reader falls back to its own heuristic then.
//
// Caller must already hold g.mu.
func (g *Game) AddCounterByForEffect(placer, cardID uuid.UUID, name string, delta int) error {
	// #1282: one body with the continuation form, so the two cannot
	// drift. A caller that does anything after the placement that reads
	// the counters wants AddCounterByThenForEffect instead — this
	// returns nil with NOTHING placed when the window pauses on a
	// CR 616 ordering prompt.
	return g.AddCounterByThenForEffect(placer, cardID, name, delta, nil)
}

// ReturnFromGraveyardForEffect moves a card from a player's
// graveyard to one of: owner's hand (default), battlefield (for
// Reanimate-style effects), or library top (top-of-library). The
// dest ZoneKind is one of ZoneHand, ZoneBattlefield, ZoneLibrary.
// Errors if the card isn't in a graveyard.
//
// The card lands under its OWNER's control. Reanimation that says
// "under your control" must use
// ReturnFromGraveyardUnderControlForEffect instead — see the note
// there for why the difference is not cosmetic.
func (g *Game) ReturnFromGraveyardForEffect(cardID uuid.UUID, dest ZoneKind) error {
	return g.ReturnFromGraveyardUnderControlForEffect(cardID, dest, uuid.Nil)
}

// ReturnFromGraveyardUnderControlForEffect is
// ReturnFromGraveyardForEffect with an explicit controller for the
// battlefield case: "put target creature card from A GRAVEYARD onto
// the battlefield UNDER YOUR CONTROL" (Reanimate, Portal to
// Phyrexia). uuid.Nil means "under its owner's control", which is
// the older behaviour and what Zombify-style "from your graveyard"
// text wants.
//
// This exists because the two were previously the same thing, and
// that was a real bug the moment a card reached across the table:
// the destination was picked off the graveyard's owner, so
// reanimating an opponent's creature handed the creature back to the
// opponent — the single most valuable line in a reanimator deck,
// silently inverted. Hand / library destinations are unaffected;
// those genuinely do go to the owner's zones, whoever cast the
// spell.
//
// The controller is stamped BEFORE the events fire, not after, so
// the layer listener, the trigger harvester and anything watching
// EventETB all see the permanent under the right control. Actor on
// both events is the new controller for the same reason.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardUnderControlForEffect(cardID uuid.UUID, dest ZoneKind, controller uuid.UUID) error {
	_, err := g.returnFromGraveyardLocked(cardID, dest, controller, false)
	return err
}

// ReturnFromGraveyardTappedForEffect is
// ReturnFromGraveyardUnderControlForEffect with the CR 614
// "enters tapped" clause riding the same event
// ReturnToBattlefieldForEffect's `tapped` already does for the
// exile-return path (#1178) — Reassembling Skeleton and Drownyard
// Temple's own printed "return this card from your graveyard to the
// battlefield tapped" (#1284).
//
// A separate function rather than a fourth parameter on
// ReturnFromGraveyardUnderControlForEffect on purpose: that one has
// call sites across the catalog and the engine's own tests today, all
// of them meaning "untapped", and a bare positional bool at the end
// of an existing signature is the kind of change a diff reviews past
// without noticing which call sites silently kept the old meaning and
// which needed the new one. `tapped` is meaningless for anything but
// Dest == ZoneBattlefield, same as ReturnToBattlefieldForEffect's.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardTappedForEffect(cardID uuid.UUID, dest ZoneKind, controller uuid.UUID, tapped bool) error {
	_, err := g.returnFromGraveyardLocked(cardID, dest, controller, tapped)
	return err
}

// ReturnToBattlefieldForEffect is "return it to the battlefield"
// (optionally TAPPED) said of a card whose zone the effect does not
// know — which is exactly what a delayed trigger keyed on "when it
// dies or is exiled" is holding (#1178, earthbend.go).
//
// It dispatches on where the card actually is. Exile and a graveyard
// are the two zones such a trigger can find it in, and each already
// has its own entry point with its own CR 400.7 bookkeeping — the
// exile return re-mints the instance ID, the graveyard return leaves
// it and relies on the re-minted battlefield-entry stamp — so this is
// a ROUTER over the two, not a fourth entry primitive. Anywhere else
// (a hand, a library, the battlefield already) is ErrCardNotFound:
// the object the effect named is not where it was left, which is the
// CR 608.2b posture and the caller's to swallow.
//
// `controller` uuid.Nil means "under its owner's control", which is
// what a return naming no controller gets (CR 400.3).
//
// Returns the entering permanent's ID, or uuid.Nil when nothing
// entered — the entry pipeline can cancel or redirect the move, and it
// can PAUSE, in which case the entry completes from the resume and
// nothing is reported here.
//
// Caller must hold g.mu.
func (g *Game) ReturnToBattlefieldForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	src := g.findCardZoneLocked(cardID)
	if src == nil {
		return uuid.Nil, ErrCardNotFound
	}
	switch src.Kind {
	case ZoneExile:
		return g.ReturnFromExileToBattlefieldForEffect(cardID, controller, tapped)
	case ZoneGraveyard:
		return g.returnFromGraveyardLocked(cardID, ZoneBattlefield, controller, tapped)
	}
	return uuid.Nil, ErrCardNotFound
}

// returnFromGraveyardLocked is the shared body of the two graveyard
// entry points above, plus the `tapped` clause neither of the older
// spellings could state.
//
// `tapped` rides onto the CR 614 event rather than being OR-ed in
// afterwards, for the reason SearchLibrarySpec.TappedOnEntry and
// ReturnFromExileToBattlefieldForEffect both give: a resume reads
// ev.EntersTapped and has no idea what effect sent the card, so a
// return that paused on an entry prompt would otherwise come back
// untapped. It is meaningless for a hand or library destination and
// ignored there.
//
// Caller must hold g.mu.
func (g *Game) returnFromGraveyardLocked(cardID uuid.UUID, dest ZoneKind, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	return g.returnFromGraveyardFaceLocked(cardID, dest, controller, tapped, FaceDownNone, nil)
}

// ReturnFromGraveyardFaceDownForEffect is "return it to the
// battlefield FACE DOWN under its owner's control. It's a Forest land."
// — Yedora, Grave Gardener; and, with a controller and `tapped`,
// Missy's "under your control face down and tapped. It's a 2/2
// Cyberman artifact creature."
//
// It is the reanimation door (the CR 614 entry pipeline, the resume,
// the controller stamped before the pipeline runs) with the face-down
// state riding the entry EVENT, exactly as a morph's does — so the
// one entry finisher lands it as a CR 708.2 object whose controller
// is its only knower, and no ETB trigger or "as enters" hook fires
// because CatalogKey has already gone silent (CR 708.2a).
//
// The kind is FaceDownTurned: an effect put it here, not a keyword,
// so only the card's own morph or disguise brings it back up
// (CR 708.7, CR 702.37e). `listed` is the body the card lists
// (CR 708.2); nil is the default nameless 2/2.
//
// `controller` uuid.Nil means "under its owner's control". Returns
// the entering permanent's ID, uuid.Nil when nothing entered, and
// ErrCardNotFound when the card is not in a graveyard any more —
// the CR 400.7 answer a "return it" trigger swallows.
//
// Caller must hold g.mu. Added for #1270.
func (g *Game) ReturnFromGraveyardFaceDownForEffect(cardID, controller uuid.UUID, tapped bool, listed *FaceDownListing) (uuid.UUID, error) {
	return g.returnFromGraveyardFaceLocked(cardID, ZoneBattlefield, controller, tapped, FaceDownTurned, listed)
}

// returnFromGraveyardFaceLocked is returnFromGraveyardLocked with the
// face-down state a battlefield entry lands in. FaceDownNone is every
// face-up return.
func (g *Game) returnFromGraveyardFaceLocked(cardID uuid.UUID, dest ZoneKind, controller uuid.UUID, tapped bool, faceDown FaceDownKind, listed *FaceDownListing) (uuid.UUID, error) {
	src := g.findCardZoneLocked(cardID)
	if src == nil || src.Kind != ZoneGraveyard {
		return uuid.Nil, ErrCardNotFound
	}
	// Identify the graveyard's owner so we can route to the same
	// player's hand / library. (Graveyards are per-player; Zone.Owner
	// holds the ID.)
	ownerID := src.Owner
	owner := g.playerByIDLocked(ownerID)
	if owner == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	var destZone *Zone
	switch dest {
	case ZoneHand:
		destZone = owner.Hand
	case ZoneLibrary:
		destZone = owner.Library
	case ZoneBattlefield:
		destZone = g.Battlefield
	default:
		return uuid.Nil, ErrZoneNotFound
	}
	if destZone.Kind == ZoneBattlefield {
		newController := controller
		if newController == uuid.Nil {
			newController = ownerID
		}
		if p := g.playerByIDLocked(newController); p == nil {
			newController = ownerID
		}
		// Stamp the controller BEFORE the pipeline runs. A
		// replacement that asks "is this entering permanent mine?"
		// (Authority of the Consuls) reads Card.Controller, and while
		// the card sits in the graveyard that field still names
		// whoever controlled it when it died — which, for a
		// reanimated opponent's creature, is the wrong player.
		for i := range src.Cards {
			if src.Cards[i].InstanceID == cardID {
				src.Cards[i].Controller = newController
				break
			}
		}
		// #263: a reanimated permanent enters the battlefield like
		// any other, so the CR 614 entry pipeline has to run on it.
		// Skipping it meant a reanimated shockland ignored its own
		// enters-tapped clause and a creature reanimated under an
		// opponent's nose dodged Authority of the Consuls.
		//
		// #478: and the entry can PAUSE — a reanimated shockland is
		// asked to pay, two entry replacements queue the CR 616
		// ordering prompt. It used to be dropped on that pause; it is
		// resumable now, through the same finisher the unpaused path
		// runs, so nothing moves until the answer arrives and then the
		// permanent enters exactly as it would have.
		//
		// #1178: and the effect's own "return it TAPPED" clause rides
		// the event for the third reason in that list — a resume reads
		// ev.EntersTapped and nothing else knows what asked for the
		// return.
		return g.enterBattlefieldThroughPipelineLocked(&ReplacementEvent{
			Kind:         RepEventMove,
			Actor:        newController,
			CardID:       cardID,
			OldZone:      ZoneGraveyard,
			NewZone:      ZoneBattlefield,
			EntersTapped: tapped,
			// #1270: a face-down return rides the event the way a
			// morph's entry does, so the one finisher lands it.
			FaceDown:       faceDown,
			FaceDownListed: listed,
			entryResumable: true,
		})
	}
	// A hand or library destination is not an entry: no pipeline, no
	// ETB, nothing to pause on, and nothing for `tapped` to mean.
	if _, err := MoveCard(src, destZone, cardID); err != nil {
		return uuid.Nil, err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   ownerID,
		CardID:  cardID,
		OldZone: ZoneGraveyard,
		NewZone: destZone.Kind,
	})
	return uuid.Nil, nil
}

// SearchLibrarySpec is the full description of one "search your
// library for ..." effect. Every field except Player and Pred has a
// useful zero value, so the older positional entry points below stay
// one-line wrappers over it.
//
// The spec exists because a search is no longer a single synchronous
// mutation: when the library offers more matches than the effect may
// take, the searcher gets a real PendingChoiceSearchLibrary prompt
// and the rest of the effect has to wait for their answer. `Then`
// is that continuation.
type SearchLibrarySpec struct {
	// Player is the searcher — always the owner of the library being
	// searched. Assassin's Trophy makes the VICTIM search, so this is
	// not necessarily the resolving spell's controller.
	Player uuid.UUID

	// Source is the card that caused the search, used only as prompt
	// context on the wire. Optional.
	Source uuid.UUID

	// Pred filters the library. nil matches every card (Gamble).
	Pred func(Card) bool

	// Dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary, ZoneExile.
	Dest ZoneKind

	// FaceDown exiles the taken cards face down (CR 406.3a) when Dest
	// is ZoneExile — the zero value, FaceDownNone, is an ordinary
	// face-up exile. Rides straight onto the zoneRoute the takes are
	// routed through (searchRoute), the same way it does for every
	// other exit, so CR 614 and CR 903.9 still see the move and
	// applyFaceDownLandingLocked still decides who may look. Ignored
	// for any other Dest. No printed card in the catalog sets it yet;
	// it exists so a future "search your library for a card, exile it
	// face down" does not reopen this file.
	FaceDown FaceDownKind

	// LibraryOwner is the seat whose library is actually searched,
	// when that is not Player — Bribery's "search TARGET OPPONENT's
	// library", where the caster (Player) does the choosing but the
	// pile scanned, shuffled and left un-knowing afterwards is the
	// target's. The zero value (uuid.Nil) means "the same as Player",
	// which is what every search before #1230 already meant and what
	// every other field on this spec keeps assuming: the chooser, the
	// knower grant, the entering CONTROLLER (searchEnterBattlefieldLocked)
	// and the stamped Actor are all Player regardless, exactly as
	// Bribery's "under YOUR control" wants. Assassin's Trophy's
	// "that player searches" is not this — there Player already IS
	// the searched owner, so LibraryOwner is unset.
	LibraryOwner uuid.UUID

	// Limit is the maximum number of cards taken. <= 0 means 1.
	// Meaningless with Unbounded set.
	Limit int

	// Unbounded is "search your library for ANY NUMBER of … cards"
	// (Ugin, Eye of the Storms' −11): the searcher may take as many
	// of the matches as they choose, from zero up to all of them, and
	// nothing about the count forces a decision the way Limit's
	// "take every match" shortcut does. Separate from Limit rather
	// than a sentinel value on it (0 already means "default to 1")
	// for the reason CounterCost.Variable is its own bool beside a
	// numeric field rather than overloading one: a search that finds
	// exactly one match still has to ask, because "any number"
	// includes zero.
	Unbounded bool

	// Reveal marks the TAKEN cards known to every seated player (CR
	// "reveal"). It never reveals the cards that were merely
	// considered — see queueSearchChoiceLocked for why that
	// distinction is load-bearing.
	Reveal bool

	// Shuffle shuffles the library once the search is finished, and
	// wipes per-card knowledge with it.
	Shuffle bool

	// TappedOnEntry forces a fetched permanent tapped. It expresses
	// the FETCHING effect's printed text ("put it onto the
	// battlefield tapped" — Cultivate, Solemn Simulacrum, Path to
	// Exile), NOT the fetched card's own enters-tapped clause; the
	// CR 614 pipeline owns that one. The two are OR-ed: a Darkslick
	// Shores fetched by Cultivate is tapped because Cultivate says
	// so, and one fetched by Skyshroud Claim is tapped only if its
	// own condition fires.
	TappedOnEntry bool

	// Optional marks a "you MAY search" (CR 701.23b). It forces the
	// prompt even when the pick is otherwise forced, so the searcher
	// can decline — Assassin's Trophy's victim keeps the right to
	// refuse the land and the shuffle.
	Optional bool

	// Reason is the prompt's banner copy. Defaults to a generic
	// "Search your library".
	Reason string

	// Validate is an extra legality check on the picked SET, for
	// clauses the per-card predicate cannot express. Myriad
	// Landscape's "two basic land cards that SHARE A LAND TYPE" is
	// the case: no predicate over one card can say it. Called with
	// the chosen cards (possibly zero of them, which is always
	// legal — you may always fail to find).
	Validate func([]Card) bool

	// Then is the rest of the effect, run once the search has
	// finished — immediately when no prompt was needed, and from
	// ResolveSearchLibrary when one was. It receives the instance
	// IDs actually taken, in the order they were taken, so an effect
	// like Fabled Passage's "then ... untap THAT LAND" can name what
	// it found instead of re-deriving it.
	Then func(g *Game, found []uuid.UUID) error

	// ToTop puts the found cards on TOP of the library, after the
	// shuffle, when Dest is ZoneLibrary. It is what the "search your
	// library for a card, then shuffle and PUT THAT CARD ON TOP"
	// tutors print — Vampiric Tutor, Enlightened Tutor, Worldly
	// Tutor, Mystical Tutor, Imperial Seal.
	//
	// The ordering is the whole clause. Shuffling first and placing
	// second is what makes the card a known quantity on an unknown
	// library; doing it the other way round would shuffle the card
	// you just tutored back into the deck.
	//
	// With ToTop false, Dest: ZoneLibrary keeps its older meaning —
	// "leave it where it is", which is a search that only reveals.
	// Ignored for any other destination. Added in S22.
	ToTop bool

	// Depth is ToTop's position, for "then shuffle and put that card
	// THIRD from the top" (Long-Term Plans, ADR 0088 Decision 4): the
	// found card goes Depth cards down, through Zone.InsertFromTop, so
	// 0 and 1 are the top and a library shorter than Depth takes the
	// card on the bottom. Meaningful only with ToTop.
	Depth int
}

// libraryOwnerID is the seat whose library this search actually reads
// — LibraryOwner when it names one, Player otherwise. Every OTHER
// field on the spec (the chooser, the knower grant, the entering
// controller, the stamped Actor) already reads Player directly and is
// unaffected by this: a foreign-library search changes WHICH pile is
// scanned, shuffled and moved out of, not who is doing the searching.
func (spec SearchLibrarySpec) libraryOwnerID() uuid.UUID {
	if spec.LibraryOwner != uuid.Nil {
		return spec.LibraryOwner
	}
	return spec.Player
}

// SearchLibraryForEffect scans playerID's library for cards matching
// `pred`, moves up to `limit` of them to the given destination zone,
// reveals (via KnownBy) to all seated players when `reveal` is true,
// and shuffles the library when `shuffle` is true.
//
// dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary (library-
// top is not distinguishable from library via ZoneKind; a future
// "top N" primitive can extend this). Used by Demonic Tutor
// (dest=Hand, limit=1, reveal=false, shuffle=true), Vampiric Tutor
// (dest=Library, limit=1, ...).
//
// The searcher picks which matches they take (S22) — see
// SearchLibraryThenForEffect. This wrapper has no continuation, so
// it suits only effects that do nothing after the search.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryForEffect(
	playerID uuid.UUID,
	pred func(Card) bool,
	dest ZoneKind,
	limit int,
	reveal bool,
	shuffle bool,
) error {
	return g.SearchLibraryForEffectWithOptions(playerID, pred, dest, limit, reveal, shuffle, false)
}

// SearchLibraryForEffectWithOptions is SearchLibraryForEffect with
// the forced-tap flag. Retained as a positional wrapper because a
// dozen catalog cards call it; new callers should build a
// SearchLibrarySpec instead.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryForEffectWithOptions(
	playerID uuid.UUID,
	pred func(Card) bool,
	dest ZoneKind,
	limit int,
	reveal bool,
	shuffle bool,
	tappedOnEntry bool,
) error {
	return g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player:        playerID,
		Pred:          pred,
		Dest:          dest,
		Limit:         limit,
		Reveal:        reveal,
		Shuffle:       shuffle,
		TappedOnEntry: tappedOnEntry,
	})
}

// SearchLibraryThenForEffect is the real entry point: it runs the
// search described by spec and then spec.Then.
//
// The searcher chooses. Until S22 the engine took the first matches
// in library order, which made a fetchland's entire strategic content
// — which dual do I want? — a property of deck-list order, and made
// Myriad Landscape's "share a land type" whatever the bottom-most
// basic happened to be. That simplification is gone: when the library
// offers a real choice, the searcher gets a prompt.
//
// "A real choice" means: more matches than the effect may take, or
// an Optional ("you may search") clause, or a Validate constraint the
// whole match set does not already satisfy. With nothing to decide —
// a Rampant Growth into a library holding exactly one Forest — the
// engine takes the match and moves on, because a modal offering one
// button is worse than no modal.
//
// Hidden information (CR 400.2): a library is a hidden zone, and the
// prompt does not change that for anybody but the searcher. The
// candidates are marked known to the SEARCHER ALONE, the per-viewer
// wire filter redacts them for every other seat, and the search
// prompt's option list is withheld from non-choosers entirely so the
// NUMBER of matches does not leak either. The searcher is always the
// library's owner, so nothing crosses the table.
//
// Caller must hold g.mu.
func (g *Game) SearchLibraryThenForEffect(spec SearchLibrarySpec) error {
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	if spec.Limit <= 0 && !spec.Unbounded {
		spec.Limit = 1
	}
	if _, err := g.searchDestZoneLocked(p, spec.Dest); err != nil {
		return err
	}
	// Collect every match, not just the first `limit` — the whole
	// point of the chooser is that the searcher sees the full set.
	matches := make([]uuid.UUID, 0, len(p.Library.Cards))
	matched := make([]Card, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		if spec.Pred == nil || spec.Pred(c) {
			matches = append(matches, c.InstanceID)
			matched = append(matched, c)
		}
	}
	if len(matches) == 0 {
		// Nothing to find. Still a search: the event fires and the
		// shuffle happens, because the card said "then shuffle"
		// unconditionally.
		return g.finishSearchLocked(spec, p, nil)
	}
	if spec.Unbounded {
		// "Any number" is a ceiling, never a target (CR 701.23a): the
		// searcher may take zero up to every match, so the count is
		// always a real choice and Limit becomes the size of the
		// whole match set rather than a forced take-all.
		spec.Limit = len(matches)
	}
	if !spec.Optional && !spec.Unbounded && len(matches) <= spec.Limit &&
		(spec.Validate == nil || spec.Validate(matched)) {
		// No decision to make — take them all. The take finishes the
		// search itself (#478): a battlefield destination is an entry
		// and an entry can pause, so "then shuffle" is the last link of
		// a chain rather than the line after the loop.
		return g.executeSearchTakeLocked(spec, p, matches)
	}
	g.queueSearchChoiceLocked(spec, p, matches)
	return nil
}

// searchDestZoneLocked resolves a search's destination ZoneKind to
// the concrete zone. Caller must hold g.mu.
func (g *Game) searchDestZoneLocked(p *Player, dest ZoneKind) (*Zone, error) {
	switch dest {
	case ZoneHand:
		return p.Hand, nil
	case ZoneBattlefield:
		return g.Battlefield, nil
	case ZoneLibrary:
		return p.Library, nil
	case ZoneGraveyard:
		// Entomb, Buried Alive, Gamble's discard half — "search your
		// library for a card, put that card into your GRAVEYARD".
		// The generic MoveCard branch below handles it unchanged;
		// only this lookup was missing, which made those cards fail
		// with ErrZoneNotFound and silently find nothing. Added with
		// the roadmap's batch 02 (#295).
		return p.Graveyard, nil
	case ZoneExile:
		// #1230: "search your library for any number of … cards, exile
		// them, then shuffle" (Ugin, Eye of the Storms). Exile is a
		// single shared zone, not per-player, so p is unread here —
		// unlike the three cases above it carries no ownership
		// question at all. The generic MoveCard branch below (through
		// searchRoute / routeCardToZoneLocked, which already resolves
		// ZoneExile) handles the move unchanged.
		return g.Exile, nil
	}
	return nil, ErrZoneNotFound
}

// queueSearchChoiceLocked opens the search prompt. The candidates
// are marked known to the searcher only: they are looking through
// their own library, which is exactly the knowledge the rules grant
// them and no more. `Reveal` is deliberately NOT applied here —
// "reveal those cards" on Cultivate names the cards you take, not
// every basic you flipped past on the way, and revealing the whole
// match set would hand the table a census of the library.
//
// Caller must hold g.mu.
func (g *Game) queueSearchChoiceLocked(spec SearchLibrarySpec, p *Player, matches []uuid.UUID) {
	known := make(map[uuid.UUID]bool, len(matches))
	for _, id := range matches {
		known[id] = true
	}
	for i := range p.Library.Cards {
		if known[p.Library.Cards[i].InstanceID] {
			p.Library.Cards[i].AddKnower(spec.Player)
		}
	}
	reason := spec.Reason
	if reason == "" {
		reason = "Search your library"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceSearchLibrary,
		Chooser:      spec.Player,
		FromPlayer:   spec.Player,
		Count:        spec.Limit,
		Source:       spec.Source,
		Reason:       reason,
		SearchCards:  append([]uuid.UUID(nil), matches...),
		SearchMax:    spec.Limit,
		searchResume: &searchResumeFrame{spec: spec},
	})
}

// executeSearchTakeLocked moves the chosen cards out of the library
// and finishes the search with the ones that actually got there.
// Shared by the no-decision path and by ResolveSearchLibrary, so a
// fetched permanent enters identically either way.
//
// It does not return the found list, because it cannot: both the
// battlefield branch (#478) and the hand / graveyard branch (#931) can
// PAUSE on a player prompt, so the search finishes from a
// continuation. finishSearchLocked is the single place it ends.
//
// Caller must hold g.mu.
func (g *Game) executeSearchTakeLocked(spec SearchLibrarySpec, p *Player, ids []uuid.UUID) error {
	destZone, err := g.searchDestZoneLocked(p, spec.Dest)
	if err != nil || destZone == p.Library {
		// ZoneLibrary means the card does not leave the library.
		// With ToTop set it will be moved to the top AFTER the
		// shuffle, in finishSearchLocked — so the IDs are carried
		// through as found, and nothing is moved here. Without it the
		// clause is "leave it where it is" and only the reveal applies.
		if err != nil {
			return g.finishSearchLocked(spec, p, nil)
		}
		if spec.Reveal {
			g.revealLibraryCardsLocked(spec, p, ids)
		}
		if spec.ToTop {
			return g.finishSearchLocked(spec, p, ids)
		}
		return g.finishSearchLocked(spec, p, nil)
	}
	// S22: the reveal happens ONCE, before anything moves, and names
	// every card the search took. Cultivate's "reveal those cards" is
	// one announcement about two lands rather than two announcements
	// about one each, and revealing before the move is what the card
	// prints — the table sees the cards where they were found.
	if spec.Reveal {
		g.revealLibraryCardsLocked(spec, p, ids)
	}
	if destZone.Kind == ZoneBattlefield {
		// #478: a battlefield destination is an ENTRY, and an entry can
		// pause. The takes are therefore SEQUENCED through the entry's
		// continuation rather than looped — each card's entry starts the
		// next one and the last one finishes the search — so a Skyshroud
		// Claim whose first Forest stops for a prompt does not fetch the
		// second one over the top of the open question. The same idiom
		// the discard batch and the exit sweeps use, and the same reason:
		// the found list is carried forward BY VALUE, which is what makes
		// an undo across the prompt replay identically.
		return g.searchEnterEachThenFinishLocked(spec, ids, nil)
	}
	// #931: a hand or graveyard destination is an EXIT, and every exit
	// in the engine goes through the one primitive. This used to be a
	// raw MoveCard loop, which is exactly the shape zone_route.go was
	// written to delete: no CR 614 window, so "if a card would be put
	// into a graveyard from anywhere, exile it instead" (Rest in Peace,
	// Leyline of the Void) could not see an Entomb, and no CR 903.9
	// offer for a tutored commander.
	//
	// It is the shared batch body (routeAllThenLocked), not a loop,
	// because a leg can now PAUSE: the takes are sequenced, each from
	// the previous one's continuation, and the search finishes — the
	// EventSearchLibrary, the shuffle, spec.Then — only once they have
	// all settled. The same sequencing the battlefield branch above got
	// in #478, for the same reason.
	//
	// `found` is what the batch reports: the cards that ARRIVED where
	// the search asked (CR 400.7, landedInZoneLocked). A card an
	// "exile it instead" replacement took, and a commander that took
	// CR 903.9's offer, are not in it — the same reading the
	// battlefield branch has used since #478, where a fetched permanent
	// whose entry was replaced away is not "found" either.
	return g.routeAllThenLocked(searchRoute(spec.Player, spec.Dest, spec.FaceDown), ids, func(g *Game, found []uuid.UUID) error {
		p := g.playerByIDLocked(spec.libraryOwnerID())
		if p == nil {
			return ErrPlayerNotFound
		}
		return g.finishSearchLocked(spec, p, found)
	})
}

// searchEnterEachThenFinishLocked puts the head of `ids` onto the
// battlefield and continues with the tail from that entry's
// continuation. The empty list is the base case: the search is finished
// (EventSearchLibrary, the shuffle, spec.Then) with the cards that
// actually entered.
//
// The player is re-looked-up from the live *Game on every step rather
// than captured, on the undo-safety contract every continuation in the
// engine follows.
//
// Caller must hold g.mu.
func (g *Game) searchEnterEachThenFinishLocked(spec SearchLibrarySpec, ids, found []uuid.UUID) error {
	p := g.playerByIDLocked(spec.libraryOwnerID())
	if p == nil {
		return ErrPlayerNotFound
	}
	// A card that is no longer in the library is skipped rather than
	// fetched: it left while an earlier card's entry prompt was open.
	for len(ids) > 0 && !p.Library.Contains(ids[0]) {
		ids = ids[1:]
	}
	if len(ids) == 0 {
		return g.finishSearchLocked(spec, p, found)
	}
	head, rest := ids[0], ids[1:]
	return g.searchEnterBattlefieldLocked(spec, p, head, func(g *Game, entered uuid.UUID) error {
		out := found
		if entered != uuid.Nil {
			// A fresh slice rather than an append in place: two runs of
			// the same continuation (an undo, then the same answer
			// again) must not see each other's entry.
			out = append(append(make([]uuid.UUID, 0, len(found)+1), found...), entered)
		}
		return g.searchEnterEachThenFinishLocked(spec, rest, out)
	})
}

// searchEnterBattlefieldLocked puts one fetched card onto the
// battlefield through the CR 614 entry pipeline (#263).
//
// Before this, the search path pushed the card straight onto the
// battlefield and emitted EventZoneMove / EventETB without ever
// calling applyReplacementsLocked — so a fetched Darkslick Shores
// ignored its own "enters tapped unless" clause, every conditional
// dual cycle came in strictly better than printed, and a creature an
// opponent fetched walked past Authority of the Consuls. The ad-hoc
// TappedOnEntry flag hid it for the always-tapped cases and could not
// express a conditional one.
//
// The pipeline is consulted BEFORE the card leaves the library, for
// the same reason the exile-return path does it: a paused entry moves
// nothing, and pausing with the card already lifted out of its zone
// would strand it.
//
// #478: the entry is RESUMABLE, so a fetched permanent gets its entry
// prompt. A fetchland cracking for a shockland asks "pay 2 life so it
// enters untapped?" exactly as the play path does, and two entry
// replacements on one fetched land (Kismet plus Thalia, Heretic
// Cathar) queue the CR 616 ordering prompt and then finish the fetch
// when it is answered.
//
// That used to be a declared simplification, and the argument against
// closing it was that executeEntryToBattlefieldLocked could finish the
// MOVE but knew nothing about the search that started it — the library
// would never be shuffled, EventSearchLibrary would never fire, and the
// Then continuation (Fabled Passage's "untap that land", Gamble's
// random discard) would never run. All three of those now ride across
// the pause on ev.entryTail, so the resume finishes the search rather
// than only the move; see entry_tail.go. With two applicable
// replacements the OLD behaviour was not a graceful degradation at all:
// the prompt was still queued, the player still answered it, and the
// card was still in the library afterwards.
//
// `then` is handed the entering permanent's ID, or uuid.Nil when
// nothing entered — cancelled, redirected, or its prompt taken away. It
// runs from every terminal outcome, which is what lets the search
// sequence its remaining takes through it.
//
// Caller must hold g.mu.
func (g *Game) searchEnterBattlefieldLocked(spec SearchLibrarySpec, p *Player, id uuid.UUID, then func(g *Game, entered uuid.UUID) error) error {
	// The controller has to be stamped before the pipeline runs:
	// Authority of the Consuls asks whose permanent is entering, and
	// a self-replacement's condition ("unless you control two or
	// fewer other lands") counts the lands of whoever that is.
	for i := range p.Library.Cards {
		if p.Library.Cards[i].InstanceID == id {
			p.Library.Cards[i].Controller = spec.Player
			break
		}
	}
	_, err := g.enterBattlefieldThroughPipelineLocked(&ReplacementEvent{
		Kind:    RepEventMove,
		Actor:   spec.Player,
		CardID:  id,
		OldZone: ZoneLibrary,
		NewZone: ZoneBattlefield,
		// The fetching effect's own "put it onto the battlefield
		// TAPPED" clause is seeded onto the event rather than OR-ed
		// in after the pipeline. Same result for the inline path, and
		// it is the difference between right and wrong for the resume
		// path: a resume reads ev.EntersTapped and has no idea what
		// spell sent the card, so a Cultivate-fetched land that
		// paused for a prompt would otherwise come back untapped.
		EntersTapped:   spec.TappedOnEntry,
		entryResumable: true,
		entryTail:      &entryTail{then: then},
	})
	return err
}

// revealLibraryCardsLocked reveals the named library cards to the
// whole table — the CR 701.20 sense of "reveal" that a tutor prints
// between "search your library for a card" and "put it into your
// hand".
//
// S22: this used to be a KnownBy loop and nothing else, which made
// every tutor's reveal silent. An opponent could read the fetched
// card once it reached a hand but had no way to tell WHEN, or to
// tell a tutored card apart from a drawn one; and for a to-top tutor
// (Enlightened Tutor) the card never reached a visible zone at all,
// so the loudest downside those cards print never happened. It now
// goes through the shared reveal primitive, which does the same
// KnownBy marking and additionally announces it on the wire.
//
// Only the cards actually TAKEN are named. The matches the searcher
// merely flipped past stay private — see queueSearchChoiceLocked for
// why that distinction is load-bearing.
//
// Caller must hold g.mu.
func (g *Game) revealLibraryCardsLocked(spec SearchLibrarySpec, p *Player, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	inLibrary := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if p.Library.Contains(id) {
			inLibrary = append(inLibrary, id)
		}
	}
	reason := spec.Reason
	if reason == "" {
		reason = "revealed from library"
	}
	g.RevealForEffect(RevealSpec{
		Player: spec.Player,
		Source: spec.Source,
		Reason: reason,
		Cards:  inLibrary,
	})
}

// finishSearchLocked emits EventSearchLibrary, shuffles if the card
// said to, and runs the continuation. Caller must hold g.mu.
func (g *Game) finishSearchLocked(spec SearchLibrarySpec, p *Player, found []uuid.UUID) error {
	g.EmitEvent(Event{
		Kind:  EventSearchLibrary,
		Actor: spec.Player,
		// #1335: `p` is always the library owner (SearchLibraryThenForEffect
		// resolves it via libraryOwnerID before calling here), which
		// differs from the searcher (spec.Player) for Bribery-style
		// search of another player's library. Without this, "an
		// opponent searches THEIR library" (Archivist of Oghma) could
		// not be told apart from "an opponent searches YOUR library".
		Target: p.ID,
		Amount: len(found),
	})
	if spec.Shuffle {
		p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
		// The shuffle is what un-knows the library again: whatever
		// the searcher saw while looking, they no longer know where
		// any of it is.
		clearKnownInZoneLocked(p.Library)
	}
	// "... then shuffle and put that card on top." AFTER the shuffle,
	// which is the whole clause — the tutored card is a known
	// quantity sitting on an unknown library. Doing it before would
	// shuffle the card straight back into the deck.
	//
	// Knowledge has to be re-granted here because the shuffle just
	// wiped the whole zone, and the tutored card is the one card in
	// it nobody has forgotten: the searcher chose it and watched it
	// go on top. If the card also said "reveal", the whole table
	// watched — Enlightened Tutor tells everyone what your next draw
	// is, and that is a real cost of the card.
	if spec.ToTop && spec.Dest == ZoneLibrary {
		for i := len(found) - 1; i >= 0; i-- {
			c, err := p.Library.Remove(found[i])
			if err != nil {
				continue
			}
			c.AddKnower(spec.Player)
			if spec.Reveal {
				for _, seat := range g.Seats {
					c.AddKnower(seat.ID)
				}
			}
			p.Library.InsertFromTop(c, spec.Depth)
		}
	}
	if spec.Then != nil {
		return spec.Then(g, found)
	}
	return nil
}

// CreateTokenForEffect creates n tokens under `controller`'s control
// from `template` (CR 701.7b). The common shape, and the one nearly
// every catalog card wants: no entry clause, nothing to do
// afterwards.
//
// Since #762 it is one call on the shared creation path — the CR
// 701.7b replacement window, then an ordinary battlefield entry per
// token. See token_create.go.
//
// Caller must hold g.mu.
func (g *Game) CreateTokenForEffect(controller uuid.UUID, template Card, n int) error {
	_, err := g.CreateTokensForEffect(controller, template, n, TokenEntryOptions{})
	return err
}

// TokenEntryOptions are the creation-time modifiers a card can apply
// to the token it makes, on top of the token's printed template.
// Every field is the zero value for an ordinary "create a 1/1 Goblin"
// — CreateTokenForEffect passes an empty struct.
//
// These are properties of the CREATION, not of the token: two cards
// can make the same printed Powerstone and only one of them says
// "tapped". Keeping them here rather than baking them into the
// template in tokens.go is what stops the catalog growing a second
// near-identical constructor per variant.
type TokenEntryOptions struct {
	// Tapped enters the tokens tapped — "create a TAPPED Powerstone
	// token" (Stern Lesson), "create a 1/1 Goblin tapped and
	// attacking" minus the attacking half, which the creation's own
	// Attacking field carries.
	//
	// Since #762 it is SEEDED onto the entry event rather than
	// stamped and forgotten, the way ZoneEntryOptions.Tapped is: the
	// creation's clause and whatever an enters-tapped replacement
	// (Urabrask, Kismet, Thalia) adds on top settle in one field, and
	// no reader downstream can lose one of the two.
	Tapped bool

	// Counters are the counters each token enters with, keyed by
	// counter name. Applied before EventETB fires, so an ETB watcher
	// and the P/T recompute both see the finished object.
	//
	// They ride the entry event as EntersWithCounters (#762), which
	// is what puts them through the CR 614 counter pipeline: a
	// Doubling Season doubles them, Renata and Arwen add to them, and
	// All Will Be One sees them placed.
	Counters map[string]int

	// Keywords are granted on top of the template's printed ones —
	// the "…with haste" half of a card that pumps the token it makes.
	// Additive: the template's own keywords are kept.
	Keywords []string
}

// CreateTokensForEffect is CreateTokenForEffect with entry options,
// returning the instance IDs of the tokens it made in creation order.
// The IDs are what a card needs when the token is not the end of the
// sentence — "create a token, then sacrifice it", "…then put a
// counter on it".
//
// THE RETURNED SLICE IS EMPTY WHEN THE CREATION PAUSED. Since #762 a
// creation goes through the CR 701.7b replacement window, and a
// window with two DIFFERENT applicable effects in it (a Doubling
// Season and an Academy Manufactor) queues a CR 616 ordering prompt;
// the tokens are made when it is answered, an action later, and there
// is nothing to return here. A caller whose sentence continues past
// the tokens must hand that continuation to
// CreateTokensThenForEffect rather than read this slice on the next
// line — the same contract MillToZoneThenForEffect and
// discardCardsLocked carry.
//
// Caller must hold g.mu.
func (g *Game) CreateTokensForEffect(controller uuid.UUID, template Card, n int, opts TokenEntryOptions) ([]uuid.UUID, error) {
	var made []uuid.UUID
	err := g.CreateTokensThenForEffect(TokenCreation{
		Controller: controller,
		Groups:     []TokenGroup{{Template: template, Count: n, Entry: opts}},
	}, func(_ *Game, created []uuid.UUID) error {
		made = created
		return nil
	})
	return made, err
}

// CreateTokensAttackingForEffect is CreateTokenForEffect for "create
// N tokens … that are attacking" (Parhelion II, Hanweir Garrison):
// the tokens are put onto the battlefield already attacking
// `defender`.
//
// CR 506.3c is why the attacking-ness is part of the CREATION rather
// than something done to the tokens afterwards: a permanent PUT onto
// the battlefield attacking was never DECLARED as an attacker, so it
// fires no "whenever ~ attacks" trigger and nothing that watches
// attack declarations sees it. Setting AttackingTarget as the token
// is minted and emitting no EventAttack is exactly that rule, and
// routing through DeclareAttacker — which emits the event, checks
// summoning sickness and taps — would be wrong on all four counts.
//
// A zero `defender`, or a defender that is not a seated player,
// creates the tokens untapped and not attacking rather than erroring:
// the ability that called this has already resolved, and the tokens
// are the part of it that can still be delivered.
//
// Caller must hold g.mu.
func (g *Game) CreateTokensAttackingForEffect(controller uuid.UUID, template Card, n int, defender uuid.UUID) error {
	return g.CreateTokensThenForEffect(TokenCreation{
		Controller: controller,
		Groups:     []TokenGroup{{Template: template, Count: n}},
		Attacking:  defender,
	}, nil)
}

// ScryForEffect performs a scry N (CR 701.22): the player looks at the
// top N cards of their library and then decides which go to the bottom
// and in what order the rest go back on top.
//
// Returns the number of cards actually looked at, which is fewer than n
// when the library is short and zero when it is empty — a scry with an
// empty library is not an error, it simply does nothing, and no choice
// is queued. It is also zero when the CR 614 keyword-action window
// PAUSED on an ordering prompt, in which case nothing has been queued
// yet and the resume queues it; see ScryThenForEffect.
//
// Scry is LOOK AT, not reveal. Only the scrying player becomes a knower
// of the cards, so the per-viewer wire filter redacts them for everyone
// else. Marking every seat a knower — which is what "reveal" does, and
// what a copy-paste from SearchLibraryForEffect would give you — would
// hand the whole table the top of a library, which is a real
// information advantage rather than a cosmetic slip.
//
// Nothing moves here. The cards stay on top until the choice is
// answered, so a scry left unanswered leaves the library exactly as it
// was rather than in a half-applied order.
//
// Caller must hold g.mu.
func (g *Game) ScryForEffect(playerID, source uuid.UUID, n int) int {
	return g.ScryThenForEffect(playerID, source, n, nil)
}

// ScryThenForEffect is scry N with a continuation: `after` runs once the
// player has finished putting the cards back, with the library in the
// order they chose.
//
// This exists because of the word "then". Preordain is "Scry 2, THEN
// draw a card" — the card left on top is the card drawn, so the draw
// cannot happen in the same step that queues the choice. Running it
// eagerly would draw one of the very cards the player is still deciding
// about, and would then leave the choice unanswerable, because that
// card is no longer in the library for ResolveScry to put back.
//
// `after` still runs when the scry looked at nothing (an empty
// library): the instruction after "then" is not conditional on the
// scry having had cards to look at. It runs on every other terminal
// outcome too — a scry replaced away entirely still leaves Preordain's
// draw to happen.
//
// A scry is a KEYWORD ACTION (CR 701.22), so this opens the CR 614
// window on it before the prompt is queued: "if you would scry, scry
// that many plus one instead" rewrites the count, and the prompt is
// queued with what the window settled on. A window with two different
// replacements in it PAUSES on a CR 616 ordering prompt, and then
// nothing is queued and the returned count is 0 — the resume queues
// the scry when the order is answered. See keyword_action.go and #976.
//
// Caller must hold g.mu.
func (g *Game) ScryThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	return g.keywordLookAtTopForEffect(KeywordActionScry, PendingChoiceScry, playerID, source, n, after)
}

// SurveilThenForEffect is surveil N with a continuation (CR 701.25):
// the player looks at the top N cards of their library and puts any
// number of them into their graveyard, the rest back on top in any
// order.
//
// Scry's mechanism with the bottom-of-library leg replaced by the
// graveyard — see lookAtTopForEffect for the shared queueing and
// ResolveSurveil for what happens on submit. `after` carries anything
// the card prints after "then", for the same reason Preordain's draw
// does: it must not run until the library is in the order the player
// chose.
//
// A surveil is a keyword action (CR 701.25) on the same CR 614 window
// scry opens — same count, same "that many plus one" shape, same
// pause. See ScryThenForEffect.
//
// Caller must hold g.mu.
func (g *Game) SurveilThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	return g.keywordLookAtTopForEffect(KeywordActionSurveil, PendingChoiceSurveil, playerID, source, n, after)
}

// keywordLookAtTopForEffect is the scry / surveil entry point: it
// opens the CR 614 keyword-action window on the count and, once the
// window settles, queues the prompt through the same
// lookAtTopForEffect body every member of the family uses.
//
// LookAtTopThenForEffect deliberately does NOT come through here.
// "Look at the top N cards of your library, then put them back in any
// order" is not a keyword action — it is a sentence Ponder and
// Sensei's Divining Top print in full — so there is no "if you would
// look at the top" to replace, and routing it through a keyword-action
// window would invent one.
//
// A count of zero or less opens no window either: you would not scry,
// so there is nothing for a replacement to replace. The `after`
// continuation still runs, in lookAtTopForEffect, exactly as it did.
//
// Caller must hold g.mu.
func (g *Game) keywordLookAtTopForEffect(action KeywordAction, kind PendingChoiceKind, playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	if n <= 0 {
		return g.lookAtTopForEffect(kind, playerID, source, n, after)
	}
	looked, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              playerID,
		Source:             source,
		KeywordAction:      action,
		KeywordActionCount: n,
		keywordAction:      &keywordActionTail{choice: kind, then: after},
	})
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    playerID,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
	return looked
}

// LookAtTopThenForEffect is "look at the top N cards of your library,
// then put them back in any order" — Ponder, Sensei's Divining Top,
// Soothsaying.
//
// The scry family's third member, with no away lane: every card goes
// back on top, and the only decision is the order. `after` carries
// anything the card prints next — Ponder's "draw a card", which must
// not run until the player has decided which card is on top, for the
// same reason Preordain's must not.
//
// Caller must hold g.mu.
func (g *Game) LookAtTopThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	return g.lookAtTopForEffect(PendingChoiceLookAtTop, playerID, source, n, after)
}

// lookAtTopForEffect queues the "look at the top N cards of your
// library, then put them somewhere" prompt that scry and surveil
// share. It marks the chooser — and ONLY the chooser — a knower of
// each card, which is what makes both keywords "look at" rather than
// "reveal"; the wire redaction in protocol.FilterViewFor keys off
// exactly that.
//
// Returns how many cards the player is actually looking at, which is
// min(n, library size) and can be zero. `after` runs immediately when
// there is nothing to look at: the instruction after "then" is not
// conditional on the library having had cards in it.
//
// The banner copy comes from the prompt kind (lookAtTopReasonFor), so
// the keyword-action window that may have rewritten `n` cannot leave
// the prompt saying "Scry 2" over three cards.
//
// Caller must hold g.mu.
func (g *Game) lookAtTopForEffect(kind PendingChoiceKind, playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	reason := lookAtTopReasonFor(kind)
	runAfter := func() {
		if after != nil {
			if err := after(g); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    playerID,
					Source:   source,
					ErrorMsg: err.Error(),
				})
			}
		}
	}
	if n <= 0 {
		runAfter()
		return 0
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		runAfter()
		return 0
	}
	size := p.Library.Size()
	if size == 0 {
		runAfter()
		return 0
	}
	if n > size {
		n = size
	}
	// Library top is the LAST element, so the top n cards are the tail
	// — collected top-first so the chooser sees them in draw order.
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		idx := size - 1 - i
		p.Library.Cards[idx].AddKnower(playerID)
		ids = append(ids, p.Library.Cards[idx].InstanceID)
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:       kind,
		Chooser:    playerID,
		FromPlayer: playerID,
		Count:      len(ids),
		Source:     source,
		Reason:     reason(len(ids)),
		ScryCards:  ids,
		scryResume: after,
	})
	return len(ids)
}

// ShuffleLibraryForEffect is "shuffle your library" as a card's own
// instruction (Soothsaying's {3}{U}{U}, Ponder's "you may shuffle")
// rather than as the tail of a search. The lock-holding twin of
// ShuffleLibrary.
//
// It clears KnownBy across the whole zone for the same reason
// finishSearchLocked does: a shuffle is exactly the thing that
// un-knows a library. Anyone who had scryed, tutored or Soothsaid
// their way to knowing where a card was no longer does, and skipping
// that would leave stale knowledge on the wire — which is a real
// information leak, not a cosmetic one.
//
// A missing player is a no-op rather than an error: the instruction
// is "shuffle your library", and a seat that has left the game has
// none to shuffle.
//
// Caller must hold g.mu.
func (g *Game) ShuffleLibraryForEffect(playerID uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		return nil
	}
	p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
	clearKnownInZoneLocked(p.Library)
	// #1335: a plain shuffle is always the player's own library.
	g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: playerID, Target: playerID, Label: "shuffle"})
	return nil
}

// scryReason is the picker's banner copy, phrased as the card prints
// it.
func scryReason(n int) string {
	switch n {
	case 1:
		return "Scry 1"
	case 2:
		return "Scry 2"
	case 3:
		return "Scry 3"
	}
	return "Scry " + strconv.Itoa(n)
}

// surveilReason is the picker's banner copy, phrased as the card
// prints it.
func surveilReason(n int) string {
	switch n {
	case 1:
		return "Surveil 1"
	case 2:
		return "Surveil 2"
	case 3:
		return "Surveil 3"
	}
	return "Surveil " + strconv.Itoa(n)
}

// lookAtTopReason is the picker's banner copy. Unlike scry and
// surveil this is not a keyword, so it is phrased as the cards print
// the instruction.
func lookAtTopReason(n int) string {
	return "Look at the top " + strconv.Itoa(n) + " cards of your library"
}

// lookAtTopReasonFor picks the banner copy for a prompt kind. The
// three prompts differ only in where the cards that leave the top go,
// so the copy is the one thing that has to say which keyword the
// player is answering — and it is chosen from the KIND rather than
// passed in beside it, so the count the prompt is queued with is
// always the count the banner names. An unknown kind gets the neutral
// phrasing rather than a blank banner.
func lookAtTopReasonFor(kind PendingChoiceKind) func(int) string {
	switch kind {
	case PendingChoiceScry:
		return scryReason
	case PendingChoiceSurveil:
		return surveilReason
	}
	return lookAtTopReason
}

// ReturnFromExileToBattlefieldForEffect is the other half of a
// flicker: it takes a card sitting in exile and puts it back onto the
// battlefield. ExileCardForEffect could already send a permanent
// there; until S22 nothing could bring one back, which is why
// "exile it, then return it" had no expression at all.
//
// `controller` is who it returns under — uuid.Nil (or a player who
// has left) means "under its owner's control", which is what almost
// every card says; Thassa's "under your control" passes the
// controller explicitly. `tapped` covers "return that card to the
// battlefield tapped".
//
// The returned permanent is a NEW OBJECT (CR 400.7): it gets a fresh
// InstanceID and none of the old object's battlefield state — no
// counters, no damage, no combat declarations, no summoning-sickness
// or timestamp history, and no lingering exile-play permission. That
// is the rules-correct behaviour and it is what makes blink a removal
// answer (counters fall off, a stolen creature goes home) as well as
// an ETB engine. Returns the new instance ID; a caller that needs to
// keep referring to the permanent must use it, because the pre-exile
// ID now names nothing.
//
// The entry runs through the CR 614 replacement pipeline exactly as
// an ordinary battlefield entry does, so Authority of the Consuls
// taps the blinked creature and a self "enters tapped" replacement
// still applies. The pipeline is consulted BEFORE the card is lifted
// out of exile, so nothing is stranded between zones.
//
// #478: the entry can PAUSE. Two entry replacements on one returning
// permanent queue the CR 616 ordering prompt, and a shockland blinked
// back is asked to pay. Until now the return was DROPPED on that pause
// — the reason #909's Living Death could hang both of its remaining
// passes off one continuation. It is now resumable like every other
// entry: nothing has moved while the question is open, and the return
// completes, with its new object identity, when the answer arrives.
// The call returns uuid.Nil with a nil error in the meantime, which is
// the same "nothing entered" the cancel and redirect branches return.
//
// Both EventETB and the catalog's AsEnters hook fire, so the permanent
// re-triggers everything a fresh entry would.
//
// Caller must hold g.mu. Added in S22.
func (g *Game) ReturnFromExileToBattlefieldForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	return g.returnFromExileToBattlefieldLocked(cardID, controller, tapped, nil)
}

// ReturnFromExileToBattlefieldThenForEffect is the exile return with the
// rest of the effect as a continuation (#1327): `then` is told the
// permanent's NEW instance ID once the entry is complete, or uuid.Nil
// when nothing entered.
//
// The synchronous door cannot answer "did it enter, and under whose
// control?" when the entry pauses — a returning Clone asks what to
// copy, a blinked shockland asks for its life — because it has
// returned uuid.Nil by the time the answer arrives. Phelia, Exuberant
// Shepherd's "if it entered under your control, put a +1/+1 counter on
// Phelia" is that question, and asking it on the next line is wrong
// both ways: no counter on a paused entry, or a counter before an
// entry that could still be cancelled.
//
// `then` runs exactly once, from every terminal outcome: at once when
// nothing paused, from the resume when something did, and with
// uuid.Nil when the card is not in exile (with ErrCardNotFound
// returned), when a replacement cancelled or redirected the entry, or
// when the prompt was taken away. It takes the live *Game, on the
// undo-safety contract every continuation follows.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromExileToBattlefieldThenForEffect(cardID, controller uuid.UUID, tapped bool, then func(g *Game, entered uuid.UUID) error) error {
	_, err := g.returnFromExileToBattlefieldLocked(cardID, controller, tapped, then)
	return err
}

// returnFromExileToBattlefieldLocked is the body of both exile-return
// doors.
//
// Caller must hold g.mu.
func (g *Game) returnFromExileToBattlefieldLocked(cardID, controller uuid.UUID, tapped bool, then func(g *Game, entered uuid.UUID) error) (uuid.UUID, error) {
	if g.Exile == nil || !g.Exile.Contains(cardID) {
		if then != nil {
			if err := then(g, uuid.Nil); err != nil {
				return uuid.Nil, errors.Join(ErrCardNotFound, err)
			}
		}
		return uuid.Nil, ErrCardNotFound
	}
	// Resolve the destination controller before the pipeline runs: a
	// replacement that asks "is the entering permanent mine?"
	// (Authority of the Consuls) reads Card.Controller, and while the
	// card sits in exile that field still names whoever controlled it
	// before it left.
	newController := controller
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID != cardID {
			continue
		}
		if newController == uuid.Nil || g.playerByIDLocked(newController) == nil {
			newController = g.Exile.Cards[i].Owner
		}
		g.Exile.Cards[i].Controller = newController
		break
	}
	// The effect's own "return it TAPPED" clause is seeded onto the
	// event rather than OR-ed in after the pipeline, the way
	// SearchLibrarySpec.TappedOnEntry and ZoneEntryOptions.Tapped are:
	// a resume reads ev.EntersTapped and has no idea what effect sent
	// the card, so a Thassa blink that paused would otherwise come back
	// untapped.
	return g.enterBattlefieldThroughPipelineLocked(&ReplacementEvent{
		Kind:           RepEventMove,
		Actor:          newController,
		CardID:         cardID,
		OldZone:        ZoneExile,
		NewZone:        ZoneBattlefield,
		EntersTapped:   tapped,
		entryResumable: true,
		// CR 400.7: the returning permanent is a NEW OBJECT. The
		// finisher mints the ID and strips the old object's
		// battlefield state, on the inline path and the resumed one
		// alike — which is what the old "an exile return cannot be
		// resumed generically" note was about.
		//
		// #1327: and the caller's continuation, run from the landing
		// with the NEW ID, whether the entry settled inline or on a
		// resume.
		entryTail: &entryTail{newObject: true, then: then},
	})
}

// HandEntryOptions are the modifiers a "put a card from your hand
// onto the battlefield" effect applies to the entry it causes. Since
// #745 it is the shared ZoneEntryOptions (battlefield_put.go), because
// the library move takes exactly the same two modifiers; the name is
// kept so a hand caller reads as one.
type HandEntryOptions = ZoneEntryOptions

// PutFromHandOntoBattlefieldForEffect puts one card from its owner's
// hand onto the battlefield without casting or playing it — "you may
// put a land card from your hand onto the battlefield" (Growth
// Spiral, Eureka Moment, Broken Bond, Chulane), "put an Equipment
// card from your hand onto the battlefield" (Stoneforge Mystic).
//
// It is the MOVE half of the seam docs/engine-seams.md calls
// "Put-from-hand onto the battlefield" (#654). The PICK half already
// existed: ChooseCardsPrompt with Zone: ZoneHand and a Then
// continuation (#552). The prompt, the "you may" and the per-card
// filter stay in the catalog (effects.PutFromHandOntoBattlefield),
// because those are card text; the move is engine.
//
// The shape is deliberately the one every other effect-side entry
// uses — searchEnterBattlefieldLocked, the reanimation path in
// ReturnFromGraveyardUnderControlForEffect, and
// ReturnFromExileToBattlefieldForEffect:
//
//  1. Stamp the controller on the card WHILE IT IS STILL IN HAND.
//     Authority of the Consuls asks whose permanent is entering and a
//     self-replacement's condition counts that player's lands; both
//     read Card.Controller, which in hand still names whoever it was
//     stamped with last.
//  2. Run the CR 614 pipeline BEFORE the card leaves the hand, so a
//     CR 616 ordering prompt — which pauses — leaves the card where
//     it is rather than stranding it between zones.
//  3. Move, then apply the settled EntersTapped / EntersWithCounters.
//  4. EventZoneMove, EventETB, then the catalog's AsEnters hook, so
//     the permanent triggers everything an ordinary entry triggers.
//
// Two things it deliberately does NOT do:
//
//   - It does not count a land drop. A land an effect puts onto the
//     battlefield was not PLAYED (CR 305.4), so LandsPlayedThisTurn
//     is untouched and the enumerator still offers the turn's land
//     play. The event carries no landPlay flag, which is the only
//     thing the finisher bumps the tally on (#478).
//   - It does not run state checks. Like its neighbours it leaves
//     them to the resolution boundary the caller is already inside,
//     so a permanent that enters and dies does so once, together
//     with everything else that effect did.
//
// It DOES offer the entry's own questions since #1322 — the Clone
// choice, a shockland's life, a reveal-land's reveal — because the put
// batch it runs through is resumable now (entry_batch.go). On such a
// pause it returns uuid.Nil and the card lands when the answer
// arrives; PutFromHandOntoBattlefieldThenForEffect is the door for a
// caller that needs the result.
//
// Returns the entering permanent's ID. It is the InstanceID the card
// had in hand: every non-exile move keeps it, and only the exile
// return mints a new one. uuid.Nil with a nil error means nothing
// entered — a replacement canceled or redirected the move, or the
// pipeline paused — which is a legal outcome, not a failure.
//
// Refuses, with ErrCardNotFound, a card that is not in a hand: this
// is a hand move, and a caller holding a stale ID must not silently
// rip a permanent off the battlefield or a card out of a graveyard.
// A nonpermanent card is refused with ErrInvalidParam — CR 110.4
// lists the permanent types and an instant is not among them, so the
// card stays in hand.
//
// Caller must hold g.mu.
func (g *Game) PutFromHandOntoBattlefieldForEffect(cardID uuid.UUID, opts HandEntryOptions) (uuid.UUID, error) {
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{cardID}, ZoneHand, opts)
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}

// addManaReason is the prompt header for an effect's mana pick: the
// ordinary "one mana of any color", or the "N mana of any one color"
// form (#742) when the slot adds more than one.
func addManaReason(slot ProducedManaEntry) string {
	if slot.OneColorAmounts() {
		return "Add mana of any one color"
	}
	return "Add one mana of any color"
}

// AddManaForEffect adds the mana a SPELL or a non-mana ability
// produces to playerID's pool — Dark Ritual's "Add {B}{B}{B}", Mana
// Drain's delayed "add an amount of {C}", Jeska's Will. Every other
// mana in the engine arrives through ActivateManaAbility (CR 605); a
// spell that adds mana resolves off the stack like any other spell
// and lands its mana here, in the same resolution frame, with the
// same EventManaAdded per unit the mana-ability path emits.
//
// `produced` is the Scryfall brace grammar ParseProducedMana reads,
// pipe syntax included: a single-colour slot goes straight into the
// pool, a multi-option slot queues the same PendingChoiceMana pick a
// Birds of Paradise activation does, with the controller's commander
// identity listed first (manaPickOptionsFor).
// `source` is the card the mana is attributed to (the spell itself
// for Dark Ritual); it rides on each ManaToken.
//
// The pool still empties at the end of the step (CR 106.4), so mana
// added by a spell has to be spent in the step it resolved in — the
// printed behaviour, and the reason Dark Ritual is cast in a main
// phase.
//
// An eliminated or unseated player gets nothing and no error: the
// spell resolved, there was just nobody to give the mana to.
//
// Caller must hold g.mu.
func (g *Game) AddManaForEffect(playerID, source uuid.UUID, produced string) error {
	return g.AddManaWithOptionsForEffect(playerID, source, produced, AddManaOptions{})
}

// AddManaOptions tunes AddManaWithOptionsForEffect.
type AddManaOptions struct {
	// NarrowToCommanderIdentity intersects a multi-option pick with the
	// controller's commander colour identity instead of offering the
	// printed width identity-first: the effect-side twin of
	// ManaAbilityShape.NarrowToCommanderIdentity. Set it only when the
	// printed text says "any color in your commander's color
	// identity"; no effect in the catalog does today. Replaced #742's
	// IgnoreCommanderIdentity when the default flipped (owner decision
	// 2026-09-17). CR 903.4f (#844): with no commander, or a colourless
	// one, such a pick adds nothing and is not queued.
	NarrowToCommanderIdentity bool
}

// AddManaWithOptionsForEffect is AddManaForEffect with options.
// AddManaForEffect is this with the zero options: the printed option
// set, commander identity first.
//
// Caller must hold g.mu.
func (g *Game) AddManaWithOptionsForEffect(playerID, source uuid.UUID, produced string, opts AddManaOptions) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return nil
	}
	return g.addManaSlotsLocked(p, source, produced, opts.NarrowToCommanderIdentity, nil, addManaReason)
}

// addManaSlotsLocked is the one slot walk behind every "an effect adds
// mana" path: AddManaWithOptionsForEffect above, and #763's triggered
// mana abilities (addTriggeredManaLocked). One body rather than two,
// so a pipe from a trigger queues exactly the pick a spell's pipe
// queues and nothing has to be kept in step by hand.
//
// It has two modes, and they are the two the mana pipeline already
// had for a multi-option slot:
//
//   - `pending` nil — queue a PendingChoiceMana, the way
//     ActivateManaAbility and a resolving spell do. The player picks.
//   - `pending` non-nil — the AUTO-TAP executor's remaining colour
//     requirements: pick greedily against them with pickColorForSlot,
//     the same function materializePlanLocked uses for a tapped
//     source's own slots, and book the requirement the pick pays. The
//     auto-tapper's contract is "no further player decisions", so it
//     may not leave a prompt behind mid-cast.
//
// `reason` labels the queued prompt. Caller must hold g.mu in write
// mode and have checked the player is seated and not eliminated.
func (g *Game) addManaSlotsLocked(
	p *Player,
	source uuid.UUID,
	produced string,
	narrow bool,
	pending *[]ColorRequirement,
	reason func(ProducedManaEntry) string,
) error {
	slots, err := ParseProducedMana(produced)
	if err != nil {
		return err
	}
	// Read once, and only when a slot could use it: the identity read
	// walks every zone, and Dark Ritual's three "{B}" slots have
	// nothing to narrow or order.
	var identity commanderIdentity
	if narrow || hasMultiOptionSlot(slots) {
		identity = commanderIdentityFor(g, p)
	}
	// #1212: what the source is, if it is a permanent at all. A
	// resolving Dark Ritual is not one, and zero is the honest answer
	// for it — "mana from a Treasure" is a question about a
	// permanent's ability (CR 605.1a). A triggered mana ability's
	// source (#763, Wild Growth on a Forest) IS a permanent, and this
	// is where its kinds are recorded.
	srcKinds := g.manaSourceKindsLocked(source)
	for _, slot := range slots {
		colorOptions := manaPickOptions(slot.Options, identity, narrow)
		if len(colorOptions) == 0 {
			// CR 903.4f (#844): a "commander's color identity" effect
			// with no identity adds nothing, and an empty picker is not
			// a choice anybody can answer. An empty printed slot lands
			// here too, as it always did.
			continue
		}
		if len(slot.Options) == 1 {
			// The caller's booking for the mana the effect PRINTS; the
			// production body below books whatever a replacement ADDS.
			if pending != nil {
				bookColorRequirement(colorOptions[0], pending)
			}
			// #1222: through the one production body, which opens the
			// CR 106.12b window on the amount. fromTap is false — this
			// is a SPELL's "Add {B}{B}{B}" or a triggered mana ability's
			// own output (ADR 0074), and neither taps a permanent for
			// mana, so neither is doubled by Mana Reflection. That is
			// what the card says, not a simplification.
			g.produceManaLocked(p, source, []string{colorOptions[0]}, nil, srcKinds, false, pending)
			continue
		}
		if pending != nil {
			// The auto-tap mode: one greedy pick, minted now. A
			// one-colour-N-mana slot mints all N of it (#742).
			color := pickColorForSlot(colorOptions, pending)
			if color == "" {
				continue
			}
			for k := 1; k < slot.AmountFor(color); k++ {
				bookColorRequirement(color, pending)
			}
			g.produceManaLocked(p, source, repeatColor(color, slot.AmountFor(color)), nil, srcKinds, false, pending)
			continue
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:            PendingChoiceMana,
			Chooser:         p.ID,
			FromPlayer:      p.ID,
			Count:           1,
			Source:          source,
			Reason:          reason(slot),
			ColorOptions:    colorOptions,
			ManaAmounts:     copyManaAmounts(slot.Amounts),
			ManaSourceKinds: srcKinds,
		})
	}
	return nil
}

// SetMaxHandSizeForEffect is SetMaxHandSize's lock-free twin, for a
// SPELL that grants "you have no maximum hand size for the rest of
// the game" (Finale of Revelation) rather than a permanent's static
// ability (Reliquary Tower's Spec.NoMaxHandSize, which is derived
// from the battlefield at cleanup and lapses the moment the
// permanent does — wrong for a effect that has to outlive the spell
// that granted it, and outlive every permanent on the board).
//
// Writes the same Player.MaxHandSize field SetMaxHandSize does, and
// EffectiveMaxHandSizeLocked checks it FIRST, before consulting the
// battlefield — so this is a permanent per-player grant with no
// battlefield dependency, exactly as the printed clause reads.
//
// `value` is clamped to NoMaxHandSize (-1) for "no cap"; any other
// negative value is rejected with ErrInvalidParam. Caller must hold
// g.mu.
func (g *Game) SetMaxHandSizeForEffect(playerID uuid.UUID, value int) error {
	if value < NoMaxHandSize {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.MaxHandSize = value
	return nil
}
