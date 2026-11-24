package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 5 group C — Karn, Scion of Urza and Nicol Bolas, the Ravager
// // Nicol Bolas, the Arisen (Secret Lair Drop: Planeswalkers, #5462).

const (
	s82OldB005KarnOracle  = "060209fa-4b7d-4e08-bf8b-8759bba5641b"
	s82OldB005BolasOracle = "55e4b27e-5447-4fc2-8cae-a03e344600c6"
)

func TestS82OldB005CCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB005KarnOracle:         "Karn, Scion of Urza",
		s82OldB005BolasOracle:        "Nicol Bolas, the Ravager",
		s82OldB005BolasOracle + "#1": "Nicol Bolas, the Arisen",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func s82OldB005KarnGame(t *testing.T, loyalty int) (*game.Game, *game.Player, uuid.UUID) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	karn := effectstest.S82Walker(g, me.ID, "Karn, Scion of Urza", s82OldB005KarnOracle, loyalty)
	effectstest.AdvanceToMain(t, g)
	return g, me, karn
}

func TestS82OldB005CKarnPlusOneOpponentChoosesTheCardForYourHand(t *testing.T) {
	g, me, karn := s82OldB005KarnGame(t, 5)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	second := effectstest.SeedLibraryTop(me, "Second", "Sorcery", "")
	first := effectstest.SeedLibraryTop(me, "First", "Instant", "")
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, karn, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("the opponent should choose between the two revealed cards: %+v", pick)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) || me.Hand.Contains(first) || me.Hand.Size() != hand+1 {
		t.Errorf("hand has Second=%v First=%v size %d, want only Second (+1)", me.Hand.Contains(second), me.Hand.Contains(first), me.Hand.Size())
	}
	if n := effectstest.CounterCountInExile(g, first, "silver"); n != 1 {
		t.Errorf("First has %d silver counters in exile, want 1", n)
	}
	if got := effectstest.LoyaltyOf(g, karn); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82OldB005CKarnMinusOneTakesACardWithASilverCounter(t *testing.T) {
	g, me, karn := s82OldB005KarnGame(t, 5)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	silver, plain, theirs := uuid.New(), uuid.New(), uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: silver, Name: "Silver", TypeLine: "Instant", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"silver": 1}})
	g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"silver": 1}})
	effectstest.S82Activate(t, g, me.ID, karn, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick == nil || len(pick.ChooseCards) != 1 {
		t.Fatalf("prompt %+v, want exactly the one silver card I own", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, silver)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(silver) || me.Hand.Contains(plain) || opp.Hand.Contains(theirs) {
		t.Error("only my silver-counter card goes to my hand")
	}
	if got := effectstest.LoyaltyOf(g, karn); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB005CKarnMinusTwoMakesAConstructThatCountsArtifacts(t *testing.T) {
	g, me, karn := s82OldB005KarnGame(t, 5)
	effectstest.S82Activate(t, g, me.ID, karn, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Construct" {
			tok = c
		}
	}
	if tok == nil {
		t.Fatal("no Construct token")
	}
	if tok.CurrentPower() != 1 || tok.CurrentToughness() != 1 {
		t.Errorf("Construct is %d/%d alone, want 1/1 (itself is an artifact)", tok.CurrentPower(), tok.CurrentToughness())
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	if c := effectstest.ApaLive(g, tok.InstanceID); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("Construct is %d/%d with a second artifact, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82OldB005CKarnMinusTwoIsRefusedAtOneLoyalty(t *testing.T) {
	g, me, karn := s82OldB005KarnGame(t, 1)
	err := g.ActivateCatalogAbility(me.ID, karn, 2, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		t.Fatal("a −2 with one loyalty counter was allowed")
	}
}

// --- Nicol Bolas, the Ravager // the Arisen -----------------------------

func s82OldB005BolasCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   s82OldB005BolasOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Nicol Bolas, the Ravager", TypeLine: "Legendary Creature — Elder Dragon", ManaCost: "{1}{U}{B}{R}",
				Colors: []string{"U", "B", "R"}, Power: 4, Toughness: 4},
			{Name: "Nicol Bolas, the Arisen", TypeLine: "Legendary Planeswalker — Bolas",
				Colors: []string{"U", "B", "R"}, StartingLoyalty: 7},
		},
	}
	c.SetFace(0)
	return c
}

func TestS82OldB005CBolasEntersAndEachOpponentDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	card := effectstest.PushHandCardOfType(opp, "Spare", "Instant")
	lt := effectstest.CastCreatureCard(t, g, "Nicol Bolas, the Ravager", "Legendary Creature — Elder Dragon", s82OldB005BolasOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(card) {
		t.Error("the opponent's only card was not discarded")
	}
	if !effectstest.HasEffectiveKeyword(t, g, lt, "flying") {
		t.Error("no flying")
	}
	_ = me
}

func s82OldB005PayTransform(me *game.Player) {
	for _, c := range []string{"U", "B", "R", "C", "C", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func TestS82OldB005CBolasExilesAndReturnsTransformedAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolas := effectstest.PushBattlefieldCardWithTimestamp(g, s82OldB005BolasCard(me.ID))
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	s82OldB005PayTransform(me)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepPostcombatMain)
	s82OldB005PayTransform(me)
	effectstest.S82Activate(t, g, me.ID, bolas, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var walker *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Nicol Bolas, the Arisen" {
			walker = c
		}
	}
	if walker == nil {
		t.Fatal("no Nicol Bolas, the Arisen on the battlefield")
	}
	if !walker.IsPlaneswalker() || walker.Counters[game.CounterLoyalty] != 7 || walker.Controller != me.ID {
		t.Errorf("walker %q loyalty %d controller %v, want a planeswalker with 7 loyalty under my control",
			walker.TypeLine, walker.Counters[game.CounterLoyalty], walker.Controller)
	}
}

func s82OldB005Arisen(t *testing.T, loyalty int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	w := effectstest.S82Walker(g, me.ID, "Nicol Bolas, the Arisen", s82OldB005BolasOracle+"#1", loyalty)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp, w
}

func TestS82OldB005CArisenPlusTwoDrawsTwo(t *testing.T) {
	g, me, _, w := s82OldB005Arisen(t, 7)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, w); got != 9 {
		t.Errorf("loyalty %d, want 9", got)
	}
}

func TestS82OldB005CArisenMinusThreeDealsTenToACreature(t *testing.T) {
	g, me, opp, w := s82OldB005Arisen(t, 7)
	big := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	effectstest.S82Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, big); c.DamageMarked != 10 {
		t.Errorf("Wall has %d damage, want 10", c.DamageMarked)
	}
	// A land is not a legal target.
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)}); err == nil {
		t.Error("a land was accepted as the target")
	}
}

func TestS82OldB005CArisenMinusFourReanimatesFromAnyGraveyard(t *testing.T) {
	g, me, opp, w := s82OldB005Arisen(t, 7)
	id := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, w, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(spell)}); err == nil {
		t.Error("a sorcery card was accepted as the target")
	}
	effectstest.S82Activate(t, g, me.ID, w, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(id)})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || c.Controller != me.ID {
		t.Fatalf("Dead Bear on the battlefield %v under %v, want mine", ok, c.Controller)
	}
	if got := effectstest.LoyaltyOf(g, w); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82OldB005CArisenMinusTwelveLeavesOnlyTheBottomCard(t *testing.T) {
	g, me, opp, w := s82OldB005Arisen(t, 12)
	effectstest.SeedLibraryTop(opp, "A", "Instant", "")
	effectstest.SeedLibraryTop(opp, "B", "Instant", "")
	bottom, err := opp.Library.Bottom()
	if err != nil {
		t.Fatal(err)
	}
	effectstest.S82Activate(t, g, me.ID, w, 3, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 1 {
		t.Fatalf("library has %d cards, want 1", opp.Library.Size())
	}
	if left, _ := opp.Library.Bottom(); left.InstanceID != bottom.InstanceID {
		t.Error("the card left is not the bottom card")
	}
}

func TestS82OldB005CArisenMinusThreeIsRefusedWithTooLittleLoyalty(t *testing.T) {
	g, me, opp, w := s82OldB005Arisen(t, 2)
	big := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(big)}); err == nil {
		t.Error("−3 with two loyalty counters was allowed")
	}
}
