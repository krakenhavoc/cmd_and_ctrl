package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSkyclaveApparitionOffersOnlyCheapNonlandNontokenOpposingPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Big Thing", TypeLine: "Artifact", ManaCost: "{5}", Owner: opp.ID, Controller: opp.ID,
	})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID,
	})
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Servo", TypeLine: "Token Artifact Creature — Servo", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastSkyclaveApparition(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, rock) {
		t.Error("a mana value 3 artifact they control is not offered")
	}
	for name, id := range map[string]uuid.UUID{"a mana value 5 artifact": big, "a land": land, "a token": token, "my own permanent": mine} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("%s is offered as a target", name)
		}
	}
}

func TestSkyclaveApparitionExilesForGoodAndPaysTheOwnerWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	apparition := effectstest.CastSkyclaveApparition(t, g)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Fatal("the targeted permanent was not exiled")
	}

	effectstest.KillCreature(t, g, me.ID, apparition)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	apparition := effectstest.CastSkyclaveApparition(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, apparition)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion" {
			t.Error("an Illusion was made although nothing was exiled")
		}
	}
}

// --- Overlord of the Hauntwoods -----------------------------------

func TestOverlordOfTheHauntwoodsMakesATappedEverywhereLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	overlord := effectstest.CastFromHand(t, g, game.Card{
		Name: "Overlord of the Hauntwoods", TypeLine: "Enchantment Creature — Avatar Horror",
		OracleID: effectstest.OverlordHauntwoodsOracle, ManaCost: "{3}{G}{G}", Power: 6, Toughness: 5,
	}, game.CastSpellParams{AlternativeCost: effects.AltCostKeyImpending})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountersOn(g, overlord, game.CounterTime); n != 4 {
		t.Errorf("%d time counters on the impending Overlord, want 4", n)
	}
	if effectstest.ContainsString(effectstest.EffectiveTypes(t, g, overlord), "Creature") {
		t.Errorf("the impending Overlord is a creature: %v", effectstest.EffectiveTypes(t, g, overlord))
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
	g := effectstest.NewCatalogGame(t)
	overlord := effectstest.CastFromHand(t, g, game.Card{
		Name: "Overlord of the Hauntwoods", TypeLine: "Enchantment Creature — Avatar Horror",
		OracleID: effectstest.OverlordHauntwoodsOracle, ManaCost: "{3}{G}{G}", Power: 6, Toughness: 5,
	}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, overlord, game.CounterTime); n != 0 {
		t.Errorf("%d time counters on the hard-cast Overlord, want 0", n)
	}
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, overlord), "Creature") {
		t.Error("the hard-cast Overlord is not a creature")
	}
	if n := effectstest.CardsNamed(g, "Everywhere", g.Seats[0].ID); n != 1 {
		t.Errorf("%d Everywhere tokens, want 1", n)
	}
}

// --- Ondu Spiritdancer --------------------------------------------

func TestOnduSpiritdancerCopiesAnEnteringEnchantmentOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: effectstest.OnduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastFromHand(t, g, game.Card{Name: "Test Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Test Enchantment", me.ID); n != 2 {
		t.Fatalf("%d Test Enchantments, want the card and its token copy", n)
	}

	// A second enchantment the same turn is not offered a copy, and
	// the token's own entry was not either.
	effectstest.CastFromHand(t, g, game.Card{Name: "Second Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Second Enchantment", me.ID); n != 1 {
		t.Errorf("%d Second Enchantments, want 1 (once each turn)", n)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Error("Ondu asked about a second copy in the same turn")
		}
	}
}

func TestOnduSpiritdancerDecliningDoesNotUseUpTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: effectstest.OnduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastFromHand(t, g, game.Card{Name: "Test Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Test Enchantment", me.ID); n != 1 {
		t.Fatalf("%d Test Enchantments after declining, want 1", n)
	}

	effectstest.CastFromHand(t, g, game.Card{Name: "Second Enchantment", TypeLine: "Enchantment"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Second Enchantment", me.ID); n != 2 {
		t.Errorf("%d Second Enchantments, want 2: declining the first copy must not use up the turn", n)
	}
}

// --- Myr Battlesphere ---------------------------------------------

func TestMyrBattlesphereMakesFourMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastFromHand(t, g, game.Card{
		Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct", OracleID: effectstest.MyrBattlesphereOracle,
		ManaCost: "{7}", Power: 4, Toughness: 7,
	}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Myr", me.ID); n != 4 {
		t.Errorf("%d Myr tokens, want 4", n)
	}
}

func TestMyrBattlesphereTapsXMyrToPumpAndHitWhatItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct",
		OracleID: effectstest.MyrBattlesphereOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	var myr []uuid.UUID
	for i := 0; i < 3; i++ {
		myr = append(myr, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Myr", TypeLine: "Token Artifact Creature — Myr",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		}))
	}
	bear := effectstest.AuraBear(g, me.ID) // not a Myr: never offered
	before := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, sphere)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt for the attack trigger")
	}
	if effectstest.HasID(prompt.ChooseCards, bear) || effectstest.HasID(prompt.ChooseCards, sphere) {
		t.Errorf("a Bear or the tapped attacker is offered: %v", prompt.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, myr[0], myr[1])
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before-2 {
		t.Errorf("the defending player is at %d, want %d (X = 2 damage)", opp.Life, before-2)
	}
	if p := effectstest.EffectivePower(t, g, sphere); p != 6 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr Construct",
		OracleID: effectstest.MyrBattlesphereOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr", TypeLine: "Token Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, sphere)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("the defending player lost %d life without any Myr being tapped", before-opp.Life)
	}
	if p := effectstest.EffectivePower(t, g, sphere); p != 4 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastFromHand(t, g, game.Card{
		Name: "Rain of Riches", TypeLine: "Enchantment", OracleID: effectstest.RainOfRichesOracle, ManaCost: "{3}{R}{R}",
	}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CardsNamed(g, "Treasure", me.ID); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
}

func TestRainOfRichesGivesCascadeToTheFirstTreasurePaidSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rain of Riches", TypeLine: "Enchantment", OracleID: effectstest.RainOfRichesOracle,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	const label = "Rain of Riches — cascade"

	// Paid with plain mana: no cascade.
	effectstest.FloatForTest(g, me, "BB")
	effectstest.CastFromHandForTest(t, g, me, "Plain Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 0 {
		t.Fatalf("a spell paid with ordinary mana cascaded (%d)", n)
	}

	// Paid with a Treasure: cascade, the first time.
	effectstest.CrackATreasureFor(t, g, me, "B")
	effectstest.FloatForTest(g, me, "B")
	effectstest.CastFromHandForTest(t, g, me, "Treasure Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 1 {
		t.Fatalf("the first Treasure-paid spell produced %d cascade triggers, want 1", n)
	}

	// And not the second.
	effectstest.CrackATreasureFor(t, g, me, "B")
	effectstest.FloatForTest(g, me, "B")
	effectstest.CastFromHandForTest(t, g, me, "Second Treasure Spell", "Instant", "{1}{B}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := cascadeTriggersFor(g, label); n != 1 {
		t.Errorf("%d cascade triggers after a second Treasure-paid spell, want still 1", n)
	}
}
