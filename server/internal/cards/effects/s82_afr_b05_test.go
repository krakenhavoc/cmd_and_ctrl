package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}

func afrTreasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Treasure" {
				n++
			}
		}
	})
	return n
}

func TestAfrInspiringBardPumpsAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Inspiring Bard", "Creature — Elf Bard", afrInspiringBardOracle)
	afrPickMode(t, g, me.ID, 0)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	var p, tg int
	g.ReadSnapshot(func() { c := findBattlefieldCardForTest(g, bear); p, tg = c.CurrentPower(), c.CurrentToughness() })
	if p != 4 || tg != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tg)
	}
}

func TestAfrInspiringBardGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castAndResolveCreature(t, g, "Inspiring Bard", "Creature — Elf Bard", afrInspiringBardOracle)
	afrPickMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestAfrDawnbringerClericModes(t *testing.T) {
	t.Run("life", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		castAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 0)
		passPriorityAroundTable(t, g)
		if me.Life != life+2 {
			t.Errorf("life %d, want %d", me.Life, life+2)
		}
	})
	t.Run("destroy enchantment", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		castAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 1)
		if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, rock) || !hasID(p.PickTargetCards, ench) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		pickCard(t, g, me.ID, ench)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ench) {
			t.Error("enchantment survived")
		}
		if !g.Battlefield.Contains(rock) {
			t.Error("artifact was destroyed")
		}
	})
	t.Run("exile from graveyard", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dead := pushCatalogGraveyardCard(opp, "Dead Thing", "Creature — Bear", "", 1, 1)
		castAndResolveCreature(t, g, "Dawnbringer Cleric", "Creature — Human Cleric", afrDawnbringerClericOracle)
		afrPickMode(t, g, me.ID, 2)
		pickCard(t, g, me.ID, dead)
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(dead) {
			t.Error("the graveyard card was not exiled")
		}
	})
}

func TestAfrPlunderingBarbarianModes(t *testing.T) {
	t.Run("treasure", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		castAndResolveCreature(t, g, "Plundering Barbarian", "Creature — Dwarf Barbarian", afrPlunderingBarbarianOracl)
		afrPickMode(t, g, me.ID, 1)
		passPriorityAroundTable(t, g)
		if n := afrTreasures(g, me.ID); n != 1 {
			t.Errorf("%d Treasures, want 1", n)
		}
	})
	t.Run("destroy artifact", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		castAndResolveCreature(t, g, "Plundering Barbarian", "Creature — Dwarf Barbarian", afrPlunderingBarbarianOracl)
		afrPickMode(t, g, me.ID, 0)
		if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, ench) || !hasID(p.PickTargetCards, rock) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		pickCard(t, g, me.ID, rock)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("artifact survived")
		}
		if afrTreasures(g, me.ID) != 0 {
			t.Error("the Treasure mode also ran")
		}
	})
}

func TestAfrShamblingGhastDies(t *testing.T) {
	t.Run("treasure", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		ghast := b12Push(g, me.ID, "Shambling Ghast", "Creature — Zombie", afrShamblingGhastOracle, 1, 1)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ghast, game.DestroyOptions{}) })
		g.RunStateChecksForTest()
		afrPickMode(t, g, me.ID, 1)
		passPriorityAroundTable(t, g)
		if n := afrTreasures(g, me.ID); n != 1 {
			t.Errorf("%d Treasures, want 1", n)
		}
	})
	t.Run("shrink", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		ghast := b12Push(g, me.ID, "Shambling Ghast", "Creature — Zombie", afrShamblingGhastOracle, 1, 1)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ghast, game.DestroyOptions{}) })
		g.RunStateChecksForTest()
		afrPickMode(t, g, me.ID, 0)
		if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
			t.Fatalf("legal targets wrong: %+v", p)
		}
		pickCard(t, g, me.ID, theirs)
		passPriorityAroundTable(t, g)
		var p int
		g.ReadSnapshot(func() { p = findBattlefieldCardForTest(g, theirs).CurrentPower() })
		if p != 1 {
			t.Errorf("target power %d, want 1", p)
		}
	})
}

func TestAfrBalefulBeholderModes(t *testing.T) {
	t.Run("sacrifice", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
		mine := b12Permanent(g, me.ID, "My Aura", "Enchantment")
		castAndResolveCreature(t, g, "Baleful Beholder", "Creature — Beholder", afrBalefulBeholderOracle)
		afrPickMode(t, g, me.ID, 0)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			// a single candidate may still ask; answer it
			if p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID); p != nil {
				pickCard(t, g, opp.ID, theirs)
				passPriorityAroundTable(t, g)
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		castAndResolveCreature(t, g, "Baleful Beholder", "Creature — Beholder", afrBalefulBeholderOracle)
		afrPickMode(t, g, me.ID, 1)
		passPriorityAroundTable(t, g)
		var m, o bool
		g.ReadSnapshot(func() {
		})
		m = krgHas(t, g, mine, "menace")
		o = krgHas(t, g, theirs, "menace")
		if !m || o {
			t.Errorf("menace mine=%v theirs=%v, want true/false", m, o)
		}
	})
}

func TestAfrValorSingerPumpsAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Valor Singer", "Creature — Tiefling Bard", afrValorSingerOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	var p int
	g.ReadSnapshot(func() { p = findBattlefieldCardForTest(g, bear).CurrentPower() })
	if p != 3 {
		t.Errorf("target power %d, want 3", p)
	}
}

func TestAfrGoblinJavelineerPingsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := b12Push(g, me.ID, "Goblin Javelineer", "Creature — Goblin Warrior", afrGoblinJavelineerOracle, 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	declareAttack(t, g, opp.ID, gob)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, gob); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, wall)
	}
	passPriorityAroundTable(t, g)
	var dmg int
	g.ReadSnapshot(func() { dmg = findBattlefieldCardForTest(g, wall).DamageMarked })
	if dmg != 1 {
		t.Errorf("blocker has %d damage, want 1", dmg)
	}
}

func TestAfrManticoreDestroysOnlyADamagedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := b12Creature(g, opp.ID, "Hurt", "Creature — Bear", 2, 5)
	fresh := b12Creature(g, opp.ID, "Fresh", "Creature — Bear", 2, 5)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 5)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	castAndResolveCreature(t, g, "Manticore", "Creature — Manticore", afrManticoreOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		passPriorityAroundTable(t, g)
		p = latestPickTarget(g, me.ID)
	}
	if p == nil || !hasID(p.PickTargetCards, hurt) || hasID(p.PickTargetCards, fresh) || hasID(p.PickTargetCards, mine) {
		t.Fatalf("legal targets wrong: %+v", p)
	}
	pickCard(t, g, me.ID, hurt)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	merchant := castAndResolveCreature(t, g, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle)
	passPriorityAroundTable(t, g)
	if n := afrTreasures(g, me.ID); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
	hand := me.Hand.Size()
	fillPoolColored(me, "B", 1)
	fillPool(me, 1)
	p1g3Activate(t, g, merchant, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{afrTreasureID(g, me.ID)}})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if afrTreasures(g, me.ID) != 0 {
		t.Error("the Treasure was not sacrificed")
	}
}

func TestAfrSkullportMerchantSacrificesAnotherCreatureButNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	merchant := pushCatalogPermanent(g, me.ID, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "B", 1)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, merchant, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{merchant}}); err == nil {
		t.Error("sacrificing the merchant itself should be refused")
	}
	hand := me.Hand.Size()
	p1g3Activate(t, g, merchant, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) || me.Hand.Size() != hand+1 {
		t.Errorf("bear on battlefield %v, hand %d (want gone, %d)", g.Battlefield.Contains(bear), me.Hand.Size(), hand+1)
	}
}

func TestAfrSkullportMerchantRefusesWithoutFodder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	merchant := pushCatalogPermanent(g, me.ID, "Skullport Merchant", "Creature — Dwarf Citizen", afrSkullportMerchantOracle, false)
	fillPoolColored(me, "B", 1)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, merchant, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activation with nothing else to sacrifice should be refused")
	}
}

func TestAfrPotionOfHealing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	life := me.Life
	potion := castCatalogSpell(t, g, "Potion of Healing", "Artifact", afrPotionOfHealingOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (ETB draw)", me.Hand.Size(), hand+1)
	}
	if !g.Battlefield.Contains(potion) {
		t.Fatal("potion not on the battlefield")
	}
	p1g3Untap(g, potion)
	fillPoolColored(me, "W", 1)
	p1g3Activate(t, g, potion, 0, game.ActivateAbilityParams{})
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(potion) {
		t.Error("potion not sacrificed")
	}
}

func TestAfrPotionOfHealingNeedsWhiteMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	potion := pushCatalogPermanent(g, me.ID, "Potion of Healing", "Artifact", afrPotionOfHealingOracle, false)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, potion, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activation with only red mana should be refused")
	}
}

func TestAfrPortableHoleExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	dear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	hole := castCatalogSpell(t, g, "Portable Hole", "Artifact", afrPortableHoleOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, cheap) || hasID(p.PickTargetCards, dear) || hasID(p.PickTargetCards, land) {
		t.Fatalf("legal targets wrong: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	golem := typedCreature(g, me.ID, "Iron Golem", "Artifact Creature — Golem", afrIronGolemOracle, 5, 3)
	if !krgHas(t, g, golem, "vigilance") {
		t.Error("Iron Golem lacks vigilance")
	}
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
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
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	atk := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	golem := typedCreature(g, opp.ID, "Iron Golem", "Artifact Creature — Golem", afrIronGolemOracle, 5, 3)
	reqToBlockers(t, g, atk)
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, atk, golem); err != nil {
		t.Fatalf("blocking with the golem: %v", err)
	}
}
