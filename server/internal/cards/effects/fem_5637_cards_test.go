package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fem_5637_cards_test.go — a batch of Fallen Empires cards (#5637).
// Every test plays the real catalog entry through the engine.

const (
	femVodalianSoldiersOracle      = "0e3ac78c-8a98-4786-b86c-ecfe61e13c83"
	femOrderOfLeitburOracle        = "d8f5279f-c58e-4560-9c68-77d226a3c86c"
	femOrderOfTheEbonHandOracle    = "2f65f16f-c6c2-48f0-b556-63cf7d162000"
	femRiverMerfolkOracle          = "2f503ee5-9180-438e-8af5-2af7f8f08a3a"
	femIcatianPriestOracle         = "3a9d29ce-e8e2-402f-b0c2-873f70418234"
	femIcatianScoutOracle          = "b1a9f49a-b0e2-481f-806d-0bc321517a4f"
	femDwarvenLieutenantOracle     = "06e18c42-d158-494d-979c-c2ac074a8ee3"
	femIcatianLieutenantOracle     = "51e9af95-745b-45b9-b870-7f768578fa4f"
	femElvenLyreOracle             = "7601378d-42cb-4351-8316-8f78a3b49a85"
	femAeolipileOracle             = "0897adea-2759-40e8-a05a-c722473e1cf3"
	femHomaridShamanOracle         = "137e3974-74e1-4941-8597-10aeba6cac37"
	femGoblinWarDrumsOracle        = "29c21edd-781b-448e-824a-17bc8b8f4077"
	femImplementsOfSacrificeOracle = "74558981-4226-4961-be33-0af867d0bdf2"
	femBasalThrullOracle           = "9da50130-3f83-4968-983c-7dcba257cf1b"
	femHymnToTourachOracle         = "992d9a72-d34d-47d7-96fb-37f0efb515ef"
	femIcatianTownOracle           = "aa8b60b6-cf55-4aaa-9caa-2b17942d8269"
	femDwarvenRuinsOracle          = "cbef6ad5-717f-4b02-9c62-e5deba407ad1"
	femEbonStrongholdOracle        = "4d4b9512-d29d-4b99-9249-d37e1199a1f5"
	femHavenwoodBattlegroundOracle = "18cbb47a-85b1-48f6-a024-8c3bbffa0d87"
	femRuinsOfTrokairOracle        = "ad7b610c-e276-4c71-9b53-1ccb6d2dafd2"
	femSvyeluniteTempleOracle      = "5578cf33-62f5-456c-a58c-f744a25df79b"
	femDwarvenHoldOracle           = "4a6625bd-3dd2-45f1-8dc9-034c833fa90c"
	femHollowTreesOracle           = "3348df85-e61c-47b5-857d-c79befb38a8a"
	femIcatianStoreOracle          = "87a0e0b9-6c2d-47a4-a3ed-7e0ae62fbffc"
	femSandSilosOracle             = "48a830f1-8965-4f97-b3d8-ca98eab1ba33"
)

// femSetColors gives a test permanent printed colours, for the
// predicates and protections that read them.
func femSetColors(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
	})
}

// --- every card is in the catalog ----------------------------------------

func TestFem5637EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		femVodalianSoldiersOracle, femOrderOfLeitburOracle, femOrderOfTheEbonHandOracle,
		femRiverMerfolkOracle, femIcatianPriestOracle, femIcatianScoutOracle,
		femDwarvenLieutenantOracle, femIcatianLieutenantOracle, femElvenLyreOracle,
		femAeolipileOracle, femHomaridShamanOracle, femGoblinWarDrumsOracle,
		femImplementsOfSacrificeOracle, femBasalThrullOracle, femHymnToTourachOracle,
		femIcatianTownOracle, femDwarvenRuinsOracle, femEbonStrongholdOracle,
		femHavenwoodBattlegroundOracle, femRuinsOfTrokairOracle, femSvyeluniteTempleOracle,
		femDwarvenHoldOracle, femHollowTreesOracle, femIcatianStoreOracle, femSandSilosOracle,
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

// --- the Orders ----------------------------------------------------------

func TestFem5637OrdersHaveProtectionFirstStrikeAndAPump(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, protection, color string
	}{
		{"Order of Leitbur", "Creature — Human Cleric Knight", femOrderOfLeitburOracle, "protection from black", "W"},
		{"Order of the Ebon Hand", "Creature — Cleric Knight", femOrderOfTheEbonHandOracle, "protection from white", "B"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			knight := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			if !hasEffectiveKeyword(t, g, knight, tc.protection) {
				t.Errorf("no %s", tc.protection)
			}
			apaMana(me, tc.color)
			apaActivate(t, g, me, knight, 0, game.ActivateAbilityParams{})
			if !hasEffectiveKeyword(t, g, knight, "first strike") {
				t.Error("no first strike after the first ability")
			}
			apaMana(me, tc.color, tc.color)
			apaActivate(t, g, me, knight, 1, game.ActivateAbilityParams{})
			if p, tt := ptOf(t, g, knight); p != 3 || tt != 1 {
				t.Errorf("after the pump %d/%d, want 3/1", p, tt)
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if hasEffectiveKeyword(t, g, knight, "first strike") {
				t.Error("first strike outlasted the turn")
			}
			if p, tt := ptOf(t, g, knight); p != 2 || tt != 1 {
				t.Errorf("next turn %d/%d, want 2/1", p, tt)
			}
		})
	}
}

// --- River Merfolk ------------------------------------------------------

func TestFem5637RiverMerfolkGainsMountainwalk(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	merfolk := b12Push(g, me.ID, "River Merfolk", "Creature — Merfolk", femRiverMerfolkOracle, 2, 1)
	if hasEffectiveKeyword(t, g, merfolk, "mountainwalk") {
		t.Fatal("mountainwalk before the ability")
	}
	apaMana(me, "U")
	apaActivate(t, g, me, merfolk, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, merfolk, "mountainwalk") {
		t.Error("no mountainwalk after the ability")
	}
}

// --- the target pumps -----------------------------------------------------

func TestFem5637TargetPumps(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		mana                   []string
		power, toughness       int
		keyword                string
		tapped                 bool
	}{
		{"Icatian Priest", "Creature — Human Cleric", femIcatianPriestOracle, []string{"C", "W", "W"}, 3, 3, "", false},
		{"Icatian Scout", "Creature — Human Soldier Scout", femIcatianScoutOracle, []string{"C"}, 2, 2, "first strike", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, src, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
			if p, tt := ptOf(t, g, bear); p != tc.power || tt != tc.toughness {
				t.Errorf("target %d/%d, want %d/%d", p, tt, tc.power, tc.toughness)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("target has no %s", tc.keyword)
			}
			if got := arnTapped(t, g, src); got != tc.tapped {
				t.Errorf("source tapped = %v, want %v", got, tc.tapped)
			}
		})
	}
}

// The Scout's {T} keeps it off the turn it arrives.
func TestFem5637IcatianScoutCantActivateSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	scout := pushCatalogPermanent(g, me.ID, "Icatian Scout", "Creature — Human Soldier Scout", femIcatianScoutOracle, true)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, scout, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Error("a summoning-sick Scout paid {T}")
	}
}

func TestFem5637LieutenantsPumpOnlyTheirSubtype(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, color, ally, other string
	}{
		{"Dwarven Lieutenant", "Creature — Dwarf Soldier", femDwarvenLieutenantOracle, "R", "Creature — Dwarf", "Creature — Goblin"},
		{"Icatian Lieutenant", "Creature — Human Soldier", femIcatianLieutenantOracle, "W", "Creature — Human Soldier", "Creature — Human Cleric"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			lt := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, true)
			ally := b12Creature(g, me.ID, "Ally", tc.ally, 2, 2)
			other := b12Creature(g, me.ID, "Other", tc.other, 2, 2)
			apaMana(me, "C", tc.color)
			if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(other)}); err == nil {
				t.Fatalf("targeted a %s", tc.other)
			}
			// No {T} in the cost: summoning sickness does not matter.
			apaActivate(t, g, me, lt, 0, game.ActivateAbilityParams{Targets: cardTarget(ally)})
			if p, tt := ptOf(t, g, ally); p != 3 || tt != 2 {
				t.Errorf("ally %d/%d, want 3/2", p, tt)
			}
		})
	}
}

// --- the sacrifice artifacts ---------------------------------------------

func TestFem5637ElvenLyreIsSacrificedForPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	lyre := pushCatalogPermanent(g, me.ID, "Elven Lyre", "Artifact", femElvenLyreOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C")
	apaActivate(t, g, me, lyre, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if g.Battlefield.Contains(lyre) {
		t.Error("the Lyre was not sacrificed")
	}
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("target %d/%d, want 4/4", p, tt)
	}
}

func TestFem5637AeolipileIsSacrificedForTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	aeolipile := pushCatalogPermanent(g, me.ID, "Aeolipile", "Artifact", femAeolipileOracle, false)
	life := opp.Life
	apaMana(me, "C")
	apaActivate(t, g, me, aeolipile, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if g.Battlefield.Contains(aeolipile) {
		t.Error("the Aeolipile was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want -2", life, opp.Life)
	}
}

// --- Homarid Shaman -------------------------------------------------------

func TestFem5637HomaridShamanTapsOnlyAGreenCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	shaman := pushCatalogPermanent(g, me.ID, "Homarid Shaman", "Creature — Homarid Shaman", femHomaridShamanOracle, true)
	green := b12Creature(g, opp.ID, "Green", "Creature — Bear", 2, 2)
	red := b12Creature(g, opp.ID, "Red", "Creature — Bear", 2, 2)
	femSetColors(g, green, "G")
	femSetColors(g, red, "R")
	apaMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(red)}); err == nil {
		t.Fatal("Homarid Shaman targeted a creature that is not green")
	}
	apaActivate(t, g, me, shaman, 0, game.ActivateAbilityParams{Targets: cardTarget(green)})
	if !arnTapped(t, g, green) {
		t.Error("the green creature is untapped")
	}
}

// --- Goblin War Drums -----------------------------------------------------

func TestFem5637GoblinWarDrumsGivesYourCreaturesMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Goblin War Drums", "Enchantment", femGoblinWarDrumsOracle, false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Goblin", 1, 1)
	if !hasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("your creature has no menace")
	}
	if hasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("an opponent's creature has menace")
	}
}

// --- the mana sources ----------------------------------------------------

func TestFem5637ImplementsOfSacrificeAddsTwoOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	tools := pushCatalogPermanent(g, me.ID, "Implements of Sacrifice", "Artifact", femImplementsOfSacrificeOracle, false)
	if err := g.ActivateManaAbility(me.ID, tools, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated without the {1}")
	}
	apaMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, tools, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(tools) {
		t.Error("the Implements were not sacrificed")
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour prompt")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatal(err)
	}
	if got := poolColours(me); got["G"] != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool %v, want {G}{G} and nothing else", got)
	}
}

func TestFem5637BasalThrullTapsAndSacrificesForTwoBlack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	sick := pushCatalogPermanent(g, me.ID, "Basal Thrull", "Creature — Thrull", femBasalThrullOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Thrull paid {T}")
	}
	thrull := pushCatalogPermanent(g, me.ID, "Basal Thrull", "Creature — Thrull", femBasalThrullOracle, false)
	if err := g.ActivateManaAbility(me.ID, thrull, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(thrull) {
		t.Error("the Thrull was not sacrificed")
	}
	if got := poolColours(me); got["B"] != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool %v, want {B}{B}", got)
	}
}

// --- the spells ------------------------------------------------------------

func TestFem5637HymnToTourachDiscardsTwoAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	for _, name := range []string{"One", "Two", "Three"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand, yard, myHand := opp.Hand.Size(), opp.Graveyard.Size(), me.Hand.Size()
	// castCatalogSpell puts the Hymn in the caster's hand and casts it.
	castCatalogSpell(t, g, "Hymn to Tourach", "Sorcery", femHymnToTourachOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if n := opp.Hand.Size(); n != hand-2 {
		t.Errorf("target's hand holds %d, want %d", n, hand-2)
	}
	if n := opp.Graveyard.Size(); n != yard+2 {
		t.Errorf("target's graveyard holds %d, want %d", n, yard+2)
	}
	if n := me.Hand.Size(); n != myHand {
		t.Errorf("caster's hand holds %d, want %d", n, myHand)
	}
}

func TestFem5637IcatianTownMakesFourWhiteCitizens(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	castCatalogSpell(t, g, "Icatian Town", "Sorcery", femIcatianTownOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Citizen" || c.Controller != me.ID {
			continue
		}
		n++
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" {
			t.Errorf("token %d/%d %v, want a 1/1 white Citizen", c.Power, c.Toughness, c.Colors)
		}
	}
	if n != 4 {
		t.Errorf("%d Citizens, want 4", n)
	}
}

// --- the sacrifice lands ---------------------------------------------------

var femSacrificeLands = []struct{ name, oracle, color string }{
	{"Dwarven Ruins", femDwarvenRuinsOracle, "R"},
	{"Ebon Stronghold", femEbonStrongholdOracle, "B"},
	{"Havenwood Battleground", femHavenwoodBattlegroundOracle, "G"},
	{"Ruins of Trokair", femRuinsOfTrokairOracle, "W"},
	{"Svyelunite Temple", femSvyeluniteTempleOracle, "U"},
}

func TestFem5637SacrificeLandsEnterTappedTapForOneAndCashInForTwo(t *testing.T) {
	for _, tc := range femSacrificeLands {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			land := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !b12Card(t, g, land).Tapped {
				t.Fatal("did not enter tapped")
			}
			me := g.Seats[g.Turn.ActiveSeat]
			untapForTest(g, land)
			activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
			if got := poolColours(me); got[tc.color] != 1 || me.ManaPool.Total() != 1 {
				t.Fatalf("tap: pool %v, want one {%s}", got, tc.color)
			}
			if !g.Battlefield.Contains(land) {
				t.Fatal("the plain tap sacrificed the land")
			}
			untapForTest(g, land)
			activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{})
			if got := poolColours(me); got[tc.color] != 3 || me.ManaPool.Total() != 3 {
				t.Errorf("sacrifice: pool %v, want three {%s} in all", got, tc.color)
			}
			if g.Battlefield.Contains(land) {
				t.Error("the land was not sacrificed")
			}
		})
	}
}

// --- the storage lands -----------------------------------------------------

var femStorageLands = []struct{ name, oracle, color string }{
	{"Dwarven Hold", femDwarvenHoldOracle, "R"},
	{"Hollow Trees", femHollowTreesOracle, "G"},
	{"Icatian Store", femIcatianStoreOracle, "W"},
	{"Sand Silos", femSandSilosOracle, "U"},
}

func TestFem5637StorageLandsEnterTappedAndPayPerCounterRemoved(t *testing.T) {
	for _, tc := range femStorageLands {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			entered := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !b12Card(t, g, entered).Tapped {
				t.Error("did not enter tapped")
			}
			me := g.Seats[g.Turn.ActiveSeat]
			land := b12Push(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
			addCountersForTest(g, land, game.CounterStorage, 3)
			activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{2}})
			if got := poolColours(me); got[tc.color] != 2 || me.ManaPool.Total() != 2 {
				t.Errorf("pool %v, want two {%s}", got, tc.color)
			}
			if got := counterCount(g, land, game.CounterStorage); got != 1 {
				t.Errorf("counters left %d, want 1", got)
			}
		})
	}
}

// Declining to untap keeps the land tapped, and its controller's upkeep
// banks a storage counter on it; another player's upkeep does not.
func TestFem5637StorageLandsBankACounterWhenKeptTapped(t *testing.T) {
	for _, tc := range femStorageLands {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			land := pushTappedForTest(g, owner.ID, tc.name, tc.oracle, "Land")
			choice := advanceToUntapChoiceOf(t, g, 1)
			if err := g.ResolveUntapChoice(choice.ID, owner.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			advanceToUpkeepOf(t, g, 1)
			passPriorityAroundTable(t, g)
			if !tappedForTest(t, g, land) {
				t.Error("the land untapped though its controller declined")
			}
			if got := counterCount(g, land, game.CounterStorage); got != 1 {
				t.Errorf("storage counters %d, want 1", got)
			}
			advanceToUpkeepOf(t, g, 2)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, land, game.CounterStorage); got != 1 {
				t.Errorf("after another player's upkeep %d counters, want 1", got)
			}
		})
	}
}

// Vodalian Soldiers is vanilla: the catalog knows it and adds nothing.
func TestFem5637VodalianSoldiersIsVanilla(t *testing.T) {
	spec, ok := Lookup(femVodalianSoldiersOracle)
	if !ok {
		t.Fatal("not in the catalog")
	}
	if len(spec.Activated) != 0 || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("a vanilla body declares abilities")
	}
}
