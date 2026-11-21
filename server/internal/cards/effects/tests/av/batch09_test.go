package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// b09TappedLands seeds n tapped basic lands under `owner`.
func b09TappedLands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		id := effectstest.SeedLandOnBattlefield(g, owner, "Island", "Basic Land — Island")
		effectstest.B02bSetTapped(g, id, true)
		out = append(out, id)
	}
	return out
}

// b09UntappedCount counts how many of ids are untapped.
func b09UntappedCount(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if c, ok := effectstest.BattlefieldCard(g, id); ok && !c.Tapped {
			n++
		}
	}
	return n
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
		effectstest.B09DismalBackwaterOracle:    "Dismal Backwater",
		effectstest.B09WindScarredCragOracle:    "Wind-Scarred Crag",
		effectstest.B09SwiftwaterCliffsOracle:   "Swiftwater Cliffs",
		effectstest.B09BlossomingSandsOracle:    "Blossoming Sands",
		effectstest.B09ThornwoodFallsOracle:     "Thornwood Falls",
		effectstest.B09TranquilCoveOracle:       "Tranquil Cove",
		effectstest.B09CastleVantressOracle:     "Castle Vantress",
		effectstest.B09TalismanOfUnityOracle:    "Talisman of Unity",
		effectstest.B09RewindOracle:             "Rewind",
		effectstest.B09UnwindOracle:             "Unwind",
		effectstest.B09ArastaOracle:             "Arasta of the Endless Web",
		effectstest.B09SetessanChampionOracle:   "Setessan Champion",
		effectstest.B09HourOfReckoningOracle:    "Hour of Reckoning",
		effectstest.B09StaffOfDominationOracle:  "Staff of Domination",
		effectstest.B09SteelshapersGiftOracle:   "Steelshaper's Gift",
		effectstest.B09SpellseekerOracle:        "Spellseeker",
		effectstest.B09StrixSerenadeOracle:      "Strix Serenade",
		effectstest.B09SyphonMindOracle:         "Syphon Mind",
		effectstest.B09RoilingRegrowthOracle:    "Roiling Regrowth",
		effectstest.B09BlackcleaveCliffsOracle:  "Blackcleave Cliffs",
		effectstest.B09JhoiraOracle:             "Jhoira, Weatherlight Captain",
		effectstest.B09WallOfOmensOracle:        "Wall of Omens",
		effectstest.B09SharedRootsOracle:        "Shared Roots",
		effectstest.B09DisenchantOracle:         "Disenchant",
		effectstest.B09ArchonOfCrueltyOracle:    "Archon of Cruelty",
		effectstest.B09StoneforgeMysticOracle:   "Stoneforge Mystic",
		effectstest.B09UpTheBeanstalkOracle:     "Up the Beanstalk",
		effectstest.B09FumigateAlreadyOnMainOID: "Fumigate",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 27 cards plus Fumigate, the table lists %d", len(want))
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
		// The fastland table predates the Completeness field and
		// this batch adds a row only (the Vernal Fen precedent);
		// Fumigate is #382's file, not this batch's.
		if spec.Completeness == effects.CompletenessUnreviewed && name != "Blackcleave Cliffs" && name != "Fumigate" {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
}

// --- the cycle rows ------------------------------------------------

func TestB09GainLandsEnterTappedGainOneAndTapForBoth(t *testing.T) {
	for _, tc := range []struct{ name, oracle, a, b string }{
		{"Dismal Backwater", effectstest.B09DismalBackwaterOracle, "U", "B"},
		{"Wind-Scarred Crag", effectstest.B09WindScarredCragOracle, "R", "W"},
		{"Swiftwater Cliffs", effectstest.B09SwiftwaterCliffsOracle, "U", "R"},
		{"Blossoming Sands", effectstest.B09BlossomingSandsOracle, "G", "W"},
		{"Thornwood Falls", effectstest.B09ThornwoodFallsOracle, "G", "U"},
		{"Tranquil Cove", effectstest.B09TranquilCoveOracle, "W", "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		id := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		effectstest.Top100AssertEnteredTapped(t, g, id, tc.name)
		if effectstest.TriggerOnStack(g, id) == nil {
			t.Errorf("%s: the life is a trigger with a response window", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

func TestB09TalismanOfUnityColoredHalfHurtsColorlessHalfDoesNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.SeedPermanentWithOracle(g, me.ID, "Talisman of Unity", "Artifact", effectstest.B09TalismanOfUnityOracle)
	before := me.Life

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colorless half: %v", err)
	}
	if me.Life != before {
		t.Errorf("the {C} half hurt: %d → %d", before, me.Life)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	effectstest.B02bSetTapped(g, rock, false)
	if err := g.ActivateManaAbility(me.ID, rock, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("colored half: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the colored half must ask G or W, got %+v", pick)
	}
	spec, _ := effects.Lookup(effectstest.B09TalismanOfUnityOracle)
	if spec.ManaAbilities[1].Produced != "{G|W}" {
		t.Errorf("row produces %s, want {G|W}", spec.ManaAbilities[1].Produced)
	}
}

func TestB09BlackcleaveCliffsIsAFastland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	early := effectstest.PlayLandFromHand(t, g, "Blackcleave Cliffs", effectstest.B09BlackcleaveCliffsOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, early, "Blackcleave Cliffs with two other lands")

	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	late := effectstest.PlayLandFromHand(t, g, "Blackcleave Cliffs", effectstest.B09BlackcleaveCliffsOracle)
	effectstest.Top100AssertEnteredTapped(t, g, late, "Blackcleave Cliffs with three other lands")

	spec, _ := effects.Lookup(effectstest.B09BlackcleaveCliffsOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{B|R}" {
		t.Errorf("row produces %+v, want {B|R}", spec.ManaAbilities)
	}
}

// --- Castle Vantress -----------------------------------------------

func TestB09CastleVantressEntersTappedWithoutAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	tapped := effectstest.PlayLandFromHand(t, g, "Castle Vantress", effectstest.B09CastleVantressOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "Castle Vantress without an Island")

	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	untapped := effectstest.PlayLandFromHand(t, g, "Castle Vantress", effectstest.B09CastleVantressOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, untapped, "Castle Vantress with an Island")
}

func TestB09CastleVantressScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	castle := effectstest.PushCatalogPermanent(g, me.ID, "Castle Vantress", "Land", effectstest.B09CastleVantressOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"})
	effectstest.B09AddColorless(me, 2)

	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, castle); !c.Tapped {
		t.Error("the scry has a tap cost")
	}
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("the scry must wait for the ability to resolve")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a scry-2 prompt, got %+v", c)
	}
}

// --- Rewind / Unwind -----------------------------------------------

func TestB09RewindCountersAndUntapsFourOfYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b09TappedLands(g, me.ID, 5)
	theirs := b09TappedLands(g, opp.ID, 2)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	before := me.Life

	effectstest.CastCatalogSpell(t, g, "Rewind", "Instant", effectstest.B09RewindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) || me.Life != before {
		t.Fatal("the Bolt was not countered")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no untap prompt: \"untap up to four lands\" is a resolution-time choice")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 4 {
		t.Errorf("bounds %d..%d, want 0..4", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[:4]...) // 5 tapped lands offered, pick 4
	if got := b09UntappedCount(g, mine); got != 4 {
		t.Errorf("%d of my lands untapped, want 4", got)
	}
	if got := b09UntappedCount(g, theirs); got != 0 {
		t.Errorf("an opponent's land was untapped")
	}
}

func TestB09UnwindCountersANoncreatureAndUntapsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b09TappedLands(g, me.ID, 5)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})

	effectstest.CastCatalogSpell(t, g, "Unwind", "Instant", effectstest.B09UnwindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[:3]...) // 5 tapped lands offered, pick 3
	if got := b09UntappedCount(g, mine); got != 3 {
		t.Errorf("%d of my lands untapped, want 3", got)
	}
}

func TestB09UnwindRefusesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Ambusher", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}})
	if err := b09TryCast(t, g, "Unwind", "Instant", effectstest.B09UnwindOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature spell was accepted for 'target noncreature spell'")
	}
}

// --- Arasta of the Endless Web ------------------------------------

func TestB09ArastaMakesASpiderWhenAnOpponentCastsAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arasta := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Arasta of the Endless Web",
		TypeLine: "Legendary Enchantment Creature — Spider", OracleID: effectstest.B09ArastaOracle,
		Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, arasta), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}

	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.TriggerOnStack(g, arasta) == nil {
		t.Fatal("the Spider trigger should be on the stack above the Bolt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Spider") != 1 {
		t.Fatal("an opponent's instant should make one Spider")
	}
	spider := effectstest.FindBattlefieldByName(g, "Spider")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, spider), "reach") {
		t.Error("the Spider token should have reach")
	}

	// The controller's own instant is not "an opponent casts".
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Spider") != 1 {
		t.Error("my own instant made a Spider")
	}
}

// --- Setessan Champion --------------------------------------------

func TestB09SetessanChampionGrowsAndDrawsOnYourEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Setessan Champion", TypeLine: "Creature — Human Warrior",
		OracleID: effectstest.B09SetessanChampionOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, champion, "+1/+1"); got != 1 {
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
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, champion, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d after non-triggering entries, want still 1", got)
	}
}

// --- Hour of Reckoning --------------------------------------------

func TestB09HourOfReckoningSparesTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirBear := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2)
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	spec, _ := effects.Lookup(effectstest.B09HourOfReckoningOracle)
	if spec.TapCost == nil {
		t.Fatal("Hour of Reckoning has convoke")
	}

	effectstest.CastCatalogSpell(t, g, "Hour of Reckoning", "Sorcery", effectstest.B09HourOfReckoningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{myBear, theirBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("nontoken creature %s survived", id)
		}
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Goblin") != 2 || effectstest.CountBattlefieldNamed(g, opp.ID, "Goblin") != 1 {
		t.Error("tokens must survive Hour of Reckoning")
	}
}

// --- Staff of Domination ------------------------------------------

func TestB09StaffOfDominationTapsUntapsItselfAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := effectstest.PushCatalogPermanent(g, me.ID, "Staff of Domination", "Artifact", effectstest.B09StaffOfDominationOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)

	// {4}, {T}: Tap target creature.
	effectstest.B09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, staff, 3, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("tap ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, staff); !c.Tapped {
		t.Fatal("the Staff should be tapped after a {T} ability")
	}

	// {1}: Untap this artifact.
	effectstest.B09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("untap self: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, staff); c.Tapped {
		t.Fatal("the Staff did not untap itself")
	}

	// {3}, {T}: Untap target creature.
	effectstest.B09AddColorless(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, staff, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("untap ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("the creature was not untapped")
	}

	// {1}: untap again, then {5}, {T}: Draw a card.
	effectstest.B09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("untap self again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()
	effectstest.B09AddColorless(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, staff, 4, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want +1", before, got)
	}
}

func TestB09StaffOfDominationGainsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := effectstest.PushCatalogPermanent(g, me.ID, "Staff of Domination", "Artifact", effectstest.B09StaffOfDominationOracle, false)
	before := me.Life
	effectstest.B09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, staff, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("life ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Boots", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Steelshaper's Gift", "Sorcery", effectstest.B09SteelshapersGiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Equipment card was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sword")
	if !effectstest.B02bHandHasNamed(me, "Sword") {
		t.Error("the chosen Equipment did not reach the hand")
	}
}

func TestB09StoneforgeMysticTutorsAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
	)
	effectstest.CastAndResolveCreature(t, g, "Stoneforge Mystic", "Creature — Kor Artificer", effectstest.B09StoneforgeMysticOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt — 'you may search' must still ask")
	}
	if effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sword")
	if !effectstest.B02bHandHasNamed(me, "Sword") {
		t.Error("the Equipment did not reach the hand")
	}
	// The {1}{W},{T} ability landed with #3138 (token_purposes_and_gaps_3138_test.go).
	spec, _ := effects.Lookup(effectstest.B09StoneforgeMysticOracle)
	if len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("Stoneforge Mystic should carry its activated ability and grade full")
	}
}

// --- Spellseeker ---------------------------------------------------

func TestB09SpellseekerOffersCheapInstantsAndSorceriesAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{U}"},
		game.Card{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{B}"},
		game.Card{Name: "Dear Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
		game.Card{Name: "Cheap Creature", TypeLine: "Creature — Bear", ManaCost: "{G}"},
	)
	effectstest.CastAndResolveCreature(t, g, "Spellseeker", "Creature — Human Wizard", effectstest.B09SpellseekerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, offered := range []string{"Cheap Instant", "Cheap Sorcery"} {
		if effectstest.SearchOptionNamed(g, c, offered) == uuid.Nil {
			t.Errorf("%s should be offered", offered)
		}
	}
	for _, refused := range []string{"Dear Sorcery", "Cheap Creature"} {
		if effectstest.SearchOptionNamed(g, c, refused) != uuid.Nil {
			t.Errorf("%s should not be offered", refused)
		}
	}
	before := me.Hand.Size()
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != before {
		t.Error("declining the search put a card in hand")
	}
}

// --- Strix Serenade -----------------------------------------------

func TestB09StrixSerenadeCountersACreatureAndGivesItsControllerABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Ambusher", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"flash"}})

	effectstest.CastCatalogSpell(t, g, "Strix Serenade", "Instant", effectstest.B09StrixSerenadeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bear) {
		t.Fatal("the creature spell was not countered")
	}
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Bird") != 1 {
		t.Error("the countered spell's CONTROLLER should get the Bird")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Bird") != 0 {
		t.Error("the caster must not get the Bird")
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bird), "flying") {
		t.Error("the Bird should fly")
	}
}

func TestB09StrixSerenadeRefusesAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := b09TryCast(t, g, "Strix Serenade", "Instant", effectstest.B09StrixSerenadeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Error("an instant was accepted for 'target artifact, creature, or planeswalker spell'")
	}
}

// --- Syphon Mind ---------------------------------------------------

func TestB09SyphonMindDrawsOnePerOpponentWhoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.EmptyHandToLibrary(g, c)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Syphon Mind", "Sorcery", effectstest.B09SyphonMindOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// #651: the draw is "a card for each card discarded THIS WAY", so
	// it rides each prompt's continuation — nothing is drawn until an
	// opponent has actually pitched. Two opponents hold cards, so two
	// draws once both have answered (the cast itself is hand-neutral:
	// castCatalogSpell adds the card it then casts).
	if got := me.Hand.Size(); got != before {
		t.Errorf("the draw waits for a discard: hand %d → %d", before, got)
	}
	if effectstest.DiscardOwed(g, a.ID) != 1 || effectstest.DiscardOwed(g, b.ID) != 1 {
		t.Errorf("each opponent with a hand owes one discard: %d and %d",
			effectstest.DiscardOwed(g, a.ID), effectstest.DiscardOwed(g, b.ID))
	}
	if effectstest.DiscardOwed(g, c.ID) != 0 {
		t.Error("an opponent with no hand was asked to discard")
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Error("the caster was asked to discard")
	}
	effectstest.DiscardFromHand(t, g, a.ID)
	effectstest.DiscardFromHand(t, g, b.ID)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d → %d, want +2", before, got)
	}
}

// --- Roiling Regrowth ---------------------------------------------

func TestB09RoilingRegrowthSacrificesOnResolutionAndFetchesTwoTappedBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest B", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest C", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Roiling Regrowth", "Instant", effectstest.B09RoilingRegrowthOracle, nil)
	if !g.Battlefield.Contains(fodder) {
		t.Error("nothing is sacrificed at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Error("the land is sacrificed as the spell resolves")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 {
		t.Errorf("search max %d, want 2", c.SearchMax)
	}
	if effectstest.SearchOptionNamed(g, c, "Bayou") != uuid.Nil {
		t.Error("a nonbasic Forest was offered for 'basic land card'")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest A", "Forest B")
	for _, name := range []string{"Forest A", "Forest B"} {
		card, ok := effectstest.SearchFetchedCard(g, name)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jhoira, Weatherlight Captain",
		TypeLine: "Legendary Creature — Human Artificer", OracleID: effectstest.B09JhoiraOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})

	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("an artifact spell: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Human", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("a legendary spell: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
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
		if got := effects.B09IsHistoric(game.Card{TypeLine: tc.typeLine}); got != tc.want {
			t.Errorf("%q historic = %v, want %v", tc.typeLine, got, tc.want)
		}
	}
}

// --- Wall of Omens / Shared Roots / Disenchant ---------------------

func TestB09WallOfOmensDrawsOnEntryAndDefends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	wall := effectstest.CastAndResolveCreature(t, g, "Wall of Omens", "Creature — Wall", effectstest.B09WallOfOmensOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want +1", before, got)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, wall), "defender") {
		t.Error("printed defender did not reach the effective abilities")
	}
}

func TestB09SharedRootsFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island A", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Shared Roots", "Sorcery — Lesson", effectstest.B09SharedRootsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bayou") != uuid.Nil {
		t.Error("a nonbasic was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island A")
	card, ok := effectstest.SearchFetchedCard(g, "Island A")
	if !ok || !card.Tapped {
		t.Errorf("the basic should be on the battlefield tapped, got %+v (found %v)", card, ok)
	}
}

func TestB09DisenchantDestroysAnEnchantmentAndRefusesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	aura := effectstest.SeedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Disenchant", "Instant", effectstest.B09DisenchantOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the enchantment should be in its owner's graveyard")
	}

	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	if err := b09TryCast(t, g, "Disenchant", "Instant", effectstest.B09DisenchantOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted for 'target artifact or enchantment'")
	}
}

// --- Archon of Cruelty --------------------------------------------

// b09AssertArchonResolved walks one Archon trigger through its
// PRINTED order (#1027):
//
//	"target opponent sacrifices a creature or planeswalker of their
//	 choice, discards a card, and loses 3 life. You draw a card and
//	 gain 3 life."
//
// Four clauses, and three of them are on the far side of two prompts.
// Before #1027 the discard prompt had no continuation, so the body
// queued both questions and ran the last three clauses on the next
// line: the victim was already three life down and the caster already
// had their card while nobody had chosen anything. Every step below
// asserts what has NOT happened yet as well as what has, because
// "eventually all four happen" was true of the broken card too.
func b09AssertArchonResolved(t *testing.T, g *game.Game, me, victim *game.Player, meLife, victimLife, meHand int, theirBear uuid.UUID) {
	t.Helper()

	// 1. The sacrifice, and nothing else.
	c := effectstest.SacrificeChoiceFor(g, victim.ID)
	if c == nil {
		t.Fatal("the victim should be choosing a creature or planeswalker to sacrifice")
	}
	if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != theirBear {
		t.Errorf("sacrifice offered %v, want just the victim's own creature %v", c.SacrificeOptions, theirBear)
	}
	if effectstest.DiscardOwed(g, victim.ID) != 0 {
		t.Error("the discard is printed AFTER the sacrifice and must not be asked beside it")
	}
	b09AssertArchonPayoutPending(t, me, victim, meLife, victimLife, meHand, "with the sacrifice prompt open")

	// 2. The discard, once the sacrifice has been made.
	effectstest.AnswerSacrifice(t, g, victim.ID, theirBear)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the chosen creature was not sacrificed")
	}
	if got := effectstest.DiscardOwed(g, victim.ID); got != 1 {
		t.Fatalf("the victim owes %d discards after sacrificing, want 1", got)
	}
	b09AssertArchonPayoutPending(t, me, victim, meLife, victimLife, meHand, "with the discard prompt open")

	// 3. The last three clauses, once the card has been pitched.
	effectstest.DiscardFromHand(t, g, victim.ID)
	if victim.Life != victimLife-3 {
		t.Errorf("victim life %d → %d, want -3", victimLife, victim.Life)
	}
	if me.Life != meLife+3 {
		t.Errorf("my life %d → %d, want +3", meLife, me.Life)
	}
	if got := me.Hand.Size(); got != meHand+1 {
		t.Errorf("my hand %d → %d, want +1", meHand, got)
	}
}

// b09AssertArchonPayoutPending asserts none of Archon's last three
// clauses has happened yet.
func b09AssertArchonPayoutPending(t *testing.T, me, victim *game.Player, meLife, victimLife, meHand int, when string) {
	t.Helper()
	if victim.Life != victimLife {
		t.Errorf("the victim lost life %s: %d → %d", when, victimLife, victim.Life)
	}
	if me.Life != meLife {
		t.Errorf("I gained life %s: %d → %d", when, meLife, me.Life)
	}
	if got := me.Hand.Size(); got != meHand {
		t.Errorf("I drew %s: hand %d → %d", when, meHand, got)
	}
}

func TestB09ArchonOfCrueltyEntersAgainstATargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	theirBear := effectstest.SeedCreature(g, "Their Bear", victim.ID)
	effectstest.SeedCreature(g, "Bystander Bear", bystander.ID)
	meLife, victimLife, bystanderLife := me.Life, victim.Life, bystander.Life

	archon := effectstest.CastAndResolveCreature(t, g, "Archon of Cruelty", "Creature — Archon", effectstest.B09ArchonOfCrueltyOracle)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, archon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	meHand := me.Hand.Size()
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the trigger targets an opponent and must ask which")
	}
	if effectstest.HasID(prompt.PickTargetPlayers, me.ID) {
		t.Error("the controller is not an opponent")
	}
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	if effectstest.TriggerOnStack(g, archon) == nil {
		t.Fatal("the trigger should be on the stack after the pick")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if bystander.Life != bystanderLife || effectstest.SacrificeChoiceFor(g, bystander.ID) != nil {
		t.Error("the untargeted opponent was touched")
	}
	b09AssertArchonResolved(t, g, me, victim, meLife, victimLife, meHand, theirBear)
	if bystander.Life != bystanderLife || effectstest.DiscardOwed(g, bystander.ID) != 0 {
		t.Error("the untargeted opponent was touched by the rest of the chain")
	}
}

func TestB09ArchonOfCrueltyAttacksAgainstATargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	archon := effectstest.PushDiesCreatureForTest(g, me.ID, "Archon of Cruelty", effectstest.B09ArchonOfCrueltyOracle, "Creature — Archon", 6, 6)
	theirBear := effectstest.SeedCreature(g, "Their Bear", victim.ID)
	meLife, victimLife, meHand := me.Life, victim.Life, me.Hand.Size()

	effectstest.DeclareAttack(t, g, victim.ID, archon)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the attack trigger must ask for its opponent")
	}
	effectstest.PickPlayer(t, g, me.ID, victim.ID)
	if effectstest.TriggersOnStackFrom(g, archon) != 1 {
		t.Fatalf("one attacker, %d triggers on the stack", effectstest.TriggersOnStackFrom(g, archon))
	}
	effectstest.PassPriorityAroundTable(t, g)
	b09AssertArchonResolved(t, g, me, victim, meLife, victimLife, meHand, theirBear)
}

// --- Up the Beanstalk ---------------------------------------------

func TestB09UpTheBeanstalkDrawsOnEntryAndOnFiveDrops(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Up the Beanstalk", "Enchantment", effectstest.B09UpTheBeanstalkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("entry: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	effectstest.CastWithCost(t, g, "Big", "Creature — Giant", "{4}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("a five-drop: hand %d → %d, want +1", before, got)
	}

	before = me.Hand.Size()
	effectstest.CastWithCost(t, g, "Small", "Creature — Bear", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("a four-drop: hand %d → %d, want no draw", before, got)
	}

	// X counts at its announced value on the stack (CR 202.3e).
	before = me.Hand.Size()
	effectstest.CastXSpell(t, g, "Fireball", "Sorcery", "", "{X}{R}", 4, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("an X=4 {X}{R} spell: hand %d → %d, want +1", before, got)
	}
}
