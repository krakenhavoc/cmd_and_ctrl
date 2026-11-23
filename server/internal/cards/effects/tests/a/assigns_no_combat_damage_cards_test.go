package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ophidian: yes draws and the 1 combat damage is not dealt; no does neither.
func TestOphidianDrawsInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.NcdAttacker(g, "Ophidian", effectstest.NcdOphidian, 1, 3)
		hand, life := me.Hand.Size(), g.Seats[1].Life
		effectstest.NcdPlay(t, g, id, yes)
		drew := me.Hand.Size() - hand
		lost := effectstest.NcdLifeLost(g, life)
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
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		id := effectstest.NcdAttacker(g, "Farrel's Zealot", effectstest.NcdFarrelsZealot, 2, 2)
		victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
		life := opp.Life
		effectstest.NcdPlay(t, g, id, yes, victim)
		got, lost := effectstest.Pr6Marked(g, victim), effectstest.NcdLifeLost(g, life)
		if yes && (got != 3 || lost != 0 || !effectstest.NcdMarked(g, id)) {
			t.Errorf("yes: victim damage %d (want 3), life lost %d (want 0), marked %v", got, lost, effectstest.NcdMarked(g, id))
		}
		if !yes && (got != 0 || lost != 2 || effectstest.NcdMarked(g, id)) {
			t.Errorf("no: victim damage %d (want 0), life lost %d (want 2), marked %v", got, lost, effectstest.NcdMarked(g, id))
		}
	}
}

// Dwarven Vigilantes: damage equal to its power, read as it resolves.
func TestDwarvenVigilantesDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := effectstest.NcdAttacker(g, "Dwarven Vigilantes", effectstest.NcdVigilantes, 2, 2)
	victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
	life := opp.Life
	effectstest.NcdPlay(t, g, id, true, victim)
	if got := effectstest.Pr6Marked(g, victim); got != 2 {
		t.Errorf("victim damage = %d, want 2 (its power)", got)
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Keeper of Tresserhorn: no choice. 2 life lost and no combat damage from a 6/6.
func TestKeeperOfTresserhornDrainsInsteadOfDealingDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.NcdAttacker(g, "Keeper of Tresserhorn", effectstest.NcdKeeper, 6, 6)
	life := g.Seats[1].Life
	effectstest.NcdPlay(t, g, id, true)
	if lost := effectstest.NcdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2 (the drain, and no combat damage)", lost)
	}
}

// Floral Spuzzem: destroys a defender's artifact; will not hit an
// artifact the defender does not control.
func TestFloralSpuzzemDestroysAnArtifactInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		id := effectstest.NcdAttacker(g, "Floral Spuzzem", effectstest.NcdFloralSpuzzem, 2, 2)
		rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
		life := opp.Life
		effectstest.NcdPlay(t, g, id, yes, rock)
		gone := effectstest.FindBattlefieldCardForTest(g, rock) == nil
		lost := effectstest.NcdLifeLost(g, life)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.NcdAttacker(g, "Kukemssa Pirates", effectstest.NcdKukemssaPirates, 2, 2)
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
	life := opp.Life
	effectstest.NcdPlay(t, g, id, true, rock)
	if c := effectstest.FindBattlefieldCardForTest(g, rock); c == nil || c.Controller != me.ID {
		t.Errorf("the artifact is not under the Pirates' controller")
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
	// Still ours after the turn is cleaned up (no duration).
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if c := effectstest.FindBattlefieldCardForTest(g, rock); c == nil || c.Controller != me.ID {
		t.Errorf("the theft ended at cleanup, want indefinite")
	}
}

// Orcish Squatters: the land is ours only while we control the Squatters.
func TestOrcishSquattersTakesALandWhileItStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.NcdAttacker(g, "Orcish Squatters", effectstest.NcdSquatters, 2, 3)
	land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	life := opp.Life
	effectstest.NcdPlay(t, g, id, true, land)
	if c := effectstest.FindBattlefieldCardForTest(g, land); c == nil || c.Controller != me.ID {
		t.Fatalf("the land is not under the Squatters' controller")
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
	if c := effectstest.FindBattlefieldCardForTest(g, land); c == nil || c.Controller != opp.ID {
		t.Errorf("the land did not go home when the Squatters left")
	}
}

// Goblin Vandal: pay {R}, destroy the artifact, no combat damage. Declining
// the payment does none of it.
func TestGoblinVandalPaysToDestroyAnArtifact(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.NcdAttacker(g, "Goblin Vandal", effectstest.NcdGoblinVandal, 1, 1)
		rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Toughness: 1})
		effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		life := opp.Life
		effectstest.NcdPlay(t, g, id, pay, rock)
		gone := effectstest.FindBattlefieldCardForTest(g, rock) == nil
		lost := effectstest.NcdLifeLost(g, life)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.NcdAttacker(g, "Rysorian Badger", effectstest.NcdBadger, 2, 2)
	a := effectstest.PushGraveyardCardForTest(opp, "Dead A")
	b := effectstest.PushGraveyardCardForTest(opp, "Dead B")
	effectstest.PushGraveyardCardForTest(opp, "Dead C")
	mine, life := me.Life, opp.Life
	effectstest.NcdPlay(t, g, id, true, a, b)
	if n := opp.Graveyard.Size(); n != 1 {
		t.Errorf("defender graveyard = %d cards, want 1", n)
	}
	if got := me.Life - mine; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Bone Dancer: the top creature card of the defender's graveyard comes
// back under our control.
func TestBoneDancerReanimatesTheTopCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.NcdAttacker(g, "Bone Dancer", effectstest.NcdBoneDancer, 2, 2)
	effectstest.PushGraveyardCardForTest(opp, "Old Dead")
	top := effectstest.PushGraveyardCardForTest(opp, "Top Dead")
	life := opp.Life
	effectstest.NcdPlay(t, g, id, true)
	if c := effectstest.FindBattlefieldCardForTest(g, top); c == nil || c.Controller != me.ID {
		t.Errorf("the top creature card did not come back under our control")
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Bone Dancer with nothing to take is not asked, and hits for its power.
func TestBoneDancerWithAnEmptyGraveyardStillHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.NcdAttacker(g, "Bone Dancer", effectstest.NcdBoneDancer, 2, 2)
	life := g.Seats[1].Life
	effectstest.NcdPlay(t, g, id, true)
	if lost := effectstest.NcdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2", lost)
	}
}

// The Laccolith creatures: blocked, a yes shoots the target for the creature's
// power and the blocker takes nothing from it; a no lets combat go on.
func TestLaccolithCreaturesShootInsteadOfFighting(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power        int
	}{
		{"Laccolith Whelp", effectstest.NcdWhelp, 1},
		{"Laccolith Grunt", effectstest.NcdGrunt, 2},
		{"Laccolith Warrior", effectstest.NcdWarrior, 3},
		{"Laccolith Titan", effectstest.NcdTitan, 6},
	} {
		for _, yes := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			id := effectstest.NcdAttacker(g, tc.name, tc.oracle, tc.power, tc.power)
			blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 20})
			victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
			effectstest.NcdPlayBlocked(t, g, id, blocker, yes, victim)
			shot, onWall := effectstest.Pr6Marked(g, victim), effectstest.Pr6Marked(g, blocker)
			if yes && (shot != tc.power || onWall != 0 || !effectstest.NcdMarked(g, id)) {
				t.Errorf("%s yes: victim %d (want %d), wall %d (want 0), marked %v", tc.name, shot, tc.power, onWall, effectstest.NcdMarked(g, id))
			}
			if !yes && (shot != 0 || onWall != tc.power || effectstest.NcdMarked(g, id)) {
				t.Errorf("%s no: victim %d (want 0), wall %d (want %d), marked %v", tc.name, shot, onWall, tc.power, effectstest.NcdMarked(g, id))
			}
		}
	}
}

// Lim-Dûl's Paladin, unblocked: no combat damage and the defender loses 4.
func TestLimDulsPaladinUnblockedDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.NcdAttacker(g, "Lim-Dûl's Paladin", effectstest.NcdPaladin, 0, 3)
	life := g.Seats[1].Life
	effectstest.NcdPlay(t, g, id, true)
	if lost := effectstest.NcdLifeLost(g, life); lost != 4 {
		t.Errorf("defender lost %d, want 4", lost)
	}
}

// Lim-Dûl's Paladin, blocked: +6/+3 until end of turn.
func TestLimDulsPaladinBlockedGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := effectstest.NcdAttacker(g, "Lim-Dûl's Paladin", effectstest.NcdPaladin, 0, 3)
	blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 20})
	effectstest.NcdPlayBlocked(t, g, id, blocker, true)
	if got := effectstest.Pr6Marked(g, blocker); got != 6 {
		t.Errorf("blocker took %d, want 6", got)
	}
}

// Cloak of Confusion: the enchanted creature assigns nothing and the
// defender discards a card at random.
func TestCloakOfConfusionTradesDamageForADiscard(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
		cloak := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Cloak of Confusion", TypeLine: "Enchantment — Aura", OracleID: effectstest.NcdCloak})
		effectstest.NcdAttach(t, g, cloak, bear)
		effectstest.PushHandCard(g, opp)
		hand, life := opp.Hand.Size(), opp.Life
		effectstest.NcdPlay(t, g, bear, yes)
		discarded, lost := hand-opp.Hand.Size(), effectstest.NcdLifeLost(g, life)
		if yes && (discarded != 1 || lost != 0) {
			t.Errorf("yes: discarded %d (want 1), life lost %d (want 0)", discarded, lost)
		}
		if !yes && (discarded != 0 || lost != 2) {
			t.Errorf("no: discarded %d (want 0), life lost %d (want 2)", discarded, lost)
		}
	}
}

// Farrel's Mantle: power plus 2 to another creature, and nothing to the player.
func TestFarrelsMantleShootsForPowerPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	mantle := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Farrel's Mantle", TypeLine: "Enchantment — Aura", OracleID: effectstest.NcdMantle})
	effectstest.NcdAttach(t, g, mantle, bear)
	victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
	life := opp.Life
	effectstest.NcdPlay(t, g, bear, true, victim)
	if got := effectstest.Pr6Marked(g, victim); got != 4 {
		t.Errorf("victim took %d, want 4 (power 2 plus 2)", got)
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 {
		t.Errorf("defender lost %d, want 0", lost)
	}
}

// Laccolith Rig: the enchanted creature, once blocked, shoots instead of fighting.
func TestLaccolithRigShootsForTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 3, "G")
	rig := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Laccolith Rig", TypeLine: "Enchantment — Aura", OracleID: effectstest.NcdRig})
	effectstest.NcdAttach(t, g, rig, bear)
	blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 20})
	victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
	effectstest.NcdPlayBlocked(t, g, bear, blocker, true, victim)
	if shot, onWall := effectstest.Pr6Marked(g, victim), effectstest.Pr6Marked(g, blocker); shot != 3 || onWall != 0 {
		t.Errorf("victim %d (want 3), wall %d (want 0)", shot, onWall)
	}
}

// Lim-Dûl's Paladin's upkeep: discarding keeps it, declining sacrifices it
// and draws a card.
func TestLimDulsPaladinUpkeep(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.NcdAttacker(g, "Lim-Dûl's Paladin", effectstest.NcdPaladin, 0, 3)
		card := effectstest.PushHandCard(g, me)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		hand := me.Hand.Size()
		c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceConfirm)
		if err := g.ResolveConfirm(c.ID, me.ID, discard); err != nil {
			t.Fatal(err)
		}
		if discard {
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
			if pick == nil {
				t.Fatal("no discard prompt")
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{card}); err != nil {
				t.Fatalf("discard pick: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		alive := effectstest.FindBattlefieldCardForTest(g, id) != nil
		switch {
		case discard && (!alive || me.Hand.Size() != hand-1):
			t.Errorf("discard: alive %v, hand %d (want alive, %d)", alive, me.Hand.Size(), hand-1)
		case !discard && (alive || me.Hand.Size() != hand+1):
			t.Errorf("decline: alive %v, hand %d (want gone, %d)", alive, me.Hand.Size(), hand+1)
		}
	}
}
