package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b02PushLand puts a land on the battlefield under owner, tapped or
// not, with no events (a fixture, not an entry).
func b02PushLand(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: owner, Controller: owner, Tapped: tapped,
	})
	return id
}

func b02ToMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// --- registration --------------------------------------------------

func TestBatch02CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B02UndergroundSeaOracle:      "Underground Sea",
		effectstest.B02VolcanicIslandOracle:      "Volcanic Island",
		effectstest.B02TropicalIslandOracle:      "Tropical Island",
		effectstest.B02TundraOracle:              "Tundra",
		effectstest.B02SundownPassOracle:         "Sundown Pass",
		effectstest.B02HauntedRidgeOracle:        "Haunted Ridge",
		effectstest.B02OvergrownFarmlandOracle:   "Overgrown Farmland",
		effectstest.B02DeathcapGladeOracle:       "Deathcap Glade",
		effectstest.B02ArcaneSanctumOracle:       "Arcane Sanctum",
		effectstest.B02JungleShrineOracle:        "Jungle Shrine",
		effectstest.B02SimicGrowthChamberOracle:  "Simic Growth Chamber",
		effectstest.B02GolgariRotFarmOracle:      "Golgari Rot Farm",
		effectstest.B02SeatOfTheSynodOracle:      "Seat of the Synod",
		effectstest.B02ScavengerGroundsOracle:    "Scavenger Grounds",
		effectstest.B02InfernalGraspOracle:       "Infernal Grasp",
		effectstest.B02WitheringTormentOracle:    "Withering Torment",
		effectstest.B02DamnOracle:                "Damn",
		effectstest.B02SnapOracle:                "Snap",
		effectstest.B02ArchmageEmeritusOracle:    "Archmage Emeritus",
		effectstest.B02BalefulStrixOracle:        "Baleful Strix",
		effectstest.B02GuardianProjectOracle:     "Guardian Project",
		effectstest.B02LotusCobraOracle:          "Lotus Cobra",
		effectstest.B02AvengerOfZendikarOracle:   "Avenger of Zendikar",
		effectstest.B02LoranOfTheThirdPathOracle: "Loran of the Third Path",
		effectstest.B02SyrKonradOracle:           "Syr Konrad, the Grim",
	}
	if len(want) != 25 {
		t.Fatalf("the batch is 25 cards, the table lists %d", len(want))
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

// The twelve land rows across four tables, each producing its printed
// mana — the loop-leak canary every cycle carries.
func TestBatch02LandRowsProduceTheirPrintedMana(t *testing.T) {
	type row struct {
		produced     string
		replacements int
		triggers     int
	}
	rows := map[string]row{
		"Sundown Pass":         {"{R|W}", 1, 0},
		"Haunted Ridge":        {"{B|R}", 1, 0},
		"Overgrown Farmland":   {"{G|W}", 1, 0},
		"Deathcap Glade":       {"{B|G}", 1, 0},
		"Underground Sea":      {"{U|B}", 0, 0},
		"Volcanic Island":      {"{U|R}", 0, 0},
		"Tropical Island":      {"{G|U}", 0, 0},
		"Tundra":               {"{W|U}", 0, 0},
		"Arcane Sanctum":       {"{W|U|B}", 1, 0},
		"Jungle Shrine":        {"{R|G|W}", 1, 0},
		"Simic Growth Chamber": {"{G}{U}", 1, 1},
		"Golgari Rot Farm":     {"{B}{G}", 1, 1},
	}
	found := 0
	for _, spec := range effects.All() {
		want, ok := rows[spec.Name]
		if !ok {
			continue
		}
		found++
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", spec.Name, spec.ManaAbilities, want.produced)
		}
		if len(spec.Replacements) != want.replacements {
			t.Errorf("%s: %d replacements, want %d", spec.Name, len(spec.Replacements), want.replacements)
		}
		if len(spec.Triggered) != want.triggers {
			t.Errorf("%s: %d triggers, want %d", spec.Name, len(spec.Triggered), want.triggers)
		}
	}
	if found != len(rows) {
		t.Errorf("found %d of %d land rows", found, len(rows))
	}
}

// --- lands ---------------------------------------------------------

func TestHauntedRidgeEntersTappedWithOneOtherLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id := effectstest.PlayLandFromHand(t, g, "Haunted Ridge", effectstest.B02HauntedRidgeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Haunted Ridge")
}

func TestDeathcapGladeEntersUntappedWithTwoOtherLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.PlayLandFromHand(t, g, "Deathcap Glade", effectstest.B02DeathcapGladeOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Deathcap Glade")
}

// A dual has no drawback: untapped, and either colour on demand.
func TestUndergroundSeaEntersUntappedAndOffersBothColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Underground Sea", effectstest.B02UndergroundSeaOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Underground Sea")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("a dual must ask U or B, got %+v", pick)
	}
}

func TestArcaneSanctumEntersTappedAndOffersThreeColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Arcane Sanctum", effectstest.B02ArcaneSanctumOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Arcane Sanctum")

	ready := effectstest.SeedPermanentWithOracle(g, me.ID, "Arcane Sanctum", "Land", effectstest.B02ArcaneSanctumOracle)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 3 {
		t.Fatalf("a tri-land must ask W, U or B, got %+v", pick)
	}
}

// #1337: the bounce is a resolution-time choice (ReturnOneYouControl
// / ChoosePermanents), not a target picked when the trigger goes on
// the stack.
func TestSimicGrowthChamberEntersTappedAndBouncesAChosenLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.PlayLandFromHand(t, g, "Simic Growth Chamber", effectstest.B02SimicGrowthChamberOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Simic Growth Chamber")

	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Simic Growth Chamber queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	if !me.Hand.Contains(forest) {
		t.Error("the chosen land should be back in hand")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Chamber itself should stay")
	}
}

func TestSeatOfTheSynodIsAnArtifactLandThatTapsForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seat := effectstest.SeedPermanentWithOracle(g, me.ID, "Seat of the Synod", "Artifact Land", effectstest.B02SeatOfTheSynodOracle)
	card, _ := effectstest.BattlefieldCard(g, seat)
	if !card.IsArtifact() || !card.IsLand() {
		t.Fatalf("Seat of the Synod must be both an artifact and a land: %q", card.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, seat, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

func TestScavengerGroundsSacrificesItselfToExileAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grounds := effectstest.PushCatalogPermanent(g, me.ID, "Scavenger Grounds", "Land — Desert", effectstest.B02ScavengerGroundsOracle, false)
	myDead := effectstest.PushGraveyardCardForTest(me, "My Dead")
	theirDead := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{grounds},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(grounds) {
		t.Error("the Grounds is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myDead, theirDead} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be in exile", id)
		}
	}
}

func TestScavengerGroundsRefusesANonDesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grounds := effectstest.PushCatalogPermanent(g, me.ID, "Scavenger Grounds", "Land — Desert", effectstest.B02ScavengerGroundsOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, grounds, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{forest},
	}); err == nil {
		t.Error("a Forest was accepted for 'Sacrifice a Desert'")
	}
	if !g.Battlefield.Contains(grounds) || !g.Battlefield.Contains(forest) {
		t.Error("a refused activation must pay nothing")
	}
}

// --- removal -------------------------------------------------------

func TestInfernalGraspDestroysAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Infernal Grasp", "Instant", effectstest.B02InfernalGraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
	if me.Life != before-2 {
		t.Errorf("caster life %d -> %d, want -2", before, me.Life)
	}
}

func TestWitheringTormentDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glory := effectstest.SeedPermanentFor(g, opp.ID, "Glory", "Enchantment")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Withering Torment", "Instant", effectstest.B02WitheringTormentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: glory}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(glory) {
		t.Error("the enchantment survived")
	}
	if me.Life != before-2 {
		t.Errorf("caster life %d -> %d, want -2", before, me.Life)
	}
}

func TestWitheringTormentRefusesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	b02ToMainPhase(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Withering Torment", TypeLine: "Instant",
		OracleID: effectstest.B02WitheringTormentOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Error("an artifact was accepted for 'target creature or enchantment'")
	}
}

func TestDamnDestroysOneCreatureWhenCastForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Damn", "Sorcery", effectstest.B02DamnOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the single-target cast must not sweep")
	}
}

func TestDamnOverloadedDestroysEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	b02ToMainPhase(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Damn", TypeLine: "Sorcery",
		OracleID: effectstest.B02DamnOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overloaded cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(c) {
			t.Errorf("%s survived the overloaded Damn", c)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("Damn overloaded is creatures only")
	}
}

func TestSnapBouncesAndUntapsTwoOfYourTappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b02PushLand(g, me.ID, "Island A", true)
	b := b02PushLand(g, me.ID, "Island B", true)
	c := b02PushLand(g, me.ID, "Island C", true)
	theirs := b02PushLand(g, opp.ID, "Their Island", true)
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)

	effectstest.CastCatalogSpell(t, g, "Snap", "Instant", effectstest.B02SnapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Hand.Contains(bear) {
		t.Error("the creature should be back in its owner's hand")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b) // 3 tapped lands offered, pick 2
	untapped := 0
	for _, id := range []uuid.UUID{a, b, c} {
		if card, _ := effectstest.BattlefieldCard(g, id); !card.Tapped {
			untapped++
		}
	}
	if untapped != 2 {
		t.Errorf("%d of my lands untapped, want exactly 2", untapped)
	}
	if card, _ := effectstest.BattlefieldCard(g, theirs); !card.Tapped {
		t.Error("an opponent's land was untapped")
	}
}

// --- triggers ------------------------------------------------------

func TestArchmageEmeritusDrawsOnAnInstantNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", effectstest.B02ArchmageEmeritusOracle, false)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 for the instant", before, got)
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a creature spell, want unchanged at %d", got, before+1)
	}
}

func TestBalefulStrixFliesHasDeathtouchAndDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	strix := effectstest.CastCatalogSpell(t, g, "Baleful Strix", "Artifact Creature — Bird", effectstest.B02BalefulStrixOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(strix) {
		t.Fatal("the Strix did not resolve to the battlefield")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want +1 from the ETB", before, got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, strix)
	if !effects.EotHasAbility(abilities, "flying") || !effects.EotHasAbility(abilities, "deathtouch") {
		t.Errorf("abilities %v, want flying and deathtouch", abilities)
	}
}

func TestGuardianProjectDrawsOnlyForAFreshName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Guardian Project", "Enchantment", effectstest.B02GuardianProjectOracle, false)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d -> %d, want +1 for a fresh name", before, got)
	}

	// A second creature with the same name as one you control: no draw.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a duplicate name, want unchanged", got)
	}

	// A name matching a creature card in your graveyard: no draw.
	effectstest.PushGraveyardCardForTest(me, "Ghost")
	effectstest.CastCatalogSpell(t, g, "Ghost", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a name in the graveyard, want unchanged", got)
	}

	// A token: no draw.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a token, want unchanged", got)
	}
}

func TestLotusCobraAddsManaOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lotus Cobra", "Creature — Snake", effectstest.B02LotusCobraOracle, false)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("landfall should queue an any-colour pick, got %+v", pick)
	}
}

func TestAvengerOfZendikarMakesPlantsThenGrowsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest A", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest B", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest C", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Avenger of Zendikar", "Creature — Elemental", effectstest.B02AvengerOfZendikarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Plant"); n != 3 {
		t.Fatalf("%d Plants, want 3 (one per land YOU control)", n)
	}

	effectstest.PlayLandFromHand(t, g, "Forest D", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Plant" {
			continue
		}
		if c.Counters["+1/+1"] != 1 {
			t.Errorf("a Plant has %d +1/+1 counters after landfall, want 1", c.Counters["+1/+1"])
		}
	}
}

func TestLoranDestroysAnArtifactOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Loran of the Third Path", "Legendary Creature — Human Artificer", effectstest.B02LoranOfTheThirdPathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived Loran's ETB")
	}
}

func TestLoranTapsToDrawForYouAndAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loran := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loran of the Third Path",
		TypeLine: "Legendary Creature — Human Artificer", OracleID: effectstest.B02LoranOfTheThirdPathOracle,
		Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, loran), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
	meBefore, oppBefore := me.Hand.Size(), opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, loran, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != meBefore+1 || opp.Hand.Size() != oppBefore+1 {
		t.Errorf("hands me %d->%d opp %d->%d, want +1 each", meBefore, me.Hand.Size(), oppBefore, opp.Hand.Size())
	}
}

func TestLoranTapAbilityRefusesYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	loran := effectstest.PushCatalogPermanent(g, me.ID, "Loran of the Third Path", "Legendary Creature — Human Artificer", effectstest.B02LoranOfTheThirdPathOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, loran, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Error("the controller was accepted for 'target opponent'")
	}
}

func TestSyrKonradPingsOnDeathMillAndLeavingYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", effectstest.B02SyrKonradOracle, false)
	before := effectstest.LifeOfOpponents(g)
	assertDrained := func(step string, n int) {
		t.Helper()
		for i, b := range before {
			if got := g.Seats[i+1].Life; got != b-n {
				t.Errorf("%s: opponent %d %d -> %d, want -%d", step, i+1, b, got, n)
			}
		}
	}

	// Another creature dies (an opponent's).
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	assertDrained("death", 1)

	// A creature card milled from a library — not from the battlefield.
	effectstest.B02TopOfLibrary(opp, "Milled Bear", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	assertDrained("mill", 2)

	// A milled non-creature is silent.
	effectstest.B02TopOfLibrary(opp, "Milled Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	assertDrained("mill noncreature", 2)

	// A creature card leaving YOUR graveyard.
	mine := effectstest.PushGraveyardCardForTest(me, "Old Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	assertDrained("left my graveyard", 3)

	// A creature card leaving an OPPONENT's graveyard is not yours.
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Old Bear")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	assertDrained("left their graveyard", 3)
}

func TestSyrKonradMillsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	konrad := effectstest.PushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Legendary Creature — Human Knight", effectstest.B02SyrKonradOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, konrad, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Library.Size(); got != sizes[i]-1 {
			t.Errorf("seat %d library %d -> %d, want -1", i, sizes[i], got)
		}
	}
}
