package protocol

// log.go projects the engine's internal event log (game.Event) into
// the PUBLIC game log that rides on GameView — the per-event,
// player-facing history of the table that ADR 0033 §4 calls a
// prerequisite for the bot seat.
//
// Three properties define it, in priority order:
//
//  1. **Public.** An entry may only say what a player sitting at the
//     table could have observed. Two independent mechanisms enforce
//     that. At BUILD time, an event whose card identity lives in a
//     hidden zone at both ends (library → hand, hand → library) is
//     emitted without any card reference at all, not even an instance
//     ID — the ID alone would let a client track a tutored card and
//     recognise it when it later surfaced. A reveal gets the same
//     treatment: it names the cards the table saw by printed name
//     and never by instance ID (see revealEntry). At FILTER time, every
//     surviving card reference goes through the SAME S13.5 knower
//     predicate that redacts CardViews, which is what keeps a
//     face-down Grizzly Bears anonymous to everyone but its
//     controller. Neither mechanism is redundant: the first covers
//     "the card is in a hidden zone", the second covers "the card is
//     visible but its identity is not".
//
//  2. **Bounded.** The log rides on every snapshot frame, so it is a
//     ring of PublicLogMax entries, oldest evicted. See the constant
//     for the measured wire cost.
//
//  3. **Derived, never stored.** There is no log buffer on game.Game.
//     The log is a pure projection of game.Game.Events, which
//     snapshot_drift_test.go already classifies `carried` — so the
//     history survives an undo, a restore, and a deploy for free, and
//     there is no second copy of the same facts to drift out of sync
//     with the first. The cost is one forward pass over the event log
//     per broadcast (not per viewer: ws.Room builds the view once and
//     filters it per client).
//
// Introduced in S31 sub-PR 0.

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PublicLogMax is how many entries GameView.Log carries. Oldest
// entries fall off the front.
//
// The number is a wire-cost decision, not a UX one: the log is on
// every snapshot frame, and sub-PR 2 has a view-size budget to
// protect. TestPublicLogWireCost measures it on a saturated
// four-player board and gates on the per-entry cost; see that test's
// output for the current number.
//
// 200 entries is roughly a full turn cycle and a half of a four-player
// game, which is the window a player — or a policy — actually reasons
// over: "who wiped the board last turn", "what has attacked me twice".
// Going to a thousand would have made the log the largest single thing
// on the wire.
//
// The schema below is shaped by the same constraint. Players are seat
// INDICES rather than UUIDs (the convention TurnView.ActiveSeat
// already uses) because a uuid string costs 36 bytes and a seat index
// costs one, and this struct repeats 200 times per frame; `step` rides
// only on LogStep entries, with every other entry inheriting it from
// the last one; and no field is present that the entry's kind already
// implies.
//
// The real fix is to ship only the entries a client has not seen,
// which needs a delta frame the v0 protocol does not have. Left for
// whoever grows one.
const PublicLogMax = 200

// NoSeat is the Seat value for an entry with no responsible player —
// a state-based life loss, a spell resolving with nobody to credit.
// Matches game.NoPriority's use of -1 as "no seat" on the wire.
const NoSeat = -1

// LogKind discriminates a public log entry. Deliberately coarser than
// game.EventKind: several engine events collapse into one line a
// human would read, and the engine kinds that carry no table-visible
// meaning (mana pool bookkeeping, becomes-target, effect errors)
// produce no entry at all.
type LogKind string

const (
	// LogStep — the turn cursor entered a step. The spine of the log:
	// every other entry is read relative to the last one of these.
	LogStep LogKind = "step"
	// LogCast — a spell was put on the stack.
	LogCast LogKind = "cast"
	// LogResolve — a spell or ability resolved.
	LogResolve LogKind = "resolve"
	// LogFizzle — a spell or ability was countered by game rules
	// (CR 608.2b): every target was illegal on resolution.
	LogFizzle LogKind = "fizzle"
	// LogCounter — a spell or ability was countered by another.
	LogCounter LogKind = "counter"
	// LogZone — a card changed zones. The catch-all for card motion
	// with at least one public endpoint.
	LogZone LogKind = "zone"
	// LogDraw — a player drew cards. Never names them; consecutive
	// draws by the same player in the same step collapse into one
	// entry with a count.
	LogDraw LogKind = "draw"
	// LogLife — a player's life total changed. Amount is signed.
	LogLife LogKind = "life"
	// LogDamage — damage was dealt to a player or a permanent.
	LogDamage LogKind = "damage"
	// LogAttack — a creature was declared as an attacker.
	LogAttack LogKind = "attack"
	// LogBlock — a creature was declared as a blocker.
	LogBlock LogKind = "block"
	// LogToken — a token was created.
	LogToken LogKind = "token"
	// LogSacrifice — a permanent was sacrificed. Distinct from the
	// LogZone entry it replaces, because "sacrificed" and "destroyed"
	// are different facts to a player reading the table.
	LogSacrifice LogKind = "sacrifice"
	// LogEliminated — a player left the game (concede, or any of the
	// state-based losses).
	LogEliminated LogKind = "eliminated"
	// LogReveal — a player revealed cards (CR 701.20). The engine
	// fires one EventRevealCards per card; every event sharing a
	// RevealSeq collapses into ONE entry, with Amount the card count
	// and OldZone the zone they were revealed from. The entry never
	// carries a card_id: see revealEntry.
	LogReveal LogKind = "reveal"
)

// logRevealNamesMax bounds how many revealed card names one LogReveal
// entry's text spells out. Five is Fact or Fiction, the largest
// fixed-size reveal in the catalog; past it the text says "and N
// more". The cap is for Hermit Druid, whose whole-library reveal
// would otherwise put a decklist into one line of a 200-line log that
// rides every frame.
const logRevealNamesMax = 5

// LogEvent is one line of the public game log.
//
// The wire shape carries BOTH a rendered `text` and the structured
// fields it was rendered from. The text is what the client panel
// prints and what a policy reads cheapest; the structured fields are
// what a client needs to make an entry interactive (click a card ID
// to highlight it on the board) and what a policy needs to reason
// over card identity without parsing English.
//
// `text` is rendered per viewer: it is regenerated in FilterViewFor
// for any entry whose card identity that viewer may not see, so the
// string never says more than the fields do.
type LogEvent struct {
	// Seq is the game.Event sequence number the entry was projected
	// from — monotonic, stable across frames, and the client's key.
	// Entries that collapse several events (a multi-card draw) carry
	// the seq of the FIRST event in the run.
	Seq uint64 `json:"seq"`
	// Kind is what happened.
	Kind LogKind `json:"kind"`
	// Turn is the turn number the entry happened on. Carried forward
	// from the last LogStep entry rather than stamped by the engine,
	// so entries that precede the first step announcement of a game
	// (the opening draw) carry turn 0.
	Turn int `json:"turn,omitempty"`
	// Step is the step name (game.Step) — present ONLY on LogStep
	// entries. Every other entry belongs to the step announced by the
	// most recent LogStep entry before it; a consumer that needs the
	// step on every line carries it forward, rather than the wire
	// repeating a 14-character string 200 times per frame.
	Step string `json:"step,omitempty"`
	// Seat is the seat index of the player responsible: the caster,
	// the controller of the attacking creature, the player whose life
	// changed. NoSeat when the event has no single actor.
	Seat int `json:"seat"`
	// TargetSeat is the seat index of the player the entry acts on —
	// who was attacked, who took the damage. Nil when the entry
	// targets a card (see Target) or nothing.
	TargetSeat *int `json:"target_seat,omitempty"`
	// CardID is the card the entry is about, when the viewer is
	// entitled to know a card is involved at all. Empty for entries
	// that deliberately carry no card reference (draws; any zone
	// change with two hidden endpoints).
	CardID string `json:"card_id,omitempty"`
	// Target is the CARD the entry acts on: a counterspell's victim,
	// a blocker's attacker. Player targets ride TargetSeat instead.
	Target string `json:"target,omitempty"`
	// Amount is the signed or counting payload: life delta, damage
	// dealt, cards drawn. On a LogEliminated entry it is 1 when the
	// player conceded and 0 when they lost — a whole field for one
	// bit, on an entry that happens at most three times a game, was
	// not worth the wire bytes.
	Amount int `json:"amount,omitempty"`
	// OldZone / NewZone are set on LogZone entries.
	OldZone string `json:"old_zone,omitempty"`
	NewZone string `json:"new_zone,omitempty"`
	// Combat marks a LogDamage entry as combat damage (CR 510).
	Combat bool `json:"combat,omitempty"`
	// CombatStep says which combat damage step dealt a combat LogDamage
	// entry: "first_strike" or "regular" (CR 510.4). Copied from
	// game.Event.CombatStep, and set only when that combat had a
	// first-strike step — combat with no first strike or double strike
	// anywhere is untagged, so the tag's presence alone says there are
	// two beats to show. #187, ADR 0053 Decision 1.
	CombatStep string `json:"combat_step,omitempty"`
	// Text is the rendered, human-readable line. Always present.
	Text string `json:"text"`

	// --- unexported render inputs, dropped by encoding/json ---

	// cardName / cardKnowers mirror CardView.Name and CardView.knowers
	// for CardID. FilterViewFor runs the same knower predicate it runs
	// on every CardView; when the viewer is not a knower the name is
	// dropped and Text re-rendered without it.
	cardName    string
	cardKnowers map[string]bool
	// targetName / targetKnowers are the same for Target.
	targetName    string
	targetKnowers map[string]bool
	// actorName / targetSeatName are player display names — public,
	// never redacted, and not on the wire because the seat index plus
	// GameView.Seats already carries them.
	actorName      string
	targetSeatName string
	// revealSeq / revealIDs / revealNames are LogReveal's render
	// inputs. The instance IDs exist only between projection and name
	// resolution — resolveLogNames swaps them for printed names and
	// drops them — so the exported half of the entry never holds a
	// handle to a card revealed out of a hidden zone.
	revealSeq   uint64
	revealIDs   []string
	revealNames []string
}

// hiddenZone reports whether a zone's contents are hidden from the
// table at large. Mirrors the wholesale-hide rule FilterViewFor
// applies to opponents' hands and libraries.
func hiddenZone(z game.ZoneKind) bool {
	return z == game.ZoneHand || z == game.ZoneLibrary
}

// publicLogOf projects g.Events into the bounded public log, using v
// (the already-assembled view) to resolve card names and knower sets.
//
// Caller must hold the read lock ViewOfGame already holds.
func publicLogOf(g *game.Game, v *GameView) []LogEvent {
	seatOf := seatIndexer(v)
	ring := newLogRing(PublicLogMax)
	var turn int
	var step string
	// sacrificed remembers the card of the EventSacrifice we just
	// emitted. game.EventSacrifice fires immediately before the
	// battlefield → graveyard move it causes, so the very next zone
	// move for that card is the same fact told twice.
	var sacrificed uuid.UUID
	// revealAt maps a RevealSeq to the ring push that opened its
	// entry. Keyed rather than adjacency-based for the reason
	// game.Event.RevealSeq gives: two back-to-back reveals are two
	// announcements, and a listener may emit between the per-card
	// events of one.
	revealAt := make(map[uint64]int)

	for _, ev := range g.Events {
		e, ok := projectEvent(ev, seatOf, &turn, &step, &sacrificed)
		if !ok {
			continue
		}
		e.Turn = turn
		if e.Kind == LogStep {
			e.Step = step
		}
		if e.Kind == LogReveal && e.revealSeq != 0 {
			if at, seen := revealAt[e.revealSeq]; seen {
				if prev := ring.pushed(at); prev != nil {
					prev.Amount += e.Amount
					// Only the IDs the text can name are kept; Amount
					// carries the true count.
					if len(prev.revealIDs) < logRevealNamesMax {
						prev.revealIDs = append(prev.revealIDs, e.revealIDs...)
					}
				}
				// An opening entry already evicted from the ring is
				// not re-opened: the tail of a reveal whose head fell
				// off the log is not worth a line of its own.
				continue
			}
			revealAt[e.revealSeq] = ring.total
		}
		// Collapse a run of single-card draws by one player into one
		// "drew N cards" line. Draw is the only event the engine
		// fires per card that a player reads as one action.
		if prev := ring.last(); prev != nil && e.Kind == LogDraw && prev.Kind == LogDraw &&
			prev.Seat == e.Seat && prev.Turn == e.Turn {
			prev.Amount += e.Amount
			continue
		}
		// The mulligan window re-runs the untap hook once per keep,
		// so the same step can announce several times. Say it once.
		if prev := ring.last(); prev != nil && e.Kind == LogStep && prev.Kind == LogStep &&
			prev.Turn == e.Turn && prev.Step == e.Step && prev.Seat == e.Seat {
			continue
		}
		ring.push(e)
	}

	out := ring.drain()
	if len(out) == 0 {
		return nil
	}
	resolveLogNames(out, v)
	for i := range out {
		out[i].Text = renderLogText(out[i], out[i].cardName, out[i].targetName)
	}
	return out
}

// projectEvent maps one engine event onto a log entry, or reports
// ok=false when the event produces no line. turn / step / sacrificed
// are the running projection state and may be updated.
//
// Every `return LogEvent{}, false` here is a deliberate scoping
// decision, not an oversight: the engine emits far more than a table
// log should show.
//
// game.Event.Batch (#829) is deliberately NOT forwarded: it is the
// engine's "these happened at the same time" identity for
// OncePerBatch, and the log already collapses adjacent lines by the
// rules the table cares about (LogReveal's RevealSeq, the
// draw / step runs in projectEvents). A second grouping key on the
// wire would be a second thing to keep consistent for no rendering
// gain. RevealSeq stays the one grouping the client reads.
func projectEvent(ev game.Event, seatOf func(uuid.UUID) int, turn *int, step *string, sacrificed *uuid.UUID) (LogEvent, bool) {
	base := LogEvent{Seq: ev.Seq, Seat: seatOf(ev.Actor)}
	// Anything but the zone move that follows a sacrifice clears the
	// pending marker — the suppression is only ever for the adjacent
	// pair.
	pendingSacrifice := *sacrificed
	if ev.Kind != game.EventSacrifice {
		*sacrificed = uuid.Nil
	}

	switch ev.Kind {
	case game.EventStepBegan:
		*turn = ev.Amount
		*step = ev.Label
		base.Kind = LogStep
		return base, true

	case game.EventCast:
		base.Kind = LogCast
		base.CardID = uuidStringOrEmpty(ev.CardID)
		// The destination is always the stack, so only a non-default
		// ORIGIN is worth wire bytes — and it is the one a card can
		// care about ("whenever you cast a spell from exile").
		if ev.OldZone != game.ZoneHand {
			base.OldZone = string(ev.OldZone)
		}
		return base, true

	case game.EventResolve:
		base.Kind = LogResolve
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventFizzle:
		base.Kind = LogFizzle
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventCounterSpell:
		base.Kind = LogCounter
		// Source is the counter; Target the countered item.
		base.CardID = uuidStringOrEmpty(ev.Source)
		base.Target = uuidStringOrEmpty(ev.Target)
		return base, true

	case game.EventZoneMove:
		if pendingSacrifice != uuid.Nil && pendingSacrifice == ev.CardID &&
			ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneGraveyard {
			// Already told as "X sacrificed Y".
			return LogEvent{}, false
		}
		// Both ends hidden: the table learns only that a hand or a
		// library changed size, which the zone counts already say.
		// Emitting the instance ID here would be the leak — it would
		// let a client follow a tutored card to the battlefield.
		if hiddenZone(ev.OldZone) && hiddenZone(ev.NewZone) {
			return LogEvent{}, false
		}
		// A cast already announced this motion.
		if ev.NewZone == game.ZoneStack {
			return LogEvent{}, false
		}
		// A spell going to the graveyard off the stack is the tail of
		// "resolved" / "was countered", not a separate fact.
		if ev.OldZone == game.ZoneStack && ev.NewZone == game.ZoneGraveyard {
			return LogEvent{}, false
		}
		base.Kind = LogZone
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.OldZone = string(ev.OldZone)
		base.NewZone = string(ev.NewZone)
		return base, true

	case game.EventSacrifice:
		*sacrificed = ev.CardID
		base.Kind = LogSacrifice
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventDrawCard:
		// Never the card — a drawn card is hidden, and the owner's
		// own hand already shows it. One entry per card, collapsed by
		// the caller.
		base.Kind = LogDraw
		base.Amount = 1
		return base, true

	case game.EventChangeLife:
		if ev.Amount == 0 {
			return LogEvent{}, false
		}
		base.Kind = LogLife
		// ChangeLife carries the player on Target, not Actor.
		base.Seat = seatOf(ev.Target)
		base.Amount = ev.Amount
		base.CardID = uuidStringOrEmpty(ev.Source)
		return base, true

	case game.EventDealDamage:
		if ev.Amount <= 0 {
			return LogEvent{}, false
		}
		base.Kind = LogDamage
		base.CardID = uuidStringOrEmpty(ev.Source)
		base.Amount = ev.Amount
		base.Combat = ev.Combat
		base.CombatStep = ev.CombatStep
		// Damage lands on a player or a permanent; the seat lookup is
		// what tells the two apart.
		base.setTarget(ev.Target, seatOf)
		return base, true

	case game.EventAttack:
		base.Kind = LogAttack
		base.CardID = uuidStringOrEmpty(ev.CardID)
		// DeclareAttacker validates its target against the seats, so
		// this is always a player today — setTarget keeps the entry
		// honest if planeswalker defenders ever widen it.
		base.setTarget(ev.Target, seatOf)
		return base, true

	case game.EventBlock:
		base.Kind = LogBlock
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Target = uuidStringOrEmpty(ev.Target)
		return base, true

	case game.EventTokenCreated:
		base.Kind = LogToken
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventConcede:
		base.Kind = LogEliminated
		base.Amount = 1 // conceded, rather than lost
		return base, true

	case game.EventPlayerEliminated:
		base.Kind = LogEliminated
		return base, true

	case game.EventRevealCards:
		return revealEntry(base, ev, seatOf), true

	default:
		return LogEvent{}, false
	}
}

// revealEntry projects one EventRevealCards. The caller collapses a
// run sharing a RevealSeq into the first entry.
//
// No card_id, for anyone, whatever zone the card was revealed from.
// A reveal is almost always out of a hand or a library, and the build-
// time rule at the top of this file applies unchanged: the instance
// ID of a card in a hidden zone is a handle that outlives the moment
// (Dark Confidant's card goes to hand, Fact or Fiction's go to hand
// and graveyard, a tutor's goes back to be shuffled). The per-viewer
// knower filter cannot stand in for that rule, because a reveal makes
// every seat a knower. The reveal frame reached the same answer; see
// reveal_frame.go.
//
// A reveal the whole table saw is logged with the printed names,
// resolved from revealIDs and never gated on knowers, for the reason
// publicRevealsOf gives: the table saw the cards, and a later shuffle
// clearing KnownBy must not blank the history.
//
// A reveal with a player Target is one only that player saw ("reveal
// it to target opponent", the #170 "Show to <player>" verb). It is
// logged for everyone, with TargetSeat set and no identity at all, so
// the line never says more than the viewer saw. The engine has no
// producer for it yet; the log fails closed ahead of it.
func revealEntry(base LogEvent, ev game.Event, seatOf func(uuid.UUID) int) LogEvent {
	base.Kind = LogReveal
	base.Amount = 1
	base.OldZone = string(ev.OldZone)
	base.revealSeq = ev.RevealSeq
	if ev.Target != uuid.Nil {
		if seat := seatOf(ev.Target); seat != NoSeat {
			base.TargetSeat = &seat
		}
		return base
	}
	if ev.CardID != uuid.Nil {
		base.revealIDs = []string{ev.CardID.String()}
	}
	return base
}

// resolveLogNames fills the unexported name / knower fields on every
// entry from the already-built view. The view is the single source of
// truth for both — resolving names from anywhere else would be a
// second visibility model to keep in sync with FilterViewFor's.
//
// A card the view cannot account for (a token that has ceased to
// exist) resolves to no name, and renders as "a card". Fail closed.
func resolveLogNames(entries []LogEvent, v *GameView) {
	if v == nil {
		return
	}
	wanted := make(map[string]bool, len(entries)*2)
	for _, e := range entries {
		if e.CardID != "" {
			wanted[e.CardID] = true
		}
		if e.Target != "" {
			wanted[e.Target] = true
		}
		for i, id := range e.revealIDs {
			if i >= logRevealNamesMax {
				break
			}
			wanted[id] = true
		}
	}

	seatNames := make([]string, len(v.Seats))
	for i, s := range v.Seats {
		seatNames[i] = s.Name
	}
	nameOfSeat := func(i int) string {
		if i < 0 || i >= len(seatNames) {
			return ""
		}
		return seatNames[i]
	}

	cards := make(map[string]CardView, len(wanted))
	if len(wanted) > 0 {
		indexCardsInto(cards, wanted, v.Battlefield, v.Stack, v.Exile)
		for _, s := range v.Seats {
			indexCardsInto(cards, wanted, s.Hand, s.Library, s.Graveyard, s.Command)
		}
	}

	for i := range entries {
		e := &entries[i]
		e.actorName = nameOfSeat(e.Seat)
		if e.TargetSeat != nil {
			e.targetSeatName = nameOfSeat(*e.TargetSeat)
		}
		if c, ok := cards[e.CardID]; ok {
			e.cardName = c.Name
			e.cardKnowers = c.knowers
		}
		if c, ok := cards[e.Target]; ok {
			e.targetName = c.Name
			e.targetKnowers = c.knowers
		}
		// Revealed cards are looked up in every zone, hands and
		// libraries included, without a knower check: see
		// revealEntry. A card that has since left every tracked zone
		// renders as "a card".
		if len(e.revealIDs) > 0 {
			e.revealNames = make([]string, 0, min(len(e.revealIDs), logRevealNamesMax))
			for j, id := range e.revealIDs {
				if j >= logRevealNamesMax {
					break
				}
				e.revealNames = append(e.revealNames, cards[id].Name)
			}
			e.revealIDs = nil
		}
	}
}

// seatIndexer returns a player-UUID → seat-index lookup over the
// view's seats, answering NoSeat for uuid.Nil and for anything not
// seated. Built once per projection; g.Seats is at most four entries,
// so a linear scan is cheaper than a map.
func seatIndexer(v *GameView) func(uuid.UUID) int {
	if v == nil {
		return func(uuid.UUID) int { return NoSeat }
	}
	ids := make([]string, len(v.Seats))
	for i, s := range v.Seats {
		ids[i] = s.ID
	}
	return func(id uuid.UUID) int {
		if id == uuid.Nil {
			return NoSeat
		}
		s := id.String()
		for i, seatID := range ids {
			if seatID == s {
				return i
			}
		}
		return NoSeat
	}
}

// setTarget routes an event's Target onto the right field: a seated
// player becomes TargetSeat, anything else is treated as a card.
func (e *LogEvent) setTarget(id uuid.UUID, seatOf func(uuid.UUID) int) {
	if id == uuid.Nil {
		return
	}
	if seat := seatOf(id); seat != NoSeat {
		e.TargetSeat = &seat
		return
	}
	e.Target = id.String()
}

// indexCardsInto adds every card of the given zones whose ID is in
// `wanted` to `dst`.
func indexCardsInto(dst map[string]CardView, wanted map[string]bool, zones ...ZoneView) {
	for _, z := range zones {
		for _, c := range z.Cards {
			if wanted[c.InstanceID] {
				dst[c.InstanceID] = c
			}
		}
	}
}

// redactLogForViewer projects the public log through the viewer's
// knower predicate — the SAME closure FilterViewFor runs over every
// CardView — and re-renders the text of any entry whose card identity
// the viewer may not see.
//
// knowers are cleared on the output, exactly as redactCardForViewer
// clears them, so a second pass over an already-filtered log redacts
// rather than re-revealing.
func redactLogForViewer(src []LogEvent, isKnower func(CardView) bool) []LogEvent {
	if len(src) == 0 {
		return nil
	}
	out := make([]LogEvent, len(src))
	for i, e := range src {
		cardName, targetName := e.cardName, e.targetName
		if e.CardID != "" && !isKnower(CardView{knowers: e.cardKnowers}) {
			cardName = ""
		}
		if e.Target != "" && !isKnower(CardView{knowers: e.targetKnowers}) {
			targetName = ""
		}
		e.cardKnowers, e.targetKnowers = nil, nil
		if cardName != e.cardName || targetName != e.targetName {
			e.cardName, e.targetName = cardName, targetName
			e.Text = renderLogText(e, cardName, targetName)
		}
		out[i] = e
	}
	return out
}

// --- rendering ---

// renderLogText writes the human-readable line. cardName / targetName
// are passed in rather than read off e so the same function serves
// both the unredacted build and the per-viewer re-render.
//
// An unknown card renders as "a card" — the entry still says that
// SOMETHING moved, which is public, without saying what.
func renderLogText(e LogEvent, cardName, targetName string) string {
	card := nameOr(cardName, "a card")
	actor := nameOr(e.actorName, "someone")
	// A target is either a seated player or a card; exactly one of the
	// two fields is set, and the player half is never redacted.
	target := nameOr(targetName, "a card")
	if e.TargetSeat != nil {
		target = nameOr(e.targetSeatName, "a player")
	}

	switch e.Kind {
	case LogStep:
		return fmt.Sprintf("Turn %d — %s · %s", e.Turn, actor, prettyStep(e.Step))
	case LogCast:
		if e.OldZone != "" {
			return fmt.Sprintf("%s cast %s from %s", actor, card, prettyZone(e.OldZone))
		}
		return fmt.Sprintf("%s cast %s", actor, card)
	case LogResolve:
		return fmt.Sprintf("%s resolved", card)
	case LogFizzle:
		return fmt.Sprintf("%s was countered by game rules (no legal targets)", card)
	case LogCounter:
		return fmt.Sprintf("%s countered %s", card, target)
	case LogZone:
		return renderZoneText(e, card)
	case LogSacrifice:
		return fmt.Sprintf("%s sacrificed %s", actor, card)
	case LogDraw:
		if e.Amount == 1 {
			return fmt.Sprintf("%s drew a card", actor)
		}
		return fmt.Sprintf("%s drew %d cards", actor, e.Amount)
	case LogLife:
		if e.Amount > 0 {
			return fmt.Sprintf("%s gained %d life", actor, e.Amount)
		}
		return fmt.Sprintf("%s lost %d life", actor, -e.Amount)
	case LogDamage:
		kind := "damage"
		if e.Combat {
			kind = "combat damage"
		}
		return fmt.Sprintf("%s dealt %d %s to %s%s", card, e.Amount, kind, target, combatStepSuffix(e.CombatStep))
	case LogAttack:
		return fmt.Sprintf("%s attacks %s", card, target)
	case LogBlock:
		return fmt.Sprintf("%s blocks %s", card, target)
	case LogToken:
		return fmt.Sprintf("%s created %s", actor, card)
	case LogEliminated:
		if e.Amount == 1 {
			return fmt.Sprintf("%s conceded", actor)
		}
		return fmt.Sprintf("%s was eliminated", actor)
	case LogReveal:
		return renderRevealText(e, actor)
	default:
		return card
	}
}

// combatStepSuffix is the tail a tagged combat LogDamage line gets, so
// the log panel, bug reports and the model-backed bot tiers all see
// which combat damage step dealt it without reading combat_step
// themselves ("Fencing Ace dealt 1 combat damage to Grizzly Bears
// (first strike)"). Untagged damage — every non-combat line, and
// combat with no first-strike step — gets nothing. An unknown value
// also gets nothing rather than being printed raw.
func combatStepSuffix(step string) string {
	switch step {
	case game.CombatStepFirstStrike:
		return " (first strike)"
	case game.CombatStepRegular:
		return " (regular damage)"
	}
	return ""
}

// renderZoneText words a zone change the way a player would say it.
// The destination carries the sentence; the origin is added only when
// it is public AND not already implied.
func renderZoneText(e LogEvent, card string) string {
	switch game.ZoneKind(e.NewZone) {
	case game.ZoneBattlefield:
		return fmt.Sprintf("%s entered the battlefield", card)
	case game.ZoneGraveyard:
		if game.ZoneKind(e.OldZone) == game.ZoneBattlefield {
			return fmt.Sprintf("%s was put into the graveyard from the battlefield", card)
		}
		return fmt.Sprintf("%s was put into the graveyard", card)
	case game.ZoneExile:
		return fmt.Sprintf("%s was exiled", card)
	case game.ZoneHand:
		return fmt.Sprintf("%s was returned to hand from %s", card, prettyZone(e.OldZone))
	case game.ZoneLibrary:
		return fmt.Sprintf("%s was put into a library from %s", card, prettyZone(e.OldZone))
	case game.ZoneCommand:
		return fmt.Sprintf("%s was put into the command zone", card)
	default:
		return fmt.Sprintf("%s moved from %s to %s", card, prettyZone(e.OldZone), prettyZone(e.NewZone))
	}
}

// renderRevealText words a LogReveal entry: "P1 revealed Island from
// their library", "P1 revealed 5 cards from their library: A, B, C,
// D, E", or, for a reveal to one player, "P1 revealed a card from
// their hand to P2" with no names for anyone.
func renderRevealText(e LogEvent, actor string) string {
	from := ""
	switch game.ZoneKind(e.OldZone) {
	case "":
	case game.ZoneHand, game.ZoneLibrary:
		from = " from their " + e.OldZone
	default:
		from = " from " + prettyZone(e.OldZone)
	}
	what := "a card"
	if e.Amount > 1 {
		what = fmt.Sprintf("%d cards", e.Amount)
	}
	if e.TargetSeat != nil {
		return fmt.Sprintf("%s revealed %s%s to %s", actor, what, from, nameOr(e.targetSeatName, "a player"))
	}
	if len(e.revealNames) == 0 {
		return fmt.Sprintf("%s revealed %s%s", actor, what, from)
	}
	names := make([]string, len(e.revealNames))
	for i, n := range e.revealNames {
		names[i] = nameOr(n, "a card")
	}
	if e.Amount <= 1 {
		return fmt.Sprintf("%s revealed %s%s", actor, names[0], from)
	}
	list := strings.Join(names, ", ")
	if more := e.Amount - len(names); more > 0 {
		list += fmt.Sprintf(" and %d more", more)
	}
	return fmt.Sprintf("%s revealed %s%s: %s", actor, what, from, list)
}

func nameOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// prettyStep turns "precombat_main" into "precombat main".
func prettyStep(s string) string {
	if s == "" {
		return "start of game"
	}
	return strings.ReplaceAll(s, "_", " ")
}

// prettyZone words a zone the way the log's sentences need it.
func prettyZone(z string) string {
	switch game.ZoneKind(z) {
	case game.ZoneBattlefield:
		return "the battlefield"
	case game.ZoneGraveyard:
		return "the graveyard"
	case game.ZoneExile:
		return "exile"
	case game.ZoneCommand:
		return "the command zone"
	case game.ZoneStack:
		return "the stack"
	case game.ZoneHand:
		return "hand"
	case game.ZoneLibrary:
		return "a library"
	default:
		return "elsewhere"
	}
}

// --- ring ---

// logRing is a fixed-capacity FIFO of log entries. A plain slice with
// re-slicing would either grow without bound or copy the whole buffer
// on every eviction, and the projection runs over every event in the
// game on every broadcast — so the eviction has to be O(1).
type logRing struct {
	buf   []LogEvent
	start int
	n     int
	// total counts every push ever made, evicted or not, so a caller
	// can name an entry by its push ordinal and ask for it back.
	total int
}

func newLogRing(capacity int) *logRing {
	return &logRing{buf: make([]LogEvent, capacity)}
}

// pushed returns a pointer to the entry made by push number `ord`
// (the value of total just before that push), or nil when it has been
// evicted or never happened.
func (r *logRing) pushed(ord int) *LogEvent {
	if ord < r.total-r.n || ord >= r.total {
		return nil
	}
	return &r.buf[(r.start+r.n-(r.total-ord))%len(r.buf)]
}

func (r *logRing) push(e LogEvent) {
	r.total++
	if r.n < len(r.buf) {
		r.buf[(r.start+r.n)%len(r.buf)] = e
		r.n++
		return
	}
	r.buf[r.start] = e
	r.start = (r.start + 1) % len(r.buf)
}

// last returns a pointer to the most recently pushed entry, for the
// in-place collapse of consecutive draws, or nil when empty.
func (r *logRing) last() *LogEvent {
	if r.n == 0 {
		return nil
	}
	return &r.buf[(r.start+r.n-1)%len(r.buf)]
}

// drain returns the entries oldest-first in a fresh slice.
func (r *logRing) drain() []LogEvent {
	if r.n == 0 {
		return nil
	}
	out := make([]LogEvent, r.n)
	for i := 0; i < r.n; i++ {
		out[i] = r.buf[(r.start+i)%len(r.buf)]
	}
	return out
}
