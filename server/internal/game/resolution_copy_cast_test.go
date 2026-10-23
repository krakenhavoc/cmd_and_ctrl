package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// resolution_copy_cast_test.go — casting copies of cards during the
// resolution that made them (CR 707.12, CR 608.2g; ADR 0066's 2026-10-22
// amendment, #2947). Uncatalogued fixtures: the cards' own tests are
// Kaervek, the Punisher's and the other CopyAndCast cards'.

// resolveCopyingAbility puts an ability on the stack that copies `orig`
// with a during-resolution offer to me, resolves it, and returns the
// copies it made.
func resolveCopyingAbility(t *testing.T, g *Game, me *Player, offer CardCopyOffer, orig ...uuid.UUID) []uuid.UUID {
	t.Helper()
	var copies []uuid.UUID
	offer.Player = me.ID
	offer.DuringResolution = true
	pushAbilityItem(g, StackItemActivated, me.ID, uuid.Nil, "Fixture copier", func(g *Game, _ *StackItem) error {
		copies = g.CopyCardsAndOfferCastsForEffect(offer, orig)
		return nil
	})
	resolveTop(t, g)
	return copies
}

// resolutionCastPrompt is me's open resolution_cast prompt, or nil.
func resolutionCastPrompt(g *Game, me uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceResolutionCast && c.Chooser == me {
			return c
		}
	}
	return nil
}

// The copies are offered inside the resolution: it stays open, nobody
// can pass, and the caster casts them in the order they like, each going
// on top. The last cast closes the prompt and the resolution.
func TestResolutionCastHoldsTheResolutionAndCastsInAnyOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := copyFixtureInGraveyard(me, "Fixture A", "Sorcery", "{R}")
	b := copyFixtureInGraveyard(me, "Fixture B", "Instant", "{R}")

	copies := resolveCopyingAbility(t, g, me, freeOffer(me), a, b)
	if len(copies) != 2 {
		t.Fatalf("copies = %v, want 2", copies)
	}
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(c.DiscardOptions) != 2 {
		t.Fatalf("prompt = %+v, want one over both copies", c)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the resolution is not held open by the prompt")
	}
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want the choice gate's refusal", err)
	}
	if n := len(cardCopiesAnywhere(g)); n != 2 {
		t.Fatalf("%d copies after the refused pass, want both still on offer", n)
	}

	// B first: it ends up under A.
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[1], CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast B's copy: %v", err)
	}
	c = resolutionCastPrompt(g, me.ID)
	if c == nil || len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != copies[0] {
		t.Fatalf("after one cast the prompt is %+v, want A's copy alone", c)
	}
	if !g.ResolutionPaused() {
		t.Error("one cast closed the resolution with a copy still on offer")
	}
	// A sorcery copy is cast though the stack is not empty (CR 608.2g).
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast A's copy: %v", err)
	}
	if resolutionCastPrompt(g, me.ID) != nil || g.ResolutionPaused() {
		t.Fatal("the prompt or the resolution outlived the last cast")
	}
	ia, ib := g.StackMeta[copies[0]], g.StackMeta[copies[1]]
	if ia == nil || ib == nil || ia.Seq <= ib.Seq {
		t.Fatalf("stack items A %+v, B %+v: want A's copy above B's", ia, ib)
	}
}

// The answers a prompt refuses leave it open: someone else's prompt, a
// card it does not offer, a non-empty resolve_choice. A cast the engine
// refuses (unpaid) leaves it open too, and the decline withdraws the
// offers, so the copies cease at the boundary.
func TestResolutionCastRefusalsAndDecline(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := copyFixtureInGraveyard(me, "Fixture A", "Instant", "{R}")
	b := copyFixtureInGraveyard(me, "Fixture B", "Instant", "{R}")
	priced := CardCopyOffer{SourceName: "Test", Timing: TimingFlash, Lapse: LapseStaysInExile}

	copies := resolveCopyingAbility(t, g, me, priced, a, b)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.CastForResolutionChoice(other.ID, c.ID, copies[0], CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another player's cast = %v, want ErrNotTheChooser", err)
	}
	if err := g.CastForResolutionChoice(me.ID, c.ID, a, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("casting the original = %v, want ErrInvalidParam", err)
	}
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Error("the copy was cast with no {R}: the price is charged even off strict mana")
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{copies[0]}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("resolve_choice naming a copy = %v, want ErrInvalidParam", err)
	}
	if resolutionCastPrompt(g, me.ID) == nil {
		t.Fatal("a refused answer closed the prompt")
	}

	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if g.ResolutionPaused() || resolutionCastPrompt(g, me.ID) != nil {
		t.Error("the decline did not let the resolution finish")
	}
	if n := len(cardCopiesAnywhere(g)); n != 0 {
		t.Errorf("%d copies after the decline, want none (CR 704.5e)", n)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("original %v moved", id)
		}
	}
	for _, perm := range me.CastPermissions {
		if perm.Scope == ScopeCards {
			t.Errorf("an offer outlived the decline: %+v", perm)
		}
	}
}

// A chooser who leaves takes the prompt with them (CR 800.4a); the
// resolution finishes without the casts and nothing is left offered.
func TestResolutionCastChooserLeaving(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := copyFixtureInGraveyard(me, "Fixture A", "Instant", "{R}")
	resolveCopyingAbility(t, g, me, freeOffer(me), a)
	if resolutionCastPrompt(g, me.ID) == nil {
		t.Fatal("no prompt")
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if resolutionCastPrompt(g, me.ID) != nil {
		t.Error("the departed chooser's prompt is still open")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held by a departed chooser")
	}
}

// Outside a resolution there is nothing to hold open: the offer is the
// permission alone, as it always was.
func TestResolutionCastOutsideAResolutionQueuesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := copyFixtureInGraveyard(me, "Fixture A", "Instant", "{R}")
	offer := freeOffer(me)
	offer.DuringResolution = true
	g.WithWriteLock(func() { g.CopyCardsAndOfferCastsForEffect(offer, []uuid.UUID{a}) })
	if c := resolutionCastPrompt(g, me.ID); c != nil {
		t.Errorf("a prompt was queued with no resolution open: %+v", c)
	}
	if n := len(cardCopiesAnywhere(g)); n != 1 {
		t.Errorf("%d copies, want the one offered", n)
	}
}

// A clone (an undo entry) is untouched by a cast made through the live
// prompt, and the clone's own prompt still answers.
func TestResolutionCastCloneIsIndependent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := copyFixtureInGraveyard(me, "Fixture A", "Instant", "{R}")
	b := copyFixtureInGraveyard(me, "Fixture B", "Instant", "{R}")
	copies := resolveCopyingAbility(t, g, me, freeOffer(me), a, b)
	clone := g.Clone()

	c := resolutionCastPrompt(g, me.ID)
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	cc := resolutionCastPrompt(clone, me.ID)
	if cc == nil || len(cc.DiscardOptions) != 2 {
		t.Fatalf("clone's prompt = %+v, want both copies still offered", cc)
	}
	if err := clone.CastForResolutionChoice(me.ID, cc.ID, copies[1], CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast in the clone: %v", err)
	}
	if got := resolutionCastPrompt(clone, me.ID); got == nil || len(got.DiscardOptions) != 1 || got.DiscardOptions[0] != copies[0] {
		t.Errorf("clone's prompt after its cast = %+v, want A's copy left", got)
	}
}
