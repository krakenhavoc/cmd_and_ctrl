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
		return Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return Card{}, false
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
	Then func(g *Game) error
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
// Caller must hold g.mu.
func (g *Game) QueueDiscardChoiceForEffect(p DiscardPrompt) uuid.UUID {
	// A prompt addressed to a seat that has left the game can never
	// be answered. Unlike the empty-hand case, Then does NOT run:
	// "each other player discards a card, then you draw a card for
	// each card discarded this way" draws nothing for a player who is
	// no longer there.
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
		if p.Then != nil {
			if err := p.Then(g); err != nil {
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
	question := p.Question
	if question == "" {
		question = g.discardQuestionLocked(p.Source, n, p.UpTo)
	}
	discarder := p.Player
	opts := discardOptions{cause: discardCauseEffect, source: p.Source, then: p.Then}
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
		damageTail: g.effectDamageTailLocked(damageTailPlayer, source),
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
		damageTail: g.effectDamageTailLocked(damageTailPermanent, source),
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
func (g *Game) DiscardRandomForEffect(playerID uuid.UUID, n int) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if n <= 0 || p.Hand.Size() == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(p.Hand.Cards))
	for i, c := range p.Hand.Cards {
		ids[i] = c.InstanceID
	}
	picked := g.ChooseAtRandomForEffect(RandomDraw{Player: playerID}, ids, n)
	return g.discardCardsLocked(playerID, picked, discardOptions{cause: discardCauseEffect})
}

// LoseTheGameForEffect marks a player as losing the game — the
// consequence half of "pay {3}{U}{U}. If you don't, you lose the
// game" (Pact of Negation and the rest of the Pact cycle), and of
// every other card that says those words outright.
//
// Routed through the AttemptedEmptyDraw flag rather than eliminating
// the player on the spot, so the ability finishes resolving first and
// the loss lands at the next SBA check alongside the empty-library
// and zero-life losses. That borrows the CR 704.5b flag for a loss
// that is not a draw, and CR 104.3e actually makes an effect loss
// immediate; ADR 0057 sub-PR 2 replaces this writer with
// loseGameLocked, after which actuallyDrawCardLocked is the flag's
// only writer.
//
// Caller must hold g.mu. Added in S28.
func (g *Game) LoseTheGameForEffect(playerID uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.AttemptedEmptyDraw = true
	return nil
}

// MillNForEffect moves n cards from the top of playerID's library
// to their graveyard. Emits EventMill per card. A library holding
// fewer than n mills what it has (CR 701.17b) and nobody loses for
// it — see MillToZoneForEffect.
func (g *Game) MillNForEffect(playerID uuid.UUID, n int) error {
	_, err := g.MillToZoneForEffect(playerID, n, ZoneGraveyard, nil)
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
// `until`, when non-nil, is consulted for each card as it comes off
// the library and stops the mill AFTER the first card it accepts;
// that is Helm of Obedience's "mills cards until a creature card is
// put into their graveyard" — the card that ends it still moves. With
// `until` set, n <= 0 means "no limit but the library", so an
// unbounded mill is expressible without inventing a sentinel.
//
// Running the library out stops the run, with no error and NO loss.
// CR 701.17b: a player instructed to mill more cards than their
// library holds "mill[s] as many as possible", and only an attempt to
// DRAW from an empty library loses the game (CR 704.5b, CR 121.4).
// The same holds for "exile the top N cards" and for an `until` run
// that never finds its card — both simply end when the library does,
// exactly as ExileTopFaceDownForEffect stops on an empty library.
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
// Caller must hold g.mu.
func (g *Game) MillToZoneForEffect(playerID uuid.UUID, n int, dest ZoneKind, until func(Card) bool) ([]uuid.UUID, error) {
	ids, err := g.millPlanLocked(playerID, n, dest, until)
	if err != nil {
		return nil, err
	}
	return g.routeAllLandedLocked(millRoute(playerID, dest), ids), nil
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
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) MillToZoneThenForEffect(
	playerID uuid.UUID,
	n int,
	dest ZoneKind,
	until func(Card) bool,
	then func(g *Game, milled []uuid.UUID) error,
) error {
	ids, err := g.millPlanLocked(playerID, n, dest, until)
	if err != nil {
		return err
	}
	return g.routeAllThenLocked(millRoute(playerID, dest), ids, then)
}

// millPlanLocked validates a mill and chooses the cards it will move,
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
// `until` is answered here too, against the same pre-move copies the
// old loop consulted: it reads the card that came off the library, not
// where that card ended up, so the run it ends is the same run whether
// it is measured before the first move or after the last. That is what
// lets the plan be a plain list of IDs — which is what the batch body
// takes — and it is also the one thing Helm of Obedience's caveat is
// about: a commander whose owner takes the command zone was never put
// into a graveyard, so by CR 701.17a's letter the run should continue,
// and it stops instead.
//
// Caller must hold g.mu.
func (g *Game) millPlanLocked(playerID uuid.UUID, n int, dest ZoneKind, until func(Card) bool) ([]uuid.UUID, error) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	switch dest {
	case ZoneGraveyard, ZoneExile:
	default:
		return nil, ErrInvalidParam
	}
	unbounded := until != nil && n <= 0

	avail := len(p.Library.Cards)
	want := n
	if unbounded || want > avail {
		want = avail
	}
	ids := make([]uuid.UUID, 0, want)
	for i := 0; i < want; i++ {
		c := p.Library.Cards[avail-1-i]
		ids = append(ids, c.InstanceID)
		if until != nil && until(c) {
			break
		}
	}
	return ids, nil
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
func (g *Game) DestroyPermanentForEffect(cardID uuid.UUID) error {
	return g.destroyBattlefieldPermanentLocked(cardID)
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
// the player who exiled it — so its knowledge set is CLEARED on the
// way in and Card.FaceDown is set. An empty knowledge set is what
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
			FaceDown: true,
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
	item, ok := g.StackMeta[spellID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return ErrCardNotOnStack
	}
	if g.Stack == nil || !g.Stack.Contains(spellID) {
		return ErrCardNotOnStack
	}
	// Resolve the destination. nil → the spell's owner's graveyard,
	// or exile when that player has left the game (the primitive's
	// own fallback). Battlefield / stack are illegal — a counter that
	// "puts the spell onto the battlefield" would be a different
	// effect, and a counter MUST move the spell off the stack.
	destKind, destOwner := ZoneGraveyard, item.Owner
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
	// the counter's chosen destination, so a flashed-back spell
	// answered by Hinder is exiled rather than shuffled away — which
	// is the whole difference between a replacement effect and an
	// exile bolted onto the resolution, and the only place in the
	// engine where it is observable.
	for _, c := range g.Stack.Cards {
		if c.InstanceID == spellID && altCostExilesFromStack(c, item.AltCost) {
			destKind, destOwner = ZoneExile, uuid.Nil
			break
		}
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:        spellID,
		Dst:           destKind,
		DstOwner:      destOwner,
		Countered:     true,
		DropStackMeta: true,
	})
	return err
}

// counterAbilityLocked is the lock-free body of CounterAbility.
// Caller must hold g.mu.
func (g *Game) counterAbilityLocked(abilityID uuid.UUID) error {
	item, ok := g.StackMeta[abilityID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	if item.Kind != StackItemActivated && item.Kind != StackItemTriggered {
		return ErrCardNotOnStack
	}
	source := item.SourceCardID
	delete(g.StackMeta, abilityID)
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Source: source,
		Target: abilityID,
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
	if delta == 0 {
		return nil
	}
	if name == "" {
		return ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterTarget: cardID,
		CounterName:   name,
		CounterDelta:  delta,
	}
	out, err := g.applyReplacementsLocked(ev)
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		// errReplacementPending: the prompt queue is the caller's
		// problem; return nil so the primitive keeps moving.
		if errors.Is(err, errReplacementPending) {
			return nil
		}
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyCounterLocked(out.CounterTarget, out.CounterName, out.CounterDelta)
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
	src := g.findCardZoneLocked(cardID)
	if src == nil || src.Kind != ZoneGraveyard {
		return ErrCardNotFound
	}
	// Identify the graveyard's owner so we can route to the same
	// player's hand / library. (Graveyards are per-player; Zone.Owner
	// holds the ID.)
	ownerID := src.Owner
	owner := g.playerByIDLocked(ownerID)
	if owner == nil {
		return ErrPlayerNotFound
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
		return ErrZoneNotFound
	}
	actor := ownerID
	var entersTapped bool
	var enterCounters map[string]int
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
		actor = newController
		// #263: a reanimated permanent enters the battlefield like
		// any other, so the CR 614 entry pipeline has to run on it.
		// Skipping it meant a reanimated shockland ignored its own
		// enters-tapped clause and a creature reanimated under an
		// opponent's nose dodged Authority of the Consuls.
		ev := &ReplacementEvent{
			Kind:    RepEventMove,
			Actor:   newController,
			CardID:  cardID,
			OldZone: ZoneGraveyard,
			NewZone: ZoneBattlefield,
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			// A CR 616 ordering prompt is open. Nothing has moved —
			// the card is still in the graveyard — so the return is
			// dropped rather than stranded mid-zone. Same posture the
			// land-play and exile-return paths take.
			return nil
		}
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			g.clearReplacementEventLocked(ev.ID)
			return err
		}
		defer g.clearReplacementEventLocked(ev.ID)
		if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
			return nil
		}
		entersTapped = out.EntersTapped
		enterCounters = out.EntersWithCounters
	}
	if _, err := MoveCard(src, destZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	var oracleID string
	if destZone.Kind == ZoneBattlefield {
		for i := range destZone.Cards {
			if destZone.Cards[i].InstanceID == cardID {
				destZone.Cards[i].Controller = actor
				if entersTapped {
					destZone.Cards[i].Tapped = true
				}
				oracleID = CatalogKey(destZone.Cards[i])
				break
			}
		}
		for name, n := range enterCounters {
			_ = g.AddCounterForEffect(cardID, name, n)
		}
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   actor,
		CardID:  cardID,
		OldZone: ZoneGraveyard,
		NewZone: destZone.Kind,
	})
	if destZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, Actor: actor, CardID: cardID})
		// A reanimated permanent enters the battlefield like any
		// other, so the catalog's AsEnters / StartingLoyalty hook has to
		// run — otherwise reanimating Solemn Simulacrum fetches
		// nothing and reanimating a planeswalker gives it no loyalty.
		// Every other path onto the battlefield already fires this;
		// this one was the omission. A no-op for a card with no
		// oracle ID (tokens, fixtures) or no catalog entry.
		g.fireETBHookLocked(cardID, oracleID)
	}
	return nil
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

	// Dest is one of ZoneHand, ZoneBattlefield, ZoneLibrary.
	Dest ZoneKind

	// Limit is the maximum number of cards taken. <= 0 means 1.
	Limit int

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
	p := g.playerByIDLocked(spec.Player)
	if p == nil {
		return ErrPlayerNotFound
	}
	if spec.Limit <= 0 {
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
	if !spec.Optional && len(matches) <= spec.Limit &&
		(spec.Validate == nil || spec.Validate(matched)) {
		// No decision to make — take them all.
		found := g.executeSearchTakeLocked(spec, p, matches)
		return g.finishSearchLocked(spec, p, found)
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
// and returns the IDs that actually made it. Shared by the
// no-decision path and by ResolveSearchLibrary, so a fetched
// permanent enters identically either way.
//
// Caller must hold g.mu.
func (g *Game) executeSearchTakeLocked(spec SearchLibrarySpec, p *Player, ids []uuid.UUID) []uuid.UUID {
	destZone, err := g.searchDestZoneLocked(p, spec.Dest)
	if err != nil || destZone == p.Library {
		// ZoneLibrary means the card does not leave the library.
		// With ToTop set it will be moved to the top AFTER the
		// shuffle, in finishSearchLocked — so the IDs are returned
		// as found, and nothing is moved here. Without it the clause
		// is "leave it where it is" and only the reveal applies.
		if err != nil {
			return nil
		}
		if spec.Reveal {
			g.revealLibraryCardsLocked(spec, p, ids)
		}
		if spec.ToTop {
			return ids
		}
		return nil
	}
	// S22: the reveal happens ONCE, before anything moves, and names
	// every card the search took. Cultivate's "reveal those cards" is
	// one announcement about two lands rather than two announcements
	// about one each, and revealing before the move is what the card
	// prints — the table sees the cards where they were found.
	if spec.Reveal {
		g.revealLibraryCardsLocked(spec, p, ids)
	}
	found := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !p.Library.Contains(id) {
			continue
		}
		if destZone.Kind == ZoneBattlefield {
			moved, ok := g.searchEnterBattlefieldLocked(spec, p, id)
			if !ok {
				continue
			}
			found = append(found, moved)
			continue
		}
		if _, err := MoveCard(p.Library, destZone, id); err != nil {
			continue
		}
		g.markCardKnownInZoneLocked(destZone, id)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   spec.Player,
			CardID:  id,
			OldZone: ZoneLibrary,
			NewZone: destZone.Kind,
		})
		found = append(found, id)
	}
	return found
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
// the same reason the exile-return path does it: a CR 616 ordering
// prompt means bailing, and bailing with the card already lifted out
// of its zone would strand it. On that bail this card is simply not
// found — the same posture the land-play path takes when its own
// pipeline call pauses.
//
// DECLARED SIMPLIFICATION — this event is NOT entryResumable, so a
// fetched permanent never gets an entry prompt.
//
// The concrete case is a fetchland cracking for a shockland. On the
// play path the controller is asked "pay 2 life so it enters
// untapped?"; fetched, they are not asked, and the land enters
// tapped. Weaker than printed, never stronger — which is exactly the
// posture ReplacementEvent.entryResumable exists to enforce, and a
// strict improvement on the old behaviour, where a fetched shockland
// ignored its entry clause altogether and arrived untapped for free.
//
// Setting entryResumable here would be actively worse, not better.
// executeEntryToBattlefieldLocked can finish the MOVE, but it knows
// nothing about the search that started it: the library would never
// be shuffled, EventSearchLibrary would never fire, and the Then
// continuation — Fabled Passage's "untap that land", Gamble's random
// discard — would never run. A missing shuffle is worse than a
// missing prompt, because it silently leaks library order.
//
// A faithful version needs the search's own continuation to survive
// the entry prompt: a second frame stacked under the replacement
// one, plus a hook in the entry resume to run it. That is its own
// change, not a rider on this one.
//
// Returns the moved card's ID and whether it moved.
//
// Caller must hold g.mu.
func (g *Game) searchEnterBattlefieldLocked(spec SearchLibrarySpec, p *Player, id uuid.UUID) (uuid.UUID, bool) {
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
	ev := &ReplacementEvent{
		Kind:    RepEventMove,
		Actor:   spec.Player,
		CardID:  id,
		OldZone: ZoneLibrary,
		NewZone: ZoneBattlefield,
		// The fetching effect's own "put it onto the battlefield
		// TAPPED" clause is seeded onto the event rather than OR-ed
		// in after the pipeline. Same result for the inline path, and
		// it is the difference between right and wrong for any resume
		// path: a resume reads ev.EntersTapped and has no idea what
		// spell sent the card, so a Cultivate-fetched land that
		// paused for a prompt would otherwise come back untapped.
		EntersTapped: spec.TappedOnEntry,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return uuid.Nil, false
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return uuid.Nil, false
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		return uuid.Nil, false
	}
	moved, err := MoveCard(p.Library, g.Battlefield, id)
	if err != nil {
		return uuid.Nil, false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != moved.InstanceID {
			continue
		}
		g.Battlefield.Cards[i].Controller = spec.Player
		// out.EntersTapped carries both inputs: the fetching effect's
		// printed "tapped" clause, seeded onto the event above, and
		// whatever the CR 614 pipeline added on top.
		if out.EntersTapped {
			g.Battlefield.Cards[i].Tapped = true
		}
		break
	}
	g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
	for name, n := range out.EntersWithCounters {
		_ = g.AddCounterForEffect(moved.InstanceID, name, n)
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   spec.Player,
		CardID:  moved.InstanceID,
		OldZone: ZoneLibrary,
		NewZone: ZoneBattlefield,
	})
	g.EmitEvent(Event{Kind: EventETB, Actor: spec.Player, CardID: moved.InstanceID})
	// Same omission as the reanimation path had: a fetched permanent
	// enters like any other, so its catalog AsEnters hook runs.
	g.fireETBHookLocked(moved.InstanceID, CatalogKey(moved))
	return moved.InstanceID, true
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
		Kind:   EventSearchLibrary,
		Actor:  spec.Player,
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
			p.Library.PushTop(c)
		}
	}
	if spec.Then != nil {
		return spec.Then(g, found)
	}
	return nil
}

// CreateTokenForEffect puts n freshly-minted tokens onto the
// battlefield under `controller`'s control. Each token has a new
// InstanceID, empty ScryfallID (tokens aren't in the Scryfall-
// printing index), and KnownBy pre-populated with every seated
// player (tokens are always public). The emitted EventTokenCreated
// references the first token ID — callers loop for the others via
// the event stream.
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
	// attacking" minus the attacking half, which needs combat state
	// this struct deliberately does not touch.
	Tapped bool

	// Counters are the counters each token enters with, keyed by
	// counter name. Applied before EventETB fires, so an ETB watcher
	// and the P/T recompute both see the finished object.
	//
	// Declared gap: these do NOT run the CR 614 counter replacement
	// pipeline, so a Doubling Season does not double them. Token
	// creation does not go through the zone-move pipeline at all
	// (the token has no previous zone to move from), which is the
	// same reason Tapped is a field here rather than the
	// RepEventMove.EntersTapped an ordinary permanent uses.
	Counters map[string]int

	// Keywords are granted on top of the template's printed ones —
	// the "…with haste" half of a card that pumps the token it
	// makes. Additive: the template's own keywords are kept.
	Keywords []string
}

// CreateTokensForEffect is CreateTokenForEffect with entry options,
// returning the instance IDs of the tokens it made in creation
// order. The IDs are what a card needs when the token is not the end
// of the sentence — "create a token, then sacrifice it", "…then put
// a counter on it".
//
// Caller must hold g.mu.
func (g *Game) CreateTokensForEffect(controller uuid.UUID, template Card, n int, opts TokenEntryOptions) ([]uuid.UUID, error) {
	if n <= 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		tok := template
		tok.InstanceID = uuid.New()
		tok.Owner = controller
		tok.Controller = controller
		tok.Counters = nil
		tok.LostLastCounter = false
		tok.KnownBy = nil
		if opts.Tapped {
			// Additive, not an assignment: a caller that pre-stamped
			// Tapped on the template (the older idiom — Mary Read's
			// Treasure, Hashaton's Zombie) must keep entering tapped.
			tok.Tapped = true
		}
		if len(opts.Counters) > 0 {
			tok.Counters = make(map[string]int, len(opts.Counters))
			for name, count := range opts.Counters {
				if count > 0 {
					tok.Counters[name] = count
				}
			}
		}
		if len(opts.Keywords) > 0 {
			// Fresh slice: the template's Keywords slice is shared by
			// every token minted from it, so appending in place would
			// leak the grant onto the next one.
			kw := make([]string, 0, len(template.Keywords)+len(opts.Keywords))
			kw = append(kw, template.Keywords...)
			kw = append(kw, opts.Keywords...)
			tok.Keywords = kw
		}
		for _, seat := range g.Seats {
			tok.AddKnower(seat.ID)
		}
		// CR 506.3c / #859: a template that arrives already attacking
		// (Adeline's "tapped and attacking Human") was PUT onto the
		// battlefield attacking, not declared, so it announces no
		// EventAttack — and the attack declaration's lock-in must not
		// mistake its AttackingTarget for a staged declaration.
		if tok.AttackingTarget != uuid.Nil {
			g.noteAttackAnnouncedLocked(tok.InstanceID)
		}
		g.Battlefield.PushTop(tok)
		ids = append(ids, tok.InstanceID)
		g.EmitEvent(Event{
			Kind:   EventTokenCreated,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
	}
	return ids, nil
}

// CreateTokensAttackingForEffect is CreateTokenForEffect for
// "create N tokens … that are attacking" (Parhelion II, Hanweir
// Garrison): the tokens are put onto the battlefield already
// attacking `defender`.
//
// CR 506.3c is the reason this is a separate entry point rather than
// a flag: a permanent PUT onto the battlefield attacking was never
// DECLARED as an attacker, so it fires no "whenever ~ attacks"
// trigger and nothing that watches attack declarations sees it.
// Setting AttackingTarget directly and emitting no EventAttack is
// exactly that rule, and routing through DeclareAttacker — which
// emits the event, checks summoning sickness and taps — would be
// wrong on all four counts.
//
// A zero `defender`, or a defender that is not a seated player,
// creates the tokens untapped and not attacking rather than
// erroring: the ability that called this has already resolved, and
// the tokens are the part of it that can still be delivered.
//
// Caller must hold g.mu.
func (g *Game) CreateTokensAttackingForEffect(controller uuid.UUID, template Card, n int, defender uuid.UUID) error {
	if n <= 0 {
		return nil
	}
	attacking := defender != uuid.Nil && g.playerByIDLocked(defender) != nil
	for i := 0; i < n; i++ {
		tok := template
		tok.InstanceID = uuid.New()
		tok.Owner = controller
		tok.Controller = controller
		tok.Counters = nil
		tok.LostLastCounter = false
		tok.KnownBy = nil
		if attacking {
			tok.AttackingTarget = defender
			// CR 506.3c, again: never declared, so the lock-in
			// (attackers.go) skips it rather than announcing it at
			// the next priority boundary.
			g.noteAttackAnnouncedLocked(tok.InstanceID)
		}
		for _, seat := range g.Seats {
			tok.AddKnower(seat.ID)
		}
		g.Battlefield.PushTop(tok)
		g.EmitEvent(Event{
			Kind:   EventTokenCreated,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  controller,
			CardID: tok.InstanceID,
		})
	}
	return nil
}

// ScryForEffect performs a scry N (CR 701.22): the player looks at the
// top N cards of their library and then decides which go to the bottom
// and in what order the rest go back on top.
//
// Returns the number of cards actually looked at, which is fewer than n
// when the library is short and zero when it is empty — a scry with an
// empty library is not an error, it simply does nothing, and no choice
// is queued.
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
// scry having had cards to look at.
//
// Caller must hold g.mu.
func (g *Game) ScryThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	return g.lookAtTopForEffect(PendingChoiceScry, playerID, source, n, scryReason, after)
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
// Caller must hold g.mu.
func (g *Game) SurveilThenForEffect(playerID, source uuid.UUID, n int, after func(g *Game) error) int {
	return g.lookAtTopForEffect(PendingChoiceSurveil, playerID, source, n, surveilReason, after)
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
	return g.lookAtTopForEffect(PendingChoiceLookAtTop, playerID, source, n, lookAtTopReason, after)
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
// Caller must hold g.mu.
func (g *Game) lookAtTopForEffect(kind PendingChoiceKind, playerID, source uuid.UUID, n int, reason func(int) string, after func(g *Game) error) int {
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
	g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: playerID, Label: "shuffle"})
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
// out of exile: if it queues a CR 616 ordering prompt the card has
// not moved yet, so the return simply doesn't happen rather than
// stranding the card between zones. (That path needs two competing
// replacements on one entry; no card in the catalog produces it
// today.)
//
// Both EventETB and the catalog's AsEnters hook fire, so the permanent
// re-triggers everything a fresh entry would.
//
// Caller must hold g.mu. Added in S22.
func (g *Game) ReturnFromExileToBattlefieldForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	if g.Exile == nil || !g.Exile.Contains(cardID) {
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
	ev := &ReplacementEvent{
		Kind:    RepEventMove,
		Actor:   newController,
		CardID:  cardID,
		OldZone: ZoneExile,
		NewZone: ZoneBattlefield,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A CR 616 ordering prompt is open. Nothing has moved; the
		// card stays in exile and the return is dropped.
		return uuid.Nil, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return uuid.Nil, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		// Canceled, or redirected elsewhere by a replacement. There
		// is no generic "put it wherever the pipeline said" helper
		// for an exile source, so a redirect is treated as a cancel
		// rather than guessed at.
		return uuid.Nil, nil
	}
	card, err := g.Exile.Remove(cardID)
	if err != nil {
		return uuid.Nil, err
	}
	newID := uuid.New()
	// A blink gets a new instance ID but remains the same random source.
	if ordinal, ok := g.sourceOrdinals[cardID]; ok {
		g.sourceOrdinals[newID] = ordinal
	}
	card.InstanceID = newID
	card.Controller = newController
	card.Tapped = tapped || out.EntersTapped
	card.NextUntapSkips = nil
	card.Counters = nil
	card.LostLastCounter = false
	card.KnownBy = nil
	card.ExilePlay = ExilePlayPermission{}
	card.DamageMarked = 0
	card.MarkedLethalByDeathtouch = false
	card.AttackingTarget = uuid.Nil
	card.BlockingTarget = uuid.Nil
	card.GoadedBy = uuid.Nil
	card.FaceDown = false
	card.BattleX = 0
	card.BattleY = 0
	card.EnteredBattlefieldAt = 0
	card.SummonedThisTurn = false
	card.NamedTribe = ""
	card.ChosenColor = ""
	card.effective = nil
	g.Battlefield.PushTop(card)
	g.markCardKnownInZoneLocked(g.Battlefield, newID)
	for name, n := range out.EntersWithCounters {
		_ = g.AddCounterForEffect(newID, name, n)
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   newController,
		CardID:  newID,
		OldZone: ZoneExile,
		NewZone: ZoneBattlefield,
	})
	g.EmitEvent(Event{Kind: EventETB, Actor: newController, CardID: newID})
	g.fireETBHookLocked(newID, CatalogKey(card))
	return newID, nil
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
// Three things it deliberately does NOT do:
//
//   - It does not count a land drop. A land an effect puts onto the
//     battlefield was not PLAYED (CR 305.4), so LandsPlayedThisTurn
//     is untouched and the enumerator still offers the turn's land
//     play. That is also why this event is not flagged
//     entryResumable: the generic resume in
//     executeEntryToBattlefieldLocked bumps the tally for any land
//     arriving with no stack item, which is right for the land-play
//     branch it was written for and wrong here.
//   - It does not run state checks. Like its neighbours it leaves
//     them to the resolution boundary the caller is already inside,
//     so a permanent that enters and dies does so once, together
//     with everything else that effect did.
//   - It does not offer the Clone choice. That prompt needs a
//     resumable entry site (copy_choice.go), and this is not one for
//     the reason above — the same declared simplification the search
//     path carries, and weaker than printed rather than stronger.
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
	slots, err := ParseProducedMana(produced)
	if err != nil {
		return err
	}
	// Read once, and only when a slot could use it: the identity read
	// walks every zone, and Dark Ritual's three "{B}" slots have
	// nothing to narrow or order.
	var identity commanderIdentity
	if opts.NarrowToCommanderIdentity || hasMultiOptionSlot(slots) {
		identity = commanderIdentityFor(g, p)
	}
	for _, slot := range slots {
		colorOptions := manaPickOptions(slot.Options, identity, opts.NarrowToCommanderIdentity)
		if len(colorOptions) == 0 {
			// CR 903.4f (#844): a "commander's color identity" effect
			// with no identity adds nothing, and an empty picker is not
			// a choice anybody can answer. An empty printed slot lands
			// here too, as it always did.
			continue
		}
		if len(slot.Options) == 1 {
			p.ManaPool.AddMana(ManaToken{Color: colorOptions[0], Source: source})
			g.EmitEvent(Event{Kind: EventManaAdded, Actor: playerID, Source: source})
			continue
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:         PendingChoiceMana,
			Chooser:      playerID,
			FromPlayer:   playerID,
			Count:        1,
			Source:       source,
			Reason:       addManaReason(slot),
			ColorOptions: colorOptions,
			ManaAmounts:  copyManaAmounts(slot.Amounts),
		})
	}
	return nil
}
