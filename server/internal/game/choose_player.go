package game

import (
	"sort"
	"strconv"

	"github.com/google/uuid"
)

// choose_player.go — "choose a player" / "choose an opponent", asked
// of a named seat while an effect resolves (#929).
//
// # No new prompt kind
//
// This is a PendingChoiceOptionPick (option_pick.go) whose options are
// SEATS: one option per eligible player, labelled with that player's
// name, addressed to the chooser. The question a player choice asks —
// here is a closed list of things, pick exactly one — is the question
// that kind already asks, so every consumer answers it unchanged: the
// choice gate has its row, `internal/legal` enumerates it, the wire
// projection carries it, and the client renders the same modal. A
// `choose_player` kind would have been a second spelling of an
// existing question and four files' worth of new cases for nothing
// (the reason MayChoice rides `confirm` and a pile split rides
// `option_pick`).
//
// # The choice is at RESOLUTION, and it is not a target
//
// "Choose a player" is not "target player". A target is named at
// announce (CR 601.2c), is public from that moment, is re-checked on
// resolution (CR 608.2b), and can be answered with hexproof or a
// redirection. A chosen player is picked while the effect resolves,
// by whoever the card says, and nothing may respond to it. Slithermuse
// draws off "choose an opponent", not "target opponent", and the
// difference is what stops a Slithermuse trigger from being fizzled by
// the opponent it was aimed at.
//
// The CR 614.12 "AS this enters, choose a player" form (True-Name
// Nemesis) is a THIRD thing again — a replacement-time choice made as
// the permanent enters, stored on the permanent, and read for the rest
// of its life. That form is designed, not built: see the note on
// ChoosePlayerPrompt below and ADR 0018's #929 amendment. It waits on
// protection (#662), which is the only thing that would read it.
//
// # The answer rides the item's payload
//
// StackItem.Payload already carries "what the effect that created this
// item had to tell it" as []TargetRef (#636, reflexive triggers). A
// chosen player is exactly that shape — a TargetPlayer ref — so it
// goes there rather than on a second payload field. Targets was not an
// option: a pick_target answer REPLACES Targets wholesale, so a chosen
// player parked there would be erased by the next re-target prompt.
//
// A clause that asks twice appends twice, in the order the card asks
// (Gluntch's first, second and third player), which is what makes
// "choose a SECOND player" expressible: the caller passes the players
// already chosen as the ones to leave out. ChosenPlayersOn reads them
// back in that order; ChosenPlayerOn reads the most recent.
//
// # Bots
//
// The eligible seats are offered in DESCENDING LIFE order, ties by
// seat. That ordering is the whole bot policy: `legal.choiceMoves`
// marks an option pick's FIRST option always-legal, and docs/bot.md's
// posture for a prompt from somebody else's card is "price what you
// can see and take the first offer otherwise" — so a bot with nothing
// better to say chooses the player with the most life. It is a
// legality-and-termination policy, not a strength one; see docs/bot.md.

// ChoosePlayerPrompt is the queue-side description of a
// resolution-time player choice.
//
// Among is the CANDIDATE list the card's clause admits — every player
// for "choose a player", the chooser's opponents for "choose an
// opponent", somebody else's opponents for "choose an opponent of
// target player". The caller computes it; this function drops the
// seats that have left the game (CR 800.4a) and orders what is left.
//
// Item is the stack item the answer is recorded on, and may be nil for
// a caller that reads the answer out of Then instead. Nothing in the
// engine re-checks a recorded player against the board — it is a
// record of a choice, not a target — so an effect that cares whether
// that seat is still seated asks.
type ChoosePlayerPrompt struct {
	// Chooser answers the prompt. Required, and frequently NOT the
	// player the options are about.
	Chooser uuid.UUID

	// Among are the players the clause admits, in any order.
	Among []uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is the prompt's header — the card's own sentence.
	Question string

	// Item is the stack item the chosen player is recorded on. Nil
	// records nothing.
	Item *StackItem

	// Then receives the chosen player. Runs with g.mu held; may queue
	// further choices, which is how a chain continues.
	Then func(g *Game, chosen uuid.UUID) error
}

// QueueChoosePlayerForEffect queues a "choose a player" and returns
// its ID, or uuid.Nil when it queued nothing — no eligible seat, or a
// chooser who has left the game (QueueChoiceForEffect's CR 800.4a
// guard).
//
// When it queues nothing it records the ABSENCE of a choice on the
// item, so a later ChosenPlayerOn on the same item reads "nobody was
// chosen" rather than the PREVIOUS clause's answer. A card that asks
// twice and gets an answer only the first time would otherwise apply
// the second clause to the first clause's player, which is the one
// way this shape can be silently wrong.
//
// A caller with more of the effect to run must check the return the
// way QueueOptionPickForEffect's callers do: nothing queued means
// nothing will call Then.
//
// Caller must hold g.mu.
func (g *Game) QueueChoosePlayerForEffect(p ChoosePlayerPrompt) uuid.UUID {
	eligible := g.eligibleChosenPlayersLocked(p.Among)
	if len(eligible) == 0 {
		recordChosenPlayer(p.Item, uuid.Nil)
		return uuid.Nil
	}
	options := g.seatChoiceOptionsLocked(eligible)
	item := p.Item
	then := p.Then
	queued := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  p.Chooser,
		Source:   p.Source,
		Question: p.Question,
		Options:  options,
		// ThenSeat, not Then: the answer is the option's own seat and
		// never an index into a slice captured here (#994). A seat that
		// leaves the game is pruned off this prompt while it is open
		// (CR 800.4a, pruneDepartedSeatOptionsLocked), which renumbers
		// every option after it — so a captured candidate list would
		// record the player one seat along from the one the chooser
		// clicked. Only a detached item pointer is closed over, which
		// is the StackItem.Effect contract: the branch resolves against
		// whichever *Game an undo restores.
		ThenSeat: func(g *Game, chosen uuid.UUID) error {
			if chosen == uuid.Nil {
				return ErrInvalidParam
			}
			recordChosenPlayer(item, chosen)
			if then == nil {
				return nil
			}
			return then(g, chosen)
		},
	})
	if queued == uuid.Nil {
		recordChosenPlayer(p.Item, uuid.Nil)
	}
	return queued
}

// EventPlayerChosen records a CR 614.12 "as this enters, choose a
// player" answer in the event log, the way EventColorChosen and
// EventCreatureTypeChosen record theirs. `Target` carries the chosen
// seat and `CardID` the permanent it was chosen for, so the log reads
// "True-Name Nemesis — Alice".
//
// Only the STORED form emits it. A resolution-time choice is part of
// one effect's resolution and is reported by whatever that effect then
// does; a stored one is a lasting fact about a permanent, and the
// window in which it was made is the only place the log can say so.
const EventPlayerChosen EventKind = "player_chosen"

// QueueChoosePlayerAsEntersForEffect queues the STORED form of the
// question: "As this permanent enters, choose a player" (CR 614.12),
// True-Name Nemesis. The answer lands on `source`'s Card.ChosenPlayer
// and is read for the rest of that permanent's life. Returns the choice
// ID, or uuid.Nil when nothing was queued.
//
// THE THIRD FORM, and the differences from the two above it are the
// reason it is a separate door rather than a flag. A target is named at
// announce and re-checked at resolution; a resolution-time chosen
// player lives on one StackItem.Payload and is gone when the item is;
// this one is made as the permanent enters, is stored on the permanent,
// and outlives every stack item involved. What it shares with the
// resolution-time form is the PROMPT — the same option_pick, built by
// the same seatChoiceOptionsLocked, so it is enumerated, gated,
// projected and (CR 800.4a) PRUNED identically. A seat that leaves
// while a Nemesis is still asking comes off its buttons exactly as it
// comes off Gluntch's (#994).
//
// It is queued from the permanent's AsEnters hook, not by pausing the
// CR 614 entry pipeline, and that is S26's declared simplification
// carried forward unchanged — creature_type_choice.go has the long form
// of the argument and color_choice.go took the same call. What it costs
// here: the Nemesis is on the battlefield with no chosen player for the
// window between entering and the answer arriving. Nothing can act in
// that window (an open PendingChoice stops priority), and the direction
// is the safe one — an unchosen player is nobody, so the protection
// applies to nothing rather than to everything.
//
// `among` is the candidate list the card's clause admits; the seats
// that have left are dropped and the rest ordered by the shared
// eligibility rule. Nothing queued — no eligible seat, or a chooser who
// has gone — leaves ChosenPlayer at uuid.Nil, which reads as "protected
// from nobody".
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueChoosePlayerAsEntersForEffect(chooser, source uuid.UUID, question string, among []uuid.UUID) uuid.UUID {
	eligible := g.eligibleChosenPlayersLocked(among)
	if len(eligible) == 0 {
		return uuid.Nil
	}
	permanent := source
	return g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		Options:  g.seatChoiceOptionsLocked(eligible),
		// ThenSeat for the same reason the resolution-time form uses
		// it: the answer is the option's own seat, so a prune cannot
		// renumber it (#994).
		ThenSeat: func(g *Game, chosen uuid.UUID) error {
			return g.setChosenPlayerLocked(permanent, chosen)
		},
	})
}

// setChosenPlayerLocked stamps the CR 614.12 answer onto the permanent.
//
// The permanent is located LIVE rather than trusted from the queue: the
// prompt is asynchronous and the Nemesis can have been killed in
// response to its own entry. A missing source is not an error — the
// choice was made and simply has nowhere to land, which is what CR
// 608.2 does with an effect whose object has gone.
//
// No layer-version bump, and that is the one place this differs from
// its two siblings. A named tribe and a chosen colour are AppliesTo
// inputs to their permanent's static abilities, so the layer engine's
// cached resolution has to be invalidated when they land. A chosen
// player is not: protection rides the PRINTED keyword token, which has
// been in Abilities since the permanent entered, and the one reader
// binds the seat at check time rather than baking it into a
// characteristic.
//
// Caller must hold g.mu.
func (g *Game) setChosenPlayerLocked(source, chosen uuid.UUID) error {
	var actor uuid.UUID
	if i := findCardOnBattlefield(g, source); i >= 0 {
		g.Battlefield.Cards[i].ChosenPlayer = chosen
		actor = g.Battlefield.Cards[i].Controller
	}
	g.EmitEvent(Event{
		Kind:   EventPlayerChosen,
		Actor:  actor,
		CardID: source,
		Target: chosen,
		Label:  g.seatLabelLocked(chosen),
	})
	return nil
}

// ChosenPlayerOf returns the player chosen for the permanent `sourceID`
// currently on the battlefield, or uuid.Nil when none has been chosen
// (or the permanent is gone).
//
// The accessor exists for the reason NamedTribeOf and ChosenColorOf do:
// a caller outside this package reads the answer without reaching into
// the battlefield slice itself. protection.go reads the field directly,
// because it is already holding the card.
//
// Caller must hold either lock.
func (g *Game) ChosenPlayerOf(sourceID uuid.UUID) uuid.UUID {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		return g.Battlefield.Cards[i].ChosenPlayer
	}
	return uuid.Nil
}

// eligibleChosenPlayersLocked narrows a candidate list to the seats
// that can actually be chosen and puts them in the order the prompt
// offers them.
//
// Dropped: a duplicate, a player the game does not have, and a player
// who has left (CR 800.4a — an eliminated seat is not a player, so
// "choose a player" cannot name one). Ordered: most life first, ties
// by seat, which is the bot policy documented at the top of this file.
//
// Caller must hold g.mu.
func (g *Game) eligibleChosenPlayersLocked(ids []uuid.UUID) []uuid.UUID {
	type seat struct {
		id   uuid.UUID
		life int
		n    int
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	seats := make([]seat, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil || seen[id] {
			continue
		}
		p := g.playerByIDLocked(id)
		if p == nil || p.Eliminated {
			continue
		}
		seen[id] = true
		seats = append(seats, seat{id: id, life: p.Life, n: p.Seat})
	}
	sort.SliceStable(seats, func(i, j int) bool {
		if seats[i].life != seats[j].life {
			return seats[i].life > seats[j].life
		}
		return seats[i].n < seats[j].n
	})
	out := make([]uuid.UUID, len(seats))
	for i, s := range seats {
		out[i] = s.id
	}
	return out
}

// seatChoiceOptionsLocked turns an ordered, already-eligible seat list
// into the option list a player prompt offers.
//
// THE ONE PLACE A SEAT BECOMES AN OPTION, for both forms of the
// question — the resolution-time prompt above and the CR 614.12
// as-enters one below. Label is what the chooser reads and Player is
// what the engine re-checks (#994), and they are set together here so
// a prompt can never carry a name the pruning paths cannot match back
// to a seat.
//
// Caller must hold g.mu.
func (g *Game) seatChoiceOptionsLocked(seats []uuid.UUID) []ChoiceOption {
	options := make([]ChoiceOption, len(seats))
	for i, id := range seats {
		options[i] = ChoiceOption{Label: g.seatLabelLocked(id), Player: id}
	}
	return options
}

// seatLabelLocked is what a seat is called on an option button. The
// player's name, or "Seat N" for a seat that has none — a nameless
// button is unanswerable, and the option list is the whole of what the
// chooser reads. Caller must hold g.mu.
func (g *Game) seatLabelLocked(id uuid.UUID) string {
	p := g.playerByIDLocked(id)
	if p == nil {
		return "Seat ?"
	}
	if p.Name != "" {
		return p.Name
	}
	return "Seat " + strconv.Itoa(p.Seat+1)
}

// recordChosenPlayer appends one player choice to an item's payload.
//
// uuid.Nil records the ABSENCE of a choice as a TargetNone ref rather
// than appending nothing: a card that asks twice has to be able to
// tell "nobody could be chosen this time" from "this clause was never
// asked", and ChosenPlayerOn reads the most recent answer either way.
func recordChosenPlayer(item *StackItem, chosen uuid.UUID) {
	if item == nil {
		return
	}
	if chosen == uuid.Nil {
		item.Payload = append(item.Payload, TargetRef{Kind: TargetNone})
		return
	}
	item.Payload = append(item.Payload, TargetRef{Kind: TargetPlayer, ID: chosen})
}

// ChosenPlayerOn is the player most recently chosen for this item, or
// uuid.Nil when the last question could not be asked and when none was
// asked at all.
//
// It scans BACKWARDS to the first player-or-absence ref, so a payload
// that also carries cards (a reflexive trigger's "the creatures
// tapped this way") does not hide the answer and does not fabricate
// one.
func ChosenPlayerOn(item *StackItem) uuid.UUID {
	if item == nil {
		return uuid.Nil
	}
	for i := len(item.Payload) - 1; i >= 0; i-- {
		switch item.Payload[i].Kind {
		case TargetPlayer:
			return item.Payload[i].ID
		case TargetNone:
			return uuid.Nil
		}
	}
	return uuid.Nil
}

// ChosenPlayersOn is every player chosen for this item, in the order
// the card asked. The read behind "choose a SECOND player": pass it
// back as the seats to leave out.
//
// Absences are skipped — a clause nobody could be chosen for excludes
// nobody from the next one.
func ChosenPlayersOn(item *StackItem) []uuid.UUID {
	if item == nil {
		return nil
	}
	var out []uuid.UUID
	for _, ref := range item.Payload {
		if ref.Kind == TargetPlayer {
			out = append(out, ref.ID)
		}
	}
	return out
}
