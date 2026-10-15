package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_c_test.go — the third slice of the Cloud, Ex-SOLDIER deck
// request (#3656): Avalanche of Sector 7, Tifa and Sephiroth.

const (
	avalancheSector7Oracle = "4eb1414c-ba12-4bae-b410-79d51925d6fa"
	tifaMartialOracle      = "0836c8d1-c453-4a3a-a52a-eb66839b8542"
	sephirothFallenOracle  = "ca9ab230-51e2-4ef2-a7b9-4e325d9f9657"
)

func TestCloudDeckAvalanchePowerCountsOpponentArtifacts(t *testing.T) {
	cloudDeckWantFull(t, avalancheSector7Oracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToMain(t, g)
	av := pushCatalogPermanent(g, me.ID, "Avalanche of Sector 7", "Legendary Creature — Human Rebel", avalancheSector7Oracle, false)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	if got := effectivePower(t, g, av); got != 0 {
		t.Errorf("power %d, want 0 (my own artifacts don't count)", got)
	}
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b12Permanent(g, opp.ID, "Their Other Rock", "Artifact Creature — Construct")
	if got := effectivePower(t, g, av); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
}

func TestCloudDeckAvalancheWatchesOpponentActivationsIncludingMana(t *testing.T) {
	spec, ok := Lookup(avalancheSector7Oracle)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatalf("Avalanche registered %v with %d triggers", ok, len(spec.Triggered))
	}
	kinds := map[game.EventKind]bool{}
	for _, k := range spec.Triggered[0].Watches {
		kinds[k] = true
	}
	if !kinds[game.EventActivateAbility] || !kinds[game.EventManaAbilityActivated] {
		t.Errorf("watches %v, want both activation kinds", spec.Triggered[0].Watches)
	}
}

func TestCloudDeckTifaMeleeGrowsPerOpponentAttacked(t *testing.T) {
	cloudDeckWantFull(t, tifaMartialOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a, b := g.Seats[(seat+1)%len(g.Seats)], g.Seats[(seat+2)%len(g.Seats)]
	advanceToMain(t, g)
	tifa := pushCatalogPermanent(g, me.ID, "Tifa, Martial Artist", "Legendary Creature — Human Monk", tifaMartialOracle, false)
	bear := seedBear(g, me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tifa, a.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, b.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, tifa); got != 3 {
		// pushCatalogPermanent makes a 1/1; two opponents attacked: +2.
		t.Errorf("Tifa power %d, want 3 (1 base, +2 for two opponents)", got)
	}
}

func TestCloudDeckTifaUntapsAndAddsACombatWhenAPowerSevenCreatureConnects(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Tifa, Martial Artist", "Legendary Creature — Human Monk", tifaMartialOracle, false)
	giant := b12Creature(g, me.ID, "Giant", "Creature — Giant", 7, 7)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	c, ok := lookupBattlefield(g, giant)
	if !ok || c.Tapped {
		t.Errorf("the Giant is still tapped after Tifa's trigger: %+v", c.Tapped)
	}
	for i := 0; i < 40 && !(g.Turn.Step == game.StepBeginCombat && g.Turn.PhaseOrdinal >= 2); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat != seat {
			break
		}
	}
	if g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepBeginCombat || g.Turn.PhaseOrdinal < 2 {
		t.Errorf("no additional combat phase: seat %d step %s ordinal %d", g.Turn.ActiveSeat, g.Turn.Step, g.Turn.PhaseOrdinal)
	}
}

func TestCloudDeckSephirothCellCounterMakesModifiedCreaturesSevenFive(t *testing.T) {
	cloudDeckWantFull(t, sephirothFallenOracle)
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		advanceToMain(t, g)
		seph := pushCatalogPermanent(g, me.ID, "Sephiroth, Fallen Hero", "Legendary Creature — Human Avatar Soldier", sephirothFallenOracle, false)
		bear := seedBear(g, me.ID)

		declareAttack(t, g, opp.ID, seph)
		if pick := latestPickTarget(g, me.ID); pick != nil {
			pickCard(t, g, me.ID, bear)
		}
		passPriorityAroundTable(t, g)
		ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatal("no confirm prompt for the cell counter")
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		passPriorityAroundTable(t, g)

		wantPower := 2
		if accept {
			wantPower = 7
		}
		if got := effectivePower(t, g, bear); got != wantPower {
			t.Errorf("accept=%v: Bear power %d, want %d", accept, got, wantPower)
		}
		if got := effectiveToughness(t, g, bear); accept && got != 5 {
			t.Errorf("Bear toughness %d, want 5", got)
		}
	}
}

func TestCloudDeckSephirothReunionReturnsItTappedForASacrificedModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, "cell", 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	id := pushCatalogGraveyardCard(me, "Sephiroth, Fallen Hero", "Legendary Creature — Human Avatar Soldier", sephirothFallenOracle, 7, 5)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate The Reunion: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := lookupBattlefield(g, id)
	if !ok || !c.Tapped {
		t.Errorf("Sephiroth on battlefield %v, tapped %v; want back and tapped", ok, c.Tapped)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the modified creature was not sacrificed")
	}
}
