package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b644_cards_test.go — Planeshift's enchantments, removal, creatures,
// counterspell, land and mana, batch 644 (#5615).

const (
	b644Cover       = "a4cd8dd1-adb9-4c4c-bda9-6f6b6c167702"
	b644Twilight    = "c886a076-dbe2-4d80-8dab-537b84d0837a"
	b644Flow        = "f7544a96-46b3-4cdc-8b14-c7b6daecdabd"
	b644Bloodstock  = "dfeaec62-ca77-46dc-8eae-f627a5e1c02a"
	b644DeathBomb   = "e323d468-def5-4090-af5b-19481289168f"
	b644Greevil     = "e43b25be-fe67-4eaa-80df-85718f38f904"
	b644Thornscape  = "3f92c684-f2fd-4fe5-bfb5-e27061d2c45b"
	b644Jailer      = "423659f1-1b48-4a2b-a1dc-f39fd62c6b6a"
	b644Routes      = "65260087-48b5-4968-8584-7cc561bb312c"
	b644Trickery    = "93d0691f-17b8-458b-a624-e182fc481f49"
	b644City        = "6bb00a28-8b5a-4049-93b7-3db02de88aeb"
	b644March       = "bca7101b-e74b-42af-ac63-1b3defb4137d"
	b644Toad        = "c615540b-bf85-4e16-8ac2-9d541cba732f"
	b644BurstOracle = "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b"
	b644BoltOracle  = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB644EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b644Cover: "Cloud Cover", b644Twilight: "Keldon Twilight", b644Flow: "Destructive Flow",
		b644Bloodstock: "Phyrexian Bloodstock", b644DeathBomb: "Death Bomb", b644Greevil: "Root Greevil",
		b644Thornscape: "Thornscape Familiar", b644Jailer: "Mogg Jailer", b644Routes: "Escape Routes",
		b644Trickery: "Ertai's Trickery", b644City: "Forsaken City", b644March: "March of Souls", b644Toad: "Morgue Toad",
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

func b644Table(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func b644Perm(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func b644HandCard(g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	return c.InstanceID
}

func b644Sacrifice(t *testing.T, g *game.Game, who uuid.UUID, pick uuid.UUID) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
	if c == nil {
		t.Fatalf("no sacrifice prompt for %s", who)
	}
	if err := g.ResolveChooseCards(c.ID, who, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func TestOldB644CloudCoverReturnsAnotherPermanentTargetedByAnOpponent(t *testing.T) {
	g, _, opp := b644Table(t)
	// Cloud Cover is the opponent's; I (the active player) target their creature.
	effectstest.PushCatalogPermanent(g, opp.ID, "Cloud Cover", "Enchantment", b644Cover, false)
	theirs := b644Perm(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", b644BoltOracle, effectstest.CardRefs(theirs))
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the targeted Bear was not returned to its owner's hand")
	}
}

func TestOldB644CloudCoverIgnoresYourOwnSpellsAndItself(t *testing.T) {
	g, me, _ := b644Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Cloud Cover", "Enchantment", b644Cover, false)
	mine := b644Perm(g, me.ID, "Bear", "Creature — Bear", 2, 5, "G")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", b644BoltOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || me.Hand.Contains(mine) {
		t.Error("my own spell's target was returned although no opponent controls the spell")
	}
}

func TestOldB644KeldonTwilightSacrificesAnOldCreatureOnlyWhenNobodyAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g, me, opp := b644Table(t)
		effectstest.PushCatalogPermanent(g, me.ID, "Keldon Twilight", "Enchantment", b644Twilight, false)
		old1 := effectstest.PushVanillaCreature(g, me.ID, "Old One", 1, 1)
		old2 := effectstest.PushVanillaCreature(g, me.ID, "Old Two", 1, 1)
		fresh := effectstest.PushCatalogPermanent(g, me.ID, "Fresh", "Creature — Bear", "", true)
		if attacked {
			effectstest.DeclareAttack(t, g, opp.ID, old1)
		}
		effectstest.AdvanceToEndStep(t, g)
		if attacked {
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(old1) || !g.Battlefield.Contains(old2) {
				t.Fatal("a creature was sacrificed although creatures attacked")
			}
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if c == nil {
			t.Fatal("no sacrifice prompt")
		}
		if effectstest.HasID(c.ChooseCards, fresh) {
			t.Error("a creature that came under my control this turn is offered")
		}
		b644Sacrifice(t, g, me.ID, old2)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(old2) || !g.Battlefield.Contains(old1) || !g.Battlefield.Contains(fresh) {
			t.Error("exactly the chosen creature should have been sacrificed")
		}
	}
}

func TestOldB644DestructiveFlowSacrificesANonbasicLandEachUpkeep(t *testing.T) {
	g, _, opp := b644Table(t)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Destructive Flow", "Enchantment", b644Flow, false)
	basic := b644Perm(g, opp.ID, "Forest", "Basic Land — Forest", 0, 0)
	nonbasic := b644Perm(g, opp.ID, "Odd Land", "Land", 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID); c == nil || effectstest.HasID(c.ChooseCards, basic) || !effectstest.HasID(c.ChooseCards, nonbasic) {
		t.Fatalf("the prompt should offer only the nonbasic land: %+v", c)
	}
	b644Sacrifice(t, g, opp.ID, nonbasic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nonbasic) {
		t.Error("the nonbasic land was not sacrificed")
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land was sacrificed")
	}
}

func TestOldB644PhyrexianBloodstockDestroysAWhiteCreatureWhenItLeaves(t *testing.T) {
	g, me, opp := b644Table(t)
	stock := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Bloodstock", "Creature — Phyrexian Zombie", b644Bloodstock, false)
	white := b644Perm(g, opp.ID, "Soldier", "Creature — Soldier", 1, 1, "W")
	green := b644Perm(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(stock) })
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the leaves trigger asked for no target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(green)); err == nil {
		t.Fatal("a green creature was accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(white)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || !g.Battlefield.Contains(green) {
		t.Error("only the white creature should have been destroyed")
	}
}

func TestOldB644DeathBombSacrificesDestroysAndDrainsTwo(t *testing.T) {
	g, me, opp := b644Table(t)
	fodder := b644Perm(g, me.ID, "Fodder", "Creature — Bear", 1, 1, "G")
	victim := b644Perm(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4, "R")
	black := b644Perm(g, opp.ID, "Zombie", "Creature — Zombie", 1, 1, "B")
	card := b644HandCard(g, me, game.Card{Name: "Death Bomb", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: b644DeathBomb})
	effectstest.B06AddMana(me, "B", "C", "C", "C")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(black), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("a black creature was accepted")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(victim), Strict: true}); err == nil {
		t.Fatal("a cast without the sacrifice was accepted")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid on cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Ogre is still on the battlefield")
	}
	if opp.Life != life-2 {
		t.Errorf("the Ogre's controller's life %d → %d, want -2", life, opp.Life)
	}
}

func TestOldB644RootGreevilDestroysEnchantmentsOfTheChosenColour(t *testing.T) {
	g, me, opp := b644Table(t)
	greevil := effectstest.HmlCreature(g, me.ID, "Root Greevil", b644Greevil, 2, 3)
	white := b644Perm(g, opp.ID, "White Ench", "Enchantment", 0, 0, "W")
	green := b644Perm(g, me.ID, "Green Ench", "Enchantment", 0, 0, "G")
	colourless := b644Perm(g, opp.ID, "Plain Ench", "Enchantment", 0, 0)
	effectstest.B06AddMana(me, "G", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, greevil, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(greevil) {
		t.Error("the Greevil was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "W")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) {
		t.Error("the white enchantment survived")
	}
	if !g.Battlefield.Contains(green) || !g.Battlefield.Contains(colourless) {
		t.Error("an enchantment of another colour was destroyed")
	}
}

func TestOldB644ThornscapeFamiliarDiscountsRedAndWhiteSpells(t *testing.T) {
	g, me, _ := b644Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Thornscape Familiar", "Creature — Insect", b644Thornscape, false)
	green := b644HandCard(g, me, game.Card{Name: "Green", TypeLine: "Instant", ManaCost: "{1}{G}", Colors: []string{"G"}})
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a green spell got a discount")
	}
	me.ManaPool = game.ManaPool{}
	for _, colour := range []string{"R", "W"} {
		c := b644HandCard(g, me, game.Card{Name: colour, TypeLine: "Instant", ManaCost: "{1}{" + colour + "}", Colors: []string{colour}})
		effectstest.B06AddMana(me, colour)
		if err := g.CastSpell(me.ID, c, game.CastSpellParams{Strict: true}); err != nil {
			t.Errorf("a %s spell for one mana: %v", colour, err)
		}
		me.ManaPool = game.ManaPool{}
	}
}

func TestOldB644MoggJailerCantAttackIntoAnUntappedSmallCreature(t *testing.T) {
	for _, tc := range []struct {
		name         string
		power        int
		tapped       bool
		canAttackNow bool
	}{
		{"untapped power 2", 2, false, false},
		{"untapped power 1", 1, false, false},
		{"untapped power 3", 3, false, true},
		{"tapped power 2", 2, true, true},
	} {
		g, me, opp := b644Table(t)
		jailer := effectstest.HmlCreature(g, me.ID, "Mogg Jailer", b644Jailer, 2, 2)
		guard := effectstest.PushVanillaCreature(g, opp.ID, "Guard", tc.power, 3)
		if tc.tapped {
			g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, guard).Tapped = true })
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(jailer, opp.ID)
		if tc.canAttackNow && err != nil {
			t.Errorf("%s: the Jailer could not attack: %v", tc.name, err)
		}
		if !tc.canAttackNow && err == nil {
			t.Errorf("%s: the Jailer attacked", tc.name)
		}
	}
}

func TestOldB644EscapeRoutesReturnsAWhiteOrBlackCreatureYouControl(t *testing.T) {
	g, me, opp := b644Table(t)
	routes := effectstest.PushCatalogPermanent(g, me.ID, "Escape Routes", "Enchantment", b644Routes, false)
	white := b644Perm(g, me.ID, "Soldier", "Creature — Soldier", 1, 1, "W")
	green := b644Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b644Perm(g, opp.ID, "Zombie", "Creature — Zombie", 1, 1, "B")
	effectstest.B06AddMana(me, "U", "C", "C")
	for _, bad := range []uuid.UUID{green, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, routes, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Fatal("an illegal target was accepted")
		}
	}
	effectstest.B16Activate(t, g, me.ID, routes, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(white)})
	if g.Battlefield.Contains(white) || !me.Hand.Contains(white) {
		t.Error("the white creature did not return to hand")
	}
}

func TestOldB644ErtaisTrickeryCountersOnlyAKickedSpell(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := b644Table(t)
		burst := b644HandCard(g, me, game.Card{Name: "Burst Lightning", TypeLine: "Instant", ManaCost: "{R}", OracleID: b644BurstOracle, Colors: []string{"R"}})
		effectstest.B06AddMana(me, "R")
		p := game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}
		if kicked {
			effectstest.B06AddMana(me, "C", "C", "C", "C")
			p.OptionalCosts = []int{0}
		}
		if err := g.CastSpell(me.ID, burst, p); err != nil {
			t.Fatalf("cast Burst Lightning: %v", err)
		}
		life := opp.Life
		trick := b644HandCard(g, me, game.Card{Name: "Ertai's Trickery", TypeLine: "Instant", ManaCost: "{U}", OracleID: b644Trickery, Colors: []string{"U"}})
		effectstest.B06AddMana(me, "U")
		if err := g.CastSpell(me.ID, trick, game.CastSpellParams{Targets: effectstest.CardRefs(burst), Strict: true}); err != nil {
			t.Fatalf("cast Ertai's Trickery: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if kicked && opp.Life != life {
			t.Errorf("the kicked Burst Lightning resolved: life %d → %d", life, opp.Life)
		}
		if !kicked && opp.Life >= life {
			t.Errorf("the unkicked Burst Lightning was countered: life %d → %d", life, opp.Life)
		}
	}
}

func TestOldB644ForsakenCityStaysTappedUnlessYouExileACard(t *testing.T) {
	for _, exile := range []bool{false, true} {
		g, me, _ := b644Table(t)
		city := effectstest.PushCatalogPermanent(g, me.ID, "Forsaken City", "Land", b644City, false)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, city).Tapped = true })
		me.Hand.Cards = nil
		keep := b644HandCard(g, me, game.Card{Name: "Spare", TypeLine: "Sorcery"})
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		if effectstest.TappedForTest(t, g, city) != true {
			t.Fatal("the land untapped during the untap step")
		}
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		for i := 0; c == nil && i < 4; i++ {
			effectstest.PassPriorityAroundTable(t, g)
			c = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		}
		if c == nil {
			t.Fatal("the upkeep trigger asked for no card")
		}
		var picks []uuid.UUID
		if exile {
			picks = []uuid.UUID{keep}
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		tapped := effectstest.TappedForTest(t, g, city)
		if exile && (tapped || me.Hand.Contains(keep)) {
			t.Errorf("exiled a card: tapped=%v, still in hand=%v", tapped, me.Hand.Contains(keep))
		}
		if !exile && (!tapped || !me.Hand.Contains(keep)) {
			t.Errorf("declined: tapped=%v, still in hand=%v", tapped, me.Hand.Contains(keep))
		}
	}
}

func TestOldB644MarchOfSoulsDestroysAllAndMakesASpiritForEachControllerCreature(t *testing.T) {
	g, me, opp := b644Table(t)
	mine1 := b644Perm(g, me.ID, "A", "Creature — Bear", 1, 1, "G")
	mine2 := b644Perm(g, me.ID, "B", "Creature — Bear", 1, 1, "G")
	theirs := b644Perm(g, opp.ID, "C", "Creature — Bear", 1, 1, "G")
	card := b644HandCard(g, me, game.Card{Name: "March of Souls", TypeLine: "Sorcery", ManaCost: "{4}{W}", OracleID: b644March, Colors: []string{"W"}})
	effectstest.B06AddMana(me, "W", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine1, mine2, theirs} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature survived")
		}
	}
	count := func(owner uuid.UUID) (n int) {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spirit" && c.Controller == owner && c.IsCreature() && c.HasColorWord("W") {
				n++
			}
		}
		return n
	}
	if count(me.ID) != 2 || count(opp.ID) != 1 {
		t.Errorf("Spirits: me %d (want 2), opponent %d (want 1)", count(me.ID), count(opp.ID))
	}
}

func TestOldB644MorgueToadSacrificesForBlueAndRed(t *testing.T) {
	g, me, _ := b644Table(t)
	toad := effectstest.PushCatalogPermanent(g, me.ID, "Morgue Toad", "Creature — Frog", b644Toad, false)
	effectstest.TapForMana(t, g, me.ID, toad)
	if g.Battlefield.Contains(toad) {
		t.Error("the Toad was not sacrificed")
	}
	pool := effectstest.SortedPool(me)
	if len(pool) != 2 || pool[0] != "R" || pool[1] != "U" {
		t.Errorf("pool %v, want [R U]", pool)
	}
}
