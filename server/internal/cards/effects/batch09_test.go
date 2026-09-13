package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch09_test.go — card-level coverage for the card-coverage
// roadmap's batch 09 (#302, `edhrec_rank` 1005–1110): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b09-prefixed.

const (
	b09DismalBackwaterOracle    = "865a2194-fca0-446e-aae3-ca475cd66e00"
	b09WindScarredCragOracle    = "b0af0c54-2a59-4075-8543-d41ff20c4c87"
	b09SwiftwaterCliffsOracle   = "2f4ad084-2062-44c0-9975-15f100204531"
	b09BlossomingSandsOracle    = "45429b2c-be3b-4b2e-9bab-a059ccbda8cd"
	b09ThornwoodFallsOracle     = "ec96cde2-f1e6-495c-94e2-3e8ae79e556c"
	b09TranquilCoveOracle       = "5d641bf6-0f93-4189-8dc1-ec7ea446dade"
	b09CastleVantressOracle     = "cdf41cf4-4e77-453d-be5b-0abbbd358934"
	b09TalismanOfUnityOracle    = "e5fcc5d7-6a60-4a5b-9d02-6c30041a95b9"
	b09RewindOracle             = "bb27bfdf-fe8d-45bd-ad62-8118dce06eda"
	b09UnwindOracle             = "e38b8ecb-e7ae-474d-b6a4-29cc1aa8ccd9"
	b09ArastaOracle             = "695eea46-1535-48c5-bbb6-0b8379e77bfc"
	b09SetessanChampionOracle   = "38513660-260f-4663-b199-6473fa004a3b"
	b09HourOfReckoningOracle    = "3bc13640-03f8-4b19-b0a4-7e7cb5271c0c"
	b09StaffOfDominationOracle  = "d7888719-647d-4022-a211-822fa09f0791"
	b09SteelshapersGiftOracle   = "d9abda7e-6ca2-42ea-ab24-c542e57014f1"
	b09SpellseekerOracle        = "47a785ed-8095-4685-8daa-02c4e2b0ffcd"
	b09StrixSerenadeOracle      = "7ca3aa03-e62d-464a-9111-e754abe17f76"
	b09SyphonMindOracle         = "abc37d6c-6300-47b5-a679-9db5b83eb54f"
	b09RoilingRegrowthOracle    = "65d63518-3261-40b5-87b8-19c152b29ee3"
	b09BlackcleaveCliffsOracle  = "5ad94412-6f79-4c5d-bbd4-4ef5779a7b6d"
	b09JhoiraOracle             = "c803b788-4213-4fab-b841-7e5bbf66088e"
	b09WallOfOmensOracle        = "5f601f48-d24b-4883-9fde-b3f620e7c9ea"
	b09SharedRootsOracle        = "9e0fd3bf-f47a-4f06-8ff1-73f6bf5d1e03"
	b09DisenchantOracle         = "a7e97fa9-4b72-4548-b854-5be5f18a6f1a"
	b09ArchonOfCrueltyOracle    = "aa1a6646-c1e6-4bff-9092-43ee3e137914"
	b09StoneforgeMysticOracle   = "358789f9-7d87-411d-919e-d597da665cbd"
	b09UpTheBeanstalkOracle     = "050f5733-7c0b-4991-9a6c-7ea12ccf0ca9"
	b09FumigateAlreadyOnMainOID = "b17ea905-0696-4e58-b564-557e87236e27"
)

// b09TryCast puts a card in the active seat's hand at a main phase
// and returns the cast error instead of failing, for the "refuses an
// illegal target" tests.
func b09TryCast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine,
		OracleID: oracle, Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets})
}

// b09OpponentCastsCard is batch01OpponentCasts for an arbitrary card:
// the non-active player casts it at instant speed. A creature gets
// flash on the fixture so the sorcery gate lets it through.
func b09OpponentCastsCard(t *testing.T, g *game.Game, opp *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = opp.ID, opp.ID
	opp.Hand.PushTop(c)
	if err := g.CastSpell(opp.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// b09TappedLands seeds n tapped basic lands under `owner`.
func b09TappedLands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		id := seedLandOnBattlefield(g, owner, "Island", "Basic Land — Island")
		b02bSetTapped(g, id, true)
		out = append(out, id)
	}
	return out
}

// b09UntappedCount counts how many of ids are untapped.
func b09UntappedCount(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if c, ok := battlefieldCard(g, id); ok && !c.Tapped {
			n++
		}
	}
	return n
}

// b09AddColorless floats n colorless mana in a player's pool.
func b09AddColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Eight of them
// are rows in cycle tables (six gain lands, a Talisman, a fastland),
// so a transposed row is invisible until someone plays that exact
// card. Fumigate is the batch's 44th ready card and was already on
// main from the S23 boardwipe pass — pinned here so the table matches
// the issue.
func TestBatch09CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b09DismalBackwaterOracle:    "Dismal Backwater",
		b09WindScarredCragOracle:    "Wind-Scarred Crag",
		b09SwiftwaterCliffsOracle:   "Swiftwater Cliffs",
		b09BlossomingSandsOracle:    "Blossoming Sands",
		b09ThornwoodFallsOracle:     "Thornwood Falls",
		b09TranquilCoveOracle:       "Tranquil Cove",
		b09CastleVantressOracle:     "Castle Vantress",
		b09TalismanOfUnityOracle:    "Talisman of Unity",
		b09RewindOracle:             "Rewind",
		b09UnwindOracle:             "Unwind",
		b09ArastaOracle:             "Arasta of the Endless Web",
		b09SetessanChampionOracle:   "Setessan Champion",
		b09HourOfReckoningOracle:    "Hour of Reckoning",
		b09StaffOfDominationOracle:  "Staff of Domination",
		b09SteelshapersGiftOracle:   "Steelshaper's Gift",
		b09SpellseekerOracle:        "Spellseeker",
		b09StrixSerenadeOracle:      "Strix Serenade",
		b09SyphonMindOracle:         "Syphon Mind",
		b09RoilingRegrowthOracle:    "Roiling Regrowth",
		b09BlackcleaveCliffsOracle:  "Blackcleave Cliffs",
		b09JhoiraOracle:             "Jhoira, Weatherlight Captain",
		b09WallOfOmensOracle:        "Wall of Omens",
		b09SharedRootsOracle:        "Shared Roots",
		b09DisenchantOracle:         "Disenchant",
		b09ArchonOfCrueltyOracle:    "Archon of Cruelty",
		b09StoneforgeMysticOracle:   "Stoneforge Mystic",
		b09UpTheBeanstalkOracle:     "Up the Beanstalk",
		b09FumigateAlreadyOnMainOID: "Fumigate",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 27 cards plus Fumigate, the table lists %d", len(want))
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
		// The fastland table predates the Completeness field and
		// this batch adds a row only (the Vernal Fen precedent);
		// Fumigate is #382's file, not this batch's.
		if spec.Completeness == CompletenessUnreviewed && name != "Blackcleave Cliffs" && name != "Fumigate" {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- the cycle rows ------------------------------------------------

func TestB09GainLandsEnterTappedGainOneAndTapForBoth(t *testing.T) {
	for _, tc := range []struct{ name, oracle, a, b string }{
		{"Dismal Backwater", b09DismalBackwaterOracle, "U", "B"},
		{"Wind-Scarred Crag", b09WindScarredCragOracle, "R", "W"},
		{"Swiftwater Cliffs", b09SwiftwaterCliffsOracle, "U", "R"},
		{"Blossoming Sands", b09BlossomingSandsOracle, "G", "W"},
		{"Thornwood Falls", b09ThornwoodFallsOracle, "G", "U"},
		{"Tranquil Cove", b09TranquilCoveOracle, "W", "U"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		id := playLandFromHand(t, g, tc.name, tc.oracle)
		top100AssertEnteredTapped(t, g, id, tc.name)
		if triggerOnStack(g, id) == nil {
			t.Errorf("%s: the life is a trigger with a response window", tc.name)
		}
		passPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

func TestB09TalismanOfUnityColoredHalfHurtsColorlessHalfDoesNot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := seedPermanentWithOracle(g, me.ID, "Talisman of Unity", "Artifact", b09TalismanOfUnityOracle)
	before := me.Life

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colorless half: %v", err)
	}
	if me.Life != before {
		t.Errorf("the {C} half hurt: %d → %d", before, me.Life)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	b02bSetTapped(g, rock, false)
	if err := g.ActivateManaAbility(me.ID, rock, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colored half: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the colored half must ask G or W, got %+v", pick)
	}
	spec, _ := Lookup(b09TalismanOfUnityOracle)
	if spec.ManaAbilities[1].Produced != "{G|W}" {
		t.Errorf("row produces %s, want {G|W}", spec.ManaAbilities[1].Produced)
	}
}

func TestB09BlackcleaveCliffsIsAFastland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	early := playLandFromHand(t, g, "Blackcleave Cliffs", b09BlackcleaveCliffsOracle)
	top100AssertEnteredUntapped(t, g, early, "Blackcleave Cliffs with two other lands")

	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	late := playLandFromHand(t, g, "Blackcleave Cliffs", b09BlackcleaveCliffsOracle)
	top100AssertEnteredTapped(t, g, late, "Blackcleave Cliffs with three other lands")

	spec, _ := Lookup(b09BlackcleaveCliffsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|R}" {
		t.Errorf("row produces %+v, want {B|R}", spec.ManaAbilities)
	}
}

// --- Castle Vantress -----------------------------------------------

func TestB09CastleVantressEntersTappedWithoutAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	tapped := playLandFromHand(t, g, "Castle Vantress", b09CastleVantressOracle)
	top100AssertEnteredTapped(t, g, tapped, "Castle Vantress without an Island")

	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	untapped := playLandFromHand(t, g, "Castle Vantress", b09CastleVantressOracle)
	top100AssertEnteredUntapped(t, g, untapped, "Castle Vantress with an Island")
}

func TestB09CastleVantressScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castle := pushCatalogPermanent(g, me.ID, "Castle Vantress", "Land", b09CastleVantressOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"})
	b09AddColorless(me, 2)

	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := battlefieldCard(g, castle); !c.Tapped {
		t.Error("the scry has a tap cost")
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("the scry must wait for the ability to resolve")
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a scry-2 prompt, got %+v", c)
	}
}

// --- Rewind / Unwind -----------------------------------------------

func TestB09RewindCountersAndUntapsFourOfYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b09TappedLands(g, me.ID, 5)
	theirs := b09TappedLands(g, opp.ID, 2)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	before := me.Life

	castCatalogSpell(t, g, "Rewind", "Instant", b09RewindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) || me.Life != before {
		t.Fatal("the Bolt was not countered")
	}
	if got := b09UntappedCount(g, mine); got != 4 {
		t.Errorf("%d of my lands untapped, want 4", got)
	}
	if got := b09UntappedCount(g, theirs); got != 0 {
		t.Errorf("an opponent's land was untapped")
	}
}

func TestB09UnwindCountersANoncreatureAndUntapsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b09TappedLands(g, me.ID, 5)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})

	castCatalogSpell(t, g, "Unwind", "Instant", b09UnwindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if got := b09UntappedCount(g, mine); got != 3 {
		t.Errorf("%d of my lands untapped, want 3", got)
	}
}

func TestB09UnwindRefusesACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b09OpponentCastsCard(t, g, opp, game.Card{Name: "Ambusher", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}})
	if err := b09TryCast(t, g, "Unwind", "Instant", b09UnwindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature spell was accepted for 'target noncreature spell'")
	}
}

// --- Arasta of the Endless Web ------------------------------------

func TestB09ArastaMakesASpiderWhenAnOpponentCastsAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arasta := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Arasta of the Endless Web",
		TypeLine: "Legendary Enchantment Creature — Spider", OracleID: b09ArastaOracle,
		Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if !eotHasAbility(effectiveAbilities(t, g, arasta), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}

	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if triggerOnStack(g, arasta) == nil {
		t.Fatal("the Spider trigger should be on the stack above the Bolt")
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Spider") != 1 {
		t.Fatal("an opponent's instant should make one Spider")
	}
	spider := findBattlefieldByName(g, "Spider")
	if !eotHasAbility(effectiveAbilities(t, g, spider), "reach") {
		t.Error("the Spider token should have reach")
	}

	// The controller's own instant is not "an opponent casts".
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Spider") != 1 {
		t.Error("my own instant made a Spider")
	}
}

// --- Setessan Champion --------------------------------------------

func TestB09SetessanChampionGrowsAndDrawsOnYourEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Setessan Champion", TypeLine: "Creature — Human Warrior",
		OracleID: b09SetessanChampionOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, champion, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want +1", before, got)
	}

	// An opponent's enchantment and a creature of mine are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Glory", TypeLine: "Token Enchantment"}, 1)
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, champion, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d after non-triggering entries, want still 1", got)
	}
}

// --- Hour of Reckoning --------------------------------------------

func TestB09HourOfReckoningSparesTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := seedCreature(g, "My Bear", me.ID)
	theirBear := seedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2)
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	spec, _ := Lookup(b09HourOfReckoningOracle)
	if spec.TapCost == nil {
		t.Fatal("Hour of Reckoning has convoke")
	}

	castCatalogSpell(t, g, "Hour of Reckoning", "Sorcery", b09HourOfReckoningOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{myBear, theirBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("nontoken creature %s survived", id)
		}
	}
	if countBattlefieldNamed(g, me.ID, "Goblin") != 2 || countBattlefieldNamed(g, opp.ID, "Goblin") != 1 {
		t.Error("tokens must survive Hour of Reckoning")
	}
}

// --- Staff of Domination ------------------------------------------

func TestB09StaffOfDominationTapsUntapsItselfAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := pushCatalogPermanent(g, me.ID, "Staff of Domination", "Artifact", b09StaffOfDominationOracle, false)
	bear := seedCreature(g, "Bear", opp.ID)

	// {4}, {T}: Tap target creature.
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, staff, 3, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if c, _ := battlefieldCard(g, staff); !c.Tapped {
		t.Fatal("the Staff should be tapped after a {T} ability")
	}

	// {1}: Untap this artifact.
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("untap self: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, staff); c.Tapped {
		t.Fatal("the Staff did not untap itself")
	}

	// {3}, {T}: Untap target creature.
	b09AddColorless(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, staff, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("untap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the creature was not untapped")
	}

	// {1}: untap again, then {5}, {T}: Draw a card.
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("untap self again: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	b09AddColorless(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, staff, 4, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want +1", before, got)
	}
}

func TestB09StaffOfDominationGainsOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	staff := pushCatalogPermanent(g, me.ID, "Staff of Domination", "Artifact", b09StaffOfDominationOracle, false)
	before := me.Life
	b09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, staff, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("life ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	// Tapped, and without the {1} it stays that way.
	if err := g.ActivateCatalogAbility(me.ID, staff, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Staff activated a {T} ability")
	}
}

// --- the Equipment tutors -----------------------------------------

func TestB09SteelshapersGiftOffersOnlyEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Boots", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Steelshaper's Gift", "Sorcery", b09SteelshapersGiftOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Equipment card was offered")
	}
	answerSearchNamed(t, g, me.ID, "Sword")
	if !b02bHandHasNamed(me, "Sword") {
		t.Error("the chosen Equipment did not reach the hand")
	}
}

func TestB09StoneforgeMysticTutorsAnEquipmentAndHasNoActivatedAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	castAndResolveCreature(t, g, "Stoneforge Mystic", "Creature — Kor Artificer", b09StoneforgeMysticOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt — 'you may search' must still ask")
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact was offered")
	}
	answerSearchNamed(t, g, me.ID, "Sword")
	if !b02bHandHasNamed(me, "Sword") {
		t.Error("the Equipment did not reach the hand")
	}
	spec, _ := Lookup(b09StoneforgeMysticOracle)
	if len(spec.Activated) != 0 || spec.Completeness != CompletenessCaveats {
		t.Error("the put-onto-the-battlefield ability is a declared omission, not a half-built ability")
	}
}

// --- Spellseeker ---------------------------------------------------

func TestB09SpellseekerOffersCheapInstantsAndSorceriesAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{U}"},
		game.Card{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{B}"},
		game.Card{Name: "Dear Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
		game.Card{Name: "Cheap Creature", TypeLine: "Creature — Bear", ManaCost: "{G}"},
	)
	castAndResolveCreature(t, g, "Spellseeker", "Creature — Human Wizard", b09SpellseekerOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, offered := range []string{"Cheap Instant", "Cheap Sorcery"} {
		if searchOptionNamed(g, c, offered) == uuid.Nil {
			t.Errorf("%s should be offered", offered)
		}
	}
	for _, refused := range []string{"Dear Sorcery", "Cheap Creature"} {
		if searchOptionNamed(g, c, refused) != uuid.Nil {
			t.Errorf("%s should not be offered", refused)
		}
	}
	before := me.Hand.Size()
	answerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != before {
		t.Error("declining the search put a card in hand")
	}
}

// --- Strix Serenade -----------------------------------------------

func TestB09StrixSerenadeCountersACreatureAndGivesItsControllerABird(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b09OpponentCastsCard(t, g, opp, game.Card{Name: "Ambusher", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}})

	castCatalogSpell(t, g, "Strix Serenade", "Instant", b09StrixSerenadeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bear) {
		t.Fatal("the creature spell was not countered")
	}
	if countBattlefieldNamed(g, opp.ID, "Bird") != 1 {
		t.Error("the countered spell's CONTROLLER should get the Bird")
	}
	if countBattlefieldNamed(g, me.ID, "Bird") != 0 {
		t.Error("the caster must not get the Bird")
	}
	bird := findBattlefieldByName(g, "Bird")
	if !eotHasAbility(effectiveAbilities(t, g, bird), "flying") {
		t.Error("the Bird should fly")
	}
}

func TestB09StrixSerenadeRefusesAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := b09TryCast(t, g, "Strix Serenade", "Instant", b09StrixSerenadeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Error("an instant was accepted for 'target artifact, creature, or planeswalker spell'")
	}
}

// --- Syphon Mind ---------------------------------------------------

func TestB09SyphonMindDrawsOnePerOpponentWhoDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	emptyHandToLibrary(g, c)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Syphon Mind", "Sorcery", b09SyphonMindOracle, nil)
	passPriorityAroundTable(t, g)

	// Two opponents hold cards, so two draws (the cast itself is
	// hand-neutral: castCatalogSpell adds the card it then casts).
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d → %d, want +2", before, got)
	}
	if g.DiscardPending[a.ID] != 1 || g.DiscardPending[b.ID] != 1 {
		t.Errorf("each opponent with a hand owes one discard: %v", g.DiscardPending)
	}
	if _, owes := g.DiscardPending[c.ID]; owes {
		t.Error("an opponent with no hand was asked to discard")
	}
	if _, owes := g.DiscardPending[me.ID]; owes {
		t.Error("the caster was asked to discard")
	}
}

// --- Roiling Regrowth ---------------------------------------------

func TestB09RoilingRegrowthSacrificesAtCastAndFetchesTwoTappedBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest B", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest C", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	castWithSacrifice(t, g, "Roiling Regrowth", "Instant", b09RoilingRegrowthOracle, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice is a cost and is paid at announce")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 {
		t.Errorf("search max %d, want 2", c.SearchMax)
	}
	if searchOptionNamed(g, c, "Bayou") != uuid.Nil {
		t.Error("a nonbasic Forest was offered for 'basic land card'")
	}
	answerSearchNamed(t, g, me.ID, "Forest A", "Forest B")
	for _, name := range []string{"Forest A", "Forest B"} {
		card, ok := searchFetchedCard(g, name)
		if !ok {
			t.Errorf("%s did not reach the battlefield", name)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s entered untapped", name)
		}
	}
}

// --- Jhoira, Weatherlight Captain ---------------------------------

func TestB09JhoiraDrawsOnHistoricSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jhoira, Weatherlight Captain",
		TypeLine: "Legendary Creature — Human Artificer", OracleID: b09JhoiraOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})

	before := me.Hand.Size()
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("an artifact spell: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	castCatalogSpell(t, g, "Legend", "Legendary Creature — Human", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("a legendary spell: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("a plain creature: hand %d → %d, want no draw", before, got)
	}
}

func TestB09HistoricReadsArtifactsLegendsAndSagas(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     bool
	}{
		{"Artifact", true},
		{"Artifact Creature — Construct", true},
		{"Legendary Creature — Human", true},
		{"Legendary Enchantment", true},
		{"Enchantment — Saga", true},
		{"Creature — Bear", false},
		{"Instant", false},
		{"Basic Land — Forest", false},
	} {
		if got := b09IsHistoric(game.Card{TypeLine: tc.typeLine}); got != tc.want {
			t.Errorf("%q historic = %v, want %v", tc.typeLine, got, tc.want)
		}
	}
}

// --- Wall of Omens / Shared Roots / Disenchant ---------------------

func TestB09WallOfOmensDrawsOnEntryAndDefends(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	wall := castAndResolveCreature(t, g, "Wall of Omens", "Creature — Wall", b09WallOfOmensOracle)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want +1", before, got)
	}
	if !eotHasAbility(effectiveAbilities(t, g, wall), "defender") {
		t.Error("printed defender did not reach the effective abilities")
	}
}

func TestB09SharedRootsFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island A", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	castCatalogSpell(t, g, "Shared Roots", "Sorcery — Lesson", b09SharedRootsOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bayou") != uuid.Nil {
		t.Error("a nonbasic was offered")
	}
	answerSearchNamed(t, g, me.ID, "Island A")
	card, ok := searchFetchedCard(g, "Island A")
	if !ok || !card.Tapped {
		t.Errorf("the basic should be on the battlefield tapped, got %+v (found %v)", card, ok)
	}
}

func TestB09DisenchantDestroysAnEnchantmentAndRefusesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := seedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	castCatalogSpell(t, g, "Disenchant", "Instant", b09DisenchantOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the enchantment should be in its owner's graveyard")
	}

	bear := seedCreature(g, "Bear", opp.ID)
	if err := b09TryCast(t, g, "Disenchant", "Instant", b09DisenchantOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted for 'target artifact or enchantment'")
	}
}

// --- Archon of Cruelty --------------------------------------------

// b09AssertArchonResolved checks everything one Archon trigger does
// to `victim` and for `me`, given the life totals and hand size
// before the trigger resolved.
func b09AssertArchonResolved(t *testing.T, g *game.Game, me, victim *game.Player, meLife, victimLife, meHand int, theirBear uuid.UUID) {
	t.Helper()
	if victim.Life != victimLife-3 {
		t.Errorf("victim life %d → %d, want -3", victimLife, victim.Life)
	}
	if me.Life != meLife+3 {
		t.Errorf("my life %d → %d, want +3", meLife, me.Life)
	}
	if got := me.Hand.Size(); got != meHand+1 {
		t.Errorf("my hand %d → %d, want +1", meHand, got)
	}
	c := sacrificeChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the victim should be choosing a creature or planeswalker to sacrifice")
	}
	if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != theirBear {
		t.Errorf("sacrifice offered %v, want just the victim's own creature %v", c.SacrificeOptions, theirBear)
	}
	if g.DiscardPending[victim.ID] != 1 {
		t.Errorf("the victim owes one discard: %v", g.DiscardPending)
	}
}

func TestB09ArchonOfCrueltyEntersAgainstATargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	theirBear := seedCreature(g, "Their Bear", victim.ID)
	seedCreature(g, "Bystander Bear", bystander.ID)
	meLife, victimLife, bystanderLife := me.Life, victim.Life, bystander.Life

	archon := castAndResolveCreature(t, g, "Archon of Cruelty", "Creature — Archon", b09ArchonOfCrueltyOracle)
	if !eotHasAbility(effectiveAbilities(t, g, archon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	meHand := me.Hand.Size()
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the trigger targets an opponent and must ask which")
	}
	if hasID(prompt.PickTargetPlayers, me.ID) {
		t.Error("the controller is not an opponent")
	}
	pickPlayer(t, g, me.ID, victim.ID)
	if triggerOnStack(g, archon) == nil {
		t.Fatal("the trigger should be on the stack after the pick")
	}
	passPriorityAroundTable(t, g)

	b09AssertArchonResolved(t, g, me, victim, meLife, victimLife, meHand, theirBear)
	if bystander.Life != bystanderLife || sacrificeChoiceFor(g, bystander.ID) != nil {
		t.Error("the untargeted opponent was touched")
	}
	answerSacrifice(t, g, victim.ID, theirBear)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the chosen creature was not sacrificed")
	}
}

func TestB09ArchonOfCrueltyAttacksAgainstATargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	archon := pushDiesCreatureForTest(g, me.ID, "Archon of Cruelty", b09ArchonOfCrueltyOracle, "Creature — Archon", 6, 6)
	theirBear := seedCreature(g, "Their Bear", victim.ID)
	meLife, victimLife, meHand := me.Life, victim.Life, me.Hand.Size()

	declareAttack(t, g, victim.ID, archon)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the attack trigger must ask for its opponent")
	}
	pickPlayer(t, g, me.ID, victim.ID)
	if triggersOnStackFrom(g, archon) != 1 {
		t.Fatalf("one attacker, %d triggers on the stack", triggersOnStackFrom(g, archon))
	}
	passPriorityAroundTable(t, g)
	b09AssertArchonResolved(t, g, me, victim, meLife, victimLife, meHand, theirBear)
}

// --- Up the Beanstalk ---------------------------------------------

func TestB09UpTheBeanstalkDrawsOnEntryAndOnFiveDrops(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Up the Beanstalk", "Enchantment", b09UpTheBeanstalkOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("entry: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	castWithCost(t, g, "Big", "Creature — Giant", "{4}{G}", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("a five-drop: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	castWithCost(t, g, "Small", "Creature — Bear", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("a four-drop: hand %d → %d, want no draw", before, got)
	}

	// X counts at its announced value on the stack (CR 202.3e).
	before = me.Hand.Size()
	castXSpell(t, g, "Fireball", "Sorcery", "", "{X}{R}", 4, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("an X=4 {X}{R} spell: hand %d → %d, want +1", before, got)
	}
}
