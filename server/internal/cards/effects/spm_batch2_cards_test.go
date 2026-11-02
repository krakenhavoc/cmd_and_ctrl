package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_batch2_cards_test.go — Marvel's Spider-Man, batch 2 (#4460). Each
// test plays the real catalog entry and asserts what its text says.

const (
	spm2EddieOracle      = "fe01c96b-eb78-4477-bec2-944b2a28c778"
	spm2EzekielOracle    = "4074e897-82a4-4f57-86c2-4aab3f29ef9b"
	spm2FriendlyOracle   = "10fe402e-8d6f-4058-9f42-86f229055359"
	spm2GoblinOracle     = "67175889-02f9-4ad8-a9ab-11db44328b67"
	spm2GrowArmsOracle   = "52f03e73-3dba-44c2-b256-10e7e208b961"
	spm2GuyOracle        = "83eb74df-1c77-48b0-83be-55d7aa4d3618"
	spm2HideOracle       = "376fa374-ead1-4728-ab8b-aca230b72031"
	spm2HotDogCartOracle = "cf27f1f8-2d95-4a00-97f8-72a244837f08"
)

func TestSpm2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		damageControlCrewOracle: "Damage Control Crew",
		spm2EddieOracle:         "Eddie Brock",
		spm2EddieOracle + "#1":  "Venom, Lethal Protector",
		electrosBoltOracle:      "Electro's Bolt",
		spm2EzekielOracle:       "Ezekiel Sims, Spider-Totem",
		spm2FriendlyOracle:      "Friendly Neighborhood",
		spm2GoblinOracle:        "Green Goblin, Revenant",
		spm2GrowArmsOracle:      "Grow Extra Arms",
		spm2GuyOracle:           "Guy in the Chair",
		spm2HideOracle:          "Hide on the Ceiling",
		spm2HotDogCartOracle:    "Hot Dog Cart",
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

// --- Hot Dog Cart ---------------------------------------------------------

func TestSpm2HotDogCartMakesAFoodAndTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cart := castAndResolveCreature(t, g, "Hot Dog Cart", "Artifact", spm2HotDogCartOracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food tokens, want 1", n)
	}
	if err := g.ActivateManaAbility(me.ID, cart, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the cart must offer all five colours, got %+v", pick)
	}
}

// --- Ezekiel Sims, Spider-Totem -------------------------------------------

func TestSpm2EzekielPumpsASpiderYouControlAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Ezekiel Sims, Spider-Totem", "Legendary Creature — Spider Human Advisor", spm2EzekielOracle, 3, 5)
	spider := b12Push(g, me.ID, "Spider", "Creature — Spider", "", 1, 1)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Spider", "Creature — Spider", "", 1, 1)

	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, bear); err == nil {
		t.Fatal("a Bear is not a Spider")
	}
	if err := sgAnswerTargets(t, g, theirs); err == nil {
		t.Fatal("an opponent's Spider is not a Spider you control")
	}
	if err := sgAnswerTargets(t, g, spider); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, spider); p != 3 || th != 3 {
		t.Errorf("Spider P/T = %d/%d, want 3/3", p, th)
	}
}

func TestSpm2EzekielCanTargetHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ez := b12Push(g, me.ID, "Ezekiel Sims, Spider-Totem", "Legendary Creature — Spider Human Advisor", spm2EzekielOracle, 3, 5)
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, ez); err != nil {
		t.Fatalf("Ezekiel is a Spider: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, ez); p != 5 || th != 7 {
		t.Errorf("Ezekiel P/T = %d/%d, want 5/7", p, th)
	}
}

// --- Grow Extra Arms --------------------------------------------------------

func TestSpm2GrowExtraArmsCostsOneLessAtASpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := b12Creature(g, me.ID, "Spider", "Creature — Spider", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := hobB3Price(t, g, me, "Grow Extra Arms", "Instant", "{1}{G}", spider); got != 1 {
		t.Errorf("at a Spider: %d, want 1", got)
	}
	if got := hobB3Price(t, g, me, "Grow Extra Arms", "Instant", "{1}{G}", bear); got != 2 {
		t.Errorf("at a Bear: %d, want 2", got)
	}
}

func TestSpm2GrowExtraArmsGivesPlusFourPlusFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Grow Extra Arms", "Instant", spm2GrowArmsOracle, s58Refs(bear))
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 6 || th != 6 {
		t.Errorf("P/T %d/%d, want 6/6", p, th)
	}
}

// --- Guy in the Chair -------------------------------------------------------

func TestSpm2GuyInTheChairWebSupportCountersASpiderAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guy := pushCatalogPermanent(g, me.ID, "Guy in the Chair", "Creature — Human Advisor", spm2GuyOracle, false)
	spider := b12Creature(g, me.ID, "Spider", "Creature — Spider", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	floatForTest(g, me, "CCG")
	if err := g.ActivateCatalogAbility(me.ID, guy, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("Web Support put a counter on a Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, guy, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(spider)}); err != nil {
		t.Fatalf("Web Support: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, spider, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Spider = %d, want 1", got)
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters on the Bear = %d, want 0", got)
	}
}

func TestSpm2GuyInTheChairWebSupportIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guy := pushCatalogPermanent(g, me.ID, "Guy in the Chair", "Creature — Human Advisor", spm2GuyOracle, false)
	spider := b12Creature(g, me.ID, "Spider", "Creature — Spider", 1, 1)
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCG")
	if err := g.ActivateCatalogAbility(me.ID, guy, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(spider)}); err == nil {
		t.Fatal("Web Support was activated in the beginning of combat step")
	}
	if got := counterCount(g, spider, game.CounterPlusOne); got != 0 {
		t.Errorf("a refused activation put %d counters", got)
	}
}

func TestSpm2GuyInTheChairTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guy := pushCatalogPermanent(g, me.ID, "Guy in the Chair", "Creature — Human Advisor", spm2GuyOracle, false)
	if err := g.ActivateManaAbility(me.ID, guy, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("Guy must offer all five colours, got %+v", pick)
	}
}

// --- Green Goblin, Revenant -------------------------------------------------

func TestSpm2GreenGoblinRevenantDrawsForEveryCardDiscardedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goblin := b12Push(g, me.ID, "Green Goblin, Revenant", "Legendary Creature — Goblin Human Villain", spm2GoblinOracle, 3, 3)
	earlier := mayhemHandCard(g, me, "Earlier", "Sorcery", "{1}", "")
	keep := mayhemHandCard(g, me, "Keep", "Sorcery", "{1}", "")
	pitch := mayhemHandCard(g, me, "Pitch", "Sorcery", "{1}", "")
	advanceToMain(t, g)
	mayhemDiscard(t, g, me, earlier)
	hand := me.Hand.Size()

	declareAttack(t, g, opp.ID, goblin)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) || !me.Hand.Contains(keep) {
		t.Fatal("the chosen card is discarded and the other stays")
	}
	// One discard from the trigger, one earlier: two draws.
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d, want %d (discard 1, draw 2)", got, hand-1+2)
	}
}

func TestSpm2GreenGoblinRevenantWithAnEmptyHandStillDrawsForEarlierDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	goblin := b12Push(g, me.ID, "Green Goblin, Revenant", "Legendary Creature — Goblin Human Villain", spm2GoblinOracle, 3, 3)
	earlier := mayhemHandCard(g, me, "Earlier", "Sorcery", "{1}", "")
	advanceToMain(t, g)
	mayhemDiscard(t, g, me, earlier)
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			c := me.Hand.Cards[0]
			_, _ = me.Hand.Remove(c.InstanceID)
		}
	})

	declareAttack(t, g, opp.ID, goblin)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1: nothing to discard, one earlier discard to draw for", got)
	}
}

// --- Hide on the Ceiling ----------------------------------------------------

func TestSpm2HideOnTheCeilingExilesXAndReturnsThemAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushTypedCard(g, opp.ID, "Rock", "Artifact", "{1}")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spared := b12Creature(g, opp.ID, "Spared", "Creature — Bear", 2, 2)
	land := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")

	castXSpell(t, g, "Hide on the Ceiling", "Instant", spm2HideOracle, "{X}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) || !g.Exile.Contains(bear) {
		t.Fatal("both targets are exiled")
	}
	if !g.Battlefield.Contains(spared) || !g.Battlefield.Contains(land) {
		t.Error("an untargeted permanent was exiled")
	}
	advanceToStepInTurn(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, opp.ID, "Rock") != 1 || countBattlefieldNamed(g, me.ID, "Bear") != 1 {
		t.Error("both cards return to the battlefield under their owners' control")
	}
}

func TestSpm2HideOnTheCeilingRefusesALandAndAWrongCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hide on the Ceiling", TypeLine: "Instant",
		OracleID: spm2HideOracle, ManaCost: "{X}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("a land is neither an artifact nor a creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("one target was accepted for X=2")
	}
}

// --- Friendly Neighborhood --------------------------------------------------

func TestSpm2FriendlyNeighborhoodMakesThreeCitizensAndGrantsTheLandAPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Friendly Neighborhood", "Enchantment — Aura", spm2FriendlyOracle, cardRefs(land))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	citizens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Human Citizen" {
			citizens++
			if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Citizen") || !c.HasSubtype("Human") {
				t.Errorf("Citizen is %+v", c)
			}
		}
	}
	if citizens != 3 {
		t.Fatalf("%d Human Citizen tokens, want 3", citizens)
	}

	idx, ref := grantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("the granted pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Four creatures: three Citizens and the Bear itself.
	if p, th := effectivePTOf(t, g, bear); p != 6 || th != 6 {
		t.Errorf("Bear is %d/%d, want 6/6 (+1/+1 for each of four creatures)", p, th)
	}
}

func TestSpm2FriendlyNeighborhoodPumpIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pushAuraOnLand(t, g, me.ID, "Friendly Neighborhood", spm2FriendlyOracle, land)
	idx, ref := grantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("the pump was activated outside a main phase")
	}
}

// --- Eddie Brock // Venom, Lethal Protector ---------------------------------

// spm2EddieCard is Eddie Brock as Scryfall lays the card out: a modal
// double-faced card, front face up.
func spm2EddieCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   spm2EddieOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Eddie Brock", TypeLine: "Legendary Creature — Human Hero Villain", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
			{Name: "Venom, Lethal Protector", TypeLine: "Legendary Creature — Symbiote Hero Villain", ManaCost: "{3}{B}{R}{G}", Colors: []string{"B", "R", "G"}, Power: 5, Toughness: 5},
		},
	}
	c.SetFace(0)
	return c
}

func TestSpm2EddieBrockReturnsACreatureCardOfManaValueOneOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := spm2Graveyard(me, "Cheap", "Creature — Bear", "{G}")
	dear := spm2Graveyard(me, "Dear", "Creature — Bear", "{1}{G}")
	spell := spm2Graveyard(me, "Spell", "Instant", "{G}")
	castAndResolveCreature(t, g, "Eddie Brock", "Legendary Creature — Human Hero Villain", spm2EddieOracle)
	for _, bad := range []uuid.UUID{dear, spell} {
		if err := sgAnswerTargets(t, g, bad); err == nil {
			t.Fatal("only a creature card with mana value 1 or less is legal")
		}
	}
	if err := sgAnswerTargets(t, g, cheap); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) && countBattlefieldNamed(g, me.ID, "Cheap") != 1 {
		t.Error("the mana value 1 creature returns to the battlefield")
	}
	if !me.Graveyard.Contains(dear) || !me.Graveyard.Contains(spell) {
		t.Error("the other cards stay in the graveyard")
	}
}

func TestSpm2EddieBrockTransformsAtSorcerySpeedIntoVenom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eddie := pushBattlefieldCardWithTimestamp(g, spm2EddieCard(me.ID))
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCCBRG")
	if err := g.ActivateCatalogAbility(me.ID, eddie, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Transform was activated outside a main phase")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	advanceToStepInTurn(t, g, game.StepPostcombatMain)
	floatForTest(g, me, "CCCBRG")
	if err := g.ActivateCatalogAbility(me.ID, eddie, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := gaLayered(t, g, eddie)
	if got.Name != "Venom, Lethal Protector" {
		t.Fatalf("after Transform the permanent is %q, want Venom, Lethal Protector", got.Name)
	}
	for _, kw := range []string{"menace", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, eddie, kw) {
			t.Errorf("Venom lacks %s", kw)
		}
	}
}

func TestSpm2VenomAttackSacrificeDrawsXAndPutsAPermanentCardOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	venomCard := spm2EddieCard(me.ID)
	venomCard.SetFace(1)
	venom := pushBattlefieldCardWithTimestamp(g, venomCard)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	cheap := mayhemHandCard(g, me, "Cheap Permanent", "Creature — Bear", "{2}", "")
	dear := mayhemHandCard(g, me, "Dear Permanent", "Creature — Bear", "{5}", "")
	burn := mayhemHandCard(g, me, "Cheap Spell", "Instant", "{1}", "")
	advanceToMain(t, g)

	handBefore := me.Hand.Size()
	declareAttack(t, g, opp.ID, venom)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (you MAY sacrifice)", p.ChooseMin, p.ChooseMax)
	}
	if hasID(p.ChooseCards, venom) || !hasID(p.ChooseCards, fodder) {
		t.Error("ANOTHER creature: Venom is not offered")
	}
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the chosen creature is sacrificed")
	}
	// X is 3: three cards drawn.
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand %d, want %d (drew X=3)", got, handBefore+3)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the put offered nothing")
	}
	if !hasID(pick.ChooseCards, cheap) {
		t.Error("a permanent card with mana value 2 is offered")
	}
	if hasID(pick.ChooseCards, dear) || hasID(pick.ChooseCards, burn) {
		t.Error("mana value 5 and an instant are not offered")
	}
	answerChooseCards(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the chosen permanent card enters the battlefield")
	}
	if !me.Hand.Contains(dear) {
		t.Error("the unchosen card stays in hand")
	}
}

func TestSpm2VenomAttackDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	venomCard := spm2EddieCard(me.ID)
	venomCard.SetFace(1)
	venom := pushBattlefieldCardWithTimestamp(g, venomCard)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, venom)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Hand.Size() != hand {
		t.Error("declining sacrifices nothing and draws nothing")
	}
}
