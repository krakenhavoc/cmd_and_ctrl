package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b384_cards_test.go — OLD batch 384 (Commander 2013, #5560): six
// creatures. Each test plays the real catalog entry through the engine.

const (
	oldB384Tidal   = "1b25e262-e2df-4768-b55e-1b7b8d3ee993"
	oldB384Serene  = "2ce0d583-81ca-4dca-bde0-52f86b683afd"
	oldB384Saurian = "2fb7f844-edaf-43ef-9121-318baf9ec9ce"
	oldB384Horror  = "8267561e-bc25-4aaa-8242-f6d7ec88143e"
	oldB384Efreet  = "9abd2286-23e9-49cd-be53-39423890f35c"
	oldB384Ravager = "c7686204-0433-48cf-bbfb-5d32b6a25cc3"
)

func TestOldB384EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB384Tidal: "Tidal Force", oldB384Serene: "Serene Master", oldB384Saurian: "Brooding Saurian",
		oldB384Horror: "Hooded Horror", oldB384Efreet: "Capricious Efreet", oldB384Ravager: "Terra Ravager",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Hooded Horror ----------------------------------------------------------

func TestOldB384HoodedHorrorIsUnblockableOnlyAgainstTheMostCreatures(t *testing.T) {
	for _, tc := range []struct {
		name          string
		opp, other    int
		wantBlockable bool
	}{
		{"defender has the most", 1, 0, false},
		{"defender is tied for the most", 1, 1, false},
		{"another player has more", 1, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
			horror := effectstest.B12Push(g, me.ID, "Hooded Horror", "Creature — Horror", oldB384Horror, 4, 4)
			wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
			for i := 1; i < tc.opp; i++ {
				effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
			}
			for i := 0; i < tc.other; i++ {
				effectstest.B12Creature(g, other.ID, "Other Wall", "Creature — Wall", 0, 4)
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(horror, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			err := g.DeclareBlocker(wall, horror)
			if tc.wantBlockable && err != nil {
				t.Errorf("the Horror could not be blocked: %v", err)
			}
			if !tc.wantBlockable && err == nil {
				t.Error("the Horror was blocked")
			}
		})
	}
}

// --- Terra Ravager ----------------------------------------------------------

func TestOldB384TerraRavagerGetsPowerForEachDefendingLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	ravager := effectstest.B12Push(g, me.ID, "Terra Ravager", "Creature — Elemental Beast", oldB384Ravager, 0, 4)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, other.ID, "Other Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	effectstest.DeclareAttack(t, g, opp.ID, ravager)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, ravager); p != 3 {
		t.Errorf("power %d, want 3 (the defending player's three lands)", p)
	}
}

// --- Brooding Saurian -------------------------------------------------------

func TestOldB384BroodingSaurianReturnsOwnedPermanentsButNotTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Brooding Saurian", "Creature — Lizard", oldB384Saurian, false)
	stolen := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		g.GainControlForEffect(stolen, stolen, opp.ID, game.IndefiniteDuration(), "test steal")
		g.GainControlForEffect(token, token, opp.ID, game.IndefiniteDuration(), "test steal")
	})
	if got := effectstest.ControllerOf(t, g, stolen); got != opp.ID {
		t.Fatalf("setup: the Bear is controlled by %s, want the opponent", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, stolen); got != me.ID {
		t.Errorf("the Bear is controlled by %s, want its owner", got)
	}
	if got := effectstest.ControllerOf(t, g, token); got != opp.ID {
		t.Errorf("the token went back to %s; tokens are left alone", got)
	}
}

// --- Tidal Force ------------------------------------------------------------

func TestOldB384TidalForceTapsAPermanentAtEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tidal Force", "Creature — Elemental", oldB384Tidal, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// An opponent's upkeep: Tidal Force's controller still chooses.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.MonSettle(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("Tidal Force's controller was not asked for a target at an opponent's upkeep")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.MonSettle(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no tap-or-untap prompt")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || !c.Tapped {
		t.Error("the targeted creature was not tapped")
	}
}

// --- Serene Master ----------------------------------------------------------

func TestOldB384SereneMasterExchangesPowerWithTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Big Attacker", "Creature — Ogre", 5, 5)
	master := effectstest.B12Push(g, opp.ID, "Serene Master", "Creature — Human Monk", oldB384Serene, 0, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(master, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	if effectstest.LatestPickTarget(g, opp.ID) != nil {
		effectstest.PickCard(t, g, opp.ID, attacker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, master); p != 5 {
		t.Errorf("the Master's power is %d, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, attacker); p != 0 {
		t.Errorf("the attacker's power is %d, want 0", p)
	}
}

// --- Capricious Efreet ------------------------------------------------------

func TestOldB384CapriciousEfreetDestroysExactlyOneTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Capricious Efreet", "Creature — Efreet", oldB384Efreet, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Ogre A", "Creature — Ogre", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Ogre B", "Creature — Ogre", 3, 3)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.WalkToUpkeepOfSeat0(t, g)
	for i := 0; i < 4; i++ {
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			break
		}
		want := []uuid.UUID{mine}
		if effectstest.HasID(p.PickTargetCards, a) {
			want = []uuid.UUID{a, b}
		}
		var refs []game.TargetRef
		for _, id := range want {
			refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasID([]uuid.UUID{land}, land) && !g.Battlefield.Contains(land) {
		t.Error("the land, which was not a legal target, was destroyed")
	}
	gone := 0
	for _, id := range []uuid.UUID{mine, a, b} {
		if !g.Battlefield.Contains(id) {
			gone++
		}
	}
	if gone != 1 {
		t.Errorf("%d of the three targets were destroyed, want exactly 1", gone)
	}
}
