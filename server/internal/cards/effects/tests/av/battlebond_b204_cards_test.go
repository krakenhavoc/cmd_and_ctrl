package testsav

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battlebond_b204_cards_test.go — Battlebond, S82 older-sets batch 204
// (#5529): the partner-with creatures.

const (
	bbdB204SylviaOracle   = "22a93a6d-fe58-4135-8ac2-6b92e9bb75a3"
	bbdB204KhorvathOracle = "b96b1341-0c15-4816-be2d-5cc98b2afb63"
	bbdB204RetrieverOrcl  = "8ad2ecd5-2d50-4a54-a672-335d7e636a5a"
	bbdB204MentorOracle   = "9ad114d8-2123-4267-b30b-6f4e755bb0de"
	bbdB204ProtegeOracle  = "a03059dd-f499-407a-9346-6710b0468202"
	bbdB204GormOracle     = "a7195e9f-53fa-4fb5-b1f1-099c0a772090"
)

func TestBbdB204EveryCardIsRegisteredFullWithItsPartnerTrigger(t *testing.T) {
	for oracle, name := range map[string]string{
		bbdB204SylviaOracle:   "Sylvia Brightspear",
		bbdB204KhorvathOracle: "Khorvath Brightflame",
		bbdB204RetrieverOrcl:  "Chakram Retriever",
		bbdB204MentorOracle:   "Proud Mentor",
		bbdB204ProtegeOracle:  "Impetuous Protege",
		bbdB204GormOracle:     "Gorm the Great",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
			continue
		}
		found := false
		for _, tr := range spec.Triggered {
			found = found || strings.Contains(tr.Key, effects.PartnerWithKeyMark)
		}
		if !found {
			t.Errorf("%s has no partner-with trigger", name)
		}
	}
}

func TestBbdB204SylviaGivesYourDragonsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sylvia := effectstest.B12Push(g, me.ID, "Sylvia Brightspear", "Legendary Creature — Human Knight", bbdB204SylviaOracle, 2, 2)
	dragon := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Dragon", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Dragon", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]bool{sylvia: true, dragon: true, theirs: false, bear: false} {
		if got := effectstest.HasKeywordOnBattlefield(t, g, id, "double strike"); got != want {
			t.Errorf("double strike = %v, want %v", got, want)
		}
	}
}

func TestBbdB204KhorvathGivesYourKnightsFlyingAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	khorvath := effectstest.B12Push(g, me.ID, "Khorvath Brightflame", "Legendary Creature — Dragon", bbdB204KhorvathOracle, 3, 4)
	knight := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Human Knight", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human Knight", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, kw := range []string{"flying", "haste"} {
		for id, want := range map[uuid.UUID]bool{khorvath: true, knight: true, theirs: false, bear: false} {
			if got := effectstest.HasKeywordOnBattlefield(t, g, id, kw); got != want {
				t.Errorf("%s = %v, want %v", kw, got, want)
			}
		}
	}
}

func TestBbdB204ChakramRetrieverUntapsWhenYouCastOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	retriever := effectstest.B12Push(g, me.ID, "Chakram Retriever", "Creature — Elemental Dog", bbdB204RetrieverOrcl, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastFromHandForTest(t, g, me, "Test Spell", "Instant", "{0}", "", game.CastSpellParams{Strict: true})
	effectstest.AvB34Pick(t, g, me.ID, bear)
	if effectstest.TriggersOnStackFrom(g, retriever) != 1 {
		t.Fatalf("no cast trigger on the stack: %+v", g.PendingChoices)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindCardAnywhere(t, g, bear).Tapped {
		t.Error("the Bear stayed tapped after you cast a spell on your turn")
	}
}

func TestBbdB204ChakramRetrieverIgnoresSpellsOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	retriever := effectstest.B12Push(g, me.ID, "Chakram Retriever", "Creature — Elemental Dog", bbdB204RetrieverOrcl, 2, 4)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Test Spell", "Instant", "{0}", "", game.CastSpellParams{Strict: true})
	if n := effectstest.TriggersOnStackFrom(g, retriever); n != 0 {
		t.Errorf("%d Retriever triggers on the opponent's turn, want 0", n)
	}
}

func TestBbdB204ProudMentorTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	mentor := effectstest.B12Push(g, me.ID, "Proud Mentor", "Creature — Human Warrior", bbdB204MentorOracle, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, mentor, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindCardAnywhere(t, g, theirs).Tapped {
		t.Error("the target was not tapped")
	}
	if !effectstest.FindCardAnywhere(t, g, mentor).Tapped {
		t.Error("the Mentor's own {T} cost was not paid")
	}
}

func TestBbdB204ImpetuousProtegeGrowsByTheGreatestTappedOpposingPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	protege := effectstest.B12Push(g, me.ID, "Impetuous Protege", "Creature — Human Warrior", bbdB204ProtegeOracle, 0, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	untappedBigger := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Giant", 8, 8)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Giant", 9, 9)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{small, big, mine} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	_ = untappedBigger
	effectstest.DeclareAttack(t, g, opp.ID, protege)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, protege); p != 5 || tough != 4 {
		t.Errorf("the Protege is %d/%d, want 5/4 (the tapped 5-power creature; untapped and its own side don't count)", p, tough)
	}
}

func TestBbdB204GormMustBeBlockedByTwoOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorm := effectstest.B12Push(g, me.ID, "Gorm the Great", "Legendary Creature — Giant Warrior", bbdB204GormOracle, 2, 7)
	w1 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	w2 := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, gorm)
	if !effectstest.HasKeywordOnBattlefield(t, g, gorm, "vigilance") {
		t.Error("Gorm has no vigilance")
	}
	if err := g.DeclareBlocker(w1, gorm); err != nil {
		t.Fatalf("one blocker as a step: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("one blocker was accepted with two creatures able to block")
	}
	if err := g.DeclareBlocker(w2, gorm); err != nil {
		t.Fatalf("the second blocker: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Errorf("two blockers refused: %v", err)
	}
}
