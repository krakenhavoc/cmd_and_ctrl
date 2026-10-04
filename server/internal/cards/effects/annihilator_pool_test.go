package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// annihilator_pool_test.go — the #2073 pool (ADR 0113 §2 and owner
// decision 2): the rest of the cards annihilator unblocked.

const (
	artisanOfKozilekOracle    = "19409704-09c4-4a4b-a5a7-f95120b425db"
	nulldrifterOracle         = "ef8d9c3a-5860-45bf-8205-14e2df9a6fc8"
	eldraziRavagerOracle      = "25a73cfc-a40d-47ea-a7b1-61c7b62a9e6c"
	flayerOfLoyaltiesOracle   = "1e9053b5-5cca-486c-9dd8-b198a7b666bf"
	azlaskOracle              = "55bc7f55-f73b-40b5-8912-9bf76c129ccc"
	eldraziConscriptionOracle = "3635e4e5-e655-45f4-a444-39756f55129a"
	itThatBetraysOracle       = "b11187ab-f8a8-422b-b550-4495f96de0f2"
	pathrazerOracle           = "4bf95747-4572-49b8-b892-87fe7f910252"
	ulamogsCrusherOracle      = "597bce51-af89-4d13-9a97-667b4f4f4694"
	breakerOfCreationOracle   = "785026c5-3f26-489c-8e26-96dd3ca6bc98"
	handOfEmrakulOracle       = "b206d3bc-1203-4f00-997e-5da706346a24"
	nazgulBattleMaceOracle    = "1ea3ba2d-e88f-4879-92f2-681637094b45"
	spawnsireOracle           = "b90d2f4d-b4ea-40af-aea0-1ab4234ab80f"
	hideousTaskmasterOracle   = "8f48c43e-fa70-4c4f-bb88-be0526303453"
	idolOfFalseGodsOracle     = "b30d583b-e770-47e8-9f26-fca7ef43285c"
)

// annCard reads a battlefield card, failing the test if it is gone.
func annCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

// annAmounts is the annihilator numbers a battlefield card has.
func annAmounts(t *testing.T, g *game.Game, id uuid.UUID) []int {
	t.Helper()
	c := annCard(t, g, id)
	return game.AnnihilatorAmounts(&c)
}

// pushAnnCatalogCreature puts a ready catalog creature on the battlefield.
func pushAnnCatalogCreature(g *game.Game, owner uuid.UUID, name, oracle string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Eldrazi", OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestPrintedAnnihilatorPoolCardsCarryTheirNumber(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	for _, c := range []struct {
		name, oracle string
		want         int
	}{
		{"Artisan of Kozilek", artisanOfKozilekOracle, 2},
		{"Nulldrifter", nulldrifterOracle, 1},
		{"Eldrazi Ravager", eldraziRavagerOracle, 1},
		{"Flayer of Loyalties", flayerOfLoyaltiesOracle, 2},
		{"It That Betrays", itThatBetraysOracle, 2},
		{"Pathrazer of Ulamog", pathrazerOracle, 3},
		{"Ulamog's Crusher", ulamogsCrusherOracle, 2},
		{"Breaker of Creation", breakerOfCreationOracle, 2},
		{"Hand of Emrakul", handOfEmrakulOracle, 1},
		{"Spawnsire of Ulamog", spawnsireOracle, 1},
		{"Hideous Taskmaster", hideousTaskmasterOracle, 1},
	} {
		id := pushAnnCatalogCreature(g, me, c.name, c.oracle, 5, 5)
		if got := annAmounts(t, g, id); len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: annihilator %v, want [%d]", c.name, got, c.want)
		}
	}
}

func TestArtisanOfKozilekReturnsACreatureOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	id := putInHand(me, game.Card{Name: "Artisan of Kozilek", TypeLine: "Creature — Eldrazi",
		OracleID: artisanOfKozilekOracle, ManaCost: "{9}", Power: 10, Toughness: 9})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !annOnBattlefield(g, bear) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestNulldrifterEvokedDrawsTwoAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := putInHand(me, game.Card{Name: "Nulldrifter", TypeLine: "Creature — Eldrazi Elemental",
		OracleID: nulldrifterOracle, ManaCost: "{7}", Power: 4, Toughness: 4})
	toMain(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("CastSpell evoke: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +2 drawn -1 cast", hand, got)
	}
	if annOnBattlefield(g, id) || !me.Graveyard.Contains(id) {
		t.Error("an evoked Nulldrifter was not sacrificed")
	}
}

func TestEldraziRavagerReturnsFromTheGraveyardForTwoEldrazi(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	ravager := pushGraveyardPermanent(me, "Eldrazi Ravager", "Creature — Eldrazi", "{5}{C}")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == ravager {
			me.Graveyard.Cards[i].OracleID = eldraziRavagerOracle
		}
	}
	a := pushAnnCatalogCreature(g, me.ID, "Eldrazi Spawn", "", 0, 1)
	b := pushAnnCatalogCreature(g, me.ID, "Eldrazi Scion", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ravager, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ravager) {
		t.Error("the Ravager did not return to hand")
	}
	if annOnBattlefield(g, a) || annOnBattlefield(g, b) {
		t.Error("the two Eldrazi were not sacrificed")
	}
}

func TestFlayerOfLoyaltiesStealsAndMakesA1010Annihilator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	id := putInHand(me, game.Card{Name: "Flayer of Loyalties", TypeLine: "Creature — Eldrazi",
		OracleID: flayerOfLoyaltiesOracle, ManaCost: "{10}", Power: 10, Toughness: 10})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c := annCard(t, g, bear)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %s tapped %v; want mine and untapped", c.Controller, c.Tapped)
	}
	if c.CurrentPower() != 10 || c.CurrentToughness() != 10 {
		t.Errorf("P/T %d/%d, want 10/10", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
		t.Error("the stolen creature lacks trample or haste")
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v, want [2]", got)
	}
}

func TestAzlaskCountsColorlessDeathsAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	azlask := pushAnnCatalogCreature(g, me.ID, "Azlask, the Swelling Scourge", azlaskOracle, 2, 2)
	colorless := pushAnnCatalogCreature(g, me.ID, "Eldrazi Drone", "", 2, 2)
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
		TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	spawn := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Eldrazi Spawn",
		TypeLine: "Token Creature — Eldrazi Spawn", Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID})
	for _, id := range []uuid.UUID{colorless, green} {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	if got := experienceCounters(g, me.ID); got != 1 {
		t.Fatalf("experience counters = %d, want 1 (the colourless death only)", got)
	}
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, azlask, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := annCard(t, g, azlask); c.CurrentPower() != 3 {
		t.Errorf("Azlask power %d, want 2+1", c.CurrentPower())
	}
	s := annCard(t, g, spawn)
	if !game.HasKeyword(&s, "indestructible") || len(game.AnnihilatorAmounts(&s)) != 1 {
		t.Error("the Spawn did not gain indestructible and annihilator 1")
	}
	if a := annCard(t, g, azlask); len(game.AnnihilatorAmounts(&a)) != 0 {
		t.Error("Azlask is not a Scion or Spawn and must not gain annihilator")
	}
}

func TestEldraziConscriptionAddsASecondAnnihilator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushAnnihilatorCreature(g, me.ID, "annihilator 4")
	aura := pushCatalogPermanent(g, me.ID, "Eldrazi Conscription", "Kindred Enchantment — Eldrazi Aura", eldraziConscriptionOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	c := annCard(t, g, host)
	if c.CurrentPower() != 15 || !game.HasKeyword(&c, "trample") {
		t.Errorf("host is %d power, trample %v; want 15 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 2 {
		t.Errorf("annihilator %v, want the printed 4 and the granted 2", got)
	}
}

func TestItThatBetraysTakesAnOpponentsSacrificedPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushAnnCatalogCreature(g, me.ID, "It That Betrays", itThatBetraysOracle, 11, 11)
	land := pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spawn",
		TypeLine: "Token Creature — Eldrazi Spawn", Owner: opp.ID, Controller: opp.ID, Toughness: 1})
	mine := pushCatalogPermanent(g, me.ID, "My Forest", "Basic Land — Forest", "", false)
	for _, s := range []struct{ who, id uuid.UUID }{{opp.ID, land}, {opp.ID, token}, {me.ID, mine}} {
		if err := g.SacrificePermanent(s.who, s.id); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	c, ok := battlefieldCard(g, land)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the opponent's sacrificed land is not mine on the battlefield (on: %v)", ok)
	}
	if annOnBattlefield(g, mine) {
		t.Error("my own sacrifice came back")
	}
}

func TestPathrazerNeedsThreeBlockersAndCrusherMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	p := pushAnnCatalogCreature(g, me, "Pathrazer of Ulamog", pathrazerOracle, 9, 9)
	if rules := game.CatalogBlockRules(annCard(t, g, p).OracleID); len(rules) != 1 {
		t.Errorf("Pathrazer declares %d block rules, want the three-blocker minimum", len(rules))
	}
	if statics := game.CatalogStaticAbilities(ulamogsCrusherOracle); len(statics) == 0 {
		t.Errorf("Ulamog's Crusher declares %d statics, want attacks each combat", len(statics))
	}
}

func TestBreakerOfCreationGainsLifePerColorlessPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushAnnCatalogCreature(g, me.ID, "Drone", "", 1, 1)
	pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
		TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	id := putInHand(me, game.Card{Name: "Breaker of Creation", TypeLine: "Creature — Eldrazi",
		OracleID: breakerOfCreationOracle, ManaCost: "{6}{C}{C}", Power: 8, Toughness: 4})
	toMain(t, g)
	life := me.Life
	colorless := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsColorless() {
			colorless++
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != colorless {
		t.Errorf("gained %d life, want %d", got, colorless)
	}
}

func TestHandOfEmrakulOffersTheSpawnSacrifice(t *testing.T) {
	ac := game.AlternativeCostByKey(handOfEmrakulOracle, "sacrifice")
	if ac == nil || ac.Sacrifice == nil || ac.Sacrifice.Min != 4 || ac.Sacrifice.Max != 4 || ac.ManaCost != "" {
		t.Fatalf("alternative cost = %+v, want sacrifice exactly four and no mana", ac)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	var spawn []uuid.UUID
	for i := 0; i < 4; i++ {
		spawn = append(spawn, pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Eldrazi Spawn",
			TypeLine: "Token Creature — Eldrazi Spawn", Toughness: 1, Owner: me.ID, Controller: me.ID}))
	}
	id := putInHand(me, game.Card{Name: "Hand of Emrakul", TypeLine: "Creature — Eldrazi",
		OracleID: handOfEmrakulOracle, ManaCost: "{9}", Power: 7, Toughness: 7})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "sacrifice", AltCostIDs: spawn, Strict: true}); err != nil {
		t.Fatalf("CastSpell for four Spawn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !annOnBattlefield(g, id) {
		t.Error("Hand of Emrakul did not resolve")
	}
	for _, s := range spawn {
		if annOnBattlefield(g, s) {
			t.Error("a Spawn was not sacrificed")
		}
	}
}

func TestNazgulBattleMaceGrantsAndTakesUnlessLifeIsPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushAnnCatalogCreature(g, me.ID, "Knight", "", 3, 3)
	mace := pushCatalogPermanent(g, me.ID, "Nazgûl Battle-Mace", "Artifact — Equipment", nazgulBattleMaceOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(mace, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	c := annCard(t, g, host)
	if !game.HasKeyword(&c, "menace") || !game.HasKeyword(&c, "deathtouch") || len(game.AnnihilatorAmounts(&c)) != 1 {
		t.Fatalf("the equipped creature lacks menace, deathtouch or annihilator 1: %v", c.Effective().Abilities)
	}

	kept := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	taken := pushCatalogPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", "", false)
	life := opp.Life
	if err := g.SacrificePermanent(opp.ID, kept); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 0) // pay 3 life
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 || annOnBattlefield(g, kept) {
		t.Errorf("paying 3 life: life %d -> %d, card on battlefield %v", life, opp.Life, annOnBattlefield(g, kept))
	}
	if err := g.SacrificePermanent(opp.ID, taken); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1) // don't pay
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, taken); !ok || c.Controller != me.ID {
		t.Error("not paying did not give me the card")
	}
}

func TestSpawnsireMakesTwoSpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	id := pushAnnCatalogCreature(g, me.ID, "Spawnsire of Ulamog", spawnsireOracle, 7, 11)
	before := cardsNamed(g, "Eldrazi Spawn", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := cardsNamed(g, "Eldrazi Spawn", me.ID) - before; got != 2 {
		t.Errorf("made %d Spawn, want 2", got)
	}
}

func TestHideousTaskmasterTakesOneCreaturePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	ca := pushCatalogPermanent(g, a.ID, "Bear A", "Creature — Bear", "", false)
	cb := pushCatalogPermanent(g, b.ID, "Bear B", "Creature — Bear", "", false)
	id := putInHand(me, game.Card{Name: "Hideous Taskmaster", TypeLine: "Creature — Eldrazi",
		OracleID: hideousTaskmasterOracle, ManaCost: "{6}{R}", Power: 7, Toughness: 2})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// One "up to one" clause per opponent, asked in seat order: seat
	// 1's creature, then seat 2's. Seat 3 controls nothing, so its
	// clause is skipped without a prompt.
	for i, pick := range []uuid.UUID{ca, cb} {
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatalf("no target prompt for opponent %d", i+1)
		}
		if !hasID(p.PickTargetCards, pick) || p.PickTargetMax != 1 || p.PickTargetMin != 0 {
			t.Fatalf("clause %d offers %v (%d..%d), want up to one of that opponent's creatures",
				i+1, p.PickTargetCards, p.PickTargetMin, p.PickTargetMax)
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: pick}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a prompt for the opponent with no creatures: %q", p.Reason)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ca, cb} {
		c := annCard(t, g, id)
		if c.Controller != me.ID || !game.HasKeyword(&c, "haste") || len(game.AnnihilatorAmounts(&c)) != 1 {
			t.Errorf("%s: controller %s, haste %v, annihilator %v", c.Name, c.Controller,
				game.HasKeyword(&c, "haste"), game.AnnihilatorAmounts(&c))
		}
	}
}

func TestIdolOfFalseGodsGrowsAndBecomesAnAnnihilator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := pushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", idolOfFalseGodsOracle, false)
	drone := pushAnnCatalogCreature(g, me.ID, "Drone", "", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(drone); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, idol, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after an Eldrazi died = %d, want 1", got)
	}
	if c := annCard(t, g, idol); c.IsCreature() {
		t.Error("the Idol is a creature at one counter")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 7); err != nil {
			t.Fatalf("add counters: %v", err)
		}
	})
	c := annCard(t, g, idol)
	if !c.IsCreature() || c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("at eight counters: creature %v, %d/%d; want an 8/8 creature", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v, want [2]", got)
	}
}
