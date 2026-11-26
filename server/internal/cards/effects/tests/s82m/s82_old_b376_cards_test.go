package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o376HeroIroas   = "0d44bd43-77da-449e-b771-8a14da88b677"
	o376Ragemonger  = "0fcfc145-9371-40a8-9194-84e5fd700a8d"
	o376Serpent     = "1031bc78-7eca-44aa-9ea8-47fa67b4ce62"
	o376Brawler     = "2ddd33e1-b48e-42e2-a60a-f09f151d3c1d"
	o376Conscriptor = "54216953-c575-4e08-9bee-384721c0bc4d"
	o376Astronomer  = "6593a7c1-fb48-4e53-8e6a-339ad2cc4489"
	o376HeroLeina   = "ae15a96d-ce88-4cb9-a353-439387ca3d02"
	o376Kraken      = "c282581f-3d77-47ce-ac92-ed6217754e38"
	o376Titan       = "f8b1ac76-2f9b-454e-a20a-1cb79c23f414"
	o376Mindreaver  = "f606ab91-1f20-4351-acd3-6366b1895f72"
)

func TestO376CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o376HeroIroas: "Hero of Iroas", o376Ragemonger: "Ragemonger", o376Serpent: "Floodtide Serpent",
		o376Brawler: "Felhide Brawler", o376Conscriptor: "Akroan Conscriptor", o376Astronomer: "Meletis Astronomer",
		o376HeroLeina: "Hero of Leina Tower", o376Kraken: "Kraken of the Straits", o376Titan: "Marshmist Titan",
		o376Mindreaver: "Mindreaver",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// o376Heroic casts Hearts on Fire (+2/+1) at the creature, the spell that triggers heroic.
func o376Heroic(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(target))
}

func o376Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("not on the battlefield")
	}
	return c.Counters[game.CounterPlusOne]
}

func TestO376HeroOfIroasGetsACounterAndDiscountsYourAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	hero := effectstest.B12Push(g, me.ID, "Hero of Iroas", "Creature — Human Soldier", o376HeroIroas, 2, 2)
	o376Heroic(t, g, hero)
	effectstest.PassPriorityAroundTable(t, g)
	if n := o376Counters(t, g, hero); n != 1 {
		t.Fatalf("%d counters, want 1", n)
	}
	aura := effectstest.HandCardFull(me, "Test Aura", effectstest.AuraTypeLine, "{2}{W}", "", []string{"W"})
	effectstest.B06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(hero)}); err != nil {
		t.Fatalf("an Aura costing {2}{W} should cost {1}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := o376Counters(t, g, hero); n != 2 {
		t.Errorf("%d counters after an Aura targeted it, want 2", n)
	}
	// An opponent's own Aura is not discounted.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	_ = opp
	effectstest.ToMainForCost(t, g2)
	effectstest.B12Push(g2, opp2.ID, "Hero of Iroas", "Creature — Human Soldier", o376HeroIroas, 2, 2)
	bear := effectstest.AuraBear(g2, me2.ID)
	aura2 := effectstest.HandCardFull(me2, "Test Aura", effectstest.AuraTypeLine, "{2}{W}", "", []string{"W"})
	effectstest.B06AddMana(me2, "C", "W")
	if err := g2.CastSpell(me2.ID, aura2, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("an opponent's Hero discounted my Aura")
	}
}

func TestO376AkroanConscriptorStealsAnotherCreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	con := effectstest.B12Push(g, me.ID, "Akroan Conscriptor", "Creature — Human Shaman", o376Conscriptor, 3, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, theirs).Tapped = true })
	o376Heroic(t, g, con)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatalf("not stolen: %+v", c)
	}
	if c.Tapped {
		t.Error("the stolen creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the stolen creature has no haste")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCardByID(g, theirs); c.Controller != opp.ID {
		t.Error("control did not end with the turn")
	}
}

func TestO376HeroOfLeinaTowerPaysXForXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	hero := effectstest.B12Push(g, me.ID, "Hero of Leina Tower", "Creature — Human Warrior", o376HeroLeina, 1, 1)
	effectstest.B09AddColorless(me, 3)
	o376Heroic(t, g, hero)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PayXPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 3); err != nil {
		t.Fatalf("ResolvePayAmount: %v", err)
	}
	pay := effectstest.PayUnlessFor(g, me.ID)
	if pay == nil {
		t.Fatal("no payment prompt")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := o376Counters(t, g, hero); n != 3 {
		t.Fatalf("%d counters, want 3", n)
	}
}

func TestO376MeletisAstronomerTakesAnEnchantmentFromTheTopThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	astronomer := effectstest.B12Push(g, me.ID, "Meletis Astronomer", "Creature — Human Wizard", o376Astronomer, 1, 3)
	push := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	push("Deep Card", "Instant")
	ench := push("Ench", "Enchantment")
	push("Bear", "Creature — Bear")
	push("Sorcery", "Sorcery")
	o376Heroic(t, g, astronomer)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ench || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the one enchantment, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ench}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment is not in hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
}

func TestO376MindreaverExilesThreeAndCountersAMatchingName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mr := effectstest.B12Push(g, me.ID, "Mindreaver", "Creature — Human Wizard", o376Mindreaver, 2, 1)
	names := []string{"Alpha", "Beta", "Gamma"}
	for i := len(names) - 1; i >= 0; i-- {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: names[i], TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Extra", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	// Library top-down: Extra, Alpha, Beta, Gamma. The top three are exiled.
	lib := len(opp.Library.Cards)
	o376Heroic(t, g, mr)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != lib-3 {
		t.Fatalf("library %d -> %d, want three exiled", lib, len(opp.Library.Cards))
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.ExiledWith.ID == mr {
			exiled++
		}
	}
	if exiled != 3 {
		t.Fatalf("%d cards exiled with Mindreaver, want 3", exiled)
	}
}

func TestO376MindreaverOnlyTargetsASpellWithAnExiledName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mr := effectstest.B12Push(g, me.ID, "Mindreaver", "Creature — Human Wizard", o376Mindreaver, 2, 1)
	mrCard, _ := effectstest.BattlefieldCardByID(g, mr)
	effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{ID: mr, Epoch: mrCard.ObjectEpoch}, "Hearts on Fire", "Instant", "")
	// The opponent casts a spell named like the exiled card, and one that is not.
	same := effectstest.HandCardFull(opp, "Hearts on Fire", "Instant", "{R}", effectstest.J25b1HeartsOnFireOracle, []string{"R"})
	other := effectstest.HandCardFull(opp, "Shock", "Instant", "{R}", "", []string{"R"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	opp.ManaPool.AddMana(game.ManaToken{Color: "R"})
	opp.ManaPool.AddMana(game.ManaToken{Color: "R"})
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	if err := g.CastSpell(opp.ID, same, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast same: %v", err)
	}
	if err := g.CastSpell(opp.ID, other, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast other: %v", err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateCatalogAbility(me.ID, mr, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("countered a spell whose name is not an exiled card's")
	}
	if err := g.ActivateCatalogAbility(me.ID, mr, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: same}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(mr) {
		t.Error("Mindreaver was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(same) {
		t.Error("the spell was not countered")
	}
}

func TestO376RagemongerTakesColouredManaOffMinotaurSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.B12Push(g, me.ID, "Ragemonger", "Creature — Minotaur Shaman", o376Ragemonger, 2, 3)
	mino := effectstest.HandCardFull(me, "Test Minotaur", "Creature — Minotaur", "{2}{R}", "", []string{"R"})
	effectstest.B06AddMana(me, "C", "C")
	if err := g.CastSpell(me.ID, mino, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Minotaur costing {2}{R} should cost {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A non-Minotaur is not reduced.
	bear := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{2}{R}", "", []string{"R"})
	effectstest.B06AddMana(me, "C", "C")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Minotaur was discounted")
	}
}

func TestO376MarshmistTitanCostsItsDevotionToBlackLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Black A", TypeLine: "Creature — Test", ManaCost: "{B}{B}", Power: 1, Toughness: 1})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Black B", TypeLine: "Creature — Test", ManaCost: "{2}{B}", Power: 1, Toughness: 1})
	titan := effectstest.HandCardFull(me, "Marshmist Titan", "Creature — Giant", "{6}{B}", o376Titan, []string{"B"})
	effectstest.B06AddMana(me, "C", "C", "C", "B") // {6}{B} less devotion 3 = {3}{B}
	if err := g.CastSpell(me.ID, titan, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("with devotion 3 the Titan should cost {3}{B}: %v", err)
	}
}

func o376BrawlerBlock(t *testing.T, withMinotaur bool) error {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawler := effectstest.B12Push(g, opp.ID, "Felhide Brawler", "Creature — Minotaur", o376Brawler, 2, 2)
	if withMinotaur {
		effectstest.B12Creature(g, opp.ID, "Other Minotaur", "Creature — Minotaur", 1, 1)
	}
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.B64ReadyToAttack(g, atk)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, atk, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(brawler, atk)
}

func TestO376FelhideBrawlerCannotBlockWithoutAnotherMinotaur(t *testing.T) {
	if err := o376BrawlerBlock(t, false); err == nil {
		t.Fatal("the Brawler blocked without another Minotaur")
	}
}

func TestO376FelhideBrawlerBlocksWithAnotherMinotaur(t *testing.T) {
	if err := o376BrawlerBlock(t, true); err != nil {
		t.Fatalf("the Brawler could not block with another Minotaur: %v", err)
	}
}

func TestO376KrakenCannotBeBlockedByCreaturesWithPowerBelowItsIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.B12Push(g, me.ID, "Kraken of the Straits", "Creature — Kraken", o376Kraken, 6, 6)
	for i := 0; i < 3; i++ {
		effectstest.B12Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	}
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 2, 2)
	strong := effectstest.B12Creature(g, opp.ID, "Strong", "Creature — Ogre", 3, 3)
	effectstest.B64ReadyToAttack(g, kraken)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, kraken, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(weak, kraken); err == nil {
		t.Fatal("a power-2 creature blocked with three Islands out")
	}
	if err := g.DeclareBlocker(strong, kraken); err != nil {
		t.Fatalf("a power-3 creature could not block: %v", err)
	}
}

func TestO376FloodtideSerpentAttacksOnlyByReturningAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.B12Push(g, me.ID, "Floodtide Serpent", "Creature — Serpent", o376Serpent, 4, 4)
	ench := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.B64ReadyToAttack(g, serpent)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttackerWith(serpent, opp.ID, game.DeclareAttackersParams{}); err == nil {
		t.Fatal("attacked without returning an enchantment")
	}
	if err := g.DeclareAttackerWith(serpent, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{ReturnIDs: []uuid.UUID{rock}}}); err == nil {
		t.Fatal("returned a non-enchantment to attack")
	}
	if err := g.DeclareAttackerWith(serpent, opp.ID, game.DeclareAttackersParams{Pay: game.CombatCostPayment{ReturnIDs: []uuid.UUID{ench}}}); err != nil {
		t.Fatalf("attack with an enchantment returned: %v", err)
	}
	if g.Battlefield.Contains(ench) || !me.Hand.Contains(ench) {
		t.Error("the enchantment was not returned to its owner's hand")
	}
}
