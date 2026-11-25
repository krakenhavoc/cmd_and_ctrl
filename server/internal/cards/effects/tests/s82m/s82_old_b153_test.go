package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b153_test.go — Commander 2019, S82 set OLD batch 153 (#5517).
// Every test plays the real catalog entry through the engine.

const (
	oldB153GargoyleCastle = "963f2848-15bd-441b-a55c-635f53b7b63f"
	oldB153Skywise        = "de80c742-4da0-4d67-9711-2b4e6452e2d0"
	oldB153Faith          = "99af5247-06cf-40b2-9a7c-65ad59b1ddf6"
	oldB153Whispers       = "b7ff27b6-7591-44b9-86b9-ee12f795ec25"
	oldB153SecretPlans    = "7d0d7480-fddf-4841-bc2a-239e03f97f60"
	oldB153Kadena         = "4da13eb8-ca18-43f1-b54b-d3ddceb980d2"
	oldB153Craghorn       = "a084bca4-fe72-443e-adf4-ab5456c7e699"
	oldB153Vraska         = "0d2bbce0-8c27-4b82-ab62-51a0c3cfc9b0"
	oldB153Ghired         = "d543cc69-4b8a-4c45-99e1-84143d663f25"
)

func TestOldB153CardsAreRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB153GargoyleCastle: "Gargoyle Castle",
		oldB153Skywise:        "Pristine Skywise",
		oldB153Faith:          "Faith of the Devoted",
		oldB153Whispers:       "Malevolent Whispers",
		oldB153SecretPlans:    "Secret Plans",
		oldB153Kadena:         "Kadena, Slinking Sorcerer",
		oldB153Vraska:         "Vraska the Unseen",
		oldB153Ghired:         "Ghired's Belligerence",
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

func TestOldB153GargoyleCastleMakesAFlyingGargoyle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Gargoyle Castle", "Land", oldB153GargoyleCastle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the {5} ability for four mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.S46Tokens(g, me.ID, "Gargoyle")
	if len(toks) != 1 {
		t.Fatalf("%d Gargoyle tokens, want 1", len(toks))
	}
	c := toks[0]
	if c.Power != 3 || c.Toughness != 4 || !c.IsArtifact() || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
		t.Errorf("token %+v is not a 3/4 flying artifact creature", c)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the Gargoyle Castle was not sacrificed")
	}
}

func TestOldB153SkywiseUntapsAndGainsProtectionOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sky := effectstest.PushCatalogPermanent(g, me.ID, "Pristine Skywise", "Creature — Dragon", oldB153Skywise, false)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, sky).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "U")
	c, _ := effectstest.BattlefieldCard(g, sky)
	if c.Tapped {
		t.Error("the Skywise stayed tapped")
	}
	if !game.HasKeyword(&c, game.ProtectionFromColor("U")) {
		t.Error("the Skywise did not gain protection from blue")
	}
}

func TestOldB153SkywiseIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sky := effectstest.PushCatalogPermanent(g, me.ID, "Pristine Skywise", "Creature — Dragon", oldB153Skywise, false)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, sky).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if effectstest.PendingOfKind(g, game.PendingChoiceColor) != nil {
		t.Error("a creature spell opened a colour choice")
	}
	if c, _ := effectstest.BattlefieldCard(g, sky); !c.Tapped {
		t.Error("a creature spell untapped the Skywise")
	}
}

func TestOldB153FaithDrainsOnADiscardWhenYouPayOne(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Faith of the Devoted", "Enchantment", oldB153Faith, false)
		effectstest.S58HandCards(me, 2)
		life, oppLife := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FloatMana(t, g, me, "{C}")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		wantMe, wantOpp := life, oppLife
		if pay {
			wantMe, wantOpp = life+2, oppLife-2
		}
		if me.Life != wantMe || opp.Life != wantOpp {
			t.Errorf("pay=%v: life %d/%d, want %d/%d", pay, me.Life, opp.Life, wantMe, wantOpp)
		}
	}
}

func TestOldB153FaithIgnoresAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Faith of the Devoted", "Enchantment", oldB153Faith, false)
	effectstest.S58HandCards(opp, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's discard offered the {1}")
	}
}

func TestOldB153WhispersStealsUntapsPumpsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Malevolent Whispers", TypeLine: "Sorcery", ManaCost: "{3}{R}", OracleID: oldB153Whispers, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %v tapped %v, want me and untapped", c.Controller, c.Tapped)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("no haste")
	}
	if p := effectstest.CurrentPower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

// oldB153MorphInHand puts a Battering Craghorn (morph {1}{R}{R}) in the
// active player's hand.
func oldB153MorphInHand(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Battering Craghorn", TypeLine: "Creature — Goat Beast", OracleID: oldB153Craghorn,
		ManaCost: "{1}{R}{R}", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	return id
}

func TestOldB153SecretPlansPumpsFaceDownCreaturesAndDrawsOnTurningUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Secret Plans", "Enchantment", oldB153SecretPlans, false)
	id := effectstest.S36CastFaceDown(t, g, "Battering Craghorn", "Creature — Goat Beast", oldB153Craghorn, "morph", 3, 1)
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 2 || tough != 3 {
		t.Errorf("face-down creature is %d/%d, want 2/3", p, tough)
	}
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.MkmB1PT(t, g, other); p != 2 || tough != 2 {
		t.Errorf("a face-up creature is %d/%d, want 2/2", p, tough)
	}
	hand := me.Hand.Size()
	effectstest.S36TurnUp(t, g, me, id, "R", "R", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for turning a permanent face up", hand, me.Hand.Size())
	}
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 3 || tough != 1 {
		t.Errorf("face-up creature is %d/%d, want its printed 3/1", p, tough)
	}
}

func TestOldB153KadenaDiscountsOnlyTheFirstFaceDownCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Kadena, Slinking Sorcerer", "Legendary Creature — Snake Wizard Sorcerer", oldB153Kadena, false)
	first := oldB153MorphInHand(me)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err != nil {
		t.Fatalf("the first face-down spell should cost nothing: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The entering face-down creature drew a card, and the cast took one out of hand.
	if got := me.Hand.Size(); got != hand-1+1 {
		t.Errorf("hand %d → %d, want the cast and one draw", hand, got)
	}
	second := oldB153MorphInHand(me)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err == nil {
		t.Error("the second face-down spell cast for no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err != nil {
		t.Errorf("the second face-down spell for the full {3}: %v", err)
	}
}

func TestOldB153KadenaDiscountIsNotForAFaceUpSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Kadena, Slinking Sorcerer", "Legendary Creature — Snake Wizard Sorcerer", oldB153Kadena, false)
	id := oldB153MorphInHand(me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a face-up creature spell was discounted")
	}
}

func TestOldB153VraskaPlusOneDestroysCreaturesThatDamageHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska the Unseen", oldB153Vraska, 5)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, vraska); got != 6 {
		t.Fatalf("loyalty %d, want 6", got)
	}
	// Damage to the player, and non-combat damage to Vraska, change nothing.
	effectstest.DealCombatDamageToPlayer(g, bystander, me.ID, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bystander, Target: vraska, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bystander); !ok {
		t.Fatal("the creature was destroyed without dealing combat damage to Vraska")
	}
	effectstest.DealCombatDamageToPlayer(g, attacker, vraska, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, attacker); ok {
		t.Error("the creature that dealt combat damage to Vraska was not destroyed")
	}
	if _, ok := effectstest.BattlefieldCard(g, bystander); !ok {
		t.Error("another creature was destroyed")
	}
}

func TestOldB153VraskaMinusThreeDestroysANonlandPermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska the Unseen", oldB153Vraska, 5)
	rock := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, vraska, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)})
	if err == nil {
		t.Fatal("targeted a land")
	}
	effectstest.S82Activate(t, g, me.ID, vraska, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB153VraskaUltimateMakesThreeAssassinsThatEndTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska the Unseen", oldB153Vraska, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.S46Tokens(g, me.ID, "Assassin")
	if len(toks) != 3 {
		t.Fatalf("%d Assassins, want 3", len(toks))
	}
	effectstest.DealCombatDamageToPlayer(g, toks[0].InstanceID, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the damaged player did not lose the game")
	}
}

func TestOldB153GhiredsBelligerencePopulatesOncePerCreatureThatDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}})
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Wall", 0, 30)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}{C}{C}{C}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ghired's Belligerence", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}", OracleID: oldB153Ghired, Owner: me.ID, Controller: me.ID})
	})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, XValue: 4, Targets: effectstest.CardRefs(a, b, sturdy),
		Distribution: map[uuid.UUID]int{a: 2, b: 1, sturdy: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, a); ok {
		t.Fatal("A survived 2 damage")
	}
	if _, ok := effectstest.BattlefieldCard(g, b); !ok {
		t.Fatal("B died to 1 damage")
	}
	if got := len(effectstest.S46Tokens(g, me.ID, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens after one death, want 2 (the original and one populated copy)", got)
	}
	// B, dealt damage this way, dies later this turn: populate again.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, b).DamageMarked = 5 })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Soldier")); got != 3 {
		t.Errorf("%d Soldier tokens after the second death, want 3", got)
	}
}

func TestOldB153GhiredsBelligerenceIgnoresACreatureNotDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}})
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ghired's Belligerence", TypeLine: "Sorcery", ManaCost: "{X}{R}{R}", OracleID: oldB153Ghired, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardRefs(a), Distribution: map[uuid.UUID]int{a: 1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bystander).DamageMarked = 5 })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("%d Soldier tokens, want 1: a creature that was not dealt damage this way populated", got)
	}
}
