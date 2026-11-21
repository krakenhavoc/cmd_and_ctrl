package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mir_b1_cards_test.go — Mirage, batch 1 (#5633). Each test plays the
// real catalog entry and asserts what its text says.

const (
	mirB1ArchersOracle      = "b5867796-5898-4a75-a75a-c802407ab338"
	mirB1FetidHorrorOracle  = "4863e1cf-3ef2-4c2d-a0ea-b73ca60c660d"
	mirB1BreathstealOracle  = "17d18e2f-2d8d-449e-9356-a7351bf0f8c2"
	mirB1DirtwaterOracle    = "0126f620-1dbf-405c-b5cf-f48b7ef02d32"
	mirB1RestlessDeadOracle = "5315ba59-ca80-498e-b21a-ed87e6a4c955"
	mirB1PearlDragonOracle  = "f4e0e2e0-8d64-40a3-92da-3951a696b5f3"
	mirB1EmissaryOracle     = "8fb44b03-017f-4cf3-a79b-8f6cc33570ca"
	mirB1PatagiaOracle      = "85308fb8-e5a1-4ff1-8606-283cf2056895"
	mirB1IgneousOracle      = "7c4408eb-b3e2-407f-9018-197787edeae7"
	mirB1HarmattanOracle    = "84b1d4e1-ca8a-4eed-b5c4-364876938aee"
	mirB1BenthicOracle      = "7e8d40fd-c324-4ec2-897d-b46368b050e6"
	mirB1NettletoothOracle  = "f1d300b6-f9cf-40a9-8520-d78cd7d813cf"
	mirB1CrocodileOracle    = "e11926bf-ca14-4e1e-aca5-9ffedb91868c"
	mirB1TeferisDrakeOracle = "3605c5ac-3b35-41a5-9266-286fb1d11716"
)

func TestMirB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mirB1ArchersOracle:      "Femeref Archers",
		mirB1FetidHorrorOracle:  "Fetid Horror",
		mirB1BreathstealOracle:  "Breathstealer",
		mirB1DirtwaterOracle:    "Dirtwater Wraith",
		mirB1RestlessDeadOracle: "Restless Dead",
		mirB1PearlDragonOracle:  "Pearl Dragon",
		mirB1EmissaryOracle:     "Wildfire Emissary",
		mirB1PatagiaOracle:      "Patagia Golem",
		mirB1IgneousOracle:      "Igneous Golem",
		mirB1HarmattanOracle:    "Harmattan Efreet",
		mirB1BenthicOracle:      "Benthic Djinn",
		mirB1NettletoothOracle:  "Nettletooth Djinn",
		mirB1CrocodileOracle:    "Sandbar Crocodile",
		mirB1TeferisDrakeOracle: "Teferi's Drake",
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

func TestMirB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Dirtwater Wraith", mirB1DirtwaterOracle, []string{"swampwalk"}},
		{"Pearl Dragon", mirB1PearlDragonOracle, []string{"flying"}},
		{"Wildfire Emissary", mirB1EmissaryOracle, []string{"protection from white"}},
		{"Harmattan Efreet", mirB1HarmattanOracle, []string{"flying"}},
		{"Benthic Djinn", mirB1BenthicOracle, []string{"islandwalk"}},
		{"Sandbar Crocodile", mirB1CrocodileOracle, []string{game.KeywordPhasing}},
		{"Teferi's Drake", mirB1TeferisDrakeOracle, []string{"flying", game.KeywordPhasing}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.MirB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- Femeref Archers -----------------------------------------------------------

func TestMirB1FemerefArchersShootOnlyAnAttackingFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	archers := effectstest.MirB1Push(g, me.ID, "Femeref Archers", "Creature — Human Archer", mirB1ArchersOracle, 2, 2)
	flier := effectstest.MirB1Push(g, opp.ID, "Flier", "Creature — Bird", "", 3, 5, "flying")
	idleFlier := effectstest.MirB1Push(g, opp.ID, "Idle Flier", "Creature — Bird", "", 3, 5, "flying")
	walker := effectstest.MirB1Push(g, opp.ID, "Walker", "Creature — Ogre", "", 3, 5)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{flier, walker} {
			if c := effectstest.BattlefieldCardFor(g, id); c != nil {
				c.AttackingTarget = me.ID
			}
		}
	})
	for _, bad := range []uuid.UUID{idleFlier, walker} {
		if err := effectstest.MirB1Activate(g, me.ID, archers, 0, effectstest.CardTarget(bad)...); err == nil {
			t.Fatal("a flier that isn't attacking, or an attacker without flying, was accepted as the target")
		}
	}
	if err := effectstest.MirB1Activate(g, me.ID, archers, 0, effectstest.CardTarget(flier)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flier).DamageMarked; got != 4 {
		t.Errorf("the attacking flier has %d damage, want 4", got)
	}
	if !effectstest.FindCardAnywhere(t, g, archers).Tapped {
		t.Error("Femeref Archers did not tap")
	}
}

// --- the pumps -------------------------------------------------------------------

func TestMirB1SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		p, tough           int
		wantP, wantT       int
	}{
		{"Fetid Horror", mirB1FetidHorrorOracle, "{B}", 1, 2, 2, 3},
		{"Breathstealer", mirB1BreathstealOracle, "{B}", 2, 2, 3, 1},
		{"Dirtwater Wraith", mirB1DirtwaterOracle, "{B}", 1, 3, 2, 3},
		{"Pearl Dragon", mirB1PearlDragonOracle, "{C}{W}", 4, 4, 4, 5},
		{"Wildfire Emissary", mirB1EmissaryOracle, "{C}{R}", 2, 4, 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.MirB1Push(g, me.ID, tc.name, "Creature", tc.oracle, tc.p, tc.tough)
			effectstest.DcMana(t, g, me.ID, tc.mana)
			if err := effectstest.MirB1Activate(g, me.ID, id, 0); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d after one activation, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestMirB1BreathstealerDiesAtZeroToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MirB1Push(g, me.ID, "Breathstealer", "Creature — Nightstalker", mirB1BreathstealOracle, 2, 2)
	for i := 0; i < 2; i++ {
		effectstest.DcMana(t, g, me.ID, "{B}")
		if err := effectstest.MirB1Activate(g, me.ID, id, 0); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(id) {
		t.Error("Breathstealer survived at 0 toughness")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Breathstealer is not in its owner's graveyard")
	}
}

// --- the keyword grants ---------------------------------------------------------

func TestMirB1GolemsGainTheirKeywordUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana, keyword string
	}{
		{"Patagia Golem", mirB1PatagiaOracle, "{C}{C}{C}", "flying"},
		{"Igneous Golem", mirB1IgneousOracle, "{C}{C}", "trample"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.MirB1Push(g, me.ID, tc.name, "Artifact Creature — Golem", tc.oracle, 2, 3)
			if effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before the ability", tc.name, tc.keyword)
			}
			effectstest.DcMana(t, g, me.ID, tc.mana)
			if err := effectstest.MirB1Activate(g, me.ID, id, 0); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Errorf("%s did not gain %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestMirB1HarmattanEfreetGivesTargetCreatureFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	efreet := effectstest.MirB1Push(g, me.ID, "Harmattan Efreet", "Creature — Efreet", mirB1HarmattanOracle, 2, 2)
	bear := effectstest.MirB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}{U}{U}")
	if err := effectstest.MirB1Activate(g, me.ID, efreet, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if effectstest.FindCardAnywhere(t, g, efreet).Tapped {
		t.Error("Harmattan Efreet tapped for an ability with no tap symbol")
	}
}

// --- Restless Dead ---------------------------------------------------------------

func TestMirB1RestlessDeadRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.MirB1Push(g, me.ID, "Restless Dead", "Creature — Skeleton", mirB1RestlessDeadOracle, 1, 1)
	effectstest.DcMana(t, g, me.ID, "{B}")
	if err := effectstest.MirB1Activate(g, me.ID, dead, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead) })
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the regenerated Restless Dead was destroyed")
	}
	if !effectstest.FindCardAnywhere(t, g, dead).Tapped {
		t.Error("regeneration did not tap Restless Dead")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead) })
	if g.Battlefield.Contains(dead) {
		t.Error("one regeneration shield saved Restless Dead twice")
	}
}

// --- the upkeep djinn ------------------------------------------------------------

func TestMirB1DjinnHurtTheirControllerEachUpkeep(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		loss         int
	}{
		{"Benthic Djinn", mirB1BenthicOracle, 2},
		{"Nettletooth Djinn", mirB1NettletoothOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.MirB1Push(g, me.ID, tc.name, "Creature — Djinn", tc.oracle, 4, 4)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			before := effectstest.LifeOf(g, me.ID)
			others := map[uuid.UUID]int{}
			for _, p := range g.Seats[1:] {
				others[p.ID] = effectstest.LifeOf(g, p.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, me.ID); got != before-tc.loss {
				t.Errorf("%s's controller has %d life, want %d", tc.name, got, before-tc.loss)
			}
			for id, life := range others {
				if got := effectstest.LifeOf(g, id); got != life {
					t.Errorf("another player's life changed from %d to %d", life, got)
				}
			}
		})
	}
}

// --- phasing -----------------------------------------------------------------------

func TestMirB1PhasersPhaseOutInTheirControllersUntapStep(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Sandbar Crocodile", mirB1CrocodileOracle},
		{"Teferi's Drake", mirB1TeferisDrakeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.MirB1Push(g, me.ID, tc.name, "Creature", tc.oracle, 3, 2)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			if !g.IsPhasedOutForEffect(id) {
				t.Errorf("%s did not phase out in its controller's untap step", tc.name)
			}
		})
	}
}
