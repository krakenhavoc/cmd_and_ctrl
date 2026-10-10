package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_f_test.go — slice fra-creature-f (tracker
// #2795): Reality Fracture creatures, one or more tests each.

const (
	oracleTraxosAcademy    = "7b9a574f-6d56-421b-8fbd-94da1e6e3ce1"
	oracleTraxosScourge    = "acf4abde-4034-4fcf-b531-78c935aecd8c"
	oracleUndulatingWit    = "77b684b8-227b-40de-80b4-15d0e70914f4"
	oracleHortimancer      = "7cd0bd74-2463-4e9d-9802-69ac4ef6d543"
	oracleVenserFervent    = "9b1ae6d8-0cf3-4a67-9241-37f4bc5ddbdd"
	oracleVerdantKraken    = "83c61083-b357-4775-a855-5a9013cc2bdf"
	oracleVinelasher       = "b5918fb0-c33f-4e33-ac6d-8f94723d4296"
	oracleVraskaSoul       = "cad63b31-667b-4518-bc49-be4ffb82bc63"
	oracleVraskaGlare      = "580de511-863c-4b7a-9d2e-fcdf6e12d54b"
	oracleWinterTeam       = "254f133a-eee5-48b9-b19f-110555e49300"
	oracleWreckingGecko    = "4a9e0294-c967-4e90-b776-6d3195034a23"
	oracleYargle           = "014d0bcd-b80f-4c43-8f00-f4c31e8f3370"
	oracleYoshimaruBelov   = "c00a0be5-d9a9-4263-a87e-7bf2d128ab3a"
	oracleYoshimaruStray   = "c87da757-3daa-4874-a1e4-be0aa2adcb28"
	oracleYurikoBlade      = "f89b0143-4b19-4ab1-a96b-356b21e435ef"
	rfCrFNoncreatureOracle = "00000000-0000-0000-0000-00000000f001"
)

// rfCrFHand puts a card in the active seat's hand.
func rfCrFHand(g *game.Game, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	return id
}

// rfCrFCast puts a free creature in hand, goes to the main phase and
// casts it.
func rfCrFCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFHand(g, name, typeLine, oracle, "", power, toughness)
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// rfCrFPush seats a permanent on the battlefield under owner.
func rfCrFPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func rfCrFCounters(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func rfCrFMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

func rfCrFToughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var v int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		v = c.CurrentToughness()
	})
	return v
}

func rfCrFPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var v int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		v = c.CurrentPower()
	})
	return v
}

func rfCrFHasAbility(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	var has bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		has = game.HasKeyword(&c, kw)
	})
	return has
}

// rfCrFCastNoncreature casts a free noncreature spell with no effect.
func rfCrFCastNoncreature(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFHand(g, "Idle Trick", "Instant", rfCrFNoncreatureOracle, "", 0, 0)
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell noncreature: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- Traxos, Academy Guardian ---------------------------------------

func TestTraxosAcademyGuardianCostsTwoLessAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	const typeLine = "Legendary Artifact Creature — Dragon Construct"
	if got := selfPricedMV(t, g, me, oracleTraxosAcademy, typeLine, "{3}{U}"); got != 4 {
		t.Fatalf("Traxos with no spell cast this turn: mana value %d, want 4", got)
	}
	rfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	passPriorityAroundTable(t, g)
	if got := selfPricedMV(t, g, me, oracleTraxosAcademy, typeLine, "{3}{U}"); got != 4 {
		t.Fatalf("Traxos after a creature spell: mana value %d, want 4", got)
	}
	rfCrFCastNoncreature(t, g)
	if got := selfPricedMV(t, g, me, oracleTraxosAcademy, typeLine, "{3}{U}"); got != 2 {
		t.Fatalf("Traxos after a noncreature spell: mana value %d, want 2", got)
	}
	id := rfCrFPush(g, me.ID, "Traxos, Academy Guardian", typeLine, oracleTraxosAcademy, 1, 5)
	for _, kw := range []string{"flying", "vigilance", game.KeywordProwess} {
		if !rfCrFHasAbility(t, g, id, kw) {
			t.Errorf("Traxos lacks %s", kw)
		}
	}
}

func TestTraxosAcademyGuardianProwessPumpsItAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Traxos, Academy Guardian", "Legendary Artifact Creature — Dragon Construct",
		oracleTraxosAcademy, 1, 5)
	rfCrFCastNoncreature(t, g)
	if p := rfCrFPower(t, g, id); p != 2 {
		t.Fatalf("power %d after a noncreature spell, want 2 (prowess)", p)
	}
}

// --- Traxos, Scourge Eternal ----------------------------------------

func TestTraxosScourgeEternalUntapsOnlyForAnArtifactOrCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Traxos, Scourge Eternal", "Legendary Artifact Creature — Dragon Construct",
		oracleTraxosScourge, 5, 4)
	if err := g.TapCard(id, true); err != nil {
		t.Fatal(err)
	}
	rfCrFCastNoncreature(t, g)
	if c, _ := battlefieldCard(g, id); !c.Tapped {
		t.Fatal("a plain noncreature spell untapped Traxos")
	}
	rfCrFCast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Fatal("a creature spell did not untap Traxos")
	}

	if err := g.TapCard(id, true); err != nil {
		t.Fatal(err)
	}
	rock := rfCrFHand(g, "Rock", "Artifact", "", "", 0, 0)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Fatal("an artifact spell did not untap Traxos")
	}
}

func TestTraxosScourgeEternalStaysTappedThroughYourUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	traxos := rfCrFPush(g, me.ID, "Traxos, Scourge Eternal", "Legendary Artifact Creature — Dragon Construct",
		oracleTraxosScourge, 5, 4)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{traxos, bear} {
		if err := g.TapCard(id, true); err != nil {
			t.Fatal(err)
		}
	}
	seat := g.Turn.ActiveSeat
	advanceToUpkeepOfSeat(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOfSeat(t, g, seat)
	if c, _ := battlefieldCard(g, traxos); !c.Tapped {
		t.Fatal("Traxos untapped during its controller's untap step")
	}
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Fatal("the control creature should have untapped")
	}
}

// --- Undulating Witness ---------------------------------------------

func TestUndulatingWitnessPumpsPlusOneMinusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Undulating Witness", "Creature — Serpent", oracleUndulatingWit, 3, 5)
	rfCrFMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := rfCrFPower(t, g, id), rfCrFToughness(t, g, id); p != 4 || tough != 4 {
		t.Fatalf("P/T %d/%d after one activation, want 4/4", p, tough)
	}
	if !rfCrFHasAbility(t, g, id, "flying") {
		t.Fatal("no flying")
	}
}

func TestUndulatingWitnessBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	toMain(t, g)
	id := pushCatalogHandCard(me, "Undulating Witness", "Creature — Serpent", oracleUndulatingWit)
	rfCrFMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("landcycle: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("no basic land reached the hand")
	}
}

// --- Unflinching Hortimancer ----------------------------------------

func TestUnflinchingHortimancerGrowsOncePerLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Unflinching Hortimancer", "Creature — Human Cleric", oracleHortimancer, 2, 1)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (GainLife{Player: me.ID, Amount: 3}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := rfCrFCounters(g, id); n != 1 {
		t.Fatalf("%d counters after a 3-life gain, want 1", n)
	}
}

// --- Vinelasher Adept -----------------------------------------------

func TestVinelasherAdeptPutsThreeCountersOnTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	adept := rfCrFCast(t, g, "Vinelasher Adept", "Creature — Rhino Soldier", oracleVinelasher, 2, 4)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := rfCrFCounters(g, bear); n != 3 {
		t.Fatalf("%d counters on the bear, want 3", n)
	}
	if n := rfCrFCounters(g, adept); n != 0 {
		t.Fatalf("%d counters on the Adept, want 0", n)
	}
	if !rfCrFHasAbility(t, g, adept, "reach") {
		t.Fatal("no reach")
	}
}

// --- Wrecking Gecko -------------------------------------------------

func TestWreckingGeckoPumpsAndGainsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Wrecking Gecko", "Artifact Creature — Lizard Construct", oracleWreckingGecko, 5, 5)
	rfCrFMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}{G}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := rfCrFPower(t, g, id), rfCrFToughness(t, g, id); p != 9 || tough != 9 {
		t.Fatalf("P/T %d/%d, want 9/9", p, tough)
	}
	if !rfCrFHasAbility(t, g, id, "trample") {
		t.Fatal("no trample after the pump")
	}
}

// --- Vraska, the Cutting Glare --------------------------------------

func rfCrFLands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		rfCrFPush(g, owner, "Forest", "Basic Land — Forest", "", 0, 0)
	}
}

func TestVraskaTheCuttingGlareDestroysAndGivesATreasureWithSixLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFLands(g, me.ID, 6)
	victim := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrFCast(t, g, "Vraska, the Cutting Glare", "Legendary Creature — Gorgon Assassin", oracleVraskaGlare, 4, 4)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the target survived")
	}
	treasure := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" {
			treasure = true
			if c.Controller != opp.ID {
				t.Fatalf("the Treasure belongs to %s, want the destroyed permanent's controller", c.Controller)
			}
		}
	}
	if !treasure {
		t.Fatal("no Treasure was created")
	}
}

func TestVraskaTheCuttingGlareDoesNothingBelowSixLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFLands(g, me.ID, 5)
	victim := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrFCast(t, g, "Vraska, the Cutting Glare", "Legendary Creature — Gorgon Assassin", oracleVraskaGlare, 4, 4)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a target prompt opened with only five lands")
	}
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the opponent's creature was destroyed with five lands")
	}
	if n := b16CountNamed(g, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures with five lands, want 0", n)
	}
}

// --- Vraska, Soul of Stone ------------------------------------------

func TestVraskaSoulOfStoneMakesSculptureTreasuresAndGivesArtifactCreaturesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrFPush(g, me.ID, "Vraska, Soul of Stone", "Legendary Creature — Gorgon Wizard", oracleVraskaSoul, 3, 3)
	golem := rfCrFPush(g, me.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	rfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Sculpture Treasure"); n != 0 {
		t.Fatalf("a creature spell made %d Sculpture Treasures", n)
	}
	rfCrFCastNoncreature(t, g)
	if n := b16CountNamed(g, "Sculpture Treasure"); n != 1 {
		t.Fatalf("%d Sculpture Treasures after a noncreature spell, want 1", n)
	}
	var sculpture uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sculpture Treasure" {
			sculpture = c.InstanceID
			if !c.IsCreature() || !c.IsArtifact() || c.Power != 1 || c.Toughness != 1 {
				t.Fatalf("token is %q %d/%d", c.TypeLine, c.Power, c.Toughness)
			}
		}
	}
	if !rfCrFHasAbility(t, g, golem, "vigilance") || !rfCrFHasAbility(t, g, sculpture, "vigilance") {
		t.Fatal("artifact creatures lack vigilance")
	}
	if rfCrFHasAbility(t, g, bear, "vigilance") {
		t.Fatal("a non-artifact creature has vigilance")
	}
}

// --- Winter, Team Player --------------------------------------------

func TestWinterTeamPlayerPumpsYourCreaturesOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFPush(g, me.ID, "Winter, Team Player", "Legendary Creature — Human Warrior", oracleWinterTeam, 3, 3)
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	passPriorityAroundTable(t, g)
	if p := rfCrFPower(t, g, mine); p != 2 {
		t.Fatalf("a creature spell pumped: power %d", p)
	}
	rfCrFCastNoncreature(t, g)
	if p := rfCrFPower(t, g, mine); p != 3 {
		t.Fatalf("my creature's power %d, want 3", p)
	}
	if p := rfCrFPower(t, g, theirs); p != 2 {
		t.Fatalf("the opponent's creature's power %d, want 2", p)
	}
}

// --- Yoshimaru, Beloved Companion -----------------------------------

func TestYoshimaruBelovedCompanionAddsACounterToYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFPush(g, me.ID, "Yoshimaru, Beloved Companion", "Legendary Creature — Dog", oracleYoshimaruBelov, 2, 2)
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		for _, id := range []uuid.UUID{mine, theirs} {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				t.Fatal(err)
			}
		}
	})
	if n := rfCrFCounters(g, mine); n != 2 {
		t.Fatalf("%d counters on my creature, want 2", n)
	}
	if n := rfCrFCounters(g, theirs); n != 1 {
		t.Fatalf("%d counters on the opponent's creature, want 1", n)
	}
}

func TestYoshimaruBelovedCompanionActivationTargetsLegendaryCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yosh := rfCrFPush(g, me.ID, "Yoshimaru, Beloved Companion", "Legendary Creature — Dog", oracleYoshimaruBelov, 2, 2)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrFMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, yosh, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("a nonlegendary creature was accepted as the target")
	}
	b16Activate(t, g, me.ID, yosh, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: yosh}},
	})
	passPriorityAroundTable(t, g)
	if n := rfCrFCounters(g, yosh); n != 2 {
		t.Fatalf("%d counters on Yoshimaru, want 2 (one, plus his own replacement)", n)
	}
}

// --- Yoshimaru, Scrappy Stray ---------------------------------------

func TestYoshimaruScrappyStrayMakesYourCreatureFightTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := rfCrFPush(g, opp.ID, "Mouse", "Creature — Mouse", "", 1, 1)
	rfCrFCast(t, g, "Yoshimaru, Scrappy Stray", "Legendary Creature — Dog", oracleYoshimaruStray, 1, 1)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, mine)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature survived the fight")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("my creature died to a 1/1")
	}
}

func TestYoshimaruScrappyStrayCannotPickItselfAsTheFighter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	yosh := rfCrFCast(t, g, "Yoshimaru, Scrappy Stray", "Legendary Creature — Dog", oracleYoshimaruStray, 1, 1)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: yosh}); err == nil {
		t.Fatal("Yoshimaru was accepted as the creature that fights")
	}
	_ = mine
}

// --- Verdant Kraken -------------------------------------------------

func TestVerdantKrakenMakesAForestTentacleOnEveryPlayersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrFPush(g, me.ID, "Verdant Kraken", "Creature — Plant Kraken", oracleVerdantKraken, 6, 6)
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Forest Tentacle"); n != 1 {
		t.Fatalf("%d Forest Tentacles after an opponent's upkeep, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Forest Tentacle" {
			continue
		}
		if c.Controller != me.ID {
			t.Fatalf("the Tentacle is controlled by %s, want the Kraken's controller", c.Controller)
		}
		if !c.IsLand() || !c.IsCreature() || c.Power != 3 || c.Toughness != 3 {
			t.Fatalf("token is %q %d/%d", c.TypeLine, c.Power, c.Toughness)
		}
	}
}

// --- Yargle ---------------------------------------------------------

func TestYargleIsRegisteredAsAFullyAutomatedVanilla(t *testing.T) {
	spec, ok := Lookup(oracleYargle)
	if !ok {
		t.Fatal("Yargle is not in the catalog")
	}
	if spec.Completeness != CompletenessFull {
		t.Fatalf("completeness %v", spec.Completeness)
	}
}
