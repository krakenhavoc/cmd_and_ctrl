package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The creature is exiled, its controller reveals down to a creature
// card and puts it onto the battlefield under their control, and the
// rest is shuffled back in.
func TestReboundSpellBlessedReincarnationReplacesTheCreatureFromItsControllersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	land := effectstest.CardNamed("Revealed Land", "Land")
	beast := effectstest.CardNamed("Reincarnated Beast", "Creature — Beast")
	beast.Power, beast.Toughness = 3, 3
	effectstest.LibraryTop(opp, beast, land)
	size := opp.Library.Size()

	me, id := effectstest.CastReboundSpellAtMain(t, g, "Blessed Reincarnation", "Instant", effectstest.BlessedReincarnationOracle,
		game.TargetRef{Kind: game.TargetCard, ID: victim})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(victim) {
		t.Error("the targeted creature was not exiled")
	}
	got := effectstest.FindBattlefieldByName(g, "Reincarnated Beast")
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: big, Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	me, _ := effectstest.CastReboundSpellAtMain(t, g, "Consuming Vapors", "Sorcery", effectstest.ConsumingVaporsOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	// Only one creature: the engine may take it without asking.
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		effectstest.AnswerSacrifice(t, g, opp.ID, big)
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
	g := effectstest.NewCatalogGame(t)
	var hits []uuid.UUID
	for i := 1; i < len(g.Seats); i++ {
		opp := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		spell := effectstest.CardNamed("Stolen Sorcery", "Sorcery")
		spell.ManaCost = "{5}"
		effectstest.LibraryTop(opp, spell, effectstest.CardNamed("Exiled Land", "Land"))
		hits = append(hits, spell.InstanceID)
	}
	me, _ := effectstest.CastReboundSpellAtMain(t, g, "Fevered Suspicion", "Sorcery", effectstest.FeveredSuspicionOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range hits {
		if !g.Exile.Contains(id) {
			t.Fatalf("an opponent's nonland card was not exiled")
		}
	}
	// Cast the first one for free, during the resolution (#4250), then
	// decline the rest.
	me.ManaPool = nil
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, hits...)
	if err := g.CastSpell(me.ID, hits[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("free cast of an exiled card: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("one cast closed the \"any number\" offer")
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.CastInsideTheResolution(t, g, hits[0])
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, tl := range []string{"Instant", "Sorcery", "Artifact", "Creature — Bear", "Land"} {
		c := effectstest.CardNamed("Graveyard "+tl, tl)
		c.Owner = me.ID
		me.Graveyard.PushTop(c)
	}
	target := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Flamer")
	effectstest.CastReboundSpellAtMain(t, g, "Flame On!", "Sorcery", effectstest.FlameOnOracle,
		game.TargetRef{Kind: game.TargetCard, ID: target})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "One")
	effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Two")
	effectstest.CastReboundSpellAtMain(t, g, "Nomads' Assembly", "Sorcery", effectstest.NomadsAssemblyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Kor Soldier"); n != 2 {
		t.Errorf("Kor Soldiers %d, want 2", n)
	}
}

// The copy of a legendary creature is not legendary.
func TestReboundSpellQuantumMisalignmentCopyIsNotLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.CastReboundSpellAtMain(t, g, "Quantum Misalignment", "Sorcery", effectstest.QuantumMisalignmentOracle,
		game.TargetRef{Kind: game.TargetCard, ID: id})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Legend"); n != 2 {
		t.Fatalf("Legends on the battlefield %d, want 2 (the legend rule must not apply)", n)
	}
}

// Draws as many as the opponent holds.
func TestReboundSpellRecurringInsightDrawsTheOpponentsHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, _ := effectstest.CastReboundSpellAtMain(t, g, "Recurring Insight", "Sorcery", effectstest.RecurringInsightOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	want := me.Hand.Size() + opp.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != want {
		t.Errorf("hand %d, want %d", me.Hand.Size(), want)
	}
}

// Only an instant card comes back.
func TestReboundSpellSurrealMemoirReturnsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	instant := effectstest.CardNamed("Old Instant", "Instant")
	instant.Owner = me.ID
	sorcery := effectstest.CardNamed("Old Sorcery", "Sorcery")
	sorcery.Owner = me.ID
	me.Graveyard.PushTop(instant)
	me.Graveyard.PushTop(sorcery)
	effectstest.CastReboundSpellAtMain(t, g, "Surreal Memoir", "Sorcery", effectstest.SurrealMemoirOracle)
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			for _, p := range g.Seats {
				if p.ID != g.Seats[g.Turn.ActiveSeat].ID {
					p.Life = tc.oppLife
				}
			}
			me, _ := effectstest.CastReboundSpellAtMain(t, g, "Survival Cache", "Sorcery", effectstest.SurvivalCacheOracle)
			me.Life = 20
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, id := effectstest.CastReboundSpellAtMain(t, g, "Transpose", "Instant", effectstest.TransposeOracle)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if n := effectstest.OnBattlefieldNamed(g, "Wizard"); n != 1 {
		t.Fatalf("Wizards after the hand cast %d, want 1", n)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("Transpose was not exiled by rebound")
	}

	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("no rebound offer")
	}
	// #4250: the cast below answers the resolution_cast prompt.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("rebound cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if n := effectstest.OnBattlefieldNamed(g, "Wizard"); n != 1 {
		t.Errorf("Wizards after the rebound cast %d, want still 1", n)
	}
}

// You gain hexproof until your next turn, and 2 life.
func TestReboundSpellBlossomingCalmGivesYouHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CastReboundSpellAtMain(t, g, "Blossoming Calm", "Instant", effectstest.BlossomingCalmOracle)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if !effectstest.PlayerHasHexproof(g, me) {
		t.Error("you did not gain hexproof")
	}
}
