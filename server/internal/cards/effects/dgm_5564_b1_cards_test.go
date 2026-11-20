package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dgm_5564_b1_cards_test.go — the first batch of Dragon's Maze cards
// (#5564). Every test plays the real catalog entry through the engine.

const (
	dgmArmoredWolfRiderOracle    = "e0965cd1-18fa-4ccc-8c5f-1414c6567d67"
	dgmBaneAlleyBlackguardOracle = "1134eea5-e931-4cb2-9e30-64c227386d43"
	dgmAscendedLawmageOracle     = "4317f4f1-d339-4940-b46e-659641035595"
	dgmBatteringKrasisOracle     = "3d0c1301-4fa7-4245-a13f-f613188eb638"
	dgmMurmuringPhantasmOracle   = "35e684fd-69e3-4137-bf15-94f705191f5e"
	dgmSpikeJesterOracle         = "29f22168-21aa-4ca9-ab72-b09396a12c76"
	dgmSteepleRocOracle          = "f1ba6053-14e0-46b0-b760-b72e4fa0c5f5"
	dgmWoodlotCrawlerOracle      = "e5b11e43-fc7c-4cea-a44a-8af3f124e29e"
	dgmRiotPikerOracle           = "864c81e4-d612-4939-b4df-4b125cb8cdf2"
	dgmRakdosDrakeOracle         = "0715167b-00b1-45a4-94a0-60d4097a1428"
	dgmTitheDrinkerOracle        = "06b1fd96-a3ad-4ebd-9ea4-b3cfb4c9c43d"
	dgmSkylasherOracle           = "153a4c11-e75f-46d7-b684-17d500ea13e6"
	dgmAzoriusCluestoneOracle    = "27e04c41-f42c-4d60-8a71-ec2d7c326f64"
	dgmBorosCluestoneOracle      = "f6c3b420-9aca-4fab-b7bb-1814bd2d93f6"
	dgmDimirCluestoneOracle      = "62f834c7-4765-4b6f-812c-7ce14e13ea0d"
	dgmGolgariCluestoneOracle    = "e917ad3d-df8e-442b-918c-3aa7a10f3cc4"
	dgmGruulCluestoneOracle      = "80911151-1e72-4d10-b00a-9fcff2dd131a"
	dgmIzzetCluestoneOracle      = "3c76bb90-92fb-428b-8e94-def4cf4b6f2d"
	dgmOrzhovCluestoneOracle     = "09650a76-05c4-40a0-b861-d146b3ba87d6"
	dgmRakdosCluestoneOracle     = "b4f0bda0-1051-4ab8-b254-891f427b33ad"
	dgmSelesnyaCluestoneOracle   = "792d4afd-b3de-4be2-b68a-03a5be7b9627"
	dgmSimicCluestoneOracle      = "0b65050d-a9cd-4b6c-9114-af89bfa6682f"
	dgmMazeAbominationOracle     = "eaa72823-413e-474e-a926-aebb0a78ff31"
	dgmMazeBehemothOracle        = "d54c258f-1914-449a-857d-11f9478bff63"
	dgmMazeGliderOracle          = "2388a064-ae84-4eac-b1d9-819a40052d12"
	dgmMazeRusherOracle          = "f6ef6ac6-bf1f-41ed-bd3f-a3806a76a336"
	dgmMazeSentinelOracle        = "635360cd-0114-47d1-8a98-e6a3678652a8"
	dgmZhurTaaAncientOracle      = "6543b60c-12b0-4e69-9074-6a2b468c5234"
	dgmOpalLakeGatekeepersOracle = "a5375366-7ddc-4bd5-91c9-dc08f8beb3f6"
	dgmSaruliGatekeepersOracle   = "509ea58a-da0a-4e9a-aafc-15f4f65a7467"
	dgmSunspireGatekeepersOracle = "df2cc4ba-70b8-4875-8f84-9c268ab8d5f6"
	dgmUbulSarGatekeepersOracle  = "f4e15575-2277-4cb1-a343-e9407c050eb1"
	dgmBronzebeakMoaOracle       = "16f03175-9c57-4058-9c9d-9c4936e4c324"
	dgmViashinoFirstbladeOracle  = "766e277c-983e-488c-a885-c03c3c0d3357"
	dgmHaunterOfNightveilOracle  = "496c852b-ec62-4320-afbb-e19f534954c9"
	dgmAdventOfTheWurmOracle     = "eb62aa4b-c11b-4195-ae85-cff8f78ce17b"
	dgmFatalFumesOracle          = "359ceca4-7063-4089-8088-ba54bc6fb7cf"
	dgmPhytoburstOracle          = "d68d175f-f59c-4995-84d4-fc593280194b"
	dgmMendingTouchOracle        = "e3172ac0-213c-4da4-8b7a-8f84bc237cde"
	dgmClearAPathOracle          = "05a1575c-8d08-4845-8ae0-85e0aba55637"
	dgmMindstaticOracle          = "880e0177-f1d2-4bdc-83b4-cae6a0706ef9"
	dgmPilferedPlansOracle       = "b734cfd9-afef-41f1-8377-68c885fb3a69"
	dgmWarleadersHelixOracle     = "b1ea88f5-162b-4f08-bec3-be1f06dec392"
	dgmObzedatsAidOracle         = "f3be91ce-e99b-4a51-a0f7-a000fd9d0e14"

	dgmBearType       = "Creature — Bear"
	dgmGatekeeperType = "Creature — Human Soldier"
)

// dgmGates puts n Gates onto the battlefield under owner's control.
func dgmGates(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Gate", "Land — Gate")
	}
}

// --- every card is in the catalog ----------------------------------------

func TestDgm5564B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		dgmArmoredWolfRiderOracle, dgmBaneAlleyBlackguardOracle, dgmAscendedLawmageOracle, dgmBatteringKrasisOracle,
		dgmMurmuringPhantasmOracle, dgmSpikeJesterOracle, dgmSteepleRocOracle, dgmWoodlotCrawlerOracle,
		dgmRiotPikerOracle, dgmRakdosDrakeOracle, dgmTitheDrinkerOracle, dgmSkylasherOracle,
		dgmAzoriusCluestoneOracle, dgmBorosCluestoneOracle, dgmDimirCluestoneOracle, dgmGolgariCluestoneOracle,
		dgmGruulCluestoneOracle, dgmIzzetCluestoneOracle, dgmOrzhovCluestoneOracle, dgmRakdosCluestoneOracle,
		dgmSelesnyaCluestoneOracle, dgmSimicCluestoneOracle, dgmMazeAbominationOracle, dgmMazeBehemothOracle,
		dgmMazeGliderOracle, dgmMazeRusherOracle, dgmMazeSentinelOracle, dgmZhurTaaAncientOracle,
		dgmOpalLakeGatekeepersOracle, dgmSaruliGatekeepersOracle, dgmSunspireGatekeepersOracle, dgmUbulSarGatekeepersOracle,
		dgmBronzebeakMoaOracle, dgmViashinoFirstbladeOracle, dgmHaunterOfNightveilOracle, dgmAdventOfTheWurmOracle,
		dgmFatalFumesOracle, dgmPhytoburstOracle, dgmMendingTouchOracle, dgmClearAPathOracle,
		dgmMindstaticOracle, dgmPilferedPlansOracle, dgmWarleadersHelixOracle, dgmObzedatsAidOracle,
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

// --- vanilla bodies and printed keywords ------------------------------------

func TestDgm5564B1VanillaCreaturesResolve(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Armored Wolf-Rider", "Creature — Elf Knight", dgmArmoredWolfRiderOracle, 4, 6},
		{"Bane Alley Blackguard", "Creature — Human Rogue", dgmBaneAlleyBlackguardOracle, 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if p, tough := ptOf(t, g, id); p != tc.p || tough != tc.tough {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.p, tc.tough)
			}
		})
	}
}

func TestDgm5564B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ascended Lawmage", "Creature — Vedalken Wizard", dgmAscendedLawmageOracle, []string{"flying", "hexproof"}},
		{"Battering Krasis", "Creature — Shark Beast", dgmBatteringKrasisOracle, []string{"trample"}},
		{"Murmuring Phantasm", "Creature — Spirit", dgmMurmuringPhantasmOracle, []string{"defender"}},
		{"Spike Jester", "Creature — Goblin Warrior", dgmSpikeJesterOracle, []string{"haste"}},
		{"Steeple Roc", "Creature — Bird", dgmSteepleRocOracle, []string{"flying", "first strike"}},
		{"Woodlot Crawler", "Creature — Insect", dgmWoodlotCrawlerOracle, []string{"forestwalk", "protection from green"}},
		{"Riot Piker", "Creature — Goblin Berserker", dgmRiotPikerOracle, []string{"first strike"}},
		{"Rakdos Drake", "Creature — Drake", dgmRakdosDrakeOracle, []string{"flying"}},
		{"Tithe Drinker", "Creature — Vampire", dgmTitheDrinkerOracle, []string{"lifelink"}},
		{"Skylasher", "Creature — Insect", dgmSkylasherOracle, []string{"flash", "reach", "protection from blue"}},
		{"Maze Abomination", "Creature — Elemental", dgmMazeAbominationOracle, []string{"deathtouch"}},
		{"Maze Behemoth", "Creature — Elemental", dgmMazeBehemothOracle, []string{"trample"}},
		{"Maze Glider", "Creature — Elemental", dgmMazeGliderOracle, []string{"flying"}},
		{"Maze Rusher", "Creature — Elemental", dgmMazeRusherOracle, []string{"haste"}},
		{"Maze Sentinel", "Creature — Elemental", dgmMazeSentinelOracle, []string{"vigilance"}},
		{"Viashino Firstblade", "Creature — Lizard Soldier", dgmViashinoFirstbladeOracle, []string{"haste"}},
	}
	for _, tc := range cases {
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

// Unleash is the engine's keyword on the Drake's printed keyword list.
func TestDgm5564B1RakdosDrakeHasUnleash(t *testing.T) {
	spec, ok := Lookup(dgmRakdosDrakeOracle)
	if !ok {
		t.Fatal("Rakdos Drake is not in the catalog")
	}
	if !containsString(spec.PrintedKeywords, game.KeywordUnleash) {
		t.Errorf("printed keywords %v, want unleash", spec.PrintedKeywords)
	}
}

// Evolve: a bigger creature entering puts a +1/+1 counter on the Krasis.
func TestDgm5564B1BatteringKrasisEvolves(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	krasis := b12Push(g, me.ID, "Battering Krasis", "Creature — Shark Beast", dgmBatteringKrasisOracle, 2, 1)
	castHoldCreature(t, g, "Big Bear", dgmBearType, "", 3, 3)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, krasis); p != 3 || tough != 2 {
		t.Errorf("the Krasis is %d/%d, want 3/2", p, tough)
	}
}

// Forestwalk: a defending player who controls a Forest can't block it.
func TestDgm5564B1WoodlotCrawlerForestwalks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := b12Push(g, me.ID, "Woodlot Crawler", "Creature — Insect", dgmWoodlotCrawlerOracle, 2, 1)
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	brAttack(t, g, crawler)
	if err := g.DeclareBlocker(zombie, crawler); err == nil {
		t.Error("a player with a Forest blocked a forestwalker")
	}
}

func TestDgm5564B1RiotPikerMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	piker := b12Push(g, me.ID, "Riot Piker", "Creature — Goblin Berserker", dgmRiotPikerOracle, 2, 1)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != piker {
		t.Fatalf("pass with the Piker home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(piker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Piker attacking: %v", err)
	}
}

func TestDgm5564B1TitheDrinkerExtortsOnACast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tithe Drinker", "Creature — Vampire", dgmTitheDrinkerOracle, 2, 1)
	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell did not offer the extort payment")
	}
	l0, l1 := me.Life, opp.Life
	floatForTest(g, me, "W")
	answerPayUnless(t, g, me.ID, true)
	if me.Life != l0+3 || opp.Life != l1-1 {
		t.Errorf("life me %d->%d, opp %d->%d; want a 1-point drain on each of three opponents", l0, me.Life, l1, opp.Life)
	}
}

// Flash: Skylasher is cast outside a main phase, and it can't be countered.
func TestDgm5564B1SkylasherHasFlashAndCantBeCountered(t *testing.T) {
	if spec, _ := Lookup(dgmSkylasherOracle); !spec.CantBeCountered {
		t.Error("Skylasher can be countered")
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Skylasher", "Creature — Insect", dgmSkylasherOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Skylasher at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Skylasher did not resolve")
	}
}

// --- Cluestones ----------------------------------------------------------------

func TestDgm5564B1Cluestones(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, a, b string
	}{
		{"Azorius Cluestone", dgmAzoriusCluestoneOracle, "W", "U"},
		{"Boros Cluestone", dgmBorosCluestoneOracle, "R", "W"},
		{"Dimir Cluestone", dgmDimirCluestoneOracle, "U", "B"},
		{"Golgari Cluestone", dgmGolgariCluestoneOracle, "B", "G"},
		{"Gruul Cluestone", dgmGruulCluestoneOracle, "R", "G"},
		{"Izzet Cluestone", dgmIzzetCluestoneOracle, "U", "R"},
		{"Orzhov Cluestone", dgmOrzhovCluestoneOracle, "W", "B"},
		{"Rakdos Cluestone", dgmRakdosCluestoneOracle, "B", "R"},
		{"Selesnya Cluestone", dgmSelesnyaCluestoneOracle, "G", "W"},
		{"Simic Cluestone", dgmSimicCluestoneOracle, "G", "U"},
	} {
		t.Run(tc.name+" taps for mana", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			stone := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tap for mana: %v", err)
			}
			if pick := riderLatestManaPick(g, me.ID); pick != nil {
				if err := g.ResolveManaChoice(pick.ID, me.ID, tc.b); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
			if got := me.ManaPool.Total(); got != 1 {
				t.Errorf("pool = %d, want 1", got)
			}
		})
		t.Run(tc.name+" sacrifices to draw", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			stone := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			me.ManaPool.AddMana(game.ManaToken{Color: tc.a})
			if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("one mana paid for a two-mana cost")
			}
			me.ManaPool.AddMana(game.ManaToken{Color: tc.b})
			hand := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(stone) {
				t.Error("the Cluestone was not sacrificed as a cost")
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
		})
	}
}

// --- the Maze cycle --------------------------------------------------------------

func TestDgm5564B1MazeCycleGrantsMulticoloredCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
	}{
		{"Maze Abomination", dgmMazeAbominationOracle, "deathtouch"},
		{"Maze Behemoth", dgmMazeBehemothOracle, "trample"},
		{"Maze Glider", dgmMazeGliderOracle, "flying"},
		{"Maze Rusher", dgmMazeRusherOracle, "haste"},
		{"Maze Sentinel", dgmMazeSentinelOracle, "vigilance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b31Push(g, me.ID, tc.name, "Creature — Elemental", tc.oracle, "", 3, 3, "B")
			gold := b31Push(g, me.ID, "Gold", dgmBearType, "", "", 2, 2, "W", "U")
			mono := b31Push(g, me.ID, "Mono", dgmBearType, "", "", 2, 2, "G")
			theirs := b31Push(g, opp.ID, "Their Gold", dgmBearType, "", "", 2, 2, "R", "G")
			if !hasEffectiveKeyword(t, g, gold, tc.keyword) {
				t.Errorf("your multicolored creature has no %s", tc.keyword)
			}
			if hasEffectiveKeyword(t, g, mono, tc.keyword) {
				t.Errorf("your monocolored creature has %s", tc.keyword)
			}
			if hasEffectiveKeyword(t, g, theirs, tc.keyword) {
				t.Errorf("an opponent's multicolored creature has %s", tc.keyword)
			}
		})
	}
}

// --- Zhur-Taa Ancient ------------------------------------------------------------

func TestDgm5564B1ZhurTaaAncientDoublesEveryPlayersLand(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Zhur-Taa Ancient", "Creature — Beast", dgmZhurTaaAncientOracle, "", 7, 5, "R", "G")
	mine := b31Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	theirs := b31Push(g, them.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	tapForMana(t, g, me.ID, mine)
	if got := sortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("my pool = %v, want two red", got)
	}
	tapForMana(t, g, them.ID, theirs)
	if got := sortedPool(them); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("their pool = %v, want two black", got)
	}
}

// --- the Gatekeepers ----------------------------------------------------------------

func TestDgm5564B1SaruliGatekeepersGainsSevenWithTwoGates(t *testing.T) {
	for _, tc := range []struct {
		gates, gain int
	}{{2, 7}, {1, 0}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		dgmGates(g, me.ID, tc.gates)
		life := me.Life
		castAndResolveCreature(t, g, "Saruli Gatekeepers", "Creature — Elf Warrior", dgmSaruliGatekeepersOracle)
		passPriorityAroundTable(t, g)
		if got := me.Life - life; got != tc.gain {
			t.Errorf("with %d Gates: gained %d, want %d", tc.gates, got, tc.gain)
		}
	}
}

func TestDgm5564B1SunspireGatekeepersMakesAKnightWithTwoGates(t *testing.T) {
	for _, tc := range []struct {
		gates, knights int
	}{{2, 1}, {1, 0}} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		dgmGates(g, me.ID, tc.gates)
		castAndResolveCreature(t, g, "Sunspire Gatekeepers", dgmGatekeeperType, dgmSunspireGatekeepersOracle)
		passPriorityAroundTable(t, g)
		if n, _ := countTokensNamed(g, me.ID, "Knight"); n != tc.knights {
			t.Errorf("with %d Gates: %d Knights, want %d", tc.gates, n, tc.knights)
		}
	}
}

func TestDgm5564B1OpalLakeGatekeepersMayDrawWithTwoGates(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	dgmGates(g, me.ID, 2)
	hand := me.Hand.Size()
	castAndResolveCreature(t, g, "Opal Lake Gatekeepers", "Creature — Vedalken Soldier", dgmOpalLakeGatekeepersOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestDgm5564B1UbulSarGatekeepersShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	dgmGates(g, me.ID, 2)
	mine := b12Creature(g, me.ID, "Mine", dgmBearType, 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", dgmBearType, 3, 3)
	castAndResolveCreature(t, g, "Ubul Sar Gatekeepers", "Creature — Zombie Soldier", dgmUbulSarGatekeepersOracle)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
		t.Fatalf("legal targets wrong: %+v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

func TestDgm5564B1UbulSarGatekeepersNeedsTwoGates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	dgmGates(g, me.ID, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", dgmBearType, 3, 3)
	castAndResolveCreature(t, g, "Ubul Sar Gatekeepers", "Creature — Zombie Soldier", dgmUbulSarGatekeepersOracle)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("the trigger asked for a target with one Gate: %+v", p)
	}
	if p, tough := ptOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("the creature is %d/%d with one Gate, want 3/3", p, tough)
	}
}

// --- other creatures --------------------------------------------------------------

func TestDgm5564B1BronzebeakMoaGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	moa := b12Push(g, me.ID, "Bronzebeak Moa", "Creature — Bird", dgmBronzebeakMoaOracle, 2, 2)
	castHoldCreature(t, g, "Bear", dgmBearType, "", 2, 2)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, moa); p != 5 || tough != 5 {
		t.Errorf("the Moa is %d/%d, want 5/5", p, tough)
	}
}

func TestDgm5564B1ViashinoFirstbladeGetsTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Viashino Firstblade", "Creature — Lizard Soldier", dgmViashinoFirstbladeOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Firstblade is %d/%d, want 4/4", p, tough)
	}
}

func TestDgm5564B1HaunterOfNightveilShrinksOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Haunter of Nightveil", "Creature — Spirit", dgmHaunterOfNightveilOracle, 3, 4)
	mine := b12Creature(g, me.ID, "Mine", dgmBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", dgmBearType, 3, 3)
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 3 {
		t.Errorf("the opponent's creature is %d/%d, want 2/3", p, tough)
	}
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 2/2", p, tough)
	}
}

// --- the spells ----------------------------------------------------------------------

func TestDgm5564B1AdventOfTheWurmMakesATramplingWurm(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	castCatalogSpell(t, g, "Advent of the Wurm", "Instant", dgmAdventOfTheWurmOracle, nil)
	passPriorityAroundTable(t, g)
	wurm, ok := visBattlefieldNamed(g, me.ID, "Wurm")
	if !ok {
		t.Fatal("no Wurm token")
	}
	if p, tough := ptOf(t, g, wurm.InstanceID); p != 5 || tough != 5 {
		t.Errorf("the Wurm is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, wurm.InstanceID, "trample") {
		t.Error("the Wurm has no trample")
	}
}

func TestDgm5564B1PumpSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Fatal Fumes", "Instant", dgmFatalFumesOracle, 1, 3},
		{"Phytoburst", "Sorcery", dgmPhytoburstOracle, 10, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			opp := top100Opponent(g)
			target := b12Creature(g, opp.ID, "Target", dgmBearType, 5, 5)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(target))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, target); p != tc.p || tough != tc.tough {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
		})
	}
}

func TestDgm5564B1MendingTouchShieldsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", dgmBearType, 2, 2)
	castCatalogSpell(t, g, "Mending Touch", "Instant", dgmMendingTouchOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := regenShieldsOn(t, g, bear); got != 1 {
		t.Fatalf("%d regeneration shields, want 1", got)
	}
	b27Kill(g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the shielded creature was destroyed")
	}
}

func TestDgm5564B1ClearAPathDestroysOnlyADefender(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", dgmBearType, 2, 2)
	wall := b12Push(g, opp.ID, "Murmuring Phantasm", "Creature — Spirit", dgmMurmuringPhantasmOracle, 0, 5)
	if err := castCatalogSpellErr(t, g, "Clear a Path", "Sorcery", dgmClearAPathOracle, cardTarget(bear)); err == nil {
		t.Fatal("Clear a Path targeted a creature without defender")
	}
	castCatalogSpell(t, g, "Clear a Path", "Sorcery", dgmClearAPathOracle, cardTarget(wall))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the defender survived")
	}
}

// dgmMindstaticBoard has seat 0 cast a sorcery and seat 1 answer it with
// Mindstatic, stopping at the tax prompt.
func dgmMindstaticBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	static := handCardFull(them, "Mindstatic", "Instant", "", dgmMindstaticOracle, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, static, game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("casting Mindstatic: %v", err)
	}
	passUntilTaxed(t, g, caster.ID)
	return g, victim, caster
}

func TestDgm5564B1MindstaticCountersUnlessSixIsPaid(t *testing.T) {
	g, victim, caster := dgmMindstaticBoard(t)
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{6}" {
		t.Errorf("the tax is %q, want {6}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestDgm5564B1MindstaticLetsThePaidSpellResolve(t *testing.T) {
	g, victim, caster := dgmMindstaticBoard(t)
	for i := 0; i < 6; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	answerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

func TestDgm5564B1PilferedPlansMillsTheTargetAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(opp, "A", "B", "C")
	seedLibrary(me, "D", "E", "F")
	library, hand := opp.Library.Size(), me.Hand.Size()
	castCatalogSpell(t, g, "Pilfered Plans", "Sorcery", dgmPilferedPlansOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("the target milled %d, want 2", got)
	}
	// The spell was put into hand and cast after the count, so the hand
	// is up by exactly the two draws.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestDgm5564B1WarleadersHelixDealsFourAndGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Warleader's Helix", "Instant", dgmWarleadersHelixOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-4 {
		t.Errorf("the target is at %d, want %d", opp.Life, theirs-4)
	}
	if me.Life != mine+4 {
		t.Errorf("you are at %d, want %d", me.Life, mine+4)
	}
}

func TestDgm5564B1ObzedatsAidReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	spell := wthGraveyardCard(me, "Old Spell", "Instant")
	if err := castCatalogSpellErr(t, g, "Obzedat's Aid", "Sorcery", dgmObzedatsAidOracle, cardTarget(spell)); err == nil {
		t.Fatal("Obzedat's Aid targeted an instant card")
	}
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	castCatalogSpell(t, g, "Obzedat's Aid", "Sorcery", dgmObzedatsAidOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the permanent card did not return to the battlefield")
	}
}
