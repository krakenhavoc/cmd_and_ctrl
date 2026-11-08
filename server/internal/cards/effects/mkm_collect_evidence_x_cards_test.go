package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_collect_evidence_x_cards_test.go — #5223: "whenever you collect
// evidence" (game.EventCollectEvidence) from every form that collects, and
// collect evidence X as an effect, through the Murders at Karlov Manor
// cards that print them: Evidence Examiner, Surveillance Monitor and
// Incinerator of the Guilty. The engine half is
// game/collect_evidence_event_test.go.

const (
	mkmCEXExaminerOracle    = "a3468804-26ad-4919-b030-a1906a7029e8"
	mkmCEXMonitorOracle     = "9be5a350-c38e-425f-9d6a-8eb111132a89"
	mkmCEXIncineratorOracle = "9b7d037f-3269-4e3b-9744-a977acb69c67"
)

func TestMkmCollectEvidenceXCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmCEXExaminerOracle:    "Evidence Examiner",
		mkmCEXMonitorOracle:     "Surveillance Monitor",
		mkmCEXIncineratorOracle: "Incinerator of the Guilty",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- helpers --------------------------------------------------------------

// evidenceEvents is every collection in the log.
func evidenceEvents(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventCollectEvidence {
			out = append(out, ev)
		}
	}
	return out
}

// wantOneCollection fails unless the log holds exactly one collection, by
// `who`, of exactly `cards` (in that order) worth `total`.
func wantOneCollection(t *testing.T, g *game.Game, who uuid.UUID, total int, cards ...uuid.UUID) {
	t.Helper()
	evs := evidenceEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d collect evidence events, want 1: %+v", len(evs), evs)
	}
	if ev := evs[0]; ev.Actor != who || ev.Amount != total || !slices.Equal(ev.CardIDs, cards) {
		t.Errorf("event actor %s, amount %d, cards %v; want %s, %d, %v", ev.Actor, ev.Amount, ev.CardIDs, who, total, cards)
	}
}

// incineratorHits puts Incinerator of the Guilty on me's battlefield, has it
// deal combat damage to opp and passes until its trigger has resolved (its
// first question is then open).
func incineratorHits(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	inc := pushCatalogPermanent(g, me.ID, "Incinerator of the Guilty", "Creature — Dragon", mkmCEXIncineratorOracle, false)
	dealCombatDamageToPlayer(g, inc, opp.ID, 6)
	passPriorityAroundTable(t, g)
	return inc
}

// --- the event, from each form --------------------------------------------

// The optional additional cost: one event, after the cards are exiled.
func TestCollectEvidenceEventFromTheOptionalCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gy := mkmCEEvidence(me, "{2}", "{4}", "{9}")
	if _, err := mkmCECast(t, g, "Crimestopper Sprite", "Creature — Faerie Detective", mkmCESpriteOracle,
		mkmCECollecting(nil, gy[1], gy[0])); err != nil {
		t.Fatalf("cast: %v", err)
	}
	wantOneCollection(t, g, me.ID, 6, gy[1], gy[0])
}

// The activated ability's cost: one event per activation.
func TestCollectEvidenceEventFromTheAbilityCost(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	researcher := pushCatalogPermanent(g, me.ID, "Forensic Researcher", "Creature — Merfolk Detective", mkmCEResearcherOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	gy := mkmCEEvidence(me, "{1}", "{3}")
	if err := g.ActivateCatalogAbility(me.ID, researcher, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{cardRef(victim)}, ExileIDs: []uuid.UUID{gy[1]}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	wantOneCollection(t, g, me.ID, 3, gy[1])
}

// The effect form: one event when evidence is collected, none when the
// offer is declined.
func TestCollectEvidenceEventFromTheEffect(t *testing.T) {
	t.Run("collected", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
		gy := mkmCEEvidence(me, "{1}", "{2}", "{3}")
		mkmCEAttack(t, g, izoni, opp.ID)
		answerChooseCards(t, g, me.ID, gy[2], gy[0])
		wantOneCollection(t, g, me.ID, 4, gy[2], gy[0])
	})
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
		mkmCEEvidence(me, "{5}")
		mkmCEAttack(t, g, izoni, opp.ID)
		answerChooseCards(t, g, me.ID)
		if n := len(evidenceEvents(g)); n != 0 {
			t.Errorf("%d events for a declined offer", n)
		}
	})
}

// --- Evidence Examiner ----------------------------------------------------

// At the beginning of combat it may collect evidence 4, and that collection
// is one it investigates for.
func TestEvidenceExaminerCollectsAtCombatAndInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Evidence Examiner", "Creature — Merfolk Detective", mkmCEXExaminerOracle, false)
	gy := mkmCEEvidence(me, "{1}", "{4}")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no collect evidence prompt at the beginning of combat")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, gy[:1]); err == nil {
		t.Fatal("collect evidence 4 with a total of 1 was accepted")
	}
	answerChooseCards(t, g, me.ID, gy[1])
	passPriorityAroundTable(t, g)
	if !inExile(g, gy[1]) || !me.Graveyard.Contains(gy[0]) {
		t.Error("only the named card is exiled")
	}
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

// "Whenever you collect evidence" sees another card's collection too (the
// card's ruling), and not an opponent's.
func TestEvidenceExaminerInvestigatesForAnyCollectionOfYours(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Evidence Examiner", "Creature — Merfolk Detective", mkmCEXExaminerOracle, false)
	pushCatalogPermanent(g, opp.ID, "Evidence Examiner", "Creature — Merfolk Detective", mkmCEXExaminerOracle, false)
	researcher := pushCatalogPermanent(g, me.ID, "Forensic Researcher", "Creature — Merfolk Detective", mkmCEResearcherOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	gy := mkmCEEvidence(me, "{3}")
	if err := g.ActivateCatalogAbility(me.ID, researcher, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{cardRef(victim)}, ExileIDs: gy}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues for me, want 1", n)
	}
	if n, _ := countTokensNamed(g, opp.ID, "Clue"); n != 0 {
		t.Errorf("the opponent's Examiner investigated %d times for my collection", n)
	}
}

// --- Surveillance Monitor -------------------------------------------------

// It enters, may collect evidence 4, and its own collection makes a Thopter;
// declining makes none.
func TestSurveillanceMonitorCollectsAsItEntersForAThopter(t *testing.T) {
	for _, collect := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		gy := mkmCEEvidence(me, "{2}", "{2}")
		if _, err := mkmCECast(t, g, "Surveillance Monitor", "Creature — Vedalken Detective", mkmCEXMonitorOracle, game.CastSpellParams{}); err != nil {
			t.Fatalf("collect=%v: cast: %v", collect, err)
		}
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		if collect {
			answerChooseCards(t, g, me.ID, gy...)
		} else {
			answerChooseCards(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		want := 0
		if collect {
			want = 1
		}
		if n, _ := countTokensNamed(g, me.ID, "Thopter"); n != want {
			t.Errorf("collect=%v: %d Thopters, want %d", collect, n, want)
		}
		if collect && (!inExile(g, gy[0]) || !inExile(g, gy[1])) {
			t.Error("the evidence was not exiled")
		}
	}
}

// --- Incinerator of the Guilty --------------------------------------------

// Collect evidence X: the cards are picked, then X up to their total; an X
// above it is refused, and "when you do" deals X to each creature that
// player controls and to nothing of anyone else's.
func TestIncineratorOfTheGuiltyCollectsXAndBurnsThatPlayersBoard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := pushVanillaCreature(g, opp.ID, "Their Three", 3, 3)
	big := pushVanillaCreature(g, opp.ID, "Their Five", 5, 5)
	mine := pushVanillaCreature(g, me.ID, "My Two", 2, 2)
	gy := mkmCEEvidence(me, "{2}", "{3}", "{1}")
	incineratorHits(t, g, me, opp)
	answerMayChoice(t, g, me.ID, true)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Fatalf("card prompt %+v, want any number of the three, none included", pick)
	}
	answerChooseCards(t, g, me.ID, gy[0], gy[1])
	x := numberPrompt(t, g, me.ID)
	if x.PayAmount.Min != 0 || x.PayAmount.Max != 5 || x.PayAmount.Goal != 5 {
		t.Fatalf("X prompt %+v, want 0 to 5 opening on 5", x.PayAmount)
	}
	if err := g.ResolvePayAmount(x.ID, me.ID, 6); err == nil {
		t.Fatal("X = 6 with cards worth 5 was accepted")
	}
	if !me.Graveyard.Contains(gy[0]) || !me.Graveyard.Contains(gy[1]) {
		t.Fatal("a refused X exiled the cards")
	}
	if err := g.ResolvePayAmount(x.ID, me.ID, 3); err != nil {
		t.Fatalf("X = 3: %v", err)
	}
	wantOneCollection(t, g, me.ID, 5, gy[0], gy[1])
	if !inExile(g, gy[0]) || !inExile(g, gy[1]) || !me.Graveyard.Contains(gy[2]) {
		t.Error("only the two chosen cards are exiled")
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, small) {
		t.Error("the 3/3 survived 3 damage")
	}
	if d := damageOn(t, g, big); d != 3 {
		t.Errorf("the 5/5 has %d damage, want 3", d)
	}
	if d := damageOn(t, g, mine); d != 0 {
		t.Errorf("my own creature took %d damage", d)
	}
}

// X = 0 is always possible: an empty graveyard still collects, with no
// card, and "whenever you collect evidence" sees it; declining collects
// nothing.
func TestIncineratorOfTheGuiltyCollectsZeroOrDeclines(t *testing.T) {
	t.Run("X = 0 on an empty graveyard", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		// Surveillance Monitor, not Evidence Examiner: the Examiner's own
		// combat trigger would join the stack as the turn moves on.
		pushCatalogPermanent(g, me.ID, "Surveillance Monitor", "Creature — Vedalken Detective", mkmCEXMonitorOracle, false)
		theirs := pushVanillaCreature(g, opp.ID, "Their One", 1, 1)
		incineratorHits(t, g, me, opp)
		answerMayChoice(t, g, me.ID, true)
		if c := chooseCardsChoiceFor(g, me.ID); c != nil {
			t.Fatalf("asked for cards from an empty graveyard: %+v", c)
		}
		if c := payAmountPrompt(g, me.ID); c != nil {
			t.Fatalf("asked for X with nothing to exile: %+v", c)
		}
		wantOneCollection(t, g, me.ID, 0)
		// The Monitor's trigger and the reflexive "when you do" go on the
		// stack together.
		passPriorityAroundTable(t, g)
		answerTriggerOrderLastQueuedFirst(t, g)
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		if n, _ := countTokensNamed(g, me.ID, "Thopter"); n != 1 {
			t.Errorf("%d Thopters for collect evidence 0, want 1", n)
		}
		if !onBattlefieldByID(g, theirs) || damageOn(t, g, theirs) != 0 {
			t.Error("X = 0 dealt damage")
		}
	})
	t.Run("X = 0 with cards chosen", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		gy := mkmCEEvidence(me, "{4}")
		incineratorHits(t, g, me, opp)
		answerMayChoice(t, g, me.ID, true)
		answerChooseCards(t, g, me.ID, gy...)
		x := numberPrompt(t, g, me.ID)
		if err := g.ResolvePayAmount(x.ID, me.ID, 0); err != nil {
			t.Fatalf("X = 0: %v", err)
		}
		wantOneCollection(t, g, me.ID, 4, gy...)
	})
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		gy := mkmCEEvidence(me, "{4}")
		incineratorHits(t, g, me, opp)
		answerMayChoice(t, g, me.ID, false)
		if n := len(evidenceEvents(g)); n != 0 {
			t.Errorf("%d events for a declined offer", n)
		}
		if !me.Graveyard.Contains(gy[0]) {
			t.Error("declining exiled the card")
		}
	})
}
