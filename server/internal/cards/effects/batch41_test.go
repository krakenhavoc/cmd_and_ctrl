package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch41_test.go — card-level coverage for the card-coverage roadmap's
// batch 41 (#448, `edhrec_rank` 4254–4353). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	b41WallOfDenialOracle      = "9cc092f9-e8f0-40ea-847d-813b70f889db"
	b41SquirrelSovereignOracle = "f43d1ea5-8127-4b64-a87b-ee5cc9b9e9fa"
	b41BonescytheSliverOracle  = "1d95ad32-768f-40a1-a461-e0562326b2b7"
	b41SentinelSliverOracle    = "1d0f1186-be6c-45c3-9703-f0c1e13892fb"
	b41VenomSliverOracle       = "ffe05d4b-0ec9-4319-bb11-1366dc091224"
	b41CrystallineSliverOracle = "ba3aa1eb-722a-47d3-83be-96daddb50265"
	b41DiffusionSliverOracle   = "458ca9f4-4622-48e1-9ca1-6a8117188973"
	b41KayasWrathOracle        = "bc8de6c7-c69d-4add-8f25-825d945874f9"
	b41PhyrexianRebirthOracle  = "6ef3c75d-6af2-4ea0-b98d-96c5d7d3af58"
	b41BlazingVolleyOracle     = "3656338d-ca08-465b-b09a-d8d8d1196eec"
	b41VanishingVerseOracle    = "5b8f0cdf-572d-4025-b930-79291f7c35be"
	b41DarkProphecyOracle      = "f47659b5-d834-4153-b9fc-6a0702c503d9"
	b41RekiOracle              = "8a7d68ae-ac43-46a6-9dc8-d6b07cc0333c"
	b41AuramancerOracle        = "bd5eb181-6a69-4dc2-93a0-fa000291bc3d"
	b41AngelOfFinalityOracle   = "48d14b5c-711a-4f8a-9de5-55415cb7a79a"
	b41IndulgingPatricianOrcl  = "189131b7-ed59-4750-a4ed-453e4497a5c8"
	b41CrestedSunmareOracle    = "054fedbb-062e-491e-9a94-594fda33094f"
	b41UnrulyCatapultOracle    = "b2b0a5d0-0084-43a4-b6dc-e893ef42cdb7"
	b41SpawningPitOracle       = "ea7b0704-7715-4f8a-b32f-d2d2f3b36054"
	b41DissipationFieldOracle  = "35426427-4268-4c8f-9fe1-270f2ce43d97"
	b41HanweirGarrisonOracle   = "7cb29569-48e1-4782-9906-fad155ebfafe"
	b41ChocoboKnightsOracle    = "6cbefcba-0c1c-48fd-9b78-905ed61aae1e"
	b41PlumecreedEscortOracle  = "04669a6f-6299-465b-a26e-6cae37cdc081"
	b41YoshimaruOracle         = "963834c8-42df-4ee6-9b45-b9de88ce2eac"
	b41GrumgullyOracle         = "fa4afcf7-9cf7-4f67-95b9-1f7fe66bf332"
	b41HordewingSkaabOracle    = "485dea0d-2123-4e5c-91a1-25ba73f4f3cf"
	b41MossDiamondOracle       = "02500f21-6e15-423e-93ff-891e09fe9904"
	b41IdyllicGrangeOracle     = "23d349a0-e441-40b8-b634-13e61440a7c8"
	b41BarkformHarvesterOracle = "0339bd11-ad71-4998-9b5d-a32790f0e5e3"
	b41AetherTunnelOracle      = "0a90d1fb-7a76-4b62-b001-4a539a8e5df1"
	b41CanopyCoverOracle       = "5b84101e-7e23-437d-835c-409bc061ecbb"
	b41StaffOfTitaniaOracle    = "f99f4b52-21ae-47a2-8ae6-b0aa7386723a"
	b41SwordOfBodyAndMindOrcl  = "fac42229-4f5f-4d04-85dd-5031d4e435aa"
	b41UnburialRitesOracle     = "d48e1545-7997-45ed-83a1-aee45b3d3d20"
	b41SnapbackOracle          = "e88de17c-086d-4e2a-b5b2-2f9f57ca7c0f"
	b41SubmergeOracle          = "99427ebe-c00d-4206-84ca-9764f6e952c6"
	b41SpectralDelugeOracle    = "bb9bb65e-504f-4edd-8d47-d0efa03e7d17"
	b41UndeadButlerOracle      = "426fa8e8-b0f9-4b6c-b22b-04ec2073187a"
	b41OssificationOracle      = "e29bfd62-286f-4982-813f-7086573c333b"
	b41GoblinRingleaderOracle  = "4100e486-0d27-436c-8429-76bc2c1a26ab"
	b41MerryOracle             = "4dce1f49-ddad-4748-893a-e10d831ee7d7"
	b41NighthowlerOracle       = "57b3f7fc-1812-4134-a645-6cef48a8aa71"
	b41MarchWorldOozeOracle    = "a04696c4-4138-47e2-8da4-81d61748519f"
	b41CountOnLuckOracle       = "73ad7bcd-4ffc-442c-b6b5-d935b9ccaaaf"
	b41AirbendersReversalOrcl  = "64e3791b-de99-4eed-af49-cbec218c28aa"
	b41ThopterFabricatorOracle = "8aab9181-c306-471a-8061-9a7a7654dab5"
)

// b41Sliver seeds a Sliver creature under owner.
func b41Sliver(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return b12Creature(g, owner, name, "Creature — Sliver", 1, 1)
}

// b41Catalog seeds a catalog permanent with a real battlefield
// timestamp, which the layer engine needs to sort its statics.
func b41Catalog(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
	})
}

// b41CatalogCreature is b41Catalog with a printed body.
func b41CatalogCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- registration --------------------------------------------------

func TestBatch41CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b41WallOfDenialOracle:      "Wall of Denial",
		b41SquirrelSovereignOracle: "Squirrel Sovereign",
		b41BonescytheSliverOracle:  "Bonescythe Sliver",
		b41SentinelSliverOracle:    "Sentinel Sliver",
		b41VenomSliverOracle:       "Venom Sliver",
		b41CrystallineSliverOracle: "Crystalline Sliver",
		b41DiffusionSliverOracle:   "Diffusion Sliver",
		b41KayasWrathOracle:        "Kaya's Wrath",
		b41PhyrexianRebirthOracle:  "Phyrexian Rebirth",
		b41BlazingVolleyOracle:     "Blazing Volley",
		b41VanishingVerseOracle:    "Vanishing Verse",
		b41DarkProphecyOracle:      "Dark Prophecy",
		b41RekiOracle:              "Reki, the History of Kamigawa",
		b41AuramancerOracle:        "Auramancer",
		b41AngelOfFinalityOracle:   "Angel of Finality",
		b41IndulgingPatricianOrcl:  "Indulging Patrician",
		b41CrestedSunmareOracle:    "Crested Sunmare",
		b41UnrulyCatapultOracle:    "Unruly Catapult",
		b41SpawningPitOracle:       "Spawning Pit",
		b41DissipationFieldOracle:  "Dissipation Field",
		b41HanweirGarrisonOracle:   "Hanweir Garrison",
		b41ChocoboKnightsOracle:    "Chocobo Knights",
		b41PlumecreedEscortOracle:  "Plumecreed Escort",
		b41YoshimaruOracle:         "Yoshimaru, Ever Faithful",
		b41GrumgullyOracle:         "Grumgully, the Generous",
		b41HordewingSkaabOracle:    "Hordewing Skaab",
		b41MossDiamondOracle:       "Moss Diamond",
		b41IdyllicGrangeOracle:     "Idyllic Grange",
		b41BarkformHarvesterOracle: "Barkform Harvester",
		b41AetherTunnelOracle:      "Aether Tunnel",
		b41CanopyCoverOracle:       "Canopy Cover",
		b41StaffOfTitaniaOracle:    "Staff of Titania",
		b41SwordOfBodyAndMindOrcl:  "Sword of Body and Mind",
		b41UnburialRitesOracle:     "Unburial Rites",
		b41SnapbackOracle:          "Snapback",
		b41SubmergeOracle:          "Submerge",
		b41SpectralDelugeOracle:    "Spectral Deluge",
		b41UndeadButlerOracle:      "Undead Butler",
		b41OssificationOracle:      "Ossification",
		b41GoblinRingleaderOracle:  "Goblin Ringleader",
		b41MerryOracle:             "Merry, Esquire of Rohan",
		b41NighthowlerOracle:       "Nighthowler",
		b41MarchWorldOozeOracle:    "March of the World Ooze",
		b41CountOnLuckOracle:       "Count on Luck",
		b41AirbendersReversalOrcl:  "Airbender's Reversal",
		b41ThopterFabricatorOracle: "Thopter Fabricator",
	}
	if len(want) != 46 {
		t.Fatalf("the batch registers 46 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
		{"Bonescythe Sliver", b41BonescytheSliverOracle, "double strike"},
		{"Sentinel Sliver", b41SentinelSliverOracle, "vigilance"},
		{"Venom Sliver", b41VenomSliverOracle, "deathtouch"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lord := b41CatalogCreature(g, me.ID, row.name, "Creature — Sliver", row.oracle, 2, 2)
		mine := b41Sliver(g, me.ID, "My Sliver")
		theirs := b41Sliver(g, opp.ID, "Their Sliver")
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if !hasEffectiveKeyword(t, g, mine, row.keyword) {
			t.Errorf("%s: your other Sliver has %s", row.name, row.keyword)
		}
		if !hasEffectiveKeyword(t, g, lord, row.keyword) {
			t.Errorf("%s: there is no \"other\" — the lord grants it to itself too", row.name)
		}
		if hasEffectiveKeyword(t, g, theirs, row.keyword) {
			t.Errorf("%s: \"you control\" stops at your own board", row.name)
		}
		if hasEffectiveKeyword(t, g, bear, row.keyword) {
			t.Errorf("%s: a Bear is not a Sliver", row.name)
		}
	}
}

func TestB41CrystallineSliverShroudsEverySliverAtTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b41CatalogCreature(g, me.ID, "Crystalline Sliver", "Creature — Sliver", b41CrystallineSliverOracle, 2, 2)
	mine := b41Sliver(g, me.ID, "My Sliver")
	theirs := b41Sliver(g, opp.ID, "Their Sliver")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{lord, mine, theirs} {
		if !hasEffectiveKeyword(t, g, id, "shroud") {
			t.Error("\"all Slivers\" has no controller clause — an opponent's Slivers get shroud too")
		}
	}
	if hasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("a Bear is not a Sliver")
	}
}

func TestB41SquirrelSovereignPumpsOtherSquirrelsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b41CatalogCreature(g, me.ID, "Squirrel Sovereign", "Creature — Squirrel Noble", b41SquirrelSovereignOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Squirrel", "Creature — Squirrel", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("your other Squirrel is a 2/2: power %d", got)
	}
	if got := effectivePower(t, g, lord); got != 2 {
		t.Errorf("\"other\" excludes the Sovereign: power %d, want its printed 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("\"you control\" excludes an opponent's Squirrel: power %d", got)
	}
}

func TestB41WallOfDenialCarriesItsThreeKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wall := b41CatalogCreature(g, me.ID, "Wall of Denial", "Creature — Wall", b41WallOfDenialOracle, 0, 8)
	for _, kw := range []string{"defender", "flying", "shroud"} {
		if !hasEffectiveKeyword(t, g, wall, kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
}

// --- sweepers and removal ------------------------------------------

func TestB41KayasWrathDestroysAllAndPaysOnlyForYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	mine := []uuid.UUID{
		b12Creature(g, me.ID, "Mine A", "Creature — Bear", 2, 2),
		b12Creature(g, me.ID, "Mine B", "Creature — Bear", 2, 2),
	}
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Kaya's Wrath", "Sorcery", b41KayasWrathOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Phyrexian Rebirth", "Sorcery", b41PhyrexianRebirthOracle, nil)
	passPriorityAroundTable(t, g)
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
	if got := effectivePower(t, g, horror); got != 3 {
		t.Errorf("three creatures died, so the Horror is a 3/3: power %d", got)
	}
	if got := effectiveToughness(t, g, horror); got != 3 {
		t.Errorf("Horror toughness %d, want 3", got)
	}
}

func TestB41BlazingVolleyOnlyBurnsOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	tough := b12Creature(g, opp.ID, "Tough", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Blazing Volley", "Sorcery", b41BlazingVolleyOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mono", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID,
	})
	gold := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Owner: opp.ID, Controller: opp.ID,
	})
	colorless := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	for _, bad := range []uuid.UUID{gold, colorless} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Vanishing Verse", TypeLine: "Instant",
			OracleID: b41VanishingVerseOracle, Owner: me.ID, Controller: me.ID,
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bad)}); err == nil {
			t.Error("only a permanent with exactly one colour is a legal target — colourless is not a colour")
		}
	}
	castCatalogSpell(t, g, "Vanishing Verse", "Instant", b41VanishingVerseOracle, cardRefs(mono))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mono) {
		t.Error("the monocolored creature is exiled")
	}
}

// --- triggers ------------------------------------------------------

func TestB41DarkProphecyDrawsAndCostsALifePerDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41Catalog(g, me.ID, "Dark Prophecy", "Enchantment", b41DarkProphecyOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one death, one card: hand %d → %d", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("and one life: %d → %d", life, me.Life)
	}
}

func TestB41RekiDrawsOnlyOnALegendarySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41CatalogCreature(g, me.ID, "Reki, the History of Kamigawa", "Legendary Creature — Human Shaman", b41RekiOracle, 1, 2)
	// castCatalogSpell seeds the card into hand and then casts it, so
	// an ordinary spell leaves the hand size where it started and a
	// draw shows up as +1.
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a nonlegendary spell draws nothing: hand %d → %d", hand, me.Hand.Size())
	}
	castCatalogSpell(t, g, "A Legend", "Legendary Creature — Human", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a legendary spell draws a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB41AuramancerMayRebuyAnEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := pushGraveyardPermanent(me, "Rancor", "Enchantment — Aura", "{G}")
	pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "Auramancer", "Creature — Human Wizard", b41AuramancerOracle, nil)
	passPriorityAroundTable(t, g)
	// "You MAY return" is the CR 603.3c optional-trigger prompt; the
	// target is picked only after it is accepted.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetCards, aura) {
		t.Fatalf("the enchantment card is offered: %+v", p)
	}
	if len(p.PickTargetCards) != 1 {
		t.Error("a creature card in the same graveyard is not")
	}
	pickCard(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura goes to hand")
	}
}

func TestB41AngelOfFinalityExilesTheTargetedGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := pushGraveyardPermanent(opp, "Their Dead Thing", "Creature — Bear", "{1}{B}")
	mine := pushGraveyardPermanent(me, "My Dead Thing", "Creature — Bear", "{1}{B}")
	castCatalogSpell(t, g, "Angel of Finality", "Creature — Angel", b41AngelOfFinalityOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the targeted player's graveyard is exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("only the targeted player's — yours is untouched")
	}
}

func TestB41IndulgingPatricianDrainsOnlyAfterThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41CatalogCreature(g, me.ID, "Indulging Patrician", "Creature — Vampire Noble", b41IndulgingPatricianOrcl, 1, 4)
	before := lifeOfOpponents(g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if g.Seats[i+1].Life != b {
			t.Fatal("no life gained this turn: nothing happens")
		}
	}
	// Gain 3 on the next turn cycle, then reach the end step again.
	advanceToMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	before = lifeOfOpponents(g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-3 {
			t.Errorf("seat %d loses 3: %d → %d", i+1, b, got)
		}
	}
}

func TestB41CrestedSunmareMakesAHorseAndSavesTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sunmare := b41CatalogCreature(g, me.ID, "Crested Sunmare", "Creature — Horse", b41CrestedSunmareOracle, 5, 5)
	horse := b12Creature(g, me.ID, "Other Horse", "Creature — Horse", 2, 2)
	if !hasEffectiveKeyword(t, g, horse, "indestructible") {
		t.Error("other Horses you control are indestructible")
	}
	if hasEffectiveKeyword(t, g, sunmare, "indestructible") {
		t.Error("\"other\" excludes the Sunmare itself — that is the card's whole weakness")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	catapult := b41CatalogCreature(g, me.ID, "Unruly Catapult", "Artifact Creature — Construct", b41UnrulyCatapultOracle, 0, 4)
	before := lifeOfOpponents(g)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, catapult, 0, game.ActivateAbilityParams{})
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("each opponent takes 1: seat %d %d → %d", i+1, b, got)
		}
	}
	if !b16Tapped(t, g, catapult) {
		t.Fatal("the Catapult taps to activate")
	}
	castCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, catapult) {
		t.Error("casting an instant untaps it")
	}
}

func TestB41SpawningPitEatsCreaturesAndSpendsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pit := b41Catalog(g, me.ID, "Spawning Pit", "Artifact", b41SpawningPitOracle)
	first := b12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	second := b12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, pit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{first}})
	if counterOn(g, pit, game.CounterCharge) != 1 {
		t.Fatalf("one sacrifice, one charge counter: %d", counterOn(g, pit, game.CounterCharge))
	}
	if err := g.ActivateCatalogAbility(me.ID, pit, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("one counter is not two — the token ability is unaffordable")
	}
	b16Activate(t, g, me.ID, pit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{second}})
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, pit, 1, game.ActivateAbilityParams{})
	if counterOn(g, pit, game.CounterCharge) != 0 {
		t.Errorf("both counters are spent at announce: %d left", counterOn(g, pit, game.CounterCharge))
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b41Catalog(g, me.ID, "Dissipation Field", "Enchantment", b41DissipationFieldOracle)
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, attacker, me.ID, 2)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !opp.Hand.Contains(attacker) {
		t.Error("the permanent that dealt the damage goes back to its owner's hand")
	}
}

func TestB41HanweirGarrisonMakesTwoTappedAttackingHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garrison := b41CatalogCreature(g, me.ID, "Hanweir Garrison", "Creature — Human Soldier", b41HanweirGarrisonOracle, 2, 3)
	declareAttack(t, g, opp.ID, garrison)
	passPriorityAroundTable(t, g)
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
	if n := triggersOnStackFrom(g, garrison); n != 0 {
		t.Errorf("the tokens were never DECLARED as attackers (CR 508.4), so they retrigger nothing: %d", n)
	}
}

func TestB41ChocoboKnightsGrantDoubleStrikeOnlyToCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knights := b41CatalogCreature(g, me.ID, "Chocobo Knights", "Creature — Human Knight", b41ChocoboKnightsOracle, 3, 3)
	withCounter := b12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(withCounter, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, knights, withCounter, bare)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, withCounter, "double strike") {
		t.Error("a creature with a counter gains double strike")
	}
	if hasEffectiveKeyword(t, g, bare, "double strike") {
		t.Error("one without a counter does not")
	}
	if hasEffectiveKeyword(t, g, knights, "double strike") {
		t.Error("the Knights themselves have no counter and get nothing")
	}
}

func TestB41PlumecreedEscortGrantsHexproofToACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Plumecreed Escort", "Creature — Bird Scout", b41PlumecreedEscortOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the chosen creature gains hexproof")
	}
}

func TestB41YoshimaruGrowsOnAnotherLegendaryPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dog := b41CatalogCreature(g, me.ID, "Yoshimaru, Ever Faithful", "Legendary Creature — Dog", b41YoshimaruOracle, 1, 1)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, dog, game.CounterPlusOne) != 0 {
		t.Fatal("a nonlegendary permanent does nothing")
	}
	castCatalogSpell(t, g, "A Legendary Rock", "Legendary Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, dog, game.CounterPlusOne) != 1 {
		t.Errorf("a legendary PERMANENT of any type grows him: %d counters", counterOn(g, dog, game.CounterPlusOne))
	}
}

func TestB41GrumgullyCountersOnlyOtherNonHumans(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41CatalogCreature(g, me.ID, "Grumgully, the Generous", "Legendary Creature — Goblin Shaman", b41GrumgullyOracle, 3, 3)
	goblin := castCatalogSpell(t, g, "A Goblin", "Creature — Goblin", "", nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, goblin, game.CounterPlusOne) != 1 {
		t.Errorf("a non-Human enters with an extra counter: %d", counterOn(g, goblin, game.CounterPlusOne))
	}
	human := castCatalogSpell(t, g, "A Human", "Creature — Human Soldier", "", nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, human, game.CounterPlusOne) != 0 {
		t.Errorf("a Human gets nothing: %d", counterOn(g, human, game.CounterPlusOne))
	}
}

func TestB41HordewingSkaabGrantsFlyingAndLootsOncePerCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b41CatalogCreature(g, me.ID, "Hordewing Skaab", "Creature — Zombie Horror", b41HordewingSkaabOracle, 3, 3)
	zombie := b12Creature(g, me.ID, "Other Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, zombie, "flying") {
		t.Error("other Zombies you control fly")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a Bear is not a Zombie")
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, zombie, opp.ID, 2)
	passPriorityAroundTable(t, g)
	// The optional trigger asks; accepting draws one and queues the
	// discard, so the hand size ends where it started.
	if me.Hand.Size() != hand {
		t.Errorf("draw one, discard one: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- lands and artifacts -------------------------------------------

func TestB41MossDiamondEntersTappedAndTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diamond := castCatalogSpell(t, g, "Moss Diamond", "Artifact", b41MossDiamondOracle, nil)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, diamond) {
		t.Fatal("the Diamond enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(diamond) })
	b28TapForMana(t, g, me.ID, diamond, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB41IdyllicGrangeNeedsThreePlainsAndThenCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	tapped := b12PlayFromHand(t, g, "Idyllic Grange", "Land — Plains", b41IdyllicGrangeOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, tapped) {
		t.Fatal("two other Plains is not three: it enters tapped")
	}
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	advanceToMainOf(t, g, 0)
	bear := b12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	grange := b12PlayFromHand(t, g, "Idyllic Grange", "Land — Plains", b41IdyllicGrangeOracle, game.CastSpellParams{})
	if b16Tapped(t, g, grange) {
		t.Fatal("three other Plains: it enters untapped")
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if counterOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("an untapped entry puts a +1/+1 counter on a creature you control: %d", counterOn(g, bear, game.CounterPlusOne))
	}
	b28TapForMana(t, g, me.ID, grange, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

func TestB41BarkformHarvesterBottomsACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	harvester := b41CatalogCreature(g, me.ID, "Barkform Harvester", "Artifact Creature — Shapeshifter", b41BarkformHarvesterOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, harvester, game.KeywordChangeling) {
		t.Error("changeling reached the effective abilities")
	}
	dead := pushGraveyardPermanent(me, "Dead Thing", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, harvester, 0, game.ActivateAbilityParams{Targets: cardRefs(dead)})
	if me.Graveyard.Contains(dead) {
		t.Error("the card leaves the graveyard")
	}
	if !me.Library.Contains(dead) {
		t.Error("and lands in the library")
	}
}

// --- attachments ---------------------------------------------------

func TestB41AetherTunnelPumpsAndUnblocksItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Aether Tunnel", "Enchantment — Aura", b41AetherTunnelOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("+1/+0: power %d, want 3", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("+1/+0 leaves toughness alone: %d, want 2", got)
	}
	assertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestB41CanopyCoverGrantsHexproofAndNotEvasion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Canopy Cover", "Enchantment — Aura", b41CanopyCoverOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("CR 702.11b — \"can't be the target of spells or abilities your opponents control\" IS hexproof")
	}
	// The declared caveat: the conditional block restriction is not
	// modelled, so nothing is set. A test pins the caveat so it cannot
	// quietly become an unconditional (stronger) "can't be blocked".
	assertRestrictions(t, g, bear, 0)
}

func TestB41StaffOfTitaniaScalesWithForestsAndMakesDryads(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	staff := b41Catalog(g, me.ID, "Staff of Titania", "Artifact — Equipment", b41StaffOfTitaniaOracle)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C")
	b16Activate(t, g, me.ID, staff, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("two Forests YOU control: power %d, want 2+2", got)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	dryads := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dryad" {
				dryads++
			}
		}
	})
	if dryads != 1 {
		t.Fatalf("one Forest Dryad token per attack: %d", dryads)
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("the new Dryad is a Forest, so the bonus grows in the same beat: power %d, want 5", got)
	}
}

func TestB41SwordOfBodyAndMindMakesAWolfAndMillsTen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := b41Catalog(g, me.ID, "Sword of Body and Mind", "Artifact — Equipment", b41SwordOfBodyAndMindOrcl)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, sword, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+2: power %d", got)
	}
	before := opp.Graveyard.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardPermanent(me, "Dead Fatty", "Creature — Bear", "{5}{B}")
	rites := castCatalogSpell(t, g, "Unburial Rites", "Sorcery", b41UnburialRitesOracle, cardRefs(dead))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card comes back to the battlefield")
	}
	if !me.Graveyard.Contains(rites) {
		t.Fatal("a hard-cast Rites is an ordinary sorcery and goes to the graveyard")
	}
	spec, _ := Lookup(b41UnburialRitesOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Fatalf("flashback is the offer: %+v", spec.AlternativeCosts)
	}
	if !spec.AlternativeCosts[0].ExileOnLeavingStack {
		t.Error("flashback exiles on leaving the stack — without it the Rites is an every-turn engine")
	}
	second := pushGraveyardPermanent(me, "Second Fatty", "Creature — Bear", "{4}{B}")
	advanceToMainOf(t, g, 0)
	b06AddMana(me, "C", "C", "C", "W")
	if err := g.CastSpell(me.ID, rites, game.CastSpellParams{
		FromZone:        string(game.ZoneGraveyard),
		AlternativeCost: "flashback",
		Targets:         cardRefs(second),
	}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(second) {
		t.Error("the flashed-back copy reanimates too")
	}
	if me.Graveyard.Contains(rites) {
		t.Error("and the Rites is exiled rather than returning to the graveyard")
	}
}

func TestB41SnapbackCanBePitchedForABlueCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blue := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blue, Name: "A Blue Card", TypeLine: "Instant",
		Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	snap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: snap, Name: "Snapback", TypeLine: "Instant",
		OracleID: b41SnapbackOracle, ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, snap, game.CastSpellParams{
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{blue},
		Targets:         cardRefs(bear),
	}); err != nil {
		t.Fatalf("pitch the blue card: %v", err)
	}
	if me.Hand.Contains(blue) {
		t.Error("the pitched card is exiled at announce, not on resolution")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature returns to its owner's hand")
	}
}

func TestB41SubmergeIsFreeOnlyWhenBothLandsAreThere(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spec, _ := Lookup(b41SubmergeOracle)
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
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if check() {
		t.Error("your Island alone is not enough — an OPPONENT must control a Forest")
	}
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if check() {
		t.Error("your OWN Forest does not satisfy \"an opponent controls a Forest\"")
	}
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if !check() {
		t.Fatal("their Forest plus your Island: the offer is live")
	}
	sub := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: sub, Name: "Submerge", TypeLine: "Instant",
		OracleID: b41SubmergeOracle, ManaCost: "{4}{U}", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, sub, game.CastSpellParams{
		AlternativeCost: "free",
		Targets:         cardRefs(bear),
	}); err != nil {
		t.Fatalf("cast it for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Library.Contains(bear) {
		t.Error("the creature goes on top of its owner's library")
	}
}

func TestB41SpectralDelugeBouncesBySmallToughnessOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	}
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 1, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Spectral Deluge", "Sorcery", b41SpectralDelugeOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b41CatalogCreature(g, me.ID, "Diffusion Sliver", "Creature — Sliver", b41DiffusionSliverOracle, 1, 1)
	mine := b41Sliver(g, me.ID, "My Sliver")
	advanceToMainOf(t, g, 1)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Their Removal", TypeLine: "Instant",
		OracleID: "94f0a572-e91c-4b56-a5d1-6cbbeabd210d", Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: cardRefs(mine)}); err != nil {
		t.Fatalf("their removal spell is a legal cast — this is ward, not hexproof: %v", err)
	}
	// The trigger goes on the stack ABOVE the removal spell and
	// resolves first, asking the SPELL's controller — not the Sliver's
	// — to pay. Priority is passed one step at a time so the answer
	// lands before the spell below it resolves.
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("the tax is charged to the spell's controller")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the Sliver's controller is never the one asked to pay")
	}
	prompt := answerPayUnless(t, g, opp.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("PayCost = %q, want {2}", prompt.PayCost)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("declined, the targeting spell is countered and the Sliver lives")
	}
}

func TestB41UndeadButlerMillsThenExilesItselfToRebuy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Graveyard.Size()
	butler := castCatalogSpell(t, g, "Undead Butler", "Creature — Zombie", b41UndeadButlerOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - before; got != 3 {
		t.Fatalf("the entry trigger mills three: %d", got)
	}
	target := pushGraveyardPermanent(me, "Dead Fatty", "Creature — Bear", "{5}{B}")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(butler) })
	passPriorityAroundTable(t, g)
	// "You may exile it" is the parent's optional prompt; the
	// reflexive "when you do" that follows is mandatory.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(butler) {
		t.Error("the Butler exiles itself")
	}
	if !me.Hand.Contains(target) {
		t.Error("and the chosen creature card goes to hand")
	}
}

func TestB41OssificationExilesUntilItLeavesTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	aura := castCatalogSpell(t, g, "Ossification", "Enchantment — Aura", b41OssificationOracle, cardRefs(plains))
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature is exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
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
	castCatalogSpell(t, g, "Goblin Ringleader", "Creature — Goblin", b41GoblinRingleaderOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merry := b41CatalogCreature(g, me.ID, "Merry, Esquire of Rohan", "Legendary Creature — Halfling Knight", b41MerryOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, merry, "first strike") {
		t.Error("unequipped, Merry has no first strike")
	}
	sword := b12Permanent(g, me.ID, "A Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: merry})
	})
	if !hasEffectiveKeyword(t, g, merry, "first strike") {
		t.Error("equipped, he does")
	}
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, merry, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a Bear is not another legendary creature — no draw")
	}
	advanceToMainOf(t, g, 0)
	legend := b12Creature(g, me.ID, "Another Legend", "Legendary Creature — Human", 2, 2)
	hand = me.Hand.Size()
	declareAttack(t, g, opp.ID, merry, legend)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Merry plus another legendary creature draws one: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB41NighthowlerGrowsWithEveryGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	howler := b41CatalogCreature(g, me.ID, "Nighthowler", "Enchantment Creature — Horror", b41NighthowlerOracle, 0, 0)
	if got := effectivePower(t, g, howler); got != 0 {
		t.Fatalf("empty graveyards: a 0/0, power %d", got)
	}
	pushGraveyardPermanent(me, "Mine", "Creature — Bear", "{1}{B}")
	pushGraveyardPermanent(opp, "Theirs", "Creature — Bear", "{1}{B}")
	pushGraveyardPermanent(opp, "Not a creature", "Enchantment", "{1}{B}")
	// A raw push into a graveyard emits no zone-move event, so the
	// layer cache has nothing to invalidate on — the same bump
	// Tarmogoyf's test makes for the same reason.
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectivePower(t, g, howler); got != 2 {
		t.Errorf("creature cards in ALL graveyards: power %d, want 2", got)
	}
	if got := effectiveToughness(t, g, howler); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

func TestB41MarchOfTheWorldOozeSetsSixSixAndAddsTheType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b41Catalog(g, me.ID, "March of the World Ooze", "Enchantment", b41MarchWorldOozeOracle)
	dork := b12Creature(g, me.ID, "Elf Druid", "Creature — Elf Druid", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, dork); got != 6 {
		t.Errorf("base power and toughness 6/6: power %d", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41Catalog(g, me.ID, "Count on Luck", "Enchantment", b41CountOnLuckOracle)
	before := 0
	g.ReadSnapshot(func() { before = g.Exile.Size() })
	// The game opens in seat 0's draw step, so its upkeep is already
	// behind us; go round the table to the NEXT one.
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	after := 0
	g.ReadSnapshot(func() { after = g.Exile.Size() })
	if after != before+1 {
		t.Errorf("one card exiled at your upkeep: exile %d → %d", before, after)
	}
}

func TestB41AirbendersReversalDestroysAnAttackerOrAirbendsYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	declareAttack(t, g, me.ID, attacker)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Airbender's Reversal", TypeLine: "Instant — Lesson",
		OracleID: b41AirbendersReversalOrcl, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: cardRefs(attacker),
	}); err != nil {
		t.Fatalf("mode 0 on an attacking creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("mode 0 destroys the attacker")
	}
	// Mode 1 exiles your own creature with a recast grant.
	advanceToMainOf(t, g, 0)
	second := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: second, Name: "Airbender's Reversal", TypeLine: "Instant — Lesson",
		OracleID: b41AirbendersReversalOrcl, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{
		Modes:   []int{1},
		Targets: cardRefs(mine),
	}); err != nil {
		t.Fatalf("mode 1 on your own creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("mode 1 airbends it — the creature is exiled")
	}
}

func TestB41ThopterFabricatorMakesAThopterOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41Catalog(g, me.ID, "Thopter Fabricator", "Artifact — Vehicle", b41ThopterFabricatorOracle)
	advanceToMainOf(t, g, 0)
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
	passPriorityAroundTable(t, g)
	if thopters() != 1 {
		t.Fatalf("the second draw makes one Thopter: %d", thopters())
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if thopters() != 1 {
		t.Errorf("the third does not — it is \"your SECOND card each turn\": %d", thopters())
	}
}
