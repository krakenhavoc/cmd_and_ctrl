package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_relative_targets_test.go — #1863. A target described by what
// the source (or the creature it is attached to) is doing in combat is
// judged at announce and again at resolution, and against the source's
// last-known information once it has left the battlefield.

const (
	wallOfViperOracle    = "4fcfe370-ba33-438f-bd96-2ae560f59df9"
	oldManOfTheSeaOracle = "cce84cf1-5574-43b0-9d75-72e6451403a7"
	whipVineOracle       = "4e90ddce-1765-44e4-9b3c-e576f3a1bcd4"
	goblinSnowmanOracle  = "38e9ad30-4bbf-4b58-8bb2-47520ce351a3"
	plasmaCasterOracle   = "1c50c637-750c-4e31-a69c-915430fc3194"
)

// crtLeaveCombat takes a creature out of combat (CR 506.4).
func crtLeaveCombat(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = uuid.Nil
				g.Battlefield.Cards[i].BlockingTarget = uuid.Nil
				g.Battlefield.Cards[i].AlsoBlocking = nil
			}
		}
	})
}

func crtOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

// --- Wall of Vipers ---------------------------------------------------

func TestWallOfViperDestroysItselfAndTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Vipers", "Creature — Snake Wall", wallOfViperOracle, false)
	attacker := effectstest.CtrlPushCreature(g, me.ID, "Attacker")
	bystander := effectstest.CtrlPushCreature(g, me.ID, "Bystander")
	effectstest.CrtBlock(g, wall, attacker, opp.ID)

	// Any player may activate it: the attacker's controller does.
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wall, 0, effectstest.LtCardTarget(bystander), "C", "C", "C"); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a creature the Wall is not blocking = %v, want ErrIllegalTarget (CR 601.2c)", err)
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wall, 0, effectstest.LtCardTarget(attacker), "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if crtOnBattlefield(g, wall) || crtOnBattlefield(g, attacker) {
		t.Error("the Wall and the creature it blocked should both be destroyed")
	}
	if !crtOnBattlefield(g, bystander) {
		t.Error("the bystander was destroyed")
	}
}

func TestWallOfViperFizzlesWhenTheCreatureLeavesCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Vipers", "Creature — Snake Wall", wallOfViperOracle, false)
	attacker := effectstest.CtrlPushCreature(g, me.ID, "Attacker")
	effectstest.CrtBlock(g, wall, attacker, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wall, 0, effectstest.LtCardTarget(attacker), "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	crtLeaveCombat(g, attacker) // in response
	effectstest.PassPriorityAroundTable(t, g)
	if !crtOnBattlefield(g, wall) || !crtOnBattlefield(g, attacker) {
		t.Error("the target was no longer legal at resolution (CR 608.2b): nothing, the Wall included, should be destroyed")
	}
}

// The Wall dies in response: the ability still resolves, and the target
// is judged by what the Wall was blocking when it left (CR 608.2h).
func TestWallOfViperDestroyedInResponseStillTakesTheCreatureItBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Vipers", "Creature — Snake Wall", wallOfViperOracle, false)
	attacker := effectstest.CtrlPushCreature(g, me.ID, "Attacker")
	effectstest.CrtBlock(g, wall, attacker, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, wall, 0, effectstest.LtCardTarget(attacker), "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wall) })
	effectstest.PassPriorityAroundTable(t, g)
	if crtOnBattlefield(g, attacker) {
		t.Error("the attacker survived: the Wall's last-known block should have kept it a legal target")
	}
}

// --- Whip Vine --------------------------------------------------------

func TestWhipVineOffersOnlyAFlyingCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vine := effectstest.PushCatalogPermanent(g, opp.ID, "Whip Vine", "Creature — Plant Wall", whipVineOracle, false)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"}})
	walker := effectstest.CtrlPushCreature(g, me.ID, "Bear")
	unblockedFlier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Other Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"}})
	effectstest.CrtBlock(g, vine, flier, opp.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == unblockedFlier || g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].AttackingTarget = opp.ID
			}
		}
	})

	for _, bad := range []uuid.UUID{walker, unblockedFlier} {
		if err := effectstest.CrtActivateRaw(t, g, opp.ID, vine, 0, effectstest.LtCardTarget(bad)); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("picking %s = %v, want ErrIllegalTarget", bad, err)
		}
	}
	effectstest.RtActivate(t, g, opp.ID, vine, 0, effectstest.LtCardTarget(flier))
	c := effectstest.LayeredCard(t, g, flier)
	if !c.Tapped || len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].While == nil ||
		c.NextUntapSkips[0].While.Condition != game.WhileSourceRemainsTapped {
		t.Fatalf("flier: tapped %v, holds %+v; want tapped and held while the Vine remains tapped", c.Tapped, c.NextUntapSkips)
	}
}

// --- Goblin Snowman ---------------------------------------------------

func TestGoblinSnowmanDealsOneDamageToTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snowman := effectstest.PushCatalogPermanent(g, opp.ID, "Goblin Snowman", "Creature — Goblin", goblinSnowmanOracle, false)
	attacker := effectstest.CtrlPushCreature(g, me.ID, "Attacker")
	other := effectstest.CtrlPushCreature(g, me.ID, "Other")
	effectstest.CrtBlock(g, snowman, attacker, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, snowman, 0, effectstest.LtCardTarget(other)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a creature it is not blocking = %v, want ErrIllegalTarget", err)
	}
	effectstest.RtActivate(t, g, opp.ID, snowman, 0, effectstest.LtCardTarget(attacker))
	if got := effectstest.LayeredCard(t, g, attacker).DamageMarked; got != 1 {
		t.Errorf("damage on the blocked creature = %d, want 1", got)
	}
	if got := effectstest.LayeredCard(t, g, other).DamageMarked; got != 0 {
		t.Errorf("damage on the unblocked creature = %d, want 0", got)
	}
}

func TestGoblinSnowmanStillDealsTheDamageAfterItHasLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snowman := effectstest.PushCatalogPermanent(g, opp.ID, "Goblin Snowman", "Creature — Goblin", goblinSnowmanOracle, false)
	attacker := effectstest.CtrlPushCreature(g, me.ID, "Attacker")
	effectstest.CrtBlock(g, snowman, attacker, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, snowman, 0, effectstest.LtCardTarget(attacker)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(snowman) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, attacker).DamageMarked; got != 1 {
		t.Errorf("damage = %d, want 1 from the Snowman's last-known self (CR 608.2h)", got)
	}
}

func TestGoblinSnowmanBlockPreventsCombatDamageToAndByIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snowman := effectstest.PushDiesCreatureForTest(g, opp.ID, "Goblin Snowman", goblinSnowmanOracle, "Creature — Goblin", 1, 1)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, snowman)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g) // the block trigger
	if n := effectstest.Pr7aShields(g); n != 1 {
		t.Fatalf("%d prevention records after the block, want one for both directions", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, attacker) != 0 || effectstest.DamageMarkedOn(g, snowman) != 0 {
		t.Errorf("attacker %d, Snowman %d damage: want both prevented", effectstest.DamageMarkedOn(g, attacker), effectstest.DamageMarkedOn(g, snowman))
	}
}

// --- Old Man of the Sea -----------------------------------------------

func TestOldManOfTheSeaStealsAnEqualPowerCreatureAndLosesItWhenItGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldMan := effectstest.PushCatalogPermanent(g, me.ID, "Old Man of the Sea", "Creature — Djinn", oldManOfTheSeaOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == oldMan {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	small := effectstest.CtrlPushCreature(g, opp.ID, "Bear") // 2/2
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Giant", TypeLine: "Creature — Giant",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CrtActivateRaw(t, g, me.ID, oldMan, 0, effectstest.LtCardTarget(big)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a power-3 target of a power-2 Old Man = %v, want ErrIllegalTarget", err)
	}
	g.PlayerByIDForEffect(me.ID).ManaPool = nil
	effectstest.RtActivate(t, g, me.ID, oldMan, 0, effectstest.LtCardTarget(small))
	if got := effectstest.ControllerOf(t, g, small); got != me.ID {
		t.Fatalf("controller = %s, want the Old Man's", got)
	}
	// Its power rises above his: the control ends (ADR 0109 §3).
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(small, game.CounterPlusOne, 1) })
	effectstest.ApaLive(g, small) // the pass that raises the power
	effectstest.ApaLive(g, small) // the pass the broken condition forces (powerConditionsFailLocked)
	if got := effectstest.ControllerOf(t, g, small); got != opp.ID {
		t.Errorf("controller after it grew = %s, want back with its owner", got)
	}
}

func TestOldManOfTheSeaReleasesTheCreatureWhenItUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldMan := effectstest.PushCatalogPermanent(g, me.ID, "Old Man of the Sea", "Creature — Djinn", oldManOfTheSeaOracle, false)
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mouse", TypeLine: "Creature — Mouse",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.RtActivate(t, g, me.ID, oldMan, 0, effectstest.LtCardTarget(small))
	if got := effectstest.ControllerOf(t, g, small); got != me.ID {
		t.Fatalf("controller = %s, want the Old Man's", got)
	}
	effectstest.RtUntap(g, oldMan)
	if got := effectstest.ControllerOf(t, g, small); got != opp.ID {
		t.Error("the Old Man untapped and the creature stayed stolen")
	}
}

// --- Plasma Caster ----------------------------------------------------

func plasmaSetup(t *testing.T) (g *game.Game, me, opp *game.Player, caster, host, blocker, other uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	caster = effectstest.PushCatalogPermanent(g, me.ID, "Plasma Caster", "Artifact — Equipment", plasmaCasterOracle, false)
	host = effectstest.CtrlPushCreature(g, me.ID, "Host")
	blocker = effectstest.CtrlPushCreature(g, opp.ID, "Blocker")
	other = effectstest.CtrlPushCreature(g, opp.ID, "Bystander")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == caster {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
			}
		}
	})
	effectstest.CrtBlock(g, blocker, host, opp.ID)
	if err := g.SetEnergy(me.ID, 2); err != nil {
		t.Fatal(err)
	}
	return
}

func TestPlasmaCasterOffersOnlyBlockersOfTheEquippedCreature(t *testing.T) {
	g, me, _, caster, _, blocker, other := plasmaSetup(t)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, caster, 0, effectstest.LtCardTarget(other)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature not blocking the equipped creature = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, caster, 0, effectstest.LtCardTarget(blocker)); err != nil {
		t.Fatalf("a blocker of the equipped creature: %v", err)
	}
}

func TestPlasmaCasterUnattachedOffersNothing(t *testing.T) {
	g, me, _, caster, _, blocker, _ := plasmaSetup(t)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == caster {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{}
			}
		}
	})
	if err := effectstest.CrtActivateRaw(t, g, me.ID, caster, 0, effectstest.LtCardTarget(blocker)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an unattached Equipment = %v, want ErrIllegalTarget", err)
	}
}

func TestPlasmaCasterFlipExilesOnAWinAndDamagesOnALoss(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g, me, _, caster, _, blocker, _ := plasmaSetup(t)
		if err := effectstest.CrtActivateRaw(t, g, me.ID, caster, 0, effectstest.LtCardTarget(blocker)); err != nil {
			t.Fatal(err)
		}
		if got := g.PlayerByIDForEffect(me.ID).Energy; got != 0 {
			t.Fatalf("energy after paying = %d, want 0", got)
		}
		effectstest.PassPriorityAroundTable(t, g)
		call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if call == nil {
			t.Fatal("no coin call")
		}
		if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
			t.Fatal(err)
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		if won {
			sawWin = true
			if crtOnBattlefield(g, blocker) {
				t.Error("won the flip but the blocker is still on the battlefield")
			}
		} else {
			sawLoss = true
			c, ok := effectstest.BattlefieldCard(g, blocker)
			if !ok || c.DamageMarked != 1 {
				t.Errorf("lost the flip: blocker %+v on battlefield %v, want 1 damage", c.DamageMarked, ok)
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("never saw both outcomes (win %v, loss %v)", sawWin, sawLoss)
	}
}
