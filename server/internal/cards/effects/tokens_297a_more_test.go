package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_297a_more_test.go — the rest of slice 297-a: Skyclave
// Apparition, Overlord of the Hauntwoods, Ondu Spiritdancer, Myr
// Battlesphere and Rain of Riches.

const (
	skyclaveApparitionOracle = "d90af00a-d322-4265-9954-7b1e80702e18"
	overlordHauntwoodsOracle = "4669e8c7-fc37-4b97-9cb7-8d29b43d9176"
	onduSpiritdancerOracle   = "d2ae893f-52af-4645-8712-d62fb78f901e"
	myrBattlesphereOracle    = "c53ba31a-ba27-4e17-9a92-311acb1cab29"
	rainOfRichesOracle       = "ef1e2d3d-e977-4729-9430-eba6242e5dfe"
)

// --- Skyclave Apparition ------------------------------------------

func castSkyclaveApparition(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castFromHand(t, g, game.Card{
		Name: "Skyclave Apparition", TypeLine: "Creature — Kor Spirit", OracleID: skyclaveApparitionOracle,
		ManaCost: "{1}{W}{W}", Power: 2, Toughness: 2,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	return id
}

func TestSkyclaveApparitionOffersOnlyCheapNonlandNontokenOpposingPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Big Thing", TypeLine: "Artifact", ManaCost: "{5}", Owner: opp.ID, Controller: opp.ID,
	})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID,
	})
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Servo", TypeLine: "Token Artifact Creature — Servo", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})

	castSkyclaveApparition(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	if !hasID(p.PickTargetCards, rock) {
		t.Error("a mana value 3 artifact they control is not offered")
	}
	for name, id := range map[string]uuid.UUID{"a mana value 5 artifact": big, "a land": land, "a token": token, "my own permanent": mine} {
		if hasID(p.PickTargetCards, id) {
			t.Errorf("%s is offered as a target", name)
		}
	}
}

func TestSkyclaveApparitionExilesForGoodAndPaysTheOwnerWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	apparition := castSkyclaveApparition(t, g)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Fatal("the targeted permanent was not exiled")
	}

	killCreature(t, g, me.ID, apparition)
	passPriorityAroundTable(t, g)

	// The exile is permanent: nothing came back.
	if !g.Exile.Contains(rock) {
		t.Error("the exiled card came back; Skyclave Apparition's exile is not temporary")
	}
	var illusion *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Illusion" {
			illusion = &g.Battlefield.Cards[i]
		}
	}
	if illusion == nil {
		t.Fatal("no Illusion token was created")
	}
	if illusion.Controller != opp.ID {
		t.Errorf("the Illusion belongs to %s, want the exiled card's owner %s", illusion.Controller, opp.ID)
	}
	if illusion.Power != 3 || illusion.Toughness != 3 {
		t.Errorf("the Illusion is %d/%d, want 3/3 (the exiled card's mana value)", illusion.Power, illusion.Toughness)
	}
	if !illusion.HasColor("U") {
		t.Errorf("the Illusion is not blue: %v", illusion.Colors)
	}
}

func TestSkyclaveApparitionWithNothingExiledPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apparition := castSkyclaveApparition(t, g)
	passPriorityAroundTable(t, g)
	killCreature(t, g, me.ID, apparition)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion" {
			t.Error("an Illusion was made although nothing was exiled")
		}
	}
}

// --- Overlord of the Hauntwoods -----------------------------------

func TestOverlordOfTheHauntwoodsMakesATappedEverywhereLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	overlord := castFromHand(t, g, game.Card{
		Name: "Overlord of the Hauntwoods", TypeLine: "Enchantment Creature — Avatar Horror",
		OracleID: overlordHauntwoodsOracle, ManaCost: "{3}{G}{G}", Power: 6, Toughness: 5,
	}, game.CastSpellParams{AlternativeCost: AltCostKeyImpending})
	passPriorityAroundTable(t, g)

	if n := countersOn(g, overlord, game.CounterTime); n != 4 {
		t.Errorf("%d time counters on the impending Overlord, want 4", n)
	}
	if containsString(effectiveTypes(t, g, overlord), "Creature") {
		t.Errorf("the impending Overlord is a creature: %v", effectiveTypes(t, g, overlord))
	}
	var land *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Everywhere" {
			land = &g.Battlefield.Cards[i]
		}
	}
	if land == nil {
		t.Fatal("no Everywhere token was created")
	}
	if land.Controller != me.ID || !land.Tapped || !land.IsToken() || !land.IsLand() {
		t.Errorf("Everywhere is wrong: controller %s tapped %v token %v land %v", land.Controller, land.Tapped, land.IsToken(), land.IsLand())
	}
	for _, sub := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if !land.HasSubtype(sub) {
			t.Errorf("Everywhere is not a %s", sub)
		}
	}
	if n := len(game.ManaAbilitiesForCard(*land)); n != 5 {
		t.Errorf("Everywhere has %d mana abilities, want one per basic land type", n)
	}
	if len(land.Colors) != 0 {
		t.Errorf("Everywhere is coloured %v, want colorless", land.Colors)
	}
}

func TestOverlordOfTheHauntwoodsMakesOneAsAHardCastCreatureToo(t *testing.T) {
	g := newCatalogGame(t)
	overlord := castFromHand(t, g, game.Card{
		Name: "Overlord of the Hauntwoods", TypeLine: "Enchantment Creature — Avatar Horror",
		OracleID: overlordHauntwoodsOracle, ManaCost: "{3}{G}{G}", Power: 6, Toughness: 5,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, overlord, game.CounterTime); n != 0 {
		t.Errorf("%d time counters on the hard-cast Overlord, want 0", n)
	}
	if !containsString(effectiveTypes(t, g, overlord), "Creature") {
		t.Error("the hard-cast Overlord is not a creature")
	}
	if n := cardsNamed(g, "Everywhere", g.Seats[0].ID); n != 1 {
		t.Errorf("%d Everywhere tokens, want 1", n)
	}
}

// --- Ondu Spiritdancer --------------------------------------------

func TestOnduSpiritdancerCopiesAnEnteringEnchantmentOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: onduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	castFromHand(t, g, game.Card{Name: "Test Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Test Enchantment", me.ID); n != 2 {
		t.Fatalf("%d Test Enchantments, want the card and its token copy", n)
	}

	// A second enchantment the same turn is not offered a copy, and
	// the token's own entry was not either.
	castFromHand(t, g, game.Card{Name: "Second Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Second Enchantment", me.ID); n != 1 {
		t.Errorf("%d Second Enchantments, want 1 (once each turn)", n)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Error("Ondu asked about a second copy in the same turn")
		}
	}
}

func TestOnduSpiritdancerDecliningDoesNotUseUpTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: onduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	castFromHand(t, g, game.Card{Name: "Test Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Test Enchantment", me.ID); n != 1 {
		t.Fatalf("%d Test Enchantments after declining, want 1", n)
	}

	castFromHand(t, g, game.Card{Name: "Second Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Second Enchantment", me.ID); n != 2 {
		t.Errorf("%d Second Enchantments, want 2: declining the first copy must not use up the turn", n)
	}
}

// --- Myr Battlesphere ---------------------------------------------

func TestMyrBattlesphereMakesFourMyr(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castFromHand(t, g, game.Card{
		Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct", OracleID: myrBattlesphereOracle,
		ManaCost: "{7}", Power: 4, Toughness: 7,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Myr", me.ID); n != 4 {
		t.Errorf("%d Myr tokens, want 4", n)
	}
}

func TestMyrBattlesphereTapsXMyrToPumpAndHitWhatItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct",
		OracleID: myrBattlesphereOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	var myr []uuid.UUID
	for i := 0; i < 3; i++ {
		myr = append(myr, pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Myr", TypeLine: "Token Artifact Creature — Myr",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		}))
	}
	bear := auraBear(g, me.ID) // not a Myr: never offered
	before := opp.Life

	declareAttack(t, g, opp.ID, sphere)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt for the attack trigger")
	}
	if hasID(prompt.ChooseCards, bear) || hasID(prompt.ChooseCards, sphere) {
		t.Errorf("a Bear or the tapped attacker is offered: %v", prompt.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, myr[0], myr[1])
	passPriorityAroundTable(t, g)

	if opp.Life != before-2 {
		t.Errorf("the defending player is at %d, want %d (X = 2 damage)", opp.Life, before-2)
	}
	if p := effectivePower(t, g, sphere); p != 6 {
		t.Errorf("the Battlesphere has power %d, want 6", p)
	}
	for i, id := range myr {
		c, _ := g.LookupCardForEffect(id)
		if want := i < 2; c.Tapped != want {
			t.Errorf("Myr %d tapped = %v, want %v", i, c.Tapped, want)
		}
	}
}

func TestMyrBattlesphereTappingNoMyrDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct",
		OracleID: myrBattlesphereOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr", TypeLine: "Token Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life
	declareAttack(t, g, opp.ID, sphere)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("the defending player lost %d life without any Myr being tapped", before-opp.Life)
	}
	if p := effectivePower(t, g, sphere); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

// --- Rain of Riches -----------------------------------------------

func cascadeTriggersFor(g *game.Game, label string) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Label == label {
			n++
		}
	}
	return n
}

func TestRainOfRichesMakesTwoTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castFromHand(t, g, game.Card{
		Name: "Rain of Riches", TypeLine: "Enchantment", OracleID: rainOfRichesOracle, ManaCost: "{3}{R}{R}",
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Treasure", me.ID); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
}

func TestRainOfRichesGivesCascadeToTheFirstTreasurePaidSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rain of Riches", TypeLine: "Enchantment", OracleID: rainOfRichesOracle,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	const label = "Rain of Riches — cascade"

	// Paid with plain mana: no cascade.
	floatForTest(g, me, "BB")
	castFromHandForTest(t, g, me, "Plain Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 0 {
		t.Fatalf("a spell paid with ordinary mana cascaded (%d)", n)
	}

	// Paid with a Treasure: cascade, the first time.
	crackATreasureFor(t, g, me, "B")
	floatForTest(g, me, "B")
	castFromHandForTest(t, g, me, "Treasure Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 1 {
		t.Fatalf("the first Treasure-paid spell produced %d cascade triggers, want 1", n)
	}

	// And not the second.
	crackATreasureFor(t, g, me, "B")
	floatForTest(g, me, "B")
	castFromHandForTest(t, g, me, "Second Treasure Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 1 {
		t.Errorf("%d cascade triggers after a second Treasure-paid spell, want still 1", n)
	}
}
