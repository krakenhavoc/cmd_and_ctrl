package game

import "github.com/google/uuid"

// entry_batch_keyed.go — #4253 slice 3: a simultaneous entry (entry_batch.go)
// paused on one card's question, written as data.
//
// A paused batch member's event carries the whole batch on its entry
// tail: the members, the cursor, the settled events of the members whose
// windows have run, the putting effect's options and the caller's
// continuation. All of it is data except that continuation, so a batch
// whose continuation is nothing, a synchronous door's result
// (entryBatch.done / landed) or a registered tail (entryBatch.keyed) is
// written with the frame, and the question is a restore point: a
// shockland put by Genesis Wave, a reveal land put by Arboreal Grazer.
//
// Since #4404 two more batches are data. Created tokens entering together
// (myriad, Saheeli the Gifted, Gisa the Hellraiser: TokenCreation.Together)
// are members whose objects live only in Game.enteringTokens, which a keyed
// frame writes (GameSnapshot.EnteringTokens); each member's token flag, its
// entry clause and what it attacks are written with it. The creation's
// synchronous door result (entryBatch.report) is not written, as
// tokenEntryRun.report is not: a creation that paused has returned to its
// door already. An Attraction's open (attractionOpenRun) is the rest of an
// "open N Attractions", and the card entering is in no zone the index
// reaches but through Game.attractionEntering, which is written beside it
// (GameSnapshot.AttractionEntering). Each takes a key of its own
// (ReplacementResumeMoveTokenBatch, ReplacementResumeMoveAttractionOpen): a
// binary from before them would read a token member as a card in no zone,
// or finish an open as an ordinary batch and lose the opens still to come.
//
// Not keyed: a batch that has recorded an error, one that is done, and one
// whose settled member events would not be keyed on their own.

// ReplacementResumeMoveTokenBatch and ReplacementResumeMoveAttractionOpen
// are the frame keys of a paused simultaneous entry of created tokens and
// of an Attraction's open (#4404). On-disk identities: never renamed, never
// reused.
const (
	ReplacementResumeMoveTokenBatch     = "replacement-resume/move-token-batch"
	ReplacementResumeMoveAttractionOpen = "replacement-resume/move-attraction-open"
)

// entryBatchKeyable reports whether a paused batch can be written as data.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) entryBatchKeyable(b *entryBatch) bool {
	if b == nil || b.firstErr.set || b.done {
		return false
	}
	for _, m := range b.members {
		if m.token && !g.stagedTokensAreData() {
			return false
		}
		if m.out != nil && !g.replacementEventKeyableAs(m.out, m.token) {
			return false
		}
	}
	return true
}

// entryBatchHasToken reports whether a batch enters created tokens.
func entryBatchHasToken(b *entryBatch) bool {
	for _, m := range b.members {
		if m.token {
			return true
		}
	}
	return false
}

// entryBatchOwnKey is the key a batch of its own kind takes (#4404), or ""
// for an ordinary batch.
func entryBatchOwnKey(b *entryBatch) string {
	switch {
	case b == nil:
	case b.attraction != nil:
		return ReplacementResumeMoveAttractionOpen
	case entryBatchHasToken(b):
		return ReplacementResumeMoveTokenBatch
	}
	return ""
}

// snapshotEntryBatchOwnKey is entryBatchOwnKey over the mirror.
func snapshotEntryBatchOwnKey(b *entryBatchSnapshot) string {
	if b == nil {
		return ""
	}
	if b.Attraction != nil {
		return ReplacementResumeMoveAttractionOpen
	}
	for _, m := range b.Members {
		if m.Token {
			return ReplacementResumeMoveTokenBatch
		}
	}
	return ""
}

// entryBatchSnapshot is a paused batch on disk.
type entryBatchSnapshot struct {
	ExchangeSource    *uuid.UUID                 `json:"exchangeSource,omitempty"`
	ExchangeSacrifice *uuid.UUID                 `json:"exchangeSacrifice,omitempty"`
	Source            *uuid.UUID                 `json:"source,omitempty"`
	Members           []entryBatchMemberSnapshot `json:"members"`
	Next              int                        `json:"next,omitempty"`
	Then              *keyedTailSnapshot         `json:"then,omitempty"`
	Opts              zoneEntryOptionsSnapshot   `json:"opts"`
	// Discard is a keyed discard batch's leg that became this entry
	// (#4253 slice 9, entryBatch.discard).
	Discard *discardEntryRun `json:"discard,omitempty"`
	// Attraction is the rest of an "open N Attractions" whose open this
	// entry is (#4404, entryBatch.attraction). Additive within v9, under
	// ReplacementResumeMoveAttractionOpen.
	Attraction *attractionOpenRun `json:"attraction,omitempty"`
}

// entryBatchMemberSnapshot is one card of a paused batch. Out is its
// settled event, absent while its window has not run or when it does
// not enter.
type entryBatchMemberSnapshot struct {
	CardID          uuid.UUID                 `json:"cardId"`
	From            ZoneKind                  `json:"from"`
	Controller      uuid.UUID                 `json:"controller"`
	PriorController *uuid.UUID                `json:"priorController,omitempty"`
	Out             *replacementEventSnapshot `json:"out,omitempty"`
	// Token, TokenEntry and Attacking are a created token's membership
	// (#4404): the object is in GameSnapshot.EnteringTokens, its entry
	// clause and the player it enters attacking are here. Additive within
	// v9, under ReplacementResumeMoveTokenBatch.
	Token      bool                `json:"token,omitempty"`
	TokenEntry *tokenEntrySnapshot `json:"tokenEntry,omitempty"`
	Attacking  *uuid.UUID          `json:"attacking,omitempty"`
}

// zoneEntryOptionsSnapshot is a ZoneEntryOptions on disk.
type zoneEntryOptionsSnapshot struct {
	Controller               *uuid.UUID            `json:"controller,omitempty"`
	PutByEnteringPlayer      bool                  `json:"putByEnteringPlayer,omitempty"`
	Tapped                   bool                  `json:"tapped,omitempty"`
	Attacking                *uuid.UUID            `json:"attacking,omitempty"`
	FaceDown                 FaceDownKind          `json:"faceDown,omitempty"`
	FaceDownListed           *FaceDownListing      `json:"faceDownListed,omitempty"`
	EntersWithCounters       map[string]int        `json:"entersWithCounters,omitempty"`
	EntersWithCountersIfType []EntryCountersIfType `json:"entersWithCountersIfType,omitempty"`
	AttachedTo               *TargetRef            `json:"attachedTo,omitempty"`
	EntersAs                 *EntersAs             `json:"entersAs,omitempty"`
}

func copyFaceDownListing(l *FaceDownListing) *FaceDownListing {
	if l == nil {
		return nil
	}
	return &FaceDownListing{Types: copyStrings(l.Types), Subtypes: copyStrings(l.Subtypes), Power: l.Power, Toughness: l.Toughness}
}

func snapshotZoneEntryOptions(o ZoneEntryOptions) zoneEntryOptionsSnapshot {
	out := zoneEntryOptionsSnapshot{
		Controller: uuidPtrOrNil(o.Controller), PutByEnteringPlayer: o.PutByEnteringPlayer, Tapped: o.Tapped,
		Attacking: uuidPtrOrNil(o.Attacking), FaceDown: o.FaceDown, FaceDownListed: copyFaceDownListing(o.FaceDownListed),
		EntersWithCounters:       copyCounterMap(o.EntersWithCounters),
		EntersWithCountersIfType: append([]EntryCountersIfType(nil), o.EntersWithCountersIfType...),
		EntersAs:                 copyEntersAs(o.EntersAs),
	}
	if o.AttachedTo != (TargetRef{}) {
		h := o.AttachedTo
		out.AttachedTo = &h
	}
	return out
}

func restoreZoneEntryOptions(s zoneEntryOptionsSnapshot) ZoneEntryOptions {
	out := ZoneEntryOptions{
		Controller: uuidOrNil(s.Controller), PutByEnteringPlayer: s.PutByEnteringPlayer, Tapped: s.Tapped,
		Attacking: uuidOrNil(s.Attacking), FaceDown: s.FaceDown, FaceDownListed: copyFaceDownListing(s.FaceDownListed),
		EntersWithCounters:       copyCounterMap(s.EntersWithCounters),
		EntersWithCountersIfType: append([]EntryCountersIfType(nil), s.EntersWithCountersIfType...),
		EntersAs:                 copyEntersAs(s.EntersAs),
	}
	if s.AttachedTo != nil {
		out.AttachedTo = *s.AttachedTo
	}
	return out
}

// snapshotEntryBatch writes a keyable batch. Caller has checked
// entryBatchKeyable.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) snapshotEntryBatch(b *entryBatch, cen *ContinuationCensus) *entryBatchSnapshot {
	out := &entryBatchSnapshot{
		ExchangeSource: uuidPtrOrNil(b.exchangeSource), ExchangeSacrifice: uuidPtrOrNil(b.exchangeSacrifice),
		Source: uuidPtrOrNil(b.source), Next: b.next, Then: snapshotKeyedTail(b.keyed),
		Opts:    snapshotZoneEntryOptions(b.opts),
		Members: make([]entryBatchMemberSnapshot, 0, len(b.members)),
		Discard: cloneDiscardEntryRun(b.discard),
	}
	if b.attraction != nil {
		a := *b.attraction
		out.Attraction = &a
	}
	for _, m := range b.members {
		ms := entryBatchMemberSnapshot{
			CardID: m.cardID, From: m.from, Controller: m.controller, PriorController: uuidPtrOrNil(m.priorController),
		}
		if m.token {
			e := snapshotTokenEntry(m.tokenEntry)
			ms.Token, ms.TokenEntry, ms.Attacking = true, &e, uuidPtrOrNil(m.attacking)
		}
		if m.out != nil {
			ev := g.snapshotReplacementEvent(m.out, cen)
			ms.Out = &ev
		}
		out.Members = append(out.Members, ms)
	}
	return out
}

// restoreEntryBatch is snapshotEntryBatch's inverse.
func restoreEntryBatch(s *entryBatchSnapshot) *entryBatch {
	if s == nil {
		return nil
	}
	b := &entryBatch{
		exchangeSource: uuidOrNil(s.ExchangeSource), exchangeSacrifice: uuidOrNil(s.ExchangeSacrifice),
		source: uuidOrNil(s.Source), next: s.Next, keyed: restoreKeyedTail(s.Then),
		opts: restoreZoneEntryOptions(s.Opts), discard: cloneDiscardEntryRun(s.Discard),
	}
	if s.Attraction != nil {
		a := *s.Attraction
		b.attraction = &a
	}
	for _, m := range s.Members {
		bm := entryBatchMember{
			cardID: m.CardID, from: m.From, controller: m.Controller, priorController: uuidOrNil(m.PriorController),
			token: m.Token, attacking: uuidOrNil(m.Attacking),
		}
		if m.TokenEntry != nil {
			bm.tokenEntry = m.TokenEntry.options()
		}
		if m.Out != nil {
			bm.out = restoreReplacementEvent(m.Out)
		}
		b.members = append(b.members, bm)
	}
	return b
}

// unknownEntryBatchKeys is checkEffectKeys' half for a keyed batch: its
// continuation, and what its settled events name.
func unknownEntryBatchKeys(s *entryBatchSnapshot) []string {
	if s == nil {
		return nil
	}
	unknown := unknownKeyedTail(s.Then)
	if s.Discard != nil {
		unknown = append(unknown, unknownDiscardBatchKeys(s.Discard.Legs)...)
	}
	for _, m := range s.Members {
		if m.Out != nil {
			unknown = append(unknown, unknownReplacementEventKeys(m.Out)...)
		}
	}
	return unknown
}
