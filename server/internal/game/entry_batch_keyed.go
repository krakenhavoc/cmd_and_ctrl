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
// Not keyed: a batch whose continuation is a closure, one that has
// recorded an error, a batch of created tokens (their objects live only
// in Game.enteringTokens), and one whose settled member events would not
// be keyed on their own.

// entryBatchKeyable reports whether a paused batch can be written as data.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) entryBatchKeyable(b *entryBatch) bool {
	if b == nil || b.then != nil || b.firstErr != nil || b.done {
		return false
	}
	for _, m := range b.members {
		if m.token {
			return false
		}
		if m.out != nil && !g.replacementEventKeyable(m.out) {
			return false
		}
	}
	return true
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
	}
	for _, m := range b.members {
		ms := entryBatchMemberSnapshot{
			CardID: m.cardID, From: m.from, Controller: m.controller, PriorController: uuidPtrOrNil(m.priorController),
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
		opts: restoreZoneEntryOptions(s.Opts),
	}
	for _, m := range s.Members {
		bm := entryBatchMember{
			cardID: m.CardID, from: m.From, controller: m.Controller, priorController: uuidOrNil(m.PriorController),
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
	for _, m := range s.Members {
		if m.Out != nil {
			unknown = append(unknown, unknownReplacementEventKeys(m.Out)...)
		}
	}
	return unknown
}
