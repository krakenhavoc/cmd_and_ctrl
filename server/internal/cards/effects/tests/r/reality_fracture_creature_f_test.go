package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rfCrFPush seats a permanent on the battlefield under owner.
func rfCrFPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func rfCrFCounters(g *game.Game, id uuid.UUID) int {
	c, ok := effectstest.BattlefieldCard(g, id)
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
	id := effectstest.RfCrFHand(g, "Idle Trick", "Instant", effectstest.RfCrFNoncreatureOracle, "", 0, 0)
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell noncreature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Traxos, Academy Guardian ---------------------------------------

func TestTraxosAcademyGuardianCostsTwoLessAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	const typeLine = "Legendary Artifact Creature — Dragon Construct"
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.OracleTraxosAcademy, typeLine, "{3}{U}"); got != 4 {
		t.Fatalf("Traxos with no spell cast this turn: mana value %d, want 4", got)
	}
	effectstest.RfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.OracleTraxosAcademy, typeLine, "{3}{U}"); got != 4 {
		t.Fatalf("Traxos after a creature spell: mana value %d, want 4", got)
	}
	rfCrFCastNoncreature(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.OracleTraxosAcademy, typeLine, "{3}{U}"); got != 2 {
		t.Fatalf("Traxos after a noncreature spell: mana value %d, want 2", got)
	}
	id := rfCrFPush(g, me.ID, "Traxos, Academy Guardian", typeLine, effectstest.OracleTraxosAcademy, 1, 5)
	for _, kw := range []string{"flying", "vigilance", game.KeywordProwess} {
		if !rfCrFHasAbility(t, g, id, kw) {
			t.Errorf("Traxos lacks %s", kw)
		}
	}
}

func TestTraxosAcademyGuardianProwessPumpsItAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Traxos, Academy Guardian", "Legendary Artifact Creature — Dragon Construct",
		effectstest.OracleTraxosAcademy, 1, 5)
	rfCrFCastNoncreature(t, g)
	if p := rfCrFPower(t, g, id); p != 2 {
		t.Fatalf("power %d after a noncreature spell, want 2 (prowess)", p)
	}
}

// --- Traxos, Scourge Eternal ----------------------------------------

func TestTraxosScourgeEternalUntapsOnlyForAnArtifactOrCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Traxos, Scourge Eternal", "Legendary Artifact Creature — Dragon Construct",
		effectstest.OracleTraxosScourge, 5, 4)
	if err := g.TapCard(id, true); err != nil {
		t.Fatal(err)
	}
	rfCrFCastNoncreature(t, g)
	if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
		t.Fatal("a plain noncreature spell untapped Traxos")
	}
	effectstest.RfCrFCast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Fatal("a creature spell did not untap Traxos")
	}

	if err := g.TapCard(id, true); err != nil {
		t.Fatal(err)
	}
	rock := effectstest.RfCrFHand(g, "Rock", "Artifact", "", "", 0, 0)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Fatal("an artifact spell did not untap Traxos")
	}
}

func TestTraxosScourgeEternalStaysTappedThroughYourUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	traxos := rfCrFPush(g, me.ID, "Traxos, Scourge Eternal", "Legendary Artifact Creature — Dragon Construct",
		effectstest.OracleTraxosScourge, 5, 4)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{traxos, bear} {
		if err := g.TapCard(id, true); err != nil {
			t.Fatal(err)
		}
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOfSeat(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	if c, _ := effectstest.BattlefieldCard(g, traxos); !c.Tapped {
		t.Fatal("Traxos untapped during its controller's untap step")
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Fatal("the control creature should have untapped")
	}
}

// --- Undulating Witness ---------------------------------------------

func TestUndulatingWitnessPumpsPlusOneMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Undulating Witness", "Creature — Serpent", effectstest.OracleUndulatingWit, 3, 5)
	rfCrFMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := rfCrFPower(t, g, id), rfCrFToughness(t, g, id); p != 4 || tough != 4 {
		t.Fatalf("P/T %d/%d after one activation, want 4/4", p, tough)
	}
	if !rfCrFHasAbility(t, g, id, "flying") {
		t.Fatal("no flying")
	}
}

func TestUndulatingWitnessBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.ToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Undulating Witness", "Creature — Serpent", effectstest.OracleUndulatingWit)
	rfCrFMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("landcycle: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Unflinching Hortimancer", "Creature — Human Cleric", effectstest.OracleHortimancer, 2, 1)
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (effects.GainLife{Player: me.ID, Amount: 3}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := rfCrFCounters(g, id); n != 1 {
		t.Fatalf("%d counters after a 3-life gain, want 1", n)
	}
}

// --- Vinelasher Adept -----------------------------------------------

func TestVinelasherAdeptPutsThreeCountersOnTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	adept := effectstest.RfCrFCast(t, g, "Vinelasher Adept", "Creature — Rhino Soldier", effectstest.OracleVinelasher, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrFPush(g, me.ID, "Wrecking Gecko", "Artifact Creature — Lizard Construct", effectstest.OracleWreckingGecko, 5, 5)
	rfCrFMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}{G}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFLands(g, me.ID, 6)
	victim := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.RfCrFCast(t, g, "Vraska, the Cutting Glare", "Legendary Creature — Gorgon Assassin", effectstest.OracleVraskaGlare, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFLands(g, me.ID, 5)
	victim := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.RfCrFCast(t, g, "Vraska, the Cutting Glare", "Legendary Creature — Gorgon Assassin", effectstest.OracleVraskaGlare, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a target prompt opened with only five lands")
	}
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the opponent's creature was destroyed with five lands")
	}
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures with five lands, want 0", n)
	}
}

// --- Vraska, Soul of Stone ------------------------------------------

func TestVraskaSoulOfStoneMakesSculptureTreasuresAndGivesArtifactCreaturesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrFPush(g, me.ID, "Vraska, Soul of Stone", "Legendary Creature — Gorgon Wizard", effectstest.OracleVraskaSoul, 3, 3)
	golem := rfCrFPush(g, me.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	effectstest.RfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Sculpture Treasure"); n != 0 {
		t.Fatalf("a creature spell made %d Sculpture Treasures", n)
	}
	rfCrFCastNoncreature(t, g)
	if n := effectstest.B16CountNamed(g, "Sculpture Treasure"); n != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFPush(g, me.ID, "Winter, Team Player", "Legendary Creature — Human Warrior", effectstest.OracleWinterTeam, 3, 3)
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.RfCrFCast(t, g, "Bear Cub", "Creature — Bear", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFPush(g, me.ID, "Yoshimaru, Beloved Companion", "Legendary Creature — Dog", effectstest.OracleYoshimaruBelov, 2, 2)
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrFPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		for _, id := range []uuid.UUID{mine, theirs} {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yosh := rfCrFPush(g, me.ID, "Yoshimaru, Beloved Companion", "Legendary Creature — Dog", effectstest.OracleYoshimaruBelov, 2, 2)
	bear := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrFMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, yosh, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Fatal("a nonlegendary creature was accepted as the target")
	}
	effectstest.B16Activate(t, g, me.ID, yosh, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: yosh}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := rfCrFCounters(g, yosh); n != 2 {
		t.Fatalf("%d counters on Yoshimaru, want 2 (one, plus his own replacement)", n)
	}
}

// --- Yoshimaru, Scrappy Stray ---------------------------------------

func TestYoshimaruScrappyStrayMakesYourCreatureFightTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := rfCrFPush(g, opp.ID, "Mouse", "Creature — Mouse", "", 1, 1)
	effectstest.RfCrFCast(t, g, "Yoshimaru, Scrappy Stray", "Legendary Creature — Dog", effectstest.OracleYoshimaruStray, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature survived the fight")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("my creature died to a 1/1")
	}
}

func TestYoshimaruScrappyStrayCannotPickItselfAsTheFighter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := rfCrFPush(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	yosh := effectstest.RfCrFCast(t, g, "Yoshimaru, Scrappy Stray", "Legendary Creature — Dog", effectstest.OracleYoshimaruStray, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrFPush(g, me.ID, "Verdant Kraken", "Creature — Plant Kraken", effectstest.OracleVerdantKraken, 6, 6)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Forest Tentacle"); n != 1 {
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
	spec, ok := effects.Lookup(effectstest.OracleYargle)
	if !ok {
		t.Fatal("Yargle is not in the catalog")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Fatalf("completeness %v", spec.Completeness)
	}
}
