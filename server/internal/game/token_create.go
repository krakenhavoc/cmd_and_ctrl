package game

import (
	"errors"

	"github.com/google/uuid"
)

// token_create.go is the engine's ONE token-creation path (#762).
//
// CR 111.1: "a token is a marker used to represent any permanent
// that isn't represented by a card." CR 701.7b: "to create one or
// more tokens ... is to put the specified numbers and kinds of those
// tokens onto the battlefield." Three entry points used to say that
// in Go — a straight creation, a creation with entry options, and a
// creation attacking — each with its own copy of the mint-and-push
// loop, and none of them opening a replacement window. Nothing could
// double a token, nothing could swap what kind it was, and a created
// token skipped the battlefield-entry replacements every other
// permanent gets: no enters-tapped, no enters-with-counters, no ETB
// hook.
//
// # Two events, in the order the rules apply them
//
// CR 701.7b makes "create N tokens" ONE event, so a doubler modifies
// the instruction rather than each token:
//
//  1. RepEventCreateTokens — the creation instruction. Opened ONCE
//     however many tokens it makes and whatever kinds they are.
//     Parallel Lives, Anointed Procession, Doubling Season, Primal
//     Vigor and Mondrak multiply its counts; Academy Manufactor
//     rewrites its kind set. Two Anointed Processions are two
//     objects contributing ONE declared effect, so #792's
//     identical-window skip applies and nobody is asked to order
//     them — the answer is ×4 either way. A Doubling Season and an
//     Academy Manufactor are DIFFERENT effects and the affected
//     player really is asked, because the orders differ: doubling a
//     Clue and then making one of each is two of each kind, and
//     making one of each and then doubling is also two of each —
//     but a Manufactor applied to a Clue that Doubling Season has
//     already turned into two Clues makes two of each, while a
//     Manufactor that goes first turns one Clue into one of each and
//     the Season then doubles all three.
//
//  2. One ordinary battlefield ENTRY per token the first event
//     settled on, through enterBattlefieldThroughPipelineLocked —
//     the one effect-side entry primitive #478 built for the
//     library search, the exile return and the reanimation. A token
//     is a fourth caller of it and nothing more: a RepEventMove with
//     no old zone, finished by the same
//     executeEntryToBattlefieldLocked, carrying the rest of its
//     batch on the same entryTail. That is where enters-tapped
//     (Urabrask, Kismet, Thalia), enters-with-counters (Renata,
//     Arwen, Dragonstorm Globe) and fireETBHookLocked reach a token
//     for the first time.
//
// The split is the rules' own: CR 616.1's "would create" replacements
// apply to the instruction, and entry replacements apply to each
// permanent as it enters. Collapsing them into one event would make a
// doubler and an enters-tapped effect compete for an ordering prompt
// the rules never offer.
//
// # Both halves can PAUSE, so the batch is a value carried forward
//
// A creation whose window queues a CR 616 ordering prompt returns
// with NOTHING created; the resume makes the tokens when the prompt
// is answered. An individual token's ENTRY can pause the same way
// (two distinct enters-tapped effects on one opponent's token), and
// it resumes through #478's entry frame — there is no second token
// resume. The tokens behind it in the batch ride that entry's
// entryTail rather than the next line of a loop, which is the same
// idiom discardCardsLocked uses for an exit and for the same reason:
// a batch carried forward as a value is what makes an undo across the
// prompt land where a clean run would.
//
// tokenTail is the CREATION half of that, for the window in step 1,
// which has no entry to hang a tail on yet. A caller that has more to
// do with the tokens it made — "create a token, then sacrifice it" —
// hands it over as a continuation (CreateTokensThenForEffect) instead
// of reading the returned IDs on the next line.

// TokenGroup is one KIND of token inside a creation instruction: a
// template, how many of it, and the creation-time entry options the
// instruction asked for.
//
// A group is the unit a replacement rewrites. "Create three 1/1
// Soldiers" is one group of three; Academy Manufactor turns one group
// of Clues into three groups (Clue, Food, Treasure) of the same
// count; a doubler multiplies every group's count.
type TokenGroup struct {
	// Template is the printed token, straight out of the token table
	// or built by a copy effect. Every token in the group is minted
	// from it.
	Template Card

	// Count is how many of Template the instruction creates. Zero or
	// negative means the group makes nothing, which is what a
	// cancelled group looks like.
	Count int

	// Entry is the creation's own entry clause — "create a TAPPED
	// Treasure", "create a 1/1 Soldier with a +1/+1 counter on it",
	// "…with haste". Properties of the CREATION, not of the token:
	// two cards can make the same printed Powerstone and only one of
	// them says "tapped".
	Entry TokenEntryOptions
}

// TokenCreation is one "create N tokens" instruction — the whole
// input to CreateTokensThenForEffect, and what the CR 701.7b
// replacement window is opened on.
type TokenCreation struct {
	// Controller is the player the tokens are created under the
	// control of ("create one or more tokens UNDER YOUR CONTROL").
	Controller uuid.UUID

	// Groups is what the instruction creates, one entry per kind.
	Groups []TokenGroup

	// Attacking is the player the tokens are created attacking
	// (Hanweir Garrison, Parhelion II). CR 506.3c: a permanent PUT
	// onto the battlefield attacking was never DECLARED as an
	// attacker, so no "whenever ~ attacks" trigger fires and the
	// declaration lock-in leaves it alone. uuid.Nil, or a player who
	// is not seated, creates the tokens not attacking rather than
	// erroring — the ability has already resolved, and the tokens are
	// the part of it that can still be delivered.
	Attacking uuid.UUID

	// Source is the card whose effect is creating the tokens, stamped
	// on the emitted EventTokenCreated. uuid.Nil for a creation with
	// no source card (a test fixture, an admin verb).
	Source uuid.UUID
}

// tokenTail is the CREATION event's answer to lifeTail, damageTail
// and zoneRoute.then: the rest of whatever asked for the creation,
// run with the IDs of the tokens that actually landed, once both
// pipelines have settled. The per-token ENTRY leg does not need one —
// it rides #478's entryTail, the tail every other effect-side entry
// already carries.
//
// It is cleared THROUGH the pointer as it runs (runTokenTailLocked),
// which is why cloneReplacementResume gives an undo snapshot its own
// copy: sharing the struct would let the live game's run consume the
// snapshot's continuation, so undoing the answer to a CR 616 prompt
// and answering it again would make the tokens and skip the rest of
// the card.
type tokenTail struct {
	// then is the continuation. It takes the live *Game on the
	// undo-safety contract every other continuation in the engine
	// follows, and runs with g.mu held, so it may start the next
	// move or queue the next prompt itself.
	then func(g *Game, created []uuid.UUID) error
}

// TokenCount is the total number of tokens the instruction currently
// creates, across every group. Card-facing: a replacement that only
// applies to a real creation ("if one or more tokens WOULD BE
// created") asks this rather than reading TokenGroups.
func (ev *ReplacementEvent) TokenCount() int {
	if ev == nil {
		return 0
	}
	n := 0
	for _, grp := range ev.TokenGroups {
		if grp.Count > 0 {
			n += grp.Count
		}
	}
	return n
}

// MultiplyTokens multiplies every group's count — the whole of what a
// token doubler does (CR 701.7b: "twice that many of those tokens are
// created instead"). Parallel Lives, Anointed Procession, Doubling
// Season, Primal Vigor and Mondrak all call it with 2.
//
// Multiplicative by construction, which is what CR 616.1 makes two
// doublers: applying ×2 and then ×2 to the same event is ×4, whatever
// order the affected player puts them in.
func (ev *ReplacementEvent) MultiplyTokens(n int) {
	if ev == nil || n < 0 {
		return
	}
	for i := range ev.TokenGroups {
		if ev.TokenGroups[i].Count > 0 {
			ev.TokenGroups[i].Count *= n
		}
	}
}

// ReplaceTokenKinds rewrites WHAT the instruction creates, keeping
// each group's count and entry clause: every group becomes one group
// per template.
//
// Passing no templates is a no-op rather than a cancellation: an
// effect that means "create nothing instead" cancels the event.
func (ev *ReplacementEvent) ReplaceTokenKinds(templates ...Card) {
	ev.ReplaceTokenKindsWhere(nil, templates...)
}

// ReplaceTokenKindsWhere is ReplaceTokenKinds narrowed to the groups
// whose template satisfies `pred` — Academy Manufactor's "if you
// would create a Clue, Food, or Treasure token, instead create one of
// each", which must leave the Soldiers alone when the same
// instruction makes both. A nil predicate rewrites every group.
func (ev *ReplacementEvent) ReplaceTokenKindsWhere(pred func(Card) bool, templates ...Card) {
	if ev == nil || len(templates) == 0 {
		return
	}
	out := make([]TokenGroup, 0, len(ev.TokenGroups)*len(templates))
	for _, grp := range ev.TokenGroups {
		if pred != nil && !pred(grp.Template) {
			out = append(out, grp)
			continue
		}
		for _, tmpl := range templates {
			out = append(out, TokenGroup{Template: tmpl, Count: grp.Count, Entry: grp.Entry})
		}
	}
	ev.TokenGroups = out
}

// TokenTemplatesMatch reports whether any group's template satisfies
// `pred` — the shape every "if you would create a <kind> token"
// replacement needs (Academy Manufactor's Clue / Food / Treasure,
// Xorn's Treasure, Divine Visitation's creature token).
//
// Card-facing read-only helper: a replacement must never mutate the
// event from AppliesTo.
func (ev *ReplacementEvent) TokenTemplatesMatch(pred func(Card) bool) bool {
	if ev == nil || pred == nil {
		return false
	}
	for _, grp := range ev.TokenGroups {
		if grp.Count > 0 && pred(grp.Template) {
			return true
		}
	}
	return false
}

// runTokenTailLocked runs a settled creation's continuation exactly
// once, with the IDs of the tokens that actually landed. The
// continuation is cleared before it runs, so a tail that re-enters
// the pipeline on the same event cannot run itself twice.
//
// Every TERMINAL outcome goes through here — created, cancelled
// (CR 614.10), or abandoned because the prompt was taken away —
// because a caller sequencing work behind the tokens has to be told
// even when the answer is "none", or it waits forever. That is the
// call #808 made for the life tail and #853 for the route tail.
//
// Caller must hold g.mu.
func (g *Game) runTokenTailLocked(ev *ReplacementEvent, created []uuid.UUID) error {
	if ev == nil {
		return nil
	}
	return runTokenTailValueLocked(g, ev.tokenTail, created)
}

// runTokenTailValueLocked is runTokenTailLocked for a tail held
// somewhere other than on an event — the batch's own "done", which
// outlives the per-token entry events it is threaded through.
//
// Caller must hold g.mu.
func runTokenTailValueLocked(g *Game, tail *tokenTail, created []uuid.UUID) error {
	if tail == nil || tail.then == nil {
		return nil
	}
	then := tail.then
	tail.then = nil
	return then(g, created)
}

// CreateTokensThenForEffect is the engine's one token-creation entry
// point: it opens the CR 701.7b replacement window on `spec`, and
// once the window settles it enters every token the instruction ended
// up creating, each through the ordinary battlefield-entry pipeline.
//
// `then` is the rest of the effect, run with the IDs of the tokens
// that actually landed, in creation order. It runs from the landing
// whether or not anything paused — inline when nothing did, and from
// the CR 616 resume when something did — which is the whole reason it
// is a continuation rather than a return value. A caller with nothing
// to do afterwards passes nil.
//
// Caller must hold g.mu.
func (g *Game) CreateTokensThenForEffect(spec TokenCreation, then func(g *Game, created []uuid.UUID) error) error {
	groups := make([]TokenGroup, 0, len(spec.Groups))
	for _, grp := range spec.Groups {
		if grp.Count > 0 {
			groups = append(groups, grp)
		}
	}
	ev := &ReplacementEvent{
		Kind:            RepEventCreateTokens,
		Actor:           spec.Controller,
		Source:          spec.Source,
		TokenController: spec.Controller,
		TokenGroups:     groups,
		TokenAttacking:  spec.Attacking,
	}
	if then != nil {
		ev.tokenTail = &tokenTail{then: then}
	}
	if len(groups) == 0 {
		// Nothing to create. Still a terminal outcome: a caller
		// sequenced behind the tail has to be told.
		return g.runTokenTailLocked(ev, nil)
	}
	return g.createTokensLocked(ev)
}

// createTokensLocked runs the CR 701.7b window for ev and, once it
// settles, creates what the window left. Shared by the inline path
// above and by the CR 616 resume
// (applyResolvedReplacementEventLocked), so a paused creation and an
// unpaused one cannot drift apart.
//
// Caller must hold g.mu.
func (g *Game) createTokensLocked(ev *ReplacementEvent) error {
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the creation now.
		// NOTHING has been created: no token exists, no event was
		// emitted, and the caller's continuation is still owed.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement: the tokens are simply
		// not created. The caller's continuation still runs.
		return g.runTokenTailLocked(ev, nil)
	}
	return g.applyResolvedTokenCreationLocked(out)
}

// applyResolvedTokenCreationLocked mints the tokens a settled
// RepEventCreateTokens describes and enters them one at a time.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedTokenCreationLocked(ev *ReplacementEvent) error {
	batch := make([]stagedToken, 0, ev.TokenCount())
	attacking := ev.TokenAttacking != uuid.Nil && g.playerByIDLocked(ev.TokenAttacking) != nil
	for _, grp := range ev.TokenGroups {
		for i := 0; i < grp.Count; i++ {
			tok := g.mintTokenLocked(grp, ev.TokenController)
			if attacking {
				tok.AttackingTarget = ev.TokenAttacking
			}
			batch = append(batch, stagedToken{card: tok, entry: grp.Entry})
		}
	}
	if len(batch) == 0 {
		return g.runTokenTailLocked(ev, nil)
	}
	return g.enterCreatedTokensLocked(batch, nil, ev.Source, ev.tokenTail)
}

// stagedToken is one minted-but-not-yet-entered token: the object
// itself plus the creation's entry clause for it.
type stagedToken struct {
	card  Card
	entry TokenEntryOptions
}

// mintTokenLocked builds one token object from a group's template.
// Everything per-instance is reset (a fresh InstanceID, no counters,
// no knowers) and everything the CREATION asked for that belongs on
// the object rather than on the entry event is stamped here.
//
// Caller must hold g.mu.
func (g *Game) mintTokenLocked(grp TokenGroup, controller uuid.UUID) Card {
	tok := grp.Template
	tok.InstanceID = uuid.New()
	tok.Owner = controller
	tok.Controller = controller
	tok.Counters = nil
	tok.LostLastCounter = false
	tok.KnownBy = nil
	if grp.Entry.Tapped {
		// Additive, not an assignment: a caller that pre-stamped
		// Tapped on the template (the older idiom — Mary Read's
		// Treasure, Hashaton's Zombie) must keep entering tapped.
		// The ENTRY event carries the same instruction, so this
		// clause and an enters-tapped replacement settle in one field
		// and neither can lose the other.
		tok.Tapped = true
	}
	if len(grp.Entry.Keywords) > 0 {
		// Fresh slice: the template's Keywords slice is shared by
		// every token minted from it, so appending in place would
		// leak the grant onto the next one.
		kw := make([]string, 0, len(grp.Template.Keywords)+len(grp.Entry.Keywords))
		kw = append(kw, grp.Template.Keywords...)
		kw = append(kw, grp.Entry.Keywords...)
		tok.Keywords = kw
	}
	for _, seat := range g.Seats {
		tok.AddKnower(seat.ID)
	}
	return tok
}

// enterCreatedTokensLocked enters the batch one token at a time,
// each through enterBattlefieldThroughPipelineLocked — the same entry
// primitive a library search, an exile return and a reanimation go
// through since #478. There is no token entry path.
//
// The rest of the batch is the CURRENT token's entryTail rather than
// the next line of a loop: an entry can pause on a CR 616 ordering
// prompt, and the tokens behind it must land on the far side of that
// prompt rather than being created into a board whose first token has
// not settled. `created` is carried forward as a value for the same
// reason discardCardsLocked carries its shrinking slice — so an undo
// across the prompt lands where a clean run would.
//
// Caller must hold g.mu.
func (g *Game) enterCreatedTokensLocked(batch []stagedToken, created []uuid.UUID, source uuid.UUID, done *tokenTail) error {
	if len(batch) == 0 {
		return runTokenTailValueLocked(g, done, created)
	}
	next, rest := batch[0], batch[1:]
	g.enteringTokens = append(g.enteringTokens, next.card)
	ev := &ReplacementEvent{
		Kind:   RepEventMove,
		Actor:  next.card.Controller,
		Source: source,
		CardID: next.card.InstanceID,
		// No OLD zone: a token is created on the battlefield and came
		// from nowhere (CR 111.1). Every entry replacement in the
		// catalog keys on NewZone, so none of them notices, and the
		// empty old zone is what the finisher reads to push the staged
		// token instead of moving a card.
		NewZone:      ZoneBattlefield,
		NewZoneOwner: next.card.Controller,
		EntersTapped: next.entry.Tapped,
		// The entry may pause, and #478's frame finishes it. Nothing is
		// skipped by resuming one: there is no shuffle owed and no new
		// object identity to mint.
		entryResumable: true,
	}
	for name, n := range next.entry.Counters {
		ev.AddCounterAtETB(name, n)
	}
	ev.entryTail = &entryTail{then: func(g *Game, entered uuid.UUID) error {
		// uuid.Nil means nothing entered — the window cancelled or
		// redirected it, or its prompt was taken away. A token that
		// never reached the battlefield never existed (CR 111.1), and
		// the rest of the batch still does.
		g.dropEnteringTokenLocked(next.card.InstanceID)
		if entered != uuid.Nil {
			created = append(created, entered)
		}
		return g.enterCreatedTokensLocked(rest, created, source, done)
	}}
	_, err := g.enterBattlefieldThroughPipelineLocked(ev)
	return err
}

// dropEnteringTokenLocked removes a staged token that will never
// enter. A token exists only on the battlefield (CR 111.1), so one
// whose entry was replaced away or abandoned simply ceases to be.
// No-op for an ID that was never staged, which is every entry that is
// a real card coming out of a real zone.
//
// Caller must hold g.mu.
func (g *Game) dropEnteringTokenLocked(id uuid.UUID) {
	for i := range g.enteringTokens {
		if g.enteringTokens[i].InstanceID != id {
			continue
		}
		g.enteringTokens = append(g.enteringTokens[:i], g.enteringTokens[i+1:]...)
		return
	}
}

// takeEnteringTokenLocked pops the staged token `id` and returns it.
// ok is false when nothing is staged under that ID — the ordinary
// answer for every entry that is a real card coming out of a real
// zone.
//
// Caller must hold g.mu.
func (g *Game) takeEnteringTokenLocked(id uuid.UUID) (Card, bool) {
	for i := range g.enteringTokens {
		if g.enteringTokens[i].InstanceID != id {
			continue
		}
		tok := g.enteringTokens[i]
		g.enteringTokens = append(g.enteringTokens[:i], g.enteringTokens[i+1:]...)
		return tok, true
	}
	return Card{}, false
}

// enteringTokenLocked reads a staged token without removing it, so
// LookupCardForEffect can answer for a token whose entry window is
// open. That is what lets Urabrask the Hidden, Kismet and Thalia read
// the entering permanent's controller and card types the same way
// they read a creature entering from a hand.
//
// Caller must hold g.mu.
func (g *Game) enteringTokenLocked(id uuid.UUID) (Card, bool) {
	for i := range g.enteringTokens {
		if g.enteringTokens[i].InstanceID == id {
			return g.enteringTokens[i], true
		}
	}
	return Card{}, false
}
