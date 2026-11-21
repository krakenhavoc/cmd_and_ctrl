package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b41Sliver seeds a Sliver creature under owner.
func b41Sliver(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Sliver", 1, 1)
}

// --- registration --------------------------------------------------

func TestBatch41CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B41WallOfDenialOracle:      "Wall of Denial",
		effectstest.B41SquirrelSovereignOracle: "Squirrel Sovereign",
		effectstest.B41BonescytheSliverOracle:  "Bonescythe Sliver",
		effectstest.B41SentinelSliverOracle:    "Sentinel Sliver",
		effectstest.B41VenomSliverOracle:       "Venom Sliver",
		effectstest.B41CrystallineSliverOracle: "Crystalline Sliver",
		effectstest.B41DiffusionSliverOracle:   "Diffusion Sliver",
		effectstest.B41KayasWrathOracle:        "Kaya's Wrath",
		effectstest.B41PhyrexianRebirthOracle:  "Phyrexian Rebirth",
		effectstest.B41BlazingVolleyOracle:     "Blazing Volley",
		effectstest.B41VanishingVerseOracle:    "Vanishing Verse",
		effectstest.B41DarkProphecyOracle:      "Dark Prophecy",
		effectstest.B41RekiOracle:              "Reki, the History of Kamigawa",
		effectstest.B41AuramancerOracle:        "Auramancer",
		effectstest.B41AngelOfFinalityOracle:   "Angel of Finality",
		effectstest.B41IndulgingPatricianOrcl:  "Indulging Patrician",
		effectstest.B41CrestedSunmareOracle:    "Crested Sunmare",
		effectstest.B41UnrulyCatapultOracle:    "Unruly Catapult",
		effectstest.B41SpawningPitOracle:       "Spawning Pit",
		effectstest.B41DissipationFieldOracle:  "Dissipation Field",
		effectstest.B41HanweirGarrisonOracle:   "Hanweir Garrison",
		effectstest.B41ChocoboKnightsOracle:    "Chocobo Knights",
		effectstest.B41PlumecreedEscortOracle:  "Plumecreed Escort",
		effectstest.B41YoshimaruOracle:         "Yoshimaru, Ever Faithful",
		effectstest.B41GrumgullyOracle:         "Grumgully, the Generous",
		effectstest.B41HordewingSkaabOracle:    "Hordewing Skaab",
		effectstest.B41MossDiamondOracle:       "Moss Diamond",
		effectstest.B41IdyllicGrangeOracle:     "Idyllic Grange",
		effectstest.B41BarkformHarvesterOracle: "Barkform Harvester",
		effectstest.B41AetherTunnelOracle:      "Aether Tunnel",
		effectstest.B41CanopyCoverOracle:       "Canopy Cover",
		effectstest.B41StaffOfTitaniaOracle:    "Staff of Titania",
		effectstest.B41SwordOfBodyAndMindOrcl:  "Sword of Body and Mind",
		effectstest.B41UnburialRitesOracle:     "Unburial Rites",
		effectstest.B41SnapbackOracle:          "Snapback",
		effectstest.B41SubmergeOracle:          "Submerge",
		effectstest.B41SpectralDelugeOracle:    "Spectral Deluge",
		effectstest.B41UndeadButlerOracle:      "Undead Butler",
		effectstest.B41OssificationOracle:      "Ossification",
		effectstest.B41GoblinRingleaderOracle:  "Goblin Ringleader",
		effectstest.B41MerryOracle:             "Merry, Esquire of Rohan",
		effectstest.B41NighthowlerOracle:       "Nighthowler",
		effectstest.B41MarchWorldOozeOracle:    "March of the World Ooze",
		effectstest.B41CountOnLuckOracle:       "Count on Luck",
		effectstest.B41AirbendersReversalOrcl:  "Airbender's Reversal",
		effectstest.B41ThopterFabricatorOracle: "Thopter Fabricator",
	}
	if len(want) != 46 {
		t.Fatalf("the batch registers 46 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- lords ---------------------------------------------------------

// The three modern Slivers say "Sliver creatures YOU CONTROL", with no
// "other". Crystalline Sliver says "ALL Slivers". The two wordings are
// one field apart and are the reason all four are tested together.
func TestB41SliverLordsGrantToTheBoardsTheyName(t *testing.T) {
	for _, row := range []struct {
		name, oracle, keyword string
	}{
		{"Bonescythe Sliver", effectstest.B41BonescytheSliverOracle, "double strike"},
		{"Sentinel Sliver", effectstest.B41SentinelSliverOracle, "vigilance"},
		{"Venom Sliver", effectstest.B41VenomSliverOracle, "deathtouch"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lord := effectstest.B41CatalogCreature(g, me.ID, row.name, "Creature — Sliver", row.oracle, 2, 2)
		mine := b41Sliver(g, me.ID, "My Sliver")
		theirs := b41Sliver(g, opp.ID, "Their Sliver")
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if !effectstest.HasEffectiveKeyword(t, g, mine, row.keyword) {
			t.Errorf("%s: your other Sliver has %s", row.name, row.keyword)
		}
		if !effectstest.HasEffectiveKeyword(t, g, lord, row.keyword) {
			t.Errorf("%s: there is no \"other\" — the lord grants it to itself too", row.name)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, row.keyword) {
			t.Errorf("%s: \"you control\" stops at your own board", row.name)
		}
		if effectstest.HasEffectiveKeyword(t, g, bear, row.keyword) {
			t.Errorf("%s: a Bear is not a Sliver", row.name)
		}
	}
}

func TestB41CrystallineSliverShroudsEverySliverAtTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B41CatalogCreature(g, me.ID, "Crystalline Sliver", "Creature — Sliver", effectstest.B41CrystallineSliverOracle, 2, 2)
	mine := b41Sliver(g, me.ID, "My Sliver")
	theirs := b41Sliver(g, opp.ID, "Their Sliver")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{lord, mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "shroud") {
			t.Error("\"all Slivers\" has no controller clause — an opponent's Slivers get shroud too")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("a Bear is not a Sliver")
	}
}

func TestB41SquirrelSovereignPumpsOtherSquirrelsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := effectstest.B41CatalogCreature(g, me.ID, "Squirrel Sovereign", "Creature — Squirrel Noble", effectstest.B41SquirrelSovereignOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Squirrel", "Creature — Squirrel", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("your other Squirrel is a 2/2: power %d", got)
	}
	if got := effectstest.EffectivePower(t, g, lord); got != 2 {
		t.Errorf("\"other\" excludes the Sovereign: power %d, want its printed 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("\"you control\" excludes an opponent's Squirrel: power %d", got)
	}
}

func TestB41WallOfDenialCarriesItsThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.B41CatalogCreature(g, me.ID, "Wall of Denial", "Creature — Wall", effectstest.B41WallOfDenialOracle, 0, 8)
	for _, kw := range []string{"defender", "flying", "shroud"} {
		if !effectstest.HasEffectiveKeyword(t, g, wall, kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
}

// --- sweepers and removal ------------------------------------------

func TestB41KayasWrathDestroysAllAndPaysOnlyForYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	mine := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Mine A", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, me.ID, "Mine B", "Creature — Bear", 2, 2),
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Kaya's Wrath", "Sorcery", effectstest.B41KayasWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range append(mine, theirs) {
		if g.Battlefield.Contains(id) {
			t.Error("every creature is destroyed, yours and theirs")
		}
	}
	if me.Life != before+2 {
		t.Errorf("you gain 1 per creature YOU controlled: %d → %d, want +2", before, me.Life)
	}
}

func TestB41PhyrexianRebirthLeavesAHorrorTheSizeOfTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Phyrexian Rebirth", "Sorcery", effectstest.B41PhyrexianRebirthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	horror := uuid.Nil
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Horror" {
				horror = c.InstanceID
			}
		}
	})
	if horror == uuid.Nil {
		t.Fatal("the Horror token was not created")
	}
	if got := effectstest.EffectivePower(t, g, horror); got != 3 {
		t.Errorf("three creatures died, so the Horror is a 3/3: power %d", got)
	}
	if got := effectstest.EffectiveToughness(t, g, horror); got != 3 {
		t.Errorf("Horror toughness %d, want 3", got)
	}
}

func TestB41BlazingVolleyOnlyBurnsOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	tough := effectstest.B12Creature(g, opp.ID, "Tough", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blazing Volley", "Sorcery", effectstest.B41BlazingVolleyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opposing 1/1 takes lethal")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own 1/1 is untouched — \"your opponents control\"")
	}
	if !g.Battlefield.Contains(tough) {
		t.Error("1 damage does not kill a 2/2")
	}
}

func TestB41VanishingVerseExilesOnlyMonocoloredPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mono", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID,
	})
	gold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Owner: opp.ID, Controller: opp.ID,
	})
	colorless := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{gold, colorless} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Vanishing Verse", TypeLine: "Instant",
			OracleID: effectstest.B41VanishingVerseOracle, Owner: me.ID, Controller: me.ID,
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Error("only a permanent with exactly one colour is a legal target — colourless is not a colour")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Vanishing Verse", "Instant", effectstest.B41VanishingVerseOracle, effectstest.CardRefs(mono))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mono) {
		t.Error("the monocolored creature is exiled")
	}
}

// --- triggers ------------------------------------------------------

func TestB41DarkProphecyDrawsAndCostsALifePerDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41Catalog(g, me.ID, "Dark Prophecy", "Enchantment", effectstest.B41DarkProphecyOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one death, one card: hand %d → %d", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("and one life: %d → %d", life, me.Life)
	}
}

func TestB41RekiDrawsOnlyOnALegendarySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41CatalogCreature(g, me.ID, "Reki, the History of Kamigawa", "Legendary Creature — Human Shaman", effectstest.B41RekiOracle, 1, 2)
	// castCatalogSpell seeds the card into hand and then casts it, so
	// an ordinary spell leaves the hand size where it started and a
	// draw shows up as +1.
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a nonlegendary spell draws nothing: hand %d → %d", hand, me.Hand.Size())
	}
	effectstest.CastCatalogSpell(t, g, "A Legend", "Legendary Creature — Human", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a legendary spell draws a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB41AuramancerMayRebuyAnEnchantmentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aura := effectstest.PushGraveyardPermanent(me, "Rancor", "Enchantment — Aura", "{G}")
	effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Auramancer", "Creature — Human Wizard", effectstest.B41AuramancerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// "You MAY return" is the CR 603.3c optional-trigger prompt; the
	// target is picked only after it is accepted.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, aura) {
		t.Fatalf("the enchantment card is offered: %+v", p)
	}
	if len(p.PickTargetCards) != 1 {
		t.Error("a creature card in the same graveyard is not")
	}
	effectstest.PickCard(t, g, me.ID, aura)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura goes to hand")
	}
}

func TestB41AngelOfFinalityExilesTheTargetedGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.PushGraveyardPermanent(opp, "Their Dead Thing", "Creature — Bear", "{1}{B}")
	mine := effectstest.PushGraveyardPermanent(me, "My Dead Thing", "Creature — Bear", "{1}{B}")
	effectstest.CastCatalogSpell(t, g, "Angel of Finality", "Creature — Angel", effectstest.B41AngelOfFinalityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the targeted player's graveyard is exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("only the targeted player's — yours is untouched")
	}
}

func TestB41IndulgingPatricianDrainsOnlyAfterThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41CatalogCreature(g, me.ID, "Indulging Patrician", "Creature — Vampire Noble", effectstest.B41IndulgingPatricianOrcl, 1, 4)
	before := effectstest.LifeOfOpponents(g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if g.Seats[i+1].Life != b {
			t.Fatal("no life gained this turn: nothing happens")
		}
	}
	// Gain 3 on the next turn cycle, then reach the end step again.
	effectstest.AdvanceToMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	before = effectstest.LifeOfOpponents(g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-3 {
			t.Errorf("seat %d loses 3: %d → %d", i+1, b, got)
		}
	}
}

func TestB41CrestedSunmareMakesAHorseAndSavesTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sunmare := effectstest.B41CatalogCreature(g, me.ID, "Crested Sunmare", "Creature — Horse", effectstest.B41CrestedSunmareOracle, 5, 5)
	horse := effectstest.B12Creature(g, me.ID, "Other Horse", "Creature — Horse", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, horse, "indestructible") {
		t.Error("other Horses you control are indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, sunmare, "indestructible") {
		t.Error("\"other\" excludes the Sunmare itself — that is the card's whole weakness")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	found := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Horse" && c.Controller == me.ID {
				found++
			}
		}
	})
	if found != 1 {
		t.Errorf("one 5/5 Horse token at the end step: %d", found)
	}
}

func TestB41UnrulyCatapultPingsTheTableAndUntapsOnASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	catapult := effectstest.B41CatalogCreature(g, me.ID, "Unruly Catapult", "Artifact Creature — Construct", effectstest.B41UnrulyCatapultOracle, 0, 4)
	before := effectstest.LifeOfOpponents(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, catapult, 0, game.ActivateAbilityParams{})
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("each opponent takes 1: seat %d %d → %d", i+1, b, got)
		}
	}
	if !effectstest.B16Tapped(t, g, catapult) {
		t.Fatal("the Catapult taps to activate")
	}
	effectstest.CastCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, catapult) {
		t.Error("casting an instant untaps it")
	}
}

func TestB41SpawningPitEatsCreaturesAndSpendsTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pit := effectstest.B41Catalog(g, me.ID, "Spawning Pit", "Artifact", effectstest.B41SpawningPitOracle)
	first := effectstest.B12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	second := effectstest.B12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, pit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{first}})
	if effectstest.CounterOn(g, pit, game.CounterCharge) != 1 {
		t.Fatalf("one sacrifice, one charge counter: %d", effectstest.CounterOn(g, pit, game.CounterCharge))
	}
	if err := g.ActivateCatalogAbility(me.ID, pit, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("one counter is not two — the token ability is unaffordable")
	}
	effectstest.B16Activate(t, g, me.ID, pit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{second}})
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, pit, 1, game.ActivateAbilityParams{})
	if effectstest.CounterOn(g, pit, game.CounterCharge) != 0 {
		t.Errorf("both counters are spent at announce: %d left", effectstest.CounterOn(g, pit, game.CounterCharge))
	}
	spawn := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spawn" {
				spawn++
			}
		}
	})
	if spawn != 1 {
		t.Errorf("one 2/2 Spawn token: %d", spawn)
	}
}

func TestB41DissipationFieldBouncesThePermanentThatHitYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B41Catalog(g, me.ID, "Dissipation Field", "Enchantment", effectstest.B41DissipationFieldOracle)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, attacker, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !opp.Hand.Contains(attacker) {
		t.Error("the permanent that dealt the damage goes back to its owner's hand")
	}
}

func TestB41HanweirGarrisonMakesTwoTappedAttackingHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garrison := effectstest.B41CatalogCreature(g, me.ID, "Hanweir Garrison", "Creature — Human Soldier", effectstest.B41HanweirGarrisonOracle, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, garrison)
	effectstest.PassPriorityAroundTable(t, g)
	humans := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Human" || c.Controller != me.ID {
				continue
			}
			humans++
			if !c.Tapped {
				t.Error("the tokens enter tapped")
			}
			if c.AttackingTarget != opp.ID {
				t.Error("and attacking the player the Garrison is attacking")
			}
		}
	})
	if humans != 2 {
		t.Errorf("two Humans: %d", humans)
	}
	if n := effectstest.TriggersOnStackFrom(g, garrison); n != 0 {
		t.Errorf("the tokens were never DECLARED as attackers (CR 508.4), so they retrigger nothing: %d", n)
	}
}

func TestB41ChocoboKnightsGrantDoubleStrikeOnlyToCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knights := effectstest.B41CatalogCreature(g, me.ID, "Chocobo Knights", "Creature — Human Knight", effectstest.B41ChocoboKnightsOracle, 3, 3)
	withCounter := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(withCounter, game.CounterPlusOne, 1) })
	effectstest.DeclareAttack(t, g, opp.ID, knights, withCounter, bare)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, withCounter, "double strike") {
		t.Error("a creature with a counter gains double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, bare, "double strike") {
		t.Error("one without a counter does not")
	}
	if effectstest.HasEffectiveKeyword(t, g, knights, "double strike") {
		t.Error("the Knights themselves have no counter and get nothing")
	}
}

func TestB41PlumecreedEscortGrantsHexproofToACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Plumecreed Escort", "Creature — Bird Scout", effectstest.B41PlumecreedEscortOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the chosen creature gains hexproof")
	}
}

func TestB41YoshimaruGrowsOnAnotherLegendaryPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dog := effectstest.B41CatalogCreature(g, me.ID, "Yoshimaru, Ever Faithful", "Legendary Creature — Dog", effectstest.B41YoshimaruOracle, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, dog, game.CounterPlusOne) != 0 {
		t.Fatal("a nonlegendary permanent does nothing")
	}
	effectstest.CastCatalogSpell(t, g, "A Legendary Rock", "Legendary Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, dog, game.CounterPlusOne) != 1 {
		t.Errorf("a legendary PERMANENT of any type grows him: %d counters", effectstest.CounterOn(g, dog, game.CounterPlusOne))
	}
}

func TestB41GrumgullyCountersOnlyOtherNonHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41CatalogCreature(g, me.ID, "Grumgully, the Generous", "Legendary Creature — Goblin Shaman", effectstest.B41GrumgullyOracle, 3, 3)
	goblin := effectstest.CastCatalogSpell(t, g, "A Goblin", "Creature — Goblin", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, goblin, game.CounterPlusOne) != 1 {
		t.Errorf("a non-Human enters with an extra counter: %d", effectstest.CounterOn(g, goblin, game.CounterPlusOne))
	}
	human := effectstest.CastCatalogSpell(t, g, "A Human", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, human, game.CounterPlusOne) != 0 {
		t.Errorf("a Human gets nothing: %d", effectstest.CounterOn(g, human, game.CounterPlusOne))
	}
}

func TestB41HordewingSkaabGrantsFlyingAndLootsOncePerCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B41CatalogCreature(g, me.ID, "Hordewing Skaab", "Creature — Zombie Horror", effectstest.B41HordewingSkaabOracle, 3, 3)
	zombie := effectstest.B12Creature(g, me.ID, "Other Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, zombie, "flying") {
		t.Error("other Zombies you control fly")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a Bear is not a Zombie")
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, zombie, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	// The optional trigger asks; accepting draws one and queues the
	// discard, so the hand size ends where it started.
	if me.Hand.Size() != hand {
		t.Errorf("draw one, discard one: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- lands and artifacts -------------------------------------------

func TestB41MossDiamondEntersTappedAndTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	diamond := effectstest.CastCatalogSpell(t, g, "Moss Diamond", "Artifact", effectstest.B41MossDiamondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, diamond) {
		t.Fatal("the Diamond enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(diamond) })
	effectstest.B28TapForMana(t, g, me.ID, diamond, "G")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB41IdyllicGrangeNeedsThreePlainsAndThenCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	tapped := effectstest.B12PlayFromHand(t, g, "Idyllic Grange", "Land — Plains", effectstest.B41IdyllicGrangeOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Fatal("two other Plains is not three: it enters tapped")
	}
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToMainOf(t, g, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	grange := effectstest.B12PlayFromHand(t, g, "Idyllic Grange", "Land — Plains", effectstest.B41IdyllicGrangeOracle, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g, grange) {
		t.Fatal("three other Plains: it enters untapped")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("an untapped entry puts a +1/+1 counter on a creature you control: %d", effectstest.CounterOn(g, bear, game.CounterPlusOne))
	}
	effectstest.B28TapForMana(t, g, me.ID, grange, "W")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

func TestB41BarkformHarvesterBottomsACardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	harvester := effectstest.B41CatalogCreature(g, me.ID, "Barkform Harvester", "Artifact Creature — Shapeshifter", effectstest.B41BarkformHarvesterOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, harvester, game.KeywordChangeling) {
		t.Error("changeling reached the effective abilities")
	}
	dead := effectstest.PushGraveyardPermanent(me, "Dead Thing", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, harvester, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	if me.Graveyard.Contains(dead) {
		t.Error("the card leaves the graveyard")
	}
	if !me.Library.Contains(dead) {
		t.Error("and lands in the library")
	}
}

// --- attachments ---------------------------------------------------

func TestB41AetherTunnelPumpsAndUnblocksItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Aether Tunnel", "Enchantment — Aura", effectstest.B41AetherTunnelOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("+1/+0: power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("+1/+0 leaves toughness alone: %d, want 2", got)
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestB41CanopyCoverGrantsHexproofAndNotEvasion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Canopy Cover", "Enchantment — Aura", effectstest.B41CanopyCoverOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("CR 702.11b — \"can't be the target of spells or abilities your opponents control\" IS hexproof")
	}
	// The declared caveat: the conditional block restriction is not
	// modelled, so nothing is set. A test pins the caveat so it cannot
	// quietly become an unconditional (stronger) "can't be blocked".
	effectstest.AssertRestrictions(t, g, bear, 0)
}

func TestB41StaffOfTitaniaScalesWithForestsAndMakesDryads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	staff := effectstest.B41Catalog(g, me.ID, "Staff of Titania", "Artifact — Equipment", effectstest.B41StaffOfTitaniaOracle)
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, staff, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("two Forests YOU control: power %d, want 2+2", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	dryads := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Forest Dryad" {
				dryads++
			}
		}
	})
	if dryads != 1 {
		t.Fatalf("one Forest Dryad token per attack: %d", dryads)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("the new Dryad is a Forest, so the bonus grows in the same beat: power %d, want 5", got)
	}
}

func TestB41SwordOfBodyAndMindMakesAWolfAndMillsTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := effectstest.B41Catalog(g, me.ID, "Sword of Body and Mind", "Artifact — Equipment", effectstest.B41SwordOfBodyAndMindOrcl)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, sword, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+2: power %d", got)
	}
	before := opp.Graveyard.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	wolves := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Wolf" && c.Controller == me.ID {
				wolves++
			}
		}
	})
	if wolves != 1 {
		t.Errorf("you get one 2/2 Wolf: %d", wolves)
	}
	if got := opp.Graveyard.Size() - before; got != 10 {
		t.Errorf("the damaged player mills ten: %d", got)
	}
}

// --- spells --------------------------------------------------------

func TestB41UnburialRitesReanimatesFromHandAndFromTheYard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardPermanent(me, "Dead Fatty", "Creature — Bear", "{5}{B}")
	rites := effectstest.CastCatalogSpell(t, g, "Unburial Rites", "Sorcery", effectstest.B41UnburialRitesOracle, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card comes back to the battlefield")
	}
	if !me.Graveyard.Contains(rites) {
		t.Fatal("a hard-cast Rites is an ordinary sorcery and goes to the graveyard")
	}
	spec, _ := effects.Lookup(effectstest.B41UnburialRitesOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Fatalf("flashback is the offer: %+v", spec.AlternativeCosts)
	}
	if !spec.AlternativeCosts[0].ExileOnLeavingStack {
		t.Error("flashback exiles on leaving the stack — without it the Rites is an every-turn engine")
	}
	second := effectstest.PushGraveyardPermanent(me, "Second Fatty", "Creature — Bear", "{4}{B}")
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "C", "C", "C", "W")
	if err := g.CastSpell(me.ID, rites, game.CastSpellParams{
		FromZone:        string(game.ZoneGraveyard),
		AlternativeCost: "flashback",
		Targets:         effectstest.CardRefs(second),
	}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(second) {
		t.Error("the flashed-back copy reanimates too")
	}
	if me.Graveyard.Contains(rites) {
		t.Error("and the Rites is exiled rather than returning to the graveyard")
	}
}

func TestB41SnapbackCanBePitchedForABlueCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blue := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blue, Name: "A Blue Card", TypeLine: "Instant",
		Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	snap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: snap, Name: "Snapback", TypeLine: "Instant",
		OracleID: effectstest.B41SnapbackOracle, ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, snap, game.CastSpellParams{
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{blue},
		Targets:         effectstest.CardRefs(bear),
	}); err != nil {
		t.Fatalf("pitch the blue card: %v", err)
	}
	if me.Hand.Contains(blue) {
		t.Error("the pitched card is exiled at announce, not on resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
}

func TestB41SubmergeIsFreeOnlyWhenBothLandsAreThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(effectstest.B41SubmergeOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Condition == nil {
		t.Fatal("the free cast is gated by a condition")
	}
	cond := spec.AlternativeCosts[0].Condition
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = cond(g, me.ID) })
		return ok
	}
	if check() {
		t.Fatal("no lands yet")
	}
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if check() {
		t.Error("your Island alone is not enough — an OPPONENT must control a Forest")
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if check() {
		t.Error("your OWN Forest does not satisfy \"an opponent controls a Forest\"")
	}
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if !check() {
		t.Fatal("their Forest plus your Island: the offer is live")
	}
	sub := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: sub, Name: "Submerge", TypeLine: "Instant",
		OracleID: effectstest.B41SubmergeOracle, ManaCost: "{4}{U}", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, sub, game.CastSpellParams{
		AlternativeCost: "free",
		Targets:         effectstest.CardRefs(bear),
	}); err != nil {
		t.Fatalf("cast it for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Library.Contains(bear) {
		t.Error("the creature goes on top of its owner's library")
	}
}

func TestB41SpectralDelugeBouncesBySmallToughnessOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	}
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 1, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Spectral Deluge", "Sorcery", effectstest.B41SpectralDelugeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("toughness 2 with two Islands is swept")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("toughness 3 is not")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("\"your opponents control\" — your own creatures stay")
	}
}

func TestB41DiffusionSliverTaxesAnOpponentTargetingYourSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B41CatalogCreature(g, me.ID, "Diffusion Sliver", "Creature — Sliver", effectstest.B41DiffusionSliverOracle, 1, 1)
	mine := b41Sliver(g, me.ID, "My Sliver")
	effectstest.AdvanceToMainOf(t, g, 1)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Their Removal", TypeLine: "Instant",
		OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(mine)}); err != nil {
		t.Fatalf("their removal spell is a legal cast — this is ward, not hexproof: %v", err)
	}
	// The trigger goes on the stack ABOVE the removal spell and
	// resolves first, asking the SPELL's controller — not the Sliver's
	// — to pay. Priority is passed one step at a time so the answer
	// lands before the spell below it resolves.
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the tax is charged to the spell's controller")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the Sliver's controller is never the one asked to pay")
	}
	prompt := effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("PayCost = %q, want {2}", prompt.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("declined, the targeting spell is countered and the Sliver lives")
	}
}

func TestB41UndeadButlerMillsThenExilesItselfToRebuy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Graveyard.Size()
	butler := effectstest.CastCatalogSpell(t, g, "Undead Butler", "Creature — Zombie", effectstest.B41UndeadButlerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got != 3 {
		t.Fatalf("the entry trigger mills three: %d", got)
	}
	target := effectstest.PushGraveyardPermanent(me, "Dead Fatty", "Creature — Bear", "{5}{B}")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(butler) })
	effectstest.PassPriorityAroundTable(t, g)
	// "You may exile it" is the parent's optional prompt; the
	// reflexive "when you do" that follows is mandatory.
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(butler) {
		t.Error("the Butler exiles itself")
	}
	if !me.Hand.Contains(target) {
		t.Error("and the chosen creature card goes to hand")
	}
}

func TestB41OssificationExilesUntilItLeavesTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	aura := effectstest.CastCatalogSpell(t, g, "Ossification", "Enchantment — Aura", effectstest.B41OssificationOracle, effectstest.CardRefs(plains))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature is exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	// CR 610.3 (#1729): the return is a one-shot effect owed at the
	// next CR 704.3 boundary, with nothing on the stack.
	g.RunStateChecksForTest()
	// It comes back as a NEW OBJECT with a fresh instance ID
	// (CR 400.7), so it is found by name rather than by the ID that
	// went into exile.
	back := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Their Threat" {
				continue
			}
			back++
			if c.Controller != opp.ID {
				t.Error("CR 610.3 — it returns under its OWNER's control, not the Aura controller's")
			}
		}
	})
	if back != 1 {
		t.Fatalf("destroying the Aura gives the permanent back: %d copies on the battlefield", back)
	}
}

func TestB41GoblinRingleaderTakesTheGoblinsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	// PushTop puts each card on top, so the last four pushed are the
	// top four the Ringleader reveals.
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant",
			Owner: me.ID, Controller: me.ID,
		})
	}
	var goblins []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		goblins = append(goblins, id)
		me.Library.PushTop(game.Card{
			InstanceID: id, Name: "A Goblin", TypeLine: "Creature — Goblin",
			Owner: me.ID, Controller: me.ID,
		})
	}
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Not a Goblin", TypeLine: "Creature — Bear",
			Owner: me.ID, Controller: me.ID,
		})
	}
	effectstest.CastCatalogSpell(t, g, "Goblin Ringleader", "Creature — Goblin", effectstest.B41GoblinRingleaderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range goblins {
		if !me.Hand.Contains(id) {
			t.Error("every Goblin card revealed this way goes to hand")
		}
	}
	if me.Hand.Size() == 0 {
		t.Fatal("the hand did not grow at all")
	}
}

func TestB41MerryDrawsOnlyWhenAnotherLegendAttacksAndFirstStrikesOnlyWhenEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merry := effectstest.B41CatalogCreature(g, me.ID, "Merry, Esquire of Rohan", "Legendary Creature — Halfling Knight", effectstest.B41MerryOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, merry, "first strike") {
		t.Error("unequipped, Merry has no first strike")
	}
	sword := effectstest.B12Permanent(g, me.ID, "A Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: merry})
	})
	if !effectstest.HasEffectiveKeyword(t, g, merry, "first strike") {
		t.Error("equipped, he does")
	}
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, merry, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a Bear is not another legendary creature — no draw")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	legend := effectstest.B12Creature(g, me.ID, "Another Legend", "Legendary Creature — Human", 2, 2)
	hand = me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, merry, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Merry plus another legendary creature draws one: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB41NighthowlerGrowsWithEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	howler := effectstest.B41CatalogCreature(g, me.ID, "Nighthowler", "Enchantment Creature — Horror", effectstest.B41NighthowlerOracle, 0, 0)
	if got := effectstest.EffectivePower(t, g, howler); got != 0 {
		t.Fatalf("empty graveyards: a 0/0, power %d", got)
	}
	effectstest.PushGraveyardPermanent(me, "Mine", "Creature — Bear", "{1}{B}")
	effectstest.PushGraveyardPermanent(opp, "Theirs", "Creature — Bear", "{1}{B}")
	effectstest.PushGraveyardPermanent(opp, "Not a creature", "Enchantment", "{1}{B}")
	// A raw push into a graveyard emits no zone-move event, so the
	// layer cache has nothing to invalidate on — the same bump
	// Tarmogoyf's test makes for the same reason.
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectstest.EffectivePower(t, g, howler); got != 2 {
		t.Errorf("creature cards in ALL graveyards: power %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, howler); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

func TestB41MarchOfTheWorldOozeSetsSixSixAndAddsTheType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B41Catalog(g, me.ID, "March of the World Ooze", "Enchantment", effectstest.B41MarchWorldOozeOracle)
	dork := effectstest.B12Creature(g, me.ID, "Elf Druid", "Creature — Elf Druid", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, dork); got != 6 {
		t.Errorf("base power and toughness 6/6: power %d", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("\"creatures you control\" — theirs is untouched: power %d", got)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != dork {
				continue
			}
			if !c.HasSubtype("Ooze") {
				t.Error("it is an Ooze")
			}
			if !c.HasSubtype("Elf") {
				t.Error("IN ADDITION to its other types — it is still an Elf")
			}
		}
	})
	// Counters land on top of the new base: the 7b SET replaces the
	// printed 1/1 and the counters are added after it, so the dork is
	// an 8/8 rather than a 6/6. Read through CurrentPower, which is
	// where the engine folds counters into the post-layer view.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(dork, game.CounterPlusOne, 2) })
	got := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == dork {
				got = c.CurrentPower()
			}
		}
	})
	if got != 8 {
		t.Errorf("two +1/+1 counters on a new 6/6 base: power %d, want 8", got)
	}
}

func TestB41CountOnLuckExilesTheTopCardAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41Catalog(g, me.ID, "Count on Luck", "Enchantment", effectstest.B41CountOnLuckOracle)
	before := 0
	g.ReadSnapshot(func() { before = g.Exile.Size() })
	// The game opens in seat 0's draw step, so its upkeep is already
	// behind us; go round the table to the NEXT one.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	after := 0
	g.ReadSnapshot(func() { after = g.Exile.Size() })
	if after != before+1 {
		t.Errorf("one card exiled at your upkeep: exile %d → %d", before, after)
	}
}

func TestB41AirbendersReversalDestroysAnAttackerOrAirbendsYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Airbender's Reversal", TypeLine: "Instant — Lesson",
		OracleID: effectstest.B41AirbendersReversalOrcl, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: effectstest.CardRefs(attacker),
	}); err != nil {
		t.Fatalf("mode 0 on an attacking creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("mode 0 destroys the attacker")
	}
	// Mode 1 exiles your own creature with a recast grant.
	effectstest.AdvanceToMainOf(t, g, 0)
	second := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: second, Name: "Airbender's Reversal", TypeLine: "Instant — Lesson",
		OracleID: effectstest.B41AirbendersReversalOrcl, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{
		Modes:   []int{1},
		Targets: effectstest.CardRefs(mine),
	}); err != nil {
		t.Fatalf("mode 1 on your own creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("mode 1 airbends it — the creature is exiled")
	}
}

func TestB41ThopterFabricatorMakesAThopterOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B41Catalog(g, me.ID, "Thopter Fabricator", "Artifact — Vehicle", effectstest.B41ThopterFabricatorOracle)
	effectstest.AdvanceToMainOf(t, g, 0)
	thopters := func() int {
		n := 0
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Thopter" && c.Controller == me.ID {
					n++
				}
			}
		})
		return n
	}
	// The draw step already drew the turn's first card.
	if thopters() != 0 {
		t.Fatal("the first draw of the turn makes nothing")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if thopters() != 1 {
		t.Fatalf("the second draw makes one Thopter: %d", thopters())
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if thopters() != 1 {
		t.Errorf("the third does not — it is \"your SECOND card each turn\": %d", thopters())
	}
}
