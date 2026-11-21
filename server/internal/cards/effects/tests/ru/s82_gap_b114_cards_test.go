package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b114_cards_test.go — S82 GAP batch 114 (Commander Masters, #5948).
// Every test plays the real catalog card through the engine.

const (
	gapB114DaysUndoing  = "810a6cf5-e9f4-48e1-8e06-3e0af85633ce"
	gapB114MindsAglow   = "b259f4e4-04a9-40e0-85e9-34b4b74ee388"
	gapB114Tutor        = "90f54959-2c9b-4b8a-84c9-d6893eb43553"
	gapB114Gatewatch    = "e17ea7eb-362e-4c61-97cd-537f3921bced"
	gapB114HarshMercy   = "dcd65c9b-4aaa-42af-869f-de179ae57c9f"
	gapB114WandStrike   = "d6239ada-c72a-49b2-882c-6dae1dd366e3"
	gapB114PlaneswalkTy = "Legendary Planeswalker — Test"
)

func TestGapB114AllAreFull(t *testing.T) {
	for o, name := range map[string]string{
		gapB114DaysUndoing: "Day's Undoing", gapB114MindsAglow: "Minds Aglow",
		gapB114Tutor: "Personal Tutor", gapB114Gatewatch: "Deploy the Gatewatch",
		gapB114HarshMercy: "Harsh Mercy", gapB114WandStrike: "Wanderer's Strike",
	} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB114DaysUndoingWheelsTheTableAndEndsYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.SeedGraveyardCard(t, g, "My Fodder", "Instant", "")
	theirs := effectstest.HandCard(opp, "Their Fodder", "Instant")
	effectstest.CastCatalogSpell(t, g, "Day's Undoing", "Sorcery", gapB114DaysUndoing, nil)
	turn := g.Turn.Seq
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("my hand is %d, want 7", got)
	}
	if got := opp.Hand.Size(); got != 7 {
		t.Errorf("their hand is %d, want 7", got)
	}
	if me.Graveyard.Contains(mine) {
		t.Error("my graveyard card was not shuffled into my library")
	}
	_ = theirs
	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat == me.Seat {
		t.Errorf("the turn did not end: turn %d seat %d", g.Turn.Seq, g.Turn.ActiveSeat)
	}
}

// Resolved for a player who is not the active player ("if it's your turn" is
// false), the spell only wheels the table.
func TestGapB114DaysUndoingOnAnOpponentsTurnDoesNotEndTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(gapB114DaysUndoing)
	turn := g.Turn.Seq
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: opp.ID, SourceCardID: uuid.New()}
		if err := spec.OnResolve(item, effects.NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Turn.Seq != turn || g.Turn.ActiveSeat != active.Seat {
		t.Errorf("the turn ended: turn %d seat %d", g.Turn.Seq, g.Turn.ActiveSeat)
	}
	if got := opp.Hand.Size(); got != 7 {
		t.Errorf("the caster's hand is %d, want 7", got)
	}
}

func TestGapB114WanderersStrikeExilesThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, mine).Counters = map[string]int{"+1/+1": 1} })
	effectstest.CastCatalogSpell(t, g, "Wanderer's Strike", "Sorcery", gapB114WandStrike, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target was not exiled")
	}
	if !g.Exile.Contains(victim) {
		t.Error("the target is not in exile")
	}
	if got := effectstest.PlusOneCounters(g, mine); got != 2 {
		t.Errorf("%d +1/+1 counters on my Bear, want 2 (proliferated)", got)
	}
}

func TestGapB114WanderersStrikeNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Wanderer's Strike", "Sorcery", gapB114WandStrike, effectstest.CardRefs(land)); err == nil {
		t.Error("targeted a land")
	}
}

func TestGapB114PersonalTutorFindsOnlyASorceryAndPutsItOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wrath", TypeLine: "Sorcery"})
	top := effectstest.TutorTopOfLibrary(t, g, me, "Personal Tutor", "Sorcery", gapB114Tutor, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored sorcery", top.Name)
	}
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card")
	}
	if me.Hand.Contains(needle) {
		t.Error("the tutored card went to hand")
	}
}

func TestGapB114DeployTheGatewatchPutsUpToTwoPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.Mm2Library(me,
		game.Card{Name: "Walker A", TypeLine: gapB114PlaneswalkTy, Counters: map[string]int{game.CounterLoyalty: 4}},
		effectstest.Mm2Spell("S1"),
		game.Card{Name: "Walker B", TypeLine: gapB114PlaneswalkTy, Counters: map[string]int{game.CounterLoyalty: 3}},
		game.Card{Name: "Walker C", TypeLine: gapB114PlaneswalkTy, Counters: map[string]int{game.CounterLoyalty: 3}},
		effectstest.Mm2CreatureCard("Bear", 2), effectstest.Mm2Spell("S2"), effectstest.Mm2Spell("S3"), effectstest.Mm2Spell("Deep"))
	effectstest.CastCatalogSpell(t, g, "Deploy the Gatewatch", "Sorcery", gapB114Gatewatch, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("%d candidates, want the three planeswalkers", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2], ids[3]}); err == nil {
		t.Fatal("three planeswalkers were accepted: up to two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:1] {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s is not on the battlefield", id)
		}
	}
	if !g.Battlefield.Contains(ids[2]) {
		t.Error("the second planeswalker is not on the battlefield")
	}
	for _, id := range []uuid.UUID{ids[1], ids[3], ids[4], ids[5], ids[6], ids[7]} {
		if z := effectstest.Mm2Zone(g, id); z != game.ZoneLibrary {
			t.Errorf("card %s in %v, want the library (the rest go to the bottom)", id, z)
		}
	}
	if n := me.Library.Size(); n != 6 {
		t.Errorf("library %d, want 6", n)
	}
}

func TestGapB114HarshMercyDestroysEveryCreatureNotOfAChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	elf := effectstest.B12Creature(g, s[0].ID, "Elf", "Creature — Elf Warrior", 2, 2)
	goblin := effectstest.B12Creature(g, s[1].ID, "Goblin", "Creature — Goblin", 2, 2)
	human := effectstest.B12Creature(g, s[2].ID, "Human", "Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, s[3].ID, "Bear", "Creature — Bear", 2, 2)
	shapeshifter := effectstest.B12Creature(g, s[3].ID, "Shapeshifter", "Creature — Shapeshifter", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, shapeshifter).Keywords = []string{"changeling"} })
	effectstest.CastAndPause(t, g, "Harsh Mercy", "Sorcery", gapB114HarshMercy)
	for i, tribe := range []string{"Elf", "Goblin", "Sliver"} {
		c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
		if c == nil || c.Chooser != s[i].ID {
			t.Fatalf("prompt %d: want one owed by seat %d, got %+v", i, i, c)
		}
		if !g.Battlefield.Contains(bear) {
			t.Fatal("creatures were destroyed before the last seat answered")
		}
		effectstest.AnswerCreatureType(t, g, s[i].ID, tribe)
	}
	effectstest.AnswerCreatureType(t, g, s[3].ID, "Human")
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"Elf": {elf, true}, "Goblin": {goblin, true}, "Human": {human, true}, "Bear": {bear, false}, "Shapeshifter": {shapeshifter, true}} {
		if got := g.Battlefield.Contains(tc.id); got != tc.want {
			t.Errorf("%s on the battlefield = %v, want %v", name, got, tc.want)
		}
	}
}

// gapB114Answer drives every open join-forces prompt: each seat's number prompt
// is answered with its entry in amounts (0 when absent) and every payment is made.
func gapB114Answer(t *testing.T, g *game.Game, amounts map[uuid.UUID]int) {
	t.Helper()
	for i := 0; i < 24; i++ {
		var next *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoicePayAmount || c.Kind == game.PendingChoicePayUnless) {
				next = c
				break
			}
		}
		if next == nil {
			return
		}
		if next.Kind == game.PendingChoicePayAmount {
			if err := g.ResolvePayAmount(next.ID, next.Chooser, amounts[next.Chooser]); err != nil {
				t.Fatalf("answer number for %s: %v", next.Chooser, err)
			}
		} else if err := g.ResolvePayUnless(next.ID, next.Chooser, true); err != nil {
			t.Fatalf("pay for %s: %v", next.Chooser, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("prompts never stopped")
}

func TestGapB114MindsAglowEveryoneDrawsTheTotalPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.StockLibraryFor(g, s[0], 20)
	effectstest.StockLibraryFor(g, s[1], 20)
	effectstest.StockLibraryFor(g, s[2], 20)
	effectstest.StockLibraryFor(g, s[3], 20)
	hands := make([]int, 4)
	effectstest.CastCatalogSpell(t, g, "Minds Aglow", "Sorcery", gapB114MindsAglow, nil)
	effectstest.FillPool(s[0], 2)
	effectstest.FillPool(s[1], 1)
	effectstest.FillPool(s[2], 4)
	for i := range s {
		hands[i] = s[i].Hand.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Seat 2 keeps its mana; seats 0 and 1 pay 2 and 1; seat 3 has none.
	gapB114Answer(t, g, map[uuid.UUID]int{s[0].ID: 2, s[1].ID: 1})
	for i := range s {
		if got := s[i].Hand.Size() - hands[i]; got != 3 {
			t.Errorf("seat %d drew %d, want 3", i, got)
		}
	}
	if n := s[0].ManaPool.Total(); n != 0 {
		t.Errorf("seat 0 has %d mana left, want 0", n)
	}
	if n := s[2].ManaPool.Total(); n != 4 {
		t.Errorf("seat 2 has %d mana left, want its 4 kept", n)
	}
}

func TestGapB114MindsAglowWithNothingPaidDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.CastCatalogSpell(t, g, "Minds Aglow", "Sorcery", gapB114MindsAglow, nil)
	effectstest.FillPool(s[0], 3)
	hand := s[1].Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	gapB114Answer(t, g, nil)
	if got := s[1].Hand.Size(); got != hand {
		t.Errorf("seat 1 drew %d with nothing paid", got-hand)
	}
}
