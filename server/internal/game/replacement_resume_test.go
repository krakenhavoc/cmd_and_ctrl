package game

import (
	"testing"

	"github.com/google/uuid"
)

// replacement_resume_test.go — #1156, the in-app report "Curiosity
// should go to the yard … instead it went to my enchantments in the
// battlefield".
//
// The CR 614 pipeline can PAUSE. When it does, the caller that asked
// for the mutation returns early and says so in as many words —
// moveCardByRefLocked's "Nothing has moved, so there is nothing for
// the state checks to see; the resume lands the card" — and the
// resume then lands the card and returned straight to the action
// layer. Nobody ran the state-based actions, so whatever the move
// made illegal stayed on the board until some later action happened
// to run them.
//
// The reporter's board is the ordinary Commander one: an Aura on a
// commander creature. The commander leaves, the owner is asked
// CR 903.9's "put it in the command zone instead?", and the Aura is
// left attached to a card that is no longer on the battlefield — no
// CR 704.5m, no graveyard, and the client draws the orphan in the
// enchantments row because that is what an unattached enchantment
// looks like. Since #539 made that window open on every exit "from
// anywhere", it is how a commander usually leaves.
//
// The fix is one call at the tail both resume paths share
// (finishReplacementResumeLocked), and the tests below are deliberately
// NOT all about attachments: the point is that answering a replacement
// prompt is a CR 117.5 boundary like every other Resolve* entry point
// in pending_choice.go, so anything the pass owes gets done. Same
// shape as #370's discard_selection, pinned in
// cards/effects/inapp_trigger_reports_test.go.

// pushResumeCommander seats a commander creature and fires the
// zone-move event, so the layer listener has stamped it before the
// attachment SBA reads its effective types.
func pushResumeCommander(g *Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  id,
		Name:        "Atraxa, Praetors' Voice",
		TypeLine:    "Legendary Creature — Phyrexian Angel Horror",
		Power:       4,
		Toughness:   4,
		Owner:       owner,
		Controller:  owner,
		IsCommander: true,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	return id
}

// answerTheCommanderPrompt answers the one queued CR 903.9 optional
// replacement, asserting there is exactly one.
func answerTheCommanderPrompt(t *testing.T, g *Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected one CR 903.9 prompt, got %d pending choices", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceOptionalReplacement)
	}
	if err := g.ResolveOptionalReplacement(p.ID, chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// The report itself: the Aura's host takes the command zone, and
// CR 704.5m has to run on the far side of the prompt.
func TestOptionalReplacementResumeRunsTheAttachmentSBA(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cmd := pushResumeCommander(g, me.ID)
	aura := pushAttachTestCard(g, me.ID, "Curiosity", "Enchantment — Aura")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: cmd}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, cmd); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	// The prompt gates the move: nothing has left the battlefield yet,
	// so the Aura is still legally attached and must NOT have moved.
	if !g.Battlefield.Contains(cmd) {
		t.Fatal("the commander left the battlefield before the prompt was answered")
	}
	if _, ok := battlefieldCardByID(g, aura); !ok {
		t.Fatal("the Aura fell off while its host was still on the battlefield")
	}

	answerTheCommanderPrompt(t, g, me.ID, true)

	if !me.Command.Contains(cmd) {
		t.Fatal("the commander did not reach the command zone")
	}
	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Error("CR 704.5m: the Aura stayed on the battlefield after its host took the command zone")
	}
	if !graveyardHas(me, aura) {
		t.Error("CR 704.5m: the Aura is not in its owner's graveyard")
	}
}

// Declining the offer is the same boundary. The commander goes to the
// graveyard the ordinary way and the Aura still has no host.
func TestDeclinedOptionalReplacementResumeRunsTheAttachmentSBA(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cmd := pushResumeCommander(g, me.ID)
	aura := pushAttachTestCard(g, me.ID, "Curiosity", "Enchantment — Aura")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: cmd})
	})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, cmd); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	answerTheCommanderPrompt(t, g, me.ID, false)

	if !me.Graveyard.Contains(cmd) {
		t.Fatal("the declined commander did not reach the graveyard")
	}
	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Error("CR 704.5m: the Aura stayed on the battlefield after its host died")
	}
}

// A CATALOGUED Aura takes the other branch of attachmentLegalLocked —
// its own enchant clause, re-asked against a host that is gone — and
// has to come out the same way. The two branches are the reason
// #1156's report could not be answered by looking at the SBA alone:
// both were already right, and neither was being run.
func TestOptionalReplacementResumeSweepsACataloguedAuraToo(t *testing.T) {
	withCatalogTargetSpec(t, func(string) *TargetSpec { return enchantCreatureSpec() })
	g := newActiveGame(t)
	me := g.Seats[0]
	cmd := pushResumeCommander(g, me.ID)
	aura := pushCataloguedAura(g, me.ID, "Curiosity")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: cmd})
	})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, cmd); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	answerTheCommanderPrompt(t, g, me.ID, true)

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Error("CR 704.5m: the catalogued Aura stayed on the battlefield")
	}
	if !graveyardHas(me, aura) {
		t.Error("CR 704.5m: the catalogued Aura is not in its owner's graveyard")
	}
}

// Not an attachment rule at all, and that is the point: answering the
// prompt is the CR 117.5 boundary, so EVERY state-based action the
// board owes is performed there. A creature shrunk to zero toughness
// while the move was paused is swept by the same pass (CR 704.5f).
func TestOptionalReplacementResumeRunsEveryStateBasedAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cmd := pushResumeCommander(g, me.ID)

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, cmd); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	// Shrunk to nothing behind the open prompt, so nothing has looked
	// at it yet. AddCounterForEffect is the *ForEffect surface: it
	// runs inside a held lock and does not run the checks itself.
	doomed := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: doomed,
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: doomed, OldZone: ZoneHand, NewZone: ZoneBattlefield})
		if err := g.AddCounterForEffect(doomed, "-1/-1", 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if _, ok := battlefieldCardByID(g, doomed); !ok {
		t.Fatal("setup: the shrunken creature is not on the battlefield")
	}

	answerTheCommanderPrompt(t, g, me.ID, true)

	if _, ok := battlefieldCardByID(g, doomed); ok {
		t.Error("CR 704.5f: a 0-toughness creature survived the replacement resume")
	}
}
