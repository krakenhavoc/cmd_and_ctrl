package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_vz_cards_test.go — Wilds of Eldraine, the V–Z batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeVZOutriderOracle  = "8fc049e7-a16d-4b48-b930-5c346f6718b9"
	woeVZCourageOracle   = virtueOfCourageOracleID
	woeVZLoyaltyOracle   = virtueOfLoyaltyOracleID
	woeVZStrengthOracle  = virtueOfStrengthOracleID
	woeVZVerminOracle    = "8d36d80e-9934-4b56-bf3c-2b18a6da3d2d"
	woeVZTabbyOracle     = "a7736614-cdd3-43bd-ab3e-461e3e24ed40"
	woeVZWingsOracle     = "17f8e8cb-cd1b-4d94-9efa-ce268e8e7096"
	woeVZSweettoothOrcl  = "9760c396-fc7d-41bc-8f03-5ee359eaec00"
	woeVZWerefoxOracle   = "d5ee2ced-29f4-430f-962e-2f930b92624c"
	woeVZVisitorOracle   = "10330a79-a416-40da-bf5a-8b7cbd9a0498"
	woeVZMentorOracle    = "f4e61b03-e6c8-4751-9720-b91fdecd57da"
	woeVZFrenzyOracle    = "f46befcb-6a7f-4f5e-8634-393cbf4b3308"
	woeVZAcolyteOracle   = woodlandAcolyteOracleID
	woeVZBearType        = "Creature — Bear"
	woeVZEnchantmentType = "Enchantment"
)

func TestWoeVZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeVZOutriderOracle:        "Verdant Outrider",
		woeVZCourageOracle:         "Virtue of Courage",
		woeVZCourageOracle + "#1":  "Embereth Blaze",
		woeVZLoyaltyOracle:         "Virtue of Loyalty",
		woeVZLoyaltyOracle + "#1":  "Ardenvale Fealty",
		woeVZStrengthOracle:        "Virtue of Strength",
		woeVZStrengthOracle + "#1": "Garenbrig Growth",
		woeVZVerminOracle:          "Voracious Vermin",
		woeVZTabbyOracle:           "Warehouse Tabby",
		woeVZWingsOracle:           "Water Wings",
		woeVZSweettoothOrcl:        "Welcome to Sweettooth",
		woeVZWerefoxOracle:         "Werefox Bodyguard",
		woeVZVisitorOracle:         "Wicked Visitor",
		woeVZMentorOracle:          "Wildwood Mentor",
		woeVZFrenzyOracle:          "Witchstalker Frenzy",
		woeVZAcolyteOracle:         "Woodland Acolyte",
		woeVZAcolyteOracle + "#1":  "Mend the Wilds",
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

// woeVZVirtue puts an Adventure card into p's hand, on its enchantment face.
func woeVZVirtue(p *game.Player, oracle, name, cost, adventure, adventureType, adventureCost, colour string) game.Card {
	return woeB1Adventure(p, oracle,
		game.Face{Name: name, TypeLine: woeVZEnchantmentType, ManaCost: cost, Colors: []string{colour}},
		game.Face{Name: adventure, TypeLine: adventureType, ManaCost: adventureCost, Colors: []string{colour}})
}

// --- Verdant Outrider --------------------------------------------------------------

func TestWoeVZVerdantOutriderBarsLowPowerBlockersThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	outrider := b12Push(g, me.ID, "Verdant Outrider", "Creature — Human Knight", woeVZOutriderOracle, 4, 2)
	small := pushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
	advanceToMain(t, g)
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, outrider, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	lciB1Attack(t, g, outrider, opp.ID)
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(small, outrider), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, outrider); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
}

// --- Virtue of Courage // Embereth Blaze -------------------------------------------

func TestWoeVZVirtueOfCourageExilesThatManyOnNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Virtue of Courage", woeVZEnchantmentType, woeVZCourageOracle, false)
	source := b12Creature(g, me.ID, "Pinger", woeVZBearType, 2, 2)
	plTop(me, "Deep Card", "Instant", "{U}")
	top := []uuid.UUID{
		plTop(me, "Top Two", "Basic Land — Forest", ""),
		plTop(me, "Top One", woeVZBearType, "{1}{G}"),
	}
	lib := me.Library.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 2) })
	assistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-2 {
		t.Fatalf("library %d -> %d, want two exiled", lib, me.Library.Size())
	}
	for _, id := range top {
		if !b12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %s is not exiled and playable by me", id)
		}
	}
	// Noncombat damage to myself does nothing.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 1) })
	assistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-2 {
		t.Errorf("damage to myself exiled cards: library %d", me.Library.Size())
	}
}

func TestWoeVZEmberethBlazeDealsTwoToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	card := woeVZVirtue(me, woeVZCourageOracle, "Virtue of Courage", "{3}{R}{R}", "Embereth Blaze", "Instant — Adventure", "{1}{R}", "R")
	life := opp.Life
	woeB1CastAdventure(t, g, me, card, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
}

// --- Virtue of Loyalty // Ardenvale Fealty ----------------------------------------

func TestWoeVZVirtueOfLoyaltyCountersAndUntapsYourCreaturesAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Virtue of Loyalty", woeVZEnchantmentType, woeVZLoyaltyOracle, false)
	a := b12Creature(g, me.ID, "Mine A", woeVZBearType, 2, 2)
	b := b12Creature(g, me.ID, "Mine B", woeVZBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", woeVZBearType, 2, 2)
	for _, id := range []uuid.UUID{a, b, theirs} {
		tapCard(g, id)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := plusOneCounters(g, id); n != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", n)
		}
		if isTapped(g, id) {
			t.Error("my creature is still tapped")
		}
	}
	if n := plusOneCounters(g, theirs); n != 0 {
		t.Errorf("the opponent's creature has %d counters, want 0", n)
	}
	if !isTapped(g, theirs) {
		t.Error("the opponent's creature was untapped")
	}
}

func TestWoeVZArdenvaleFealtyMakesAVigilantKnight(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := woeVZVirtue(me, woeVZLoyaltyOracle, "Virtue of Loyalty", "{3}{W}{W}", "Ardenvale Fealty", "Instant — Adventure", "{1}{W}", "W")
	woeB1CastAdventure(t, g, me, card, nil)
	if got := b27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Fatalf("%d Knights, want 1", got)
	}
	knight := findBattlefieldByName(g, "Knight")
	if p, tough := effectivePTOf(t, g, knight); p != 2 || tough != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", p, tough)
	}
	assertKeywords(t, g, knight, "vigilance")
}

// --- Virtue of Strength // Garenbrig Growth ---------------------------------------

func TestWoeVZVirtueOfStrengthTriplesOnlyYourBasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Virtue of Strength", woeVZEnchantmentType, woeVZStrengthOracle, false)

	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	tapForMana(t, g, me.ID, forest)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("a basic Forest made %d mana, want 3", got)
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "G" {
			t.Errorf("three times as much of THAT mana: got %q, want G", tok.Color)
		}
	}

	grove := pushLandFor(g, me.ID, "Nonbasic Grove", "Land — Forest")
	before := me.ManaPool.Total()
	tapForMana(t, g, me.ID, grove)
	if got := me.ManaPool.Total() - before; got != 1 {
		t.Errorf("a nonbasic Forest made %d mana, want 1", got)
	}

	island := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	tapForMana(t, g, opp.ID, island)
	if got := opp.ManaPool.Total(); got != 1 {
		t.Errorf("the opponent's Island made %d mana, want 1", got)
	}
}

func TestWoeVZGarenbrigGrowthReturnsACreatureOrLandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	dead := pushCatalogGraveyardCard(me, "Dead Bear", woeVZBearType, "", 2, 2)
	spell := pushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	card := woeVZVirtue(me, woeVZStrengthOracle, "Virtue of Strength", "{5}{G}{G}", "Garenbrig Growth", "Sorcery — Adventure", "{G}", "G")
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(spell)}); err == nil {
		t.Fatal("an instant card was a legal target")
	}
	woeB1CastAdventure(t, g, me, card, cardRefs(dead))
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// --- Voracious Vermin --------------------------------------------------------------

func TestWoeVZVoraciousVerminMakesARatAndGrowsWhenAnotherOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	vermin := avB1ResolveCreature(t, g, me, "Voracious Vermin", "Creature — Rat", woeVZVerminOracle)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens, want 1", got)
	}
	rat := uuid.Nil
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rat" && c.Controller == me.ID && c.InstanceID != vermin {
				rat = c.InstanceID
			}
		}
	})
	assertRestrictions(t, g, rat, game.CantBlock)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeVZBearType, 2, 2)
	woeB1Destroy(t, g, theirs)
	if n := plusOneCounters(g, vermin); n != 0 {
		t.Fatalf("%d counters after an opponent's creature died, want 0", n)
	}
	woeB1Destroy(t, g, rat)
	if n := plusOneCounters(g, vermin); n != 1 {
		t.Errorf("%d counters after the Rat died, want 1", n)
	}
}

// --- Warehouse Tabby ---------------------------------------------------------------

func TestWoeVZWarehouseTabbyMakesARatWhenYourEnchantmentDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Warehouse Tabby", "Creature — Cat", woeVZTabbyOracle, false)
	woeB1Destroy(t, g, b12Permanent(g, opp.ID, "Their Anthem", woeVZEnchantmentType))
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("%d Rats after an opponent's enchantment died, want 0", got)
	}
	woeB1Destroy(t, g, b12Permanent(g, me.ID, "My Anthem", woeVZEnchantmentType))
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rats after your enchantment died, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
}

func TestWoeVZWarehouseTabbyGainsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	tabby := b12Push(g, me.ID, "Warehouse Tabby", "Creature — Cat", woeVZTabbyOracle, 1, 1)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Fatal("deathtouch before the ability")
	}
	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, tabby, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Error("no deathtouch after the ability resolved")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Error("deathtouch outlived the turn")
	}
}

// --- Water Wings -------------------------------------------------------------------

func TestWoeVZWaterWingsMakesYourCreatureAFourFourFlierWithHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	mine := b12Creature(g, me.ID, "My Bear", woeVZBearType, 1, 1)
	castCatalogSpell(t, g, "Water Wings", "Instant", woeVZWingsOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	assertKeywords(t, g, mine, "flying", "hexproof")
	advanceToNextSeatsTurn(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d on the next turn, want 1/1", p, tough)
	}
}

func TestWoeVZWaterWingsTargetsOnlyYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeVZBearType, 1, 1)
	advanceToMain(t, g)
	id := handCardFull(me, "Water Wings", "Instant", "{1}{U}", woeVZWingsOracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(theirs)}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
}

// --- Welcome to Sweettooth ---------------------------------------------------------

func TestWoeVZWelcomeToSweettoothChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seat := g.Turn.ActiveSeat
	mine := b12Creature(g, me.ID, "My Bear", woeVZBearType, 2, 2)
	castCatalogSpell(t, g, "Welcome to Sweettooth", "Enchantment — Saga", woeVZSweettoothOrcl, nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Fatalf("chapter I: %d Humans, want 1", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("chapter II: %d Foods, want 1", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, mine); n != 2 {
		t.Errorf("chapter III put %d counters, want 2 (one plus one Food)", n)
	}
}

// --- Werefox Bodyguard -------------------------------------------------------------

func TestWoeVZWerefoxBodyguardExilesANonFoxUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, opp.ID, "Their Bear", woeVZBearType, 2, 2)
	fox := b12Creature(g, opp.ID, "Their Fox", "Creature — Fox", 2, 2)
	werefox := avB1ResolveCreature(t, g, me, "Werefox Bodyguard", "Creature — Elf Fox Knight", woeVZWerefoxOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(fox)); err == nil {
		t.Fatal("a Fox is not a legal target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(werefox)); err == nil {
		t.Fatal("the Bodyguard itself is not a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the Bear should be in exile")
	}
	life := me.Life
	fillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, werefox, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want 2 gained", life, me.Life)
	}
	if avB1Named(g, opp.ID, "Their Bear") == uuid.Nil {
		t.Error("the exiled Bear did not return when the Bodyguard left")
	}
}

// --- Wicked Visitor ----------------------------------------------------------------

func TestWoeVZWickedVisitorDrainsWhenYourEnchantmentDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Wicked Visitor", "Creature — Nightmare", woeVZVisitorOracle, false)
	mine, oppLife := me.Life, opp.Life
	woeB1Destroy(t, g, b12Permanent(g, opp.ID, "Their Anthem", woeVZEnchantmentType))
	if opp.Life != oppLife {
		t.Fatalf("an opponent's enchantment drained: %d -> %d", oppLife, opp.Life)
	}
	woeB1Destroy(t, g, b12Permanent(g, me.ID, "My Anthem", woeVZEnchantmentType))
	for _, p := range g.Seats {
		want := oppLife - 1
		if p.ID == me.ID {
			want = mine
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, want)
		}
	}
}

// --- Wildwood Mentor ---------------------------------------------------------------

func TestWoeVZWildwoodMentorGrowsOnYourTokensAndPumpsAnotherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	mentor := b12Push(g, me.ID, "Wildwood Mentor", "Creature — Treefolk", woeVZMentorOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", woeVZBearType, 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Clue", TypeLine: "Token Artifact — Clue"}, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, mentor); n != 0 {
		t.Fatalf("%d counters after an opponent's token entered, want 0", n)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, game.Card{Name: "My Clue", TypeLine: "Token Artifact — Clue"}, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, mentor); n != 1 {
		t.Fatalf("%d counters after your token entered, want 1", n)
	}
	declareAttack(t, g, opp.ID, mentor, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, mentor) {
		t.Error("the Mentor offered itself as a target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (+2/+2 from a 2-power Mentor)", p, tough)
	}
}

// --- Witchstalker Frenzy -----------------------------------------------------------

func TestWoeVZWitchstalkerFrenzyCostsLessPerAttackerAndDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "Witchstalker Frenzy", "Instant", "{3}{R}", woeVZFrenzyOracle, nil); got != 4 {
		t.Errorf("price with no attackers %d, want 4", got)
	}
	a := b12Creature(g, me.ID, "A", woeVZBearType, 2, 2)
	b := b12Creature(g, me.ID, "B", woeVZBearType, 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepPostcombatMain)
	if got := woeILPrice(t, g, me.ID, "Witchstalker Frenzy", "Instant", "{3}{R}", woeVZFrenzyOracle, nil); got != 2 {
		t.Errorf("price after two attackers %d, want 2", got)
	}
	big := b12Creature(g, opp.ID, "Their Five", woeVZBearType, 5, 5)
	bigger := b12Creature(g, opp.ID, "Their Six", woeVZBearType, 6, 6)
	castCatalogSpell(t, g, "Witchstalker Frenzy", "Instant", woeVZFrenzyOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 5/5 survived 5 damage")
	}
	if !g.Battlefield.Contains(bigger) {
		t.Error("an untargeted creature died")
	}
}

// --- Woodland Acolyte // Mend the Wilds --------------------------------------------

func TestWoeVZWoodlandAcolyteDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	avB1ResolveCreature(t, g, me, "Woodland Acolyte", "Creature — Human Cleric", woeVZAcolyteOracle)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestWoeVZMendTheWildsPutsAPermanentCardOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	dead := pushCatalogGraveyardCard(me, "Dead Bear", woeVZBearType, "", 2, 2)
	spell := pushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	card := woeB1Adventure(me, woeVZAcolyteOracle,
		game.Face{Name: "Woodland Acolyte", TypeLine: "Creature — Human Cleric", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Mend the Wilds", TypeLine: "Instant — Adventure", ManaCost: "{G}", Colors: []string{"G"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(spell)}); err == nil {
		t.Fatal("an instant card was a legal target")
	}
	woeB1CastAdventure(t, g, me, card, cardRefs(dead))
	if top := libraryTopIDs(me, 1); len(top) != 1 || top[0] != dead {
		t.Errorf("library top %v, want the Bear card", top)
	}
}
