package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rebound_spells_test.go — the rebound cards whose own text is more
// than one primitive (#1854, ADR 0107 §3). Rebound itself is pinned in
// game/rebound_test.go and rebound_cards_test.go.

const (
	blessedReincarnationOracle = "ca29588c-f117-418e-be9e-fa2ee89862ca"
	blossomingCalmOracle       = "489a60f1-83f8-465b-918f-7d63d4f76d14"
	consumingVaporsOracle      = "a10ce333-48d7-499b-9355-f381f2395497"
	feveredSuspicionOracle     = "51545f6e-d0e3-4b4c-bc73-c731f85e26b0"
	flameOnOracle              = "e715d33e-3d60-4623-b843-db0b906ae98b"
	nomadsAssemblyOracle       = "a9a0088a-3c86-4baf-9754-79006f733ce1"
	quantumMisalignmentOracle  = "197582b9-4c86-41a3-ad0c-789a0db8e087"
	recurringInsightOracle     = "c6815bbb-24a5-4c9f-bf30-6190bc766d05"
	surrealMemoirOracle        = "fa482ded-b24c-4f9e-948e-d0620e907b8b"
	survivalCacheOracle        = "5fb8be5a-3666-4680-84e2-341cb269df07"
	transposeOracle            = "77d5f298-06b4-49d1-9b33-f1f176665ba2"
)

// castReboundSpellAtMain seeds a catalog card in the active seat's
// hand at its main phase and casts it.
func castReboundSpellAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me, id
}

// libraryTop pushes cards on top of a library, last argument on top.
func libraryTop(p *game.Player, cards ...game.Card) {
	for _, c := range cards {
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
	}
}

func cardNamed(name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	return c
}

// The creature is exiled, its controller reveals down to a creature
// card and puts it onto the battlefield under their control, and the
// rest is shuffled back in.
func TestReboundSpellBlessedReincarnationReplacesTheCreatureFromItsControllersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	land := cardNamed("Revealed Land", "Land")
	beast := cardNamed("Reincarnated Beast", "Creature — Beast")
	beast.Power, beast.Toughness = 3, 3
	libraryTop(opp, beast, land)
	size := opp.Library.Size()

	me, id := castReboundSpellAtMain(t, g, "Blessed Reincarnation", "Instant", blessedReincarnationOracle,
		game.TargetRef{Kind: game.TargetCard, ID: victim})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(victim) {
		t.Error("the targeted creature was not exiled")
	}
	got := findBattlefieldByName(g, "Reincarnated Beast")
	if got == uuid.Nil {
		t.Fatal("the revealed creature card did not enter")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == got && c.Controller != opp.ID {
			t.Errorf("the creature entered under %v, want its library's owner", c.Controller)
		}
	}
	if !opp.Library.Contains(land.InstanceID) || opp.Library.Size() != size-1 {
		t.Errorf("the revealed land was not shuffled back (library %d, want %d)", opp.Library.Size(), size-1)
	}
	if !g.Exile.Contains(id) {
		t.Error("Blessed Reincarnation was not exiled by rebound")
	}
	_ = me
}

// The opponent sacrifices, and you gain its toughness.
func TestReboundSpellConsumingVaporsGainsTheSacrificedCreaturesToughness(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: big, Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	me, _ := castReboundSpellAtMain(t, g, "Consuming Vapors", "Sorcery", consumingVaporsOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	life := me.Life
	passPriorityAroundTable(t, g)
	// Only one creature: the engine may take it without asking.
	if c := sacrificeChoiceFor(g, opp.ID); c != nil {
		answerSacrifice(t, g, opp.ID, big)
	}
	if !opp.Graveyard.Contains(big) {
		t.Fatal("the creature was not sacrificed")
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
}

// Each opponent exiles down to a nonland card; you may cast those for
// free right away, and a pass leaves the rest in exile.
func TestReboundSpellFeveredSuspicionOffersEachOpponentsNonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	var hits []uuid.UUID
	for i := 1; i < len(g.Seats); i++ {
		opp := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		spell := cardNamed("Stolen Sorcery", "Sorcery")
		spell.ManaCost = "{5}"
		libraryTop(opp, spell, cardNamed("Exiled Land", "Land"))
		hits = append(hits, spell.InstanceID)
	}
	me, _ := castReboundSpellAtMain(t, g, "Fevered Suspicion", "Sorcery", feveredSuspicionOracle)
	passPriorityAroundTable(t, g)
	for _, id := range hits {
		if !g.Exile.Contains(id) {
			t.Fatalf("an opponent's nonland card was not exiled")
		}
	}
	// Cast the first one for free, at sorcery speed or not.
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, hits[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("free cast of an exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(hits[0]) {
		t.Error("the cast card is still in exile")
	}
	// The window has closed: the others stay in exile, uncastable.
	if !g.Exile.Contains(hits[1]) {
		t.Fatal("an uncast card left exile")
	}
	if err := g.CastSpell(me.ID, hits[1], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("an uncast card was still castable after the window closed")
	}
}

// X counts noncreature, nonland cards in your graveyard; the creature
// also flies.
func TestReboundSpellFlameOnCountsNoncreatureNonlandCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, tl := range []string{"Instant", "Sorcery", "Artifact", "Creature — Bear", "Land"} {
		c := cardNamed("Graveyard "+tl, tl)
		c.Owner = me.ID
		me.Graveyard.PushTop(c)
	}
	target := pushCreatureToBattlefieldForTest(g, me.ID, "Flamer")
	castReboundSpellAtMain(t, g, "Flame On!", "Sorcery", flameOnOracle,
		game.TargetRef{Kind: game.TargetCard, ID: target})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != target {
			continue
		}
		if n := c.Counters[game.CounterPlusOne]; n != 3 {
			t.Errorf("+1/+1 counters %d, want 3", n)
		}
		if !game.HasKeyword(&c, "flying") {
			t.Error("the creature did not gain flying")
		}
	}
}

// One Kor Soldier per creature you control, counted once.
func TestReboundSpellNomadsAssemblyMakesOneSoldierPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCreatureToBattlefieldForTest(g, me.ID, "One")
	pushCreatureToBattlefieldForTest(g, me.ID, "Two")
	castReboundSpellAtMain(t, g, "Nomads' Assembly", "Sorcery", nomadsAssemblyOracle)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Kor Soldier"); n != 2 {
		t.Errorf("Kor Soldiers %d, want 2", n)
	}
}

// The copy of a legendary creature is not legendary.
func TestReboundSpellQuantumMisalignmentCopyIsNotLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castReboundSpellAtMain(t, g, "Quantum Misalignment", "Sorcery", quantumMisalignmentOracle,
		game.TargetRef{Kind: game.TargetCard, ID: id})
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Legend"); n != 2 {
		t.Fatalf("Legends on the battlefield %d, want 2 (the legend rule must not apply)", n)
	}
}

// Draws as many as the opponent holds.
func TestReboundSpellRecurringInsightDrawsTheOpponentsHandSize(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, _ := castReboundSpellAtMain(t, g, "Recurring Insight", "Sorcery", recurringInsightOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	want := me.Hand.Size() + opp.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != want {
		t.Errorf("hand %d, want %d", me.Hand.Size(), want)
	}
}

// Only an instant card comes back.
func TestReboundSpellSurrealMemoirReturnsAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	instant := cardNamed("Old Instant", "Instant")
	instant.Owner = me.ID
	sorcery := cardNamed("Old Sorcery", "Sorcery")
	sorcery.Owner = me.ID
	me.Graveyard.PushTop(instant)
	me.Graveyard.PushTop(sorcery)
	castReboundSpellAtMain(t, g, "Surreal Memoir", "Sorcery", surrealMemoirOracle)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(instant.InstanceID) {
		t.Error("the instant card did not return")
	}
	if !me.Graveyard.Contains(sorcery.InstanceID) {
		t.Error("a sorcery card left the graveyard")
	}
}

// The draw needs more life than at least one opponent, after the gain.
func TestReboundSpellSurvivalCacheDrawsOnlyAheadOfAnOpponent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		oppLife  int
		wantDraw bool
	}{
		{"ahead", 20, true},
		{"behind", 40, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			for _, p := range g.Seats {
				if p.ID != g.Seats[g.Turn.ActiveSeat].ID {
					p.Life = tc.oppLife
				}
			}
			me, _ := castReboundSpellAtMain(t, g, "Survival Cache", "Sorcery", survivalCacheOracle)
			me.Life = 20
			hand := me.Hand.Size()
			passPriorityAroundTable(t, g)
			if me.Life != 22 {
				t.Errorf("life %d, want 22", me.Life)
			}
			if got := me.Hand.Size() > hand; got != tc.wantDraw {
				t.Errorf("drew = %v, want %v", got, tc.wantDraw)
			}
		})
	}
}

// From hand: loot, lose 1, and a Wizard. The rebound cast from exile
// makes no Wizard.
func TestReboundSpellTransposeMakesAWizardOnlyWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, id := castReboundSpellAtMain(t, g, "Transpose", "Instant", transposeOracle)
	life := me.Life
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if n := onBattlefieldNamed(g, "Wizard"); n != 1 {
		t.Fatalf("Wizards after the hand cast %d, want 1", n)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("Transpose was not exiled by rebound")
	}

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("no rebound offer")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("rebound cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	if n := onBattlefieldNamed(g, "Wizard"); n != 1 {
		t.Errorf("Wizards after the rebound cast %d, want still 1", n)
	}
}

// You gain hexproof until your next turn, and 2 life.
func TestReboundSpellBlossomingCalmGivesYouHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := castReboundSpellAtMain(t, g, "Blossoming Calm", "Instant", blossomingCalmOracle)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if !playerHasHexproof(g, me) {
		t.Error("you did not gain hexproof")
	}
}
