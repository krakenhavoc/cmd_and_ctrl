package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// trek_b26_cards_test.go — Star Trek, batch 26 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB26SpockOracle    = "bdfe5809-17d4-4239-a9ee-6accdef99035"
	trekB26TravelerOracle = "b61718eb-dc6a-4199-b7c7-42eccf4158e8"
	trekB26VashOracle     = "ec5a9dde-b89d-4c54-8563-4a7170f5cefd"
	trekB26WorfOracle     = "2ecda53d-ac24-4b28-aed8-2a5fb1449f85"
	trekB26YearOracle     = "83f426db-3b25-4407-afd0-0465e3900c24"
	trekB26ZeroOracle     = "746e494e-fdc4-415e-8a82-d35e11ebeccb"
	trekB26ShuttleOracle  = "4ccc91f2-4d6e-4d64-b438-3d8b4361040a"
	trekB26CombadgeOracle = "ccd0e613-17c1-4a97-b519-de0376d596f0"
	trekB26EntityOracle   = "ca6f43d8-1498-4a13-b1cc-2d135b24a8b6"
)

func TestTrekB26EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB26SpockOracle:    "Spock, Always Your Friend",
		trekB26TravelerOracle: "The Traveler",
		trekB26VashOracle:     "Vash, Freelance Archaeologist",
		trekB26WorfOracle:     "Worf, Honor-Bound Warrior",
		trekB26YearOracle:     "Year of Hell",
		trekB26ZeroOracle:     "Zero, Incorporeal Sensate",
		trekB26ShuttleOracle:  "Class-F Shuttlecraft",
		trekB26CombadgeOracle: "Combadge",
		trekB26EntityOracle:   "Crystalline Entity",
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

// --- fixtures ------------------------------------------------------------------

// trekB26Named counts the permanents named `name` that p controls, and how
// many of them are tokens.
func trekB26Named(g *game.Game, p uuid.UUID, name string) (all, tokens int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.Name == name {
				all++
				if IsToken(c) {
					tokens++
				}
			}
		}
	})
	return all, tokens
}

// trekB26Attack declares the attackers at defender and settles what the
// declaration triggers.
func trekB26Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
}

// trekB26CantBeBlocked reports whether the creature carries a
// can't-be-blocked restriction.
func trekB26CantBeBlocked(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return auraRestrictions(t, g, id)&game.CantBeBlocked != 0
}

// trekB26Land puts a land onto the battlefield under owner's control.
func trekB26Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner})
}

// --- Spock, Always Your Friend ---------------------------------------------------

func TestTrekB26SpockTakesANoncreatureNonlandCardFromTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Relic", TypeLine: "Artifact"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	bolt, bear, forest, relic, deep := ids[0], ids[1], ids[2], ids[3], ids[4]
	if top, _ := me.Library.Top(); top.InstanceID != bolt {
		t.Fatal("non-vacuity: Bolt is not on top")
	}
	castHoldCreature(t, g, "Spock, Always Your Friend", "Legendary Creature — Vulcan Human Scientist", trekB26SpockOracle, 2, 2)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no dig prompt")
	}
	if len(pick.ChooseCards) != 2 || !slices.Contains(pick.ChooseCards, relic) || !slices.Contains(pick.ChooseCards, bolt) {
		t.Fatalf("offered %v, want only the Relic and the Bolt (not the land, the creature or the fifth card)", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatal(err)
	}
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("the rest are put on the bottom in an order the player chooses")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{bear, forest, bolt}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Hand.Contains(relic) {
		t.Fatal("the Relic did not reach the hand")
	}
	for _, id := range []uuid.UUID{forest, bear, bolt} {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Error("a card not taken left the library")
		}
	}
	if top, _ := me.Library.Top(); top.InstanceID != deep {
		t.Error("the rest did not go to the bottom: the fifth card is not on top")
	}
	if bottom := me.Library.Cards[0].InstanceID; bottom != bolt {
		t.Error("the player's order was not kept: the Bolt is not at the bottom")
	}
}

// --- The Traveler -----------------------------------------------------------------

func TestTrekB26TravelerCountsDifferentlyNamedLandsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	traveler := b12Push(g, me.ID, "The Traveler", "Legendary Creature — Scout", trekB26TravelerOracle, 0, 3)
	if got := currentPower(t, g, traveler); got != 0 {
		t.Fatalf("power %d with no lands, want 0", got)
	}
	trekB26Land(g, me.ID, "Forest", "Basic Land — Forest")
	trekB26Land(g, me.ID, "Forest", "Basic Land — Forest")
	trekB26Land(g, me.ID, "Island", "Basic Land — Island")
	trekB26Land(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	trekB26Land(g, opp.ID, "Swamp", "Basic Land — Swamp")
	if got := currentPower(t, g, traveler); got != 3 {
		t.Errorf("power %d, want 3 (Forest, Island, Snow-Covered Forest; not the second Forest, not the opponent's Swamp)", got)
	}
	if got := layeredCard(t, g, traveler).CurrentToughness(); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	for _, kw := range []string{"reach", "vigilance"} {
		if !hasEffectiveKeyword(t, g, traveler, kw) {
			t.Errorf("The Traveler has no %s", kw)
		}
	}
}

// --- Vash, Freelance Archaeologist ----------------------------------------------

func TestTrekB26VashSurveilsOneAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Vash, Freelance Archaeologist", "Legendary Creature — Human Rogue Scientist", trekB26VashOracle, 2, 4)
	ids := seedSearchLibrary(owner,
		game.Card{Name: "Top", TypeLine: "Instant"},
		game.Card{Name: "Second", TypeLine: "Instant"},
	)
	advanceToUpkeepOf(t, g, 1)
	monSettle(t, g)
	var ask *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == owner.ID {
			ask = c
		}
	}
	if ask == nil {
		t.Fatal("no surveil prompt at Vash's controller's upkeep")
	}
	if len(ask.ScryCards) != 1 || ask.ScryCards[0] != ids[0] {
		t.Fatalf("surveil looks at %v, want only the top card", ask.ScryCards)
	}
	if err := g.ResolveSurveil(ask.ID, owner.ID, []uuid.UUID{ids[0]}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !owner.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card is not in the graveyard")
	}
}

func TestTrekB26VashDrawsOnceEachTurnWhenCardsLeaveYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	vash := b12Push(g, me.ID, "Vash, Freelance Archaeologist", "Legendary Creature — Human Rogue Scientist", trekB26VashOracle, 2, 4)
	seedSearchLibrary(me,
		game.Card{Name: "Draw C", TypeLine: "Instant"},
		game.Card{Name: "Draw B", TypeLine: "Instant"},
		game.Card{Name: "Draw A", TypeLine: "Instant"},
	)
	a := pushGraveyardCardTyped(me, "Old A", "Sorcery")
	b := pushGraveyardCardTyped(me, "Old B", "Sorcery")
	c := pushGraveyardCardTyped(me, "Old C", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Old", "Sorcery")
	hand := me.Hand.Size()

	// An opponent's graveyard is not yours.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{theirs}) })
	monSettle(t, g)
	if triggersOnStackFrom(g, vash) != 0 || me.Hand.Size() != hand {
		t.Fatal("a card leaving an opponent's graveyard triggered Vash")
	}

	// Two cards leaving together: one trigger, one card drawn.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d after two cards left your graveyard, want 1", got)
	}

	// Later the same turn: no more.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{c}) })
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d this turn, want 1: the ability triggers only once each turn", got)
	}
}

// --- Worf, Honor-Bound Warrior ----------------------------------------------------

func trekB26PushWorf(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Worf, Honor-Bound Warrior", "Legendary Creature — Klingon Warrior Officer", trekB26WorfOracle, 1, 1)
}

func TestTrekB26WorfDoublesHisPowerWhenYouAttackWithEightTotalPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worf := trekB26PushWorf(g, me.ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(worf, game.CounterPlusOne, 2) })
	big := pushVanillaCreature(g, me.ID, "Big Ally", 5, 5)
	// 3 + 5 = 8: Worf does not have to attack.
	trekB26Attack(t, g, opp.ID, big)
	if got := currentPower(t, g, worf); got != 6 {
		t.Errorf("Worf's power %d after you attacked with 8 total power, want 6 (3 doubled)", got)
	}
}

func TestTrekB26WorfNeedsEightTotalPowerAsYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worf := trekB26PushWorf(g, me.ID)
	ally := pushVanillaCreature(g, me.ID, "Ally", 6, 6)
	// An opponent's power is not "creatures you control".
	pushVanillaCreature(g, opp.ID, "Their Giant", 9, 9)
	trekB26Attack(t, g, opp.ID, ally, worf)
	if got := currentPower(t, g, worf); got != 1 {
		t.Errorf("Worf's power %d with 7 total power, want 1", got)
	}
}

func TestTrekB26WorfTapsTwoArtifactsOrCreaturesForACounterAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	worf := trekB26PushWorf(g, me.ID)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, worf, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{worf, land}, Strict: true}); err == nil {
		t.Fatal("a land paid for \"artifacts and/or creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, worf, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{worf, relic}, Strict: true}); err != nil {
		t.Fatalf("tapping Worf and an artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, worf); got != 1 {
		t.Fatalf("Worf has %d +1/+1 counters, want 1", got)
	}
	if !layeredCard(t, g, relic).Tapped || !layeredCard(t, g, worf).Tapped {
		t.Error("the cost did not tap both")
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, worf, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Error("one untapped creature paid for two")
	}
}

func TestTrekB26WorfCantActivateOutsideSorceryTiming(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worf := trekB26PushWorf(g, me.ID)
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	trekB26Attack(t, g, opp.ID, a)
	if err := g.ActivateCatalogAbility(me.ID, worf, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{worf, b}, Strict: true}); err == nil {
		t.Error("activated during combat")
	}
}

// --- Year of Hell -----------------------------------------------------------------

func TestTrekB26YearOfHellEachPlayerDrawsAndTakesTwoEachChapter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	advanceToMain(t, g)
	life := map[uuid.UUID]int{}
	hand := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID], hand[p.ID] = p.Life, p.Hand.Size()
	}
	saga := castCatalogSpell(t, g, "Year of Hell", "Enchantment — Saga", trekB26YearOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := p.Hand.Size() - hand[p.ID]; got != 1 {
			t.Errorf("seat %d drew %d on chapter I, want 1", p.Seat, got)
		}
		if got := life[p.ID] - p.Life; got != 2 {
			t.Errorf("seat %d took %d on chapter I, want 2", p.Seat, got)
		}
	}
	if !g.Battlefield.Contains(saga) {
		t.Fatal("the Saga left after chapter I")
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := life[p.ID] - p.Life; got != 6 {
			t.Errorf("seat %d took %d over three chapters, want 6", p.Seat, got)
		}
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- Zero, Incorporeal Sensate ----------------------------------------------------

// trekB26ZeroEnters casts Zero from the active seat and targets opp; the
// look-and-pick is left open.
func trekB26ZeroEnters(t *testing.T, g *game.Game, opp *game.Player) uuid.UUID {
	t.Helper()
	zero := castHoldCreature(t, g, "Zero, Incorporeal Sensate", "Legendary Artifact Creature — Medusan Scientist", trekB26ZeroOracle, 2, 2)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	return zero
}

func TestTrekB26ZeroExilesANonlandCardItsOwnerMayCastForTwoMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = revealHand(opp, rhForest(), game.Card{Name: "Spark", TypeLine: "Instant", ManaCost: "{R}"})
	})
	forest, spark := ids[0], ids[1]
	zero := trekB26ZeroEnters(t, g, opp)
	if !hasEffectiveKeyword(t, g, zero, "lifelink") {
		t.Error("Zero has no lifelink")
	}
	pick := openVariant(t, g)
	if pick.Chooser != me.ID || slices.Contains(pick.DiscardOptions, forest) || !slices.Contains(pick.DiscardOptions, spark) {
		t.Fatalf("pick: chooser %v options %v; want Zero's controller choosing among nonland cards", pick.Chooser, pick.DiscardOptions)
	}
	third := g.Seats[2].ID
	for _, c := range opp.Hand.Cards {
		if c.IsKnownTo(third) {
			t.Errorf("%s: a third player saw it; Zero looks, it does not reveal", c.Name)
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spark}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	monSettle(t, g)
	if !g.Exile.Contains(spark) || opp.Hand.Contains(spark) {
		t.Fatal("the chosen card was not exiled")
	}

	var price game.CastPrice
	var err error
	g.ReadSnapshot(func() {
		c, _ := exileCard(g, spark)
		price, err = g.PriceCastForEffect(opp.ID, c, game.CastSpellParams{FromZone: "exile"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := price.Total.String(); got != "{2}{R}" {
		t.Errorf("price = %s, want {2}{R}", got)
	}
	if err := g.CastSpell(me.ID, spark, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("Zero's controller cast the exiled card; only its owner may")
	}

	// Zero leaving does not end the owner's permission.
	b15Destroy(t, g, zero)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if enumeratedFor(g, opp.ID, spark) {
		t.Error("the enumerator offered the card for {R}")
	}
	if err := g.CastSpell(opp.ID, spark, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the owner cast it for {R}; it costs {2} more")
	}
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if !enumeratedFor(g, opp.ID, spark) {
		t.Error("the enumerator did not offer the card for {2}{R}")
	}
	if err := g.CastSpell(opp.ID, spark, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("the owner could not cast it for {2}{R}: %v", err)
	}
	_ = legal.TypeCastSpell
}

func TestTrekB26ZeroMayExileNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = revealHand(opp, game.Card{Name: "Spark", TypeLine: "Instant", ManaCost: "{R}"})
	})
	trekB26ZeroEnters(t, g, opp)
	if err := g.ResolvePendingChoice(openVariant(t, g).ID, me.ID, nil); err != nil {
		t.Fatalf("choosing nothing: %v", err)
	}
	monSettle(t, g)
	if !opp.Hand.Contains(ids[0]) || g.Exile.Contains(ids[0]) {
		t.Error("a card was exiled though none was chosen")
	}
}

func TestTrekB26ZeroTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Zero, Incorporeal Sensate", "Legendary Artifact Creature — Medusan Scientist", trekB26ZeroOracle, 2, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("Zero's controller was accepted as the target opponent")
	}
}

// --- Class-F Shuttlecraft ---------------------------------------------------------

func TestTrekB26ShuttlecraftTapsForAnyColorAndFliesWithVigilanceAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ship := pushCatalogPermanent(g, me.ID, "Class-F Shuttlecraft", "Artifact — Spacecraft", trekB26ShuttleOracle, false)
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("an unstationed Spacecraft is a creature")
	}
	if err := g.ActivateManaAbility(me.ID, ship, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if c := latestChoiceOfKind(g, game.PendingChoiceMana); c == nil || len(c.ColorOptions) != 5 {
		t.Fatal("the mana ability does not offer each of the five colors")
	}
	resolveManaPickAnyColor(t, g, me.ID)
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool holds %d mana, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(ship) })
	crew := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Crew", TypeLine: "Creature — Human", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("a creature at 5 charge counters")
	}
	crew2 := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Crew Two", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	station(t, g, me, ship, crew2)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 2 || c.CurrentToughness() != 4 ||
		!game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") {
		t.Errorf("at 6+ a 2/4 flying, vigilance artifact creature; creature=%v %d/%d flying=%v vigilance=%v",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"))
	}
}

// --- Combadge ---------------------------------------------------------------------

func trekB26CastCombadge(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castFromHand(t, g, game.Card{Name: "Combadge", TypeLine: "Artifact — Communicator Equipment", OracleID: trekB26CombadgeOracle, ManaCost: "{3}"}, game.CastSpellParams{})
}

func TestTrekB26CombadgeCopiesItselfWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekB26CastCombadge(t, g)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if all, tokens := trekB26Named(g, me.ID, "Combadge"); all != 2 || tokens != 1 {
		t.Errorf("%d Combadges (%d tokens), want the card and one token copy", all, tokens)
	}
}

func TestTrekB26CombadgeCounteredIsStillCopied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	badge := trekB26CastCombadge(t, g)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(badge); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if all, tokens := trekB26Named(g, me.ID, "Combadge"); all != 1 || tokens != 1 {
		t.Errorf("%d Combadges (%d tokens), want only the token copy", all, tokens)
	}
}

func TestTrekB26CombadgeMakesTheWearerUnblockableWithAnotherCommunicatorOut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := pushVanillaCreature(g, me.ID, "Away Team A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Away Team B", 2, 2)
	badgeA := pushCatalogPermanent(g, me.ID, "Combadge", "Artifact — Communicator Equipment", trekB26CombadgeOracle, false)
	equipTo(t, g, me.ID, badgeA, a)

	// Alone, the Communicator on A does nothing.
	trekB26Attack(t, g, opp.ID, a)
	if trekB26CantBeBlocked(t, g, a) {
		t.Fatal("A is unblockable with no other creature wearing a Communicator")
	}

	g2 := newCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	a = pushVanillaCreature(g2, me.ID, "Away Team A", 2, 2)
	b = pushVanillaCreature(g2, me.ID, "Away Team B", 2, 2)
	badgeA = pushCatalogPermanent(g2, me.ID, "Combadge", "Artifact — Communicator Equipment", trekB26CombadgeOracle, false)
	badgeB := pushCatalogPermanent(g2, me.ID, "Combadge", "Artifact — Communicator Equipment", trekB26CombadgeOracle, false)
	equipTo(t, g2, me.ID, badgeA, a)
	equipTo(t, g2, me.ID, badgeB, b)
	plain := pushVanillaCreature(g2, me.ID, "Redshirt", 2, 2)
	trekB26Attack(t, g2, opp.ID, a, b, plain)
	if !trekB26CantBeBlocked(t, g2, a) || !trekB26CantBeBlocked(t, g2, b) {
		t.Error("each wearer should be unblockable while another creature wears a Communicator")
	}
	if trekB26CantBeBlocked(t, g2, plain) {
		t.Error("a creature with no Combadge became unblockable")
	}
}

// --- Crystalline Entity -----------------------------------------------------------

func TestTrekB26CrystallineEntityCastDestroysAllNonartifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	entity := castHoldCreature(t, g, "Crystalline Entity", "Artifact Creature — Elemental", trekB26EntityOracle, 7, 7)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a nonartifact creature survived")
	}
	for _, id := range []uuid.UUID{entity, golem, relic} {
		if !g.Battlefield.Contains(id) {
			t.Error("an artifact was destroyed")
		}
	}
	if !hasEffectiveKeyword(t, g, entity, "flying") {
		t.Error("Crystalline Entity has no flying")
	}
}

func TestTrekB26CrystallineEntityNotCastDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b12Push(g, me.ID, "Crystalline Entity", "Artifact Creature — Elemental", trekB26EntityOracle, 7, 7)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("an Entity that was not cast destroyed a creature")
	}
}
