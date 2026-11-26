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
	oldB239Diviner  = "4bf308fc-a54e-4fba-b0fd-3bedb4db3b18"
	oldB239Samut    = "fffdc2ac-bde4-4e4c-a5bd-0e6c6e49ad91"
	oldB239Nissa    = "ac905661-37c4-495a-ac76-96032c23a144"
	oldB239Thirst   = "5f55d41a-ef93-4d4f-ab55-13ea887e7af2"
	oldB239Wildfire = "195736a9-d8b0-4c74-abe9-0289aa83c67d"
	oldB239Splendor = "1099588c-cc26-4523-8bd2-90e1e1433f53"
	oldB239Hold     = "a3639381-e401-409e-8f07-0a3bc736c50d"
	oldB239Bolas    = "524b00ea-0372-40f9-a6a1-76a8497a1ea6"
	oldB239Camel    = "2eac0c6a-b289-4b75-9571-59869e62d180"
	oldB239Mummy    = "ea30f395-86fc-444f-8c9e-b9876287084e"
)

func TestOldB239CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB239Diviner: "Dune Diviner", oldB239Samut: "Samut, the Tested", oldB239Nissa: "Nissa, Genesis Mage",
		oldB239Thirst: "Unquenchable Thirst", oldB239Wildfire: "Wildfire Eternal", oldB239Splendor: "Overwhelming Splendor",
		oldB239Hold: "Desert's Hold", oldB239Bolas: "Nicol Bolas, the Deceiver", oldB239Camel: "Solitary Camel",
		oldB239Mummy: "Unraveling Mummy",
	}
	for oracle, name := range want {
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

func oldB239Desert(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Test Desert", "Land — Desert", "", 0, 0)
}

func oldB239Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, OracleID: oracle,
		TypeLine: "Legendary Planeswalker — Test", Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner})
}

func TestOldB239DuneDivinerTapsADesertToGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	diviner := effectstest.B12Push(g, me.ID, "Dune Diviner", "Creature — Snake Cleric", oldB239Diviner, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, diviner, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no Desert to tap")
	}
	desert := oldB239Desert(g, me.ID)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, diviner, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{desert}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if !effectstest.IsTapped(g, desert) {
		t.Error("the Desert was not tapped")
	}
}

func TestOldB239SolitaryCamelHasLifelinkOnlyWithADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	camel := effectstest.B12Push(g, me.ID, "Solitary Camel", "Creature — Camel", oldB239Camel, 3, 2)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, camel, "lifelink") {
		t.Fatal("lifelink with no Desert")
	}
	desert := oldB239Desert(g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, camel, "lifelink") {
		t.Error("no lifelink with a Desert on the battlefield")
	}
	g.WithWriteLock(func() { g.Battlefield.Remove(desert) })
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Test Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
	})
	if !effectstest.HasEffectiveKeyword(t, g, camel, "lifelink") {
		t.Error("no lifelink with a Desert card in the graveyard")
	}
}

func TestOldB239UnravelingMummyGrantsOnlyToAnAttackingZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mummy := effectstest.B12Push(g, me.ID, "Unraveling Mummy", "Creature — Zombie", oldB239Mummy, 2, 3)
	zombie := effectstest.B12Creature(g, me.ID, "Test Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B64ReadyToAttack(g, zombie)
	effectstest.B64ReadyToAttack(g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, mummy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(zombie)}); err == nil {
		t.Fatal("targeted a Zombie that is not attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{zombie, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("attack: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.ApaMana(me, "C", "W", "B")
	if err := g.ActivateCatalogAbility(me.ID, mummy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("targeted an attacking non-Zombie")
	}
	if err := g.ActivateCatalogAbility(me.ID, mummy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(zombie)}); err != nil {
		t.Fatalf("lifelink: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ApaMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, mummy, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(zombie)}); err != nil {
		t.Fatalf("deathtouch: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, zombie, kw) {
			t.Errorf("the attacking Zombie has no %s", kw)
		}
	}
}

func TestOldB239WildfireEternalOffersAFreeInstantOrSorceryWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	we := effectstest.B12Push(g, me.ID, "Wildfire Eternal", "Creature — Zombie Jackal Cleric", oldB239Wildfire, 1, 4)
	effectstest.B64ReadyToAttack(g, we)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := effectstest.HandCardFull(me, "Test Sorcery", "Sorcery", "{4}{R}", "", []string{"R"})
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, we, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil {
		t.Fatalf("no free-cast offer after the unblocked attack: %+v", g.PendingChoices)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the spell left the hand before the offer was answered")
	}
}

func TestOldB239OverwhelmingSplendorShrinksCreaturesAndLimitsActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Overwhelming Splendor", TypeLine: "Enchantment — Aura Curse", OracleID: oldB239Splendor, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the enchanted player's creature is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your own creature is %d/%d, want 3/3", p, tough)
	}
	// An activated ability of the enchanted player is refused; yours isn't.
	woodland := effectstest.PushCatalogPermanent(g, opp.ID, "Blighted Woodland", "Land", "02679a2e-303d-412f-87d8-0a37a8ca259c", false)
	effectstest.ApaMana(opp, "C", "C", "C", "G")
	err := g.ActivateCatalogAbility(opp.ID, woodland, 0, game.ActivateAbilityParams{Strict: true})
	var refused *game.CantActivateError
	if !errors.As(err, &refused) || refused.Source != id {
		t.Errorf("the enchanted player's activation was not refused by the Aura: %v", err)
	}
}

func TestOldB239UnquenchableThirstTapsOnlyWithADesertAndKeepsItTapped(t *testing.T) {
	for _, desert := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
		if desert {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Test Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
		}
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Unquenchable Thirst", TypeLine: "Enchantment — Aura", OracleID: oldB239Thirst, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(ogre)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.IsTapped(g, ogre); got != desert {
			t.Errorf("desert=%v: tapped = %v", desert, got)
		}
	}
}

func TestOldB239DesertsHoldGainsLifeWithADesertAndPinsTheCreature(t *testing.T) {
	for _, desert := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
		if desert {
			oldB239Desert(g, me.ID)
		}
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Desert's Hold", TypeLine: "Enchantment — Aura", OracleID: oldB239Hold, Owner: me.ID, Controller: me.ID})
		life := me.Life
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(ogre)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if desert {
			want = life + 3
		}
		if me.Life != want {
			t.Errorf("desert=%v: life %d, want %d", desert, me.Life, want)
		}
		if !effectstest.SbCantBlock(t, g, ogre) {
			t.Errorf("desert=%v: the enchanted creature can block", desert)
		}
	}
}

func TestOldB239NicolBolasPlusThreePunishesEachOpponentAndMinusThreeDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolas := oldB239Walker(g, me.ID, "Nicol Bolas, the Deceiver", oldB239Bolas, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			lives[p.ID] = p.Life
		}
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		effectstest.BlbSTPickOption(t, g, p.ID, "Lose 3 life")
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != lives[p.ID]-3 {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, lives[p.ID]-3)
		}
	}
	if n := effectstest.CounterCount(g, bolas, game.CounterLoyalty); n != 8 {
		t.Errorf("loyalty %d, want 8", n)
	}
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, bolas, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		// Only one loyalty ability per turn: a second activation is refused.
		t.Fatal("a second loyalty ability was activated in one turn")
	}
	_ = hand
}

func TestOldB239NicolBolasMinusThreeAndMinusElevenResolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolas := oldB239Walker(g, me.ID, "Nicol Bolas, the Deceiver", oldB239Bolas, 12)
	victim := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Victim", "Creature — Bear", 2, 2)
	for i := 0; i < 10; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bolas, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived −3")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}

	// Next turn's worth: a fresh game for the ultimate.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	bolas = oldB239Walker(g, me.ID, "Nicol Bolas, the Deceiver", oldB239Bolas, 12)
	for i := 0; i < 10; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	hand = me.Hand.Size()
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, bolas, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−11: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 7
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
	if me.Hand.Size() != hand+7 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+7)
	}
}

func TestOldB239SamutPlusOneMinusTwoAndMinusSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	samut := oldB239Walker(g, me.ID, "Samut, the Tested", oldB239Samut, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, samut, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("no double strike")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	samut = oldB239Walker(g, me.ID, "Samut, the Tested", oldB239Samut, 4)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, samut, 1, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, effectstest.PlayerTarget(opp.ID)},
		Distribution: map[uuid.UUID]int{a: 1, opp.ID: 1}}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if n := effectstest.B12Card(t, g, a).DamageMarked; n != 1 {
		t.Errorf("damage on the creature %d, want 1", n)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	samut = oldB239Walker(g, me.ID, "Samut, the Tested", oldB239Samut, 7)
	var found []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		found = append(found, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Library Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, samut, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, found...)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Library Bear" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d searched creatures on the battlefield, want 2", n)
	}
}

func TestOldB239NissaUntapsPumpsAndDigsTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := oldB239Walker(g, me.ID, "Nissa, Genesis Mage", oldB239Nissa, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == bear || id == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear, land)}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) || effectstest.IsTapped(g, land) {
		t.Error("a target stayed tapped")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	nissa = oldB239Walker(g, me.ID, "Nissa, Genesis Mage", oldB239Nissa, 5)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("%d/%d, want 7/7", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	nissa = oldB239Walker(g, me.ID, "Nissa, Genesis Mage", oldB239Nissa, 10)
	var creature, spell uuid.UUID
	for i := 9; i >= 0; i-- {
		id := uuid.New()
		c := game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		switch i {
		case 0:
			c = game.Card{InstanceID: id, Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
			creature = id
		case 1:
			spell = id
		}
		me.Library.PushTop(c)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the creature card did not enter the battlefield")
	}
	if g.Battlefield.Contains(spell) {
		t.Error("a nonpermanent card entered the battlefield")
	}
}
