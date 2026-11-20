package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// thb_5514_b1_cards_test.go — Theros Beyond Death, batch 1 (#5514). Every
// test plays the real catalog entry through the engine.

const (
	thbB1NyxbornBruteOracle      = "986ba716-3921-4f64-9acf-703200d666f6"
	thbB1NyxbornColossusOracle   = "4a74cccb-5d75-4362-9c9e-2622096544fe"
	thbB1NyxbornCourserOracle    = "a9b345ea-3e34-48cf-8ab8-887c2b67df9c"
	thbB1NyxbornMarauderOracle   = "1552c6d1-f20a-4303-acf2-2fdd432a98e4"
	thbB1NyxbornSeaguardOracle   = "48cc79b8-f7a5-474b-bd88-7cd6ae673086"
	thbB1MossViperOracle         = "7d6f6ee4-77c2-4986-91c7-7dee962fc1a7"
	thbB1RiptideTurtleOracle     = "cb7e076c-a695-4c79-9f20-498370c81380"
	thbB1NyleasForerunnerOracle  = "bf0948b7-f7af-4d99-a8aa-ea7e6a69232d"
	thbB1ThunderingChariotOracle = "589ea35b-c64e-4cf6-81c5-806e9463e992"
	thbB1BrineGiantOracle        = "4ecbaf6a-aa03-45c8-a09e-6c7fc598ab9e"
	thbB1BronzeSwordOracle       = "14691783-1246-45ed-a439-c86d7b5eef62"
	thbB1EliteInstructorOracle   = "6536f72e-ff4e-4cda-bc7d-909145adf829"
	thbB1RumblingSentryOracle    = "7126772a-d586-4859-81ec-46bbe9d1a678"
	thbB1ThaumaturgesOracle      = "74a1a484-7c35-466a-b097-9aaf6fc2073b"
	thbB1GrimPhysicianOracle     = "381539ce-fc05-4855-80da-b0807a7c07c2"
	thbB1PheresBandOracle        = "01499544-6850-49f0-a66e-113df19c4340"
	thbB1IrreverentOracle        = "05abc57a-839d-4376-986c-8e3855e34239"
	thbB1HeroPrideOracle         = "f3a2d80c-de6f-4b0e-aeb1-78639d55a6e6"
	thbB1HeroWindsOracle         = "711893a5-2f4c-4957-90f5-2b1d6492b175"
	thbB1HeroNyxbornOracle       = "19e483e7-a0a5-4164-a5ea-dab51c07a3d7"
	thbB1DiscordantPiperOracle   = "bb0107f4-723a-4f2c-8e4b-237d5f668e3c"
	thbB1SphinxMindbreakerOracle = "a5270cf7-1eb3-4907-9f47-d787f61940d3"
	thbB1ScavengingHarpyOracle   = "94ea649c-9d4f-4882-a947-e2df4993b996"
	thbB1CarelessCelebOracle     = "56afad48-c5e5-4fbe-8889-15bcab234c31"
	thbB1DemonOfLoathingOracle   = "2e55fc03-779c-4a71-beb7-aa0ade738c24"
	thbB1HyraxTowerScoutOracle   = "c1261a72-ccbd-4be9-bc3b-c1bcb807eb85"
	thbB1AspectManticoreOracle   = "899f3758-bb06-4f9c-81eb-7a9002f03245"
	thbB1AspectLampreyOracle     = "7d3c25bb-5175-4fb1-992f-62a44722f535"
	thbB1MinionsReturnOracle     = "6cbe76c6-2b64-4389-b5f9-0b221cd92312"
	thbB1MiresGraspOracle        = "1009d2a2-a145-46e1-ba14-7a42735062fc"
	thbB1OreadOracle             = "ebbcbaa4-9dc2-4098-b6a2-393375d86f16"
	thbB1WitnessOracle           = "b46adfcd-fd7a-40c5-85f4-1cfee43a4c48"
	thbB1EidolonPhilosophyOracle = "65acea17-c900-4574-b22f-910285552aa2"
	thbB1SoulreaperOracle        = "508ce3a9-9050-4494-9adf-7f9cd6cc2f9f"
	thbB1OmenForgeOracle         = "e11e2dc8-b28e-415d-9a5d-0eb46da8945c"
	thbB1OmenDeadOracle          = "0aa9ee99-69a8-452b-a0d5-001384bdff3c"
	thbB1OmenHuntOracle          = "53295b03-717d-481d-9d5c-212c5f30bbbf"
	thbB1OmenSunOracle           = "08740dda-0629-4b6f-bbcc-c8088105b5e7"
	thbB1FinalDeathOracle        = "1be3eea4-eaff-469c-9689-abadac1e848c"
	thbB1FinalFlareOracle        = "34f9a03e-ca95-4d3e-9e63-7297b5513481"
	thbB1RelentlessOracle        = "d39c378b-f0a2-400f-af40-a632e24b6d72"
	thbB1MemoryDrainOracle       = "1a61929e-14bd-4f21-b173-00818d49743d"
	thbB1EatToExtinctionOracle   = "bf5f348e-d748-48e2-b483-740109e71176"
	thbB1FatefulEndOracle        = "a5495ab6-081a-42be-b50c-f89f15a587c8"
	thbB1ChainToMemoryOracle     = "a3805bb3-544c-4142-a6df-c60fa38d625a"
	thbB1SternDismissalOracle    = "319d5cf8-ab72-401e-b04e-c13a7a6c1aac"
	thbB1TriumphantSurgeOracle   = "ff281103-7a9b-4eb1-b81e-b8cd137b6e98"
	thbB1FuneralRitesOracle      = "771210f7-0f00-4438-a886-a1f716e3d80e"

	thbB1BearType = "Creature — Bear"
)

// thbB1AnswerPlayer answers the open pick_target prompt with a player.
func thbB1AnswerPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// thbB1Attached puts a catalog Aura or Equipment on the battlefield
// already attached to host.
func thbB1Attached(g *game.Game, owner, host uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// --- every card is in the catalog ----------------------------------------

func TestThb5514B1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		thbB1NyxbornBruteOracle: "Nyxborn Brute", thbB1NyxbornColossusOracle: "Nyxborn Colossus",
		thbB1NyxbornCourserOracle: "Nyxborn Courser", thbB1NyxbornMarauderOracle: "Nyxborn Marauder",
		thbB1NyxbornSeaguardOracle: "Nyxborn Seaguard", thbB1MossViperOracle: "Moss Viper",
		thbB1RiptideTurtleOracle: "Riptide Turtle", thbB1NyleasForerunnerOracle: "Nylea's Forerunner",
		thbB1ThunderingChariotOracle: "Thundering Chariot", thbB1BrineGiantOracle: "Brine Giant",
		thbB1BronzeSwordOracle: "Bronze Sword", thbB1EliteInstructorOracle: "Elite Instructor",
		thbB1RumblingSentryOracle: "Rumbling Sentry", thbB1ThaumaturgesOracle: "Thaumaturge's Familiar",
		thbB1GrimPhysicianOracle: "Grim Physician", thbB1PheresBandOracle: "Pheres-Band Brawler",
		thbB1IrreverentOracle: "Irreverent Revelers", thbB1HeroPrideOracle: "Hero of the Pride",
		thbB1HeroWindsOracle: "Hero of the Winds", thbB1HeroNyxbornOracle: "Hero of the Nyxborn",
		thbB1DiscordantPiperOracle: "Discordant Piper", thbB1SphinxMindbreakerOracle: "Sphinx Mindbreaker",
		thbB1ScavengingHarpyOracle: "Scavenging Harpy", thbB1CarelessCelebOracle: "Careless Celebrant",
		thbB1DemonOfLoathingOracle: "Demon of Loathing", thbB1HyraxTowerScoutOracle: "Hyrax Tower Scout",
		thbB1AspectManticoreOracle: "Aspect of Manticore", thbB1AspectLampreyOracle: "Aspect of Lamprey",
		thbB1MinionsReturnOracle: "Minion's Return", thbB1MiresGraspOracle: "Mire's Grasp",
		thbB1OreadOracle: "Oread of Mountain's Blaze", thbB1WitnessOracle: "Witness of Tomorrows",
		thbB1EidolonPhilosophyOracle: "Eidolon of Philosophy", thbB1SoulreaperOracle: "Soulreaper of Mogis",
		thbB1OmenForgeOracle: "Omen of the Forge", thbB1OmenDeadOracle: "Omen of the Dead",
		thbB1OmenHuntOracle: "Omen of the Hunt", thbB1OmenSunOracle: "Omen of the Sun",
		thbB1FinalDeathOracle: "Final Death", thbB1FinalFlareOracle: "Final Flare",
		thbB1RelentlessOracle: "Relentless Pursuit", thbB1MemoryDrainOracle: "Memory Drain",
		thbB1EatToExtinctionOracle: "Eat to Extinction", thbB1FatefulEndOracle: "Fateful End",
		thbB1ChainToMemoryOracle: "Chain to Memory", thbB1SternDismissalOracle: "Stern Dismissal",
		thbB1TriumphantSurgeOracle: "Triumphant Surge", thbB1FuneralRitesOracle: "Funeral Rites",
	}
	if len(want) != 48 {
		t.Fatalf("the batch registers 48 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

// --- printed keywords -------------------------------------------------------

func TestThb5514B1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Moss Viper", "Creature — Snake", thbB1MossViperOracle, []string{"deathtouch"}},
		{"Riptide Turtle", "Creature — Turtle", thbB1RiptideTurtleOracle, []string{"flash", "defender"}},
		{"Nylea's Forerunner", "Enchantment Creature — Beast", thbB1NyleasForerunnerOracle, []string{"trample"}},
		{"Thaumaturge's Familiar", "Artifact Creature — Bird", thbB1ThaumaturgesOracle, []string{"flying"}},
		{"Hero of the Winds", "Creature — Human Soldier", thbB1HeroWindsOracle, []string{"flying"}},
		{"Sphinx Mindbreaker", "Creature — Sphinx", thbB1SphinxMindbreakerOracle, []string{"flying"}},
		{"Scavenging Harpy", "Creature — Harpy", thbB1ScavengingHarpyOracle, []string{"flying"}},
		{"Demon of Loathing", "Creature — Demon", thbB1DemonOfLoathingOracle, []string{"flying", "trample"}},
		{"Witness of Tomorrows", "Enchantment Creature — Sphinx", thbB1WitnessOracle, []string{"flying"}},
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

func TestThb5514B1NyleasForerunnerGivesYourOtherCreaturesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Nylea's Forerunner", "Enchantment Creature — Beast", thbB1NyleasForerunnerOracle, 5, 3)
	mine := b12Creature(g, me.ID, "Mine", thbB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", thbB1BearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("your other creature has no trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
}

func TestThb5514B1ThunderingChariotCrewsIntoAFirstStrikingTrampler(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chariot := b12Push(g, me.ID, "Thundering Chariot", "Artifact — Vehicle", thbB1ThunderingChariotOracle, 3, 3)
	crew := b12Creature(g, me.ID, "Crew", "Creature — Test", 1, 1)
	advanceToMain(t, g)
	if isCreatureNow(g, chariot) {
		t.Fatal("the Vehicle is a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, chariot, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isCreatureNow(g, chariot) {
		t.Fatal("the crewed Vehicle is not a creature")
	}
	for _, kw := range []string{"first strike", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, chariot, kw) {
			t.Errorf("the crewed Chariot has no %s", kw)
		}
	}
}

func TestThb5514B1BrineGiantCostsOneLessPerEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, thbB1BrineGiantOracle, "Creature — Giant", "{6}{U}"); got != 7 {
		t.Fatalf("bare = %d, want 7", got)
	}
	b12Permanent(g, me.ID, "Aura A", "Enchantment")
	b12Permanent(g, me.ID, "Aura B", "Enchantment")
	b12Permanent(g, me.ID, "Rock", "Artifact")
	if got := selfPricedMV(t, g, me, thbB1BrineGiantOracle, "Creature — Giant", "{6}{U}"); got != 5 {
		t.Errorf("with two enchantments = %d, want 5", got)
	}
}

func TestThb5514B1BronzeSwordGivesPlusTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", thbB1BearType, 2, 2)
	thbB1Attached(g, me.ID, bear, "Bronze Sword", "Artifact — Equipment", thbB1BronzeSwordOracle)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

// --- enters triggers ----------------------------------------------------------

func TestThb5514B1EliteInstructorLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Elite Instructor", "Creature — Human Wizard", thbB1EliteInstructorOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestThb5514B1EntersAndScries(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Rumbling Sentry", "Creature — Giant", thbB1RumblingSentryOracle},
		{"Thaumaturge's Familiar", "Artifact Creature — Bird", thbB1ThaumaturgesOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 3)
			passPriorityAroundTable(t, g)
			if scryChoiceFor(g, me.ID) == nil {
				t.Error("no scry prompt")
			}
		})
	}
}

func TestThb5514B1HyraxTowerScoutUntapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
	b16Tap(g, bear)
	castHoldCreature(t, g, "Hyrax Tower Scout", "Creature — Human Scout", thbB1HyraxTowerScoutOracle, 3, 3)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if hmlTapped(t, g, bear) {
		t.Error("the target is still tapped")
	}
}

func TestThb5514B1PheresBandBrawlerFights(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
	brawler := castHoldCreature(t, g, "Pheres-Band Brawler", "Creature — Centaur Warrior", thbB1PheresBandOracle, 4, 4)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the fought creature survived 4 damage")
	}
	if got := findCardAnywhere(t, g, brawler).DamageMarked; got != 2 {
		t.Errorf("the Brawler has %d damage, want 2", got)
	}
}

func TestThb5514B1IrreverentRevelersHasteMode(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Irreverent Revelers", "Creature — Satyr", thbB1IrreverentOracle, 2, 2)
	// With no artifact to destroy only the haste bullet is offered.
	chooseModeNow(t, g, g.Seats[g.Turn.ActiveSeat].ID, []int{1}, 1)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, id, "haste") {
		t.Error("the Revelers did not gain haste")
	}
}

func TestThb5514B1IrreverentRevelersDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	ench := b12Permanent(g, opp.ID, "Pact", "Enchantment")
	castHoldCreature(t, g, "Irreverent Revelers", "Creature — Satyr", thbB1IrreverentOracle, 2, 2)
	chooseModeNow(t, g, g.Seats[g.Turn.ActiveSeat].ID, []int{0, 1}, 0)
	answerPickTargetIfOpen(t, g, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("an enchantment was destroyed")
	}
}

func TestThb5514B1HeroOfTheNyxbornMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Hero of the Nyxborn", "Enchantment Creature — Human Soldier", thbB1HeroNyxbornOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Human Soldier", me.ID); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

func TestThb5514B1SphinxMindbreakerMillsEachOpponentTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	libs := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		libs[s.ID] = s.Library.Size()
	}
	castHoldCreature(t, g, "Sphinx Mindbreaker", "Creature — Sphinx", thbB1SphinxMindbreakerOracle, 6, 6)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := 10
		if s.ID == me.ID {
			want = 0
		}
		if got := libs[s.ID] - s.Library.Size(); got != want {
			t.Errorf("%s milled %d, want %d", s.Name, got, want)
		}
	}
}

func TestThb5514B1ScavengingHarpyExilesAnOpponentsCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := pushGraveyardCardForTest(opp, "Old Card")
	castHoldCreature(t, g, "Scavenging Harpy", "Creature — Harpy", thbB1ScavengingHarpyOracle, 2, 1)
	answerPickTarget(t, g, card)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("the card was not exiled")
	}
}

// --- heroic -------------------------------------------------------------------

func TestThb5514B1HeroicPumpsYourTeam(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Hero of the Pride", "Creature — Cat Soldier", thbB1HeroPrideOracle},
		{"Hero of the Winds", "Creature — Human Soldier", thbB1HeroWindsOracle},
		{"Hero of the Nyxborn", "Enchantment Creature — Human Soldier", thbB1HeroNyxbornOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			advanceToMain(t, g)
			hero := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			bear := b12Creature(g, me.ID, "Bear", thbB1BearType, 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
			castCatalogSpell(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(hero))
			passPriorityAroundTable(t, g)
			if p, _ := ptOf(t, g, bear); p != 3 {
				t.Errorf("Bear power %d, want 3 (the Hero's +1)", p)
			}
			if p, _ := ptOf(t, g, theirs); p != 2 {
				t.Errorf("an opponent's creature is %d power, want 2", p)
			}
		})
	}
}

// --- dies and combat-damage triggers -------------------------------------------

func TestThb5514B1GrimPhysicianShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	physician := b12Push(g, me.ID, "Grim Physician", "Creature — Zombie", thbB1GrimPhysicianOracle, 1, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b27Kill(g, physician)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, ogre); p != 2 || tough != 2 {
		t.Errorf("the Ogre is %d/%d, want 2/2", p, tough)
	}
}

func TestThb5514B1DiscordantPiperLeavesAGoat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	piper := b12Push(g, me.ID, "Discordant Piper", "Creature — Zombie Satyr", thbB1DiscordantPiperOracle, 2, 1)
	b27Kill(g, piper)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Goat", me.ID); n != 1 {
		t.Errorf("%d Goats, want 1", n)
	}
}

func TestThb5514B1CarelessCelebrantDealsTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celebrant := b12Push(g, me.ID, "Careless Celebrant", "Creature — Satyr Shaman", thbB1CarelessCelebOracle, 2, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b27Kill(g, celebrant)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

func TestThb5514B1DemonOfLoathingMakesTheDamagedPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	demon := b12Push(g, me.ID, "Demon of Loathing", "Creature — Demon", thbB1DemonOfLoathingOracle, 7, 7)
	theirs := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
	bystander := b12Creature(g, other.ID, "Bystander", thbB1BearType, 2, 2)
	dealCombatDamage(g, me.ID, demon, opp.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the damaged player kept their only creature")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a player not dealt damage sacrificed a creature")
	}
}

// --- Auras and Equipment ----------------------------------------------------------

func TestThb5514B1AspectOfManticoreGivesFirstStrikeThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", thbB1BearType, 2, 2)
	castCatalogSpell(t, g, "Aspect of Manticore", auraTypeLine, thbB1AspectManticoreOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !hasKeywordOf(g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

func TestThb5514B1AspectOfLampreyDiscardsTwoAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d2dSetHand(opp, 3)
	bear := b12Creature(g, me.ID, "Bear", thbB1BearType, 2, 2)
	castCatalogSpell(t, g, "Aspect of Lamprey", auraTypeLine, thbB1AspectLampreyOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	thbB1AnswerPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
	if !hasKeywordOf(g, bear, "lifelink") {
		t.Error("the enchanted creature has no lifelink")
	}
}

func TestThb5514B1MinionsReturnStealsTheDeadCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := auraBear(g, opp.ID)
	auraCast(t, g, "Minion's Return", thbB1MinionsReturnOracle, bear)
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears under the Aura's controller, want 1", n)
	}
}

func TestThb5514B1MiresGraspShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	thbB1Attached(g, me.ID, ogre, "Mire's Grasp", auraTypeLine, thbB1MiresGraspOracle)
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the enchanted Ogre is %d/%d, want 1/1", p, tough)
	}
}

// --- activated abilities -------------------------------------------------------

func TestThb5514B1OreadRummages(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	oread := b12Push(g, me.ID, "Oread of Mountain's Blaze", "Enchantment Creature — Nymph", thbB1OreadOracle, 1, 3)
	socB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	floatForTest(g, me, "RRR")
	apaActivate(t, g, me, oread, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}})
	if !me.Graveyard.Contains(pitched) || me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("rummage: hand %d, library %d -> %d", me.Hand.Size(), lib, me.Library.Size())
	}
}

func TestThb5514B1WitnessOfTomorrowsScries(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	witness := b12Push(g, me.ID, "Witness of Tomorrows", "Enchantment Creature — Sphinx", thbB1WitnessOracle, 3, 4)
	floatForTest(g, me, "UUUU")
	apaActivate(t, g, me, witness, 0, game.ActivateAbilityParams{})
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestThb5514B1EidolonOfPhilosophyDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	eidolon := b12Push(g, me.ID, "Eidolon of Philosophy", "Enchantment Creature — Spirit", thbB1EidolonPhilosophyOracle, 1, 2)
	hand := me.Hand.Size()
	floatForTest(g, me, "UUUUUUU")
	apaActivate(t, g, me, eidolon, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(eidolon) {
		t.Error("the Eidolon was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestThb5514B1SoulreaperOfMogisSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	reaper := b12Push(g, me.ID, "Soulreaper of Mogis", "Enchantment Creature — Minotaur Shaman", thbB1SoulreaperOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", thbB1BearType, 2, 2)
	hand := me.Hand.Size()
	floatForTest(g, me, "BBB")
	apaActivate(t, g, me, reaper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- the Omens ----------------------------------------------------------------------

func TestThb5514B1OmensSacrificeToScryTwo(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Omen of the Forge", thbB1OmenForgeOracle, "RRR"},
		{"Omen of the Dead", thbB1OmenDeadOracle, "BBB"},
		{"Omen of the Hunt", thbB1OmenHuntOracle, "GGG"},
		{"Omen of the Sun", thbB1OmenSunOracle, "WWW"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			omen := pushCatalogPermanent(g, me.ID, tc.name, "Enchantment", tc.oracle, false)
			floatForTest(g, me, tc.mana)
			apaActivate(t, g, me, omen, 0, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(omen) || !me.Graveyard.Contains(omen) {
				t.Error("the Omen was not sacrificed")
			}
			if scryChoiceFor(g, me.ID) == nil {
				t.Error("no scry prompt")
			}
		})
	}
}

func TestThb5514B1OmenOfTheForgeDealsTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Omen of the Forge", "Enchantment", thbB1OmenForgeOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

func TestThb5514B1OmenOfTheDeadReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := pushGraveyardCardForTest(me, "Old Bear")
	castCatalogSpell(t, g, "Omen of the Dead", "Enchantment", thbB1OmenDeadOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, card)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the creature card did not return to hand")
	}
}

func TestThb5514B1OmenOfTheHuntFetchesATappedBasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Omen of the Hunt", "Enchantment", thbB1OmenHuntOracle, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was not fetched")
	}
	if !hmlTapped(t, g, forest) {
		t.Error("the fetched land entered untapped")
	}
}

func TestThb5514B1OmenOfTheSunMakesTwoSoldiersAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Omen of the Sun", "Enchantment", thbB1OmenSunOracle, nil)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Human Soldier", me.ID); n != 2 {
		t.Errorf("%d Human Soldier tokens, want 2", n)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- the spells -------------------------------------------------------------------

func TestThb5514B1FinalDeathExiles(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
	castCatalogSpell(t, g, "Final Death", "Instant", thbB1FinalDeathOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestThb5514B1FinalFlareSacrificesThenDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	victim := b12Creature(g, me.ID, "Victim", thbB1BearType, 2, 2)
	wurm := b12Creature(g, opp.ID, "Wurm", "Creature — Wurm", 6, 6)
	id := b13HandCard(me, "Final Flare", "Instant", thbB1FinalFlareOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(wurm)}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(wurm), SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if got := findCardAnywhere(t, g, wurm).DamageMarked; got != 5 {
		t.Errorf("the target has %d damage, want 5", got)
	}
}

func TestThb5514B1RelentlessPursuitTakesACreatureAndALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := trekB48Library(me,
		game.Card{Name: "Bear", TypeLine: thbB1BearType},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Other Bear", TypeLine: thbB1BearType},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "Relentless Pursuit", "Sorcery", thbB1RelentlessOracle, nil)
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0], lib[1]}); err != nil {
		t.Fatalf("a creature and a land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[0]) || !me.Hand.Contains(lib[1]) {
		t.Error("the picks are not in hand")
	}
	for _, i := range []int{2, 3} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[4]) {
		t.Error("the fifth card left the library")
	}
}

func TestThb5514B1MemoryDrainCountersThenScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Memory Drain", "Instant", thbB1MemoryDrainOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestThb5514B1EatToExtinctionExilesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", thbB1BearType, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Eat to Extinction", "Instant", thbB1EatToExtinctionOracle, cardRefs(rock)); err == nil {
		t.Fatal("Eat to Extinction targeted an artifact")
	}
	castCatalogSpell(t, g, "Eat to Extinction", "Instant", thbB1EatToExtinctionOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestThb5514B1FatefulEndDealsThreeThenScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	castCatalogSpell(t, g, "Fateful End", "Instant", thbB1FatefulEndOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestThb5514B1ChainToMemoryShrinksThenScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wurm := b12Creature(g, opp.ID, "Wurm", "Creature — Wurm", 5, 5)
	castCatalogSpell(t, g, "Chain to Memory", "Instant", thbB1ChainToMemoryOracle, cardRefs(wurm))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, wurm); p != 1 || tough != 5 {
		t.Errorf("the Wurm is %d/%d, want 1/5", p, tough)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestThb5514B1SternDismissalBouncesAnOpponentsEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", thbB1BearType, 2, 2)
	pact := b12Permanent(g, opp.ID, "Pact", "Enchantment")
	if err := castCatalogSpellErr(t, g, "Stern Dismissal", "Instant", thbB1SternDismissalOracle, cardRefs(mine)); err == nil {
		t.Fatal("Stern Dismissal targeted its caster's own creature")
	}
	castCatalogSpell(t, g, "Stern Dismissal", "Instant", thbB1SternDismissalOracle, cardRefs(pact))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pact) || !opp.Hand.Contains(pact) {
		t.Error("the enchantment did not return to its owner's hand")
	}
}

func TestThb5514B1TriumphantSurgeDestroysABigCreatureAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Small", thbB1BearType, 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Wurm", 4, 4)
	if err := castCatalogSpellErr(t, g, "Triumphant Surge", "Instant", thbB1TriumphantSurgeOracle, cardRefs(small)); err == nil {
		t.Fatal("Triumphant Surge targeted a creature with power 3")
	}
	life := me.Life
	castCatalogSpell(t, g, "Triumphant Surge", "Instant", thbB1TriumphantSurgeOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the creature survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestThb5514B1FuneralRitesDrawsLosesAndMills(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hand, lib, life := me.Hand.Size(), me.Library.Size(), me.Life
	castCatalogSpell(t, g, "Funeral Rites", "Sorcery", thbB1FuneralRitesOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if got := lib - me.Library.Size(); got != 4 {
		t.Errorf("the library lost %d cards, want 4 (two drawn, two milled)", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}
