package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b603_cards_test.go — Mirrodin, S82 OLD batch 603 (#5606). Each
// test plays the real catalog entry and asserts what its text says.

const (
	oldB603War      = "07947d01-ebb1-46b7-9bdd-ea00e034805b"
	oldB603Choker   = "3241abc1-d739-4ac3-93d1-7fc2898b943b"
	oldB603Prototyp = "496e0fbb-cb01-41d6-9c56-23e3ae2bb4f7"
	oldB603Strider  = "6e52e48f-b7a0-4195-8fff-8baed7b7d3fc"
	oldB603Ogre     = "0a64b683-7017-44cb-923f-b512fe3886f6"
	oldB603Nova     = "30dc2a1a-c17e-4e95-9dd8-d0e9be1dcd77"
	oldB603Dust     = "56828166-eaa3-4711-91b6-401a3e3b733f"
	oldB603Detonate = "daa90a75-c600-41bd-9311-ec21cf51480b"
	oldB603Rain     = "e9774695-b2f1-4014-8504-04b2339be568"
)

func TestOldB603CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB603War: "War Elemental", oldB603Choker: "Jinxed Choker", oldB603Prototyp: "Myr Prototype",
		oldB603Strider: "Slith Strider", oldB603Ogre: "Rustmouth Ogre", oldB603Nova: "Soul Nova",
		oldB603Dust: "Turn to Dust", oldB603Detonate: "Detonate", oldB603Rain: "Molten Rain",
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

func oldB603Damage(g *game.Game, source, target uuid.UUID, n int) {
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, target, n); err != nil {
			panic(err)
		}
	})
	g.RunStateChecksForTest()
}

func TestOldB603WarElementalIsSacrificedUnlessAnOpponentWasDealtDamage(t *testing.T) {
	for _, hurt := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if hurt {
			oldB603Damage(g, me.ID, opp.ID, 1)
		}
		elemental := effectstest.CastAndResolveCreature(t, g, "War Elemental", "Creature — Elemental", oldB603War)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(elemental) != hurt {
			t.Errorf("opponent dealt damage=%v: the Elemental on the battlefield=%v, want %v", hurt, g.Battlefield.Contains(elemental), hurt)
		}
	}
}

func TestOldB603WarElementalGrowsByTheDamageDealtToOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elemental := effectstest.B12Push(g, me.ID, "War Elemental", "Creature — Elemental", oldB603War, 1, 1)
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	oldB603Damage(g, pinger, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.RfPT(t, g, elemental); p != 4 || tough != 4 {
		t.Errorf("after 3 damage to an opponent the Elemental is %d/%d, want 4/4", p, tough)
	}
	oldB603Damage(g, pinger, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.RfPT(t, g, elemental); p != 4 {
		t.Errorf("damage to its controller changed the Elemental to %d power", p)
	}
	oldB603Damage(g, pinger, g.Seats[2].ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.RfPT(t, g, elemental); p != 6 {
		t.Errorf("after 2 damage to another opponent the Elemental has %d power, want 6", p)
	}
}

func TestOldB603MyrPrototypeGrowsAndTaxesItsAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	proto := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Myr Prototype", TypeLine: "Artifact Creature — Myr",
		OracleID: oldB603Prototyp, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	oppProto := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Myr Prototype", TypeLine: "Artifact Creature — Myr",
		OracleID: oldB603Prototyp, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 2}})
	_ = oppProto
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 1, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, proto); c.Counters["+1/+1"] != 1 {
		t.Fatalf("%d +1/+1 counters after the upkeep, want 1", c.Counters["+1/+1"])
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(proto, opp.ID); err == nil {
		t.Fatal("the Prototype attacked without paying {1} for its counter")
	}
	effectstest.B06AddMana(me, "R")
	if err := g.DeclareAttacker(proto, opp.ID); err != nil {
		t.Fatalf("the Prototype could not attack with {1} available: %v", err)
	}
	if err := g.DeclareAttacker(plain, opp.ID); err != nil {
		t.Fatalf("an untaxed creature could not attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(oppProto, plain); err != nil {
		t.Fatalf("staging the block: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); !errors.Is(err, game.ErrBlockTaxUnpaid) {
		t.Fatalf("finishing the block with no mana: %v, want ErrBlockTaxUnpaid", err)
	}
	// The opponent's Prototype grew a third counter in its own upkeep on the way here.
	effectstest.B06AddMana(opp, "R", "R")
	if err := g.FinishBlocks(opp.ID); !errors.Is(err, game.ErrBlockTaxUnpaid) {
		t.Fatalf("finishing the block with {2} for three counters: %v, want ErrBlockTaxUnpaid", err)
	}
	effectstest.B06AddMana(opp, "R")
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Errorf("the Prototype could not block with {3} available for its three counters: %v", err)
	}
}

func TestOldB603SlithStriderDrawsWhenBlockedAndGrowsWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strider := effectstest.B12Push(g, me.ID, "Slith Strider", "Creature — Slith", oldB603Strider, 1, 1)
	blockerA := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 5)
	blockerB := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 5)
	effectstest.BrAttack(t, g, strider)
	hand := me.Hand.Size()
	for _, b := range []uuid.UUID{blockerA, blockerB} {
		if err := g.DeclareBlocker(b, strider); err != nil {
			t.Fatalf("block: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d after a double block, want exactly one card drawn", hand, me.Hand.Size())
	}
	if c, _ := effectstest.BattlefieldCard(g, strider); c.Counters["+1/+1"] != 0 {
		t.Error("a blocked Strider got a counter")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	strider2 := effectstest.B12Push(g2, me2.ID, "Slith Strider", "Creature — Slith", oldB603Strider, 1, 1)
	life := g2.Seats[1].Life
	effectstest.BrAttack(t, g2, strider2)
	effectstest.AdvanceTo(t, g2, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Seats[1].Life != life-1 {
		t.Fatalf("the unblocked Strider did not connect (life %d → %d)", life, g2.Seats[1].Life)
	}
	if c, _ := effectstest.BattlefieldCard(g2, strider2); c.Counters["+1/+1"] != 1 {
		t.Errorf("%d +1/+1 counters after connecting, want 1", c.Counters["+1/+1"])
	}
}

func TestOldB603RustmouthOgreMayDestroyAnArtifactOfTheDamagedPlayer(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		ogre := effectstest.B12Push(g, me.ID, "Rustmouth Ogre", "Creature — Ogre", oldB603Ogre, 5, 4)
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		elsewhere := effectstest.B12Permanent(g, other.ID, "Elsewhere Rock", "Artifact")
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.BrAttack(t, g, ogre)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.B04WaitForPick(t, g, me.ID)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil || effectstest.HasID(p.PickTargetCards, elsewhere) || effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, rock) {
			t.Fatalf("pick offered %+v: only the damaged player's artifacts", p)
		}
		effectstest.PickCard(t, g, me.ID, rock)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) == yes {
			t.Errorf("yes=%v: the artifact on the battlefield=%v", yes, g.Battlefield.Contains(rock))
		}
		if !g.Battlefield.Contains(elsewhere) || !g.Battlefield.Contains(bear) {
			t.Error("a permanent that was not a legal target was destroyed")
		}
	}
}

func TestOldB603SoulNovaExilesTheAttackerAndItsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 1, 1)
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: attacker}})
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Other Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: idle}})
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if _, err := effectstest.RprB19Cast(t, g, opp, "Soul Nova", "Instant", "{3}{W}{W}", oldB603Nova, "{W}{W}{W}{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	g.WithWriteLock(func() { opp.ManaPool = game.ManaPool{} })
	if _, err := effectstest.RprB19Cast(t, g, opp, "Soul Nova", "Instant", "{3}{W}{W}", oldB603Nova, "{W}{W}{W}{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, attacker) || !effectstest.InExile(g, sword) {
		t.Error("the attacker or its Equipment was not exiled")
	}
	if !g.Battlefield.Contains(idle) || !g.Battlefield.Contains(other) {
		t.Error("something that was not the target or attached to it left the battlefield")
	}
}

func TestOldB603TurnToDustDestroysAnEquipmentAndAddsGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sword := effectstest.B12Permanent(g, g.Seats[1].ID, "Sword", "Artifact — Equipment")
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if _, err := effectstest.RprB19Cast(t, g, me, "Turn to Dust", "Instant", "{G}", oldB603Dust, "{G}", game.CastSpellParams{Targets: effectstest.CardTarget(rock)}); err == nil {
		t.Fatal("a non-Equipment artifact was accepted as the target")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if _, err := effectstest.RprB19Cast(t, g, me, "Turn to Dust", "Instant", "{G}", oldB603Dust, "{G}", game.CastSpellParams{Targets: effectstest.CardTarget(sword)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sword) || !g.Battlefield.Contains(rock) {
		t.Error("the Equipment survived or the plain artifact was destroyed")
	}
	if n := len(me.ManaPool.Units()); n != 1 {
		t.Errorf("%d mana in the pool, want the {G} the spell adds", n)
	}
}

func TestOldB603DetonateNeedsManaValueXAndHurtsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	three := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	if _, err := effectstest.RprB19Cast(t, g, me, "Detonate", "Sorcery", "{X}{R}", oldB603Detonate, "{C}{C}{R}", game.CastSpellParams{XValue: 2, Targets: effectstest.CardTarget(three)}); err == nil {
		t.Fatal("an artifact with mana value 3 was accepted for X = 2")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	life := opp.Life
	if _, err := effectstest.RprB19Cast(t, g, me, "Detonate", "Sorcery", "{X}{R}", oldB603Detonate, "{C}{C}{R}", game.CastSpellParams{XValue: 2, Targets: effectstest.CardTarget(two)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) || !g.Battlefield.Contains(three) {
		t.Error("the artifact with mana value 2 survived or the one with mana value 3 was destroyed")
	}
	if opp.Life != life-2 {
		t.Errorf("the artifact's controller's life %d → %d, want -2", life, opp.Life)
	}
}

func TestOldB603MoltenRainHurtsOnlyForANonbasicLand(t *testing.T) {
	for _, basic := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		typeLine := "Land"
		if basic {
			typeLine = "Basic Land — Mountain"
		}
		land := effectstest.PushLandFor(g, opp.ID, "Target Land", typeLine)
		life := opp.Life
		if _, err := effectstest.RprB19Cast(t, g, me, "Molten Rain", "Sorcery", "{1}{R}{R}", oldB603Rain, "{C}{R}{R}", game.CastSpellParams{Targets: effectstest.CardTarget(land)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) {
			t.Errorf("basic=%v: the land was not destroyed", basic)
		}
		want := life - 2
		if basic {
			want = life
		}
		if opp.Life != want {
			t.Errorf("basic=%v: the land's controller's life %d → %d, want %d", basic, life, opp.Life, want)
		}
	}
}

func TestOldB603JinxedChokerChangesHandsWithACounterAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	choker := effectstest.B12Push(g, me.ID, "Jinxed Choker", "Artifact", oldB603Choker, 0, 0)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, choker)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("the Choker is controlled by %v, want the targeted opponent", c.Controller)
	}
	if c.Counters["charge"] != 1 {
		t.Errorf("%d charge counters, want 1", c.Counters["charge"])
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("its new controller's life %d → %d, want -1 for the one charge counter", life, opp.Life)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	for i, want := range []int{2, 1} {
		effectstest.B06AddMana(opp, "C", "C", "C")
		if err := g.ActivateCatalogAbility(opp.ID, choker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, opp.ID, i)
		if c, _ := effectstest.BattlefieldCard(g, choker); c.Counters["charge"] != want {
			t.Errorf("after activation %d: %d charge counters, want %d", i, c.Counters["charge"], want)
		}
	}
}

func TestOldB603DetonateStillHurtsTheControllerOfAnIndestructibleArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tough Rock", TypeLine: "Artifact", ManaCost: "{1}",
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	life := opp.Life
	if _, err := effectstest.RprB19Cast(t, g, me, "Detonate", "Sorcery", "{X}{R}", oldB603Detonate, "{C}{R}", game.CastSpellParams{XValue: 1, Targets: effectstest.CardTarget(tough)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible artifact was destroyed")
	}
	if opp.Life != life-1 {
		t.Errorf("the artifact's controller's life %d → %d, want -1 even though it survived", life, opp.Life)
	}
}
