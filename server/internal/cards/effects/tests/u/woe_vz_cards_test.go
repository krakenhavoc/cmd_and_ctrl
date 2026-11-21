package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsv "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/v"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_vz_cards_test.go — Wilds of Eldraine, the V–Z batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeVZOutriderOracle  = "8fc049e7-a16d-4b48-b930-5c346f6718b9"
	woeVZCourageOracle   = cardsv.VirtueOfCourageOracleID
	woeVZLoyaltyOracle   = cardsv.VirtueOfLoyaltyOracleID
	woeVZStrengthOracle  = cardsv.VirtueOfStrengthOracleID
	woeVZVerminOracle    = "8d36d80e-9934-4b56-bf3c-2b18a6da3d2d"
	woeVZTabbyOracle     = "a7736614-cdd3-43bd-ab3e-461e3e24ed40"
	woeVZWingsOracle     = "17f8e8cb-cd1b-4d94-9efa-ce268e8e7096"
	woeVZSweettoothOrcl  = "9760c396-fc7d-41bc-8f03-5ee359eaec00"
	woeVZWerefoxOracle   = "d5ee2ced-29f4-430f-962e-2f930b92624c"
	woeVZVisitorOracle   = "10330a79-a416-40da-bf5a-8b7cbd9a0498"
	woeVZMentorOracle    = "f4e61b03-e6c8-4751-9720-b91fdecd57da"
	woeVZFrenzyOracle    = "f46befcb-6a7f-4f5e-8634-393cbf4b3308"
	woeVZAcolyteOracle   = cardsw.WoodlandAcolyteOracleID
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

// woeVZVirtue puts an Adventure card into p's hand, on its enchantment face.
func woeVZVirtue(p *game.Player, oracle, name, cost, adventure, adventureType, adventureCost, colour string) game.Card {
	return woeB1Adventure(p, oracle,
		game.Face{Name: name, TypeLine: woeVZEnchantmentType, ManaCost: cost, Colors: []string{colour}},
		game.Face{Name: adventure, TypeLine: adventureType, ManaCost: adventureCost, Colors: []string{colour}})
}

// --- Verdant Outrider --------------------------------------------------------------

func TestWoeVZVerdantOutriderBarsLowPowerBlockersThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	outrider := effectstest.B12Push(g, me.ID, "Verdant Outrider", "Creature — Human Knight", woeVZOutriderOracle, 4, 2)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, outrider, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.LciB1Attack(t, g, outrider, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(small, outrider), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, outrider); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
}

// --- Virtue of Courage // Embereth Blaze -------------------------------------------

func TestWoeVZVirtueOfCourageExilesThatManyOnNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Virtue of Courage", woeVZEnchantmentType, woeVZCourageOracle, false)
	source := effectstest.B12Creature(g, me.ID, "Pinger", woeVZBearType, 2, 2)
	effectstest.PlTop(me, "Deep Card", "Instant", "{U}")
	top := []uuid.UUID{
		effectstest.PlTop(me, "Top Two", "Basic Land — Forest", ""),
		effectstest.PlTop(me, "Top One", woeVZBearType, "{1}{G}"),
	}
	lib := me.Library.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 2) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-2 {
		t.Fatalf("library %d -> %d, want two exiled", lib, me.Library.Size())
	}
	for _, id := range top {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %s is not exiled and playable by me", id)
		}
	}
	// Noncombat damage to myself does nothing.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 1) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-2 {
		t.Errorf("damage to myself exiled cards: library %d", me.Library.Size())
	}
}

func TestWoeVZEmberethBlazeDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	card := woeVZVirtue(me, woeVZCourageOracle, "Virtue of Courage", "{3}{R}{R}", "Embereth Blaze", "Instant — Adventure", "{1}{R}", "R")
	life := opp.Life
	woeB1CastAdventure(t, g, me, card, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
}

// --- Virtue of Loyalty // Ardenvale Fealty ----------------------------------------

func TestWoeVZVirtueOfLoyaltyCountersAndUntapsYourCreaturesAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Virtue of Loyalty", woeVZEnchantmentType, woeVZLoyaltyOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Mine A", woeVZBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Mine B", woeVZBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", woeVZBearType, 2, 2)
	for _, id := range []uuid.UUID{a, b, theirs} {
		effectstest.TapCard(g, id)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.PlusOneCounters(g, id); n != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", n)
		}
		if effectstest.IsTapped(g, id) {
			t.Error("my creature is still tapped")
		}
	}
	if n := effectstest.PlusOneCounters(g, theirs); n != 0 {
		t.Errorf("the opponent's creature has %d counters, want 0", n)
	}
	if !effectstest.IsTapped(g, theirs) {
		t.Error("the opponent's creature was untapped")
	}
}

func TestWoeVZArdenvaleFealtyMakesAVigilantKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	card := woeVZVirtue(me, woeVZLoyaltyOracle, "Virtue of Loyalty", "{3}{W}{W}", "Ardenvale Fealty", "Instant — Adventure", "{1}{W}", "W")
	woeB1CastAdventure(t, g, me, card, nil)
	if got := effectstest.B27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Fatalf("%d Knights, want 1", got)
	}
	knight := effectstest.FindBattlefieldByName(g, "Knight")
	if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 2 || tough != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", p, tough)
	}
	effectstest.AssertKeywords(t, g, knight, "vigilance")
}

// --- Virtue of Strength // Garenbrig Growth ---------------------------------------

func TestWoeVZVirtueOfStrengthTriplesOnlyYourBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Virtue of Strength", woeVZEnchantmentType, woeVZStrengthOracle, false)

	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("a basic Forest made %d mana, want 3", got)
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "G" {
			t.Errorf("three times as much of THAT mana: got %q, want G", tok.Color)
		}
	}

	grove := effectstest.PushLandFor(g, me.ID, "Nonbasic Grove", "Land — Forest")
	before := me.ManaPool.Total()
	effectstest.TapForMana(t, g, me.ID, grove)
	if got := me.ManaPool.Total() - before; got != 1 {
		t.Errorf("a nonbasic Forest made %d mana, want 1", got)
	}

	island := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.TapForMana(t, g, opp.ID, island)
	if got := opp.ManaPool.Total(); got != 1 {
		t.Errorf("the opponent's Island made %d mana, want 1", got)
	}
}

func TestWoeVZGarenbrigGrowthReturnsACreatureOrLandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	dead := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", woeVZBearType, "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	card := woeVZVirtue(me, woeVZStrengthOracle, "Virtue of Strength", "{5}{G}{G}", "Garenbrig Growth", "Sorcery — Adventure", "{G}", "G")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(spell)}); err == nil {
		t.Fatal("an instant card was a legal target")
	}
	woeB1CastAdventure(t, g, me, card, effectstest.CardRefs(dead))
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// --- Voracious Vermin --------------------------------------------------------------

func TestWoeVZVoraciousVerminMakesARatAndGrowsWhenAnotherOfYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	vermin := effectstest.AvB1ResolveCreature(t, g, me, "Voracious Vermin", "Creature — Rat", woeVZVerminOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
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
	effectstest.AssertRestrictions(t, g, rat, game.CantBlock)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeVZBearType, 2, 2)
	effectstest.WoeB1Destroy(t, g, theirs)
	if n := effectstest.PlusOneCounters(g, vermin); n != 0 {
		t.Fatalf("%d counters after an opponent's creature died, want 0", n)
	}
	effectstest.WoeB1Destroy(t, g, rat)
	if n := effectstest.PlusOneCounters(g, vermin); n != 1 {
		t.Errorf("%d counters after the Rat died, want 1", n)
	}
}

// --- Warehouse Tabby ---------------------------------------------------------------

func TestWoeVZWarehouseTabbyMakesARatWhenYourEnchantmentDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Warehouse Tabby", "Creature — Cat", woeVZTabbyOracle, false)
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, opp.ID, "Their Anthem", woeVZEnchantmentType))
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("%d Rats after an opponent's enchantment died, want 0", got)
	}
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, me.ID, "My Anthem", woeVZEnchantmentType))
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rats after your enchantment died, want 1", got)
	}
	effectstest.AssertRestrictions(t, g, effectstest.FindBattlefieldByName(g, "Rat"), game.CantBlock)
}

func TestWoeVZWarehouseTabbyGainsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	tabby := effectstest.B12Push(g, me.ID, "Warehouse Tabby", "Creature — Cat", woeVZTabbyOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Fatal("deathtouch before the ability")
	}
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, tabby, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Error("no deathtouch after the ability resolved")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, tabby, "deathtouch") {
		t.Error("deathtouch outlived the turn")
	}
}

// --- Water Wings -------------------------------------------------------------------

func TestWoeVZWaterWingsMakesYourCreatureAFourFourFlierWithHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeVZBearType, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Water Wings", "Instant", woeVZWingsOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	effectstest.AssertKeywords(t, g, mine, "flying", "hexproof")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d on the next turn, want 1/1", p, tough)
	}
}

func TestWoeVZWaterWingsTargetsOnlyYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeVZBearType, 1, 1)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Water Wings", "Instant", "{1}{U}", woeVZWingsOracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
}

// --- Welcome to Sweettooth ---------------------------------------------------------

func TestWoeVZWelcomeToSweettoothChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	seat := g.Turn.ActiveSeat
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeVZBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Welcome to Sweettooth", "Enchantment — Saga", woeVZSweettoothOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Human"); got != 1 {
		t.Fatalf("chapter I: %d Humans, want 1", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("chapter II: %d Foods, want 1", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, mine); n != 2 {
		t.Errorf("chapter III put %d counters, want 2 (one plus one Food)", n)
	}
}

// --- Werefox Bodyguard -------------------------------------------------------------

func TestWoeVZWerefoxBodyguardExilesANonFoxUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", woeVZBearType, 2, 2)
	fox := effectstest.B12Creature(g, opp.ID, "Their Fox", "Creature — Fox", 2, 2)
	werefox := effectstest.AvB1ResolveCreature(t, g, me, "Werefox Bodyguard", "Creature — Elf Fox Knight", woeVZWerefoxOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(fox)); err == nil {
		t.Fatal("a Fox is not a legal target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(werefox)); err == nil {
		t.Fatal("the Bodyguard itself is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the Bear should be in exile")
	}
	life := me.Life
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, werefox, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want 2 gained", life, me.Life)
	}
	if effectstest.AvB1Named(g, opp.ID, "Their Bear") == uuid.Nil {
		t.Error("the exiled Bear did not return when the Bodyguard left")
	}
}

// --- Wicked Visitor ----------------------------------------------------------------

func TestWoeVZWickedVisitorDrainsWhenYourEnchantmentDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Wicked Visitor", "Creature — Nightmare", woeVZVisitorOracle, false)
	mine, oppLife := me.Life, opp.Life
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, opp.ID, "Their Anthem", woeVZEnchantmentType))
	if opp.Life != oppLife {
		t.Fatalf("an opponent's enchantment drained: %d -> %d", oppLife, opp.Life)
	}
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, me.ID, "My Anthem", woeVZEnchantmentType))
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	mentor := effectstest.B12Push(g, me.ID, "Wildwood Mentor", "Creature — Treefolk", woeVZMentorOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeVZBearType, 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Clue", TypeLine: "Token Artifact — Clue"}, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, mentor); n != 0 {
		t.Fatalf("%d counters after an opponent's token entered, want 0", n)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, game.Card{Name: "My Clue", TypeLine: "Token Artifact — Clue"}, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, mentor); n != 1 {
		t.Fatalf("%d counters after your token entered, want 1", n)
	}
	effectstest.DeclareAttack(t, g, opp.ID, mentor, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, mentor) {
		t.Error("the Mentor offered itself as a target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (+2/+2 from a 2-power Mentor)", p, tough)
	}
}

// --- Witchstalker Frenzy -----------------------------------------------------------

func TestWoeVZWitchstalkerFrenzyCostsLessPerAttackerAndDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "Witchstalker Frenzy", "Instant", "{3}{R}", woeVZFrenzyOracle, nil); got != 4 {
		t.Errorf("price with no attackers %d, want 4", got)
	}
	a := effectstest.B12Creature(g, me.ID, "A", woeVZBearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", woeVZBearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if got := woeILPrice(t, g, me.ID, "Witchstalker Frenzy", "Instant", "{3}{R}", woeVZFrenzyOracle, nil); got != 2 {
		t.Errorf("price after two attackers %d, want 2", got)
	}
	big := effectstest.B12Creature(g, opp.ID, "Their Five", woeVZBearType, 5, 5)
	bigger := effectstest.B12Creature(g, opp.ID, "Their Six", woeVZBearType, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Witchstalker Frenzy", "Instant", woeVZFrenzyOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 5/5 survived 5 damage")
	}
	if !g.Battlefield.Contains(bigger) {
		t.Error("an untargeted creature died")
	}
}

// --- Woodland Acolyte // Mend the Wilds --------------------------------------------

func TestWoeVZWoodlandAcolyteDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.AvB1ResolveCreature(t, g, me, "Woodland Acolyte", "Creature — Human Cleric", woeVZAcolyteOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestWoeVZMendTheWildsPutsAPermanentCardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	dead := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", woeVZBearType, "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	card := woeB1Adventure(me, woeVZAcolyteOracle,
		game.Face{Name: "Woodland Acolyte", TypeLine: "Creature — Human Cleric", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Mend the Wilds", TypeLine: "Instant — Adventure", ManaCost: "{G}", Colors: []string{"G"}})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(spell)}); err == nil {
		t.Fatal("an instant card was a legal target")
	}
	woeB1CastAdventure(t, g, me, card, effectstest.CardRefs(dead))
	if top := effectstest.LibraryTopIDs(me, 1); len(top) != 1 || top[0] != dead {
		t.Errorf("library top %v, want the Bear card", top)
	}
}
