package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m20_5518_b1_cards_test.go — Core Set 2020, batch 1 (#5518). Every test
// plays the real catalog entry through the engine.

const (
	m20ImperialOutriderOracle     = "76cd10a0-36af-4f56-87c1-f72dee07b27c"
	m20MoatPiranhasOracle         = "40057917-4a88-4316-bc5b-efc0ada1ddff"
	m20NimbleBirdstickerOracle    = "2f047fcb-c198-44ac-aa1e-31447f6c6aba"
	m20LightningStormkinOracle    = "126844f5-3a95-48cb-80f2-d7e4b0005350"
	m20WoodlandMysticOracle       = "6996c354-d3fa-40c3-8534-27d2f5b74756"
	m20DaggersailAeronautOracle   = "7c0ef3b7-b368-4a5b-8cbe-bf1f47610f74"
	m20DawningAngelOracle         = "a7d88510-d2ce-4dd6-938c-964cc74ff7d7"
	m20HealerOfTheGladeOracle     = "2701b912-b672-4764-97f7-51c444f754f2"
	m20SavannahSageOracle         = "7e6cfec7-3b54-43db-853a-27c164776a5f"
	m20FerociousPupOracle         = "a1112750-8d38-49e0-ab1c-77110b2bbc4d"
	m20HowlingGiantOracle         = "c616242a-95f0-4cd8-99a1-9237cfea0e57"
	m20OctoprophetOracle          = "af6cfc23-b64b-4c8a-94b4-69d2b01cb3b9"
	m20SilverbackShamanOracle     = "fc011947-b496-400e-99b8-b368068ba79b"
	m20ScholarOfTheAgesOracle     = "1bb1ad32-9ce5-4f14-a408-e142ff30c679"
	m20BattalionFootSoldierOracle = "f6f39a85-28b9-49f3-ab90-2ff99169de42"
	m20EtherealElkOracle          = "82d643ef-8380-4fdc-a88d-ed205324dbd9"
	m20SorinsGuideOracle          = "36f5f161-b25e-48e4-b8f9-e6c36dfefe5b"
	m20YaroksFenlurkerOracle      = "6bb60f1f-9295-41ce-bdcd-3fe67f8e175c"
	m20GruesomeScourgerOracle     = "1b2993fa-b212-4eaa-b3e8-f79c2f2e4cbe"
	m20TomeboundLichOracle        = "6d243164-e21d-46c5-943e-9a2eb759c742"
	m20ScamperingScorcherOracle   = "4899595e-e626-432c-9c9f-f9d5c243a64b"
	m20SavageGorgerOracle         = "307d20fd-a9c9-4da9-b990-26c653054640"
	m20RottingRegisaurOracle      = "acdbaa99-18c9-4aa1-8797-69837ca7d8f5"
	m20ScorchSpitterOracle        = "4cdc2120-7051-45b8-b877-87ffaf396ee7"
	m20LavakinBrawlerOracle       = "fa4ad622-5021-44fe-92af-0de115059eb6"
	m20SorcererOfTheFangOracle    = "c1d5a631-a963-43a4-a5c0-1ba23bd88d0f"
	m20PyroclasticElementalOracle = "1f37e3a6-ce3a-472f-ac45-0d9b23a31a11"
	m20OgreSiegebreakerOracle     = "1432db6a-9466-48f6-9011-60728c437608"
	m20GoblinBirdGrabberOracle    = "780f8be6-d477-43ca-92ba-0fe480bbd54d"
	m20BrightwoodTrackerOracle    = "f8d44858-1d21-43fd-b193-59ddf7c53785"
	m20ApostleOracle              = "0b49dbdc-4f1d-4895-b263-a48c7b7e6b44"
	m20ManifoldKeyOracle          = "370c37f7-2484-40e8-8fd9-c10cf6f5d5f2"
	m20RetributiveWandOracle      = "b550a94d-3a92-416e-b651-510cdbdd05f1"
	m20UnchainedBerserkerOracle   = "32e7befb-c083-4e17-aaeb-7247f317bbcf"
	m20ThicketCrasherOracle       = "998e850a-b3c7-4ab2-85ac-25beef98ab20"
	m20CreepingTrailblazerOracle  = "94431642-cfcf-48f4-9e0b-fcddc5b051ef"
	m20CelestialMessengerOracle   = "cf40caa2-cd5b-4372-ad4a-7f4cf564241e"
	m20ViviensCrocodileOracle     = "b49791f9-525c-4b04-954d-d7678c5565e3"
	m20DarkRemedyOracle           = "d7cc744b-2fbd-40e6-8a59-392535c17454"
	m20EngulfingEruptionOracle    = "27b4fee0-ae46-4d15-96cd-aff8f3009de8"
	m20CaptivatingGyreOracle      = "41011f09-ad8a-4eea-aa52-d20a219bc94c"
	m20EternalIsolationOracle     = "8f69a103-d557-4d28-847e-aa94a440e908"
	m20NoxiousGraspOracle         = "b17f4960-9603-4a81-8dc8-e09e8b822357"
	m20FryOracle                  = "3a747893-a514-4689-8486-ecccf654fecb"
	m20RecklessAirStrikeOracle    = "89247308-e16f-4f60-a2a2-0b5e297c1fc1"
	m20ChandraNoviceOracle        = "882619f7-69b4-4ce0-be4e-1e34e608f925"

	m20BearType      = "Creature — Bear"
	m20ElementalType = "Creature — Elemental"
)

// m20Planeswalker puts a planeswalker with the given subtype and loyalty
// under `owner`.
func m20Planeswalker(g *game.Game, owner uuid.UUID, name, subtype string, loyalty int) uuid.UUID {
	id := b12Permanent(g, owner, name, "Legendary Planeswalker — "+subtype)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}

// m20Tap taps a battlefield permanent.
func m20Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
}

// m20Flyer puts a creature with flying under `owner`.
func m20Flyer(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: power, Toughness: toughness, Keywords: []string{"flying"}})
}

// m20NamedYouControl lists the permanents named `name` that `controller`
// controls.
func m20NamedYouControl(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// --- every card is in the catalog ----------------------------------------

func TestM20B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m20ImperialOutriderOracle, m20MoatPiranhasOracle, m20NimbleBirdstickerOracle, m20LightningStormkinOracle,
		m20WoodlandMysticOracle, m20DaggersailAeronautOracle, m20DawningAngelOracle, m20HealerOfTheGladeOracle,
		m20SavannahSageOracle, m20FerociousPupOracle, m20HowlingGiantOracle, m20OctoprophetOracle,
		m20SilverbackShamanOracle, m20ScholarOfTheAgesOracle, m20BattalionFootSoldierOracle, m20EtherealElkOracle,
		m20SorinsGuideOracle, m20YaroksFenlurkerOracle, m20GruesomeScourgerOracle, m20TomeboundLichOracle,
		m20ScamperingScorcherOracle, m20SavageGorgerOracle, m20RottingRegisaurOracle, m20ScorchSpitterOracle,
		m20LavakinBrawlerOracle, m20SorcererOfTheFangOracle, m20PyroclasticElementalOracle, m20OgreSiegebreakerOracle,
		m20GoblinBirdGrabberOracle, m20BrightwoodTrackerOracle, m20ApostleOracle, m20ManifoldKeyOracle,
		m20RetributiveWandOracle, m20UnchainedBerserkerOracle, m20ThicketCrasherOracle, m20CreepingTrailblazerOracle,
		m20CelestialMessengerOracle, m20ViviensCrocodileOracle, m20DarkRemedyOracle, m20EngulfingEruptionOracle,
		m20CaptivatingGyreOracle, m20EternalIsolationOracle, m20NoxiousGraspOracle, m20FryOracle,
		m20RecklessAirStrikeOracle, m20ChandraNoviceOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -----------------------------------------------------

func TestM20B1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Moat Piranhas", "Creature — Fish", m20MoatPiranhasOracle, []string{"defender"}},
		{"Nimble Birdsticker", "Creature — Goblin", m20NimbleBirdstickerOracle, []string{"reach"}},
		{"Lightning Stormkin", "Creature — Elemental Wizard", m20LightningStormkinOracle, []string{"flying", "haste"}},
		{"Dawning Angel", "Creature — Angel", m20DawningAngelOracle, []string{"flying"}},
		{"Howling Giant", "Creature — Giant Druid", m20HowlingGiantOracle, []string{"reach"}},
		{"Silverback Shaman", "Creature — Ape Shaman", m20SilverbackShamanOracle, []string{"trample"}},
		{"Ethereal Elk", "Creature — Elk Spirit", m20EtherealElkOracle, []string{"trample"}},
		{"Tomebound Lich", "Creature — Zombie Wizard", m20TomeboundLichOracle, []string{"deathtouch", "lifelink"}},
		{"Savage Gorger", "Creature — Vampire", m20SavageGorgerOracle, []string{"flying"}},
		{"Apostle of Purifying Light", "Creature — Human Cleric", m20ApostleOracle, []string{"protection from black"}},
		{"Unchained Berserker", "Creature — Human Berserker", m20UnchainedBerserkerOracle, []string{"protection from white"}},
		{"Thicket Crasher", "Creature — Elemental Rhino", m20ThicketCrasherOracle, []string{"trample"}},
		{"Celestial Messenger", "Creature — Bird Spirit", m20CelestialMessengerOracle, []string{"flash", "flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestM20B1ImperialOutriderIsVanilla(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Imperial Outrider", "Creature — Human Knight", m20ImperialOutriderOracle, 1, 5)
	if p, tough := ptOf(t, g, id); p != 1 || tough != 5 {
		t.Errorf("Imperial Outrider is %d/%d, want 1/5", p, tough)
	}
}

// Daggersail Aeronaut flies only during its controller's turn.
func TestM20B1DaggersailAeronautFliesOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := b12Push(g, me.ID, "Daggersail Aeronaut", "Creature — Goblin", m20DaggersailAeronautOracle, 3, 2)
	theirs := b12Push(g, opp.ID, "Daggersail Aeronaut", "Creature — Goblin", m20DaggersailAeronautOracle, 3, 2)
	if !hasKeywordOnBattlefield(t, g, mine, "flying") {
		t.Error("the active player's Aeronaut has no flying")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("an Aeronaut has flying on another player's turn")
	}
}

func TestM20B1WoodlandMysticTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Woodland Mystic", "Creature — Elf Druid", m20WoodlandMysticOracle, false)
	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "G" {
		t.Errorf("mana pool %+v, want 1xG", me.ManaPool)
	}
}

// --- enters and dies triggers --------------------------------------------

func TestM20B1EntersAndGainsLife(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		gain                   int
	}{
		{"Dawning Angel", "Creature — Angel", m20DawningAngelOracle, 4},
		{"Healer of the Glade", m20ElementalType, m20HealerOfTheGladeOracle, 3},
		{"Savannah Sage", "Creature — Cat Cleric", m20SavannahSageOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			passPriorityAroundTable(t, g)
			if me.Life != life+tc.gain {
				t.Errorf("life %d, want %d", me.Life, life+tc.gain)
			}
		})
	}
}

func TestM20B1EntersAndMakesWolves(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		wolves                 int
	}{
		{"Ferocious Pup", "Creature — Wolf", m20FerociousPupOracle, 1},
		{"Howling Giant", "Creature — Giant Druid", m20HowlingGiantOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			passPriorityAroundTable(t, g)
			if got := battlefieldNamed(g, me.ID, "Wolf"); got != tc.wolves {
				t.Errorf("%d Wolf tokens, want %d", got, tc.wolves)
			}
		})
	}
}

func TestM20B1OctoprophetScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Octoprophet", "Creature — Octopus", m20OctoprophetOracle, 3, 3)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry was asked")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(scry.ScryCards))
	}
}

func TestM20B1SilverbackShamanDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ape := b12Push(g, me.ID, "Silverback Shaman", "Creature — Ape Shaman", m20SilverbackShamanOracle, 5, 4)
	hand := me.Hand.Size()
	b27Kill(g, ape)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestM20B1ScholarOfTheAgesReturnsTwoSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	instant := wthGraveyardCard(me, "Old Instant", "Instant")
	sorcery := wthGraveyardCard(me, "Old Sorcery", "Sorcery")
	creature := wthGraveyardCard(me, "Old Bear", m20BearType)
	castHoldCreature(t, g, "Scholar of the Ages", "Creature — Human Wizard", m20ScholarOfTheAgesOracle, 3, 3)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: creature}}); err == nil {
		t.Fatal("a creature card was accepted as a target")
	}
	answerTriggerTargets(t, g, me.ID, instant, sorcery)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(instant) || !me.Hand.Contains(sorcery) {
		t.Error("the instant and sorcery cards did not return to hand")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature card left the graveyard")
	}
}

func TestM20B1BattalionFootSoldierFetchesItsNamesakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := trekB1Library(g, me, "Battalion Foot Soldier", "Creature — Human Soldier", m20BattalionFootSoldierOracle)
	b := trekB1Library(g, me, "Battalion Foot Soldier", "Creature — Human Soldier", m20BattalionFootSoldierOracle)
	trekB1Library(g, me, "Other Soldier", "Creature — Human Soldier", "")
	castHoldCreature(t, g, "Battalion Foot Soldier", "Creature — Human Soldier", m20BattalionFootSoldierOracle, 2, 2)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search was offered when the Foot Soldier entered")
	}
	if searchOptionNamed(g, c, "Other Soldier") != uuid.Nil {
		t.Error("the search offered a card with another name")
	}
	answerSearchByID(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, a) != game.ZoneHand || b12ZoneOf(g, b) != game.ZoneHand {
		t.Errorf("the namesakes are in %v and %v, want both in hand", b12ZoneOf(g, a), b12ZoneOf(g, b))
	}
}

func TestM20B1PartnerSearchFetchesFromTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, fetches string
	}{
		{"Ethereal Elk", m20EtherealElkOracle, "Creature — Elk Spirit", "Vivien, Nature's Avenger"},
		{"Sorin's Guide", m20SorinsGuideOracle, "Creature — Vampire", "Sorin, Vampire Lord"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			lgKnowLibrary(me)
			pw := pushGraveyardCardWithTypeLineID(me, tc.fetches, "Legendary Planeswalker")
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, true)
			answerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			answerSearchByID(t, g, me.ID, pw)
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the graveyard", tc.fetches)
			}
		})
	}
}

func TestM20B1YaroksFenlurkerMakesEachOpponentExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	d2dSetHand(opp, 2)
	d2dSetHand(other, 2)
	castHoldCreature(t, g, "Yarok's Fenlurker", "Creature — Horror", m20YaroksFenlurkerOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, opp.ID) == nil || chooseCardsChoiceFor(g, other.ID) == nil {
		t.Fatal("an opponent with cards in hand was not asked to exile one")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the Fenlurker's controller was asked to exile a card")
	}
}

func TestM20B1YaroksFenlurkerPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Yarok's Fenlurker", "Creature — Horror", m20YaroksFenlurkerOracle, 1, 1)
	apaMana(me, "B", "C", "C")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, id); p != 2 || tough != 2 {
		t.Errorf("the Fenlurker is %d/%d, want 2/2", p, tough)
	}
}

func TestM20B1GruesomeScourgerCountsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	b12Creature(g, opp.ID, "Their Bear", m20BearType, 2, 2)
	life := opp.Life
	castHoldCreature(t, g, "Gruesome Scourger", "Creature — Orc Warrior", m20GruesomeScourgerOracle, 3, 3)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
}

func TestM20B1TomeboundLichLootsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Tomebound Lich", "Creature — Zombie Wizard", m20TomeboundLichOracle, 1, 3)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestM20B1TomeboundLichLootsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	lich := b12Push(g, me.ID, "Tomebound Lich", "Creature — Zombie Wizard", m20TomeboundLichOracle, 1, 3)
	hand := me.Hand.Size()
	dealCombatDamage(g, me.ID, lich, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestM20B1ScamperingScorcherMakesHastyElementals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Scampering Scorcher", m20ElementalType, m20ScamperingScorcherOracle, 1, 1)
	passPriorityAroundTable(t, g)
	tokens := m20NamedYouControl(g, me.ID, "Elemental")
	if len(tokens) != 2 {
		t.Fatalf("%d Elemental tokens, want 2", len(tokens))
	}
	for _, id := range tokens {
		if !hasKeywordOnBattlefield(t, g, id, "haste") {
			t.Error("an Elemental token has no haste")
		}
	}
}

// --- other triggers ---------------------------------------------------------

func TestM20B1SavageGorgerGrowsOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		gorger := b12Push(g, me.ID, "Savage Gorger", "Creature — Vampire", m20SavageGorgerOracle, 1, 1)
		advanceToMain(t, g)
		if lost {
			g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		want := 0
		if lost {
			want = 1
		}
		if got := countersOn(g, gorger, game.CounterPlusOne); got != want {
			t.Errorf("opponent lost life %v: %d counters, want %d", lost, got, want)
		}
	}
}

func TestM20B1RottingRegisaurDiscardsEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Rotting Regisaur", "Creature — Zombie Dinosaur", m20RottingRegisaurOracle, 7, 6)
	walkToUpkeepOfSeat0(t, g)
	d2dSetHand(me, 2)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestM20B1ScorchSpitterPingsTheDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	spitter := b12Push(g, me.ID, "Scorch Spitter", "Creature — Elemental Lizard", m20ScorchSpitterOracle, 1, 1)
	oppLife, otherLife := lifeOf(g, opp.ID), lifeOf(g, other.ID)
	declareAttack(t, g, opp.ID, spitter)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("defending player at %d, want %d", got, oppLife-1)
	}
	if got := lifeOf(g, other.ID); got != otherLife {
		t.Errorf("a bystander lost life: %d", got)
	}
}

func TestM20B1LavakinBrawlerCountsElementals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	brawler := b12Push(g, me.ID, "Lavakin Brawler", "Creature — Elemental Warrior", m20LavakinBrawlerOracle, 2, 4)
	b12Creature(g, me.ID, "Spark", m20ElementalType, 1, 1)
	b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	b12Creature(g, opp.ID, "Their Spark", m20ElementalType, 1, 1)
	declareAttack(t, g, opp.ID, brawler)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, brawler); p != 4 {
		t.Errorf("the Brawler has power %d, want 4", p)
	}
}

// --- activated abilities ----------------------------------------------------

func TestM20B1SorcererOfTheFangHitsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	id := b12Push(g, me.ID, "Sorcerer of the Fang", "Creature — Human Wizard Sorcerer", m20SorcererOfTheFangOracle, 1, 3)
	apaMana(me, "B", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{playerTarget(me.ID)}}); err == nil {
		t.Fatal("the Sorcerer targeted its controller")
	}
	life := opp.Life
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	if !hmlTapped(t, g, id) {
		t.Error("the Sorcerer did not tap")
	}
}

func TestM20B1PyroclasticElementalPingsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	id := b12Push(g, me.ID, "Pyroclastic Elemental", m20ElementalType, m20PyroclasticElementalOracle, 5, 4)
	bear := b12Creature(g, opp.ID, "Bear", m20BearType, 2, 2)
	apaMana(me, "R", "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("the Elemental targeted a creature")
	}
	life := opp.Life
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestM20B1OgreSiegebreakerDestroysADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	ogre := b12Push(g, me.ID, "Ogre Siegebreaker", "Creature — Ogre Berserker", m20OgreSiegebreakerOracle, 4, 3)
	fresh := b12Creature(g, opp.ID, "Fresh", m20BearType, 2, 2)
	hurt := b12Creature(g, opp.ID, "Hurt", m20BearType, 2, 3)
	b27Damage(g, ogre, hurt, 1)
	g.RunStateChecksForTest()
	apaMana(me, "B", "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ogre, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(fresh)}); err == nil {
		t.Fatal("the Siegebreaker targeted a creature dealt no damage this turn")
	}
	apaActivate(t, g, me, ogre, 0, game.ActivateAbilityParams{Targets: cardTarget(hurt)})
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
}

func TestM20B1GoblinBirdGrabberNeedsAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	grabber := b12Push(g, me.ID, "Goblin Bird-Grabber", "Creature — Goblin", m20GoblinBirdGrabberOracle, 2, 1)
	apaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, grabber, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a creature with flying")
	}
	m20Flyer(g, me.ID, 1, 1)
	apaActivate(t, g, me, grabber, 0, game.ActivateAbilityParams{})
	if !hasKeywordOnBattlefield(t, g, grabber, "flying") {
		t.Error("the Bird-Grabber has no flying")
	}
}

func TestM20B1BrightwoodTrackerTakesACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	tracker := b12Push(g, me.ID, "Brightwood Tracker", "Creature — Elf Scout", m20BrightwoodTrackerOracle, 2, 4)
	beast := uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Land", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery"},
		{InstanceID: beast, Name: "Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5},
		{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	apaMana(me, "G", "C", "C", "C", "C", "C")
	apaActivate(t, g, me, tracker, 0, game.ActivateAbilityParams{})
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	answerChooseCards(t, g, me.ID, beast)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(beast) {
		t.Error("the creature card is not in hand")
	}
	if !hmlTapped(t, g, tracker) {
		t.Error("the Tracker did not tap")
	}
}

func TestM20B1ApostleExilesAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	apostle := b12Push(g, me.ID, "Apostle of Purifying Light", "Creature — Human Cleric", m20ApostleOracle, 2, 1)
	dead := pushGraveyardCardForTest(opp, "Dead")
	apaMana(me, "C", "C")
	apaActivate(t, g, me, apostle, 0, game.ActivateAbilityParams{Targets: cardTarget(dead)})
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
}

func TestM20B1ManifoldKeyUntapsAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	key := pushCatalogPermanent(g, me.ID, "Manifold Key", "Artifact", m20ManifoldKeyOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	m20Tap(g, rock)
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(key)}); err == nil {
		t.Fatal("the Key targeted itself")
	}
	apaActivate(t, g, me, key, 0, game.ActivateAbilityParams{Targets: cardTarget(rock)})
	if hmlTapped(t, g, rock) {
		t.Error("the other artifact is still tapped")
	}
	if !hmlTapped(t, g, key) {
		t.Error("the Key did not tap")
	}
}

func TestM20B1ManifoldKeyMakesACreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	key := pushCatalogPermanent(g, me.ID, "Manifold Key", "Artifact", m20ManifoldKeyOracle, false)
	raider := b12Creature(g, me.ID, "Raider", m20BearType, 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", m20BearType, 2, 2)
	apaMana(me, "C", "C", "C")
	apaActivate(t, g, me, key, 1, game.ActivateAbilityParams{Targets: cardTarget(raider)})
	brAttack(t, g, raider)
	if err := g.DeclareBlocker(blocker, raider); err == nil {
		t.Error("a creature blocked the unblockable attacker")
	}
}

func TestM20B1RetributiveWandPingsAndBurnsOnTheWayOut(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	wand := pushCatalogPermanent(g, me.ID, "Retributive Wand", "Artifact", m20RetributiveWandOracle, false)
	life := opp.Life
	apaMana(me, "C", "C", "C")
	apaActivate(t, g, me, wand, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d after the ping, want %d", opp.Life, life-1)
	}
	b27Kill(g, wand)
	g.RunStateChecksForTest()
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("the opponent is at %d after the Wand died, want %d", opp.Life, life-6)
	}
}

// --- statics ----------------------------------------------------------------

func TestM20B1UnchainedBerserkerPumpsWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	berserker := b12Push(g, me.ID, "Unchained Berserker", "Creature — Human Berserker", m20UnchainedBerserkerOracle, 1, 1)
	if p, _ := ptOf(t, g, berserker); p != 1 {
		t.Fatalf("power %d at rest, want 1", p)
	}
	brAttack(t, g, berserker)
	if p, tough := ptOf(t, g, berserker); p != 3 || tough != 1 {
		t.Errorf("attacking Berserker is %d/%d, want 3/1", p, tough)
	}
}

func TestM20B1ThicketCrasherGivesOtherElementalsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	b12Push(g, me.ID, "Thicket Crasher", "Creature — Elemental Rhino", m20ThicketCrasherOracle, 4, 3)
	spark := b12Creature(g, me.ID, "Spark", m20ElementalType, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Spark", m20ElementalType, 1, 1)
	if !hasKeywordOnBattlefield(t, g, spark, "trample") {
		t.Error("another Elemental you control has no trample")
	}
	if hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("a non-Elemental has trample")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's Elemental has trample")
	}
}

func TestM20B1CreepingTrailblazerAnthemAndPump(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	blazer := b12Push(g, me.ID, "Creeping Trailblazer", m20ElementalType, m20CreepingTrailblazerOracle, 2, 2)
	spark := b12Creature(g, me.ID, "Spark", m20ElementalType, 1, 1)
	if p, tough := ptOf(t, g, spark); p != 2 || tough != 1 {
		t.Errorf("the other Elemental is %d/%d, want 2/1", p, tough)
	}
	if p, _ := ptOf(t, g, blazer); p != 2 {
		t.Errorf("the Trailblazer pumps itself: power %d, want 2", p)
	}
	apaMana(me, "R", "G", "C", "C")
	apaActivate(t, g, me, blazer, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, blazer); p != 4 || tough != 4 {
		t.Errorf("the Trailblazer is %d/%d with two Elementals, want 4/4", p, tough)
	}
}

func TestM20B1PlaneswalkerBuddies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, subtype string
		power, toughness                int
	}{
		{"Celestial Messenger", "Creature — Bird Spirit", m20CelestialMessengerOracle, "Yanling", 3, 2},
		{"Vivien's Crocodile", "Creature — Crocodile Spirit", m20ViviensCrocodileOracle, "Vivien", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			m20Planeswalker(g, opp.ID, "Their Walker", tc.subtype, 3)
			m20Planeswalker(g, me.ID, "Other Walker", "Gideon", 3)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("%d/%d without your %s, want %d/%d", p, tough, tc.subtype, tc.power, tc.toughness)
			}
			m20Planeswalker(g, me.ID, "My Walker", tc.subtype, 3)
			if p, tough := ptOf(t, g, id); p != tc.power+1 || tough != tc.toughness+1 {
				t.Errorf("%d/%d with your %s, want %d/%d", p, tough, tc.subtype, tc.power+1, tc.toughness+1)
			}
		})
	}
}

// --- spells -------------------------------------------------------------------

func TestM20B1DarkRemedyPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	castCatalogSpell(t, g, "Dark Remedy", "Instant", m20DarkRemedyOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 3/5", p, tough)
	}
}

func TestM20B1EngulfingEruptionDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 7)
	castCatalogSpell(t, g, "Engulfing Eruption", "Sorcery", m20EngulfingEruptionOracle, cardTarget(wall))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, wall).DamageMarked; got != 5 {
		t.Errorf("the creature has %d damage, want 5", got)
	}
}

func TestM20B1CaptivatingGyreBouncesThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	a := b12Creature(g, opp.ID, "A", m20BearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", m20BearType, 2, 2)
	c := b12Creature(g, me.ID, "C", m20BearType, 2, 2)
	targets := append(append(cardTarget(a), cardTarget(b)...), cardTarget(c)...)
	castCatalogSpell(t, g, "Captivating Gyre", "Sorcery", m20CaptivatingGyreOracle, targets)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the opponent's creatures did not return to their owner's hand")
	}
	if !me.Hand.Contains(c) {
		t.Error("your creature did not return to your hand")
	}
}

func TestM20B1EternalIsolationTucksAPowerfulCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	small := b12Creature(g, opp.ID, "Small", m20BearType, 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	if err := castCatalogSpellErr(t, g, "Eternal Isolation", "Sorcery", m20EternalIsolationOracle, cardTarget(small)); err == nil {
		t.Fatal("Eternal Isolation targeted a creature with power 3")
	}
	castCatalogSpell(t, g, "Eternal Isolation", "Sorcery", m20EternalIsolationOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !opp.Library.Contains(big) {
		t.Fatal("the creature was not put into its owner's library")
	}
	if bottom := opp.Library.Cards[0].InstanceID; bottom != big {
		t.Error("the creature is not on the bottom of the library")
	}
}

func TestM20B1NoxiousGraspDestroysGreenOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	blue := b12Creature(g, opp.ID, "Blue", "Creature — Merfolk", 2, 2)
	hmlSetColors(g, blue, "U")
	green := b12Creature(g, opp.ID, "Green", m20BearType, 2, 2)
	hmlSetColors(g, green, "G")
	if err := castCatalogSpellErr(t, g, "Noxious Grasp", "Instant", m20NoxiousGraspOracle, cardTarget(blue)); err == nil {
		t.Fatal("Noxious Grasp targeted a blue creature")
	}
	life := me.Life
	castCatalogSpell(t, g, "Noxious Grasp", "Instant", m20NoxiousGraspOracle, cardTarget(green))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the green creature survived")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestM20B1FryBurnsWhiteOrBlue(t *testing.T) {
	spec, ok := Lookup(m20FryOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Fry is not registered as uncounterable")
	}
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	red := b12Creature(g, opp.ID, "Red", "Creature — Goblin", 1, 7)
	hmlSetColors(g, red, "R")
	white := b12Creature(g, opp.ID, "White", "Creature — Wall", 0, 7)
	hmlSetColors(g, white, "W")
	if err := castCatalogSpellErr(t, g, "Fry", "Instant", m20FryOracle, cardTarget(red)); err == nil {
		t.Fatal("Fry targeted a red creature")
	}
	castCatalogSpell(t, g, "Fry", "Instant", m20FryOracle, cardTarget(white))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, white).DamageMarked; got != 5 {
		t.Errorf("the white creature has %d damage, want 5", got)
	}
}

func TestM20B1RecklessAirStrikeEachMode(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	walker := b12Creature(g, opp.ID, "Walker", m20BearType, 5, 5)
	flyer := m20Flyer(g, opp.ID, 5, 5)
	if err := wthCastWithModes(t, g, "Reckless Air Strike", "Sorcery", m20RecklessAirStrikeOracle, []int{0}, cardTarget(walker)); err == nil {
		t.Fatal("the damage bullet targeted a creature without flying")
	}
	castModal(t, g, "Reckless Air Strike", "Sorcery", m20RecklessAirStrikeOracle, []int{0}, cardTarget(flyer))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, flyer).DamageMarked; got != 3 {
		t.Errorf("the flyer has %d damage, want 3", got)
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castModal(t, g, "Reckless Air Strike", "Sorcery", m20RecklessAirStrikeOracle, []int{1}, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

// --- planeswalker ---------------------------------------------------------------

func m20ChandraTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Novice Pyromancer", "Legendary Planeswalker — Chandra", m20ChandraNoviceOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 5) })
	advanceToMain(t, g)
	return g, me, opp, chandra
}

func TestM20B1ChandraNovicePumpsElementals(t *testing.T) {
	g, me, _, chandra := m20ChandraTable(t)
	spark := b12Creature(g, me.ID, "Spark", m20ElementalType, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", m20BearType, 2, 2)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	if p, _ := ptOf(t, g, spark); p != 3 {
		t.Errorf("the Elemental has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the non-Elemental has power %d, want 2", p)
	}
}

func TestM20B1ChandraNoviceAddsTwoRed(t *testing.T) {
	g, me, _, chandra := m20ChandraTable(t)
	b16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("the pool holds %d, want 2", n)
	}
	for _, u := range me.ManaPool.Units() {
		if u.Color != "R" {
			t.Errorf("a %s mana in the pool, want only red", u.Color)
		}
	}
}

func TestM20B1ChandraNoviceDealsTwo(t *testing.T) {
	g, me, opp, chandra := m20ChandraTable(t)
	life := opp.Life
	b16Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}
