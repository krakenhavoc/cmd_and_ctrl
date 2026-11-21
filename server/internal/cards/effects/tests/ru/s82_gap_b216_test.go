package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB216EarthquakeOracle = "1903200d-0350-48a9-ac7c-b851f7df647f"
	gapB216ErinisOracle     = "1b76a54d-5a9f-4b1b-a075-4b4c4de8ba0a"
	gapB216LinnormOracle    = "1bf04a2a-7f96-412c-9d91-034bef26e1b6"
	gapB216DeflectionOracle = gapB216LinnormOracle + "#1"
	gapB216DurnanOracle     = "21d3858f-441d-4bb2-888e-8b147ecad7e9"
	gapB216SwineOracle      = "411cf990-972c-4ee3-9ae4-98602f269407"
	gapB216MuralistsOracle  = "4dfad4ec-9695-4f3a-8269-4cc00c26204c"
	gapB216SwarmOracle      = "61f09fcc-11cd-4999-a43a-54488b19861d"
	gapB216BadgerOracle     = "96aab05a-49bc-418a-9980-f3322955de1e"
	gapB216DigDeepOracle    = gapB216BadgerOracle + "#1"
	gapB216DruidOracle      = "acf54a85-0e9e-43fb-99d9-c223c02f13c4"
	gapB216CircleOracle     = "b6cc351e-6afd-4427-998d-e06e430cd0e6"
	gapB216EmeraldOracle    = "c5632f18-7904-4a29-82be-6a6ef8e85e18"
	gapB216WaveOracle       = gapB216EmeraldOracle + "#1"
	gapB216EttercapOracle   = "e0b5b3af-af89-4b91-8b55-947ec095d6f2"
	gapB216WebShotOracle    = gapB216EttercapOracle + "#1"
)

func TestGapB216EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB216EarthquakeOracle: "Earthquake Dragon",
		gapB216ErinisOracle:     "Erinis, Gloom Stalker",
		gapB216LinnormOracle:    "Dread Linnorm",
		gapB216DeflectionOracle: "Scale Deflection",
		gapB216DurnanOracle:     "Durnan of the Yawning Portal",
		gapB216SwineOracle:      "Carefree Swinemaster",
		gapB216MuralistsOracle:  "Draconic Muralists",
		gapB216SwarmOracle:      "Cloakwood Swarmkeeper",
		gapB216BadgerOracle:     "Colossal Badger",
		gapB216DigDeepOracle:    "Dig Deep",
		gapB216DruidOracle:      "Druid of the Emerald Grove",
		gapB216CircleOracle:     "Circle of the Land Druid",
		gapB216EmeraldOracle:    "Emerald Dragon",
		gapB216WaveOracle:       "Dissonant Wave",
		gapB216EttercapOracle:   "Ettercap",
		gapB216WebShotOracle:    "Web Shot",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB216Hand(me *game.Player, name, typeLine, oracle, cost string, p, tgh int) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: p, Toughness: tgh, Owner: me.ID, Controller: me.ID})
	return id
}

func gapB216Mana(t *testing.T, g *game.Game, me *game.Player, generic int, colored string) {
	t.Helper()
	effectstest.FloatMana(t, g, me, strings.Repeat("{C}", generic)+colored)
}

// --- Earthquake Dragon ---------------------------------------------------

func TestGapB216EarthquakeDragonCostsLessByTheTotalManaValueOfYourDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// A 6-mana Dragon of mine and a 9-mana Dragon of an opponent's: only mine counts.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Dragon", TypeLine: "Creature — Dragon",
		ManaCost: "{4}{R}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon",
		ManaCost: "{7}{R}{R}", Power: 7, Toughness: 7, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	id := gapB216Hand(me, "Earthquake Dragon", "Creature — Elemental Dragon", gapB216EarthquakeOracle, "{14}{G}", 10, 10)
	gapB216Mana(t, g, me, 7, "{G}") // {14}{G} less 6 is {8}{G}; seven generic is one short
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for one mana too few")
	}
	gapB216Mana(t, g, me, 1, "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {8}{G}: %v", err)
	}
}

func TestGapB216EarthquakeDragonReturnsFromTheGraveyardForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Earthquake Dragon", TypeLine: "Creature — Elemental Dragon", OracleID: gapB216EarthquakeOracle,
		ManaCost: "{14}{G}", Power: 10, Toughness: 10, Owner: me.ID, Controller: me.ID})
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	gapB216Mana(t, g, me, 2, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the Dragon did not return to hand")
	}
	if effectstest.B12ZoneOf(g, land) == game.ZoneBattlefield {
		t.Error("the land was not sacrificed")
	}
}

// --- Erinis --------------------------------------------------------------

func TestGapB216ErinisReturnsALandFromYourGraveyardWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	erinis := effectstest.B12Push(g, me.ID, "Erinis, Gloom Stalker", "Legendary Creature — Halfling Ranger", gapB216ErinisOracle, 3, 3)
	land := effectstest.B17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	spell := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.DeclareAttack(t, g, opp.ID, erinis)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, land); z != game.ZoneBattlefield {
		t.Errorf("the land is in %s, want the battlefield", z)
	}
	if z := effectstest.B12ZoneOf(g, spell); z != game.ZoneGraveyard {
		t.Errorf("a nonland card moved to %s", z)
	}
}

// --- Dread Linnorm // Scale Deflection -----------------------------------

func TestGapB216LinnormCantBeBlockedByPowerThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	linnorm := effectstest.B12Push(g, me.ID, "Dread Linnorm", "Creature — Snake Dragon", gapB216LinnormOracle, 7, 6)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, linnorm)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, linnorm); err == nil {
		t.Error("a power-3 creature blocked the Linnorm")
	}
	if err := g.DeclareBlocker(big, linnorm); err != nil {
		t.Errorf("a power-4 creature could not block: %v", err)
	}
}

func TestGapB216ScaleDeflectionCountersUntapsAndGivesHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, bear)
	effectstest.B18CastFromHand(t, g, me, "Scale Deflection", "Instant — Adventure", gapB216DeflectionOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if effectstest.B13Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the creature lacks hexproof")
	}
}

// --- Carefree Swinemaster ------------------------------------------------

func TestGapB216SwinemasterMakesATappedAttackingBoarWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		swine := effectstest.B12Push(g, me.ID, "Carefree Swinemaster", "Creature — Gnome Ranger", gapB216SwineOracle, 1, 4)
		effectstest.DeclareAttack(t, g, opp.ID, swine)
		effectstest.PassPriorityAroundTable(t, g)
		gapB216Mana(t, g, me, 1, "{G}")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		n := effectstest.B16CountNamed(g, "Boar")
		if !pay && n != 0 {
			t.Errorf("declined: %d Boars, want none", n)
		}
		if pay {
			if n != 1 {
				t.Fatalf("paid: %d Boars, want 1", n)
			}
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Boar" && (!c.Tapped || c.AttackingTarget != opp.ID || c.Power != 2 || c.Toughness != 2) {
					t.Errorf("Boar %d/%d tapped=%v attacking=%v, want a tapped 2/2 attacking %v", c.Power, c.Toughness, c.Tapped, c.AttackingTarget, opp.ID)
				}
			}
		}
	}
}

// --- Draconic Muralists --------------------------------------------------

func TestGapB216MuralistsSearchesForADragonWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.B12Push(g, me.ID, "Draconic Muralists", "Creature — Dragon Bard", gapB216MuralistsOracle, 4, 3)
	dragon := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: dragon, Name: "Library Dragon", TypeLine: "Creature — Dragon", Owner: me.ID, Controller: me.ID})
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Library Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.B18Kill(t, g, m)
	effectstest.PassPriorityAroundTable(t, g)
	// The search prompt: take the Dragon.
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceSearchLibrary {
			effectstest.BeginSearchIfAsked(t, g, c)
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{dragon}); err != nil {
				t.Fatalf("ResolveSearch: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dragon) {
		t.Error("the Dragon did not reach the hand")
	}
	if me.Hand.Contains(other) {
		t.Error("a non-Dragon was fetched")
	}
}

// --- Cloakwood Swarmkeeper -----------------------------------------------

func TestGapB216SwarmkeeperGetsOneCounterPerBatchOfYourTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	k := effectstest.B12Push(g, me.ID, "Cloakwood Swarmkeeper", "Creature — Elf Ranger", gapB216SwarmOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, k, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after two tokens at once, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, k, game.CounterPlusOne); n != 1 {
		t.Errorf("an opponent's token gave a counter (%d)", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, k, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after another token, want 2", n)
	}
}

// --- Circle of the Land Druid --------------------------------------------

func TestGapB216CircleMillsFourWhenYouSayYesAndReturnsALandWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B13PlayAs(t, g, 0, "Circle of the Land Druid", "Creature — Gnome Druid", gapB216CircleOracle)
	grave := me.Graveyard.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - grave; got != 4 {
		t.Errorf("graveyard grew by %d, want 4", got)
	}
	land := effectstest.B17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	effectstest.B18Kill(t, g, id)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
}

func TestGapB216CircleMayDeclineToMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B13PlayAs(t, g, 0, "Circle of the Land Druid", "Creature — Gnome Druid", gapB216CircleOracle)
	grave := me.Graveyard.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != grave {
		t.Errorf("graveyard %d -> %d after declining", grave, me.Graveyard.Size())
	}
}

// --- Colossal Badger // Dig Deep -----------------------------------------

func TestGapB216BadgerGainsThreeLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.B13PlayAs(t, g, 0, "Colossal Badger", "Creature — Badger", gapB216BadgerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestGapB216DigDeepPutsACounterForEachCreatureCardMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Top four cards: two creatures, a land and an instant.
	for _, c := range []struct{ name, typeLine string }{{"Instant", "Instant"}, {"Forest", "Basic Land — Forest"}, {"Elf", "Creature — Elf"}, {"Ogre", "Creature — Ogre"}} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.typeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.B18CastFromHand(t, g, me, "Dig Deep", "Sorcery — Adventure", gapB216DigDeepOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (the Elf and the Ogre)", n)
	}
}

// --- Druid of the Emerald Grove ------------------------------------------

func TestGapB216DruidPutsTheBasicsWhereTheD20Says(t *testing.T) {
	for i := 0; i < 40; i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		var basics []uuid.UUID
		for _, name := range []string{"Forest", "Island"} {
			id := uuid.New()
			basics = append(basics, id)
			me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Basic Land — " + name, Owner: me.ID, Controller: me.ID})
		}
		effectstest.B13PlayAs(t, g, 0, "Druid of the Emerald Grove", "Creature — Dwarf Druid", gapB216DruidOracle)
		effectstest.PassPriorityAroundTable(t, g)
		for step := 0; step < 4; step++ {
			if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); p != nil {
				if err := g.ResolveSearchLibrary(p.ID, me.ID, basics); err != nil {
					t.Fatalf("ResolveSearch: %v", err)
				}
			}
			if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); p != nil {
				if err := g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards[:1]); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		rolls := effectstest.RandomEvents(g, game.EventRollDie)
		if len(rolls) != 1 {
			t.Fatalf("%d die rolls, want 1", len(rolls))
		}
		roll := rolls[0].Amount
		onBF, inHand := 0, 0
		for _, id := range basics {
			switch effectstest.B12ZoneOf(g, id) {
			case game.ZoneBattlefield:
				onBF++
				if !effectstest.B13Tapped(t, g, id) {
					t.Errorf("roll %d: a land entered untapped", roll)
				}
			case game.ZoneHand:
				inHand++
			}
		}
		wantBF := 0
		switch {
		case roll >= 20:
			wantBF = 2
		case roll >= 10:
			wantBF = 1
		}
		if onBF != wantBF || inHand != 2-wantBF {
			t.Errorf("roll %d: %d on the battlefield, %d in hand, want %d and %d", roll, onBF, inHand, wantBF, 2-wantBF)
		}
	}
}

// --- Dissonant Wave and Web Shot -----------------------------------------

func TestGapB216DissonantWaveOnlyTargetsAbilitiesFromNoncreatureSources(t *testing.T) {
	spec, ok := effects.Lookup(gapB216WaveOracle)
	if !ok || spec.Targets == nil || !spec.Targets.Abilities {
		t.Fatalf("Dissonant Wave targets %+v, want an ability on the stack", spec.Targets)
	}
}

func TestGapB216WebShotDestroysOnlyACreatureWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.B12Creature(g, opp.ID, "Flier", "Creature — Bird", 1, 1)
	ground := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	if err := g.CastSpell(me.ID, gapB216Hand(me, "Web Shot", "Instant — Adventure", gapB216WebShotOracle, "", 0, 0),
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ground}}}); err == nil {
		t.Error("Web Shot targeted a creature without flying")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B18CastFromHand(t, g, me, "Web Shot", "Instant — Adventure", gapB216WebShotOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flier}}})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, flier); z != game.ZoneGraveyard {
		t.Errorf("the flier is in %s, want the graveyard", z)
	}
}

// --- Durnan of the Yawning Portal ----------------------------------------

func TestGapB216DurnanExilesACreatureFromTheTopFourAndLetsYouCastItWithUndaunted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	durnan := effectstest.B12Push(g, me.ID, "Durnan of the Yawning Portal", "Legendary Creature — Human Warrior", gapB216DurnanOracle, 3, 3)
	var ogre uuid.UUID
	var others []uuid.UUID
	for _, c := range []struct{ name, typeLine, cost string }{
		{"Instant", "Instant", "{U}"}, {"Forest", "Basic Land — Forest", ""}, {"Ogre", "Creature — Ogre", "{4}{R}"}, {"Elf", "Creature — Elf", "{G}"},
	} {
		id := uuid.New()
		if c.name == "Ogre" {
			ogre = id
		} else {
			others = append(others, id)
		}
		me.Library.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	}
	size := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, durnan)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("not asked which creature card to exile")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d cards offered, want the two creature cards", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ogre}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, ogre); z != game.ZoneExile {
		t.Fatalf("the Ogre is in %s, want exile", z)
	}
	if got := me.Library.Size(); got != size-1 {
		t.Errorf("library %d -> %d, want one card gone (the rest go back to the bottom)", size, got)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("not asked to order the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, others, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	// {4}{R} costs {1} less for each of the three opponents: {1}{R}.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast the Ogre for one mana, below its undaunted price of {1}{R}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the exiled Ogre for {1}{R}: %v", err)
	}
}
