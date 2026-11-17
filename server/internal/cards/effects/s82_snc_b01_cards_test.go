package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	sncB01MidnightAssassin = "f394d2bf-89aa-44c1-b473-8c7a8874a9a3"
	sncB01FleetfootDancer  = "8f606acf-db16-42cd-8907-dad88c4a1d54"
	sncB01CrookedCustodian = "9c39bc53-52d3-4610-920e-5607a625b01f"
	sncB01WreckingCrew     = "533c4956-87f8-4bae-8924-166fc2e9ec1f"
	sncB01ChromeCat        = "0e2f8758-7457-4b54-ae61-a9c68cea4a23"
	sncB01JewelThief       = "470f0bbf-a0c9-4216-96f0-d400dac17aef"
	sncB01Glittermonger    = "184c763c-b52a-471c-bd87-d01729ebe5e4"
	sncB01BackupAgent      = "df9a3fe0-aca8-4adb-9ef6-45616025925b"
	sncB01SpeakeasyServer  = "4e08e07f-ca30-4e34-b207-d8fe62c03777"
	sncB01HighRiseSawjack  = "0049438b-e37e-4c57-820e-f2ddf3322170"
	sncB01ObscuraInitiate  = "bdc3a563-86e0-4457-b00b-85c03e62a930"
	sncB01CabarettiInit    = "5e0d877c-6ce3-462b-9746-7fc921c93be9"
	sncB01RiveteersInit    = "c756d786-9aff-4d2e-9c06-6ae01d5fca2e"
)

func TestSncB01CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		sncB01MidnightAssassin: "Midnight Assassin", sncB01FleetfootDancer: "Fleetfoot Dancer",
		sncB01CrookedCustodian: "Crooked Custodian", sncB01WreckingCrew: "Wrecking Crew",
		sncB01ChromeCat: "Chrome Cat", sncB01JewelThief: "Jewel Thief",
		sncB01Glittermonger: "Glittermonger", sncB01BackupAgent: "Backup Agent",
		sncB01SpeakeasyServer: "Speakeasy Server", sncB01HighRiseSawjack: "High-Rise Sawjack",
		sncB01ObscuraInitiate: "Obscura Initiate", sncB01CabarettiInit: "Cabaretti Initiate",
		sncB01RiveteersInit: "Riveteers Initiate",
	}
	for oracle, name := range want {
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

func TestSncB01PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Midnight Assassin", "Creature — Vampire Assassin", sncB01MidnightAssassin, []string{"flying", "deathtouch"}},
		{"Fleetfoot Dancer", "Creature — Elf Druid", sncB01FleetfootDancer, []string{"trample", "lifelink", "haste"}},
		{"Wrecking Crew", "Creature — Human Warrior", sncB01WreckingCrew, []string{"reach", "trample"}},
		{"Jewel Thief", "Creature — Cat Rogue", sncB01JewelThief, []string{"vigilance", "trample"}},
		{"Speakeasy Server", "Creature — Bird Citizen", sncB01SpeakeasyServer, []string{"flying"}},
		{"High-Rise Sawjack", "Creature — Elf Citizen", sncB01HighRiseSawjack, []string{"reach"}},
		{"Obscura Initiate", "Creature — Bird Citizen", sncB01ObscuraInitiate, []string{"flying"}},
	} {
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.kws {
			if !hasKeywordOf(g, id, kw) {
				t.Errorf("%s has no %s", tc.name, kw)
			}
		}
	}
}

func TestSncB01CrookedCustodianEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Crooked Custodian", "Creature — Ogre Rogue", sncB01CrookedCustodian)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Crooked Custodian did not resolve onto the battlefield")
	}
	if !c.Tapped {
		t.Error("Crooked Custodian entered untapped")
	}
}

func TestSncB01ChromeCatScriesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top", "Next")
	castAndResolveCreature(t, g, "Chrome Cat", "Artifact Creature — Cat", sncB01ChromeCat)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("entering queued no scry prompt")
	}
}

func TestSncB01JewelThiefMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := countTreasuresControlledBy(g, me.ID)
	castAndResolveCreature(t, g, "Jewel Thief", "Creature — Cat Rogue", sncB01JewelThief)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 1 {
		t.Errorf("made %d Treasures, want 1", got)
	}
}

func TestSncB01GlittermongerTapsForATreasureOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Glittermonger", "Creature — Elf Rogue", sncB01Glittermonger, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestSncB01BackupAgentPutsACounterOnTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Backup Agent", "Creature — Human Citizen", sncB01BackupAgent)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, foe); p != 3 || tough != 3 {
		t.Errorf("target P/T = %d/%d, want 3/3", p, tough)
	}
}

func TestSncB01SpeakeasyServerGainsOnePerOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := me.Life
	castAndResolveCreature(t, g, "Speakeasy Server", "Creature — Bird Citizen", sncB01SpeakeasyServer)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 (the Server itself and their creature do not count)", life, me.Life)
	}
}

func TestSncB01HighRiseSawjackPumpsOnlyBlockingAFlier(t *testing.T) {
	for _, flier := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		saw := b12Push(g, opp.ID, "High-Rise Sawjack", "Creature — Elf Citizen", sncB01HighRiseSawjack, 2, 3)
		kws := []string(nil)
		if flier {
			kws = []string{"flying"}
		}
		atk := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Bird",
			Power: 2, Toughness: 2, Keywords: kws, Owner: me.ID, Controller: me.ID,
		})
		attackIntoBlocks(t, g, opp.ID, atk)
		if err := g.DeclareBlocker(saw, atk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		lockInBlocks(t, g)
		passPriorityAroundTable(t, g)
		want := 2
		if flier {
			want = 4
		}
		if p, tough := eoeCDPT(t, g, saw); p != want || tough != 3 {
			t.Errorf("flier=%v: Sawjack is %d/%d, want %d/3", flier, p, tough, want)
		}
	}
}

func TestSncB01InitiatesGainTheirKeywordForEitherHybridHalf(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, kw string
		pays                       [2]string
	}{
		{"Obscura Initiate", "Creature — Bird Citizen", sncB01ObscuraInitiate, "lifelink", [2]string{"{C}{W}", "{C}{B}"}},
		{"Cabaretti Initiate", "Creature — Raccoon Citizen", sncB01CabarettiInit, "double strike", [2]string{"{C}{C}{R}", "{C}{C}{W}"}},
		{"Riveteers Initiate", "Creature — Lizard Citizen", sncB01RiveteersInit, "deathtouch", [2]string{"{C}{B}", "{C}{G}"}},
	} {
		for _, pay := range tc.pays {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatalf("%s: activated with an empty pool", tc.name)
			}
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatalf("%s: activated with one colourless mana", tc.name)
			}
			if hasKeywordOf(g, id, tc.kw) {
				t.Fatalf("%s: has %s before paying", tc.name, tc.kw)
			}
			g2 := newCatalogGame(t)
			me2 := g2.Seats[0]
			id2 := b12Push(g2, me2.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if err := g2.AddManaForEffect(me2.ID, uuid.Nil, pay); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g2.ActivateCatalogAbility(me2.ID, id2, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("%s with %s: %v", tc.name, pay, err)
			}
			passPriorityAroundTable(t, g2)
			if !hasKeywordOf(g2, id2, tc.kw) {
				t.Errorf("%s paid with %s: no %s", tc.name, pay, tc.kw)
			}
		}
	}
}
