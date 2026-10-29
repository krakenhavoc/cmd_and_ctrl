package game

import (
	"reflect"
	"sort"

	"github.com/google/uuid"
)

// replacement_resume_keyed.go — #4253: a CR 616 prompt's resume frame
// written as data, so a table waiting on one is a restore point.
//
// Every CR 616 prompt (a "may" replacement's yes/no, riot's question, the
// ordering prompt of CR 616.1) pauses one ReplacementEvent and resumes it
// through PendingChoice.replacementResume: the event, the replacements
// gathered for it, and — on the Game — the CR 614.5 once-per-event marks.
// Until now none of that was written to a restore point, so the census
// counted the frame (ChoiceResumeFrames) and the table was not one.
//
// # What the frame is, on disk
//
// The frame is four things, and each is data:
//
//   - The event: its kind and its fields (replacementEventSnapshot). Its
//     own continuation is either data already (a draw's drawTail and
//     drawThen, a replaced draw's or learn's registered body, an expanded
//     draw's pending siblings) or is engine code chosen by the event's
//     kind (finishSettledReplacementLocked). A field that held a Go
//     closure — a route's `then`, an entry batch, a token, life, counter,
//     damage or mill tail — kept the frame off the disk, counted by the
//     census. #4361 deleted the last of them (S18 reclassified the
//     frame's closure-ledger lines keyed and rebuilt): what still keeps
//     a frame off the disk is an event kind this file does not key, a
//     gate field, or a card or stack item the mirrors cannot write whole.
//   - The pending replacement set: the IDs of the gathered effects, in
//     gathered order. The effects themselves are rebuilt by the running
//     binary, by gathering the restored event again and picking those
//     IDs, because every source a gather reads (the board, the
//     graveyards, the entering card, its look-ahead, the scoped records)
//     is in the restore point. The object serials the catalog IDs pack
//     (#3482) are written beside the prompts, so the IDs mean the same
//     objects after the restore.
//   - The marks: which effects have already applied to this event.
//   - A registered key for the tail that runs once the window settles,
//     one per event kind (the three keys below). It rides the prompt's
//     `pickThen`, so a binary from before this change, which checks a
//     pickThen it does not know as a revealed-hand pick's continuation,
//     refuses the file (ErrUnknownEffectKey) instead of restoring a prompt
//     with no frame behind it, which would wedge the table.
//
// # Event kinds
//
// Keyed: a draw (dredge, Underrealm Lich, Forbidden Crypt, an ordering
// prompt between a doubler and a dredge); a keyword action whose tail is
// data (learn, so Retriever Phoenix; a scry, surveil, proliferate or
// earthbend with no "then"); a zone move — an exit whose route carries
// no continuation (a commander's CR 903.9b bounce) and a battlefield
// entry with no batch and no entry continuation (riot, unleash, a
// permanent spell's entry, whose resolving stack item is written with
// it); and a counter placement or a life change with no tail (an
// ordering prompt between Hardened Scales and Doubling Season, or
// between two life-gain replacements).
//
// Not keyed when this was written (the frame stayed a closure, counted
// by the census): a discard, a token creation, damage, a mill, a step
// transition, a countered spell, and a keyword action, move, entry,
// counter placement or life change whose tail held a closure, and the
// entry of a staged token. #4253 and #4361 keyed most of those, and no
// tail holds a closure any more; replacementEventFields' gates (a step
// transition, a countered spell, mana, an extra turn, dice, combat
// damage) and replacementEventKeyable say what is still counted.
//
// # Prompt kinds
//
// Keyed: every prompt kind whose resolve path reads nothing but the
// frame and the prompt's own written fields (keyedReplacementPromptKind).
// The CR 616 three (optional_replacement, replacement_order, entry_riot)
// and the entry questions that pause through the same frame: copy_target,
// entry_controller, the four entry card choices (reveal, discard,
// sacrifice, choose permanents), aura_host, entry_read_ahead and
// entry_pay_life (#4253 slice 2). An entry card choice's live zone is the
// one its prompt kind names (liveZone), so it needs nothing more (#4361
// S7b: the zone-only chooseCardsFrame it held is gone).

// The registered tails, one per keyed event kind. On-disk identities:
// never renamed, never reused. Ledgered as "resume <key>".
const (
	ReplacementResumeDraw          = "replacement-resume/draw"
	ReplacementResumeKeywordAction = "replacement-resume/keyword-action"
	ReplacementResumeMove          = "replacement-resume/move"
	ReplacementResumeCounter       = "replacement-resume/counter"
	ReplacementResumeLife          = "replacement-resume/life"
	// #4253 slice 3: a token creation instruction (CR 701.7a).
	ReplacementResumeCreateTokens = "replacement-resume/create-tokens"

	// #4253 slice 3: the same event kinds when the event carries a tail
	// that is data (eventCarriesDataTail). A separate key so that a
	// binary from before the tail was carried, which would silently drop
	// the JSON field it does not know and restore the prompt without the
	// rest of its effect, refuses the file instead (ErrUnknownEffectKey).
	ReplacementResumeMoveTail          = "replacement-resume/move-tail"
	ReplacementResumeCounterTail       = "replacement-resume/counter-tail"
	ReplacementResumeLifeTail          = "replacement-resume/life-tail"
	ReplacementResumeKeywordActionTail = "replacement-resume/keyword-action-tail"

	// #4253 slice 6: a move whose single-permanent entry carries the
	// caller's "then" as a registered tail (entryTail.keyed). Its own key
	// again: a binary from before it knows the move-tail key but drops the
	// field, and would restore the prompt without the rest of its card.
	ReplacementResumeMoveEntryTail = "replacement-resume/move-entry-tail"

	// #4253 slice 9: a move that is one leg of a keyed route batch
	// (zoneRoute.batchLegs, with the batch's last-known copies), a
	// discard (RepEventDiscard, which no key covered before; its legs
	// ride zoneRoute.discardLegs), and an entry that a keyed discard leg
	// became (Obstinate Baloth, entryBatch.discard). Keys of their own:
	// a binary from before them knows the move-tail key and would drop
	// the fields it does not know.
	ReplacementResumeMoveBatch    = "replacement-resume/move-batch"
	ReplacementResumeDiscard      = "replacement-resume/discard"
	ReplacementResumeDiscardEntry = "replacement-resume/discard-entry"

	// #4253 slice 10a: a mill instruction's amount window (RepEventMill,
	// which no key covered before; its tail is millTail's data half), and
	// a sacrifice that is one half of an exchange (zoneRoute's settled
	// entry, Goblin Welder). Keys of their own: a binary from before them
	// does not key a mill at all, and knows the move key but would drop
	// the exchange's entry.
	ReplacementResumeMill         = "replacement-resume/mill"
	ReplacementResumeMoveExchange = "replacement-resume/move-exchange"

	// #4517: a move carrying a reflexive trigger a replacement declared
	// on it (reflexiveAfterMove, "exile it instead. When you do, …").
	// It takes precedence over every other move key: a binary from
	// before it knows those keys and would drop the trigger.
	ReplacementResumeMoveReflexive = "replacement-resume/move-reflexive"
)

// replacementResumeTails maps each tail key to the event kind it
// finishes. The tail itself is finishSettledReplacementLocked's arm for
// that kind, reached through the prompt kind's own resolve path.
var replacementResumeTails = map[string]ReplacementEventKind{
	ReplacementResumeDraw:          RepEventDraw,
	ReplacementResumeKeywordAction: RepEventKeywordAction,
	ReplacementResumeMove:          RepEventMove,
	ReplacementResumeCounter:       RepEventCounter,
	ReplacementResumeLife:          RepEventLife,
	ReplacementResumeCreateTokens:  RepEventCreateTokens,

	ReplacementResumeMoveTail:          RepEventMove,
	ReplacementResumeMoveEntryTail:     RepEventMove,
	ReplacementResumeCounterTail:       RepEventCounter,
	ReplacementResumeLifeTail:          RepEventLife,
	ReplacementResumeKeywordActionTail: RepEventKeywordAction,

	ReplacementResumeMoveBatch:    RepEventMove,
	ReplacementResumeDiscard:      RepEventDiscard,
	ReplacementResumeDiscardEntry: RepEventMove,

	ReplacementResumeMill:         RepEventMill,
	ReplacementResumeMoveExchange: RepEventMove,

	ReplacementResumeMoveReflexive: RepEventMove,

	ReplacementResumeDamage: RepEventDamage,

	// #4362: combat damage (combat_damage_keyed.go).
	ReplacementResumeCombatDamage: RepEventDamage,

	// #4404: a damage event with redirected parts
	// (damage_redirected_keyed.go).
	ReplacementResumeDamageRedirected:       RepEventDamage,
	ReplacementResumeCombatDamageRedirected: RepEventDamage,

	// #4404: a simultaneous entry of created tokens, and an Attraction's
	// open (entry_batch_keyed.go).
	ReplacementResumeMoveTokenBatch:     RepEventMove,
	ReplacementResumeMoveAttractionOpen: RepEventMove,
}

// replacementResumeBaseKeys is the key a frame of each kind is written
// under, and replacementResumeTailedKeys the key when its event carries
// a data tail.
var (
	replacementResumeBaseKeys = map[ReplacementEventKind]string{
		RepEventDraw:          ReplacementResumeDraw,
		RepEventKeywordAction: ReplacementResumeKeywordAction,
		RepEventMove:          ReplacementResumeMove,
		RepEventCounter:       ReplacementResumeCounter,
		RepEventLife:          ReplacementResumeLife,
		RepEventCreateTokens:  ReplacementResumeCreateTokens,
		RepEventDiscard:       ReplacementResumeDiscard,
		RepEventMill:          ReplacementResumeMill,
		RepEventDamage:        ReplacementResumeDamage,
	}
	replacementResumeTailedKeys = map[ReplacementEventKind]string{
		RepEventMove:          ReplacementResumeMoveTail,
		RepEventCounter:       ReplacementResumeCounterTail,
		RepEventLife:          ReplacementResumeLifeTail,
		RepEventKeywordAction: ReplacementResumeKeywordActionTail,
		// A token creation's key is new with its tail, and so is a
		// discard's.
		RepEventCreateTokens: ReplacementResumeCreateTokens,
		RepEventDiscard:      ReplacementResumeDiscard,
		RepEventMill:         ReplacementResumeMill,
		// #4253: a damage event's key is new with its frame.
		RepEventDamage: ReplacementResumeDamage,
	}
)

// eventCarriesDataTail reports whether a keyable event carries a tail
// written as data (#4253 slice 3), so its frame takes the tailed key.
func eventCarriesDataTail(ev *ReplacementEvent) bool {
	if t := ev.entryTail; t != nil && (t.search != nil || t.batch != nil || t.tokens != nil || !t.keyed.IsZero()) {
		return true
	}
	if r := ev.zoneRoute; r != nil && (r.faceDownLegs != nil || r.batchLegs != nil || r.discardLegs != nil || r.simultaneousEntry != nil) {
		return true
	}
	if t := ev.keywordAction; t != nil && !t.keyed.IsZero() {
		return true
	}
	if t := ev.damageTail; t != nil && (!t.keyed.IsZero() || t.each != nil) {
		return true
	}
	return ev.counterTail != nil || ev.lifeTail != nil || ev.tokenTail != nil || ev.mill != nil
}

// replacementResumeTailedKeyOf is the key a keyable event is written under
// when it carries a data tail: its kind's tailed key, or the move-entry-tail
// key when the tail is a single permanent's entry tail (#4253 slice 6).
func replacementResumeTailedKeyOf(ev *ReplacementEvent) string {
	if ev.Kind == RepEventMove && len(ev.reflexiveAfterMove) > 0 {
		return ReplacementResumeMoveReflexive
	}
	if ev.Kind == RepEventMove && ev.entryTail != nil && ev.entryTail.batch != nil && ev.entryTail.batch.discard != nil {
		return ReplacementResumeDiscardEntry
	}
	if ev.Kind == RepEventMove && ev.entryTail != nil {
		if key := entryBatchOwnKey(ev.entryTail.batch); key != "" {
			return key
		}
	}
	if ev.Kind == RepEventMove && ev.zoneRoute != nil && ev.zoneRoute.batchLegs != nil {
		return ReplacementResumeMoveBatch
	}
	if ev.Kind == RepEventMove && ev.zoneRoute != nil && ev.zoneRoute.simultaneousEntry != nil {
		return ReplacementResumeMoveExchange
	}
	if ev.Kind == RepEventMove && ev.entryTail != nil && !ev.entryTail.keyed.IsZero() {
		return ReplacementResumeMoveEntryTail
	}
	return replacementResumeTailedKeys[ev.Kind]
}

// snapshotTailedKeyOf is replacementResumeTailedKeyOf over the mirror.
func snapshotTailedKeyOf(ev *replacementEventSnapshot, kind ReplacementEventKind) string {
	if kind == RepEventMove && len(ev.ReflexiveAfterMove) > 0 {
		return ReplacementResumeMoveReflexive
	}
	if kind == RepEventMove && ev.EntryTail != nil && ev.EntryTail.Batch != nil && ev.EntryTail.Batch.Discard != nil {
		return ReplacementResumeDiscardEntry
	}
	if kind == RepEventMove && ev.EntryTail != nil {
		if key := snapshotEntryBatchOwnKey(ev.EntryTail.Batch); key != "" {
			return key
		}
	}
	if kind == RepEventMove && ev.ZoneRoute != nil && ev.ZoneRoute.BatchLegs != nil {
		return ReplacementResumeMoveBatch
	}
	if kind == RepEventMove && ev.ZoneRoute != nil && ev.ZoneRoute.Exchange != nil {
		return ReplacementResumeMoveExchange
	}
	if kind == RepEventMove && ev.EntryTail != nil && ev.EntryTail.Keyed != nil {
		return ReplacementResumeMoveEntryTail
	}
	return replacementResumeTailedKeys[kind]
}

// snapshotCarriesDataTail is eventCarriesDataTail over the mirror.
func snapshotCarriesDataTail(ev *replacementEventSnapshot) bool {
	if t := ev.EntryTail; t != nil && (t.Search != nil || t.Batch != nil || t.Tokens != nil || t.Keyed != nil) {
		return true
	}
	if r := ev.ZoneRoute; r != nil && (r.FaceDownLegs != nil || r.BatchLegs != nil || r.DiscardLegs != nil || r.Exchange != nil) {
		return true
	}
	if t := ev.KeywordActionTail; t != nil && t.Then != nil {
		return true
	}
	if t := ev.DamageTail; t != nil && (t.Then != nil || t.Each != nil) {
		return true
	}
	return ev.CounterTail != nil || ev.LifeTail != nil || ev.TokenTail != nil || ev.MillTail != nil
}

// KnownReplacementResume reports whether key names a registered tail.
func KnownReplacementResume(key string) bool {
	_, ok := replacementResumeTails[key]
	return ok
}

// replacementResumeKeyFor is the tail key for an event kind, or "".
func replacementResumeKeyFor(kind ReplacementEventKind) string {
	return replacementResumeBaseKeys[kind]
}

// keyedReplacementPromptKind reports whether a prompt of this kind
// resumes through a resolve path that reads nothing but the frame and
// the prompt's own written fields: ResolveOptionalReplacement,
// ResolveReplacementOrder, ResolveEntryRiot, ResolveCopyTarget,
// ResolveEntryController, ResolveEntryCardChoice, ResolveAuraHost,
// ResolveEntryReadAhead and ResolveEntryPayLife.
func keyedReplacementPromptKind(k PendingChoiceKind) bool {
	switch k {
	case PendingChoiceOptionalReplacement, PendingChoiceReplacementOrder, PendingChoiceEntryRiot,
		// #4253 slice 2: the entry questions that share the frame.
		PendingChoiceCopyTarget, PendingChoiceEntryController, PendingChoiceAuraHost,
		PendingChoiceEntryReadAhead, PendingChoiceEntryPayLife:
		return true
	}
	return isEntryCardChoiceKind(k)
}

// keyedPromptNeedsApplicable reports whether a keyed prompt kind's
// resolve path reads the frame's first gathered effect. An aura_host
// prompt is the one that does not: CR 303.4f's question is asked by the
// apply-loop itself, once nothing is left to apply, so its frame holds
// the event alone.
func keyedPromptNeedsApplicable(k PendingChoiceKind) bool {
	return k != PendingChoiceAuraHost
}

// entryCardChoiceZone is the zone an entry card choice's prompt kind
// picks from: EntryCardChoice.zone for the declaration that queues it.
func entryCardChoiceZone(k PendingChoiceKind) ZoneKind {
	if k == PendingChoiceEntrySacrifice || k == PendingChoiceEntryChoosePermanents {
		return ZoneBattlefield
	}
	return ZoneHand
}

// registeredReplacementResumeKeys lists the ledger lines for the tails
// and for the three body registries a keyed event can name: "resume
// <key>", "drawinstead <key>", "drawthen <key>", "keywordinstead <key>".
func registeredReplacementResumeKeys() []string {
	var out []string
	for k := range replacementResumeTails {
		out = append(out, "resume "+k)
	}
	drawInsteadMu.RLock()
	for k := range drawInsteadBodies {
		out = append(out, "drawinstead "+k)
	}
	drawInsteadMu.RUnlock()
	drawThenMu.RLock()
	for k := range drawThenBodies {
		out = append(out, "drawthen "+k)
	}
	drawThenMu.RUnlock()
	keywordInsteadMu.RLock()
	for k := range keywordInsteadBodies {
		out = append(out, "keywordinstead "+k)
	}
	keywordInsteadMu.RUnlock()
	out = append(out, registeredTailKeys()...)
	out = append(out, registeredMillUntilKeys()...)
	sort.Strings(out)
	return out
}

// KnownDrawInstead, KnownDrawThen and KnownKeywordInstead report whether
// a body is registered under key.
func KnownDrawInstead(key string) bool {
	drawInsteadMu.RLock()
	defer drawInsteadMu.RUnlock()
	_, ok := drawInsteadBodies[key]
	return ok
}

func KnownDrawThen(key string) bool {
	drawThenMu.RLock()
	defer drawThenMu.RUnlock()
	_, ok := drawThenBodies[key]
	return ok
}

func KnownKeywordInstead(key string) bool {
	keywordInsteadMu.RLock()
	defer keywordInsteadMu.RUnlock()
	_, ok := keywordInsteadBodies[key]
	return ok
}

// --- the event's fields -----------------------------------------------

// replacementEventField says what a keyed frame does with one field of
// ReplacementEvent.
type replacementEventField uint8

const (
	// repFieldCarried: written with the frame and put back.
	repFieldCarried replacementEventField = iota + 1
	// repFieldGate: not written. A frame whose event has it set is not
	// keyed (it is counted by the census, as before): it holds a closure,
	// or it belongs to an event kind this file does not key.
	repFieldGate
	// repFieldCache: not written; the running binary recomputes it.
	repFieldCache
)

// replacementEventFields classifies every field of ReplacementEvent.
// TestReplacementEventFieldsAreClassified fails on a field missing from
// it, and an unclassified field is treated as a gate, so a new field
// keeps its frames off the disk until someone decides.
var replacementEventFields = map[string]replacementEventField{
	"ID":            repFieldCarried,
	"Kind":          repFieldCarried,
	"Canceled":      repFieldCarried,
	"Actor":         repFieldCarried,
	"Source":        repFieldCarried,
	"mustSettleNow": repFieldCarried,

	// A draw.
	"DrawPlayer":     repFieldCarried,
	"DrawCount":      repFieldCarried,
	"drawPending":    repFieldCarried,
	"drawInstead":    repFieldCarried,
	"drawTail":       repFieldCarried,
	"drawThen":       repFieldCarried,
	"keywordInstead": repFieldCarried,

	// A keyword action. Its tail's data half (a proliferate's lists or
	// its "ask", a scry's or surveil's prompt kind, an earthbend's land)
	// is carried; a tail with a closure "then" keeps the frame off the
	// disk (replacementFrameKeyable).
	// Learn has no tail at all.
	"KeywordAction":      repFieldCarried,
	"KeywordActionCount": repFieldCarried,
	"keywordAction":      repFieldCarried, // its data half; see replacementFrameKeyable

	// A zone move and a battlefield entry.
	"CardID":                    repFieldCarried,
	"OldZone":                   repFieldCarried,
	"NewZone":                   repFieldCarried,
	"NewZoneOwner":              repFieldCarried,
	"ShuffleDestinationLibrary": repFieldCarried,
	"ExiledWith":                repFieldCarried,
	"ExileCounter":              repFieldCarried,
	"Destruction":               repFieldCarried,
	"CantBeRegenerated":         repFieldCarried,
	"StateBasedDestruction":     repFieldCarried,
	"LeaveCause":                repFieldCarried,
	"EntersTapped":              repFieldCarried,
	"EntersPrepared":            repFieldCarried,
	"EntersDevoured":            repFieldCarried,
	"EntersWithHaste":           repFieldCarried,
	"lookAhead":                 repFieldCache,
	"entryRefused":              repFieldCarried,
	"EntersUnlocked":            repFieldCarried,
	"EntersAttacking":           repFieldCarried,
	"FaceDown":                  repFieldCarried,
	"FaceDownListed":            repFieldCarried,
	"EntersAsCopyOf":            repFieldCarried,
	"copySourceID":              repFieldCarried,
	"copyUntilEndOfTurn":        repFieldCarried,
	"auraHost":                  repFieldCarried,
	"auraHostFixed":             repFieldCarried,
	"auraHostSettled":           repFieldCarried,
	"entersAsAura":              repFieldCarried,
	"entersAs":                  repFieldCarried,
	"exchangeTextWith":          repFieldCarried,
	"stackItem":                 repFieldCarried,
	"putByEntering":             repFieldCarried,
	"EntersWithCounters":        repFieldCarried,
	"entryResumable":            repFieldCarried,
	"entryTail":                 repFieldCarried, // its data half; see replacementFrameKeyable
	"landPlay":                  repFieldCarried,
	"landPlayer":                repFieldCarried,
	"entryControllerSet":        repFieldCarried,
	"entryControllerPrior":      repFieldCarried,
	"zoneRoute":                 repFieldCarried, // its data half; see replacementFrameKeyable
	"lifeAfterMove":             repFieldCarried,
	"extraTurnsAfterMove":       repFieldCarried,
	"reflexiveAfterMove":        repFieldCarried,
	"asCommanderMove":           repFieldCarried,
	"commanderAnswer":           repFieldCarried,

	// A counter placement and a life change (#4253 slice 2). Their
	// tails hold closures, so a placement or change with one gates.
	"CounterTarget":              repFieldCarried,
	"CounterName":                repFieldCarried,
	"CounterDelta":               repFieldCarried,
	"CounterPlayer":              repFieldCarried,
	"StartingLoyalty":            repFieldCarried,
	"CounterPlacer":              repFieldCarried,
	"CounterFromCombatDamage":    repFieldCarried,
	"CounterFromCost":            repFieldCarried,
	"CounterFromTurnBasedAction": repFieldCarried,
	"LifePlayer":                 repFieldCarried,
	"LifeDelta":                  repFieldCarried,
	"LifeFromRadiation":          repFieldCarried,
	// #4253 slice 3: a tail's data half — a keyed continuation, or a
	// lose-life batch's remaining players — is carried; a closure gates
	// (replacementEventKeyable).
	"counterTail": repFieldCarried,
	"lifeTail":    repFieldCarried,

	// A token creation (#4253 slice 3). Its tail's data half — a keyed
	// continuation — is carried; a closure `then` gates
	// (replacementFrameKeyable), and so does a template the card mirror
	// cannot write whole.
	"TokenController":         repFieldCarried,
	"TokenGroups":             repFieldCarried,
	"TokenAttacking":          repFieldCarried,
	"TokenTogether":           repFieldCarried,
	"TokenAttackChoicesLabel": repFieldCarried,
	"tokenTail":               repFieldCarried, // its data half; see replacementEventKeyable

	// A discard (#4253 slice 9). Its continuation is the route's
	// discardLegs, carried with the route.
	"DiscardPlayer":          repFieldCarried,
	"DiscardCause":           repFieldCarried,
	"DiscardCauseController": repFieldCarried,

	// A mill instruction's amount window (#4253 slice 10a). Its tail is
	// data (#4361 S9 deleted the closure `then`).
	"MillPlayer": repFieldCarried,
	"MillCount":  repFieldCarried,
	"mill":       repFieldCarried, // its data half; see replacementEventKeyable

	// A damage event (#4253, damage_keyed.go). Its tail's data half is
	// carried. Combat damage is carried since #4362, under a key of its own
	// (combat_damage_keyed.go), and redirected parts since #4404
	// (damage_redirected_keyed.go).
	"DamageSource":   repFieldCarried,
	"DamageTarget":   repFieldCarried,
	"DamageAmount":   repFieldCarried,
	"DamageInstance": repFieldCarried,
	"SourceLKI":      repFieldCarried,
	"damageTail":     repFieldCarried, // its data half; see damageEventIsData
	"IsCombatDamage": repFieldCarried,

	// The event kinds this file does not key.
	"ManaPlayer":         repFieldGate,
	"ManaSource":         repFieldGate,
	"ManaColors":         repFieldGate,
	"ManaFromTap":        repFieldGate,
	"StepTransitionStep": repFieldGate,
	"StepTransitionSeat": repFieldGate,
	"ExtraTurnSeat":      repFieldGate,
	"ExtraTurnRef":       repFieldGate,
	"CounterSpellID":     repFieldGate,
	"CounterBy":          repFieldGate,
	"CounterBySource":    repFieldGate,
	"counterInstead":     repFieldGate,
	// #4307: a dice roll's window never pauses (mustSettleNow).
	"RollPlayer":  repFieldGate,
	"RollSides":   repFieldGate,
	"RollCount":   repFieldGate,
	"rollShields": repFieldGate,
}

// replacementEventGatesClear reports whether every gate field of ev is
// at its zero value. An unclassified field counts as a gate.
func replacementEventGatesClear(ev *ReplacementEvent) bool {
	v := reflect.ValueOf(ev).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		class := replacementEventFields[t.Field(i).Name]
		if class == repFieldCarried || class == repFieldCache {
			continue
		}
		if !v.Field(i).IsZero() {
			return false
		}
	}
	return true
}

// replacementFrameKeyable reports the tail key a prompt's frame is
// written under, or "" when the frame stays a closure.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) replacementFrameKeyable(c *PendingChoice) string {
	if c == nil || !keyedReplacementPromptKind(c.Kind) || c.costCommanderResume != nil {
		return ""
	}
	f := c.replacementResume
	if f == nil || f.ev == nil || (len(f.applicable) == 0 && keyedPromptNeedsApplicable(c.Kind)) {
		return ""
	}
	ev := f.ev
	key := replacementResumeKeyFor(ev.Kind)
	if key == "" || !g.replacementEventKeyable(ev) {
		return ""
	}
	if eventCarriesDataTail(ev) || (ev.Kind == RepEventMove && len(ev.reflexiveAfterMove) > 0) {
		key = replacementResumeTailedKeyOf(ev)
	}
	if ev.Kind == RepEventDamage {
		// #4362: combat damage has a key of its own, so a binary that
		// would resume it as noncombat damage refuses the file. #4404: so
		// does an event with redirected parts, which a binary from before
		// them would drop.
		key = damageEventResumeKey(ev)
	}
	if !applicableRegathered(f.applicable) || frameDepartedUnregatherable(f) {
		return ""
	}
	return key
}

// applicableRegathered reports whether a restored binary could gather
// every effect of a frame again by its ID.
func applicableRegathered(applicable []activeReplacement) bool {
	for _, a := range applicable {
		if a.id >= testReplacementIDBase && a.id < builtinReplacementIDBase {
			// A test-injected effect is not in the restore point, so the
			// restored binary could not gather it again.
			return false
		}
	}
	return true
}

// replacementEventKeyable reports whether an event, its tail included,
// is data: every gate field clear and no tail holding a closure. A
// paused simultaneous entry's settled member events are checked the same
// way (entryBatchKeyable).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) replacementEventKeyable(ev *ReplacementEvent) bool {
	return g.replacementEventKeyableAs(ev, false)
}

// replacementEventKeyableAs is replacementEventKeyable for an event that
// may be a settled member of a batch of created tokens (#4404): its token
// is staged, and the batch, not the event, carries the rest of the entry
// (entryBatchKeyable checks the staged tokens).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) replacementEventKeyableAs(ev *ReplacementEvent, batchToken bool) bool {
	if !replacementEventGatesClear(ev) {
		return false
	}
	if t := ev.keywordAction; t != nil && !reflect.ValueOf(t.armyToken).IsZero() {
		// #4253 slice 10b: an amass's Army token template is written
		// through the card mirror, as a token group's is; one the mirror
		// cannot carry whole keeps the frame off the disk.
		var scratch ContinuationCensus
		_ = snapshotCard(t.armyToken, &scratch)
		if !scratch.Empty() {
			return false
		}
	}
	if ev.CardID != uuid.Nil && !batchToken {
		if _, staged := g.enteringTokenLocked(ev.CardID); staged {
			// A created token's entry: the token is in no zone, only in
			// Game.enteringTokens. #4253 slice 3: that is written with a
			// keyed frame (GameSnapshot.EnteringTokens), so the entry is
			// data when the rest of its batch is (entryTail.tokens, or
			// since #4404 a simultaneous entry, entryTail.batch, checked
			// below) and every staged token is written whole.
			t := ev.entryTail
			if t == nil || (t.tokens == nil && t.batch == nil) || !g.stagedTokensAreData() {
				return false
			}
		}
	}
	// #4361 S17: the rest of a replaced move's life changes (ADR 0178,
	// lifeTail.move) is written with the life tail, and its route rides
	// the event's zoneRoute, checked below like any other.
	if ev.Kind == RepEventDamage && !g.damageEventIsData(ev) {
		return false
	}
	for _, grp := range ev.TokenGroups {
		var scratch ContinuationCensus
		_ = snapshotCard(grp.Template, &scratch)
		if !scratch.Empty() {
			return false
		}
	}
	if r := ev.zoneRoute; r != nil {
		// #4253 slice 10a: an exchange's settled entry is data
		// (settledExchangeEntry), written beside the route. #4361 S17: a
		// route's continuation is always data now (zoneRoute.then is
		// gone).
		//
		// #4253 slice 9: a keyed batch's last-known copies are written
		// beside its run, through the card mirror; any other route with
		// them, or copies the mirror cannot write whole, stays a closure.
		if len(r.simultaneousExit) > 0 && (r.batchLegs == nil || !routeBatchLKIIsData(r.simultaneousExit)) {
			return false
		}
	}
	if t := ev.entryTail; t != nil {
		if t.batch != nil && !g.entryBatchKeyable(t.batch) {
			// #4253 slice 3: a simultaneous entry is data when its
			// continuation is and its settled members are.
			return false
		}
	}
	if ev.stackItem != nil {
		// The resolving spell a permanent's entry carries. Written with
		// the frame through the stack-item mirror; one the mirror cannot
		// carry whole keeps the frame off the disk.
		var scratch ContinuationCensus
		_ = snapshotStackItem(g, ev.stackItem, &scratch)
		if !scratch.Empty() {
			return false
		}
	}
	return true
}

// --- the mirror -------------------------------------------------------

// replacementResumeSnapshot is a keyed frame on disk
// (pendingChoiceSnapshot.ReplacementResume). Its tail key rides the
// prompt's pickThen.
type replacementResumeSnapshot struct {
	Event replacementEventSnapshot `json:"event"`
	// Applicable is the pending replacement set, in gathered order:
	// replacementResumeFrame.applicable by ID. Restore gathers the event
	// again and picks these.
	Applicable []ReplacementEffectID `json:"applicable"`
	// Applied is Game.replacementsAppliedThisEvent for the event, sorted:
	// the effects that have already applied to it (CR 614.5).
	Applied []ReplacementEffectID `json:"applied,omitempty"`
}

// replacementEventSnapshot is a ReplacementEvent's carried fields.
type replacementEventSnapshot struct {
	ID            ReplacementEventID   `json:"id"`
	Kind          ReplacementEventKind `json:"kind"`
	Canceled      bool                 `json:"canceled,omitempty"`
	Actor         *uuid.UUID           `json:"actor,omitempty"`
	Source        *uuid.UUID           `json:"source,omitempty"`
	MustSettleNow bool                 `json:"mustSettleNow,omitempty"`

	DrawPlayer     *uuid.UUID            `json:"drawPlayer,omitempty"`
	DrawCount      int                   `json:"drawCount,omitempty"`
	DrawPending    []pendingDrawSnapshot `json:"drawPending,omitempty"`
	DrawInstead    *insteadRunSnapshot   `json:"drawInstead,omitempty"`
	DrawTail       int                   `json:"drawTail,omitempty"`
	DrawThen       *drawThenSnapshot     `json:"drawThen,omitempty"`
	KeywordInstead *insteadRunSnapshot   `json:"keywordInstead,omitempty"`

	KeywordAction      KeywordAction `json:"keywordAction,omitempty"`
	KeywordActionCount int           `json:"keywordActionCount,omitempty"`

	CardID                    *uuid.UUID          `json:"cardId,omitempty"`
	OldZone                   ZoneKind            `json:"oldZone,omitempty"`
	NewZone                   ZoneKind            `json:"newZone,omitempty"`
	NewZoneOwner              *uuid.UUID          `json:"newZoneOwner,omitempty"`
	ShuffleDestinationLibrary bool                `json:"shuffleDestinationLibrary,omitempty"`
	ExiledWith                *PermissionCardRef  `json:"exiledWith,omitempty"`
	ExileCounter              string              `json:"exileCounter,omitempty"`
	Destruction               bool                `json:"destruction,omitempty"`
	CantBeRegenerated         bool                `json:"cantBeRegenerated,omitempty"`
	StateBasedDestruction     bool                `json:"stateBasedDestruction,omitempty"`
	LeaveCause                LeaveCause          `json:"leaveCause,omitempty"`
	EntersTapped              bool                `json:"entersTapped,omitempty"`
	EntersPrepared            bool                `json:"entersPrepared,omitempty"`
	EntersDevoured            int                 `json:"entersDevoured,omitempty"`
	EntersWithHaste           bool                `json:"entersWithHaste,omitempty"`
	EntryRefused              bool                `json:"entryRefused,omitempty"`
	EntersUnlocked            DoorMask            `json:"entersUnlocked,omitempty"`
	EntersAttacking           *uuid.UUID          `json:"entersAttacking,omitempty"`
	FaceDown                  FaceDownKind        `json:"faceDown,omitempty"`
	FaceDownListed            *FaceDownListing    `json:"faceDownListed,omitempty"`
	EntersAsCopyOf            *PrintedValues      `json:"entersAsCopyOf,omitempty"`
	CopySource                *uuid.UUID          `json:"copySource,omitempty"`
	CopyUntilEndOfTurn        bool                `json:"copyUntilEndOfTurn,omitempty"`
	AuraHost                  *TargetRef          `json:"auraHost,omitempty"`
	AuraHostFixed             bool                `json:"auraHostFixed,omitempty"`
	AuraHostSettled           bool                `json:"auraHostSettled,omitempty"`
	EntersAsAura              bool                `json:"entersAsAura,omitempty"`
	EntersAs                  *EntersAs           `json:"entersAs,omitempty"`
	ExchangeTextWith          *uuid.UUID          `json:"exchangeTextWith,omitempty"`
	StackItem                 *stackItemSnapshot  `json:"stackItem,omitempty"`
	PutByEntering             bool                `json:"putByEntering,omitempty"`
	EntersWithCounters        map[string]int      `json:"entersWithCounters,omitempty"`
	EntryResumable            bool                `json:"entryResumable,omitempty"`
	EntryTail                 *entryTailSnapshot  `json:"entryTail,omitempty"`
	LandPlay                  bool                `json:"landPlay,omitempty"`
	LandPlayer                *uuid.UUID          `json:"landPlayer,omitempty"`
	EntryControllerSet        bool                `json:"entryControllerSet,omitempty"`
	EntryControllerPrior      *uuid.UUID          `json:"entryControllerPrior,omitempty"`
	ZoneRoute                 *zoneRouteSnapshot  `json:"zoneRoute,omitempty"`
	LifeAfterMove             []moveLifeSnapshot  `json:"lifeAfterMove,omitempty"`
	ExtraTurnsAfterMove       []moveTurnSnapshot  `json:"extraTurnsAfterMove,omitempty"`
	AsCommanderMove           bool                `json:"asCommanderMove,omitempty"`
	CommanderAnswer           commanderZoneAnswer `json:"commanderAnswer,omitempty"`

	// #4517: reflexive triggers a replacement declared on the move.
	ReflexiveAfterMove []moveReflexiveSnapshot `json:"reflexiveAfterMove,omitempty"`

	// #4253 slice 2: a keyword action's tail (its data half), a counter
	// placement and a life change.
	KeywordActionTail          *keywordActionTailSnapshot `json:"keywordActionTail,omitempty"`
	CounterTarget              *uuid.UUID                 `json:"counterTarget,omitempty"`
	CounterName                string                     `json:"counterName,omitempty"`
	CounterDelta               int                        `json:"counterDelta,omitempty"`
	CounterPlayer              *uuid.UUID                 `json:"counterPlayer,omitempty"`
	StartingLoyalty            bool                       `json:"startingLoyalty,omitempty"`
	CounterPlacer              *uuid.UUID                 `json:"counterPlacer,omitempty"`
	CounterFromCombatDamage    bool                       `json:"counterFromCombatDamage,omitempty"`
	CounterFromCost            bool                       `json:"counterFromCost,omitempty"`
	CounterFromTurnBasedAction bool                       `json:"counterFromTurnBasedAction,omitempty"`
	LifePlayer                 *uuid.UUID                 `json:"lifePlayer,omitempty"`
	LifeDelta                  int                        `json:"lifeDelta,omitempty"`
	LifeFromRadiation          bool                       `json:"lifeFromRadiation,omitempty"`
	// #4253 slice 9: a discard.
	DiscardPlayer          *uuid.UUID   `json:"discardPlayer,omitempty"`
	DiscardCause           DiscardCause `json:"discardCause,omitempty"`
	DiscardCauseController *uuid.UUID   `json:"discardCauseController,omitempty"`

	// #4253 slice 3: a token creation instruction.
	TokenController         *uuid.UUID           `json:"tokenController,omitempty"`
	TokenGroups             []tokenGroupSnapshot `json:"tokenGroups,omitempty"`
	TokenAttacking          *uuid.UUID           `json:"tokenAttacking,omitempty"`
	TokenTogether           bool                 `json:"tokenTogether,omitempty"`
	TokenAttackChoicesLabel string               `json:"tokenAttackChoicesLabel,omitempty"`
	TokenTail               *tokenTailSnapshot   `json:"tokenTail,omitempty"`
	// A counter placement's and a life change's tail, as data.
	CounterTail *counterTailSnapshot `json:"counterTail,omitempty"`
	LifeTail    *lifeTailSnapshot    `json:"lifeTail,omitempty"`

	// #4253 slice 10a: a mill instruction's amount window.
	MillPlayer *uuid.UUID        `json:"millPlayer,omitempty"`
	MillCount  int               `json:"millCount,omitempty"`
	MillTail   *millTailSnapshot `json:"millTail,omitempty"`

	// #4253: a damage event (damage_keyed.go).
	DamageSource   *uuid.UUID          `json:"damageSource,omitempty"`
	DamageTarget   *uuid.UUID          `json:"damageTarget,omitempty"`
	DamageAmount   int                 `json:"damageAmount,omitempty"`
	DamageInstance DamageInstance      `json:"damageInstance,omitempty"`
	SourceLKI      *Characteristic     `json:"sourceLki,omitempty"`
	DamageTail     *damageTailSnapshot `json:"damageTail,omitempty"`
	// #4362: combat damage, written under ReplacementResumeCombatDamage.
	IsCombatDamage bool `json:"isCombatDamage,omitempty"`
}

// millTailSnapshot is a millTail's data half, present exactly when the
// event had a tail: a tail with a closure `then` is never written.
// Sequenced is whether the keyed door asked for one (millTail.keyed set,
// even to the zero tail), which sequences the mill.
type millTailSnapshot struct {
	Dest      ZoneKind           `json:"dest"`
	AllMoved  bool               `json:"allMoved,omitempty"`
	Sequenced bool               `json:"sequenced,omitempty"`
	Then      *keyedTailSnapshot `json:"then,omitempty"`
	Run       *millUntilRun      `json:"run,omitempty"`
}

// exchangeEntrySnapshot is a settledExchangeEntry on disk (#4253 slice
// 10a): the settled half of an exchange, landed when the exit settles.
type exchangeEntrySnapshot struct {
	CardID             uuid.UUID        `json:"cardId"`
	Actor              *uuid.UUID       `json:"actor,omitempty"`
	NewZoneOwner       *uuid.UUID       `json:"newZoneOwner,omitempty"`
	Attacking          *uuid.UUID       `json:"attacking,omitempty"`
	CopySource         *uuid.UUID       `json:"copySource,omitempty"`
	OldZone            ZoneKind         `json:"oldZone,omitempty"`
	NewZone            ZoneKind         `json:"newZone,omitempty"`
	Tapped             bool             `json:"tapped,omitempty"`
	Prepared           bool             `json:"prepared,omitempty"`
	Haste              bool             `json:"haste,omitempty"`
	CopyUntilEndOfTurn bool             `json:"copyUntilEndOfTurn,omitempty"`
	Devoured           int              `json:"devoured,omitempty"`
	Counters           map[string]int   `json:"counters,omitempty"`
	Copy               *PrintedValues   `json:"copy,omitempty"`
	Unlocked           DoorMask         `json:"unlocked,omitempty"`
	FaceDown           FaceDownKind     `json:"faceDown,omitempty"`
	Listed             *FaceDownListing `json:"listed,omitempty"`
}

func snapshotExchangeEntry(e *settledExchangeEntry) *exchangeEntrySnapshot {
	if e == nil {
		return nil
	}
	return &exchangeEntrySnapshot{
		CardID: e.CardID, Actor: uuidPtrOrNil(e.Actor), NewZoneOwner: uuidPtrOrNil(e.NewZoneOwner),
		Attacking: uuidPtrOrNil(e.Attacking), CopySource: uuidPtrOrNil(e.CopySource),
		OldZone: e.OldZone, NewZone: e.NewZone, Tapped: e.Tapped, Prepared: e.Prepared, Haste: e.Haste,
		CopyUntilEndOfTurn: e.CopyUntilEndOfTurn, Devoured: e.Devoured, Counters: copyStringIntMap(e.Counters),
		Copy: copyPrintedValues(e.Copy), Unlocked: e.Unlocked, FaceDown: e.FaceDown, Listed: copyFaceDownListing(e.Listed),
	}
}

func restoreExchangeEntry(s *exchangeEntrySnapshot) *settledExchangeEntry {
	if s == nil {
		return nil
	}
	return &settledExchangeEntry{
		CardID: s.CardID, Actor: uuidOrNil(s.Actor), NewZoneOwner: uuidOrNil(s.NewZoneOwner),
		Attacking: uuidOrNil(s.Attacking), CopySource: uuidOrNil(s.CopySource),
		OldZone: s.OldZone, NewZone: s.NewZone, Tapped: s.Tapped, Prepared: s.Prepared, Haste: s.Haste,
		CopyUntilEndOfTurn: s.CopyUntilEndOfTurn, Devoured: s.Devoured, Counters: copyStringIntMap(s.Counters),
		Copy: copyPrintedValues(s.Copy), Unlocked: s.Unlocked, FaceDown: s.FaceDown, Listed: copyFaceDownListing(s.Listed),
	}
}

func snapshotMillTail(t *millTail) *millTailSnapshot {
	if t == nil {
		return nil
	}
	out := &millTailSnapshot{Dest: t.dest, AllMoved: t.allMoved, Run: cloneMillUntilRun(t.run)}
	if t.keyed != nil {
		out.Sequenced = true
		out.Then = snapshotKeyedTail(*t.keyed)
	}
	return out
}

func restoreMillTail(s *millTailSnapshot) *millTail {
	if s == nil {
		return nil
	}
	out := &millTail{dest: s.Dest, allMoved: s.AllMoved, run: cloneMillUntilRun(s.Run)}
	if s.Sequenced {
		k := restoreKeyedTail(s.Then)
		out.keyed = &k
	}
	return out
}

// counterTailSnapshot is a counterTail's data half, present exactly when
// the event had a tail.
type counterTailSnapshot struct {
	Then *keyedTailSnapshot `json:"then,omitempty"`
}

// lifeTailSnapshot is a lifeTail's data half, present exactly when the
// event had a tail.
type lifeTailSnapshot struct {
	Then *keyedTailSnapshot `json:"then,omitempty"`
	Each *lifeEachRun       `json:"each,omitempty"`
	// Move is the rest of a replaced move's declared life changes (ADR
	// 0178, lifeTail.move, #4361 S17); the move's route rides the event's
	// ZoneRoute. Additive within v9.
	Move *moveLifeRestSnapshot `json:"move,omitempty"`
}

// moveLifeRestSnapshot is a moveLifeRest.
type moveLifeRestSnapshot struct {
	Remaining     []moveLifeSnapshot `json:"remaining,omitempty"`
	HasRoute      bool               `json:"hasRoute,omitempty"`
	MustSettleNow bool               `json:"mustSettleNow,omitempty"`
}

func snapshotMoveLifeRest(r *moveLifeRest) *moveLifeRestSnapshot {
	if r == nil {
		return nil
	}
	out := &moveLifeRestSnapshot{HasRoute: r.hasRoute, MustSettleNow: r.mustSettleNow}
	for _, l := range r.remaining {
		out.Remaining = append(out.Remaining, moveLifeSnapshot{Source: uuidPtrOrNil(l.Source), Player: l.Player, Delta: l.Delta})
	}
	return out
}

func restoreMoveLifeRest(s *moveLifeRestSnapshot) *moveLifeRest {
	if s == nil {
		return nil
	}
	out := &moveLifeRest{hasRoute: s.HasRoute, mustSettleNow: s.MustSettleNow}
	for _, l := range s.Remaining {
		out.remaining = append(out.remaining, lifeInstruction{Source: uuidOrNil(l.Source), Player: l.Player, Delta: l.Delta})
	}
	return out
}

// tokenGroupSnapshot is a TokenGroup: its template through the card
// mirror.
type tokenGroupSnapshot struct {
	Template       cardSnapshot       `json:"template"`
	Count          int                `json:"count"`
	Entry          tokenEntrySnapshot `json:"entry"`
	Attacking      *uuid.UUID         `json:"attacking,omitempty"`
	AttackOpponent *uuid.UUID         `json:"attackOpponent,omitempty"`
	CreatedWith    PermissionCardRef  `json:"createdWith"`
	Blocking       *uuid.UUID         `json:"blocking,omitempty"`
}

// tokenTailSnapshot is a tokenTail's data half: present exactly when the
// event had a tail. A tail with a closure is never written; a
// synchronous door's result is not either.
type tokenTailSnapshot struct {
	Then *keyedTailSnapshot `json:"then,omitempty"`
}

// keywordActionTailSnapshot is a keywordActionTail's data half: a tail
// with a "then", an amass's Army or "the Army you amassed" is never
// written. Present (even empty) exactly when the event had a tail.
type keywordActionTailSnapshot struct {
	Cards   []uuid.UUID       `json:"cards,omitempty"`
	Players []uuid.UUID       `json:"players,omitempty"`
	Ask     bool              `json:"ask,omitempty"`
	Choice  PendingChoiceKind `json:"choice,omitempty"`
	Land    *uuid.UUID        `json:"land,omitempty"`
	// #4253 slice 3: a keyed "then" (Preordain's draw).
	Then *keyedTailSnapshot `json:"then,omitempty"`
	// #4253 slice 10b: an amass's two printed parameters (CR 701.47a),
	// the Army token template through the card mirror and the creature
	// type.
	ArmyToken   *cardSnapshot `json:"armyToken,omitempty"`
	ArmySubtype string        `json:"armySubtype,omitempty"`
}

// pendingDrawSnapshot is a pendingDraw: an expanded draw's sibling.
type pendingDrawSnapshot struct {
	Player  uuid.UUID             `json:"player"`
	Count   int                   `json:"count"`
	Applied []ReplacementEffectID `json:"applied,omitempty"`
}

// insteadRunSnapshot is a recorded draw-instead or keyword-instead body:
// its registered key and the card it belongs to.
type insteadRunSnapshot struct {
	Key    string     `json:"key"`
	Source *uuid.UUID `json:"source,omitempty"`
}

// drawThenSnapshot is a DrawThen.
type drawThenSnapshot struct {
	Key    string     `json:"key"`
	Player *uuid.UUID `json:"player,omitempty"`
	Source *uuid.UUID `json:"source,omitempty"`
	N      int        `json:"n,omitempty"`
}

// entryTailSnapshot is an entryTail's data half: a tail with a `then`
// or a batch is never written.
type entryTailSnapshot struct {
	NewObject   bool        `json:"newObject,omitempty"`
	CraftedWith []ObjectRef `json:"craftedWith,omitempty"`
	BatchIndex  int         `json:"batchIndex,omitempty"`
	// #4253 slice 3: a library search's remaining work, the
	// simultaneous entry this event is one card of, and a token batch's
	// remaining entries.
	Search *searchEntryRun     `json:"search,omitempty"`
	Batch  *entryBatchSnapshot `json:"batch,omitempty"`
	Tokens *tokenEntryRun      `json:"tokens,omitempty"`
	// #4253 slice 6: the caller's "then" of a single permanent's return,
	// as a registered tail.
	Keyed *keyedTailSnapshot `json:"keyed,omitempty"`
}

// zoneRouteSnapshot is a zoneRoute's data half: a route with a `then`,
// a simultaneous exit or an exchange entry is never written.
type zoneRouteSnapshot struct {
	CardID              *uuid.UUID          `json:"cardId,omitempty"`
	Dst                 ZoneKind            `json:"dst,omitempty"`
	DstOwner            *uuid.UUID          `json:"dstOwner,omitempty"`
	Actor               *uuid.UUID          `json:"actor,omitempty"`
	ToBottom            bool                `json:"toBottom,omitempty"`
	Depth               int                 `json:"depth,omitempty"`
	FaceDown            FaceDownKind        `json:"faceDown,omitempty"`
	HiddenBy            *PermissionCardRef  `json:"hiddenBy,omitempty"`
	ExiledWith          *PermissionCardRef  `json:"exiledWith,omitempty"`
	ExiledWithAbility   *AbilityRef         `json:"exiledWithAbility,omitempty"`
	Mill                bool                `json:"mill,omitempty"`
	Discard             bool                `json:"discard,omitempty"`
	DiscardCause        DiscardCause        `json:"discardCause,omitempty"`
	Source              *uuid.UUID          `json:"source,omitempty"`
	MustSettleNow       bool                `json:"mustSettleNow,omitempty"`
	Countered           bool                `json:"countered,omitempty"`
	CauseKind           MoveCauseKind       `json:"causeKind,omitempty"`
	CauseController     *uuid.UUID          `json:"causeController,omitempty"`
	CauseItem           *uuid.UUID          `json:"causeItem,omitempty"`
	Sacrifice           bool                `json:"sacrifice,omitempty"`
	ViaBattlefieldLeave bool                `json:"viaBattlefieldLeave,omitempty"`
	Destruction         bool                `json:"destruction,omitempty"`
	CantBeRegenerated   bool                `json:"cantBeRegenerated,omitempty"`
	StateBased          bool                `json:"stateBased,omitempty"`
	AsCommander         bool                `json:"asCommander,omitempty"`
	CommanderAnswer     commanderZoneAnswer `json:"commanderAnswer,omitempty"`
	// #4253 slice 3: one leg of a face-down exile batch.
	FaceDownLegs *faceDownExileRun `json:"faceDownLegs,omitempty"`
	// #4253 slice 9: one leg of a keyed route batch with the batch's
	// last-known copies (CR 603.10), and one leg of a discard batch.
	BatchLegs        *routeBatchRun   `json:"batchLegs,omitempty"`
	SimultaneousExit []cardSnapshot   `json:"simultaneousExit,omitempty"`
	DiscardLegs      *discardBatchRun `json:"discardLegs,omitempty"`
	// #4253 slice 10a: the settled entry an exchange's exit lands with
	// (zoneRoute.simultaneousEntry, Goblin Welder).
	Exchange *exchangeEntrySnapshot `json:"exchange,omitempty"`
}

// moveLifeSnapshot is a lifeInstruction (ADR 0178).
type moveLifeSnapshot struct {
	Source *uuid.UUID `json:"source,omitempty"`
	Player uuid.UUID  `json:"player"`
	Delta  int        `json:"delta"`
}

// moveTurnSnapshot is an extraTurnInstruction (ADR 0188).
type moveTurnSnapshot struct {
	Source *uuid.UUID `json:"source,omitempty"`
	Player uuid.UUID  `json:"player"`
}

// moveReflexiveSnapshot is a reflexiveMoveInstruction (#4517, ADR 0178
// amendment of 2026-10-29). Body is a ReflexiveBody key, checked at
// restore.
type moveReflexiveSnapshot struct {
	Source       *uuid.UUID    `json:"source,omitempty"`
	SourceObject *ObjectRef    `json:"sourceObject,omitempty"`
	Controller   uuid.UUID     `json:"controller"`
	Zone         ZoneKind      `json:"zone"`
	Label        string        `json:"label"`
	Body         string        `json:"body"`
	Params       *EffectParams `json:"params,omitempty"`
}

func snapshotMoveReflexive(r reflexiveMoveInstruction) moveReflexiveSnapshot {
	out := moveReflexiveSnapshot{Source: uuidPtrOrNil(r.Source), Controller: r.Controller, Zone: r.Zone, Label: r.Label, Body: r.Body}
	if r.SourceObject != (ObjectRef{}) {
		o := r.SourceObject
		out.SourceObject = &o
	}
	out.Params = effectParamsOrNil(r.Params)
	return out
}

func restoreMoveReflexive(s moveReflexiveSnapshot) reflexiveMoveInstruction {
	out := reflexiveMoveInstruction{Source: uuidOrNil(s.Source), Controller: s.Controller, Zone: s.Zone, Label: s.Label, Body: s.Body}
	if s.SourceObject != nil {
		out.SourceObject = *s.SourceObject
	}
	out.Params = effectParamsValue(s.Params)
	return out
}

// replacementSerialSnapshot is one entry of Game.replacementObjectSerials.
type replacementSerialSnapshot struct {
	Object ObjectRef `json:"object"`
	Serial int       `json:"serial"`
}

func permissionRefPtr(r PermissionCardRef) *PermissionCardRef {
	if r == (PermissionCardRef{}) {
		return nil
	}
	return &r
}

func permissionRefValue(r *PermissionCardRef) PermissionCardRef {
	if r == nil {
		return PermissionCardRef{}
	}
	return *r
}

func insteadRunOrNil(key string, source uuid.UUID) *insteadRunSnapshot {
	if key == "" {
		return nil
	}
	return &insteadRunSnapshot{Key: key, Source: uuidPtrOrNil(source)}
}

// snapshotReplacementResume writes a keyable frame. Caller has checked
// replacementFrameKeyable.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) snapshotReplacementResume(f *replacementResumeFrame, cen *ContinuationCensus) *replacementResumeSnapshot {
	out := &replacementResumeSnapshot{Event: g.snapshotReplacementEvent(f.ev, cen)}
	out.Applicable = make([]ReplacementEffectID, 0, len(f.applicable))
	for _, a := range f.applicable {
		out.Applicable = append(out.Applicable, a.id)
	}
	for id, applied := range g.replacementsAppliedThisEvent[f.ev.ID] {
		if applied {
			out.Applied = append(out.Applied, id)
		}
	}
	sort.Slice(out.Applied, func(i, j int) bool { return out.Applied[i] < out.Applied[j] })
	return out
}

func (g *Game) snapshotReplacementEvent(ev *ReplacementEvent, cen *ContinuationCensus) replacementEventSnapshot {
	out := replacementEventSnapshot{
		ID:                        ev.ID,
		Kind:                      ev.Kind,
		Canceled:                  ev.Canceled,
		Actor:                     uuidPtrOrNil(ev.Actor),
		Source:                    uuidPtrOrNil(ev.Source),
		MustSettleNow:             ev.mustSettleNow,
		DrawPlayer:                uuidPtrOrNil(ev.DrawPlayer),
		DrawCount:                 ev.DrawCount,
		DrawInstead:               insteadRunOrNil(ev.drawInstead.ref.key, ev.drawInstead.source),
		DrawTail:                  ev.drawTail,
		KeywordInstead:            insteadRunOrNil(ev.keywordInstead.ref.key, ev.keywordInstead.source),
		KeywordAction:             ev.KeywordAction,
		KeywordActionCount:        ev.KeywordActionCount,
		CardID:                    uuidPtrOrNil(ev.CardID),
		OldZone:                   ev.OldZone,
		NewZone:                   ev.NewZone,
		NewZoneOwner:              uuidPtrOrNil(ev.NewZoneOwner),
		ShuffleDestinationLibrary: ev.ShuffleDestinationLibrary,
		ExiledWith:                permissionRefPtr(ev.ExiledWith),
		ExileCounter:              ev.ExileCounter,
		Destruction:               ev.Destruction,
		CantBeRegenerated:         ev.CantBeRegenerated,
		StateBasedDestruction:     ev.StateBasedDestruction,
		LeaveCause:                ev.LeaveCause,
		EntersTapped:              ev.EntersTapped,
		EntersPrepared:            ev.EntersPrepared,
		EntersDevoured:            ev.EntersDevoured,
		EntersWithHaste:           ev.EntersWithHaste,
		EntryRefused:              ev.entryRefused,
		EntersUnlocked:            ev.EntersUnlocked,
		EntersAttacking:           uuidPtrOrNil(ev.EntersAttacking),
		FaceDown:                  ev.FaceDown,
		EntersAsCopyOf:            copyPrintedValues(ev.EntersAsCopyOf),
		CopySource:                uuidPtrOrNil(ev.copySourceID),
		CopyUntilEndOfTurn:        ev.copyUntilEndOfTurn,
		AuraHostFixed:             ev.auraHostFixed,
		AuraHostSettled:           ev.auraHostSettled,
		EntersAsAura:              ev.entersAsAura,
		EntersAs:                  copyEntersAs(ev.entersAs),
		ExchangeTextWith:          uuidPtrOrNil(ev.exchangeTextWith),
		PutByEntering:             ev.putByEntering,
		EntersWithCounters:        copyStringIntMap(ev.EntersWithCounters),
		EntryResumable:            ev.entryResumable,
		LandPlay:                  ev.landPlay,
		LandPlayer:                uuidPtrOrNil(ev.landPlayer),
		EntryControllerSet:        ev.entryControllerSet,
		EntryControllerPrior:      uuidPtrOrNil(ev.entryControllerPrior),
		AsCommanderMove:           ev.asCommanderMove,
		CommanderAnswer:           ev.commanderAnswer,

		CounterTarget:              uuidPtrOrNil(ev.CounterTarget),
		CounterName:                ev.CounterName,
		CounterDelta:               ev.CounterDelta,
		CounterPlayer:              uuidPtrOrNil(ev.CounterPlayer),
		StartingLoyalty:            ev.StartingLoyalty,
		CounterPlacer:              uuidPtrOrNil(ev.CounterPlacer),
		CounterFromCombatDamage:    ev.CounterFromCombatDamage,
		CounterFromCost:            ev.CounterFromCost,
		CounterFromTurnBasedAction: ev.CounterFromTurnBasedAction,
		LifePlayer:                 uuidPtrOrNil(ev.LifePlayer),
		LifeDelta:                  ev.LifeDelta,
		LifeFromRadiation:          ev.LifeFromRadiation,
		DiscardPlayer:              uuidPtrOrNil(ev.DiscardPlayer),
		DiscardCause:               ev.DiscardCause,
		DiscardCauseController:     uuidPtrOrNil(ev.DiscardCauseController),

		TokenController:         uuidPtrOrNil(ev.TokenController),
		TokenAttacking:          uuidPtrOrNil(ev.TokenAttacking),
		TokenTogether:           ev.TokenTogether,
		TokenAttackChoicesLabel: ev.TokenAttackChoicesLabel,
	}
	out.TokenGroups = snapshotTokenGroups(ev.TokenGroups, cen)
	if t := ev.tokenTail; t != nil {
		out.TokenTail = &tokenTailSnapshot{Then: snapshotKeyedTail(t.keyed)}
	}
	if t := ev.counterTail; t != nil {
		out.CounterTail = &counterTailSnapshot{Then: snapshotKeyedTail(t.keyed)}
	}
	if t := ev.lifeTail; t != nil {
		out.LifeTail = &lifeTailSnapshot{Then: snapshotKeyedTail(t.keyed), Each: cloneLifeEachRun(t.each), Move: snapshotMoveLifeRest(t.move)}
	}
	out.MillPlayer = uuidPtrOrNil(ev.MillPlayer)
	out.MillCount = ev.MillCount
	out.MillTail = snapshotMillTail(ev.mill)
	if t := ev.keywordAction; t != nil {
		out.KeywordActionTail = &keywordActionTailSnapshot{
			Cards: copyUUIDs(t.cards), Players: copyUUIDs(t.players), Ask: t.ask, Choice: t.choice,
			Land: uuidPtrOrNil(t.land), Then: snapshotKeyedTail(t.keyed),
			ArmySubtype: t.armySubtype,
		}
		if !reflect.ValueOf(t.armyToken).IsZero() {
			army := snapshotCard(t.armyToken, cen)
			out.KeywordActionTail.ArmyToken = &army
		}
	}
	for _, p := range ev.drawPending {
		out.DrawPending = append(out.DrawPending, pendingDrawSnapshot{
			Player: p.player, Count: p.count, Applied: copyReplacementEffectIDs(p.applied),
		})
	}
	if ev.drawThen.Ref.key != "" {
		out.DrawThen = &drawThenSnapshot{
			Key: ev.drawThen.Ref.key, Player: uuidPtrOrNil(ev.drawThen.Player),
			Source: uuidPtrOrNil(ev.drawThen.Source), N: ev.drawThen.N,
		}
	}
	if l := ev.FaceDownListed; l != nil {
		out.FaceDownListed = &FaceDownListing{
			Types: copyStrings(l.Types), Subtypes: copyStrings(l.Subtypes), Power: l.Power, Toughness: l.Toughness,
		}
	}
	if ev.auraHost != (TargetRef{}) {
		h := ev.auraHost
		out.AuraHost = &h
	}
	if ev.stackItem != nil {
		it := snapshotStackItem(g, ev.stackItem, cen)
		out.StackItem = &it
	}
	if t := ev.entryTail; t != nil {
		out.EntryTail = &entryTailSnapshot{
			NewObject: t.newObject, CraftedWith: append([]ObjectRef(nil), t.craftedWith...), BatchIndex: t.batchIndex,
			Search: cloneSearchEntryRun(t.search), Tokens: cloneTokenEntryRun(t.tokens),
			Keyed: snapshotKeyedTail(t.keyed),
		}
		if t.batch != nil {
			out.EntryTail.Batch = g.snapshotEntryBatch(t.batch, cen)
		}
	}
	if r := ev.zoneRoute; r != nil {
		out.ZoneRoute = &zoneRouteSnapshot{
			CardID: uuidPtrOrNil(r.CardID), Dst: r.Dst, DstOwner: uuidPtrOrNil(r.DstOwner),
			Actor: uuidPtrOrNil(r.Actor), ToBottom: r.ToBottom, Depth: r.Depth, FaceDown: r.FaceDown,
			HiddenBy: permissionRefPtr(r.HiddenBy), ExiledWith: permissionRefPtr(r.ExiledWith),
			Mill: r.Mill, Discard: r.Discard, DiscardCause: r.DiscardCause,
			Source: uuidPtrOrNil(r.Source), MustSettleNow: r.MustSettleNow, Countered: r.Countered,
			CauseKind: r.Cause.Kind, CauseController: uuidPtrOrNil(r.Cause.Controller), CauseItem: uuidPtrOrNil(r.Cause.Item),
			Sacrifice: r.Sacrifice, ViaBattlefieldLeave: r.ViaBattlefieldLeave, Destruction: r.Destruction,
			CantBeRegenerated: r.CantBeRegenerated, StateBased: r.StateBased, AsCommander: r.AsCommander,
			CommanderAnswer: r.commanderAnswer, FaceDownLegs: cloneFaceDownExileRun(r.faceDownLegs),
			BatchLegs: cloneRouteBatchRun(r.batchLegs), DiscardLegs: cloneDiscardBatchRun(r.discardLegs),
			Exchange: snapshotExchangeEntry(r.simultaneousEntry),
		}
		for _, c := range r.simultaneousExit {
			out.ZoneRoute.SimultaneousExit = append(out.ZoneRoute.SimultaneousExit, snapshotCard(c, cen))
		}
		if r.ExiledWithAbility != (AbilityRef{}) {
			a := r.ExiledWithAbility
			out.ZoneRoute.ExiledWithAbility = &a
		}
	}
	for _, l := range ev.lifeAfterMove {
		out.LifeAfterMove = append(out.LifeAfterMove, moveLifeSnapshot{Source: uuidPtrOrNil(l.Source), Player: l.Player, Delta: l.Delta})
	}
	for _, x := range ev.extraTurnsAfterMove {
		out.ExtraTurnsAfterMove = append(out.ExtraTurnsAfterMove, moveTurnSnapshot{Source: uuidPtrOrNil(x.Source), Player: x.Player})
	}
	for _, r := range ev.reflexiveAfterMove {
		out.ReflexiveAfterMove = append(out.ReflexiveAfterMove, snapshotMoveReflexive(r))
	}
	snapshotDamageEvent(ev, &out)
	return out
}

// restoreReplacementEvent is snapshotReplacementEvent's inverse. The
// look-ahead cache is left empty, for the running binary to fill.
func restoreReplacementEvent(s *replacementEventSnapshot) *ReplacementEvent {
	ev := &ReplacementEvent{
		ID:                        s.ID,
		Kind:                      s.Kind,
		Canceled:                  s.Canceled,
		Actor:                     uuidOrNil(s.Actor),
		Source:                    uuidOrNil(s.Source),
		mustSettleNow:             s.MustSettleNow,
		DrawPlayer:                uuidOrNil(s.DrawPlayer),
		DrawCount:                 s.DrawCount,
		drawTail:                  s.DrawTail,
		KeywordAction:             s.KeywordAction,
		KeywordActionCount:        s.KeywordActionCount,
		CardID:                    uuidOrNil(s.CardID),
		OldZone:                   s.OldZone,
		NewZone:                   s.NewZone,
		NewZoneOwner:              uuidOrNil(s.NewZoneOwner),
		ShuffleDestinationLibrary: s.ShuffleDestinationLibrary,
		ExiledWith:                permissionRefValue(s.ExiledWith),
		ExileCounter:              s.ExileCounter,
		Destruction:               s.Destruction,
		CantBeRegenerated:         s.CantBeRegenerated,
		StateBasedDestruction:     s.StateBasedDestruction,
		LeaveCause:                s.LeaveCause,
		EntersTapped:              s.EntersTapped,
		EntersPrepared:            s.EntersPrepared,
		EntersDevoured:            s.EntersDevoured,
		EntersWithHaste:           s.EntersWithHaste,
		entryRefused:              s.EntryRefused,
		EntersUnlocked:            s.EntersUnlocked,
		EntersAttacking:           uuidOrNil(s.EntersAttacking),
		FaceDown:                  s.FaceDown,
		EntersAsCopyOf:            copyPrintedValues(s.EntersAsCopyOf),
		copySourceID:              uuidOrNil(s.CopySource),
		copyUntilEndOfTurn:        s.CopyUntilEndOfTurn,
		auraHostFixed:             s.AuraHostFixed,
		auraHostSettled:           s.AuraHostSettled,
		entersAsAura:              s.EntersAsAura,
		entersAs:                  copyEntersAs(s.EntersAs),
		exchangeTextWith:          uuidOrNil(s.ExchangeTextWith),
		putByEntering:             s.PutByEntering,
		EntersWithCounters:        copyStringIntMap(s.EntersWithCounters),
		entryResumable:            s.EntryResumable,
		landPlay:                  s.LandPlay,
		landPlayer:                uuidOrNil(s.LandPlayer),
		entryControllerSet:        s.EntryControllerSet,
		entryControllerPrior:      uuidOrNil(s.EntryControllerPrior),
		asCommanderMove:           s.AsCommanderMove,
		commanderAnswer:           s.CommanderAnswer,

		CounterTarget:              uuidOrNil(s.CounterTarget),
		CounterName:                s.CounterName,
		CounterDelta:               s.CounterDelta,
		CounterPlayer:              uuidOrNil(s.CounterPlayer),
		StartingLoyalty:            s.StartingLoyalty,
		CounterPlacer:              uuidOrNil(s.CounterPlacer),
		CounterFromCombatDamage:    s.CounterFromCombatDamage,
		CounterFromCost:            s.CounterFromCost,
		CounterFromTurnBasedAction: s.CounterFromTurnBasedAction,
		LifePlayer:                 uuidOrNil(s.LifePlayer),
		LifeDelta:                  s.LifeDelta,
		LifeFromRadiation:          s.LifeFromRadiation,
		DiscardPlayer:              uuidOrNil(s.DiscardPlayer),
		DiscardCause:               s.DiscardCause,
		DiscardCauseController:     uuidOrNil(s.DiscardCauseController),

		TokenController:         uuidOrNil(s.TokenController),
		TokenAttacking:          uuidOrNil(s.TokenAttacking),
		TokenTogether:           s.TokenTogether,
		TokenAttackChoicesLabel: s.TokenAttackChoicesLabel,
	}
	ev.TokenGroups = restoreTokenGroups(s.TokenGroups)
	if t := s.TokenTail; t != nil {
		ev.tokenTail = &tokenTail{keyed: restoreKeyedTail(t.Then)}
	}
	if t := s.CounterTail; t != nil {
		ev.counterTail = &counterTail{keyed: restoreKeyedTail(t.Then)}
	}
	if t := s.LifeTail; t != nil {
		ev.lifeTail = &lifeTail{keyed: restoreKeyedTail(t.Then), each: cloneLifeEachRun(t.Each), move: restoreMoveLifeRest(t.Move)}
	}
	ev.MillPlayer = uuidOrNil(s.MillPlayer)
	ev.MillCount = s.MillCount
	ev.mill = restoreMillTail(s.MillTail)
	if t := s.KeywordActionTail; t != nil {
		ev.keywordAction = &keywordActionTail{
			cards: copyUUIDs(t.Cards), players: copyUUIDs(t.Players), ask: t.Ask, choice: t.Choice,
			land: uuidOrNil(t.Land), keyed: restoreKeyedTail(t.Then), armySubtype: t.ArmySubtype,
		}
		if t.ArmyToken != nil {
			ev.keywordAction.armyToken = restoreCard(t.ArmyToken)
		}
	}
	for _, p := range s.DrawPending {
		ev.drawPending = append(ev.drawPending, pendingDraw{
			player: p.Player, count: p.Count, applied: copyReplacementEffectIDs(p.Applied),
		})
	}
	if r := s.DrawInstead; r != nil {
		ev.drawInstead = drawInsteadRun{ref: DrawInsteadRef{key: r.Key}, source: uuidOrNil(r.Source)}
	}
	if r := s.KeywordInstead; r != nil {
		ev.keywordInstead = keywordInsteadRun{ref: KeywordInsteadRef{key: r.Key}, source: uuidOrNil(r.Source)}
	}
	if d := s.DrawThen; d != nil {
		ev.drawThen = DrawThen{Ref: DrawThenRef{key: d.Key}, Player: uuidOrNil(d.Player), Source: uuidOrNil(d.Source), N: d.N}
	}
	if l := s.FaceDownListed; l != nil {
		ev.FaceDownListed = &FaceDownListing{
			Types: copyStrings(l.Types), Subtypes: copyStrings(l.Subtypes), Power: l.Power, Toughness: l.Toughness,
		}
	}
	if s.AuraHost != nil {
		ev.auraHost = *s.AuraHost
	}
	if s.StackItem != nil {
		ev.stackItem, _ = restoreStackItem(s.StackItem)
	}
	if t := s.EntryTail; t != nil {
		ev.entryTail = &entryTail{
			newObject: t.NewObject, craftedWith: append([]ObjectRef(nil), t.CraftedWith...), batchIndex: t.BatchIndex,
			search: cloneSearchEntryRun(t.Search), batch: restoreEntryBatch(t.Batch), tokens: cloneTokenEntryRun(t.Tokens),
			keyed: restoreKeyedTail(t.Keyed),
		}
	}
	if r := s.ZoneRoute; r != nil {
		ev.zoneRoute = &zoneRoute{
			CardID: uuidOrNil(r.CardID), Dst: r.Dst, DstOwner: uuidOrNil(r.DstOwner),
			Actor: uuidOrNil(r.Actor), ToBottom: r.ToBottom, Depth: r.Depth, FaceDown: r.FaceDown,
			HiddenBy: permissionRefValue(r.HiddenBy), ExiledWith: permissionRefValue(r.ExiledWith),
			Mill: r.Mill, Discard: r.Discard, DiscardCause: r.DiscardCause,
			Source: uuidOrNil(r.Source), MustSettleNow: r.MustSettleNow, Countered: r.Countered,
			Cause:     MoveCause{Kind: r.CauseKind, Controller: uuidOrNil(r.CauseController), Item: uuidOrNil(r.CauseItem)},
			Sacrifice: r.Sacrifice, ViaBattlefieldLeave: r.ViaBattlefieldLeave, Destruction: r.Destruction,
			CantBeRegenerated: r.CantBeRegenerated, StateBased: r.StateBased, AsCommander: r.AsCommander,
			commanderAnswer: r.CommanderAnswer, faceDownLegs: cloneFaceDownExileRun(r.FaceDownLegs),
			batchLegs: cloneRouteBatchRun(r.BatchLegs), discardLegs: cloneDiscardBatchRun(r.DiscardLegs),
			simultaneousEntry: restoreExchangeEntry(r.Exchange),
		}
		for i := range r.SimultaneousExit {
			ev.zoneRoute.simultaneousExit = append(ev.zoneRoute.simultaneousExit, restoreCard(&r.SimultaneousExit[i]))
		}
		if r.ExiledWithAbility != nil {
			ev.zoneRoute.ExiledWithAbility = *r.ExiledWithAbility
		}
	}
	for _, l := range s.LifeAfterMove {
		ev.lifeAfterMove = append(ev.lifeAfterMove, lifeInstruction{Source: uuidOrNil(l.Source), Player: l.Player, Delta: l.Delta})
	}
	for _, x := range s.ExtraTurnsAfterMove {
		ev.extraTurnsAfterMove = append(ev.extraTurnsAfterMove, extraTurnInstruction{Source: uuidOrNil(x.Source), Player: x.Player})
	}
	for _, r := range s.ReflexiveAfterMove {
		ev.reflexiveAfterMove = append(ev.reflexiveAfterMove, restoreMoveReflexive(r))
	}
	restoreDamageEvent(s, ev)
	return ev
}

// unknownReplacementResumeKeys is checkEffectKeys' half for a keyed
// frame: its tail, the bodies its event names, and its stack item.
func unknownReplacementResumeKeys(c *pendingChoiceSnapshot) []string {
	var unknown []string
	f := c.ReplacementResume
	kind, known := replacementResumeTails[c.PickThen]
	switch {
	case !known:
		unknown = append(unknown, "replacement resume "+c.PickThen)
	case f == nil:
		unknown = append(unknown, "replacement resume "+c.PickThen+" with no frame")
	case f.Event.Kind != kind:
		unknown = append(unknown, "replacement resume "+c.PickThen+" over a "+string(f.Event.Kind)+" event")
	case kind == RepEventDamage && c.PickThen != snapshotDamageResumeKey(&f.Event):
		// #4362: combat damage is written under its own key, and only
		// combat damage is. #4404: likewise an event with redirected
		// parts.
		unknown = append(unknown, "replacement resume "+c.PickThen+" over a damage event of the other kind")
	case kind == RepEventDamage && snapshotDamageIsCombat(&f.Event) && snapshotCarriesDataTail(&f.Event):
		// #4362: combat damage owes its caller nothing.
		unknown = append(unknown, "replacement resume "+c.PickThen+" over an event with a data tail")
	case kind == RepEventMove && (len(f.Event.ReflexiveAfterMove) > 0) != (c.PickThen == ReplacementResumeMoveReflexive):
		// #4517: a declared reflexive trigger is written under its own
		// key, and only an event carrying one is.
		unknown = append(unknown, "replacement resume "+c.PickThen+" over a move whose reflexive triggers it does not match")
	case kind != RepEventDamage && snapshotCarriesDataTail(&f.Event) && c.PickThen != snapshotTailedKeyOf(&f.Event, kind):
		// #4253 slice 3: a tail written as data under the key a binary
		// that drops it would accept.
		unknown = append(unknown, "replacement resume "+c.PickThen+" over an event with a data tail")
	}
	if f == nil {
		return unknown
	}
	return append(unknown, unknownReplacementEventKeys(&f.Event)...)
}

// unknownReplacementEventKeys is the event's half: the bodies, tails and
// stack item it names.
func unknownReplacementEventKeys(ev *replacementEventSnapshot) []string {
	var unknown []string
	if r := ev.DrawInstead; r != nil && !KnownDrawInstead(r.Key) {
		unknown = append(unknown, "draw-instead body "+r.Key)
	}
	if r := ev.KeywordInstead; r != nil && !KnownKeywordInstead(r.Key) {
		unknown = append(unknown, "keyword-instead body "+r.Key)
	}
	if t := ev.EntryTail; t != nil {
		unknown = append(unknown, unknownSearchEntryKeys(t.Search)...)
		unknown = append(unknown, unknownEntryBatchKeys(t.Batch)...)
		unknown = append(unknown, unknownTokenEntryKeys(t.Tokens)...)
		unknown = append(unknown, unknownKeyedTail(t.Keyed)...)
	}
	if t := ev.TokenTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
	}
	if r := ev.ZoneRoute; r != nil {
		unknown = append(unknown, unknownFaceDownExileKeys(r.FaceDownLegs)...)
		unknown = append(unknown, unknownRouteBatchKeys(r.BatchLegs)...)
		unknown = append(unknown, unknownDiscardBatchKeys(r.DiscardLegs)...)
	}
	if t := ev.CounterTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
	}
	if t := ev.LifeTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
		if t.Each != nil {
			unknown = append(unknown, unknownKeyedTail(t.Each.Then)...)
		}
	}
	if t := ev.MillTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
		unknown = append(unknown, unknownMillUntilKeys(t.Run)...)
	}
	if t := ev.DamageTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
		unknown = append(unknown, unknownDamageEachKeys(t.Each)...)
	}
	for _, r := range ev.ReflexiveAfterMove {
		if !KnownEffectBody(r.Body) {
			unknown = append(unknown, "reflexive body "+r.Body)
		}
	}
	if d := ev.DrawThen; d != nil && !KnownDrawThen(d.Key) {
		unknown = append(unknown, "draw-then body "+d.Key)
	}
	if t := ev.KeywordActionTail; t != nil && t.Choice != "" && !KnownChoiceKind(t.Choice) {
		// The prompt a settled scry or surveil queues.
		unknown = append(unknown, "keyword-action prompt kind "+string(t.Choice))
	}
	if t := ev.KeywordActionTail; t != nil {
		unknown = append(unknown, unknownKeyedTail(t.Then)...)
	}
	if it := ev.StackItem; it != nil {
		if it.Body != "" && !KnownEffectBody(it.Body) {
			unknown = append(unknown, "stack-item body "+it.Body)
		}
		unknown = append(unknown, unknownAbilityRef(*it)...)
	}
	return unknown
}

// restoreReplacementSerials puts the object serials back.
func (g *Game) restoreReplacementSerials(serials []replacementSerialSnapshot, next int) {
	if len(serials) == 0 && next == 0 {
		return
	}
	g.replacementObjectSerials = make(map[ObjectRef]int, len(serials))
	for _, s := range serials {
		g.replacementObjectSerials[s.Object] = s.Serial
		if s.Serial >= next {
			next = s.Serial + 1
		}
	}
	g.nextReplacementObjectSerial = next
	g.replacementSerialsCompactAt = 2*len(g.replacementObjectSerials) + 64
}

// snapshotReplacementSerials writes the object serials, sorted.
func (g *Game) snapshotReplacementSerials() ([]replacementSerialSnapshot, int) {
	out := make([]replacementSerialSnapshot, 0, len(g.replacementObjectSerials))
	for ref, serial := range g.replacementObjectSerials {
		out = append(out, replacementSerialSnapshot{Object: ref, Serial: serial})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Serial < out[j].Serial })
	return out, g.nextReplacementObjectSerial
}

// rebuildReplacementResumeLocked rebuilds a keyed frame on a restored
// game: the event, its marks, and the pending set gathered again by ID.
// An ID the gather no longer finds — the effect is gone from the running
// binary's catalog — is kept as a lost entry, which every resolve path
// treats as an effect that no longer applies (CR 616.1f).
//
// Caller must hold g.mu (write): the gather fills layer caches and mints
// serials.
func (g *Game) rebuildReplacementResumeLocked(s *replacementResumeSnapshot) *replacementResumeFrame {
	ev := restoreReplacementEvent(&s.Event)
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	marks := make(map[ReplacementEffectID]bool, len(s.Applied))
	for _, id := range s.Applied {
		marks[id] = true
	}
	g.replacementsAppliedThisEvent[ev.ID] = marks
	if uint64(ev.ID) > g.nextReplacementEventID.Load() {
		g.nextReplacementEventID.Store(uint64(ev.ID))
	}
	// #4253: a damage event's instance is one a later instance must not
	// reuse.
	g.damageInstanceSeq = max(g.damageInstanceSeq, uint64(ev.DamageInstance))
	gathered := g.gatherActiveReplacementsLocked(ev)
	byID := make(map[ReplacementEffectID]activeReplacement, len(gathered))
	for _, a := range gathered {
		byID[a.id] = a
	}
	frame := &replacementResumeFrame{ev: ev}
	for _, id := range s.Applicable {
		a, ok := byID[id]
		if !ok {
			a = activeReplacement{id: id, lost: true}
		}
		frame.applicable = append(frame.applicable, a)
	}
	return frame
}
