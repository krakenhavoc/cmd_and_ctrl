package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b649_cards_test.go — OLD batch 649 (Invasion, #5616): nine kicker
// cards. Each test plays the real catalog entry through the engine.

const (
	oldB649Aggressor = "84c00644-d0e2-418b-b931-4e972ecb3366"
	oldB649Barricade = "8531b38a-8cc9-452b-8373-3febc733a95b"
	oldB649Verduran  = "915bd598-c3dd-4982-bdcc-5904a6f17436"
	oldB649Growth    = "96c65faf-859c-4436-a551-5538409a9792"
	oldB649Benalish  = "99f2e195-5da5-4239-87e0-ed552643e7b5"
	oldB649Elite     = "9e112393-962b-461e-a605-9f8d6a1ebaa4"
	oldB649Surge     = "a151c45f-0b9e-41af-90b3-fafa12e3bbce"
	oldB649Soldier   = "a56d301b-8f4c-4f12-8ee0-34c6b66faee0"
	oldB649Thicket   = "a85390b6-e90b-4d0c-9573-eb67ce01a0aa"
)

func TestOldB649EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB649Aggressor: "Kavu Aggressor", oldB649Barricade: "Prison Barricade", oldB649Verduran: "Verduran Emissary",
		oldB649Growth: "Explosive Growth", oldB649Benalish: "Benalish Emissary", oldB649Elite: "Llanowar Elite",
		oldB649Surge: "Canopy Surge", oldB649Soldier: "Ardent Soldier", oldB649Thicket: "Thicket Elemental",
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

// oldB649Cast casts a card from the active player's hand, kicked or not.
func oldB649Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, kicked bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s (kicked=%v): %v", name, kicked, err)
	}
	return id
}

func TestOldB649KickedCreaturesEnterWithCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, want int
	}{
		{"Kavu Aggressor", "Creature — Kavu", oldB649Aggressor, 3, 2, 1},
		{"Prison Barricade", "Creature — Wall", oldB649Barricade, 1, 3, 1},
		{"Llanowar Elite", "Creature — Elf", oldB649Elite, 1, 1, 5},
		{"Ardent Soldier", "Creature — Human Soldier", oldB649Soldier, 1, 2, 1},
	} {
		t.Run(tc.name+" kicked", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := oldB649Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, true, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d counters, want %d", got, tc.want)
			}
			if p, _ := effectstest.PtOf(t, g, id); p != tc.power+tc.want {
				t.Errorf("power %d, want %d", p, tc.power+tc.want)
			}
		})
		t.Run(tc.name+" unkicked", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := oldB649Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, false, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 0 {
				t.Errorf("%d counters unkicked, want 0", got)
			}
		})
	}
}

func TestOldB649KavuAggressorCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kavu := effectstest.B12Push(g, opp.ID, "Kavu Aggressor", "Creature — Kavu", oldB649Aggressor, 3, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if err := g.DeclareBlocker(kavu, attacker); err == nil {
		t.Error("Kavu Aggressor blocked")
	}
}

func TestOldB649PrisonBarricadeAttacksOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		opp := g.Seats[(seat+1)%len(g.Seats)]
		id := oldB649Cast(t, g, "Prison Barricade", "Creature — Wall", oldB649Barricade, 1, 3, kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
		effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
		err := g.DeclareAttacker(id, opp.ID)
		if kicked && err != nil {
			t.Errorf("a kicked Prison Barricade can't attack: %v", err)
		}
		if !kicked && err == nil {
			t.Error("an unkicked Prison Barricade attacked through defender")
		}
	}
}

func TestOldB649EmissariesDestroyOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, victimType string
		power, toughness                   int
	}{
		{"Verduran Emissary", "Creature — Human Wizard", oldB649Verduran, "Artifact", 2, 3},
		{"Benalish Emissary", "Creature — Human Wizard", oldB649Benalish, "Basic Land — Plains", 1, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			victim := effectstest.B12Permanent(g, opp.ID, "Victim", tc.victimType)
			oldB649Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, false, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(victim) {
				t.Fatal("an unkicked cast destroyed the permanent")
			}
			oldB649Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, true, nil)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the kicked Emissary did not destroy the permanent")
			}
		})
	}
}

func TestOldB649ExplosiveGrowthPumps(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		want   int
	}{{false, 4}, {true, 7}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		oldB649Cast(t, g, "Explosive Growth", "Instant", oldB649Growth, 0, 0, tc.kicked,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != tc.want || tough != tc.want {
			t.Errorf("kicked=%v: %d/%d, want %d/%d", tc.kicked, p, tough, tc.want, tc.want)
		}
	}
}

func TestOldB649CanopySurgeHitsFlyersAndPlayers(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		dmg    int
	}{{false, 1}, {true, 4}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Big Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
			Power: 1, Toughness: 9, Owner: opp.ID, Controller: opp.ID,
		})
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 9)
		myLife, oppLife := me.Life, opp.Life
		oldB649Cast(t, g, "Canopy Surge", "Sorcery", oldB649Surge, 0, 0, tc.kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageMarkedOn(g, bird); got != tc.dmg {
			t.Errorf("kicked=%v: the flyer took %d, want %d", tc.kicked, got, tc.dmg)
		}
		if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
			t.Errorf("kicked=%v: the ground creature took %d, want 0", tc.kicked, got)
		}
		if me.Life != myLife-tc.dmg || opp.Life != oppLife-tc.dmg {
			t.Errorf("kicked=%v: lives %d/%d, want %d/%d", tc.kicked, me.Life, opp.Life, myLife-tc.dmg, oppLife-tc.dmg)
		}
	}
}

func TestOldB649ThicketElementalRevealsToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	prey := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Prey", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	before := len(me.Library.Cards)
	oldB649Cast(t, g, "Thicket Elemental", "Creature — Elemental", oldB649Thicket, 4, 4, true, nil)
	for i := 0; i < 8; i++ {
		if effectstest.LatestConfirmFor(g, me.ID) != nil {
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			continue
		}
		if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(prey) {
		t.Error("the creature card did not enter")
	}
	if z := effectstest.B12ZoneOf(g, land); z != game.ZoneLibrary {
		t.Errorf("the land is in %s, want the library", z)
	}
	if len(me.Library.Cards) != before-1 {
		t.Errorf("library %d, want %d", len(me.Library.Cards), before-1)
	}
}

func TestOldB649ThicketElementalUnkickedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	prey := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Prey", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oldB649Cast(t, g, "Thicket Elemental", "Creature — Elemental", oldB649Thicket, 4, 4, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(prey) {
		t.Error("an unkicked Thicket Elemental put a creature onto the battlefield")
	}
}
