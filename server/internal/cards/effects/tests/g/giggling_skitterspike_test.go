package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gigglingSkitterspikeOracle = "ca02cb31-9850-48bf-bc8d-ed86a4d4d4f7"

func TestGigglingSkitterspikeIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(gigglingSkitterspikeOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Giggling Skitterspike: %+v %v", spec, ok)
	}
	if len(spec.Triggered) != 1 || len(spec.Activated) != 1 {
		t.Fatalf("triggers %d, activated %d, want 1 and 1", len(spec.Triggered), len(spec.Activated))
	}
}

func pushSkitterspike(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Giggling Skitterspike", "Artifact Creature — Toy", gigglingSkitterspikeOracle, false)
}

func TestGigglingSkitterspikeAttackingHitsEachOpponentForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spike := pushSkitterspike(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == spike {
				g.Battlefield.Cards[i].Power = 3
			}
		}
	})
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(spike, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for i, s := range g.Seats[1:] {
		// The trigger's 3 lands on every opponent; the defender may also
		// have taken combat damage by now, so it is a ceiling for them.
		if i == 0 && s.Life > before[s.ID]-3 {
			t.Errorf("defender life %d, want at most %d", s.Life, before[s.ID]-3)
		}
		if i > 0 && s.Life != before[s.ID]-3 {
			t.Errorf("seat %d life %d, want %d", i+1, s.Life, before[s.ID]-3)
		}
	}
	if me.Life != before[me.ID] {
		t.Errorf("controller life %d → %d, want unchanged", before[me.ID], me.Life)
	}
}

func TestGigglingSkitterspikeBlockingHitsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me := g.Seats[0], g.Seats[1]
	spike := pushSkitterspike(g, me.ID)
	attacker := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Attacker", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	before := attackerSeat.Life
	if err := g.DeclareBlocker(spike, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if attackerSeat.Life != before-1 {
		t.Errorf("attacker's controller life %d → %d, want -1", before, attackerSeat.Life)
	}
}

func TestGigglingSkitterspikeFiresWhenASpellTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spike := pushSkitterspike(g, me.ID)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	bolt := game.NewCard("Test Removal", them.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	them.Hand.PushTop(bolt)
	before := them.Life
	if err := g.CastSpell(them.ID, bolt.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spike}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-1 {
		t.Errorf("targeting player's life %d → %d, want -1", before, them.Life)
	}
}

func TestGigglingSkitterspikeIgnoresATargetingAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := pushSkitterspike(g, me.ID)
	src, _ := g.LookupCardForEffect(id)
	spec, _ := effects.Lookup(gigglingSkitterspikeOracle)
	ev := game.Event{Kind: game.EventBecomesTarget, CardID: id, Source: uuid.New()}
	if spec.Triggered[0].AppliesTo(ev, &src, src.Effective(), g) {
		t.Error("an ability that targets it triggers it")
	}
}

func TestGigglingSkitterspikeMonstrosityPutsFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spike := pushSkitterspike(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCC")
	effectstest.ActivateMonstrosity(t, g, me.ID, spike, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, spike); n != 5 {
		t.Errorf("+1/+1 counters = %d, want 5", n)
	}
	if !effectstest.MonstrousNow(g, spike) {
		t.Error("not monstrous")
	}
}
