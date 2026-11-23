package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mh2_5501_b1_cards_test.go — Modern Horizons 2, batch 1 (#5501). Every
// test plays the real catalog entry through the engine.

const (
	mh2B1HealersFlockOracle       = "0b3b2fb1-d5fe-4d14-985a-85f13cd60d9c"
	mh2B1ArcboundPrototypeOracle  = "ac1003b4-1f7c-4801-9dd3-5cb7d9881af8"
	mh2B1ArcboundMouserOracle     = "5e59199c-facc-46b4-9681-988de0cfaba2"
	mh2B1ArcboundWhelpOracle      = "36875030-8094-4e2c-860d-c0820825a24e"
	mh2B1JadeAvengerOracle        = "177248b7-f3fd-42f8-b325-2625a2bb28ff"
	mh2B1KitchenImpOracle         = "eaa2ad31-065d-4aa3-a39c-5e491dd59022"
	mh2B1MarbleGargoyleOracle     = "8a1fd464-760c-43d9-829a-424945d91eb9"
	mh2B1JewelEyedCobraOracle     = "bbfee53c-30a6-4ed0-beef-5c254be75fc5"
	mh2B1BottleGolemsOracle       = "46334ae3-e496-4f15-8033-0f42ee0ae375"
	mh2B1StormGodsOracleOracle    = "21613fac-d959-44ba-9bcb-0a30b5e6545e"
	mh2B1GuardianKirinOracle      = "e9dbfda3-e162-49f0-b9bf-f8213e27c17d"
	mh2B1MyrScraplingOracle       = "1a4ffc1e-2f5e-446a-b1e8-32385b3c083b"
	mh2B1ParcelMyrOracle          = "9e660f66-0a94-4694-95e0-25296799b193"
	mh2B1CryptkeeperOracle        = "8c846841-1b44-4bc5-b803-76baaa217548"
	mh2B1VedalkenInfiltratorOracl = "1d6aa6ab-7cfc-463f-a8ba-007058933ebf"
	mh2B1FoulWatcherOracle        = "b50dd71f-3664-41d2-be06-57729064a591"
	mh2B1ClatteringAugurOracle    = "d88f249f-5bf1-448a-91d1-f81538dfd475"
	mh2B1SkophosReaverOracle      = "4bc0bdd5-32da-46df-b624-29a709efc676"
	mh2B1HellMongrelOracle        = "64aa732c-d939-48c7-b1d6-c73fa1eedb40"
	mh2B1RevolutionistOracle      = "d00607b1-77ce-4b89-ad04-763d0d390c5c"
	mh2B1DreadmawOracle           = "cbde4b9d-1549-4892-b8ca-38256a6ac1ba"
	mh2B1MentalJourneyOracle      = "94e71162-7c9d-46f4-90e8-413a1b08a198"
	mh2B1WorldWearyOracle         = "de6e0f80-3a9e-4499-aec1-2513a0264538"
	mh2B1LoseFocusOracle          = "1cea6439-7ae5-4887-8c33-7da9fb36e2d4"
	mh2B1MineCollapseOracle       = "83df0f44-86ca-4adb-9f3a-da1ba6dd7a6d"
	mh2B1LensFlareOracle          = "e39bcfb8-cc45-42ab-a4b6-003f44d6b578"
	mh2B1ArchfiendOracle          = "fe84966e-0f0c-4df8-b61c-e9693492dfd8"
	mh2B1SoulOfMigrationOracle    = "f90ebf35-08f9-4277-92d9-75165f32a84e"
	mh2B1DiscipleOfTheSunOracle   = "3ea41848-7a11-4608-88c1-7a6e76e812ca"
	mh2B1FunnelWebRecluseOracle   = "6148fcf5-466a-43aa-a0f3-d3d9fae4d883"
	mh2B1FloodhoundOracle         = "6dd8ba09-f117-4514-9e09-f9f9f5fb5025"
	mh2B1EtheriumSpinnerOracle    = "b501128b-2797-401b-8aec-923050081cae"
	mh2B1DreyKeeperOracle         = "7223ef44-0bfb-4108-810d-134174ff26b0"
	mh2B1CapturedByLagacsOracle   = "db1a29f9-1608-4502-a366-89048c2e9bda"
	mh2B1TragicFallOracle         = "1b7e325d-24ee-42e3-b06f-6c91ec5280c3"
	mh2B1ViashinoLashclawOracle   = "02358f98-99d5-45a3-b7e3-f2cbf08fa52c"
	mh2B1VerminGorgerOracle       = "7b81596b-9487-4be1-a417-6913fa72299d"
	mh2B1RavingVisionaryOracle    = "83157972-3a0b-40e2-a80a-7bfb2a8627ee"
	mh2B1SlagStriderOracle        = "2b154d55-37dd-4e5f-aa99-d4bc547cdec0"
	mh2B1SteelfinWhaleOracle      = "a35ff288-0ba0-47de-9848-4ff4bef6b91b"
	mh2B1GlimmerBairnOracle       = "118cdf8b-44fb-4c3e-86c4-e7f4b1267646"
	mh2B1SweepTheSkiesOracle      = "7596fe49-c9cf-4415-ac1b-156cbbf9604e"
	mh2B1RadiantEpicureOracle     = "cdea4344-24b6-4659-a830-c51d4b348af5"

	mh2B1BearType = "Creature — Bear"
	mh2B1AuraType = "Enchantment — Aura"
)

// mh2B1Ogre seeds a vanilla 4/4 that survives three damage.
func mh2B1Ogre(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Creature(g, owner, "Ogre", "Creature — Ogre", 4, 4)
}

// mh2B1CastPaid casts a catalog card from `me`'s hand under the strict
// mana gate with X announced as x, paying from exactly the floated pool,
// and resolves the spell.
func mh2B1CastPaid(t *testing.T, g *game.Game, me *game.Player, c game.Card, x int, pool string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return c.InstanceID
}

// --- every card is in the catalog ----------------------------------------

func TestMh2B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mh2B1HealersFlockOracle: "Healer's Flock", mh2B1ArcboundPrototypeOracle: "Arcbound Prototype",
		mh2B1ArcboundMouserOracle: "Arcbound Mouser", mh2B1ArcboundWhelpOracle: "Arcbound Whelp",
		mh2B1JadeAvengerOracle: "Jade Avenger", mh2B1KitchenImpOracle: "Kitchen Imp",
		mh2B1MarbleGargoyleOracle: "Marble Gargoyle", mh2B1JewelEyedCobraOracle: "Jewel-Eyed Cobra",
		mh2B1BottleGolemsOracle: "Bottle Golems", mh2B1StormGodsOracleOracle: "Storm God's Oracle",
		mh2B1GuardianKirinOracle: "Guardian Kirin", mh2B1MyrScraplingOracle: "Myr Scrapling",
		mh2B1ParcelMyrOracle: "Parcel Myr", mh2B1CryptkeeperOracle: "Tormod's Cryptkeeper",
		mh2B1VedalkenInfiltratorOracl: "Vedalken Infiltrator", mh2B1FoulWatcherOracle: "Foul Watcher",
		mh2B1ClatteringAugurOracle: "Clattering Augur", mh2B1SkophosReaverOracle: "Skophos Reaver",
		mh2B1HellMongrelOracle: "Hell Mongrel", mh2B1RevolutionistOracle: "Revolutionist",
		mh2B1DreadmawOracle: "Phantasmal Dreadmaw", mh2B1MentalJourneyOracle: "Mental Journey",
		mh2B1WorldWearyOracle: "World-Weary", mh2B1LoseFocusOracle: "Lose Focus",
		mh2B1MineCollapseOracle: "Mine Collapse", mh2B1LensFlareOracle: "Lens Flare",
		mh2B1ArchfiendOracle: "Archfiend of Sorrows", mh2B1SoulOfMigrationOracle: "Soul of Migration",
		mh2B1DiscipleOfTheSunOracle: "Disciple of the Sun", mh2B1FunnelWebRecluseOracle: "Funnel-Web Recluse",
		mh2B1FloodhoundOracle: "Floodhound", mh2B1EtheriumSpinnerOracle: "Etherium Spinner",
		mh2B1DreyKeeperOracle: "Drey Keeper", mh2B1CapturedByLagacsOracle: "Captured by Lagacs",
		mh2B1TragicFallOracle: "Tragic Fall", mh2B1ViashinoLashclawOracle: "Viashino Lashclaw",
		mh2B1VerminGorgerOracle: "Vermin Gorger", mh2B1RavingVisionaryOracle: "Raving Visionary",
		mh2B1SlagStriderOracle: "Slag Strider", mh2B1SteelfinWhaleOracle: "Steelfin Whale",
		mh2B1GlimmerBairnOracle: "Glimmer Bairn", mh2B1SweepTheSkiesOracle: "Sweep the Skies",
		mh2B1RadiantEpicureOracle: "Radiant Epicure",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- printed keywords -----------------------------------------------------

func TestMh2B1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Healer's Flock", "Creature — Bird", mh2B1HealersFlockOracle, []string{"flying", "lifelink"}},
		{"Arcbound Mouser", "Artifact Creature — Cat", mh2B1ArcboundMouserOracle, []string{"lifelink"}},
		{"Arcbound Whelp", "Artifact Creature — Dragon", mh2B1ArcboundWhelpOracle, []string{"flying"}},
		{"Kitchen Imp", "Creature — Imp", mh2B1KitchenImpOracle, []string{"flying", "haste"}},
		{"Marble Gargoyle", "Artifact Creature — Gargoyle", mh2B1MarbleGargoyleOracle, []string{"flying"}},
		{"Jewel-Eyed Cobra", "Creature — Snake", mh2B1JewelEyedCobraOracle, []string{"deathtouch"}},
		{"Bottle Golems", "Artifact Creature — Golem", mh2B1BottleGolemsOracle, []string{"trample"}},
		{"Guardian Kirin", "Creature — Kirin", mh2B1GuardianKirinOracle, []string{"flying"}},
		{"Tormod's Cryptkeeper", "Artifact Creature — Golem", mh2B1CryptkeeperOracle, []string{"vigilance"}},
		{"Foul Watcher", "Creature — Nightmare Bird", mh2B1FoulWatcherOracle, []string{"flying"}},
		{"Phantasmal Dreadmaw", "Creature — Dinosaur Illusion", mh2B1DreadmawOracle, []string{"trample"}},
		{"Archfiend of Sorrows", "Creature — Demon", mh2B1ArchfiendOracle, []string{"flying"}},
		{"Soul of Migration", "Creature — Elemental", mh2B1SoulOfMigrationOracle, []string{"flying"}},
		{"Disciple of the Sun", "Creature — Human Cleric", mh2B1DiscipleOfTheSunOracle, []string{"lifelink"}},
		{"Funnel-Web Recluse", "Creature — Spider", mh2B1FunnelWebRecluseOracle, []string{"reach"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- modular --------------------------------------------------------------

func TestMh2B1ArcboundsEnterWithTheirModularCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		counters               int
	}{
		{"Arcbound Prototype", "Artifact Creature — Assembly-Worker", mh2B1ArcboundPrototypeOracle, 2},
		{"Arcbound Mouser", "Artifact Creature — Cat", mh2B1ArcboundMouserOracle, 1},
		{"Arcbound Whelp", "Artifact Creature — Dragon", mh2B1ArcboundWhelpOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			id := effectstest.MrdEnter(t, g, me, tc.name, tc.typeLine, tc.oracle)
			if got := mrdPlusOneCounters(t, g, id); got != tc.counters {
				t.Errorf("%s entered with %d +1/+1 counters, want %d", tc.name, got, tc.counters)
			}
		})
	}
}

func TestMh2B1ArcboundPrototypePassesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	proto := effectstest.MrdEnter(t, g, me, "Arcbound Prototype", "Artifact Creature — Assembly-Worker", mh2B1ArcboundPrototypeOracle)
	thopter := effectstest.MrdPermanent(g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "{0}", 0, 2)
	effectstest.B25Destroy(g, proto)
	effectstest.AnswerPickTarget(t, g, thopter)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, thopter, game.CounterPlusOne); got != 2 {
		t.Errorf("the Thopter got %d counters, want the Prototype's 2", got)
	}
}

// --- bushido --------------------------------------------------------------

func TestMh2B1JadeAvengerGrowsWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Jade Avenger", "Creature — Frog Samurai", mh2B1JadeAvengerOracle, 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("Jade Avenger is %d/%d after becoming blocked, want 4/4", p, tough)
	}
}

// --- madness --------------------------------------------------------------

func TestMh2B1MadnessCostsAreDeclared(t *testing.T) {
	for oracle, want := range map[string]string{
		mh2B1KitchenImpOracle:    "{B}",
		mh2B1SkophosReaverOracle: "{1}{R}",
		mh2B1HellMongrelOracle:   "{2}{B}",
		mh2B1RevolutionistOracle: "{3}{R}",
	} {
		spec, _ := effects.Lookup(oracle)
		if spec.Madness != want {
			t.Errorf("%s: madness %q, want %q", spec.Name, spec.Madness, want)
		}
	}
}

// Discarded through Faithless Looting, Kitchen Imp is exiled and cast for
// {B}, and arrives with haste.
func TestMh2B1KitchenImpIsCastForItsMadnessCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.TspB1Loot(t, g, "Kitchen Imp", "Creature — Imp", mh2B1KitchenImpOracle)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Kitchen Imp did not resolve")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Kitchen Imp has no haste")
	}
}

// --- pumps ----------------------------------------------------------------

func TestMh2B1ManaPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		mana                   []string
		wantP, wantT           int
	}{
		{"Marble Gargoyle", "Artifact Creature — Gargoyle", mh2B1MarbleGargoyleOracle, 2, 2, []string{"W"}, 2, 3},
		{"Storm God's Oracle", "Enchantment Creature — Human Shaman", mh2B1StormGodsOracleOracle, 1, 3, []string{"C"}, 2, 2},
		{"Arcbound Whelp", "Artifact Creature — Dragon", mh2B1ArcboundWhelpOracle, 2, 2, []string{"R"}, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := effectstest.PtOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestMh2B1GlimmerBairnSacrificesATokenToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bairn := effectstest.B12Push(g, me.ID, "Glimmer Bairn", "Creature — Ouphe", mh2B1GlimmerBairnOracle, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh2B1BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, bairn, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a nontoken creature paid the sacrifice")
	}
	token := effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	effectstest.ApaActivate(t, g, me, bairn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{token}})
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, bairn); p != 3 || tough != 4 {
		t.Errorf("Glimmer Bairn is %d/%d, want 3/4", p, tough)
	}
}

func TestMh2B1HellMongrelDiscardsToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mongrel := effectstest.B12Push(g, me.ID, "Hell Mongrel", "Creature — Nightmare Dog", mh2B1HellMongrelOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, mongrel, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mongrel); p != 5 || tough != 4 {
		t.Errorf("Hell Mongrel is %d/%d, want 5/4", p, tough)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

// --- statics --------------------------------------------------------------

func TestMh2B1VedalkenInfiltratorCantBeBlockedAndGrowsWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := effectstest.B12Push(g, me.ID, "Vedalken Infiltrator", "Creature — Vedalken Rogue", mh2B1VedalkenInfiltratorOracl, 1, 3)
	effectstest.B12Permanent(g, me.ID, "A", "Artifact")
	effectstest.B12Permanent(g, me.ID, "B", "Artifact")
	if p, tough := effectstest.EffectivePTOf(t, g, inf); p != 1 || tough != 3 {
		t.Errorf("with two artifacts it is %d/%d, want 1/3", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "C", "Artifact")
	if p, tough := effectstest.EffectivePTOf(t, g, inf); p != 2 || tough != 3 {
		t.Errorf("with three artifacts it is %d/%d, want 2/3", p, tough)
	}
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, inf)
	if err := g.DeclareBlocker(wall, inf); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}

func TestMh2B1FoulWatcherSurveilsAndGrowsWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top A", "Top B")
	effectstest.CastHoldCreature(t, g, "Foul Watcher", "Creature — Nightmare Bird", mh2B1FoulWatcherOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice = %+v, want one card", c)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	watcher := effectstest.B12Push(g, me.ID, "Foul Watcher", "Creature — Nightmare Bird", mh2B1FoulWatcherOracle, 1, 2)
	effectstest.B24SeedTypes(me, 3)
	effectstest.S29Bump(g)
	if p, tough := effectstest.S29Size(t, g, watcher); p != 1 || tough != 2 {
		t.Fatalf("three card types: %d/%d, want 1/2", p, tough)
	}
	effectstest.GyCard(me, "Old Rock", "Artifact", "{1}", "")
	effectstest.S29Bump(g)
	if p, tough := effectstest.S29Size(t, g, watcher); p != 2 || tough != 2 {
		t.Errorf("four card types: %d/%d, want 2/2", p, tough)
	}
}

func TestMh2B1SkophosReaverGetsPlusTwoOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reaver := effectstest.B12Push(g, me.ID, "Skophos Reaver", "Creature — Minotaur Warrior", mh2B1SkophosReaverOracle, 2, 3)
	if p, tough := effectstest.EoeCDPT(t, g, reaver); p != 4 || tough != 3 {
		t.Errorf("on my turn %d/%d, want 4/3", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EoeCDPT(t, g, reaver); p != 2 || tough != 3 {
		t.Errorf("on their turn %d/%d, want 2/3", p, tough)
	}
}

func TestMh2B1WorldWearyShrinksTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	effectstest.CastCatalogSpell(t, g, "World-Weary", mh2B1AuraType, mh2B1WorldWearyOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 2/2", p, tough)
	}
}

func TestMh2B1CapturedByLagacsBarsTheHostAndSupportsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirs := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Theirs", mh2B1BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", mh2B1BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Captured by Lagacs", mh2B1AuraType, mh2B1CapturedByLagacsOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(mine)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("the supported creature has %d +1/+1 counters, want 1", got)
	}
	c := effectstest.LayeredCard(t, g, theirs)
	if !game.Restricted(&c, game.CantAttack) || !game.Restricted(&c, game.CantBlock) {
		t.Error("the enchanted creature can still attack or block")
	}
}

// --- triggered abilities --------------------------------------------------

func TestMh2B1JewelEyedCobraLeavesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cobra := effectstest.B12Push(g, me.ID, "Jewel-Eyed Cobra", "Creature — Snake", mh2B1JewelEyedCobraOracle, 3, 1)
	effectstest.B27Kill(g, cobra)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestMh2B1BottleGolemsGainLifeEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golems := effectstest.B12Push(g, me.ID, "Bottle Golems", "Artifact Creature — Golem", mh2B1BottleGolemsOracle, 3, 3)
	life := me.Life
	effectstest.B27Kill(g, golems)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMh2B1StormGodsOracleDealsThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oracle := effectstest.B12Push(g, me.ID, "Storm God's Oracle", "Enchantment Creature — Human Shaman", mh2B1StormGodsOracleOracle, 1, 3)
	ogre := mh2B1Ogre(g, opp.ID)
	effectstest.B27Kill(g, oracle)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the target has %d damage, want 3", got)
	}
}

func TestMh2B1GuardianKirinGrowsWhenAnotherCreatureYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirin := effectstest.B12Push(g, me.ID, "Guardian Kirin", "Creature — Kirin", mh2B1GuardianKirinOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", mh2B1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", mh2B1BearType, 2, 2)
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, kirin, game.CounterPlusOne); got != 0 {
		t.Fatalf("an opponent's creature dying gave %d counters", got)
	}
	effectstest.B27Kill(g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, kirin, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after your creature died, want 1", got)
	}
}

func TestMh2B1ClatteringAugurDrawsCantBlockAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	augur := effectstest.CastHoldCreature(t, g, "Clattering Augur", "Creature — Skeleton Shaman", mh2B1ClatteringAugurOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	c := effectstest.LayeredCard(t, g, augur)
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("Clattering Augur can block")
	}

	effectstest.B27Kill(g, augur)
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(augur) {
		t.Fatal("the Augur is not in the graveyard")
	}
	effectstest.ApaMana(me, "C", "C", "B", "B")
	effectstest.ApaActivate(t, g, me, augur, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(augur) {
		t.Error("the Augur did not return to hand")
	}
}

func TestMh2B1RevolutionistReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
	effectstest.CastHoldCreature(t, g, "Revolutionist", "Creature — Human Wizard", mh2B1RevolutionistOracle, 3, 3)
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant card did not return to hand")
	}
}

func TestMh2B1DiscipleOfTheSunReturnsACheapPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.GyCard(me, "Old Bear", mh2B1BearType, "{1}{G}", "")
	effectstest.CastHoldCreature(t, g, "Disciple of the Sun", "Creature — Human Cleric", mh2B1DiscipleOfTheSunOracle, 3, 3)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestMh2B1PhantasmalDreadmawIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	maw := effectstest.B12Push(g, me.ID, "Phantasmal Dreadmaw", "Creature — Dinosaur Illusion", mh2B1DreadmawOracle, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Tragic Fall", "Instant", mh2B1TragicFallOracle, effectstest.CardTarget(maw))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(maw) || !me.Graveyard.Contains(maw) {
		t.Error("the Dreadmaw was not sacrificed when it became a target")
	}
}

func TestMh2B1ArchfiendOfSorrowsShrinksOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", mh2B1BearType, 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", mh2B1BearType, 2, 2)
	big := mh2B1Ogre(g, opp.ID)
	effectstest.CastHoldCreature(t, g, "Archfiend of Sorrows", "Creature — Demon", mh2B1ArchfiendOracle, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("an opponent's 2/2 survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, big); p != 2 || tough != 2 {
		t.Errorf("an opponent's 4/4 is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 2/2", p, tough)
	}
}

func TestMh2B1SoulOfMigrationMakesTwoBirds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Soul of Migration", "Creature — Elemental", mh2B1SoulOfMigrationOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Bird"); n != 2 {
		t.Errorf("%d Birds, want 2", n)
	}
}

func TestMh2B1DreyKeeperMakesSquirrelsAndPumpsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keeper := effectstest.CastHoldCreature(t, g, "Drey Keeper", "Creature — Elf Druid", mh2B1DreyKeeperOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	squirrels := effectstest.BattlefieldIDsNamed(g, "Squirrel")
	if len(squirrels) != 2 {
		t.Fatalf("%d Squirrels, want 2", len(squirrels))
	}
	effectstest.ApaMana(me, "C", "C", "C", "B")
	effectstest.ApaActivate(t, g, me, keeper, 0, game.ActivateAbilityParams{})
	for _, id := range squirrels {
		if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 1 {
			t.Errorf("a Squirrel is %d/%d, want 2/1", p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Error("a Squirrel has no menace")
		}
	}
	if p, _ := effectstest.PtOf(t, g, keeper); p != 2 {
		t.Errorf("the Keeper itself has power %d, want 2", p)
	}
}

func TestMh2B1FunnelWebRecluseInvestigatesOnlyIfACreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Funnel-Web Recluse", "Creature — Spider", mh2B1FunnelWebRecluseOracle, 3, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Clue"); n != 0 {
		t.Fatalf("%d Clues with no creature dead this turn, want 0", n)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.B12Creature(g, me.ID, "Victim", mh2B1BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Kill(g, victim)
	g.RunStateChecksForTest()
	effectstest.CastHoldCreature(t, g, "Funnel-Web Recluse", "Creature — Spider", mh2B1FunnelWebRecluseOracle, 3, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues after a creature died, want 1", n)
	}
}

func TestMh2B1EtheriumSpinnerMakesAThopterForABigSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Etherium Spinner", "Artifact Creature — Human Wizard", mh2B1EtheriumSpinnerOracle, 2, 1)
	small := effectstest.HandCardFull(me, "Small Spell", "Sorcery", "{2}{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the small spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Thopter"); n != 0 {
		t.Fatalf("%d Thopters for a mana value 3 spell, want 0", n)
	}
	big := effectstest.HandCardFull(me, "Big Spell", "Sorcery", "{3}{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the big spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Thopter"); n != 1 {
		t.Errorf("%d Thopters for a mana value 4 spell, want 1", n)
	}
}

func TestMh2B1SteelfinWhaleUntapsWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whale := effectstest.B12Push(g, me.ID, "Steelfin Whale", "Creature — Whale", mh2B1SteelfinWhaleOracle, 3, 4)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, whale).Tapped = true })
	effectstest.DstToken(t, g, me.ID, "My Thopter", "Token Artifact Creature — Thopter")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, whale) {
		t.Error("the Whale is still tapped after an artifact entered")
	}
}

func TestMh2B1RadiantEpicureDrainsPerColourSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opps := []*game.Player{}
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
		if s.ID != me.ID {
			opps = append(opps, s)
		}
	}
	mh2B1CastPaid(t, g, me, game.Card{Name: "Radiant Epicure", TypeLine: "Creature — Vampire Wizard", ManaCost: "{4}{B}",
		Power: 5, Toughness: 5, Colors: []string{"B"}, OracleID: mh2B1RadiantEpicureOracle}, 0, "BRGCC")
	effectstest.PassPriorityAroundTable(t, g)
	for _, o := range opps {
		if o.Life != lives[o.ID]-3 {
			t.Errorf("an opponent is at %d, want %d", o.Life, lives[o.ID]-3)
		}
	}
	if me.Life != lives[me.ID]+3 {
		t.Errorf("you are at %d, want %d", me.Life, lives[me.ID]+3)
	}
}

// --- activated abilities --------------------------------------------------

func TestMh2B1MyrScraplingSacrificesToPutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	myr := effectstest.B12Push(g, me.ID, "Myr Scrapling", "Artifact Creature — Myr", mh2B1MyrScraplingOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh2B1BearType, 2, 2)
	effectstest.ApaActivate(t, g, me, myr, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if g.Battlefield.Contains(myr) {
		t.Error("the Scrapling was not sacrificed")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

func TestMh2B1ParcelMyrSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	myr := effectstest.B12Push(g, me.ID, "Parcel Myr", "Artifact Creature — Clue Myr", mh2B1ParcelMyrOracle, 2, 1)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, myr, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(myr) {
		t.Error("the Myr was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestMh2B1TormodsCryptkeeperExilesAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	keeper := effectstest.B12Push(g, me.ID, "Tormod's Cryptkeeper", "Artifact Creature — Golem", mh2B1CryptkeeperOracle, 3, 2)
	one := effectstest.PushGraveyardCardForTest(opp, "One")
	two := effectstest.PushGraveyardCardForTest(opp, "Two")
	effectstest.ApaActivate(t, g, me, keeper, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(keeper) {
		t.Error("the Cryptkeeper was not sacrificed")
	}
	if opp.Graveyard.Size() != 0 || !g.Exile.Contains(one) || !g.Exile.Contains(two) {
		t.Error("the target player's graveyard was not exiled")
	}
}

func TestMh2B1FloodhoundInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hound := effectstest.B12Push(g, me.ID, "Floodhound", "Creature — Elemental Dog", mh2B1FloodhoundOracle, 1, 2)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, hound, 0, game.ActivateAbilityParams{})
	if n := effectstest.B27CountNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
	if !effectstest.HmlTapped(t, g, hound) {
		t.Error("the Floodhound did not tap")
	}
}

func TestMh2B1ViashinoLashclawGivesYourCreaturesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lash := effectstest.B12Push(g, me.ID, "Viashino Lashclaw", "Creature — Lizard Warrior", mh2B1ViashinoLashclawOracle, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", mh2B1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", mh2B1BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, lash, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("your creature has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
}

func TestMh2B1VerminGorgerDrainsForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	gorger := effectstest.B12Push(g, me.ID, "Vermin Gorger", "Creature — Vampire", mh2B1VerminGorgerOracle, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, gorger, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{gorger}}); err == nil {
		t.Fatal("the Gorger sacrificed itself")
	}
	victim := effectstest.B12Creature(g, me.ID, "Victim", mh2B1BearType, 2, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.ApaActivate(t, g, me, gorger, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID] + 2
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestMh2B1RavingVisionaryLootsAndDrawsWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	seer := effectstest.B12Push(g, me.ID, "Raving Visionary", "Creature — Merfolk Wizard", mh2B1RavingVisionaryOracle, 1, 1)
	effectstest.ApaMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, seer, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the delirium draw was activated without four card types")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.B24SeedTypes(me, 4)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C", "U")
	effectstest.ApaActivate(t, g, me, seer, 1, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d after the delirium draw, want %d", got, hand+1)
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me = effectstest.Top100ActiveSeat(g)
	seer = effectstest.B12Push(g, me.ID, "Raving Visionary", "Creature — Merfolk Wizard", mh2B1RavingVisionaryOracle, 1, 1)
	hand = me.Hand.Size()
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, seer, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestMh2B1SlagStriderSacrificesAnArtifactToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	strider := effectstest.B12Push(g, me.ID, "Slag Strider", "Creature — Elemental", mh2B1SlagStriderOracle, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	life := opp.Life
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
	})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestMh2B1AffinityForArtifacts(t *testing.T) {
	for _, oracle := range []string{mh2B1LensFlareOracle, mh2B1SlagStriderOracle, mh2B1SteelfinWhaleOracle} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.SelfCostModifiers) != 1 {
			t.Errorf("%s has %d self cost modifiers, want affinity", spec.Name, len(spec.SelfCostModifiers))
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, n := range []string{"A", "B", "C", "D"} {
		effectstest.B12Permanent(g, me.ID, n, "Artifact")
	}
	whale := mh2B1CastPaid(t, g, me, game.Card{Name: "Steelfin Whale", TypeLine: "Creature — Whale", ManaCost: "{5}{U}",
		Power: 3, Toughness: 4, Colors: []string{"U"}, OracleID: mh2B1SteelfinWhaleOracle}, 0, "UC")
	if !g.Battlefield.Contains(whale) {
		t.Error("Steelfin Whale was not cast for {1}{U} with four artifacts")
	}
}

// --- spells ---------------------------------------------------------------

func TestMh2B1MentalJourneyDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mental Journey", "Instant", mh2B1MentalJourneyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand is %d, want %d", got, hand+3)
	}
}

func TestMh2B1BasicLandcyclersHaveTheirRow(t *testing.T) {
	for oracle, label := range map[string]string{
		mh2B1MentalJourneyOracle: "Basic landcycling {1}{U}",
		mh2B1WorldWearyOracle:    "Basic landcycling {1}{B}",
	} {
		spec, _ := effects.Lookup(oracle)
		found := false
		for _, a := range spec.Activated {
			if len(a.Label) >= len(label) && a.Label[:len(label)] == label {
				found = true
			}
		}
		if !found {
			t.Errorf("%s has no %q row", spec.Name, label)
		}
	}
}

func TestMh2B1TragicFallIsBiggerWhenHellbent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.D2dSetHand(me, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Tragic Fall", "Instant", mh2B1TragicFallOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 3 || tough != 3 {
		t.Fatalf("with cards in hand the target is %d/%d, want 3/3", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	huge := effectstest.B12Creature(g, opp.ID, "Huge", "Creature — Giant", 12, 12)
	effectstest.CastCatalogSpell(t, g, "Tragic Fall", "Instant", mh2B1TragicFallOracle, effectstest.CardTarget(huge))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(huge) {
		t.Error("a 12/12 survived the hellbent -13/-13")
	}
}

func TestMh2B1LensFlareHitsOnlyACombatant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raider := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 6, 6)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Human", 6, 6)
	effectstest.BrAttack(t, g, raider)
	flare := effectstest.B13HandCard(me, "Lens Flare", "Instant", mh2B1LensFlareOracle)
	if err := g.CastSpell(me.ID, flare, game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("Lens Flare targeted a creature that isn't in combat")
	}
	if err := g.CastSpell(me.ID, flare, game.CastSpellParams{Targets: effectstest.CardTarget(raider)}); err != nil {
		t.Fatalf("Lens Flare at the attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, raider).DamageMarked; got != 5 {
		t.Errorf("the attacker has %d damage, want 5", got)
	}
}

func TestMh2B1MineCollapseSacrificesAMountainOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	collapse := effectstest.HandCardFull(me, "Mine Collapse", "Instant", "{3}{R}", mh2B1MineCollapseOracle, []string{"R"})
	ogre := effectstest.B12Creature(g, them.ID, "Ogre", "Creature — Ogre", 5, 5)
	island := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := mmqB1Cast(g, me.ID, collapse, "sacrifice", []uuid.UUID{island}, effectstest.CardRefs(ogre)); err == nil {
		t.Fatal("an Island paid \"sacrifice a Mountain\"")
	}
	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := mmqB1Cast(g, me.ID, collapse, "sacrifice", []uuid.UUID{mountain}, effectstest.CardRefs(ogre)); err != nil {
		t.Fatalf("Mine Collapse for a Mountain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mountain) {
		t.Error("the Mountain was not sacrificed")
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("a 5/5 survived 5 damage")
	}

	spec, _ := effects.Lookup(mh2B1MineCollapseOracle)
	cond := spec.AlternativeCosts[0].Condition
	if cond == nil || !cond(g, g.Seats[g.Turn.ActiveSeat].ID) {
		t.Error("the active player is not offered the sacrifice")
	}
	if cond(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID) {
		t.Error("a player whose turn it isn't is offered the sacrifice")
	}
}

func TestMh2B1LoseFocusCountersUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.B13HandCard(me, "Lose Focus", "Instant", mh2B1LoseFocusOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the spell's controller was not asked to pay {2}")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the unpaid spell was not countered")
	}
}

func TestMh2B1LoseFocusHasReplicate(t *testing.T) {
	spec, _ := effects.Lookup(mh2B1LoseFocusOracle)
	if len(spec.OptionalCosts) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("Lose Focus has %d optional costs and %d triggers, want replicate's one of each", len(spec.OptionalCosts), len(spec.Triggered))
	}
}

func TestMh2B1SweepTheSkiesMakesAThopterPerColourSpent(t *testing.T) {
	for _, tc := range []struct {
		x    int
		pool string
		want int
	}{{0, "UU", 1}, {2, "UUWB", 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		mh2B1CastPaid(t, g, me, game.Card{Name: "Sweep the Skies", TypeLine: "Sorcery", ManaCost: "{X}{U}{U}",
			Colors: []string{"U"}, OracleID: mh2B1SweepTheSkiesOracle}, tc.x, tc.pool)
		if n := effectstest.B27CountNamed(g, me.ID, "Thopter"); n != tc.want {
			t.Errorf("%s: %d Thopters, want %d", tc.pool, n, tc.want)
		}
	}
}
