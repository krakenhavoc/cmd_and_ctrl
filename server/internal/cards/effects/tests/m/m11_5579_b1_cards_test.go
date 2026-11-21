package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m11_5579_b1_cards_test.go — the first batch of Magic 2011 cards (#5579).
// Every test plays the real catalog entry through the engine.

const (
	m11ArmoredCancrixOracle       = "445bb343-9758-459b-970f-47f1705f1e55"
	m11BaronyVampireOracle        = "e886756a-298d-4f79-8699-8bfc9345352e"
	m11MaritimeGuardOracle        = "0bb1c704-47f1-4d0c-8158-cfcd0b5e4d06"
	m11NetherHorrorOracle         = "d4647f16-b287-46a2-9c38-93881c290f76"
	m11StoneGolemOracle           = "7cfb32af-ac59-4459-ac9e-2762488e9180"
	m11AssaultGriffinOracle       = "802b9333-c256-410a-a635-a22bfeda679f"
	m11CloudCrusaderOracle        = "2ef481fa-ff20-426a-af16-fbf761ea5eaf"
	m11GarruksCompanionOracle     = "32fc8fa7-7e0a-4d4e-85cf-2f98b3fc6ecf"
	m11GreaterBasiliskOracle      = "6e42e7dc-acd9-4125-a90d-75fd8da204b3"
	m11SacredWolfOracle           = "ccb64a70-cbdd-4991-a28d-b137032a4f4d"
	m11WallOfVinesOracle          = "51ce0158-c9a5-4fa5-9704-46ab02f01b86"
	m11RottingLegionOracle        = "6e5fd8b4-d81a-4ed6-b389-21445351bb43"
	m11AetherAdeptOracle          = "e076067c-77b7-4a37-a62d-51434392dd53"
	m11AuguryOwlOracle            = "44c5f8e1-d9b8-4067-a60a-1ecc8bd11145"
	m11LilianasSpecterOracle      = "1a7e0e55-41c3-4fd8-b718-55b59b5cab90"
	m11ManicVandalOracle          = "9c12bdbf-80cd-4edb-bebb-b4efcf532272"
	m11TirelessMissionariesOracle = "a730e699-834a-4aac-9b3b-29d6917fc9b3"
	m11WarPriestOfThuneOracle     = "2e44fd5c-5c7f-4ca8-90cd-34015047eb26"
	m11SquadronHawkOracle         = "00ab9841-934f-4a66-a98d-68d01661b1c9"
	m11RocEggOracle               = "16dfd80a-9847-45b8-954f-2af9b7cfa21d"
	m11ArcRunnerOracle            = "9bbee1a5-2eb0-4fad-9524-4ad9cc5ca998"
	m11AirServantOracle           = "49638d96-e83e-4aab-bafb-d6183a6e3ff6"
	m11BrindleBoarOracle          = "1c4f6893-4e83-4756-b324-1de0d06ea2c4"
	m11EmberHaulerOracle          = "7d313958-49ca-4592-97c7-20f12c8468a8"
	m11NightwingShadeOracle       = "e679ef01-8d12-4246-850a-efed7f4316b2"
	m11WaterServantOracle         = "f8c32c80-9693-4ae4-a387-21ec8a3cdd40"
	m11GargoyleSentinelOracle     = "893f319f-6e3d-4230-bb65-4f9179b18fad"
	m11BrittleEffigyOracle        = "34909cf0-780d-4bc7-9926-9f7625b8f200"
	m11PhantomBeastOracle         = "13ca7c14-5db2-4d17-912a-fdeabcdcd9f8"
	m11AjanisMantraOracle         = "2f47661e-a107-4d98-8773-0a068a63df49"
	m11LeylineOfVitalityOracle    = "cbf2b19e-32a3-489f-9979-ba2748c52845"
	m11EarthServantOracle         = "7ce30379-3f07-4b94-b2d6-a89b4763971b"
	m11HarborSerpentOracle        = "599fb8b1-8750-4832-869a-cddc3784087a"
	m11DryadsFavorOracle          = "af740170-af87-487c-abab-a6ccdea82d97"
	m11VolcanicStrengthOracle     = "7c60b7bd-06cc-487f-ba42-b382ac429b12"
	m11QuagSicknessOracle         = "c7f2c629-0d3a-4556-951e-6bc5d74fd7bc"
	m11WarlordsAxeOracle          = "ebe43cc6-42fe-44f6-b198-ef0212bf82cb"
	m11BackToNatureOracle         = "b8d8ece6-397e-4bc9-b86c-7caa1d675d0f"
	m11BloodTitheOracle           = "c82665b7-d011-4bdc-b4a2-eca5861b16d2"
	m11CallToMindOracle           = "72e2b883-a284-4ab6-8174-1ca289141a49"
	m11HornetStingOracle          = "25edeff3-7727-4b80-9459-2bb6b8587fb9"
	m11HuntersFeastOracle         = "df206fb3-2d98-439e-9f1a-161bdbdc47b2"
	m11JacesIngenuityOracle       = "ce588fd3-642e-4df9-8061-6fcfc6e81270"
	m11StabbingPainOracle         = "604ad5b3-7f11-455e-9442-65b7484238e5"
	m11ThunderStrikeOracle        = "a5f25e80-7a6a-4629-8201-901e0fcd4a02"
	m11DiminishOracle             = "3e554ab4-8179-4d3e-84ba-9a67b51cff7f"
	m11InciteOracle               = "304da0e0-9852-4ba5-a896-3c6bd703a64c"
	m11TimeReversalOracle         = "9a63f6fc-dcf8-4529-a598-ca3d511f29e2"

	m11BearType   = "Creature — Bear"
	m11IslandType = "Basic Land — Island"
)

// m11AnswerSearchAllNamed takes every card the open search offers that is
// named `name`, and returns how many it took.
func m11AnswerSearchAllNamed(t *testing.T, g *game.Game, chooser uuid.UUID, name string) int {
	t.Helper()
	c := effectstest.SearchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	var picks []uuid.UUID
	for _, id := range c.SearchCards {
		var named bool
		g.ReadSnapshot(func() {
			card, ok := g.LookupCardForEffect(id)
			named = ok && card.Name == name
		})
		if named {
			picks = append(picks, id)
		}
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	return len(picks)
}

// m11DeclareAttacker walks to the declare attackers step and declares id
// against seat 1, returning what the engine said.
func m11DeclareAttacker(t *testing.T, g *game.Game, id uuid.UUID) error {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	return g.DeclareAttacker(id, g.Seats[1].ID)
}

// --- every card is in the catalog ---------------------------------------------

func TestM115579EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m11ArmoredCancrixOracle, m11BaronyVampireOracle, m11MaritimeGuardOracle, m11NetherHorrorOracle,
		m11StoneGolemOracle, m11AssaultGriffinOracle, m11CloudCrusaderOracle, m11GarruksCompanionOracle,
		m11GreaterBasiliskOracle, m11SacredWolfOracle, m11WallOfVinesOracle, m11RottingLegionOracle,
		m11AetherAdeptOracle, m11AuguryOwlOracle, m11LilianasSpecterOracle, m11ManicVandalOracle,
		m11TirelessMissionariesOracle, m11WarPriestOfThuneOracle, m11SquadronHawkOracle, m11RocEggOracle,
		m11ArcRunnerOracle, m11AirServantOracle, m11BrindleBoarOracle, m11EmberHaulerOracle,
		m11NightwingShadeOracle, m11WaterServantOracle, m11GargoyleSentinelOracle, m11BrittleEffigyOracle,
		m11PhantomBeastOracle, m11AjanisMantraOracle, m11LeylineOfVitalityOracle, m11EarthServantOracle,
		m11HarborSerpentOracle, m11DryadsFavorOracle, m11VolcanicStrengthOracle, m11QuagSicknessOracle,
		m11WarlordsAxeOracle, m11BackToNatureOracle, m11BloodTitheOracle, m11CallToMindOracle,
		m11HornetStingOracle, m11HuntersFeastOracle, m11JacesIngenuityOracle, m11StabbingPainOracle,
		m11ThunderStrikeOracle, m11DiminishOracle, m11InciteOracle, m11TimeReversalOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// The vanilla creatures carry no rules text at all.
func TestM115579VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{
		m11ArmoredCancrixOracle, m11BaronyVampireOracle, m11MaritimeGuardOracle, m11NetherHorrorOracle, m11StoneGolemOracle,
	} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.PrintedKeywords)+len(spec.Triggered)+len(spec.Activated)+len(spec.Static)+len(spec.Replacements) != 0 {
			t.Errorf("%s declares rules text", spec.Name)
		}
	}
}

// --- printed keywords -------------------------------------------------------------

func TestM115579KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Assault Griffin", "Creature — Griffin", m11AssaultGriffinOracle, []string{"flying"}},
		{"Cloud Crusader", "Creature — Human Knight", m11CloudCrusaderOracle, []string{"flying", "first strike"}},
		{"Garruk's Companion", "Creature — Beast", m11GarruksCompanionOracle, []string{"trample"}},
		{"Greater Basilisk", "Creature — Basilisk", m11GreaterBasiliskOracle, []string{"deathtouch"}},
		{"Sacred Wolf", "Creature — Wolf", m11SacredWolfOracle, []string{"hexproof"}},
		{"Wall of Vines", "Creature — Plant Wall", m11WallOfVinesOracle, []string{"defender", "reach"}},
		{"Augury Owl", "Creature — Bird", m11AuguryOwlOracle, []string{"flying"}},
		{"Liliana's Specter", "Creature — Specter", m11LilianasSpecterOracle, []string{"flying"}},
		{"Squadron Hawk", "Creature — Bird", m11SquadronHawkOracle, []string{"flying"}},
		{"Roc Egg", "Creature — Bird Egg", m11RocEggOracle, []string{"defender"}},
		{"Arc Runner", "Creature — Elemental Ox", m11ArcRunnerOracle, []string{"haste"}},
		{"Air Servant", "Creature — Elemental", m11AirServantOracle, []string{"flying"}},
		{"Nightwing Shade", "Creature — Shade", m11NightwingShadeOracle, []string{"flying"}},
		{"Gargoyle Sentinel", "Artifact Creature — Gargoyle", m11GargoyleSentinelOracle, []string{"defender"}},
		{"Harbor Serpent", "Creature — Serpent", m11HarborSerpentOracle, []string{"islandwalk"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Hexproof: an opponent's spell can't target Sacred Wolf.
func TestM115579SacredWolfCantBeTargetedByAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	wolf := effectstest.B12Push(g, opp.ID, "Sacred Wolf", "Creature — Wolf", m11SacredWolfOracle, 3, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Hornet Sting", "Instant", m11HornetStingOracle, effectstest.CardTarget(wolf)); err == nil {
		t.Error("an opponent's spell targeted a creature with hexproof")
	}
}

// --- enters ------------------------------------------------------------------------

func TestM115579RottingLegionEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Rotting Legion", "Creature — Zombie", m11RottingLegionOracle, 4, 5)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Rotting Legion did not resolve")
	}
	if !effectstest.HmlTapped(t, g, id) {
		t.Error("Rotting Legion entered untapped")
	}
}

func TestM115579AetherAdeptBouncesATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m11BearType, 2, 2)
	adept := effectstest.CastHoldCreature(t, g, "Aether Adept", "Creature — Human Wizard", m11AetherAdeptOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(adept) {
		t.Error("the Adept left the battlefield")
	}
}

func TestM115579AuguryOwlScriesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Augury Owl", "Creature — Bird", m11AuguryOwlOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 3 {
		t.Errorf("scrying %d cards, want 3", len(c.ScryCards))
	}
}

func TestM115579LilianasSpecterMakesAnOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, s := range g.Seats {
		effectstest.D2dSetHand(s, 2)
	}
	effectstest.CastHoldCreature(t, g, "Liliana's Specter", "Creature — Specter", m11LilianasSpecterOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	owed := 0
	for _, s := range g.Seats {
		if s.ID != me.ID {
			owed += effectstest.DiscardOwed(g, s.ID)
		}
	}
	if owed < 1 {
		t.Error("no opponent was asked to discard")
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 0 {
		t.Errorf("the controller owes %d discards", n)
	}
}

func TestM115579ManicVandalDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m11BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Manic Vandal", "Creature — Human Warrior", m11ManicVandalOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the artifact target")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestM115579TirelessMissionariesGainThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Tireless Missionaries", "Creature — Human Cleric", m11TirelessMissionariesOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestM115579WarPriestOfThuneMayDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.CastHoldCreature(t, g, "War Priest of Thune", "Creature — Human Cleric", m11WarPriestOfThuneOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, aura)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
}

func TestM115579SquadronHawkFetchesOtherHawks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: m11BearType},
		game.Card{Name: "Squadron Hawk", TypeLine: "Creature — Bird", OracleID: m11SquadronHawkOracle},
		game.Card{Name: "Squadron Hawk", TypeLine: "Creature — Bird", OracleID: m11SquadronHawkOracle},
	)
	effectstest.CastHoldCreature(t, g, "Squadron Hawk", "Creature — Bird", m11SquadronHawkOracle, 1, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after agreeing")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a card not named Squadron Hawk was offered")
	}
	if n := m11AnswerSearchAllNamed(t, g, me.ID, "Squadron Hawk"); n != 2 {
		t.Fatalf("the search offered %d Hawks, want 2", n)
	}
	hawks := 0
	for _, card := range me.Hand.Cards {
		if card.Name == "Squadron Hawk" {
			hawks++
		}
	}
	if hawks != 2 {
		t.Errorf("%d Hawks in hand, want 2", hawks)
	}
}

// --- leaves and end step -------------------------------------------------------------

func TestM115579RocEggMakesABirdWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.B12Push(g, me.ID, "Roc Egg", "Creature — Bird Egg", m11RocEggOracle, 0, 3)
	effectstest.B27Kill(g, egg)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	bird, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Bird")
	if !ok {
		t.Fatal("no Bird token")
	}
	if bird.Power != 3 || bird.Toughness != 3 {
		t.Errorf("the Bird is %d/%d, want 3/3", bird.Power, bird.Toughness)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bird.InstanceID, "flying") {
		t.Error("the Bird has no flying")
	}
}

func TestM115579ArcRunnerIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	runner := effectstest.B12Push(g, me.ID, "Arc Runner", "Creature — Elemental Ox", m11ArcRunnerOracle, 5, 1)
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(runner) {
		t.Error("Arc Runner survived the end step")
	}
}

func TestM115579PhantomBeastIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	beast := effectstest.B12Push(g, opp.ID, "Phantom Beast", "Creature — Illusion Beast", m11PhantomBeastOracle, 4, 5)
	effectstest.CastCatalogSpell(t, g, "Hornet Sting", "Instant", m11HornetStingOracle, effectstest.CardTarget(beast))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beast) {
		t.Error("the Beast survived being targeted")
	}
}

// --- activated abilities -------------------------------------------------------------

func TestM115579AirServantTapsOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	servant := effectstest.B12Push(g, me.ID, "Air Servant", "Creature — Elemental", m11AirServantOracle, 4, 3)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", m11BearType, 2, 2)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.ApaMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, servant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(walker)}); err == nil {
		t.Fatal("Air Servant targeted a creature without flying")
	}
	effectstest.ApaActivate(t, g, me, servant, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(flyer)})
	if !effectstest.HmlTapped(t, g, flyer) {
		t.Error("the flyer is untapped")
	}
}

func TestM115579BrindleBoarGainsFourAsItIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	boar := effectstest.B12Push(g, me.ID, "Brindle Boar", "Creature — Boar", m11BrindleBoarOracle, 2, 2)
	life := me.Life
	effectstest.ApaActivate(t, g, me, boar, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(boar) {
		t.Error("the Boar was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestM115579EmberHaulerDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	hauler := effectstest.B12Push(g, me.ID, "Ember Hauler", "Creature — Goblin", m11EmberHaulerOracle, 2, 2)
	life := opp.Life
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, hauler, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(hauler) {
		t.Error("the Hauler was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
}

func TestM115579NightwingShadePumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	shade := effectstest.B12Push(g, me.ID, "Nightwing Shade", "Creature — Shade", m11NightwingShadeOracle, 2, 2)
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, shade, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, shade); p != 3 || tough != 3 {
		t.Errorf("the Shade is %d/%d, want 3/3", p, tough)
	}
}

func TestM115579WaterServantTradesEachWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	servant := effectstest.B12Push(g, me.ID, "Water Servant", "Creature — Elemental", m11WaterServantOracle, 3, 4)
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, servant, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, servant); p != 4 || tough != 3 {
		t.Errorf("after +1/-1 the Servant is %d/%d, want 4/3", p, tough)
	}
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, servant, 1, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, servant); p != 3 || tough != 4 {
		t.Errorf("after -1/+1 the Servant is %d/%d, want 3/4", p, tough)
	}
}

func TestM115579GargoyleSentinelLosesDefenderAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	gargoyle := effectstest.B12Push(g, me.ID, "Gargoyle Sentinel", "Artifact Creature — Gargoyle", m11GargoyleSentinelOracle, 3, 3)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, gargoyle, 0, game.ActivateAbilityParams{})
	if effectstest.HasEffectiveKeyword(t, g, gargoyle, "defender") {
		t.Error("the Sentinel still has defender")
	}
	if !effectstest.HasEffectiveKeyword(t, g, gargoyle, "flying") {
		t.Error("the Sentinel has no flying")
	}
}

func TestM115579BrittleEffigyExilesACreatureAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effigy := effectstest.PushCatalogPermanent(g, me.ID, "Brittle Effigy", "Artifact", m11BrittleEffigyOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m11BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, effigy, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	if !g.Exile.Contains(effigy) {
		t.Error("the Effigy was not exiled to pay its cost")
	}
}

// --- triggered and static abilities ---------------------------------------------------

func TestM115579AjanisMantraMayGainOneEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Ajani's Mantra", "Enchantment", m11AjanisMantraOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life := me.Life
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestM115579LeylineOfVitalityToughensAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	if spec, _ := effects.Lookup(m11LeylineOfVitalityOracle); spec.OpeningHand == nil {
		t.Error("Leyline of Vitality has no opening-hand action")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Leyline of Vitality", "Enchantment", m11LeylineOfVitalityOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m11BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m11BearType, 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 2/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Bear", m11BearType, "", 2, 2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after a creature entered, want %d", me.Life, life+1)
	}
}

func TestM115579EarthServantCountsYourMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Mountain", "Basic Land — Mountain")
	servant := effectstest.B12Push(g, me.ID, "Earth Servant", "Creature — Elemental", m11EarthServantOracle, 4, 4)
	if p, tough := effectstest.PtOf(t, g, servant); p != 4 || tough != 6 {
		t.Errorf("the Servant is %d/%d, want 4/6", p, tough)
	}
}

func TestM115579HarborSerpentNeedsFiveIslands(t *testing.T) {
	setup := func(islands int) (*game.Game, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < islands; i++ {
			owner := me.ID
			if i%2 == 1 {
				owner = opp.ID
			}
			effectstest.SeedLandOnBattlefield(g, owner, "Island", m11IslandType)
		}
		return g, effectstest.B12Push(g, me.ID, "Harbor Serpent", "Creature — Serpent", m11HarborSerpentOracle, 5, 5)
	}
	g, serpent := setup(4)
	if err := m11DeclareAttacker(t, g, serpent); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("with four Islands: err = %v, want ErrCantAttack", err)
	}
	g, serpent = setup(5)
	if err := m11DeclareAttacker(t, g, serpent); err != nil {
		t.Fatalf("with five Islands, two of them an opponent's: %v", err)
	}
}

// --- Auras and Equipment ---------------------------------------------------------------

func TestM115579DryadsFavorGrantsForestwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", m11BearType, 2, 2)
	effectstest.AuraCast(t, g, "Dryad's Favor", m11DryadsFavorOracle, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "forestwalk") {
		t.Error("the enchanted creature has no forestwalk")
	}
}

func TestM115579VolcanicStrengthPumpsAndGrantsMountainwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", m11BearType, 2, 2)
	effectstest.AuraCast(t, g, "Volcanic Strength", m11VolcanicStrengthOracle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "mountainwalk") {
		t.Error("the enchanted creature has no mountainwalk")
	}
}

func TestM115579QuagSicknessCountsTheAurasControllersSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.AuraCast(t, g, "Quag Sickness", m11QuagSicknessOracle, theirs)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the enchanted creature is %d/%d, want 1/1", p, tough)
	}
}

func TestM115579WarlordsAxePumpsTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	axe := effectstest.PushCatalogPermanent(g, me.ID, "Warlord's Axe", effectstest.EquipTypeLine, m11WarlordsAxeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m11BearType, 2, 2)
	effectstest.EquipTo(t, g, me.ID, axe, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("the equipped creature is %d/%d, want 5/3", p, tough)
	}
}

// --- the spells -------------------------------------------------------------------------

func TestM115579BackToNatureDestroysEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m11BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Back to Nature", "Instant", m11BackToNatureOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestM115579BloodTitheDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.CastCatalogSpell(t, g, "Blood Tithe", "Sorcery", m11BloodTitheOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	opponents := 0
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		opponents++
		if s.Life != lives[s.ID]-3 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]-3)
		}
	}
	if me.Life != lives[me.ID]+3*opponents {
		t.Errorf("life %d, want %d", me.Life, lives[me.ID]+3*opponents)
	}
}

func TestM115579CallToMindReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	good := effectstest.WthGraveyardCard(me, "Good", "Instant")
	bad := effectstest.WthGraveyardCard(me, "Bad", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Call to Mind", "Sorcery", m11CallToMindOracle, effectstest.CardTarget(bad)); err == nil {
		t.Fatal("Call to Mind targeted an artifact card")
	}
	effectstest.CastCatalogSpell(t, g, "Call to Mind", "Sorcery", m11CallToMindOracle, effectstest.CardTarget(good))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(good) || me.Graveyard.Contains(good) {
		t.Error("the card did not return to hand")
	}
}

func TestM115579HornetStingDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Hornet Sting", "Instant", m11HornetStingOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestM115579HuntersFeastGainsSixForEachTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.CastCatalogSpell(t, g, "Hunters' Feast", "Sorcery", m11HuntersFeastOracle,
		[]game.TargetRef{effectstest.PlayerTarget(me.ID), effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID]
		if s.ID == me.ID || s.ID == opp.ID {
			want += 6
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestM115579JacesIngenuityDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Jace's Ingenuity", "Instant", m11JacesIngenuityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand is %d, want %d", got, hand+3)
	}
}

func TestM115579StabbingPainShrinksAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Stabbing Pain", "Instant", m11StabbingPainOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 2 || tough != 2 {
		t.Errorf("the Ogre is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HmlTapped(t, g, ogre) {
		t.Error("the Ogre is untapped")
	}
}

func TestM115579ThunderStrikePumpsAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m11BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Thunder Strike", "Instant", m11ThunderStrikeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear has no first strike")
	}
}

func TestM115579DiminishSetsBaseOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Diminish", "Instant", m11DiminishOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1", p, tough)
	}
}

func TestM115579InciteMakesItRedAndAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", m11BearType, 2, 2)
	effectstest.HmlSetColors(g, bear, "G")
	effectstest.CastCatalogSpell(t, g, "Incite", "Instant", m11InciteOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	var colors []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear {
				colors = c.Effective().Colors
			}
		}
	})
	if len(colors) != 1 || colors[0] != "R" {
		t.Errorf("the Bear's colours are %v, want only red", colors)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Errorf("pass with the Bear home = %v, want a refusal naming it", err)
	}
}

func TestM115579TimeReversalWheelsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.D2dSetHand(me, 2)
	effectstest.D2dSetHand(opp, 4)
	mine := effectstest.PushGraveyardCardForTest(me, "My Fodder")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Fodder")
	spell := effectstest.CastCatalogSpell(t, g, "Time Reversal", "Sorcery", m11TimeReversalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range []*game.Player{me, opp} {
		if got := s.Hand.Size(); got != 7 {
			t.Errorf("%s holds %d cards, want 7", s.Name, got)
		}
	}
	if me.Graveyard.Contains(mine) || opp.Graveyard.Contains(theirs) {
		t.Error("a graveyard card was not shuffled away")
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Time Reversal was not exiled")
	}
}
