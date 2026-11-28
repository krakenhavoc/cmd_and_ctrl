package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o614Deathgazer = "0f772f35-6d98-4ac8-88e5-02539737078b"
	o614Kraken     = "104a8f76-d178-4a33-bb59-e8767da3d982"
	o614Barbarians = "7e162cc1-9482-4812-8897-209ec2927ec0"
	o614Armodon    = "4c9f5359-249e-488b-9a7e-db1a43c2b1f4"
	o614Gargadon   = "7fe21ade-01f1-45ff-9b70-bc6302c3d284"
	o614Rats       = "23720758-f572-45e3-9bec-f6d3b974d7f5"
	o614Cockroach  = "c8b49b28-d364-4cf0-a2a5-56f6f7dc1b22"
	o614LoneWolf   = "f77dfcb6-72b0-4d50-8177-6f6c57a034b9"
	o614Okk        = "f835a8d3-e117-4669-9952-6e9ce13b35fa"
)

func TestS82OldB614EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		o614Deathgazer: "Deathgazer", o614Kraken: "Tidal Kraken", o614Barbarians: "Balduvian Barbarians",
		o614Armodon: "Trained Armodon", o614Gargadon: "Lesser Gargadon", o614Rats: "Swarm of Rats",
		o614Cockroach: "Giant Cockroach", o614LoneWolf: "Lone Wolf", o614Okk: "Okk",
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

func TestS82OldB614TidalKrakenCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.PushCatalogPermanent(g, me.ID, "Tidal Kraken", "Creature — Kraken", o614Kraken, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kraken, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, kraken); err == nil {
		t.Error("the Kraken was blocked")
	}
}

func TestS82OldB614SwarmOfRatsCountsYourRats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swarm := effectstest.PushCatalogPermanent(g, me.ID, "Swarm of Rats", "Creature — Rat", o614Rats, false)
	if got := effectstest.EffectivePower(t, g, swarm); got != 1 {
		t.Errorf("alone it is %d/1, want 1", got)
	}
	effectstest.B12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	if got := effectstest.EffectivePower(t, g, swarm); got != 3 {
		t.Errorf("with two other Rats it is %d, want 3 (the opponent's Rat and the Bear do not count)", got)
	}
	if _, tough := effectstest.EffectivePTOf(t, g, swarm); tough != 1 {
		t.Errorf("toughness %d, want 1", tough)
	}
}

func TestS82OldB614LoneWolfMayAssignDamageAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	wolf := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Lone Wolf", "Creature — Wolf", o614LoneWolf, false)
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.ApaLive(g, wolf).Effective().AssignsAsThoughUnblocked {
		t.Error("Lone Wolf cannot assign its damage as though unblocked")
	}
	if effectstest.ApaLive(g, bear).Effective().AssignsAsThoughUnblocked {
		t.Error("another creature got the ability")
	}
}

func TestS82OldB614OkkNeedsAGreaterPowerCreatureToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	okk := effectstest.PushCatalogPermanent(g, me.ID, "Okk", "Creature — Goblin", o614Okk, false)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(okk, opp.ID); err != nil {
		t.Fatalf("staging Okk: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with Okk alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(big, opp.ID); err != nil {
		t.Fatalf("the Ogre was refused: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, okk); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("Okk is not attacking beside the Ogre")
	}
}

func TestS82OldB614OkkNeedsAGreaterPowerCreatureToBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := effectstest.PushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := effectstest.PushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	okk := effectstest.PushCatalogPermanent(g, opp.ID, "Okk", "Creature — Goblin", o614Okk, false)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a1, Target: opp.ID}, {Attacker: a2, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	var re *game.BlockRefusedError
	if err := g.DeclareBlocker(okk, a1); !errors.As(err, &re) {
		t.Fatalf("Okk blocking alone: %v, want a refusal", err)
	}
	if err := g.DeclareBlocker(big, a1); err != nil {
		t.Fatalf("the Ogre block: %v", err)
	}
	if err := g.DeclareBlocker(okk, a2); err != nil {
		t.Fatalf("Okk beside the greater-power Ogre: %v", err)
	}
}

// o614SacrificeLand advances the step until the "sacrifice a land" prompt opens
// (a block locks in as priority passes) and answers it with the given land.
func o614SacrificeLand(t *testing.T, g *game.Game, seat, land uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		if pick := effectstest.LatestChooseCardsFor(g, seat); pick != nil {
			if err := g.ResolveChooseCards(pick.ID, seat, []uuid.UUID{land}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
		_, _ = g.AdvanceStep()
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestS82OldB614LesserGargadonSacrificesALandWhenItAttacksAndBlocks(t *testing.T) {
	// Attacking.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garg := effectstest.PushCatalogPermanent(g, me.ID, "Lesser Gargadon", "Creature — Beast", o614Gargadon, false)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.DeclareAttack(t, g, opp.ID, garg)
	effectstest.PassPriorityAroundTable(t, g)
	o614SacrificeLand(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the attack did not cost a land")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's land was sacrificed")
	}

	// Blocking.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	garg = effectstest.PushCatalogPermanent(g, opp.ID, "Lesser Gargadon", "Creature — Beast", o614Gargadon, false)
	land = effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, garg)
	effectstest.PassPriorityAroundTable(t, g)
	o614SacrificeLand(t, g, opp.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the block did not cost a land")
	}
}

func TestS82OldB614DeathgazerDestroysANonblackCreatureAtEndOfCombatEitherWay(t *testing.T) {
	// Deathgazer attacks and is blocked by a nonblack creature.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gazer := effectstest.PushCatalogPermanent(g, me.ID, "Deathgazer", "Creature — Lizard", o614Deathgazer, false)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, gazer, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Fatal("the blocker was destroyed before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the nonblack blocker survived end of combat")
	}

	// Deathgazer blocks a nonblack attacker.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 0, 9)
	gazer = effectstest.PushCatalogPermanent(g, opp.ID, "Deathgazer", "Creature — Lizard", o614Deathgazer, false)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, gazer)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the nonblack attacker survived end of combat")
	}
}

func TestS82OldB614DeathgazerLeavesABlackCreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gazer := effectstest.PushCatalogPermanent(g, me.ID, "Deathgazer", "Creature — Lizard", o614Deathgazer, false)
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Wall", TypeLine: "Creature — Zombie", Colors: []string{"B"},
		Power: 0, Toughness: 9, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, gazer, black)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(black) {
		t.Error("a black creature was destroyed")
	}
}
