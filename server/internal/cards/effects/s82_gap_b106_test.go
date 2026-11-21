package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b106_test.go — S82 GAP batch 106 (#5948). Each test plays the real
// catalog entry.

const (
	gapB106SpitebellowsOracle = "fee18f57-4ae5-4df8-af3b-f82ed2157ac3"
	gapB106SpitefulOracle     = "c4d2c804-1448-4639-b41b-0a86bac083b9"
	gapB106SubiraOracle       = "9e2e3272-d4e1-4329-b1d2-f393b3511b82"
	gapB106MelekOracle        = "095d9719-0db6-43de-8e4f-a0035a4c65ed"
	gapB106AkiriOracle        = "c03ae426-2536-4c20-a73a-a24e816f76b0"
	gapB106LavabellyOracle    = "334f673e-7cfa-432c-93aa-41adfbb99113"
	gapB106AzamiOracle        = "0f8b97fe-3e5e-47c2-9a9d-7f77482aa159"
)

func TestGapB106CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB106SpitebellowsOracle: "Spitebellows",
		gapB106SpitefulOracle:     "Spiteful Sliver",
		gapB106SubiraOracle:       "Subira, Tulzidi Caravanner",
		gapB106MelekOracle:        "Melek, Izzet Paragon",
		gapB106AkiriOracle:        "Akiri, Fearless Voyager",
		gapB106LavabellyOracle:    "Lavabelly Sliver",
		gapB106AzamiOracle:        "Azami, Lady of Scrolls",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Spitebellows ------------------------------------------------------------

func TestGapB106SpitebellowsEvokedDealsSixWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Big Victim", "Creature — Ogre", 6, 6)
	id := putInHand(me, game.Card{Name: "Spitebellows", TypeLine: "Creature — Elemental", OracleID: gapB106SpitebellowsOracle,
		ManaCost: "{5}{R}", Power: 6, Toughness: 1})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("CastSpell evoke: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 6/6 survived an evoked Spitebellows' leave trigger")
	}
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the evoked Spitebellows was not sacrificed")
	}
}

func TestGapB106SpitebellowsDamagesWhenItDiesAsAnOrdinaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bellows := b12Push(g, me.ID, "Spitebellows", "Creature — Elemental", gapB106SpitebellowsOracle, 6, 1)
	victim := b12Creature(g, opp.ID, "Big Victim", "Creature — Ogre", 6, 6)
	small := b12Creature(g, opp.ID, "Small Victim", "Creature — Bear", 2, 2)
	b27Kill(g, bellows)
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the chosen creature survived six damage")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the untargeted creature died")
	}
}

// --- Spiteful Sliver ---------------------------------------------------------

func TestGapB106SpitefulSliverReflectsDamageDealtToASliverOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Spiteful Sliver", "Creature — Sliver", gapB106SpitefulOracle, 2, 2)
	other := b12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 3, 5)
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	life := opp.Life
	effectivePTOf(t, g, other) // settle the layers so the granted trigger exists
	cmdr2B08Damage(g, src, other, 3)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d, want 3 (that much)", got)
	}
}

func TestGapB106SpitefulSliverDoesNotCoverNonSlivers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Spiteful Sliver", "Creature — Sliver", gapB106SpitefulOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	life := opp.Life
	effectivePTOf(t, g, bear)
	cmdr2B08Damage(g, src, bear, 3)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a damaged Bear reflected damage: life %d -> %d", life, opp.Life)
	}
}

// --- Lavabelly Sliver --------------------------------------------------------

func TestGapB106LavabellySliverPingsAndGainsWhenASliverEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lavabelly Sliver", "Creature — Sliver", gapB106LavabellyOracle, 2, 2)
	toMain(t, g)
	my, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Plain Sliver", "Creature — Sliver", "", nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-1 || me.Life != my+1 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d, want -1 and +1", theirs, opp.Life, my, me.Life)
	}
}

func TestGapB106LavabellySliverTriggersOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	my, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Lavabelly Sliver", "Creature — Sliver", gapB106LavabellyOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-1 || me.Life != my+1 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d, want -1 and +1", theirs, opp.Life, my, me.Life)
	}
}

func TestGapB106LavabellySliverIgnoresANonSliverEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lavabelly Sliver", "Creature — Sliver", gapB106LavabellyOracle, 2, 2)
	toMain(t, g)
	theirs := opp.Life
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs {
		t.Errorf("a Bear entering cost the opponent %d life", theirs-opp.Life)
	}
}

// --- Azami, Lady of Scrolls --------------------------------------------------

func TestGapB106AzamiTapsAWizardToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	azami := b12Push(g, me.ID, "Azami, Lady of Scrolls", "Legendary Creature — Human Wizard", gapB106AzamiOracle, 0, 2)
	wiz := b12Creature(g, me.ID, "Apprentice", "Creature — Human Wizard", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	if err := ravB2Activate(t, g, me.ID, azami, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a non-Wizard paid the tap cost")
	}
	if err := ravB2Activate(t, g, me.ID, azami, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{wiz}}); err != nil {
		t.Fatalf("tapping a Wizard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Subira, Tulzidi Caravanner ----------------------------------------------

func TestGapB106SubiraMakesAnotherSmallCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	subira := b12Push(g, me.ID, "Subira, Tulzidi Caravanner", "Legendary Creature — Human Shaman", gapB106SubiraOracle, 2, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	advanceToMain(t, g)
	if err := ravB2Activate(t, g, me.ID, subira, 0, game.ActivateAbilityParams{Targets: cardRefs(big)}, "C"); err == nil {
		t.Error("a power-3 creature was a legal target")
	}
	if err := ravB2Activate(t, g, me.ID, subira, 0, game.ActivateAbilityParams{Targets: cardRefs(subira)}, "C"); err == nil {
		t.Error("Subira targeted herself")
	}
	ravB2MustActivate(t, g, me.ID, subira, 0, game.ActivateAbilityParams{Targets: cardRefs(small)}, "C")
	assertRestrictions(t, g, small, game.CantBeBlocked)
}

func TestGapB106SubiraDiscardsHandAndDrawsForSmallCreaturesThatConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	subira := b12Push(g, me.ID, "Subira, Tulzidi Caravanner", "Legendary Creature — Human Shaman", gapB106SubiraOracle, 2, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	advanceToMain(t, g)
	putInHand(me, game.Card{Name: "Spare One", TypeLine: "Sorcery"})
	putInHand(me, game.Card{Name: "Spare Two", TypeLine: "Sorcery"})
	seedLibrary(me, "A", "B", "C", "D")
	if err := ravB2Activate(t, g, me.ID, subira, 1, game.ActivateAbilityParams{}, "C"); err == nil {
		t.Fatal("paid {1}{R} with only colourless mana")
	}
	ravB2MustActivate(t, g, me.ID, subira, 1, game.ActivateAbilityParams{}, "C", "R")
	if me.Hand.Size() != 0 {
		t.Fatalf("hand has %d cards after discarding it", me.Hand.Size())
	}
	attackWith(t, g, opp.ID, small, big)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand is %d cards, want 1 (only the power-2 creature's damage draws)", me.Hand.Size())
	}
}

// --- Melek, Izzet Paragon ----------------------------------------------------

func TestGapB106MelekCastsInstantsAndSorceriesFromTheTopAndCopiesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	melek := pushTopSource(t, g, me.ID, "Melek, Izzet Paragon", "Legendary Creature — Weird Wizard", gapB106MelekOracle)
	if !topVisibleTo(g, me.ID, opp.ID) {
		t.Error("the top card is not revealed")
	}
	for _, c := range []game.Card{
		{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"},
	} {
		top := seedTopOfLibrary(me, c)
		if grantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
			t.Errorf("%s was opened by an instant-and-sorcery permission", c.Name)
		}
	}
	spell := seedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if !grantedPermissionOn(g, me.ID, spell, game.ZoneLibrary).Granted() {
		t.Fatal("a sorcery on top was not castable")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the sorcery off the top: %v", err)
	}
	g.RunStateChecksForTest()
	if triggerOnStack(g, melek) == nil {
		t.Fatal("no Melek trigger after casting from the library")
	}
	for i := 0; i < 8 && triggerOnStack(g, melek) != nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(copiesOnStack(g)); n != 1 {
		t.Errorf("%d copies on the stack after the trigger resolved, want 1", n)
	}
}

func TestGapB106MelekDoesNotCopyASpellCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	melek := pushTopSource(t, g, me.ID, "Melek, Izzet Paragon", "Legendary Creature — Weird Wizard", gapB106MelekOracle)
	id := mayhemHandCard(g, me, "From Hand", "Sorcery", "{0}", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if triggerOnStack(g, melek) != nil {
		t.Error("Melek triggered for a spell cast from hand")
	}
}

// --- Akiri, Fearless Voyager -------------------------------------------------

func TestGapB106AkiriDrawsOnceWhenEquippedCreaturesAttackAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Akiri, Fearless Voyager", "Legendary Creature — Kor Warrior", gapB106AkiriOracle, 3, 3)
	a := pushVanillaCreature(g, me.ID, "Armed One", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Armed Two", 2, 2)
	somB1Equip(t, g, me.ID, "Sword One", "", a)
	somB1Equip(t, g, me.ID, "Sword Two", "", b)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, a, b)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 for the one player attacked", got)
	}
}

func TestGapB106AkiriDoesNotDrawForUnequippedAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Akiri, Fearless Voyager", "Legendary Creature — Kor Warrior", gapB106AkiriOracle, 3, 3)
	bare := pushVanillaCreature(g, me.ID, "Bare", 2, 2)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bare)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards for an unequipped attacker", got)
	}
}

func TestGapB106AkiriUnattachesAnEquipmentTapsTheCreatureAndMakesItIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	akiri := b12Push(g, me.ID, "Akiri, Fearless Voyager", "Legendary Creature — Kor Warrior", gapB106AkiriOracle, 3, 3)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	sword := somB1Equip(t, g, me.ID, "Sword", "", host)
	advanceToMain(t, g)
	if err := ravB2Activate(t, g, me.ID, akiri, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without {W}")
	}
	ravB2MustActivate(t, g, me.ID, akiri, 0, game.ActivateAbilityParams{}, "W")
	answerPermanentsPick(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(sword); !ok || c.AttachedTo.Kind == game.TargetCard {
		t.Error("the Equipment is still attached")
	}
	if !hasKeywordOnBattlefield(t, g, host, "indestructible") {
		t.Error("the creature did not gain indestructible")
	}
	if c, _ := g.LookupCardForEffect(host); !c.Tapped {
		t.Error("the creature was not tapped")
	}
}
