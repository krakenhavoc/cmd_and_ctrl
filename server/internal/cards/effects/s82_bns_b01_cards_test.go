package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b01_cards_test.go — the Enchanting Tales bonus-sheet cards, batch 1
// (#5929). Every test plays the real catalog entry through the engine.

const (
	bnsAsForetoldOracle       = "e4b3e0af-5caa-47c1-b5be-c56a875cbb15"
	bnsCompulsionOracle       = "3fafb6b2-5cae-45b6-8550-3ff8daa02802"
	bnsCopyEnchantmentOracle  = "87bc0b1f-49ef-4c06-9049-2c148b467882"
	bnsDarkTutelageOracle     = "0c6ce7c9-316b-4171-a8fa-ca26af1eecb7"
	bnsGreaterAuramancyOracle = "e6e888ea-9935-4df2-a344-c54637e55f43"
	bnsHatchingPlansOracle    = "cebc9fcc-25ec-4366-a914-b95afbd134b0"
	bnsKarmicJusticeOracle    = "9ea58595-fcb9-455c-9558-b6a3fb2b83c5"
	bnsLeylineAbundanceOracle = "1197a595-3afc-4f18-b868-bbb096771922"
	bnsLeylineLightningOracle = "bb87ec08-bb1b-43d8-8809-92904cbba685"
	bnsPrismaticOmenOracle    = "d90882ae-6572-4214-8d9c-0bdfb75f9be2"
	bnsSneakAttackOracle      = "e7aee670-e6ac-4b3a-b3d1-7b7013db8f3d"
	bnsDragonMantleOracle     = "0cc5aaa1-07bf-4872-ac72-303cef9fb143"
	bnsFrayingSanityOracle    = "ab4e9284-7904-413e-97b0-69793be79cc5"
	bnsKnightlyValorOracle    = "9ba6cec6-97bf-4a83-9ffa-77e187d0945b"
	bnsStabWoundOracle        = "aad212b2-57ba-4e70-869b-636842eb7a88"
	bnsDeadWeightOracle       = "b1804304-fac1-4b19-a48d-6ade9407972a"
)

// bnsAttachTo marks the Aura `aura` as attached to `host` the way a resolved
// Aura spell leaves it.
func bnsAttachTo(g *game.Game, aura, host uuid.UUID) {
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
}

// --- Sneak Attack ------------------------------------------------------------

func TestBnsSneakAttackPutsAnyCreatureInWithHasteAndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	attack := pushCatalogPermanent(g, me.ID, "Sneak Attack", "Enchantment", bnsSneakAttackOracle, false)
	big := rfdHandCard(me, "Big", "Creature — Giant", "", "", 9, 9)
	floatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, attack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Fatal("the creature did not enter")
	}
	if !hasEffectiveKeyword(t, g, big, "haste") {
		t.Error("the creature did not gain haste")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the creature was not sacrificed at the next end step")
	}
}

func TestBnsSneakAttackDecliningMakesNoDelayedTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	attack := pushCatalogPermanent(g, me.ID, "Sneak Attack", "Enchantment", bnsSneakAttackOracle, false)
	small := rfdHandCard(me, "Small", "Creature — Goblin", "", "", 1, 1)
	floatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, attack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !me.Hand.Contains(small) {
		t.Error("declining the put still moved the creature")
	}
}

// --- Dark Tutelage -----------------------------------------------------------

func TestBnsDarkTutelageRevealsTheTopCardAndChargesItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	flipped := seedLibraryTop(owner, "Gray Merchant of Asphodel", "Creature — Zombie", "{2}{B}{B}")
	pushPermanentForTest(g, owner.ID, "Dark Tutelage", bnsDarkTutelageOracle, "Enchantment")
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	life := owner.Life
	_ = life
	if !owner.Hand.Contains(flipped) {
		t.Fatal("the revealed card never reached hand")
	}
	if owner.Life != 40-4 {
		t.Errorf("life = %d, want 36 (the flipped card's mana value is 4)", owner.Life)
	}
}

func TestBnsDarkTutelageDoesNotTriggerOnAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[2]
	flipped := seedLibraryTop(owner, "Gray Merchant of Asphodel", "Creature — Zombie", "{2}{B}{B}")
	pushPermanentForTest(g, owner.ID, "Dark Tutelage", bnsDarkTutelageOracle, "Enchantment")
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if owner.Hand.Contains(flipped) {
		t.Error("Dark Tutelage triggered in an opponent's upkeep")
	}
}

// --- Stab Wound --------------------------------------------------------------

func TestBnsStabWoundShrinksAndDrainsTheHostsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	aura := pushCatalogPermanent(g, me.ID, "Stab Wound", "Enchantment — Aura", bnsStabWoundOracle, false)
	bnsAttachTo(g, aura, host)
	if p, tt := ptOf(t, g, host); p != 2 || tt != 2 {
		t.Fatalf("the enchanted creature is %d/%d, want 2/2", p, tt)
	}
	myLife, oppLife := me.Life, opp.Life
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the host's controller is at %d, want %d", opp.Life, oppLife-2)
	}
	// Another player's upkeep does not trigger it.
	advanceToStepOf(t, g, 2, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the host's controller is at %d after another player's upkeep, want %d", opp.Life, oppLife-2)
	}
	if me.Life != myLife {
		t.Errorf("the Aura's controller lost life: %d", me.Life)
	}
}

func TestBnsStabWoundKillsAToughnessTwoCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	host := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Stab Wound", "Enchantment — Aura", bnsStabWoundOracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("a 2/2 survived -2/-2")
	}
}

// --- Knightly Valor ----------------------------------------------------------

func TestBnsKnightlyValorPumpsGrantsVigilanceAndMakesAKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Knightly Valor", "Enchantment — Aura", bnsKnightlyValorOracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, host, "vigilance") {
		t.Error("the enchanted creature has no vigilance")
	}
	var knight uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && c.Controller == me.ID {
			knight = c.InstanceID
		}
	}
	if knight == uuid.Nil {
		t.Fatal("no Knight token")
	}
	if p, tt := ptOf(t, g, knight); p != 2 || tt != 2 || !hasEffectiveKeyword(t, g, knight, "vigilance") {
		t.Errorf("the Knight is %d/%d, want a 2/2 with vigilance", p, tt)
	}
}

// --- Dragon Mantle -----------------------------------------------------------

func TestBnsDragonMantleDrawsAndGrantsFirebreathing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()
	aura := castCatalogSpell(t, g, "Dragon Mantle", "Enchantment — Aura", bnsDragonMantleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	// The Aura left the hand for the stack, then a card was drawn.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (the Aura came and left, one card drawn)", got, handBefore+1)
	}
	_ = aura
	idx, ref := grantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted ability")
	}
	floatForTest(g, me, "RR")
	b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, host); p != 4 {
		t.Errorf("power = %d, want 4 after two activations", p)
	}
}

// --- Compulsion --------------------------------------------------------------

func TestBnsCompulsionLootsAndSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	comp := pushCatalogPermanent(g, me.ID, "Compulsion", "Enchantment", bnsCompulsionOracle, false)
	seedLibrary(me, "A", "B", "C")
	discard := rfdHandCard(me, "Discard Me", "Sorcery", "", "", 0, 0)
	floatForTest(g, me, "UUUU")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, comp, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate the discard ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(discard) || me.Hand.Size() != hand {
		t.Errorf("hand = %d (discarded card still there: %v), want %d", me.Hand.Size(), me.Hand.Contains(discard), hand)
	}
	if err := g.ActivateCatalogAbility(me.ID, comp, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the sacrifice ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(comp) {
		t.Error("Compulsion was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Hatching Plans ----------------------------------------------------------

func TestBnsHatchingPlansDrawsThreeWhenItDiesButNotWhenBounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	plans := pushPermanentForTest(g, me.ID, "Hatching Plans", bnsHatchingPlansOracle, "Enchantment")
	hand := me.Hand.Size()
	destroyNow(t, g, plans)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+3)
	}
}

func TestBnsHatchingPlansIsQuietWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	plans := pushPermanentForTest(g, me.ID, "Hatching Plans", bnsHatchingPlansOracle, "Enchantment")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(plans) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (exile is not a graveyard)", me.Hand.Size(), hand)
	}
}

// --- Leyline of Abundance ----------------------------------------------------

func TestBnsLeylineOfAbundanceAddsGreenForCreatureManaAndPutsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Leyline of Abundance", "Enchantment", bnsLeylineAbundanceOracle, false)
	elf := b21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", actLlanowarElvesOracle, 1, 1, "G")
	forest := avB14Forest(g, me.ID)
	g.WithWriteLock(func() { me.ManaPool = nil })
	tapForMana(t, g, me.ID, elf)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("a creature tapped for mana: pool %v, want {G}{G}", got)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 {
		t.Errorf("a land that is not a creature: pool %v, want one {G}", got)
	}
}

func TestBnsLeylineOfAbundanceActivationPutsACounterOnEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leyline := pushCatalogPermanent(g, me.ID, "Leyline of Abundance", "Enchantment", bnsLeylineAbundanceOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatForTest(g, me, "GGGGGGGG")
	if err := g.ActivateCatalogAbility(me.ID, leyline, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", n)
		}
	}
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

// --- Leyline of Lightning ----------------------------------------------------

func TestBnsLeylineOfLightningPingsWhenYouPayOneOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushCatalogPermanent(g, me.ID, "Leyline of Lightning", "Enchantment", bnsLeylineLightningOracle, false)
	life := opp.Life
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	floatForTest(g, me, "R")
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no {1} payment was offered")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, the {1} was not paid", me.ManaPool.Total())
	}
}

func TestBnsLeylineOfLightningDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushCatalogPermanent(g, me.ID, "Leyline of Lightning", "Enchantment", bnsLeylineLightningOracle, false)
	life := opp.Life
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	floatForTest(g, me, "R")
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent at %d, want %d", opp.Life, life)
	}
}

// --- Prismatic Omen ----------------------------------------------------------

func TestBnsPrismaticOmenMakesYourLandsEveryBasicType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Prismatic Omen", bnsPrismaticOmenOracle, "Enchantment")
	mine := avB14Forest(g, me.ID)
	theirs := avB14Forest(g, opp.ID)
	c := layeredCard(t, g, mine)
	for _, st := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if !c.HasSubtype(st) {
			t.Errorf("my land lacks %s", st)
		}
	}
	if layeredCard(t, g, theirs).HasSubtype("Island") {
		t.Error("an opponent's land gained a type")
	}
}

// --- Greater Auramancy -------------------------------------------------------

func TestBnsGreaterAuramancyGivesShroudToOtherEnchantmentsAndEnchantedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auramancy := pushPermanentForTest(g, me.ID, "Greater Auramancy", bnsGreaterAuramancyOracle, "Enchantment")
	other := pushPermanentForTest(g, me.ID, "Other Enchantment", "", "Enchantment")
	theirEnch := pushPermanentForTest(g, opp.ID, "Their Enchantment", "", "Enchantment")
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	enchanted := b12Creature(g, me.ID, "Enchanted", "Creature — Bear", 2, 2)
	theirCreature := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	aura := pushPermanentForTest(g, opp.ID, "Test Aura", "", "Enchantment — Aura")
	bnsAttachTo(g, aura, enchanted)
	aura2 := pushPermanentForTest(g, opp.ID, "Test Aura", "", "Enchantment — Aura")
	bnsAttachTo(g, aura2, theirCreature)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"Greater Auramancy itself", auramancy, false},
		{"another enchantment I control", other, true},
		{"an opponent's enchantment", theirEnch, false},
		{"a creature with no Aura", bare, false},
		{"my enchanted creature", enchanted, true},
		{"an opponent's enchanted creature", theirCreature, false},
	} {
		if got := hasEffectiveKeyword(t, g, tc.id, "shroud"); got != tc.want {
			t.Errorf("%s: shroud = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// --- Copy Enchantment --------------------------------------------------------

func TestBnsCopyEnchantmentEntersAsACopyOfAnyEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushPermanentForTest(g, opp.ID, "Hatching Plans", bnsHatchingPlansOracle, "Enchantment")
	bear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := castCatalogSpell(t, g, "Copy Enchantment", "Enchantment", bnsCopyEnchantmentOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !hasID(pc.CopyOptions, theirs) || hasID(pc.CopyOptions, bear) {
		t.Errorf("CopyOptions = %v, want the enchantment and not the creature", pc.CopyOptions)
	}
	resolveWithCopyChoice(t, g, theirs)
	c := layeredCard(t, g, id)
	if c.Name != "Hatching Plans" || !c.IsEnchantment() {
		t.Errorf("the copy is %q (%s), want a Hatching Plans enchantment", c.Name, c.TypeLine)
	}
}

// --- As Foretold -------------------------------------------------------------

func TestBnsAsForetoldGrowsAndOffersAFreeCastUpToItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToMainForTest(t, g)
	foretold := pushPermanentForTest(g, me.ID, "As Foretold", bnsAsForetoldOracle, "Enchantment")
	key := GrantedAltCostZeroUpToTimeCounters
	spell := handSorcery(me, "Two Drop", "{1}{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), key) != nil {
		t.Fatal("As Foretold offered a free cast with no time counters")
	}
	// Its own upkeep, next round: a counter, then a two-mana spell is free.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, foretold, game.CounterTime); n != 1 {
		t.Fatalf("time counters = %d, want 1", n)
	}
	advanceToMainForTest(t, g)
	if offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), key) != nil {
		t.Error("one time counter offered a free cast of a mana value 2 spell")
	}
	one := handSorcery(me, "One Drop", "{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, one, game.ZoneHand), key) == nil {
		t.Fatal("one time counter did not offer a free cast of a mana value 1 spell")
	}
	if err := g.CastSpell(me.ID, one, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Once each turn.
	another := handSorcery(me, "Another One Drop", "{R}")
	if offerWithKey(grantedCardOffers(g, me.ID, another, game.ZoneHand), key) != nil {
		t.Error("As Foretold offered a second free cast in one turn")
	}
}

// --- Fraying Sanity ----------------------------------------------------------

func TestBnsFrayingSanityMillsAsManyCardsAsWentToTheirGraveyardThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := pushPermanentForTest(g, me.ID, "Fraying Sanity", bnsFrayingSanityOracle, "Enchantment — Aura Curse")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	})
	seedLibrary(opp, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8")
	advanceToMainOf(t, g, 1)
	// Two cards reach opp's graveyard this turn: a discard and a dying creature.
	discard := rfdHandCard(opp, "Discarded", "Sorcery", "", "", 0, 0)
	g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(opp.ID, []uuid.UUID{discard}, uuid.Nil, game.KeyedTail{}) })
	dying := b12Creature(g, opp.ID, "Dying", "Creature — Bear", 1, 1)
	destroyNow(t, g, dying)
	// A token is not a card.
	token := b12Creature(g, opp.ID, "Token", "Token Creature — Bear", 1, 1)
	destroyNow(t, g, token)
	before := opp.Library.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d cards, want 2 (the discard and the creature, not the token)", got)
	}
}

func TestBnsFrayingSanityMillsNothingWhenNothingWentToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := pushPermanentForTest(g, me.ID, "Fraying Sanity", bnsFrayingSanityOracle, "Enchantment — Aura Curse")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	})
	seedLibrary(opp, "L1", "L2", "L3")
	advanceToMainOf(t, g, 1)
	before := opp.Library.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 0 {
		t.Errorf("milled %d cards, want 0", got)
	}
}

// --- Karmic Justice ----------------------------------------------------------

func bnsKarmicTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	opp = g.Seats[g.Turn.ActiveSeat]
	me = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushPermanentForTest(g, me.ID, "Karmic Justice", bnsKarmicJusticeOracle, "Enchantment")
	return g, me, opp
}

func TestBnsKarmicJusticeDestroysAPermanentOfTheOpponentWhoDestroyedYours(t *testing.T) {
	g, me, opp := bnsKarmicTable(t)
	rock := pushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	bystander := pushTestPermanent(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, game.Card{Name: "Third Rock", TypeLine: "Artifact"})
	castCatalogSpell(t, g, "Shatter", "Instant", leaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	pick := game.TargetRef{Kind: game.TargetCard, ID: theirs}
	assistCardSettle(t, g, me.ID, &pick, true)
	if g.Battlefield.Contains(rock) {
		t.Fatal("setup: the artifact was not destroyed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's permanent was not destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a third player's permanent was destroyed")
	}
}

func TestBnsKarmicJusticeIgnoresMyOwnDestroyAndCreatures(t *testing.T) {
	g, me, opp := bnsKarmicTable(t)
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	golem := pushTestPermanent(g, me.ID, game.Card{Name: "My Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2})
	// The opponent destroys an artifact CREATURE of mine: not a noncreature permanent.
	castCatalogSpell(t, g, "Shatter", "Instant", leaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: golem}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Fatal("setup: the artifact creature was not destroyed")
	}
	if latestPickTarget(g, me.ID) != nil || assistCardTriggerPrompt(g, me.ID) != nil {
		t.Error("Karmic Justice triggered for a creature")
	}
	// A state-based or bare destruction nobody's spell caused is not an opponent's.
	mine := pushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	destroyNow(t, g, mine)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || assistCardTriggerPrompt(g, me.ID) != nil {
		t.Error("Karmic Justice triggered with no spell or ability behind the destruction")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's permanent was destroyed")
	}
}

func TestBnsKarmicJusticeIgnoresMyOwnSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushPermanentForTest(g, me.ID, "Karmic Justice", bnsKarmicJusticeOracle, "Enchantment")
	rock := pushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	castCatalogSpell(t, g, "Shatter", "Instant", leaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || assistCardTriggerPrompt(g, me.ID) != nil {
		t.Error("Karmic Justice triggered for my own spell")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's permanent was destroyed")
	}
}

func TestBnsKarmicJusticeMayDecline(t *testing.T) {
	g, me, opp := bnsKarmicTable(t)
	rock := pushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	castCatalogSpell(t, g, "Shatter", "Instant", leaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	pick := game.TargetRef{Kind: game.TargetCard, ID: theirs}
	assistCardSettle(t, g, me.ID, &pick, false)
	if !g.Battlefield.Contains(theirs) {
		t.Error("declining the may still destroyed the permanent")
	}
}

func TestBnsKarmicJusticeCannotTargetAThirdPlayersPermanent(t *testing.T) {
	g, me, opp := bnsKarmicTable(t)
	rock := pushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	third := pushTestPermanent(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID, game.Card{Name: "Third Rock", TypeLine: "Artifact"})
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	castCatalogSpell(t, g, "Shatter", "Instant", leaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: third}); err == nil {
		t.Error("a permanent the destroying opponent does not control was accepted")
	}
}

// --- Copy Enchantment copying an Aura ------------------------------------------

func TestBnsCopyEnchantmentCopyingAnAuraEnchantsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	host := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	aura := pushCatalogPermanent(g, opp.ID, "Dead Weight", "Enchantment — Aura", bnsDeadWeightOracle, false)
	bnsAttachTo(g, aura, host)
	id := castCatalogSpell(t, g, "Copy Enchantment", "Enchantment", bnsCopyEnchantmentOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	resolveWithCopyChoice(t, g, aura)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		answerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: host})
		passPriorityAroundTable(t, g)
	}
	copyCard := layeredCard(t, g, id)
	if copyCard.Name != "Dead Weight" || !copyCard.IsAura() {
		t.Fatalf("the copy is %q (%s), want a Dead Weight Aura", copyCard.Name, copyCard.TypeLine)
	}
	if copyCard.AttachedTo.ID != host {
		t.Errorf("the copied Aura is attached to %v, want the creature %v", copyCard.AttachedTo.ID, host)
	}
	if p, tt := ptOf(t, g, host); p != 2 || tt != 2 {
		t.Errorf("the creature is %d/%d, want 2/2 under two Dead Weights", p, tt)
	}
}
