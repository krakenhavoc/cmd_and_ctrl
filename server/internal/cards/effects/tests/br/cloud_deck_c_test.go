package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCloudDeckAvalanchePowerCountsOpponentArtifacts(t *testing.T) {
	cloudDeckWantFull(t, effectstest.AvalancheSector7Oracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	av := effectstest.PushCatalogPermanent(g, me.ID, "Avalanche of Sector 7", "Legendary Creature — Human Rebel", effectstest.AvalancheSector7Oracle, false)
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if got := effectstest.EffectivePower(t, g, av); got != 0 {
		t.Errorf("power %d, want 0 (my own artifacts don't count)", got)
	}
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Other Rock", "Artifact Creature — Construct")
	if got := effectstest.EffectivePower(t, g, av); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
}

func TestCloudDeckAvalancheWatchesOpponentActivationsIncludingMana(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.AvalancheSector7Oracle)
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
	cloudDeckWantFull(t, effectstest.TifaMartialOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a, b := g.Seats[(seat+1)%len(g.Seats)], g.Seats[(seat+2)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tifa := effectstest.PushCatalogPermanent(g, me.ID, "Tifa, Martial Artist", "Legendary Creature — Human Monk", effectstest.TifaMartialOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tifa, a.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, b.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, tifa); got != 3 {
		// pushCatalogPermanent makes a 1/1; two opponents attacked: +2.
		t.Errorf("Tifa power %d, want 3 (1 base, +2 for two opponents)", got)
	}
}

func TestCloudDeckTifaUntapsAndAddsACombatWhenAPowerSevenCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tifa, Martial Artist", "Legendary Creature — Human Monk", effectstest.TifaMartialOracle, false)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 7, 7)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.LookupBattlefield(g, giant)
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
	cloudDeckWantFull(t, effectstest.SephirothFallenOracle)
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		effectstest.AdvanceToMain(t, g)
		seph := effectstest.PushCatalogPermanent(g, me.ID, "Sephiroth, Fallen Hero", "Legendary Creature — Human Avatar Soldier", effectstest.SephirothFallenOracle, false)
		bear := effectstest.SeedBear(g, me.ID)

		effectstest.DeclareAttack(t, g, opp.ID, seph)
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			effectstest.PickCard(t, g, me.ID, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatal("no confirm prompt for the cell counter")
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)

		wantPower := 2
		if accept {
			wantPower = 7
		}
		if got := effectstest.EffectivePower(t, g, bear); got != wantPower {
			t.Errorf("accept=%v: Bear power %d, want %d", accept, got, wantPower)
		}
		if got := effectstest.EffectiveToughness(t, g, bear); accept && got != 5 {
			t.Errorf("Bear toughness %d, want 5", got)
		}
	}
}

func TestCloudDeckSephirothReunionReturnsItTappedForASacrificedModifiedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, "cell", 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	id := effectstest.PushCatalogGraveyardCard(me, "Sephiroth, Fallen Hero", "Legendary Creature — Human Avatar Soldier", effectstest.SephirothFallenOracle, 7, 5)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate The Reunion: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.LookupBattlefield(g, id)
	if !ok || !c.Tapped {
		t.Errorf("Sephiroth on battlefield %v, tapped %v; want back and tapped", ok, c.Tapped)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the modified creature was not sacrificed")
	}
}
