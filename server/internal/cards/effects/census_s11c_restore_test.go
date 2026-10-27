package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s11c_restore_test.go — #4361 S11c: the last engine route batches
// whose continuation was a closure (zoneRoute.then) now carry it as data,
// and so does the hand-size discard. Each test pauses the move on a CR 616
// ordering prompt over a card named Needle (two probes that replace the
// move with itself), restores the table STRICTLY (the boot path) while the
// prompt is open, answers on the restored table, and checks what the
// continuation did:
//
//   - a surveil's graveyard leg (engine/surveil-routed): the EventSurveil
//     and the card's own "then" after it;
//   - a library search to a graveyard (routeBatchRun.Search): the search
//     finishes (EventSearchLibrary) and runs its ThenTail;
//   - Hinder's leg (engine/library-order-leg-landed on a keyed counter
//     route, engine/countered-to-library): the countered spell lands where
//     its controller put it;
//   - Guile's exile instead of Hinder's counter
//     (engine/counter-exiled-instead): the free cast is offered after the
//     restored answer;
//   - the CR 514.1 hand-size discard (engine/cleanup-discarded): the
//     cleanup step's bookkeeping runs and the turn moves on.

// libraryOrderProbeOracles are two test cards, each "if a card named Needle
// would be put into a library, it is put into that library instead": two
// declared effects, so a Needle's move to a library asks for their order
// (CR 616.1). Catalog rows, so a restored table gathers them again.
var libraryOrderProbeOracles = []string{"test-4361-library-order-a", "test-4361-library-order-b"}

func init() {
	for i, oracle := range libraryOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Library Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneLibrary {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Library Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// s11cGainLife gains its controller one life, plus one for each ID it is
// told, plus the amount it is told: a "then" a test can read back.
var s11cGainLife = CardTail("test/4361-s11c-gain-life", func(ctx *Context, o game.TailOutcome) error {
	return ctx.Game.ChangePlayerLifeForEffect(uuid.Nil, ctx.Controller(), 1+len(o.IDs)+o.Amount)
})

// s11cOnlyPrompt is the one open prompt, of kind k.
func s11cOnlyPrompt(t *testing.T, g *game.Game, k game.PendingChoiceKind) *game.PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != k {
		t.Fatalf("pending choices %+v, want one %q prompt", g.PendingChoices, k)
	}
	return g.PendingChoices[0]
}

func s11cHasEvent(g *game.Game, kind game.EventKind) *game.Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return &g.Events[i]
		}
	}
	return nil
}

// --- a surveil ---------------------------------------------------------------

func TestASurveilsGraveyardLegIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryTopForTest(me, "Needle")
	filler := pushLibraryTopForTest(me, "Filler")
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	life := me.Life
	g.WithWriteLock(func() {
		g.SurveilKeyedForEffect(me.ID, uuid.Nil, 2, game.KeyedTail{Ref: s11cGainLife, Player: me.ID})
	})
	ask := s11cOnlyPrompt(t, g, game.PendingChoiceSurveil)
	if err := g.ResolveSurveil(ask.ID, me.ID, []uuid.UUID{filler, needle}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)
	if s11cHasEvent(g, game.EventSurveil) != nil {
		t.Fatal("the surveil was announced before its last card settled")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = seatByID(g, me.ID)
	if !me.Graveyard.Contains(filler) || !me.Library.Contains(needle) {
		t.Fatal("the restored table does not have the first card in the graveyard and the Needle on the library")
	}
	answerExileOrder(t, g)
	if !me.Graveyard.Contains(needle) {
		t.Fatal("the Needle is not in the graveyard after the restored answer")
	}
	ev := s11cHasEvent(g, game.EventSurveil)
	if ev == nil || ev.Amount != 2 || ev.LookedAt != 2 || ev.Actor != me.ID {
		t.Fatalf("EventSurveil = %+v, want two cards put into the graveyard of two looked at", ev)
	}
	// The card's "then" is told how many stayed on top (none).
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 from the surveil's keyed then", life, me.Life)
	}
}

// --- a library search to a graveyard ----------------------------------------

func TestASearchToAGraveyardIsARestorePointAtTheSecondCardsOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryTopForTest(me, "Needle")
	filler := pushLibraryTopForTest(me, "Filler")
	seedOrderProbes(g, me.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	life := me.Life
	g.WithWriteLock(func() {
		err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player:   me.ID,
			Pred:     func(game.Card) bool { return true },
			Dest:     game.ZoneGraveyard,
			Limit:    2,
			Shuffle:  true,
			ThenTail: game.KeyedTail{Ref: s11cGainLife, Player: me.ID},
		})
		if err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	ask := s11cOnlyPrompt(t, g, game.PendingChoiceSearchLibrary)
	if err := g.ResolveSearchLibrary(ask.ID, me.ID, []uuid.UUID{filler, needle}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = seatByID(g, me.ID)
	if !me.Graveyard.Contains(filler) || !me.Library.Contains(needle) {
		t.Fatal("the restored table does not have the first take in the graveyard and the Needle in the library")
	}
	if s11cHasEvent(g, game.EventSearchLibrary) != nil {
		t.Fatal("the search finished before its last take settled")
	}
	answerExileOrder(t, g)
	if !me.Graveyard.Contains(needle) {
		t.Fatal("the Needle is not in the graveyard after the restored answer")
	}
	if s11cHasEvent(g, game.EventSearchLibrary) == nil {
		t.Error("the search did not finish after the restored answer")
	}
	// The search's ThenTail is told the two cards found.
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 from the search's keyed then", life, me.Life)
	}
}

// --- Hinder ------------------------------------------------------------------

// castNeedleAt casts a Shock named Needle from caster's hand at a player.
func castNeedleAt(t *testing.T, g *game.Game, caster *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Needle", TypeLine: "Instant",
		OracleID: b5ShockOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at}},
	}); err != nil {
		t.Fatalf("CastSpell Needle: %v", err)
	}
	return id
}

// hinderANeedle has seat 1 cast a Needle at seat 0, seat 0 Hinder it and
// put it on the bottom of its owner's library, and returns the table with
// the counter's move open on whatever the caller seeded.
func hinderANeedle(t *testing.T, seed func(g *game.Game, me *game.Player)) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seed(g, me)
	seedLibrary(opp, "Their Top")
	needle := castNeedleAt(t, g, opp, me.ID)
	castCounterAt(t, g, me, "Hinder", hinderOracle, needle)
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("Hinder's top-or-bottom choice was not asked")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{needle}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)
	if !g.Stack.Contains(needle) {
		t.Fatal("the Needle left the stack before the ordering prompt was answered")
	}
	return g, me, opp, needle
}

func TestHindersCounterToTheLibraryIsARestorePoint(t *testing.T) {
	g, _, opp, needle := hinderANeedle(t, func(g *game.Game, me *game.Player) {
		seedOrderProbes(g, me.ID, libraryOrderProbeOracles, "Library Order Probe")
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	opp = seatByID(g, opp.ID)
	if g.Stack.Contains(needle) {
		t.Fatal("the Needle is still on the stack after the restored answer")
	}
	cards := opp.Library.Cards
	if len(cards) == 0 || cards[0].InstanceID != needle {
		t.Fatal("the Needle is not on the bottom of its owner's library after the restored answer")
	}
	if s11cHasEvent(g, game.EventCounterSpell) == nil {
		t.Error("the Needle was not countered")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts open after Hinder finished", len(g.PendingChoices))
	}
}

// Guile turns Hinder's counter into an exile: the exile's ordering prompt
// is a restore point, and the restored answer offers Guile's controller
// the free cast during Hinder's resolution (#4286).
func TestGuilesExileInsteadOfHindersCounterIsARestorePoint(t *testing.T) {
	g, me, _, needle := hinderANeedle(t, func(g *game.Game, me *game.Player) {
		pushGuile(g, me.ID)
		seedExileOrderProbes(g, me.ID)
	})

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	if !g.Exile.Contains(needle) {
		t.Fatal("the Needle is not in exile after the restored answer")
	}
	if s11cHasEvent(g, game.EventCounterSpell) != nil {
		t.Error("an exiled-instead spell was reported as countered")
	}
	offer := s11cOnlyPrompt(t, g, game.PendingChoiceResolutionCast)
	if offer.Chooser != me.ID {
		t.Errorf("the free cast is offered to %s, want Guile's controller", offer.Chooser)
	}
}

// --- the hand-size discard ---------------------------------------------------

func TestTheHandSizeDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	seedOrderProbes(g, p.ID, graveyardOrderProbeOracles, "Graveyard Order Probe")
	turn := g.Turn.Seq
	needle := cleanupDiscard(t, g, 0, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	s11cOnlyPrompt(t, g, game.PendingChoiceReplacementOrder)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDiscard)
	p = seatByID(g, p.ID)
	if !p.Hand.Contains(needle) || g.DiscardPending[p.ID] != 1 {
		t.Fatalf("the restored table has the discard settled (pending %v)", g.DiscardPending)
	}
	answerExileOrder(t, g)
	if !p.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not discarded after the restored answer")
	}
	if len(g.DiscardPending) != 0 {
		t.Errorf("the hand-size debt is still owed: %v", g.DiscardPending)
	}
	if g.Turn.Seq == turn && g.Turn.Step == game.StepCleanup {
		t.Error("the cleanup step did not move on after the restored answer")
	}
}
