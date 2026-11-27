package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b713_cards_test.go — Portal Three Kingdoms, removal, batch 713
// (S82, #5622). Each test plays the real catalog entry through the engine.

const (
	oldB713MisfortuneOracle = "01b778d7-0965-48ab-a81f-e29f8a233d07"
	oldB713TripWireOracle   = "5751df7a-aa9a-4a27-b7ab-445ec4bcbdee"
	oldB713PoisonOracle     = "6d920c99-8a85-4b20-97c2-7aa0c79ae70a"
	oldB713MazeOracle       = "9097457c-ceb3-4944-8f01-c28ddefbb270"
	oldB713MarchOracle      = "a1b87d5d-9493-4001-8c7e-553f40de61ea"
	oldB713HordeOracle      = "b775b83c-0c3d-437d-9d3c-1d5c9cd418a8"
)

func TestOldB713EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB713MisfortuneOracle: "Misfortune's Gain", oldB713TripWireOracle: "Trip Wire",
		oldB713PoisonOracle: "Poison Arrow", oldB713MazeOracle: "Eightfold Maze",
		oldB713MarchOracle: "Guan Yu's 1,000-Li March", oldB713HordeOracle: "Ravaging Horde",
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

func oldB713Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string, targets ...uuid.UUID) error {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: refs})
}

func TestOldB713MisfortunesGainPaysTheOwnerNotTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// opp's creature, stolen by the third seat: the OWNER (opp) gains.
	stolen := effectstest.B12Creature(g, opp.ID, "Stolen", "Creature — Test", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, stolen).Controller = g.Seats[2].ID })
	lives := effectstest.B27Lives(g)
	if err := oldB713Cast(t, g, me, "Misfortune's Gain", "Sorcery", "{3}{W}", oldB713MisfortuneOracle, "CCCW", stolen); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, stolen); ok {
		t.Error("the creature was not destroyed")
	}
	got := effectstest.B27Lives(g)
	if got[1] != lives[1]+4 {
		t.Errorf("owner life %d, want %d", got[1], lives[1]+4)
	}
	if got[0] != lives[0] || got[2] != lives[2] {
		t.Errorf("the caster or the controller gained life: %v -> %v", lives, got)
	}
}

func TestOldB713TripWireOnlyTargetsHorsemanship(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SfPush(g, me.ID, "Plain", "Creature — Test", 2, []string{"G"})
	rider := effectstest.SfPush(g, me.ID, "Rider", "Creature — Test", 2, []string{"G"}, "horsemanship")
	if err := oldB713Cast(t, g, me, "Trip Wire", "Sorcery", "{2}{G}", oldB713TripWireOracle, "CCG", plain); err == nil {
		t.Fatal("Trip Wire targeted a creature without horsemanship")
	}
	if err := oldB713Cast(t, g, me, "Trip Wire", "Sorcery", "{2}{G}", oldB713TripWireOracle, "CCG", rider); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, rider); ok {
		t.Error("the horsemanship creature survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, plain); !ok {
		t.Error("the other creature was destroyed")
	}
}

func TestOldB713PoisonArrowDestroysNonblackAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	black := effectstest.SfPush(g, g.Seats[1].ID, "Black", "Creature — Test", 2, []string{"B"})
	green := effectstest.SfPush(g, g.Seats[1].ID, "Green", "Creature — Test", 2, []string{"G"})
	if err := oldB713Cast(t, g, me, "Poison Arrow", "Sorcery", "{4}{B}{B}", oldB713PoisonOracle, "CCCCBB", black); err == nil {
		t.Fatal("Poison Arrow targeted a black creature")
	}
	before := me.Life
	if err := oldB713Cast(t, g, me, "Poison Arrow", "Sorcery", "{4}{B}{B}", oldB713PoisonOracle, "CCCCBB", green); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, green); ok {
		t.Error("the green creature survived")
	}
	if me.Life != before+3 {
		t.Errorf("life %d, want %d", me.Life, before+3)
	}
}

func TestOldB713EightfoldMazeNeedsBeingAttackedAndAnAttackingTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	if err := oldB713Cast(t, g, opp, "Eightfold Maze", "Instant", "{2}{W}", oldB713MazeOracle, "CCW", atk); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, atk)
	if err := oldB713Cast(t, g, opp, "Eightfold Maze", "Instant", "{2}{W}", oldB713MazeOracle, "CCW", home); err == nil {
		t.Fatal("Eightfold Maze targeted a creature that isn't attacking")
	}
	if err := oldB713Cast(t, g, g.Seats[2], "Eightfold Maze", "Instant", "{2}{W}", oldB713MazeOracle, "CCW", atk); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	if err := oldB713Cast(t, g, opp, "Eightfold Maze", "Instant", "{2}{W}", oldB713MazeOracle, "CCW", atk); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, atk); ok {
		t.Error("the attacker survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, home); !ok {
		t.Error("the creature that wasn't attacking was destroyed")
	}
}

func TestOldB713GuanYusMarchDestroysEveryTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mineTapped := effectstest.SfPush(g, me.ID, "Mine Tapped", "Creature — Test", 2, []string{"W"})
	theirsTapped := effectstest.SfPush(g, opp.ID, "Theirs Tapped", "Creature — Test", 2, []string{"R"})
	theirsUntapped := effectstest.SfPush(g, opp.ID, "Theirs Untapped", "Creature — Test", 2, []string{"R"})
	tappedLand := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.B02bSetTapped(g, mineTapped, true)
	effectstest.B02bSetTapped(g, theirsTapped, true)
	effectstest.B02bSetTapped(g, tappedLand, true)
	if err := oldB713Cast(t, g, me, "Guan Yu's 1,000-Li March", "Sorcery", "{4}{W}{W}", oldB713MarchOracle, "CCCCWW"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mineTapped, theirsTapped} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("a tapped creature survived")
		}
	}
	for _, id := range []uuid.UUID{theirsUntapped, tappedLand} {
		if _, ok := effectstest.BattlefieldCard(g, id); !ok {
			t.Error("an untapped creature or a tapped land was destroyed")
		}
	}
}

func TestOldB713RavagingHordeDestroysTheTargetLandWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	other := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.CastHoldCreature(t, g, "Ravaging Horde", "Creature — Human Soldier", oldB713HordeOracle, 3, 3)
	effectstest.AnswerPickTargetIfOpen(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	_, a := effectstest.BattlefieldCard(g, land)
	_, b := effectstest.BattlefieldCard(g, other)
	if a == b {
		t.Errorf("exactly one land should be destroyed: forest on battlefield %v, island %v", a, b)
	}
}
