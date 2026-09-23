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
	"strconv"
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
//
// Which engine kinds those are is not a matter of taste any more:
// log_event_kind_gate_test.go reads every declared game.EventKind and
// fails unless projectEvent has an arm for it OR it is listed as a
// deliberate silence with a reason (#984).
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
	LogRoll   LogKind = "roll"
	LogFlip   LogKind = "flip"
	// LogChooseColor — a player answered a "choose a color" prompt
	// (CR 105.4): Coldsteel Heart as it enters, Wash Out as it
	// resolves. `Choice` is the colour LETTER and CardID the card the
	// colour was chosen for.
	LogChooseColor LogKind = "choose_color"
	// LogChooseType — a player answered a "choose a creature type"
	// prompt (CR 614.12): Cavern of Souls, Door of Destinies.
	// `Choice` is the type as the engine canonicalised it ("Elf").
	LogChooseType LogKind = "choose_type"
	// LogChooseName — a player answered an "as this enters, choose a
	// card name" prompt (CR 614.12): Pithing Needle, Phyrexian
	// Revoker, Sorcerous Spyglass. `Choice` is the name as the player
	// typed it, trimmed and otherwise untouched — CR 201.2 admits any
	// card name, so there is no canonical spelling to report (#1210).
	LogChooseName LogKind = "choose_name"
	// LogChoosePlayer — a player answered an "as this enters, choose
	// a player" prompt (CR 614.12): True-Name Nemesis. The answer is
	// a SEAT, so it rides TargetSeat like every other player
	// reference in this log and `Choice` is empty.
	LogChoosePlayer LogKind = "choose_player"
	// LogChooseCards — a player answered one of #1214's three
	// resolution-time picks (CR 608.2): an opponent choosing from a
	// set you revealed, somebody choosing among another player's
	// permanents, a seat choosing N of its own.
	//
	// `Amount` is how many cards were chosen and is always present;
	// `CardID` names the one card for the common single-card pick and
	// is empty for any other count, so a viewer who may not identify
	// it reads "chose a card" and one who may reads its name. `Target`
	// is the card that ASKED — the resolving spell or the permanent
	// whose ability it was — because "chose 2 cards" says nothing
	// without it.
	//
	// The line exists for the reason LogChooseColor does (#1023 /
	// #984): a decision the table watched somebody make is a decision
	// the history has to carry. Without it the only trace of Tragic
	// Arrogance is a column of sacrifices with nobody's name on them,
	// and the only trace of Gifts Ungiven is two cards appearing in a
	// graveyard.
	LogChooseCards LogKind = "choose_cards"
	// LogControl — a permanent changed controller (CR 613.1b).
	// `Seat` is the player who GAINED control and `TargetSeat` the
	// one who lost it, which is one sentence for a gain, an
	// exchange (CR 701.12) and a duration expiring alike.
	LogControl LogKind = "control"
	// LogSpecialAction — a player took a CR 116.2 special action:
	// foretell, suspend. `Label` is the action as the card prints
	// it ("Foretell {2}"), and the entry is the only thing that
	// says WHICH action a card leaving a hand for exile was.
	LogSpecialAction LogKind = "special_action"
	// LogCycle — a player cycled a card (CR 702.29b). It REPLACES
	// the LogZone line for the discard that paid the cost, which is
	// the same motion in words that do not say "cycled".
	LogCycle LogKind = "cycle"
	// LogCounters — the count of one counter kind on one card
	// changed (CR 122). `Label` is the kind ("+1/+1") and `Amount`
	// the count AFTER the change, which is what the engine's event
	// carries; a placement and a removal are the same line with a
	// different number. Not every kind gets one — see
	// counterKindIsNarrated.
	LogCounters LogKind = "counters"
	// LogScry / LogSurveil — a player finished a scry (CR 701.22)
	// or a surveil (CR 701.25). NEVER a card, for anyone: the cards
	// are hidden at both ends and the entry carries no card
	// reference at all. `Amount` is how many the table watched go to
	// the bottom of the library, or into the graveyard, and
	// `LookedAt` is the size of the action — "scry 2". Both halves
	// are public; neither identifies a card.
	LogScry    LogKind = "scry"
	LogSurveil LogKind = "surveil"
	// LogSagaChapter — a lore counter advanced a Saga onto a chapter
	// (CR 714.2b); `Amount` is the chapter number.
	LogSagaChapter LogKind = "saga_chapter"
	// LogClassLevel — a Class permanent's level designation changed
	// (CR 716.2); `Amount` is the new level. A level is not a
	// counter, so LogCounters cannot say it.
	LogClassLevel LogKind = "class_level"
	// LogSettings — a table setting changed (ADR 0075 §2.3). `Label`
	// is the setting's key ("undo_limit", "allow_spawn") and `Choice`
	// its NEW value as text; the old value is deliberately not on the
	// wire, because the sentence the table needs is "undos are 3 now",
	// not a diff.
	//
	// The one log kind that is not about the game: it is about the
	// rules the game is being played under, which is exactly why it
	// has to be written down. A budget that quietly halved mid-game is
	// the argument this line exists to prevent.
	LogSettings LogKind = "settings"
	// LogSpawn — the host or the admin put cards or tokens onto the
	// table from nowhere (ADR 0075 §2.4). `Label` is the card or
	// token name, `Amount` the count, `NewZone` where they went and
	// `TargetSeat` whose zone it was. This line is the ONLY thing
	// that tells the table a Treasure was handed out rather than
	// earned, which is the condition the feature ships under; a
	// spawn into a hidden zone names the zone and not the card.
	LogSpawn LogKind = "spawn"
	// LogTransform — a permanent was turned over to its other face
	// (CR 701.27a). `Label` is the name of the face it turned FROM,
	// which is the only place that name survives: the card's own name
	// is already the new face by the time the entry is projected.
	//
	// Narrated rather than silent, unlike the other board-state
	// changes: a card physically turning over is the clearest case
	// there is of a thing a player announces out loud, and a reader
	// scrolling back wants to know WHEN it happened, which the board
	// alone cannot say. Added in S46 (ADR 0079, #343).
	LogTransform LogKind = "transform"
	// LogPhaseOut / LogPhaseIn — a permanent phased out or in
	// (CR 702.26). #1199, ADR 0084.
	//
	// Narrated rather than silent, for LogTransform's reason and one
	// more that is stronger here. A phase-out is not a zone change
	// (CR 702.26d), so no LogZone entry says it, and the board simply
	// STOPS SHOWING the permanent — which is indistinguishable from a
	// permanent that died unless the log says which. Teferi's
	// Protection removes a whole board this way, and a reader
	// scrolling back has no other way to learn that it came back
	// rather than never left.
	LogPhaseOut LogKind = "phase_out"
	LogPhaseIn  LogKind = "phase_in"
	// LogStorm — a storm trigger settled on its count (CR 702.40a).
	// `card_id` is the storm spell and `amount` the count, which is
	// how many copies of it are about to be created and is zero for
	// the turn's first spell. #1238, ADR 0086.
	//
	// Narrated because nothing else narrates the number, and the
	// number is the card. A spell copy is created and not cast
	// (CR 707.10), so it emits no engine event and produces no line;
	// the trigger's own resolution is a LogResolve with no card_id,
	// which renders as "a card resolved". Without this entry the
	// table watches N Grapeshots appear from nowhere. Same argument
	// as LogSagaChapter and LogClassLevel: a mechanic whose number is
	// otherwise unobservable gets a line.
	LogStorm LogKind = "storm"
	// LogTurnFaceDown — a permanent that was face up on the
	// battlefield was turned face down (CR 708.2a). `card_id` is the
	// permanent; `target` is the object that did it (Ixidron, Cyber
	// Conversion), which is public where the permanent's identity no
	// longer is. #1209, ADR 0082's 2026-09-23 amendment.
	//
	// Narrated for LogTransform's reason and LogPhaseOut's stronger
	// one. Turning face down is not a zone change and not a
	// transform, so no other line says it; the board simply stops
	// showing a card the table could read a second ago and starts
	// showing a nameless 2/2, which is indistinguishable from the
	// creature having been exiled and replaced unless the log says
	// which.
	//
	// IT NAMES NOBODY, and that is not this projection's decision. A
	// CR 708.2 object has no name for ANY viewer, its controller
	// included (ADR 0069 decision 6: `CardView.Name` is the effective
	// characteristic's, and CR 708.2a leaves it empty; the controller
	// gets the art through `scryfall_id` and `face_visible` instead).
	// So resolveLogNames finds no name to put on the entry, and the
	// house fallback "a card" is the honest rendering — there is
	// nothing here to leak and nothing to withhold. `card_id` is
	// still on the entry, so a client can point at the permanent on
	// the board, which is the half of the identity that IS public.
	//
	// The reverse direction has no line at all: EventTurnedFaceUp's
	// silence row says the CR 116.2g special action's own line
	// already carries it, and there is no special action here to
	// carry this one.
	LogTurnFaceDown LogKind = "turn_face_down"
)

// The three choose-a-value kinds are separate rather than one "chose
// something" kind with a discriminator, because the discriminator
// would BE the kind: a client tones, filters and (one day) icons a log
// line by `kind`, and "Elf" and "Katara" are not the same row to a
// reader. They share one field (`Choice`) and one sentence shape
// ("<player> chose <value> for <card>"), which is the part worth
// having in one place. #984.

// The eight kinds after them are #1021: six of the silences #984 wrote
// down read as gaps rather than decisions, and each one is now a line.
// They are eight kinds for six decisions because scry and surveil are
// two keywords with two payoffs (game.EventSurveil says why) and a
// Saga chapter is not a Class level — a client tones and filters by
// `kind`, so collapsing either pair would take a distinction away from
// the reader to save a constant.

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
	// LookedAt is the SIZE of a LogScry / LogSurveil entry's keyword
	// action — the "2" in "scry 2" — where Amount is how many cards
	// the table then watched move. Zero on every other kind, and
	// zero on a scry whose event predates the field. Copied from
	// game.Event.LookedAt, which says why the two are separate
	// numbers. #1036.
	LookedAt int `json:"looked_at,omitempty"`
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
	// Random outcomes are public. One entry groups a whole instruction.
	Sides   int      `json:"sides,omitempty"`
	Results []int    `json:"results,omitempty"`
	Faces   []string `json:"faces,omitempty"`
	Call    string   `json:"call,omitempty"`
	Wins    int      `json:"wins,omitempty"`
	// Choice is the VALUE a player named at a "choose a ..." prompt:
	// the colour letter on a LogChooseColor entry ("G"), the
	// canonical creature type on a LogChooseType one ("Elf"). A
	// chosen PLAYER rides TargetSeat instead, for the reason every
	// other player in this struct does — a seat index costs one byte
	// and GameView.Seats already names it.
	//
	// Redacted with the card's name. CR 105.4 makes the answer public
	// at the table, but the answer also IDENTIFIES the card that asked
	// ("Elf" names Cavern of Souls as loudly as loyalty says
	// "planeswalker"), which is why #781 strips CardView.ChosenColor /
	// NamedTribe from a non-knower — so redactLogForViewer clears this
	// alongside the name rather than leaving the same fact on the wire
	// under a line that no longer says it.
	Choice string `json:"choice,omitempty"`
	// Label is the PRINTED NAME of the thing an entry is about, when
	// the kind needs one that is not a card: the special action as
	// the card prints it ("Foretell {2}") on a LogSpecialAction
	// entry, the counter kind ("+1/+1") on a LogCounters one.
	//
	// Redacted with the card's name, exactly as Choice is and for the
	// same reason (#781, #1021). "Foretell {2}" prices a card the
	// viewer may not identify — redactCardForViewer already strips
	// alternative_costs from a face-down card for precisely that leak
	// — and redactCardForViewer strips `counters` too, so a line
	// saying "+1/+1" under a card it no longer names would put back
	// what the CardView filter took away.
	Label string `json:"label,omitempty"`
	// ActorIsHost marks an entry whose actor held the table (ADR
	// 0075 §2.1) when they took it. Set only on the kinds where the
	// authority is the point — today LogSpawn, where "who put this
	// Treasure here" and "were they allowed to" are the same
	// question.
	//
	// Stamped by the ROOM, not by the projection: the host is a
	// property of the room, not of the engine, so publicLogOf cannot
	// know it. See StampHostOnLog, which the room calls on the same
	// view it stamps PlayerView.IsHost on. An entry from a table with
	// no host, or from a spawner who is not the host (the dev route
	// lets anyone at a preview table spawn), is left false — which is
	// the honest answer, not a missing one.
	ActorIsHost bool `json:"actor_is_host,omitempty"`
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
	batchSeq    uint64
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
	randomAt := make(map[uint64]int)

	for _, ev := range g.Events {
		e, ok := projectEvent(ev, seatOf, &turn, &step, &sacrificed)
		if !ok {
			continue
		}
		e.Turn = turn
		if e.Kind == LogStep {
			e.Step = step
		}
		if (e.Kind == LogRoll || e.Kind == LogFlip) && e.batchSeq != 0 {
			if at, seen := randomAt[e.batchSeq]; seen {
				if prev := ring.pushed(at); prev != nil {
					prev.Results = append(prev.Results, e.Results...)
					prev.Faces = append(prev.Faces, e.Faces...)
					prev.Wins += e.Wins
				}
				continue
			}
			randomAt[e.batchSeq] = ring.total
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
		// CR 702.29b: a cycling's cost DISCARDED the card, and the
		// LogZone line that motion already produced is the same fact
		// this entry names with the word a player uses. Replace it
		// in place — EventCycle is emitted immediately after the cost
		// is paid, so the zone line is the last one pushed — and keep
		// the zone line's seq, the convention a collapsed run follows
		// everywhere else in this function.
		if prev := ring.last(); prev != nil && e.Kind == LogCycle && prev.Kind == LogZone &&
			prev.CardID == e.CardID && game.ZoneKind(prev.NewZone) == game.ZoneGraveyard {
			e.Seq = prev.Seq
			*prev = e
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
	case game.EventRollDie, game.EventFlipCoin:
		base.CardID = uuidStringOrEmpty(ev.Source)
		base.batchSeq = ev.BatchSeq
		if ev.Kind == game.EventRollDie {
			base.Kind = LogRoll
			base.Sides = ev.Sides
			base.Results = []int{ev.Amount}
		} else {
			base.Kind = LogFlip
			base.Faces = []string{ev.Label}
			base.Call = ev.Call
			if ev.Won {
				base.Wins = 1
			}
		}
		return base, true

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

	case game.EventColorChosen:
		// CR 105.4: the choice is made at the table and heard by
		// everyone. #983 put it on the card; this puts it in the
		// history, which is the only place that says WHEN it was made
		// and by whom (#984).
		base.Kind = LogChooseColor
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Choice = ev.Label
		return base, true

	case game.EventCreatureTypeChosen:
		base.Kind = LogChooseType
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Choice = ev.Label
		return base, true

	case game.EventCardNameChosen:
		base.Kind = LogChooseName
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Choice = ev.Label
		return base, true

	case game.EventPlayerChosen:
		base.Kind = LogChoosePlayer
		base.CardID = uuidStringOrEmpty(ev.CardID)
		// Deliberately NOT setTarget: the answer is always a seat, and
		// setTarget's card fallback would put a raw player UUID on the
		// wire as `target` for a seat the view no longer carries. An
		// unseatable answer leaves the entry with no target at all,
		// which renders as "a player".
		if seat := seatOf(ev.Target); seat != NoSeat {
			base.TargetSeat = &seat
		}
		return base, true

	case game.EventCardsChosen:
		// #1214, CR 608.2. Actor chose, Source asked, CardID is the one
		// card when there was one. The asking card rides `Target` so
		// it goes through resolveLogNames' redaction like any other
		// card reference — it is usually a spell on the stack, which
		// is public, but a face-down or already-gone source should not
		// be named out of this line either.
		base.Kind = LogChooseCards
		base.Amount = ev.Amount
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Target = uuidStringOrEmpty(ev.Source)
		return base, true

	case game.EventControlChanged:
		// CR 613.1b, #930 / #1008. "Ian gained control of Grizzly
		// Bears" is a thing a player says out loud, and until #1021
		// the log said nothing: the card simply appeared under a
		// different controller on the next frame.
		//
		// Actor GAINED control and Target LOST it, which is the one
		// shape materialiseControlLocked emits for a gain, an
		// exchange (CR 701.12) and an expiry alike — Act of Treason's
		// creature going home at cleanup is this entry with the two
		// seats the other way round.
		base.Kind = LogControl
		base.CardID = uuidStringOrEmpty(ev.CardID)
		if seat := seatOf(ev.Target); seat != NoSeat {
			base.TargetSeat = &seat
		}
		return base, true

	case game.EventSettingsChanged:
		// ADR 0075 §2.3: every change is logged, because the settings
		// are the rules the table agreed to play under and a change to
		// them is a thing announced out loud. One event per field that
		// actually moved, so one line per field.
		//
		// Actor is uuid.Nil for the server admin — Seat is NoSeat and
		// the sentence says "The admin" rather than naming a seat.
		base.Kind = LogSettings
		base.Label = ev.Label
		base.Choice = ev.SettingNew
		return base, true

	case game.EventSpecialAction:
		// CR 116.2. The card's motion out of a hand is told by its
		// LogZone line — and that line cannot say WHICH action it was,
		// because a foretell, a suspend and a discard all read as "a
		// card left a hand". The Label is the word, as the card prints
		// it ("Foretell {2}"), which is also what the player said out
		// loud when they paid for it.
		base.Kind = LogSpecialAction
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Label = ev.Label
		return base, true

	case game.EventSpawned:
		// ADR 0075 §2.4. The whole justification for allowing a
		// production spawn is that the table can see it happen: a
		// spawned Treasure is indistinguishable from an earned one on
		// the board, and this line is the difference.
		base.Kind = LogSpawn
		base.Amount = ev.Amount
		base.NewZone = string(ev.NewZone)
		if seat := seatOf(ev.Target); seat != NoSeat {
			base.TargetSeat = &seat
		}
		// A spawn into a hand or a library names the ZONE and not the
		// card, for the same reason a hand → library move carries no
		// card reference at all: the name is hidden information about
		// a hidden zone, and the entry redaction downstream cannot
		// reach it — this entry has no CardID for the knower
		// predicate to key on. Dropped here, at build time, so it is
		// dropped for everyone including the spawner.
		if !hiddenZone(ev.NewZone) {
			base.Label = ev.Label
		}
		return base, true

	case game.EventCycle:
		// CR 702.29b. The cost's discard already produced a LogZone
		// line for the same motion; publicLogOf REPLACES it with this
		// one rather than printing both, the way the sacrifice pair is
		// one line. Without it a Drake Haven table cannot read why a
		// token appeared.
		base.Kind = LogCycle
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventCounterPlaced:
		// The loudest kind the engine emits — a Commander turn moves
		// dozens — so the log narrates the changes that are not
		// already a line somewhere else. counterKindIsNarrated is that
		// rule, in one place.
		if !counterKindIsNarrated(ev.Label) {
			return LogEvent{}, false
		}
		base.Kind = LogCounters
		// applyCounterLocked names the card on Target and carries NO
		// actor: a counter arrives from a resolved spell, a paid cost,
		// a trigger and the CR 704.5q cancel, and "who did it" is not
		// a fact those four share. The entry is card-shaped for the
		// same reason LogResolve is.
		base.CardID = uuidStringOrEmpty(ev.Target)
		base.Label = ev.Label
		// Amount is the count AFTER the change, which is what the
		// event carries; zero or less means the last one came off.
		base.Amount = ev.Amount
		return base, true

	case game.EventScry, game.EventSurveil:
		// CR 701.22 / CR 701.25, and the one arm in this switch that
		// deliberately names NO card, for anyone — not the cards that
		// were looked at (hidden at both ends, so the build-time rule
		// at the top of this file applies) and not the spell that
		// scried either, which the LogCast / LogResolve line above
		// already named. An entry that carries no card reference
		// cannot leak one.
		//
		// Two counts, and the line says both (#1036). Amount is the
		// MOTION the table watched — how many cards went to the
		// bottom of the library, or into the graveyard. LookedAt is
		// the SIZE of the action, the "scry 2" a player announces out
		// loud; #1021 shipped without it because it was on no event,
		// and it is on the event now.
		//
		// Neither of them names a card, which is the reason the size
		// is safe to say: "scry 2" tells the table how big the look
		// was, not what was in it.
		base.Kind = LogScry
		if ev.Kind == game.EventSurveil {
			base.Kind = LogSurveil
		}
		base.Amount = ev.Amount
		base.LookedAt = ev.LookedAt
		return base, true

	case game.EventSagaChapter:
		// CR 714.2b. A chapter firing is a beat of the turn, and until
		// #1021 only the chapter ability's own resolve line marked it
		// — a line that names the Saga and not the chapter.
		base.Kind = LogSagaChapter
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Amount = ev.Amount
		return base, true

	case game.EventClassLevel:
		// CR 716.2. A level is not a counter (CR 716.2b), so there is
		// no counter line for this one to be implied by.
		base.Kind = LogClassLevel
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Amount = ev.Amount
		return base, true

	case game.EventPhaseOut, game.EventPhaseIn:
		// CR 702.26. Not a zone move (CR 702.26d), so no LogZone entry
		// says it — this is the only line the table gets, and without
		// it a permanent silently vanishes from the board and is read
		// as dead.
		base.Kind = LogPhaseOut
		if ev.Kind == game.EventPhaseIn {
			base.Kind = LogPhaseIn
		}
		base.CardID = uuidStringOrEmpty(ev.CardID)
		return base, true

	case game.EventTurnedFaceDown:
		// CR 708.2a, #1209. Not a zone move and not a transform
		// (CR 701.27b is explicit that the two are different game
		// actions), so no other entry says it.
		//
		// The card reference is the permanent and is redacted for
		// every seat but its controller, which is the rule (CR 708.5)
		// arriving through the ordinary knower predicate. The SOURCE
		// goes in Target instead of in Label, so that it is redacted
		// on its own terms rather than travelling with the name it is
		// not: Ixidron is a public permanent and stays named in a
		// line that has stopped naming what it hit.
		base.Kind = LogTurnFaceDown
		base.CardID = uuidStringOrEmpty(ev.CardID)
		if ev.Source != ev.CardID {
			base.Target = uuidStringOrEmpty(ev.Source)
		}
		return base, true

	case game.EventTransform:
		// CR 701.27a. Not a zone move (CR 712.18), so no LogZone entry
		// says it — this is the only line the table gets, and without
		// it a permanent silently becomes a different card.
		base.Kind = LogTransform
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Label = ev.Label
		return base, true

	case game.EventStorm:
		// CR 702.40a, #1238. The count is the card, and nothing else
		// says it: a spell copy emits no event (CR 707.10 — it is
		// created, not cast) and the trigger's LogResolve carries no
		// card, so N copies would otherwise arrive unexplained.
		// Entered even at zero — "storm count 0" is the turn's first
		// spell, and a reader counting copies wants to see that the
		// trigger resolved and found nothing to copy rather than
		// wonder whether it fired.
		base.Kind = LogStorm
		base.CardID = uuidStringOrEmpty(ev.CardID)
		base.Amount = ev.Amount
		return base, true

	default:
		return LogEvent{}, false
	}
}

// counterKindIsNarrated is the RULE #1021 asked for before
// EventCounterPlaced could have an arm: WHICH counter changes are
// worth a line.
//
// Every kind is narrated except the two whose changes the log already
// tells somewhere else. That direction — an allowlist of silences
// rather than of lines — is deliberate: a homebrew or a newly printed
// counter kind is a thing a player would announce ("it has two stun
// counters now"), and the failure mode of the other direction is the
// one #984 was filed about, a fact nothing says.
//
//   - loyalty (CR 606.5): it moves on every activation and on every
//     point of damage a planeswalker takes, and BOTH of those are
//     already lines — the resolve of the ability, the LogDamage entry
//     — while the walker's CardView carries the total. A Commander
//     turn would otherwise spend a dozen log lines counting loyalty.
//   - lore (CR 714.2b): the LogSagaChapter line below says the same
//     advance, in the words the card prints. One fact, one line — the
//     rule the sacrifice pair follows.
//
// A lore counter that crosses NO chapter (a Saga past its final
// chapter gaining another) is therefore silent in both places, which
// is the right answer: nothing happened that a chapter ability or a
// reader cares about.
func counterKindIsNarrated(kind string) bool {
	switch kind {
	case game.CounterLoyalty, game.CounterLore:
		return false
	}
	return true
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

// amountNamesTheCard reports whether this entry's Amount is a
// CHARACTERISTIC of the card it names rather than a public quantity.
//
// Damage, life and a draw count are public whoever the card is; a
// counter total, a Saga's chapter and a Class's level are all read off
// the permanent, and redactCardForViewer strips every one of them from
// a viewer who may not identify it (`counters` directly, the rest with
// the type line and the abilities). So on a redacted entry they travel
// with the name — see redactLogForViewer.
func (e LogEvent) amountNamesTheCard() bool {
	switch e.Kind {
	case LogCounters, LogSagaChapter, LogClassLevel:
		return true
	}
	return false
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
			// The chosen value goes with the name it identifies. The
			// guard is on the DROP (the name was there and is not any
			// more), not on "no name": a card the view can no longer
			// account for was never redacted, and "P1 chose green for
			// a card" is the honest line for it.
			if cardName == "" {
				e.Choice = ""
				// #1021: and so does the printed label, for the same
				// reason — "Foretell {2}" prices the card the line no
				// longer names, which is the leak redactCardForViewer
				// clears alternative_costs to close.
				e.Label = ""
				if e.amountNamesTheCard() {
					// And the number with it: redactCardForViewer
					// strips CardView.counters from a non-knower, so a
					// count under a line that says "a card" would hand
					// back exactly what the card filter took away.
					e.Amount = 0
				}
			}
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
	case LogRoll, LogFlip:
		return renderRandomLogText(e, actor, card)
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
	case LogChooseColor:
		return fmt.Sprintf("%s chose %s for %s", actor, nameOr(game.ColorName(e.Choice), "a color"), card)
	case LogChooseType:
		return fmt.Sprintf("%s chose %s for %s", actor, nameOr(e.Choice, "a creature type"), card)
	case LogChooseName:
		return fmt.Sprintf("%s named %s for %s", actor, nameOr(e.Choice, "a card"), card)
	case LogChoosePlayer:
		// `target` is already the seat name here (or "a player"): a
		// chosen player never rides Target, so the card branch above
		// cannot have claimed it.
		if e.TargetSeat == nil {
			return fmt.Sprintf("%s chose a player for %s", actor, card)
		}
		return fmt.Sprintf("%s chose %s for %s", actor, target, card)
	case LogChooseCards:
		// #1214. `card` is the one card chosen (or "a card" for a
		// viewer who may not identify it), `target` the card that
		// asked. A count other than one names no card at all, so the
		// line says how many rather than guessing which.
		switch {
		case e.Amount == 0:
			return fmt.Sprintf("%s chose nothing for %s", actor, target)
		case e.Amount == 1 && e.CardID != "":
			return fmt.Sprintf("%s chose %s for %s", actor, card, target)
		case e.Amount == 1:
			return fmt.Sprintf("%s chose a card for %s", actor, target)
		default:
			return fmt.Sprintf("%s chose %d cards for %s", actor, e.Amount, target)
		}
	case LogControl:
		// `target` is the seat that lost control, for the same reason.
		if e.TargetSeat == nil {
			return fmt.Sprintf("%s gained control of %s", actor, card)
		}
		return fmt.Sprintf("%s gained control of %s from %s", actor, card, target)
	case LogSpecialAction:
		// An empty label is a redacted one (see redactLogForViewer) or
		// a special action a card declared without printing a name for;
		// the line still says a special action happened, which is what
		// the zone move on its own cannot.
		if e.Label == "" {
			return fmt.Sprintf("%s took a special action on %s", actor, card)
		}
		return fmt.Sprintf("%s used %s on %s", actor, e.Label, card)
	case LogCycle:
		return fmt.Sprintf("%s cycled %s", actor, card)
	case LogCounters:
		return renderCountersText(e, card)
	case LogScry, LogSurveil:
		return renderLookText(e, actor)
	case LogSagaChapter:
		if e.Amount <= 0 {
			return fmt.Sprintf("%s advanced a chapter", card)
		}
		return fmt.Sprintf("%s reached chapter %d", card, e.Amount)
	case LogClassLevel:
		if e.Amount <= 0 {
			return fmt.Sprintf("%s gained a level", card)
		}
		return fmt.Sprintf("%s became level %d", card, e.Amount)
	case LogSettings:
		return renderSettingsText(e)
	case LogSpawn:
		return renderSpawnText(e, target)
	case LogStorm:
		// "Grapeshot — storm count 3". The em dash is the separator
		// every storm-shaped stack label in the catalog already uses,
		// so the log line and the stack overlay read the same way. A
		// viewer who may not identify the card still gets the number:
		// the count is public (it is a fact about the turn's casts,
		// which every seat watched) and only the name is redacted.
		return fmt.Sprintf("%s — storm count %d", card, e.Amount)
	case LogTurnFaceDown:
		// `card` is "a card" for everyone, because the permanent has
		// stopped having a name (see LogTurnFaceDown's comment); it
		// is still resolved through the ordinary path so the line
		// improves on its own if that ever changes. `target` is the
		// object that did it — public, and the half of this line that
		// carries the information.
		if e.Target == "" {
			return fmt.Sprintf("%s was turned face down", card)
		}
		return fmt.Sprintf("%s turned %s face down", target, card)
	case LogPhaseOut:
		return fmt.Sprintf("%s phased out", card)
	case LogPhaseIn:
		return fmt.Sprintf("%s phased in", card)
	case LogTransform:
		// The card name is the face it turned INTO — viewOfCard reads
		// the active face — and Label is the one it turned from. Label
		// is cleared with the name on a redacted entry, so the fallback
		// is the line a viewer who may not identify the card reads.
		if e.Label == "" {
			return fmt.Sprintf("%s transformed", card)
		}
		return fmt.Sprintf("%s transformed into %s", e.Label, card)
	default:
		return card
	}
}

// renderSettingsText words a table-settings change the way the host
// would say it out loud (ADR 0075 §2.3): "Luke (host) set undos to 3
// per turn", "Luke (host) allowed spawning".
//
// The "(host)" is not read off the view — the log is projected inside
// ViewOfGame, before Room.stampHostLocked decides which seat wears the
// crown, so it is not available here. It is instead true BY
// CONSTRUCTION: the only two paths to EventSettingsChanged are the
// PATCH route and the set_table_settings action, and both refuse
// anyone who is not the host or the admin. An admin change has no seat
// at all (Actor is uuid.Nil), and says so.
//
// A key the table does not know about renders generically rather than
// being dropped: a setting nobody can name still changed, and a line
// that says so is better than a silence. The default arm is what a
// future field gets for free until someone gives it a sentence.
func renderSettingsText(e LogEvent) string {
	who := nameOr(e.actorName, "someone")
	if e.Seat == NoSeat {
		who = "The admin"
	} else {
		who += " (host)"
	}
	switch e.Label {
	case game.SettingUndoLimit:
		switch e.Choice {
		case strconv.Itoa(game.UndoUnlimited):
			return fmt.Sprintf("%s made undos unlimited", who)
		case "0":
			return fmt.Sprintf("%s turned undos off", who)
		case "1":
			return fmt.Sprintf("%s set undos to 1 per turn", who)
		}
		return fmt.Sprintf("%s set undos to %s per turn", who, e.Choice)
	case game.SettingUndoScope:
		if e.Choice == string(game.UndoScopeHostAny) {
			return fmt.Sprintf("%s may now undo anyone's action", who)
		}
		return fmt.Sprintf("%s limited undo to each player's own actions", who)
	case game.SettingStartingLife:
		return fmt.Sprintf("%s set starting life to %s", who, e.Choice)
	case game.SettingCommanderDamage:
		return fmt.Sprintf("%s set lethal commander damage to %s", who, e.Choice)
	case game.SettingBotPace:
		return fmt.Sprintf("%s set the bots' pace to %s", who, e.Choice)
	case game.SettingAllowSpawn:
		if e.Choice == "true" {
			return fmt.Sprintf("%s allowed spawning", who)
		}
		return fmt.Sprintf("%s disallowed spawning", who)
	}
	if e.Label == "" {
		return fmt.Sprintf("%s changed a table setting", who)
	}
	return fmt.Sprintf("%s set %s to %s", who, e.Label, nameOr(e.Choice, "its default"))
}

// renderSpawnText words a spawn (ADR 0075 §2.4): who did it, under
// what authority, what arrived, how many, and whose zone it landed
// in — "Luke (host) spawned 2 × Treasure onto Ana's battlefield."
//
// An empty Label is a spawn into a hidden zone, which names the zone
// and not the card by construction (see projectEvent). An unseated
// actor is the server admin: a spawn has exactly two authorised
// callers on a production table, so "someone" — renderLogText's
// fallback for an actorless entry — would be a worse answer than the
// one we have.
func renderSpawnText(e LogEvent, target string) string {
	who := e.actorName
	switch {
	case who == "":
		who = "The admin"
	case e.ActorIsHost:
		who += " (host)"
	}

	what := "a card"
	switch {
	case e.Label != "" && e.Amount > 1:
		what = fmt.Sprintf("%d × %s", e.Amount, e.Label)
	case e.Label != "":
		what = e.Label
	case e.Amount > 1:
		what = fmt.Sprintf("%d cards", e.Amount)
	}

	switch game.ZoneKind(e.NewZone) {
	case game.ZoneBattlefield:
		return fmt.Sprintf("%s spawned %s onto %s's battlefield", who, what, target)
	case game.ZoneExile:
		// Exile is shared (ZoneRef.Owner is nil for it), so there is
		// no "whose" to name.
		return fmt.Sprintf("%s spawned %s into exile", who, what)
	default:
		return fmt.Sprintf("%s spawned %s into %s's %s", who, what, target, spawnZoneWord(e.NewZone))
	}
}

// spawnZoneWord is prettyZone without the article, because a spawn
// line puts a possessive in front of it ("Ana's graveyard", not
// "Ana's the graveyard").
func spawnZoneWord(z string) string {
	switch game.ZoneKind(z) {
	case game.ZoneHand:
		return "hand"
	case game.ZoneGraveyard:
		return "graveyard"
	case game.ZoneLibrary:
		return "library"
	case game.ZoneCommand:
		return "command zone"
	default:
		return "zone"
	}
}

// StampHostOnLog marks the log entries whose actor is the seat this
// view flags as host, and re-renders their text.
//
// It is a second pass rather than part of publicLogOf because the
// host lives on the ROOM (ws/host.go) and the projection only has the
// engine. ws.Room.stampHostLocked calls it immediately after it
// stamps PlayerView.IsHost, on the same view, so the two can never
// disagree about who held the table.
//
// Only LogSpawn entries are touched: the marker is there to say a
// spawn was authorised, and putting "(host)" on every line a host
// happened to produce would make it noise rather than a signal.
func StampHostOnLog(v *GameView) {
	if v == nil {
		return
	}
	host := NoSeat
	for i, s := range v.Seats {
		if s.IsHost {
			host = i
			break
		}
	}
	if host == NoSeat {
		return
	}
	for i := range v.Log {
		e := &v.Log[i]
		if e.Kind != LogSpawn || e.Seat != host {
			continue
		}
		e.ActorIsHost = true
		e.Text = renderLogText(*e, e.cardName, e.targetName)
	}
}

// renderCountersText words a counter change the way a player reads the
// card: the kind, and the count that is on it NOW. A placement and a
// removal are the same sentence with a different number, because the
// engine's event carries the post-change total and nothing else — see
// game.EventCounterPlaced, and game.EventSagaChapter for what a rule
// that needed the delta had to do instead.
//
// An empty kind is a redacted one (redactLogForViewer clears it with
// the card's name), and the count goes with it, so the line says only
// that something changed.
func renderCountersText(e LogEvent, card string) string {
	if e.Label == "" {
		return fmt.Sprintf("%s's counters changed", card)
	}
	if e.Amount <= 0 {
		return fmt.Sprintf("%s has no %s counters left", card, e.Label)
	}
	if e.Amount == 1 {
		return fmt.Sprintf("%s now has 1 %s counter", card, e.Label)
	}
	return fmt.Sprintf("%s now has %d %s counters", card, e.Amount, e.Label)
}

// renderLookText words a finished scry or surveil. It says the keyword
// and two COUNTS and never a card — see the arm in projectEvent.
//
// The size (`LookedAt`) is the number the player announced: "scry 2".
// The motion (`Amount`) is what the table then watched happen — cards
// going to the bottom of the library (CR 701.22b), or into the
// graveyard (CR 701.25a). A scry that moved nothing still happened and
// still says so, because "they kept what was on top" is the half of
// the decision an opponent gets to read.
//
// A size of zero drops the number rather than printing "scried 0":
// that is what an EventScry recorded before #1036 decodes as, and what
// a hand-built event in a test carries, and "P1 scried and put 1 card
// on the bottom" is the honest line for an event that does not know
// how big the look was.
func renderLookText(e LogEvent, actor string) string {
	verb, where := "scried", "on the bottom"
	if e.Kind == LogSurveil {
		verb, where = "surveilled", "into their graveyard"
	}
	if e.LookedAt > 0 {
		verb = fmt.Sprintf("%s %d", verb, e.LookedAt)
	}
	if e.Amount <= 0 {
		return fmt.Sprintf("%s %s and kept every card on top", actor, verb)
	}
	if e.Amount == 1 {
		return fmt.Sprintf("%s %s and put 1 card %s", actor, verb, where)
	}
	return fmt.Sprintf("%s %s and put %d cards %s", actor, verb, e.Amount, where)
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
