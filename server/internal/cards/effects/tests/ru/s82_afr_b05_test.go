package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	afrInspiringBardOracle      = "88443470-f696-4f9b-b34d-50be8f1d4bc2"
	afrBalefulBeholderOracle    = "e7645ac4-6e05-426a-a62d-a33b30c33102"
	afrDawnbringerClericOracle  = "0146f430-da82-48d3-a576-c2b05101797f"
	afrPlunderingBarbarianOracl = "da2d27f7-cced-4118-bddc-41712e37e916"
	afrShamblingGhastOracle     = "e5e6229a-3422-4afd-9a87-86daa2d31d3a"
	afrValorSingerOracle        = "3e0d39e2-d99f-4729-899b-c7bcd2b53615"
	afrGoblinJavelineerOracle   = "e35c8206-98d0-44ab-8770-823a5ff7ab72"
	afrManticoreOracle          = "fe35b55e-52ae-428e-b27e-817ee78b5d8c"
	afrSkullportMerchantOracle  = "64ed5c35-8185-4e68-8028-b7b632ef5828"
	afrPotionOfHealingOracle    = "11fb3fee-ae84-4bf3-b834-63d3b8987df3"
	afrPortableHoleOracle       = "5e57a8cd-f4f0-40b2-84a0-e002751c0d40"
	afrIronGolemOracle          = "43990801-ed4a-49ff-a96e-03faa302214f"
)

func afrPickMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}

func TestAfrInspiringBardPumpsAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Inspiring Bard", "Creature — Elf Bard", afrInspiringBardOracle)
	afrPickMode(t, g, me.ID, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	var p, tg int
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, bear)
		p, tg = c.CurrentPower(), c.CurrentToughness()
	})
	if p != 4 || tg != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tg)
	}
}

func TestAfrInspiringBardGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Inspiring Bard", "Creature — Elf Bard", afrInspiringBardOracle)
	afrPickMode(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestAfrDawnbringerClericModes(t *testing.T) {
	t.Run("life", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+2 {
			t.Errorf("life %d, want %d", me.Life, life+2)
		}
	})
	t.Run("destroy enchantment", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		effectstest.CastAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 1)
		if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, rock) || !effectstest.HasID(p.PickTargetCards, ench) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		effectstest.PickCard(t, g, me.ID, ench)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ench) {
			t.Error("enchantment survived")
		}
		if !g.Battlefield.Contains(rock) {
			t.Error("artifact was destroyed")
		}
	})
	t.Run("exile from graveyard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dead := effectstest.PushCatalogGraveyardCard(opp, "Dead Thing", "Creature — Bear", "", 1, 1)
		effectstest.CastAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 2)
		effectstest.PickCard(t, g, me.ID, dead)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(dead) {
			t.Error("the graveyard card was not exiled")
		}
	})
}

func TestAfrPlunderingBarbarianModes(t *testing.T) {
	t.Run("treasure", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.CastAndResolveCreature(t, g, "Plundering Barbarian", "Creature — Dwarf Barbarian", afrPlunderingBarbarianOracl)
		afrPickMode(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.AfrTreasures(g, me.ID); n != 1 {
			t.Errorf("%d Treasures, want 1", n)
		}
	})
	t.Run("destroy artifact", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		effectstest.CastAndResolveCreature(t, g, "Plundering Barbarian", "Creature — Dwarf Barbarian", afrPlunderingBarbarianOracl)
		afrPickMode(t, g, me.ID, 0)
		if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, ench) || !effectstest.HasID(p.PickTargetCards, rock) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		effectstest.PickCard(t, g, me.ID, rock)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("artifact survived")
		}
		if effectstest.AfrTreasures(g, me.ID) != 0 {
			t.Error("the Treasure mode also ran")
		}
	})
}

func TestAfrShamblingGhastDies(t *testing.T) {
	t.Run("treasure", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ghast := effectstest.B12Push(g, me.ID, "Shambling Ghast", "Creature — Zombie", afrShamblingGhastOracle, 1, 1)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ghast, game.DestroyOptions{}) })
		g.RunStateChecksForTest()
		afrPickMode(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.AfrTreasures(g, me.ID); n != 1 {
			t.Errorf("%d Treasures, want 1", n)
		}
	})
	t.Run("shrink", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		ghast := effectstest.B12Push(g, me.ID, "Shambling Ghast", "Creature — Zombie", afrShamblingGhastOracle, 1, 1)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ghast, game.DestroyOptions{}) })
		g.RunStateChecksForTest()
		afrPickMode(t, g, me.ID, 0)
		if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		var p int
		g.ReadSnapshot(func() { p = effectstest.FindBattlefieldCardForTest(g, theirs).CurrentPower() })
		if p != 1 {
			t.Errorf("target power %d, want 1", p)
		}
	})
}

func TestAfrBalefulBeholderModes(t *testing.T) {
	t.Run("sacrifice", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		mine := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
		effectstest.CastAndResolveCreature(t, g, "Baleful Beholder", "Creature — Beholder", afrBalefulBeholderOracle)
		afrPickMode(t, g, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			// a single candidate may still ask; answer it
			if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID); p != nil {
				effectstest.PickCard(t, g, opp.ID, theirs)
				effectstest.PassPriorityAroundTable(t, g)
			}
		}
		if g.Battlefield.Contains(theirs) {
			t.Error("the opponent's enchantment was not sacrificed")
		}
		if !g.Battlefield.Contains(mine) {
			t.Error("your own enchantment was sacrificed")
		}
	})
	t.Run("menace", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Baleful Beholder", "Creature — Beholder", afrBalefulBeholderOracle)
		afrPickMode(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		var m, o bool
		g.ReadSnapshot(func() {
		})
		m = effectstest.KrgHas(t, g, mine, "menace")
		o = effectstest.KrgHas(t, g, theirs, "menace")
		if !m || o {
			t.Errorf("menace mine=%v theirs=%v, want true/false", m, o)
		}
	})
}

func TestAfrValorSingerPumpsAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Valor Singer", "Creature — Tiefling Bard", afrValorSingerOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var p int
	g.ReadSnapshot(func() { p = effectstest.FindBattlefieldCardForTest(g, bear).CurrentPower() })
	if p != 3 {
		t.Errorf("target power %d, want 3", p)
	}
}

func TestAfrGoblinJavelineerPingsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := effectstest.B12Push(g, me.ID, "Goblin Javelineer", "Creature — Goblin Warrior", afrGoblinJavelineerOracle, 1, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.DeclareAttack(t, g, opp.ID, gob)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, gob); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, wall)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var dmg int
	g.ReadSnapshot(func() { dmg = effectstest.FindBattlefieldCardForTest(g, wall).DamageMarked })
	if dmg != 1 {
		t.Errorf("blocker has %d damage, want 1", dmg)
	}
}

func TestAfrManticoreDestroysOnlyADamagedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt", "Creature — Bear", 2, 5)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", "Creature — Bear", 2, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 5)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	effectstest.CastAndResolveCreature(t, g, "Manticore", "Creature — Manticore", afrManticoreOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityAroundTable(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p == nil || !effectstest.HasID(p.PickTargetCards, hurt) || effectstest.HasID(p.PickTargetCards, fresh) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("legal targets wrong: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !g.Battlefield.Contains(fresh) || !g.Battlefield.Contains(mine) {
		t.Error("a wrong creature was destroyed")
	}
}

func afrTreasureID(g *game.Game, owner uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Treasure" {
				id = c.InstanceID
			}
		}
	})
	return id
}

func TestAfrSkullportMerchantTreasureAndSacrificeDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	merchant := effectstest.CastAndResolveCreature(t, g, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AfrTreasures(g, me.ID); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
	hand := me.Hand.Size()
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPool(me, 1)
	effectstest.P1g3Activate(t, g, merchant, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{afrTreasureID(g, me.ID)}})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if effectstest.AfrTreasures(g, me.ID) != 0 {
		t.Error("the Treasure was not sacrificed")
	}
}

func TestAfrSkullportMerchantSacrificesAnotherCreatureButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	merchant := effectstest.PushCatalogPermanent(g, me.ID, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, merchant, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{merchant}}); err == nil {
		t.Error("sacrificing the merchant itself should be refused")
	}
	hand := me.Hand.Size()
	effectstest.P1g3Activate(t, g, merchant, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) || me.Hand.Size() != hand+1 {
		t.Errorf("bear on battlefield %v, hand %d (want gone, %d)", g.Battlefield.Contains(bear), me.Hand.Size(), hand+1)
	}
}

func TestAfrSkullportMerchantRefusesWithoutFodder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	merchant := effectstest.PushCatalogPermanent(g, me.ID, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle, false)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, merchant, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activation with nothing else to sacrifice should be refused")
	}
}

func TestAfrPotionOfHealing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	life := me.Life
	potion := effectstest.CastCatalogSpell(t, g, "Potion of Healing", "Artifact", afrPotionOfHealingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (ETB draw)", me.Hand.Size(), hand+1)
	}
	if !g.Battlefield.Contains(potion) {
		t.Fatal("potion not on the battlefield")
	}
	effectstest.P1g3Untap(g, potion)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.P1g3Activate(t, g, potion, 0, game.ActivateAbilityParams{})
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(potion) {
		t.Error("potion not sacrificed")
	}
}

func TestAfrPotionOfHealingNeedsWhiteMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	potion := effectstest.PushCatalogPermanent(g, me.ID, "Potion of Healing", "Artifact", afrPotionOfHealingOracle, false)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, potion, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activation with only red mana should be refused")
	}
}

func TestAfrPortableHoleExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	dear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	hole := effectstest.CastCatalogSpell(t, g, "Portable Hole", "Artifact", afrPortableHoleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, cheap) || effectstest.HasID(p.PickTargetCards, dear) || effectstest.HasID(p.PickTargetCards, land) {
		t.Fatalf("legal targets wrong: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Fatal("the cheap permanent was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hole, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
	back := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Cheap" && c.Controller == opp.ID {
				back++
			}
		}
	})
	if back != 1 {
		t.Errorf("%d copies returned, want 1", back)
	}
}

func TestAfrIronGolemMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	golem := effectstest.TypedCreature(g, me.ID, "Iron Golem", "Artifact Creature — Golem", afrIronGolemOracle, 5, 3)
	if !effectstest.KrgHas(t, g, golem, "vigilance") {
		t.Error("Iron Golem lacks vigilance")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the golem home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(golem, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass after attacking: %v", err)
	}
}

func TestAfrIronGolemMustBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	golem := effectstest.TypedCreature(g, opp.ID, "Iron Golem", "Artifact Creature — Golem", afrIronGolemOracle, 5, 3)
	effectstest.ReqToBlockers(t, g, atk)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, atk, golem); err != nil {
		t.Fatalf("blocking with the golem: %v", err)
	}
}
