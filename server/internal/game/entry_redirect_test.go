package game

import (
	"testing"

	"github.com/google/uuid"
)

// entry_redirect_test.go — ADR 0098's engine half: a battlefield entry
// that a replacement REDIRECTS elsewhere, and the decline rule for a
// question nobody can be asked.
//
// Before ADR 0098 no catalog card redirected an entry, and the entry
// finishers disagreed about what one meant. Three treated it as a
// cancel (the card stayed where it was) and two — stack resolution and
// the land play — did not look at the destination at all and put the
// card onto the battlefield anyway. These tests use a synthetic card
// whose own replacement unconditionally rewrites its entry to its
// owner's graveyard, so every site is exercised without a prompt.

const redirectTestOracle = "test-redirected-entry"

// stubRedirectedEntry registers a card whose own "if this would enter,
// put it into its owner's graveyard instead" is mandatory.
func stubRedirectedEntry(t *testing.T) {
	t.Helper()
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		redirectTestOracle: {{
			Watches:         []EventKind{EventZoneMove},
			SelfReplacement: true,
			Label:           "Test redirect",
			AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
			Replace: func(ev *ReplacementEvent, _ *Game, src *Card) error {
				ev.NewZone = ZoneGraveyard
				ev.NewZoneOwner = src.Owner
				return nil
			},
		}},
	})
}

// TestARedirectedPermanentSpellGoesWhereItsWindowSentIt is the stack
// resolution site. It used to call the entry finisher whatever the
// settled destination was, so the artifact entered the battlefield.
func TestARedirectedPermanentSpellGoesWhereItsWindowSentIt(t *testing.T) {
	g := newActiveGame(t)
	stubRedirectedEntry(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(Card{InstanceID: id, Name: "Redirected Rock", TypeLine: "Artifact",
		OracleID: redirectTestOracle, Owner: me.ID, Controller: me.ID})
	toMainPhase(t, g)
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if _, ok := battlefieldCardByID(g, id); ok {
		t.Fatal("the redirected artifact entered the battlefield anyway")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the redirected artifact is not in its owner's graveyard")
	}
	if g.Stack.Contains(id) {
		t.Error("the redirected artifact was left on the stack")
	}
	for _, ev := range g.Events {
		if ev.Kind == EventETB && ev.CardID == id {
			t.Error("EventETB fired for a permanent that never entered (CR 614.6)")
		}
	}
}

// TestARedirectedLandPlaySpendsTheLandDrop is the land-play site, and
// ADR 0098 owner decision 7: the play happened (CR 116.2a), so the
// turn's land drop is used even though the land never entered.
func TestARedirectedLandPlaySpendsTheLandDrop(t *testing.T) {
	g := newActiveGame(t)
	stubRedirectedEntry(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(Card{InstanceID: id, Name: "Redirected Land", TypeLine: "Land",
		OracleID: redirectTestOracle, Owner: me.ID, Controller: me.ID})
	toMainPhase(t, g)
	before := g.LandsPlayedThisTurn[me.ID]
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	if _, ok := battlefieldCardByID(g, id); ok {
		t.Fatal("the redirected land entered the battlefield anyway")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the redirected land is not in its owner's graveyard")
	}
	if got := g.LandsPlayedThisTurn[me.ID]; got != before+1 {
		t.Errorf("lands played this turn = %d, want %d: a redirected land play still spends the drop", got, before+1)
	}
}

// TestARedirectedEffectEntryIsMovedNotCancelled is the effect-side
// entry primitive (enterBattlefieldThroughPipelineLocked), reached
// through an exile return. It used to treat the redirect as a cancel,
// leaving the card in exile.
func TestARedirectedEffectEntryIsMovedNotCancelled(t *testing.T) {
	g := newActiveGame(t)
	stubRedirectedEntry(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.Exile.PushTop(Card{InstanceID: id, Name: "Redirected Rock", TypeLine: "Artifact",
		OracleID: redirectTestOracle, Owner: me.ID, Controller: me.ID})
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() { entered, err = g.ReturnFromExileToBattlefieldForEffect(id, me.ID, false) })
	if err != nil {
		t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
	}
	if entered != uuid.Nil {
		t.Errorf("reported %s as entered; nothing entered", entered)
	}
	if g.Exile.Contains(id) {
		t.Fatal("the redirected card was left in exile (the old cancel reading)")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the redirected card is not in its owner's graveyard")
	}
}

// TestARedirectedBatchMemberIsMoved is the simultaneous "put onto the
// battlefield" batch: a member whose window redirected it is moved as
// the batch lands, and its sibling still enters.
func TestARedirectedBatchMemberIsMoved(t *testing.T) {
	g := newActiveGame(t)
	stubRedirectedEntry(t)
	me := g.Seats[g.Turn.ActiveSeat]
	redirected := uuid.New()
	plain := uuid.New()
	me.Hand.PushTop(Card{InstanceID: redirected, Name: "Redirected Rock", TypeLine: "Artifact",
		OracleID: redirectTestOracle, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(Card{InstanceID: plain, Name: "Plain Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID})
	var got []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		err = g.PutOntoBattlefieldTogetherThenForEffect([]BatchEntry{
			{CardID: redirected, From: ZoneHand},
			{CardID: plain, From: ZoneHand},
		}, ZoneEntryOptions{}, func(_ *Game, entered []uuid.UUID) error {
			got = entered
			return nil
		})
	})
	if err != nil {
		t.Fatalf("PutOntoBattlefieldTogetherThenForEffect: %v", err)
	}
	if !me.Graveyard.Contains(redirected) {
		t.Error("the redirected member is not in its owner's graveyard")
	}
	if _, ok := battlefieldCardByID(g, plain); !ok {
		t.Error("the other member did not enter")
	}
	if len(got) != 1 || got[0] != plain {
		t.Errorf("entered = %v, want just the plain rock", got)
	}
}

// TestAnUnaskedDeclineIsReplaceRunsItsReplace is ADR 0098 Decision 5.
// A window that cannot ask (mustSettleNow, an ordering whose chooser
// has left) used to mark every question-asking effect applied WITHOUT
// running Replace. For a "may" that is the weaker branch; for a
// shockland, a reveal-land or a Mox Diamond, whose Replace IS the "you
// didn't" branch, it was the stronger one.
func TestAnUnaskedDeclineIsReplaceRunsItsReplace(t *testing.T) {
	g := newActiveGame(t)
	tapped := func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		ev.EntersTapped = true
		return nil
	}
	cases := []struct {
		name   string
		effect ReplacementEffect
		want   bool
	}{
		{"shockland (EntryLifeCost)", ReplacementEffect{EntryLifeCost: 2, Replace: tapped}, true},
		{"reveal-land (EntryCardChoice)", ReplacementEffect{EntryCardChoice: &EntryCardChoice{Max: 1}, Replace: tapped}, true},
		{"a \"may\" (Optional)", ReplacementEffect{Optional: true, Replace: tapped}, false},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := &ReplacementEvent{ID: ReplacementEventID(9000 + i), Kind: RepEventMove, NewZone: ZoneBattlefield}
			g.WithWriteLock(func() {
				if g.replacementsAppliedThisEvent == nil {
					g.replacementsAppliedThisEvent = map[ReplacementEventID]map[ReplacementEffectID]bool{}
				}
				g.replacementsAppliedThisEvent[ev.ID] = map[ReplacementEffectID]bool{}
				rest := g.skipQuestionsLocked(ev, []activeReplacement{{id: 1, effect: tc.effect}})
				if len(rest) != 0 {
					t.Errorf("skipQuestionsLocked kept %d effects; a question-asking effect is never in the gathered order", len(rest))
				}
				if !g.replacementsAppliedThisEvent[ev.ID][1] {
					t.Error("the skipped effect was not marked applied")
				}
			})
			if ev.EntersTapped != tc.want {
				t.Errorf("EntersTapped = %v, want %v", ev.EntersTapped, tc.want)
			}
		})
	}
}
