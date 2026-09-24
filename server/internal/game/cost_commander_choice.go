package game

import (
	"github.com/google/uuid"
)

// cost_commander_choice.go — #1397: a commander moved to PAY A COST is
// offered CR 903.9, and the payment never pauses to ask.
//
// Paying a cost moves cards: a discard (Thrill of Possibility, Fauna
// Shaman, cycling), a return to hand (ninjutsu, Quirion Ranger), an
// exile from hand or graveyard (Cadaverous Bloom, Grim Lavamancer,
// scavenge), a sacrifice (Viscera Seer, Ashnod's Altar, Village Rites)
// and an alternative cost's card (Force of Will's pitch, escape's
// exiles). Any of those cards can be a commander, and CR 903.9 gives
// its OWNER — who is not always the payer: a stolen commander is
// sacrificed and returned by the player who controls it — the choice
// of the command zone instead.
//
// Two things pulled against each other before this file:
//
//   - CR 601.2h / 602.2b pay an announcement's costs as ONE
//     INDIVISIBLE STEP, and CR 605.3a gives a mana ability no priority
//     window at all. So the discard, return and exile payers set
//     zoneRoute.MustSettleNow, which SKIPPED the CR 903.9 question —
//     the commander went to the graveyard, the hand or exile and its
//     owner was never asked. That is #1397.
//   - The sacrifice payers and the alternative-cost payer did not set
//     it, so a sacrificed commander's move PAUSED mid-payment. The
//     ability went onto the stack with the commander still on the
//     battlefield, and it could be spent AGAIN — a second Ashnod's
//     Altar activation naming the same commander was accepted while
//     the first prompt was open. The pause was the honest answer to
//     the question and the wrong answer to the cost.
//
// The resolution (ADR 0013 §5af) is to ask BEFORE paying. Every
// announcement that moves cards as a cost — ActivateCatalogAbility,
// CastSpell, ActivateManaAbility — validates its payload in full, then
// asks askCostCommanderLocked whether any card the payment is about to
// move is a commander whose owner has not answered yet. If one is, the
// announcement is PARKED: nothing has been paid, nothing has moved,
// and a CR 903.9 prompt is queued to that card's owner. The owner's
// answer re-runs the whole announcement with the answer attached, so
// the payment that follows is the ordinary indivisible one — it
// settles every move inline, and each commander's move carries the
// answer its owner already gave (zoneRoute.commanderAnswer) instead of
// a question.
//
// Why this is rules-faithful rather than a workaround. The owner's
// choice is about WHERE the card goes, and nothing between the answer
// and the payment can change what the payment is: the announcement is
// re-validated from scratch when it resumes, and if the payer spent
// the card on something else in the meantime, the re-run refuses and
// nothing is paid. The one thing the owner learns early — that the
// payer means to pay this way — is on the table in paper too, where
// the payer names the card and the owner says "command zone".
//
// What it is NOT is a new prompt shape. The question is the existing
// optional_replacement yes/no, so the client, the legal-move
// enumerator and the bots answer it without knowing a cost is behind
// it; ResolveOptionalReplacement routes the answer here by the frame
// the prompt carries (costCommanderResume) instead of by its kind.

// commanderZoneAnswer is a CR 903.9 answer given BEFORE the move it
// is about, carried on a zoneRoute onto the ReplacementEvent the move
// opens. The zero value is "not asked", which leaves the built-in
// exactly as it always was.
type commanderZoneAnswer uint8

const (
	commanderZoneUnasked commanderZoneAnswer = iota
	// commanderZoneAccept: the owner chose the command zone. The
	// built-in applies as a MANDATORY replacement for this one event,
	// so a cost move that cannot pause (mustSettleNow) still applies
	// it rather than skipping it as a question.
	commanderZoneAccept
	// commanderZoneDecline: the owner chose the ordinary destination.
	// The built-in is not gathered for this event at all.
	commanderZoneDecline
)

// commanderAnswerFor reads one card's answer out of an announcement's
// answer map. A nil map is "nothing asked yet".
func commanderAnswerFor(answers map[uuid.UUID]bool, id uuid.UUID) commanderZoneAnswer {
	a, ok := answers[id]
	switch {
	case !ok:
		return commanderZoneUnasked
	case a:
		return commanderZoneAccept
	}
	return commanderZoneDecline
}

// costCommanderFrame is a parked announcement: the cost payment that
// is waiting on one commander's owner, and how to make it again once
// they have answered.
//
// Immutable once queued. The undo snapshot shares it with the live
// game (clone.go copies the PendingChoice by value), which is safe
// only because nothing writes it: the answer builds a NEW map rather
// than adding to `answers`, so a rewind into the open prompt replays
// the same announcement with the same earlier answers.
type costCommanderFrame struct {
	// payer is the player making the announcement. Not always the
	// chooser: the chooser is the commander's owner.
	payer uuid.UUID
	// card is the commander this prompt is about.
	card uuid.UUID
	// answers are the owners' answers already given for OTHER cards
	// of the same payment — a Village Rites paid with two partners'
	// worth of commanders asks twice, one card at a time.
	answers map[uuid.UUID]bool
	// announce re-runs the announcement under g.mu with `answers`
	// attached. It captures its arguments by value and reads the game
	// only through the *Game it is handed, so a replay against a
	// restored game acts on that game.
	announce func(g *Game, answers map[uuid.UUID]bool) error
}

// refusePausedCostCardsLocked is the other half of that gate (#1445),
// asked of the same `moving` list immediately before
// askCostCommanderLocked.
//
// Paying a cost no longer pauses — that is what the ask-first model
// above bought. But an EFFECT still can: a commander destroyed by Doom
// Blade, exiled out of a graveyard by Bojuka Bog or discarded by Mind
// Rot opens its owner's CR 903.9 prompt and waits where it was, on the
// battlefield, in the graveyard or in the hand, until the owner
// answers. That card is already on its way out. Naming it to a cost in
// the meantime — an Ashnod's Altar crack, a Village Rites, a Force of
// Will pitch, a scavenge — would spend an object the effect has
// already taken, and the destroy's own prompt would then be withdrawn
// as stale. Before this, one commander could be both destroyed and
// sacrificed for mana (CR 118.3: one object pays one cost).
//
// So a payment that would move a card whose exit is paused is refused
// outright with ErrChoicePending — the same answer the table gives any
// other action that needs a prompt answered first — and nothing is
// paid or asked. It is a refusal rather than a park: the question
// blocking this payment is somebody else's, already on the table, and
// once it is answered the card has gone and the payment is simply no
// longer available.
//
// #1427: the same holds for a payment that only TAPS a card — its own
// {T}, a crew, a convoke or waterbend tap, "tap an untapped creature
// you control". Nothing moves, so the destroy's prompt would even
// survive it, but the permanent being tapped has already left as far
// as the rules are concerned: a destroyed commander Birds of Paradise
// tapping for {G} while its owner decides is mana from an object that
// is gone. So the callers pass a second list, `tapping`, and it gets
// the same answer. It is kept apart from `moving` because only a card
// that moves is asked CR 903.9 by askCostCommanderLocked; tapping a
// live commander asks nothing.
//
// #1474: and the same holds for the two uses of a paused card that are
// in neither list, which is why the gate takes any number of them:
//
//   - CASTING it, or playing it as a land. CR 601.2a moves the spell
//     to the stack before a single cost is paid, so the card being
//     cast was never in `moving`, and a commander an effect was
//     exiling out of its owner's hand went onto the stack with the
//     exile's prompt still open. castSpellLocked asks the gate about
//     the card itself before anything else is judged, from whatever
//     zone it is being cast or played out of.
//   - ACTIVATING an ability of it, whatever the cost. A counter it
//     removes or adds, a loyalty cost, and an ability that costs
//     nothing at all neither move nor tap the source, and each is the
//     same use of an object that is already gone. So both activation
//     paths pass the SOURCE, together with every permanent a
//     counter-removal component names ("remove a +1/+1 counter from a
//     creature you control").
//
// Every list gets the same answer. They are separate arguments only
// because the callers build them for different reasons.
//
// Caller must hold g.mu.
func (g *Game) refusePausedCostCardsLocked(lists ...[]uuid.UUID) error {
	for _, ids := range lists {
		for _, id := range ids {
			if g.zoneChangePausedLocked(id) {
				return ErrChoicePending
			}
		}
	}
	return nil
}

// CardExitPausedForEffect is the gate above asked about one card, for
// the readers that have to agree with it without being an announcement:
// the protocol view clears `castable_here` on a card the gate would
// refuse to cast (#1474), so the zone browser never renders a button
// CastSpell answers with ErrChoicePending.
//
// Caller must hold g.mu (read or write).
func (g *Game) CardExitPausedForEffect(cardID uuid.UUID) bool {
	return g.refusePausedCostCardsLocked([]uuid.UUID{cardID}) != nil
}

// askCostCommanderLocked is the one gate every cost-paying
// announcement passes after it has validated its payload and before it
// pays anything.
//
// `moving` is every card the payment is about to move — the discards,
// the returns, the exiles, the sacrifices, the alternative cost's
// cards, and the source itself when the cost moves it. For each one
// that is a commander (Card.IsCommander, the same test the CR 903.9
// built-in makes) with no answer yet in `answers`:
//
//   - an owner who is gone (CR 800.4a) cannot be asked, and their
//     commander left the game with them anyway, so it is recorded as a
//     decline and the walk goes on;
//   - otherwise a CR 903.9 prompt is queued to the OWNER and the
//     function reports asked=true. The caller returns nil without
//     paying: the announcement is parked on the prompt, and
//     `announce` makes it again when the owner answers.
//
// When nothing is left to ask it returns asked=false and the answers
// the payment should carry, which may have gained declines for gone
// owners. The map is fresh when anything was added, so a caller's
// params never alias a frame's.
//
// `what` names the announcement for the prompt ("Viscera Seer",
// "Thrill of Possibility").
//
// Caller must hold g.mu.
func (g *Game) askCostCommanderLocked(payer uuid.UUID, moving []uuid.UUID, answers map[uuid.UUID]bool, what string, announce func(*Game, map[uuid.UUID]bool) error) (bool, map[uuid.UUID]bool) {
	for _, id := range moving {
		if _, answered := answers[id]; answered {
			continue
		}
		card, ok := g.LookupCardForEffect(id)
		if !ok || !card.IsCommander {
			continue
		}
		if g.chooserGoneLocked(card.Owner) || g.playerByIDLocked(card.Owner) == nil {
			answers = withCommanderAnswer(answers, id, false)
			continue
		}
		reason := "Send " + card.Name + " to the command zone instead?"
		if what != "" {
			reason = card.Name + " is paying " + what + "'s cost. Send it to the command zone instead?"
		}
		queued := g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceOptionalReplacement,
			Chooser: card.Owner,
			Count:   1,
			Source:  id,
			Reason:  reason,
			costCommanderResume: &costCommanderFrame{
				payer:    payer,
				card:     id,
				answers:  answers,
				announce: announce,
			},
		})
		if queued == uuid.Nil {
			// Refused at the queue (the owner left between the check
			// above and here — it cannot happen under one lock, but
			// the refusal is the queue's to make). Same answer as a
			// gone owner.
			answers = withCommanderAnswer(answers, id, false)
			continue
		}
		return true, answers
	}
	return false, answers
}

// withCommanderAnswer returns a NEW map holding `answers` plus one
// more. Never writes the map it was given: a parked frame and the
// params of the announcement it re-runs can hold the same map.
func withCommanderAnswer(answers map[uuid.UUID]bool, id uuid.UUID, apply bool) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(answers)+1)
	for k, v := range answers {
		out[k] = v
	}
	out[id] = apply
	return out
}

// resolveCostCommanderChoiceLocked answers a parked announcement's
// CR 903.9 prompt: record the answer and make the announcement again.
// The prompt has already been dequeued by the caller.
//
// The re-run validates everything from scratch against the game as it
// is NOW, because the payer kept priority while the owner decided and
// may have spent the card, the mana or the source on something else.
// A re-run that no longer validates pays nothing and moves nothing —
// the announcement simply is not made — and the refusal goes back to
// the player who answered only when that is the payer, who can act on
// it; an opponent who answered a question about their commander is
// not handed "insufficient mana" for somebody else's ability, and the
// log carries it instead.
//
// A payer who has left the game while the question was open has no
// announcement left to make (CR 800.4a), so the answer is recorded
// and nothing else happens.
//
// Caller must hold g.mu.
func (g *Game) resolveCostCommanderChoiceLocked(frame *costCommanderFrame, chooserID uuid.UUID, apply bool) error {
	if frame == nil || frame.announce == nil {
		return ErrInvalidParam
	}
	if g.chooserGoneLocked(frame.payer) || g.playerByIDLocked(frame.payer) == nil {
		return nil
	}
	err := frame.announce(g, withCommanderAnswer(frame.answers, frame.card, apply))
	if err == nil {
		return nil
	}
	if chooserID == frame.payer {
		return err
	}
	g.EmitEvent(Event{
		Kind:     EventEffectError,
		Actor:    frame.payer,
		Source:   frame.card,
		ErrorMsg: "the payment the command-zone answer was for could no longer be made: " + err.Error(),
	})
	return nil
}
