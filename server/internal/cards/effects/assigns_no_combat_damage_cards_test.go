package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assigns_no_combat_damage_cards_test.go — the "attacks and isn't blocked,
// you may <do>. If you do, this creature assigns no combat damage this
// turn" cards of #3531 (CR 510.1a, the gate of ADR 0187 §4).

const (
	ncdOphidian        = "953bdefa-4639-4524-af5c-5055d1289f40"
	ncdFarrelsZealot   = "0d063c0e-66bc-4a9c-98aa-a381b9cea120"
	ncdVigilantes      = "b49baf80-237c-41ee-8a22-2b6a68a313e7"
	ncdKeeper          = "ce976cb2-2b69-4d00-918c-d4900244ac86"
	ncdFloralSpuzzem   = "994de451-14f9-466f-a56e-da052b4666e5"
	ncdKukemssaPirates = "280148c5-3a2f-48f7-a147-4578f2068073"
	ncdSquatters       = "ce936736-915f-44bb-a709-b3bdcf90a3be"
	ncdGoblinVandal    = "50bebbb9-01b7-4eb6-8efd-307c9bcd2517"
	ncdBadger          = "cf9fc317-3184-4999-84ef-945a8842096e"
	ncdBoneDancer      = "13b5c9ef-d0e3-453d-ae5e-8de4f310cc38"
)

// ncdPlay attacks seat 1 with `attacker` (nobody blocks), answers every
// prompt the trigger raises (`yes` to the "you may", `targets` to a target
// pick) until the table is quiet, and walks into the combat damage step.
func ncdPlay(t *testing.T, g *game.Game, attacker uuid.UUID, yes bool, targets ...uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	pr7bAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
loop:
	for i := 0; i < 24; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		switch {
		case open == nil && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0:
			break loop
		case open == nil:
			if err := g.PassPriority(); err != nil {
				break loop
			}
		case open.Kind == game.PendingChoicePickTarget:
			var refs []game.TargetRef
			for _, id := range targets {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			if err := g.ResolvePickTargets(open.ID, me.ID, refs); err != nil {
				t.Fatalf("pick targets: %v", err)
			}
		case open.Kind == game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(open.ID, me.ID, yes); err != nil {
				t.Fatalf("confirm: %v", err)
			}
		case open.Kind == game.PendingChoicePayUnless:
			if err := g.ResolvePayUnless(open.ID, me.ID, yes); err != nil {
				t.Fatalf("pay: %v", err)
			}
		default:
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
	}
	pr7bDamageStep(t, g)
}

// ncdLifeLost is how much life the defending seat lost to combat damage.
func ncdLifeLost(g *game.Game, before int) int { return before - g.Seats[1].Life }

func ncdAttacker(g *game.Game, name, oracle string, p, tough int) uuid.UUID {
	return apaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{Name: name, TypeLine: "Creature — Test", OracleID: oracle, Power: p, Toughness: tough})
}

func ncdMarked(g *game.Game, id uuid.UUID) bool {
	return g.PermanentAssignsNoCombatDamageForEffect(id)
}

// Ophidian: yes draws and the 1 combat damage is not dealt; no does neither.
func TestOphidianDrawsInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := ncdAttacker(g, "Ophidian", ncdOphidian, 1, 3)
		hand, life := me.Hand.Size(), g.Seats[1].Life
		ncdPlay(t, g, id, yes)
		drew := me.Hand.Size() - hand
		lost := ncdLifeLost(g, life)
		if yes && (drew != 1 || lost != 0) {
			t.Errorf("yes: drew %d (want 1), defender lost %d (want 0)", drew, lost)
		}
		if !yes && (drew != 0 || lost != 1) {
			t.Errorf("no: drew %d (want 0), defender lost %d (want 1)", drew, lost)
		}
	}
}

// Farrel's Zealot: 3 damage to the target, then none to the player.
func TestFarrelsZealotShootsInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		id := ncdAttacker(g, "Farrel's Zealot", ncdFarrelsZealot, 2, 2)
		victim := pr7Creature(g, opp.ID, "Victim", 1, "G")
		life := opp.Life
		ncdPlay(t, g, id, yes, victim)
		got, lost := pr6Marked(g, victim), ncdLifeLost(g, life)
		if yes && (got != 3 || lost != 0 || !ncdMarked(g, id)) {
			t.Errorf("yes: victim damage %d (want 3), life lost %d (want 0), marked %v", got, lost, ncdMarked(g, id))
		}
		if !yes && (got != 0 || lost != 2 || ncdMarked(g, id)) {
			t.Errorf("no: victim damage %d (want 0), life lost %d (want 2), marked %v", got, lost, ncdMarked(g, id))
		}
	}
}

// Dwarven Vigilantes: damage equal to its power, read as it resolves.
func TestDwarvenVigilantesDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := ncdAttacker(g, "Dwarven Vigilantes", ncdVigilantes, 2, 2)
	victim := pr7Creature(g, opp.ID, "Victim", 1, "G")
	life := opp.Life
	ncdPlay(t, g, id, true, victim)
	if got := pr6Marked(g, victim); got != 2 {
		t.Errorf("victim damage = %d, want 2 (its power)", got)
	}
	if lost := ncdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Keeper of Tresserhorn: no choice. 2 life lost and no combat damage from a 6/6.
func TestKeeperOfTresserhornDrainsInsteadOfDealingDamage(t *testing.T) {
	g := newCatalogGame(t)
	id := ncdAttacker(g, "Keeper of Tresserhorn", ncdKeeper, 6, 6)
	life := g.Seats[1].Life
	ncdPlay(t, g, id, true)
	if lost := ncdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2 (the drain, and no combat damage)", lost)
	}
}

// Floral Spuzzem: destroys a defender's artifact; will not hit an
// artifact the defender does not control.
func TestFloralSpuzzemDestroysAnArtifactInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		id := ncdAttacker(g, "Floral Spuzzem", ncdFloralSpuzzem, 2, 2)
		rock := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
		life := opp.Life
		ncdPlay(t, g, id, yes, rock)
		gone := findBattlefieldCardForTest(g, rock) == nil
		lost := ncdLifeLost(g, life)
		if yes && (!gone || lost != 0) {
			t.Errorf("yes: artifact gone %v (want true), life lost %d (want 0)", gone, lost)
		}
		if !yes && (gone || lost != 2) {
			t.Errorf("no: artifact gone %v (want false), life lost %d (want 2)", gone, lost)
		}
	}
}

// Kukemssa Pirates: takes the artifact for good.
func TestKukemssaPiratesStealsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := ncdAttacker(g, "Kukemssa Pirates", ncdKukemssaPirates, 2, 2)
	rock := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
	life := opp.Life
	ncdPlay(t, g, id, true, rock)
	if c := findBattlefieldCardForTest(g, rock); c == nil || c.Controller != me.ID {
		t.Errorf("the artifact is not under the Pirates' controller")
	}
	if lost := ncdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
	// Still ours after the turn is cleaned up (no duration).
	advanceTo(t, g, game.StepCleanup)
	if c := findBattlefieldCardForTest(g, rock); c == nil || c.Controller != me.ID {
		t.Errorf("the theft ended at cleanup, want indefinite")
	}
}

// Orcish Squatters: the land is ours only while we control the Squatters.
func TestOrcishSquattersTakesALandWhileItStays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := ncdAttacker(g, "Orcish Squatters", ncdSquatters, 2, 3)
	land := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	life := opp.Life
	ncdPlay(t, g, id, true, land)
	if c := findBattlefieldCardForTest(g, land); c == nil || c.Controller != me.ID {
		t.Fatalf("the land is not under the Squatters' controller")
	}
	if lost := ncdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
	if c := findBattlefieldCardForTest(g, land); c == nil || c.Controller != opp.ID {
		t.Errorf("the land did not go home when the Squatters left")
	}
}

// Goblin Vandal: pay {R}, destroy the artifact, no combat damage. Declining
// the payment does none of it.
func TestGoblinVandalPaysToDestroyAnArtifact(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := ncdAttacker(g, "Goblin Vandal", ncdGoblinVandal, 1, 1)
		rock := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
		apaPush(g, me.ID, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		life := opp.Life
		ncdPlay(t, g, id, pay, rock)
		gone := findBattlefieldCardForTest(g, rock) == nil
		lost := ncdLifeLost(g, life)
		if pay && (!gone || lost != 0) {
			t.Errorf("pay: artifact gone %v, life lost %d (want true, 0)", gone, lost)
		}
		if !pay && (gone || lost != 1) {
			t.Errorf("decline: artifact gone %v, life lost %d (want false, 1)", gone, lost)
		}
	}
}

// Rysorian Badger: exile the chosen creature cards, gain 1 each.
func TestRysorianBadgerExilesCardsAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := ncdAttacker(g, "Rysorian Badger", ncdBadger, 2, 2)
	a := pushGraveyardCardForTest(opp, "Dead A")
	b := pushGraveyardCardForTest(opp, "Dead B")
	pushGraveyardCardForTest(opp, "Dead C")
	mine, life := me.Life, opp.Life
	ncdPlay(t, g, id, true, a, b)
	if n := opp.Graveyard.Size(); n != 1 {
		t.Errorf("defender graveyard = %d cards, want 1", n)
	}
	if got := me.Life - mine; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if lost := ncdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Bone Dancer: the top creature card of the defender's graveyard comes
// back under our control.
func TestBoneDancerReanimatesTheTopCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := ncdAttacker(g, "Bone Dancer", ncdBoneDancer, 2, 2)
	pushGraveyardCardForTest(opp, "Old Dead")
	top := pushGraveyardCardForTest(opp, "Top Dead")
	life := opp.Life
	ncdPlay(t, g, id, true)
	if c := findBattlefieldCardForTest(g, top); c == nil || c.Controller != me.ID {
		t.Errorf("the top creature card did not come back under our control")
	}
	if lost := ncdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Bone Dancer with nothing to take is not asked, and hits for its power.
func TestBoneDancerWithAnEmptyGraveyardStillHits(t *testing.T) {
	g := newCatalogGame(t)
	id := ncdAttacker(g, "Bone Dancer", ncdBoneDancer, 2, 2)
	life := g.Seats[1].Life
	ncdPlay(t, g, id, true)
	if lost := ncdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2", lost)
	}
}
