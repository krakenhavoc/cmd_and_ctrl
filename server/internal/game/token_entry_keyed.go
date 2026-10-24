package game

import "github.com/google/uuid"

// token_entry_keyed.go — #4253 slice 3: a token creation's remaining
// entries, carried as data.
//
// A settled creation mints its tokens and enters them one at a time,
// each from the previous token's entry continuation
// (enterCreatedTokensLocked, #762), so a token whose entry stops on a
// question (two enters-tapped effects on an opponent's token) waits with
// the rest behind it. That continuation was a closure over the minted
// tokens, and the token entering was only in Game.enteringTokens, which
// a restore point did not carry.
//
// The keyed form stages EVERY minted token of the batch in
// Game.enteringTokens up front, so the objects are in a field a restore
// point carries (GameSnapshot.EnteringTokens), and the rest of the batch
// is IDs and each token's entry clause (tokenEntryRun). A token is only
// ever looked up by its ID, and nothing outside the creation knows the
// ID of one still waiting, so staging it early is invisible. The
// creation's own continuation rides along as a keyed tail or a
// synchronous door's result; a creation whose card passes a closure
// keeps the closure form, and its frame is still counted.

// tokenDone is a creation's continuation, in whichever form it came: a
// closure, a keyed tail, or a synchronous door's result. A parameter
// only; the keyed form stores its two data halves (tokenEntryRun).
type tokenDone struct {
	then   func(g *Game, created []uuid.UUID) error
	keyed  KeyedTail
	report *[]uuid.UUID
}

// run tells the continuation which tokens landed.
func (d tokenDone) run(g *Game, created []uuid.UUID) error {
	if d.report != nil {
		*d.report = created
	}
	if d.then != nil {
		return d.then(g, created)
	}
	return g.runKeyedTailLocked(d.keyed, created, 0)
}

// closure is the continuation as one func, for the simultaneous-entry
// path, which is not keyed. Nil when there is nothing to run.
func (d tokenDone) closure() func(*Game, []uuid.UUID) error {
	if d.then == nil && d.keyed.IsZero() && d.report == nil {
		return nil
	}
	return d.run
}

// tokenEntryStep is one staged token still to enter: its ID (the object
// is in Game.enteringTokens), the creation's entry clause for it, and
// the attacker it is put onto the battlefield blocking.
type tokenEntryStep struct {
	ID       uuid.UUID          `json:"id"`
	Entry    tokenEntrySnapshot `json:"entry"`
	Blocking uuid.UUID          `json:"blocking"`
}

// tokenEntrySnapshot is a TokenEntryOptions as data.
type tokenEntrySnapshot struct {
	Tapped     bool           `json:"tapped,omitempty"`
	Counters   map[string]int `json:"counters,omitempty"`
	Keywords   []string       `json:"keywords,omitempty"`
	AttachedTo *TargetRef     `json:"attachedTo,omitempty"`
}

func snapshotTokenEntry(o TokenEntryOptions) tokenEntrySnapshot {
	out := tokenEntrySnapshot{Tapped: o.Tapped, Counters: copyCounterMap(o.Counters), Keywords: copyStrings(o.Keywords)}
	if o.AttachedTo != (TargetRef{}) {
		h := o.AttachedTo
		out.AttachedTo = &h
	}
	return out
}

func (s tokenEntrySnapshot) options() TokenEntryOptions {
	out := TokenEntryOptions{Tapped: s.Tapped, Counters: copyCounterMap(s.Counters), Keywords: copyStrings(s.Keywords)}
	if s.AttachedTo != nil {
		out.AttachedTo = *s.AttachedTo
	}
	return out
}

// tokenEntryRun is the rest of a token batch, riding the entry tail of
// the token now entering (Token, which is put blocking Blocking once it
// lands): the tokens still to enter, the tokens that have, the
// creation's source and its continuation. Never written after it is
// built, so an undo snapshot may share it.
type tokenEntryRun struct {
	Token    uuid.UUID          `json:"token"`
	Blocking uuid.UUID          `json:"blocking"`
	Rest     []tokenEntryStep   `json:"rest,omitempty"`
	Created  []uuid.UUID        `json:"created,omitempty"`
	Source   uuid.UUID          `json:"source"`
	Then     *keyedTailSnapshot `json:"then,omitempty"`
	// report is a synchronous door's result (tokenTail.report). Never
	// written: a creation that paused has returned to the door already.
	report *[]uuid.UUID
}

func cloneTokenEntryRun(r *tokenEntryRun) *tokenEntryRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = make([]tokenEntryStep, 0, len(r.Rest))
	for _, st := range r.Rest {
		st.Entry = snapshotTokenEntry(st.Entry.options()) // its own map and slice
		out.Rest = append(out.Rest, st)
	}
	out.Created = copyUUIDs(r.Created)
	if r.Then != nil {
		t := *r.Then
		t.Carry = copyUUIDs(r.Then.Carry)
		out.Then = &t
	}
	return &out
}

// enterCreatedTokensKeyedLocked is enterCreatedTokensLocked for a batch
// whose continuation is data: every token is staged now, and the batch
// walks on IDs.
//
// Caller must hold g.mu.
func (g *Game) enterCreatedTokensKeyedLocked(batch []stagedToken, created []uuid.UUID, source uuid.UUID, done tokenDone) error {
	steps := make([]tokenEntryStep, 0, len(batch))
	for _, st := range batch {
		g.enteringTokens = append(g.enteringTokens, st.card)
		steps = append(steps, tokenEntryStep{ID: st.card.InstanceID, Entry: snapshotTokenEntry(st.entry), Blocking: st.blocking})
	}
	return g.enterStagedTokensLocked(&tokenEntryRun{
		Rest: steps, Created: copyUUIDs(created), Source: source,
		Then: snapshotKeyedTail(done.keyed), report: done.report,
	})
}

// enterStagedTokensLocked enters the head of r.Rest, carrying the rest
// on its entry tail, or runs the creation's continuation when none is
// left. A step whose token is no longer staged is skipped.
//
// Caller must hold g.mu.
func (g *Game) enterStagedTokensLocked(r *tokenEntryRun) error {
	rest := r.Rest
	for len(rest) > 0 {
		if _, ok := g.enteringTokenLocked(rest[0].ID); ok {
			break
		}
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return tokenDone{keyed: restoreKeyedTail(r.Then), report: r.report}.run(g, r.Created)
	}
	next := rest[0]
	card, _ := g.enteringTokenLocked(next.ID)
	ev := createdTokenEntryEvent(card, next.Entry.options(), r.Source)
	ev.entryTail = &entryTail{tokens: &tokenEntryRun{
		Token: next.ID, Blocking: next.Blocking, Rest: rest[1:], Created: r.Created, Source: r.Source,
		Then: r.Then, report: r.report,
	}}
	_, err := g.enterBattlefieldThroughPipelineLocked(ev)
	if err != nil {
		// The batch stops here, as the closure form's does. A token
		// still staged will never enter, so it never existed (CR 111.1).
		for _, st := range rest {
			g.dropEnteringTokenLocked(st.ID)
		}
	}
	return err
}

// runTokenEntryStepLocked is the keyed form of the step closure in
// enterCreatedTokensLocked: the token that just settled leaves the
// staging area, is counted when it landed, and the batch walks on.
//
// Caller must hold g.mu.
func (g *Game) runTokenEntryStepLocked(r *tokenEntryRun, entered uuid.UUID) error {
	g.dropEnteringTokenLocked(r.Token)
	landed := r.Created
	if entered != uuid.Nil {
		landed = append(append([]uuid.UUID(nil), r.Created...), entered)
		if r.Blocking != uuid.Nil {
			g.stampEntryBlockerLocked(entered, r.Blocking)
		}
	}
	return g.enterStagedTokensLocked(&tokenEntryRun{
		Rest: r.Rest, Created: landed, Source: r.Source, Then: r.Then, report: r.report,
	})
}

// createdTokenEntryEvent is a staged token's battlefield entry event,
// for both forms of the batch.
func createdTokenEntryEvent(card Card, entry TokenEntryOptions, source uuid.UUID) *ReplacementEvent {
	ev := &ReplacementEvent{
		Kind:   RepEventMove,
		Actor:  card.Controller,
		Source: source,
		CardID: card.InstanceID,
		// No OLD zone: a token is created on the battlefield and came
		// from nowhere (CR 111.1). Every entry replacement in the
		// catalog keys on NewZone, so none of them notices, and the
		// empty old zone is what the finisher reads to push the staged
		// token instead of moving a card.
		NewZone:      ZoneBattlefield,
		NewZoneOwner: card.Controller,
		EntersTapped: entry.Tapped,
		// #1227: CR 506.3c's attack, taken off the minted token and
		// put on the ENTRY where the card door already carries it
		// (ZoneEntryOptions.Attacking). The token object keeps its own
		// AttackingTarget — that is what a replacement inspecting the
		// staged token reads, and what this line copies — but the
		// "never declared" marking is now made once, by the entry, for
		// both doors. See stampEntryAttackerLocked.
		EntersAttacking: card.AttackingTarget,
		// The entry may pause, and #478's frame finishes it. Nothing is
		// skipped by resuming one: there is no shuffle owed and no new
		// object identity to mint.
		entryResumable: true,
	}
	for name, n := range entry.Counters {
		ev.AddCounterAtETB(name, n)
	}
	// #3158: a creation that names the Aura token's host.
	if entry.AttachedTo.Kind != "" {
		ev.auraHost, ev.auraHostFixed = entry.AttachedTo, true
	}
	return ev
}

// unknownTokenEntryKeys is checkEffectKeys' half for a keyed token batch.
func unknownTokenEntryKeys(r *tokenEntryRun) []string {
	if r == nil {
		return nil
	}
	return unknownKeyedTail(r.Then)
}

// stagedTokensAreData reports whether every staged token is written
// whole by the card mirror: one holding an ability the catalog cannot
// rebuild keeps a frame that needs it off the disk.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) stagedTokensAreData() bool {
	for _, c := range g.enteringTokens {
		var scratch ContinuationCensus
		_ = snapshotCard(c, &scratch)
		if !scratch.Empty() {
			return false
		}
	}
	return true
}
