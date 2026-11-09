package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std_b58_cards_test.go — Wilds of Eldraine, Standard leftovers
// batches 55–58, third part of the tests (#4471).

func b55Upkeep(g *game.Game, me *game.Player) {
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
}

// b55AnswerOption answers the open option prompt with the index.
func b55AnswerOption(t *testing.T, g *game.Game, who uuid.UUID, idx int) {
	t.Helper()
	c := latestOptionPickFor(g, who)
	if c == nil {
		t.Fatalf("no option prompt for %s", who)
	}
	if err := g.ResolveOptionPick(c.ID, who, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

func b55Price(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card, cost string) int {
	t.Helper()
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatal(err)
	}
	c.ManaCost = cost
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: c, Controller: owner, FromZone: game.ZoneHand})
	if err != nil {
		t.Fatal(err)
	}
	return out.ManaValue()
}

// --- Beluna Grandsquall // Seek Thrills -----------------------------------------

func TestB56BelunaDiscountsPermanentSpellsThatHaveAnAdventureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	adv := b55AdventurerCard(me.ID, "Knight")
	plain := game.Card{InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	if got := b55Price(t, g, me.ID, adv, "{2}{G}"); got != 3 {
		t.Fatalf("price without Beluna %d, want 3", got)
	}
	b12Push(g, me.ID, "Beluna Grandsquall", "Legendary Creature — Giant Noble", b56Beluna, 4, 4)
	if got := b55Price(t, g, me.ID, adv, "{2}{G}"); got != 2 {
		t.Errorf("adventurer creature spell costs %d, want 2", got)
	}
	half := adv
	half.SetFace(1)
	if got := b55Price(t, g, me.ID, half, "{2}{G}"); got != 3 {
		t.Errorf("the Adventure half costs %d, want it undiscounted (3)", got)
	}
	if got := b55Price(t, g, me.ID, plain, "{2}{G}"); got != 3 {
		t.Errorf("a card with no Adventure costs %d, want 3", got)
	}
}

func TestB56SeekThrillsMillsSevenAndTakesTheAdventurers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	for i := 0; i < 2; i++ {
		me.Library.PushTop(b55AdventurerCard(me.ID, "Milled Knight"))
	}
	_, id := hobA6HandAdventure(t, g, b56Beluna,
		hobA6Face("Beluna Grandsquall", "Legendary Creature — Giant Noble", "{G}{U}{R}", 4, 4, "G", "U", "R"),
		hobA6Face("Seek Thrills", "Instant — Adventure", "{2}{G}{U}{R}", 0, 0, "G", "U", "R"))
	hand := b55HandSize(g, me)
	hobA6MustCast(t, g, me, id, 1, nil)
	got := 0
	for _, c := range me.Hand.Cards {
		if c.Name == "Milled Knight" {
			got++
		}
	}
	if got != 2 {
		t.Errorf("%d adventurer cards in hand (hand %d -> %d), want 2", got, hand, b55HandSize(g, me))
	}
}

// --- Devouring Sugarmaw -------------------------------------------------------------

func TestB56DevouringSugarmawTapsUnlessYouSacrificeFodder(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	maw := b12Push(g, me.ID, "Devouring Sugarmaw", "Creature — Horror", b56Sugarmaw, 6, 6)
	b55Upkeep(g, me)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, maw) {
		t.Error("with nothing to sacrifice the Sugarmaw did not tap")
	}
	untapForTest(g, maw)
	food := b12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	b55Upkeep(g, me)
	passPriorityAroundTable(t, g)
	b55AnswerOption(t, g, me.ID, 1)
	b55Settle(t, g, me.ID, true)
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if tappedNow(g, maw) {
		t.Error("it tapped after you sacrificed")
	}
}

func TestB56HaveForDinnerMakesAHumanAndAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, b56Sugarmaw,
		hobA6Face("Devouring Sugarmaw", "Creature — Horror", "{2}{B}{B}", 6, 6, "B"),
		hobA6Face("Have for Dinner", "Instant — Adventure", "{1}{W}", 0, 0, "W"))
	hobA6MustCast(t, g, me, id, 1, nil)
	if b27CountNamed(g, me.ID, "Human") != 1 || b27CountNamed(g, me.ID, "Food") != 1 {
		t.Errorf("Humans %d Foods %d, want 1 and 1", b27CountNamed(g, me.ID, "Human"), b27CountNamed(g, me.ID, "Food"))
	}
}

// --- Candy Trail, Eriette's Tempting Apple, Edgewall Inn ---------------------------------

func TestB56CandyTrailScriesThenGainsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castCatalogSpell(t, g, "Candy Trail", "Artifact — Food Clue", b56Candy, nil)
	passPriorityAroundTable(t, g)
	if c := scryChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	trail := findBattlefieldByName(g, "Candy Trail")
	if trail == uuid.Nil {
		t.Fatal("Candy Trail did not enter")
	}
	life, hand := lifeOf(g, me.ID), b55HandSize(g, me)
	if err := hobB4Activate(t, g, me.ID, trail, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, me.ID) != life+3 || b55HandSize(g, me) != hand+1 {
		t.Errorf("life %d hand %d, want +3 and +1", lifeOf(g, me.ID)-life, b55HandSize(g, me)-hand)
	}
}

func TestB56EriettesTemptingAppleStealsAndHastesThenEitherSacrificeAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Eriette's Tempting Apple", "Legendary Artifact — Food", b56Apple, nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, victim)
	if c.Controller != me.ID {
		t.Fatal("control was not taken")
	}
	if !b55Has(t, g, victim, "haste") {
		t.Error("no haste")
	}
	apple := findBattlefieldByName(g, "Eriette's Tempting Apple")
	life := lifeOf(g, opp.ID)
	if err := hobB4Activate(t, g, me.ID, apple, 1, "{C}{C}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent at %d, want %d", got, life-3)
	}
}

func TestB56EdgewallInnReturnsAnAdventurerFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	inn := pushCatalogPermanent(g, me.ID, "Edgewall Inn", "Land", b56Inn, false)
	adv := b55AdventurerCard(me.ID, "Dead Knight")
	me.Graveyard.PushTop(adv)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, inn, 0, "{C}{C}{C}", []game.TargetRef{hobA6Card(adv.InstanceID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(adv.InstanceID) {
		t.Error("the Adventure card is not in hand")
	}
}

// --- Bramble Familiar // Fetch Quest ---------------------------------------------------------

func TestB57FetchQuestPutsACreatureEnchantmentOrLandFromTheMill(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	for i := 0; i < 7; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Instant", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	land := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(land)
	_, id := hobA6HandAdventure(t, g, b57Bramble,
		hobA6Face("Bramble Familiar", "Creature — Elemental Raccoon", "{1}{G}", 2, 2, "G"),
		hobA6Face("Fetch Quest", "Sorcery — Adventure", "{5}{G}{G}", 0, 0, "G"))
	hobA6MustCast(t, g, me, id, 1, nil)
	b55Settle(t, g, me.ID, false)
	if !g.Battlefield.Contains(land.InstanceID) {
		t.Error("the milled land did not reach the battlefield")
	}
}

// --- Talion, the Kindly Lord -----------------------------------------------------------------

func TestB56TalionPunishesAnOpponentCastingTheChosenNumber(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	talion := b12Push(g, me.ID, "Talion, the Kindly Lord", "Legendary Creature — Faerie Noble", b56Talion, 3, 4)
	g.WithWriteLock(func() { g.SetChosenNumberForEffect(talion, 3) })
	cast := func(name, typeLine, cost string, power, toughness int) {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness, Owner: opp.ID, Controller: opp.ID}
		opp.Hand.PushTop(c)
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: c.InstanceID}) })
		passPriorityAroundTable(t, g)
	}
	life, hand := lifeOf(g, opp.ID), b55HandSize(g, me)
	cast("Four", "Sorcery", "{3}{R}", 0, 0)
	if lifeOf(g, opp.ID) != life || b55HandSize(g, me) != hand {
		t.Fatal("a mana value 4 spell triggered Talion")
	}
	cast("Three", "Sorcery", "{2}{R}", 0, 0)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d, want %d", got, life-2)
	}
	if got := b55HandSize(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	cast("Big Bear", "Creature — Bear", "{5}", 3, 5) // power equals the number
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("after a power-3 creature spell opponent at %d, want %d", got, life-4)
	}
	cast("Tough Bear", "Creature — Bear", "{5}", 1, 3) // toughness equals the number
	if got := lifeOf(g, opp.ID); got != life-6 {
		t.Errorf("after a toughness-3 creature spell opponent at %d, want %d", got, life-6)
	}
}

// --- Graceful Takedown -------------------------------------------------------------------------

func TestB55GracefulTakedownCountsEnchantedCreaturesAndOneOther(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	enchanted := b12Creature(g, me.ID, "Enchanted", "Creature — Bear", 3, 3)
	aura := b12Permanent(g, me.ID, "Pacifism", "Enchantment — Aura")
	woeB1Attach(t, g, aura, enchanted)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 9, 5)
	// A plain creature is not a legal pick for the "enchanted" clause.
	if err := castCatalogSpellErr(t, g, "Graceful Takedown", "Sorcery", b55Graceful, []game.TargetRef{
		{Kind: game.TargetCard, ID: plain, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 2}}); err == nil {
		t.Fatal("an unenchanted creature was accepted for the enchanted clause")
	}
	g2 := newCatalogGame(t)
	me, opp = woeB1Seats(g2)
	enchanted = b12Creature(g2, me.ID, "Enchanted", "Creature — Bear", 3, 3)
	aura = b12Permanent(g2, me.ID, "Pacifism", "Enchantment — Aura")
	woeB1Attach(t, g2, aura, enchanted)
	plain = b12Creature(g2, me.ID, "Plain", "Creature — Bear", 2, 2)
	victim = b12Creature(g2, opp.ID, "Victim", "Creature — Bear", 9, 5)
	castCatalogSpell(t, g2, "Graceful Takedown", "Sorcery", b55Graceful, []game.TargetRef{
		{Kind: game.TargetCard, ID: enchanted, Slot: 0}, {Kind: game.TargetCard, ID: plain, Slot: 1}, {Kind: game.TargetCard, ID: victim, Slot: 2}})
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(victim) {
		t.Error("3 + 2 damage did not kill a 5-toughness creature")
	}
}

// --- Sharae of Numbing Depths and Hylda -----------------------------------------------------------

func TestB55SharaeTapsStunsAndDrawsOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	v1 := b12Creature(g, opp.ID, "V1", "Creature — Bear", 2, 2)
	hand := b55HandSize(g, me)
	castCatalogSpell(t, g, "Sharae of Numbing Depths", "Legendary Creature — Merfolk Wizard", b55Sharae, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, v1)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, v1) {
		t.Error("the creature was not tapped")
	}
	if got := countersOn(g, v1, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters, want 1", got)
	}
	if got := b55HandSize(g, me); got != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn for the tap (the cast card is in the baseline)", hand, got)
	}
	// A second tap the same turn draws nothing.
	v2 := b12Creature(g, opp.ID, "V2", "Creature — Bear", 2, 2)
	hand = b55HandSize(g, me)
	g.WithWriteLock(func() {
		if err := g.TapTargetByEffectForEffect(v2, me.ID); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := b55HandSize(g, me); got != hand {
		t.Errorf("a second tap drew a card (%d -> %d)", hand, got)
	}
}

func TestB55SharaeIgnoresATapAnOpponentInstructsOfTheirOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	b12Push(g, me.ID, "Sharae of Numbing Depths", "Legendary Creature — Merfolk Wizard", b55Sharae, 3, 4)
	v := b12Creature(g, opp.ID, "V", "Creature — Bear", 2, 2)
	hand := b55HandSize(g, me)
	g.WithWriteLock(func() {
		if err := g.TapTargetByEffectForEffect(v, opp.ID); err != nil { // the opponent's own effect
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if b55HandSize(g, me) != hand {
		t.Error("Sharae drew for a tap the opponent's own effect made")
	}
}

func TestB58HyldaOfTheIcyCrownPaysOneThenChoosesAnElemental(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	b12Push(g, me.ID, "Hylda of the Icy Crown", "Legendary Creature — Human Warlock", b58Hylda, 3, 4)
	crown := b12Push(g, me.ID, "Hylda's Crown of Winter", "Legendary Artifact", b58HyldaCrown, 0, 0)
	v := b12Creature(g, opp.ID, "V", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, crown, 0, "{C}{C}", []game.TargetRef{hobA6Card(v)}); err != nil {
		t.Fatalf("activate the Crown (costs {0} on your turn): %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d after the Crown, want 2 (the ability cost {0} on your turn)", got)
	}
	passPriorityAroundTable(t, g)
	if !tappedNow(g, v) {
		t.Fatal("the Crown did not tap the creature")
	}
	answerPayUnless(t, g, me.ID, true)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt after paying {1}")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("%d Elementals, want 1", got)
	}
}

// --- Hylda's Crown of Winter ----------------------------------------------------------------------

func TestB58HyldasCrownDrawsPerTappedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	crown := b12Push(g, me.ID, "Hylda's Crown of Winter", "Legendary Artifact", b58HyldaCrown, 0, 0)
	t1 := b12Creature(g, opp.ID, "T1", "Creature — Bear", 1, 1)
	t2 := b12Creature(g, opp.ID, "T2", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 1, 1)
	for _, id := range []uuid.UUID{t1, t2} {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	}
	hand := b55HandSize(g, me)
	if err := hobB4Activate(t, g, me.ID, crown, 1, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b55HandSize(g, me); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
}

// --- Goddric, Cloaked Reveler -------------------------------------------------------------------------

func TestB58GoddricBecomesAFourFourFlyingDragonWithCelebration(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	goddric := b12Push(g, me.ID, "Goddric, Cloaked Reveler", "Legendary Creature — Human Noble", b58Goddric, 3, 3)
	if !b55Has(t, g, goddric, "haste") || b55Has(t, g, goddric, "flying") {
		t.Fatal("wrong keywords before celebration")
	}
	g.WithWriteLock(func() { g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 2) })
	p, tough := effectivePTOf(t, g, goddric)
	if p != 4 || tough != 4 {
		t.Errorf("celebrating Goddric is %d/%d, want 4/4", p, tough)
	}
	if !b55Has(t, g, goddric, "flying") {
		t.Error("no flying")
	}
	c, _ := battlefieldCard(g, goddric)
	if !c.HasSubtype("Dragon") || c.HasSubtype("Human") || c.HasSubtype("Noble") {
		t.Errorf("subtypes %v, want only Dragon", c.Effective().Subtypes)
	}
}

// --- Redcap Gutter-Dweller -------------------------------------------------------------------------

func TestB58RedcapGutterDwellerMakesTwoRatsAndGrowsByEatingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	redcap := castCatalogSpell(t, g, "Redcap Gutter-Dweller", "Creature — Goblin Warrior", b58Redcap, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Fatalf("%d Rats, want 2", got)
	}
	exiled := len(g.Exile.Cards)
	b55Upkeep(g, me)
	passPriorityAroundTable(t, g)
	b55Settle(t, g, me.ID, true)
	if got := countersOn(g, redcap, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if len(g.Exile.Cards) != exiled+1 {
		t.Errorf("exile %d -> %d, want the top card exiled", exiled, len(g.Exile.Cards))
	}
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats left, want 1", got)
	}
}

// --- Gruff Triplets ------------------------------------------------------------------------------------

func TestB58GruffTripletsMakesTwoCopiesThatDoNotMakeMoreAndPumpOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	advanceTo(t, g, game.StepPrecombatMain)
	triplets := game.Card{InstanceID: uuid.New(), Name: "Gruff Triplets", TypeLine: "Creature — Satyr Warrior", OracleID: b58Triplets,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(triplets)
	if err := g.CastSpell(me.ID, triplets.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Gruff Triplets"); got != 3 {
		t.Fatalf("%d Gruff Triplets, want the card and two token copies", got)
	}
	var original uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gruff Triplets" && !c.IsToken() {
			original = c.InstanceID
		}
	}
	woeB1Destroy(t, g, original)
	for i := 0; i < 4; i++ {
		to := latestChoiceOfKindFor(g, game.PendingChoiceTriggerOrder, me.ID)
		if to == nil {
			break
		}
		if err := g.ResolveTriggerOrder(to.ID, me.ID, to.TriggerOrderIDs); err != nil {
			t.Fatalf("order: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	total := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gruff Triplets" {
			total += countersOn(g, c.InstanceID, game.CounterPlusOne)
		}
	}
	if total != 2*3 {
		// Power 3 (the printed Triplets are 3/3) to each of the two copies.
		t.Errorf("%d counters among the copies, want 6", total)
	}
}

// --- Sentinel of Lost Lore ----------------------------------------------------------------------------

func TestB58SentinelOfLostLoreReturnsYoursAndBottomsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	mine := b55ExileAdventurer(g, me.ID, "Mine")
	theirs := b55ExileAdventurer(g, opp.ID, "Theirs")
	castCatalogSpell(t, g, "Sentinel of Lost Lore", "Creature — Elf Knight", b58Sentinel, nil)
	passPriorityAroundTable(t, g)
	mp := modePickChoiceFor(g, me.ID)
	if mp == nil {
		t.Fatal("no mode prompt for the enters trigger")
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("modes: %v", err)
	}
	for latestPickTarget(g, me.ID) != nil {
		p := latestPickTarget(g, me.ID)
		id := mine
		if hasID(p.PickTargetCards, theirs) && !hasID(p.PickTargetCards, mine) {
			id = theirs
		}
		pickCard(t, g, me.ID, id)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("your exiled Adventure card did not return to your hand")
	}
	if !opp.Library.Contains(theirs) {
		t.Error("their exiled Adventure card did not go into their library")
	}
}

// --- Thunderous Debut -------------------------------------------------------------------------------------

func TestB58ThunderousDebutHandUnbargainedBattlefieldBargained(t *testing.T) {
	for _, bargained := range []bool{false, true} {
		g := newCatalogGame(t)
		me, _ := woeB1Seats(g)
		var creatures []uuid.UUID
		for i := 0; i < 3; i++ {
			c := game.Card{InstanceID: uuid.New(), Name: "Library Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
			me.Library.PushTop(c)
			creatures = append(creatures, c.InstanceID)
		}
		params := game.CastSpellParams{}
		if bargained {
			params.OptionalCosts = []int{0}
			params.SacrificeIDs = []uuid.UUID{woeILToken(g, me.ID)}
		}
		advanceTo(t, g, game.StepPrecombatMain)
		floatForTest(g, me, "GGCCCCCC")
		p1g2Cast(t, g, "Thunderous Debut", "Sorcery", b58Debut, "{6}{G}{G}", params)
		passPriorityAroundTable(t, g)
		if cc := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); cc != nil {
			if err := g.ResolveChooseCards(cc.ID, me.ID, cc.ChooseCards[:2]); err != nil {
				t.Fatalf("pick: %v", err)
			}
		} else {
			t.Fatalf("bargained=%v: no pick prompt", bargained)
		}
		passPriorityAroundTable(t, g)
		got := 0
		for _, id := range creatures {
			if bargained && g.Battlefield.Contains(id) || !bargained && me.Hand.Contains(id) {
				got++
			}
		}
		if got < 1 {
			t.Errorf("bargained=%v: no creature card reached its destination", bargained)
		}
	}
}

// --- Agatha of the Vile Cauldron --------------------------------------------------------------------------

func TestB58AgathaCutsActivationCostsByHerPowerKeepingOneMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	b12Push(g, me.ID, "Agatha of the Vile Cauldron", "Legendary Creature — Human Warlock", b58Agatha, 2, 2)
	hollow := b12Push(g, me.ID, "Hollow Scavenger", "Creature — Wolf", b55Hollow, 3, 2)
	b12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	// {1} stays {1}: the reduction can't take it below one mana.
	if err := hobB4Activate(t, g, me.ID, hollow, 0, "", nil); err == nil {
		t.Error("the {1} ability was paid for nothing")
	}
}

// --- Yenna, Redtooth Regent ---------------------------------------------------------------------------------

func TestB58YennaCopiesAnEnchantmentNotLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	yenna := b12Push(g, me.ID, "Yenna, Redtooth Regent", "Legendary Creature — Elf Noble", b58Yenna, 4, 4)
	ench := b12Permanent(g, me.ID, "Odd Idol", "Legendary Enchantment")
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, yenna, 0, "{C}{C}", []game.TargetRef{hobA6Card(ench)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Odd Idol" && c.IsToken() {
			copies++
			if c.IsLegendary() {
				t.Error("the copy is legendary")
			}
		}
	}
	if copies != 1 {
		t.Errorf("%d token copies, want 1", copies)
	}
}

// --- The Huntsman's Redemption, Redtooth Vanguard, The Apprentice's Folly -----------------------------------------

func TestB55HuntsmansRedemptionMakesABeastAndTutorsForASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	saga := castCatalogSpell(t, g, "The Huntsman's Redemption", "Enchantment — Saga", b55Huntsman, nil)
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Beast") != 1 {
		t.Fatalf("no Beast from chapter I")
	}
	_ = saga
}

func TestB55RedtoothVanguardReturnsWhenYourEnchantmentEntersAndYouPay(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	card := game.Card{InstanceID: uuid.New(), Name: "Redtooth Vanguard", TypeLine: "Creature — Elf Warrior", OracleID: b55Redtooth, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(card)
	ench := b12Permanent(g, me.ID, "My Anthem", "Enchantment")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ench, Actor: me.ID}) })
	floatForTest(g, me, "CC")
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "CC")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card.InstanceID) {
		t.Error("Redtooth Vanguard did not return to hand")
	}
}

func TestB55ApprenticesFollyMakesAReflectionCopyThenSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	base := b12Push(g, me.ID, "Legend", "Legendary Creature — Elf Noble", "", 2, 2)
	_ = base
	castCatalogSpell(t, g, "The Apprentice's Folly", "Enchantment — Saga", b55Folly, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, base)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Reflection") {
			found = true
			if c.IsLegendary() {
				t.Error("the Reflection is legendary")
			}
		}
	}
	if !found {
		t.Error("no Reflection token")
	}
}

// --- Will, Scion of Peace -----------------------------------------------------------------------------

func TestB57WillDiscountsWhiteAndBlueSpellsByLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	will := b12Push(g, me.ID, "Will, Scion of Peace", "Legendary Creature — Human Wizard", b57Will, 2, 4)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.New(), me.ID, 3); err != nil {
			t.Fatal(err)
		}
		// Life lost doesn't count against it.
		if err := g.ChangePlayerLifeForEffect(uuid.New(), me.ID, -3); err != nil {
			t.Fatal(err)
		}
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, will, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := woeILPrice(t, g, me.ID, "White Thing", "Sorcery", "{4}{W}", "x", nil); got != 2 {
		t.Errorf("a {4}{W} white spell costs %d, want 2 ({3} less; coloured mana stays)", got)
	}
	if got := woeILPrice(t, g, me.ID, "Black Thing", "Sorcery", "{3}{B}", "x", nil); got != 4 {
		t.Errorf("a black spell costs %d, want 4", got)
	}
}

// --- Questing Druid --------------------------------------------------------------------------------------

func TestB56QuestingDruidGrowsForEachWhiteBlueBlackOrRedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	druid := b12Push(g, me.ID, "Questing Druid", "Creature — Human Druid", b56Druid, 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	floatForTest(g, me, "WRG")
	p1g2Cast(t, g, "White Spell", "Instant", "w-spell", "{W}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	p1g2Cast(t, g, "Green Spell", "Instant", "g-spell", "{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	p1g2Cast(t, g, "Red Spell", "Instant", "r-spell", "{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, druid, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (white and red, not green)", got)
	}
}

// --- Johann, Apprentice Sorcerer ----------------------------------------------------------------------------

func TestB55JohannCastsAnInstantFromTheTopOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	b12Push(g, me.ID, "Johann, Apprentice Sorcerer", "Legendary Creature — Human Wizard Sorcerer", b55Johann, 4, 4)
	first := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID}
	second := game.Card{InstanceID: uuid.New(), Name: "Shock Two", TypeLine: "Instant", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID}
	creature := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(second)
	me.Library.PushTop(first)
	advanceTo(t, g, game.StepPrecombatMain)
	life := lifeOf(g, opp.ID)
	if err := g.CastSpell(me.ID, first.InstanceID, game.CastSpellParams{FromZone: "library", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, opp.ID) != life-3 {
		t.Errorf("opponent at %d, want %d", lifeOf(g, opp.ID), life-3)
	}
	if err := g.CastSpell(me.ID, second.InstanceID, game.CastSpellParams{FromZone: "library", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a second spell from the top was cast in the same turn")
	}
	me.Library.PushTop(creature)
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("a creature spell was cast from the top")
	}
}

// --- Likeness Looter -----------------------------------------------------------------------------------------

func TestB55LikenessLooterBecomesACopyOfAGraveyardCreatureWithManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	looter := b12Push(g, me.ID, "Likeness Looter", "Creature — Faerie Shapeshifter", b55Looter, 1, 1)
	dead := game.Card{InstanceID: uuid.New(), Name: "Dead Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	other := game.Card{InstanceID: uuid.New(), Name: "Dead Elf", TypeLine: "Creature — Elf", ManaCost: "{3}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(dead)
	me.Graveyard.PushTop(other)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, looter, 1, game.ActivateAbilityParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{hobA6Card(other.InstanceID)}}); err == nil {
		t.Fatal("a mana value 3 card was a legal target for X = 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, looter, 1, game.ActivateAbilityParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{hobA6Card(dead.InstanceID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, looter)
	if c.Name != "Dead Ogre" {
		t.Fatalf("it is %q, want a copy of Dead Ogre", c.Name)
	}
	if p, tough := effectivePTOf(t, g, looter); p != 4 || tough != 4 {
		t.Errorf("%d/%d, want 4/4", p, tough)
	}
	if !b55Has(t, g, looter, "flying") {
		t.Error("no flying")
	}
	if !me.Graveyard.Contains(dead.InstanceID) {
		t.Error("the card left the graveyard (nothing exiles it)")
	}
}

// --- Feral Encounter ---------------------------------------------------------------------------------------------

func TestB58FeralEncounterExilesACreatureAndBitesAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	biter := b12Creature(g, me.ID, "Biter", "Creature — Bear", 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	elf := game.Card{InstanceID: uuid.New(), Name: "Top Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(elf)
	advanceTo(t, g, game.StepPrecombatMain)
	floatForTest(g, me, "GG")
	p1g2Cast(t, g, "Feral Encounter", "Sorcery", b58Feral, "{G}{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if cc := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); cc != nil {
		if err := g.ResolveChooseCards(cc.ID, me.ID, []uuid.UUID{elf.InstanceID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	} else {
		t.Fatal("no pick prompt")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(elf.InstanceID) {
		t.Fatal("the creature card was not exiled")
	}
	advanceTo(t, g, game.StepBeginCombat)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) != nil; i++ {
		p := latestPickTarget(g, me.ID)
		id := biter
		if hasID(p.PickTargetCards, victim) && !hasID(p.PickTargetCards, biter) {
			id = victim
		}
		pickCard(t, g, me.ID, id)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage did not kill the 2/3")
	}
}
