package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m12_5573_b1_cards_test.go — the first batch of Magic 2012 cards
// (#5573). Every non-vanilla test plays the real catalog entry through
// the engine.

const (
	m12B1AmphinCutthroatOracle    = "e766a5eb-684b-4939-b164-6093d15600c9"
	m12B1ArmoredWarhorseOracle    = "29233aa4-3fa0-4074-ad0f-c8e7101052f3"
	m12B1BonebreakerGiantOracle   = "69166587-3db4-4eb0-b153-b75021be9a36"
	m12B1WallOfTorchesOracle      = "674a269c-61ed-4fb5-ad74-422e581a5a91"
	m12B1AvenFleetwingOracle      = "37ad9318-6690-4a8d-bcea-9fcbca58769d"
	m12B1PeregrineGriffinOracle   = "f8848065-7075-41d8-82f6-67ace390625a"
	m12B1RustedSentinelOracle     = "bab28608-b388-4c48-92a4-689943bbbf4a"
	m12B1TormentedSoulOracle      = "ee334bd7-d344-4684-85c8-d0464bbaf2b4"
	m12B1SkywinderDrakeOracle     = "e894ccf5-fc71-48bd-bca5-cfeb05315f42"
	m12B1BloodOgreOracle          = "8bbe3f0d-41d0-4c58-ad65-6675a234cad3"
	m12B1BloodrageVampireOracle   = "b1b6a840-2dc6-4af3-869a-b3bfe64ff7c0"
	m12B1CarnageWurmOracle        = "007161cb-a44b-4443-a53a-dfaacc6cfb86"
	m12B1DuskhunterBatOracle      = "6082f84a-3af2-4007-ab88-7b4b59b5c154"
	m12B1FurybornHellkiteOracle   = "8c4aba38-7fbd-463e-8454-21037090c71e"
	m12B1GorehornMinotaursOracle  = "be8ad29c-2232-426d-bda9-a289dcc6c265"
	m12B1LurkingCrocodileOracle   = "b0df3389-5959-4ee2-a921-9324732f7ce8"
	m12B1VampireOutcastsOracle    = "25ad4e88-b5ee-4853-a4fc-0a0420c746f6"
	m12B1AzureMageOracle          = "bbf9fd47-10e1-4749-b06f-4fe72200afc7"
	m12B1CrimsonMageOracle        = "d7bf23c4-172a-42fd-b826-3b074a588610"
	m12B1OnyxMageOracle           = "8a398777-2fa5-4cac-adee-4582e59d410d"
	m12B1DriftingShadeOracle      = "fa3883c7-4fb4-48b9-847b-9a94d1988dd7"
	m12B1GideonsLawkeeperOracle   = "bae86bb0-5a97-4f6a-8a3f-4e6848bfacc0"
	m12B1JadeMageOracle           = "6daf3c59-7636-4d44-872b-67722b4868cb"
	m12B1MerfolkMesmeristOracle   = "590d524f-278d-4cf4-ae3f-efcced4517b1"
	m12B1ArbalestEliteOracle      = "1d79b7a8-2b4c-4893-afdb-879c74957288"
	m12B1ManalithOracle           = "bd9e416a-89b3-4912-be9b-49fce6a93dc9"
	m12B1GreatswordOracle         = "51024f9a-6c6e-4555-a156-92b32255a0c9"
	m12B1KiteShieldOracle         = "24f9c64a-6ce6-41cd-948e-11554d7f17fb"
	m12B1DarkFavorOracle          = "c1f4a440-638a-4e3e-8a3e-af4a4a73ec91"
	m12B1TrollhideOracle          = "397df813-c759-4246-8def-122c7ff82d88"
	m12B1BenalishVeteranOracle    = "5da2e5c7-ec04-4f1f-bb2f-7572f877393c"
	m12B1ChasmDrakeOracle         = "e0f44bb8-f07b-416c-b767-87b9fe6461aa"
	m12B1PrideGuardianOracle      = "6b2d396f-802d-4b49-8418-5a96b8a12226"
	m12B1StingerflingSpiderOracle = "056770e2-8ab7-424d-a2d8-7be6d6cc73f1"
	m12B1StonehornDignitaryOracle = "65368569-68bf-4498-9979-d2935942ca49"
	m12B1PhantasmalBearOracle     = "780da643-8d14-4f1a-a298-5b7454e9c4df"
	m12B1PhantasmalDragonOracle   = "7f32c581-4d4a-4873-8e4b-b95e13790481"
	m12B1GideonsAvengerOracle     = "5ca79852-d90d-47b0-8427-4da90779e205"
	m12B1SorinsThirstOracle       = "ff27ff37-96c0-41af-8881-a078e884e67b"
	m12B1SorinsVengeanceOracle    = "75d9c036-4f0d-4b55-b0a4-096ca84748ca"
	m12B1TasteOfBloodOracle       = "8e5026da-4859-4f30-824c-a49ac308e91a"
	m12B1WringFleshOracle         = "45bb536c-c173-4b6e-a5a8-354c74fd93a1"
	m12B1GuardiansPledgeOracle    = "32d78ffd-5bed-45e8-be6a-420149a263bb"
	m12B1HideousVisageOracle      = "bd9c5b53-9666-45ee-8f9d-ee1ae74111b6"
	m12B1StaveOffOracle           = "c77c23bc-efde-43ce-bf6f-7a8531d81d89"
	m12B1TectonicRiftOracle       = "b94fcdd2-7e50-439c-b5eb-aaa9ec7dee08"
	m12B1GarrukPrimalHunterOracle = "298a7e1a-b254-442e-bc1d-2a3e5296d849"
	m12B1JaceMemoryAdeptOracle    = "4a4dc133-541f-456b-a95f-3d707e1c95e9"

	m12B1BearType = "Creature — Bear"
)

// m12B1FillLibrary puts n filler cards on top of p's library.
func m12B1FillLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// m12B1Flyer seeds a non-catalog 1/1 creature with flying.
func m12B1Flyer(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
}

// --- every card is in the catalog ----------------------------------------

func TestM12B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m12B1AmphinCutthroatOracle, m12B1ArmoredWarhorseOracle, m12B1BonebreakerGiantOracle, m12B1WallOfTorchesOracle,
		m12B1AvenFleetwingOracle, m12B1PeregrineGriffinOracle, m12B1RustedSentinelOracle, m12B1TormentedSoulOracle,
		m12B1SkywinderDrakeOracle, m12B1BloodOgreOracle, m12B1BloodrageVampireOracle, m12B1CarnageWurmOracle,
		m12B1DuskhunterBatOracle, m12B1FurybornHellkiteOracle, m12B1GorehornMinotaursOracle, m12B1LurkingCrocodileOracle,
		m12B1VampireOutcastsOracle, m12B1AzureMageOracle, m12B1CrimsonMageOracle, m12B1OnyxMageOracle,
		m12B1DriftingShadeOracle, m12B1GideonsLawkeeperOracle, m12B1JadeMageOracle, m12B1MerfolkMesmeristOracle,
		m12B1ArbalestEliteOracle, m12B1ManalithOracle, m12B1GreatswordOracle, m12B1KiteShieldOracle,
		m12B1DarkFavorOracle, m12B1TrollhideOracle, m12B1BenalishVeteranOracle, m12B1ChasmDrakeOracle,
		m12B1PrideGuardianOracle, m12B1StingerflingSpiderOracle, m12B1StonehornDignitaryOracle, m12B1PhantasmalBearOracle,
		m12B1PhantasmalDragonOracle, m12B1GideonsAvengerOracle, m12B1SorinsThirstOracle, m12B1SorinsVengeanceOracle,
		m12B1TasteOfBloodOracle, m12B1WringFleshOracle, m12B1GuardiansPledgeOracle, m12B1HideousVisageOracle,
		m12B1StaveOffOracle, m12B1TectonicRiftOracle, m12B1GarrukPrimalHunterOracle, m12B1JaceMemoryAdeptOracle,
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

func TestM12B1VanillaCreaturesDeclareNoRules(t *testing.T) {
	for _, oracle := range []string{m12B1AmphinCutthroatOracle, m12B1ArmoredWarhorseOracle, m12B1BonebreakerGiantOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 || len(spec.Replacements) != 0 {
			t.Errorf("%s declares rules: %+v", spec.Name, spec)
		}
	}
}

// --- printed keywords -----------------------------------------------------------

func TestM12B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Wall of Torches", "Creature — Wall", m12B1WallOfTorchesOracle, []string{"defender"}},
		{"Aven Fleetwing", "Creature — Bird Soldier", m12B1AvenFleetwingOracle, []string{"flying", "hexproof"}},
		{"Peregrine Griffin", "Creature — Griffin", m12B1PeregrineGriffinOracle, []string{"flying", "first strike"}},
		{"Skywinder Drake", "Creature — Drake", m12B1SkywinderDrakeOracle, []string{"flying"}},
		{"Blood Ogre", "Creature — Ogre Warrior", m12B1BloodOgreOracle, []string{"first strike"}},
		{"Carnage Wurm", "Creature — Wurm", m12B1CarnageWurmOracle, []string{"trample"}},
		{"Duskhunter Bat", "Creature — Bat", m12B1DuskhunterBatOracle, []string{"flying"}},
		{"Furyborn Hellkite", "Creature — Dragon", m12B1FurybornHellkiteOracle, []string{"flying"}},
		{"Lurking Crocodile", "Creature — Crocodile", m12B1LurkingCrocodileOracle, []string{"islandwalk"}},
		{"Vampire Outcasts", "Creature — Vampire", m12B1VampireOutcastsOracle, []string{"lifelink"}},
		{"Drifting Shade", "Creature — Shade", m12B1DriftingShadeOracle, []string{"flying"}},
		{"Chasm Drake", "Creature — Drake", m12B1ChasmDrakeOracle, []string{"flying"}},
		{"Pride Guardian", "Creature — Cat Monk", m12B1PrideGuardianOracle, []string{"defender"}},
		{"Stingerfling Spider", "Creature — Spider", m12B1StingerflingSpiderOracle, []string{"reach"}},
		{"Phantasmal Dragon", "Creature — Dragon Illusion", m12B1PhantasmalDragonOracle, []string{"flying"}},
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

// Hexproof: an opponent's spell can't target Aven Fleetwing.
func TestM12B1AvenFleetwingCantBeTargetedByAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	fleetwing := effectstest.B12Push(g, opp.ID, "Aven Fleetwing", "Creature — Bird Soldier", m12B1AvenFleetwingOracle, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Wring Flesh", "Instant", m12B1WringFleshOracle, effectstest.CardTarget(fleetwing)); err == nil {
		t.Error("an opponent's spell targeted a creature with hexproof")
	}
}

// --- bloodthirst ----------------------------------------------------------------

func TestM12B1BloodthirstNeedsADamagedOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	vampire := effectstest.CastHoldCreature(t, g, "Bloodrage Vampire", "Creature — Vampire", m12B1BloodrageVampireOracle, 3, 1)
	if n := effectstest.CounterCount(g, vampire, game.CounterPlusOne); n != 0 {
		t.Errorf("with no opponent damaged, %d +1/+1 counters", n)
	}
}

func TestM12B1BloodthirstCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, t, n                int
	}{
		{"Blood Ogre", "Creature — Ogre Warrior", m12B1BloodOgreOracle, 2, 2, 1},
		{"Bloodrage Vampire", "Creature — Vampire", m12B1BloodrageVampireOracle, 3, 1, 1},
		{"Carnage Wurm", "Creature — Wurm", m12B1CarnageWurmOracle, 6, 6, 3},
		{"Duskhunter Bat", "Creature — Bat", m12B1DuskhunterBatOracle, 1, 1, 1},
		{"Furyborn Hellkite", "Creature — Dragon", m12B1FurybornHellkiteOracle, 6, 6, 6},
		{"Gorehorn Minotaurs", "Creature — Minotaur Warrior", m12B1GorehornMinotaursOracle, 3, 3, 2},
		{"Lurking Crocodile", "Creature — Crocodile", m12B1LurkingCrocodileOracle, 2, 2, 1},
		{"Vampire Outcasts", "Creature — Vampire", m12B1VampireOutcastsOracle, 2, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, opp := effectstest.GptB1Seats(g)
			effectstest.GptB1DamagePlayer(g, opp.ID, 1)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.t)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", n, tc.n)
			}
		})
	}
}

// --- statics and replacements ------------------------------------------------------

func TestM12B1RustedSentinelEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sentinel := effectstest.CastHoldCreature(t, g, "Rusted Sentinel", "Artifact Creature — Golem", m12B1RustedSentinelOracle, 3, 4)
	if !effectstest.HmlTapped(t, g, sentinel) {
		t.Error("Rusted Sentinel entered untapped")
	}
}

func TestM12B1TormentedSoulCantBlockAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	soul := effectstest.B12Push(g, g.Seats[0].ID, "Tormented Soul", "Creature — Spirit", m12B1TormentedSoulOracle, 1, 1)
	r := effectstest.AuraRestrictions(t, g, soul)
	if r&game.CantBlock == 0 {
		t.Error("Tormented Soul can block")
	}
	if r&game.CantBeBlocked == 0 {
		t.Error("Tormented Soul can be blocked")
	}
}

func TestM12B1SkywinderDrakeBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", m12B1BearType, 2, 2)
	flyer := m12B1Flyer(g, me.ID)
	drake := effectstest.B12Push(g, opp.ID, "Skywinder Drake", "Creature — Drake", m12B1SkywinderDrakeOracle, 3, 1)
	effectstest.BrAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(drake, walker); err == nil {
		t.Error("Skywinder Drake blocked a creature without flying")
	}
	if err := g.DeclareBlocker(drake, flyer); err != nil {
		t.Errorf("Skywinder Drake can't block a flyer: %v", err)
	}
}

// --- activated abilities -------------------------------------------------------------

func TestM12B1AzureMageDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mage := effectstest.B12Push(g, me.ID, "Azure Mage", "Creature — Human Wizard", m12B1AzureMageOracle, 2, 1)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C", "C", "U")
	effectstest.ApaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestM12B1MagesGrantAKeywordToYourCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		mana                  []string
	}{
		{"Crimson Mage", m12B1CrimsonMageOracle, "haste", []string{"R"}},
		{"Onyx Mage", m12B1OnyxMageOracle, "deathtouch", []string{"C", "B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			mage := effectstest.B12Push(g, me.ID, tc.name, "Creature — Human", tc.oracle, 2, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m12B1BearType, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m12B1BearType, 2, 2)
			if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
				t.Fatal("the Mage targeted a creature you don't control")
			}
			effectstest.ApaMana(me, tc.mana...)
			effectstest.ApaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the target has no %s", tc.keyword)
			}
		})
	}
}

func TestM12B1DriftingShadePumpsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	shade := effectstest.B12Push(g, me.ID, "Drifting Shade", "Creature — Shade", m12B1DriftingShadeOracle, 1, 1)
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, shade, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, shade); p != 2 || tough != 2 {
		t.Errorf("Drifting Shade is %d/%d, want 2/2", p, tough)
	}
}

// Gideon's Lawkeeper taps an opponent's creature, and Gideon's Avenger
// grows from that; the Lawkeeper tapping itself to pay is no opponent's.
func TestM12B1LawkeeperTapsAndAvengerGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	avenger := effectstest.B12Push(g, me.ID, "Gideon's Avenger", "Creature — Human Soldier", m12B1GideonsAvengerOracle, 2, 2)
	lawkeeper := effectstest.B12Push(g, me.ID, "Gideon's Lawkeeper", "Creature — Human Soldier", m12B1GideonsLawkeeperOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m12B1BearType, 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, lawkeeper, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Fatal("the target is untapped")
	}
	if n := effectstest.CounterCount(g, avenger, game.CounterPlusOne); n != 1 {
		t.Errorf("Gideon's Avenger has %d +1/+1 counters, want 1", n)
	}
}

func TestM12B1JadeMageMakesASaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mage := effectstest.B12Push(g, me.ID, "Jade Mage", "Creature — Human Shaman", m12B1JadeMageOracle, 2, 1)
	before := effectstest.B27CountNamed(g, me.ID, "Saproling")
	effectstest.ApaMana(me, "C", "C", "G")
	effectstest.ApaActivate(t, g, me, mage, 0, game.ActivateAbilityParams{})
	if got := effectstest.B27CountNamed(g, me.ID, "Saproling"); got != before+1 {
		t.Errorf("%d Saprolings, want %d", got, before+1)
	}
}

func TestM12B1MerfolkMesmeristMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mesmerist := effectstest.B12Push(g, me.ID, "Merfolk Mesmerist", "Creature — Merfolk Wizard", m12B1MerfolkMesmeristOracle, 1, 2)
	library := opp.Library.Size()
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, mesmerist, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestM12B1ArbalestEliteShootsACombatantAndStaysTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elite := effectstest.B12Push(g, me.ID, "Arbalest Elite", "Creature — Human Archer", m12B1ArbalestEliteOracle, 2, 3)
	raider := effectstest.B12Creature(g, me.ID, "Raider", m12B1BearType, 2, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle", m12B1BearType, 2, 4)
	effectstest.BrAttack(t, g, raider)
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("the Elite targeted a creature that isn't in combat")
	}
	effectstest.ApaMana(me, "C", "C", "W")
	effectstest.ApaActivate(t, g, me, elite, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(raider)})
	if got := effectstest.FindCardAnywhere(t, g, raider).DamageMarked; got != 3 {
		t.Errorf("the attacker has %d damage, want 3", got)
	}
	if !effectstest.HmlTapped(t, g, elite) {
		t.Error("the Elite did not tap")
	}
	skips := effectstest.FindCardAnywhere(t, g, elite).NextUntapSkips
	if len(skips) != 1 || skips[0].Player != me.ID {
		t.Errorf("the Elite's untap skips are %+v, want one for its controller", skips)
	}
}

func TestM12B1ManalithAddsAnyColor(t *testing.T) {
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := effectstest.Top100ActiveSeat(g)
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		rock := effectstest.B12Push(g, me.ID, "Manalith", "Artifact", m12B1ManalithOracle, 0, 0)
		if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("add {%s}: %v", colour, err)
		}
		if got := me.ManaPool.CountColor(colour); got != 1 {
			t.Errorf("pool holds %d {%s}, want 1", got, colour)
		}
		if !effectstest.HmlTapped(t, g, rock) {
			t.Error("Manalith did not tap")
		}
	}
}

// --- equipment and auras ---------------------------------------------------------

func TestM12B1EquipmentPumpsTheEquippedCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		p, t         int
	}{
		{"Greatsword", m12B1GreatswordOracle, 5, 2},
		{"Kite Shield", m12B1KiteShieldOracle, 2, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			gear := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact — Equipment", tc.oracle, false)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m12B1BearType, 2, 2)
			effectstest.ApaMana(me, "C", "C", "C")
			effectstest.ApaActivate(t, g, me, gear, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.t {
				t.Errorf("the equipped Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.t)
			}
		})
	}
}

func TestM12B1DarkFavorPumpsAndCostsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m12B1BearType, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Dark Favor", "Enchantment — Aura", m12B1DarkFavorOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/3", p, tough)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestM12B1TrollhidePumpsAndGrantsRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m12B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Trollhide", "Enchantment — Aura", m12B1TrollhideOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/4", p, tough)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted ability")
	}
	effectstest.FloatForTest(g, me, "GG")
	effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the regenerated Bear was destroyed")
	}
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("regeneration did not tap the Bear")
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestM12B1BenalishVeteranGrowsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	veteran := effectstest.B12Push(g, me.ID, "Benalish Veteran", "Creature — Human Soldier", m12B1BenalishVeteranOracle, 2, 2)
	effectstest.AttackWithTrigger(t, g, veteran, opp.ID, uuid.Nil)
	if p, tough := effectstest.PtOf(t, g, veteran); p != 3 || tough != 3 {
		t.Errorf("the attacking Veteran is %d/%d, want 3/3", p, tough)
	}
}

func TestM12B1ChasmDrakeGivesFlyingWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.B12Push(g, me.ID, "Chasm Drake", "Creature — Drake", m12B1ChasmDrakeOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m12B1BearType, 2, 2)
	effectstest.AttackWithTrigger(t, g, drake, opp.ID, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target has no flying")
	}
}

func TestM12B1PrideGuardianGainsLifeWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := effectstest.B12Push(g, opp.ID, "Pride Guardian", "Creature — Cat Monk", m12B1PrideGuardianOracle, 0, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", m12B1BearType, 2, 2)
	life := opp.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: guardian, Target: attacker, Amount: 1})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+3 {
		t.Errorf("life %d, want %d", opp.Life, life+3)
	}
}

func TestM12B1StingerflingSpiderMayDestroyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	flyer := m12B1Flyer(g, opp.ID)
	effectstest.CastHoldCreature(t, g, "Stingerfling Spider", "Creature — Spider", m12B1StingerflingSpiderOracle, 2, 5)
	effectstest.AnswerPickTargetIfOpen(t, g, flyer)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the creature with flying survived")
	}
}

func TestM12B1StonehornDignitarySkipsAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.CastHoldCreature(t, g, "Stonehorn Dignitary", "Creature — Rhino Soldier", m12B1StonehornDignitaryOracle, 1, 4)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.NextCombatPhaseSkipsForLocked(opp.ID); n != 1 {
		t.Errorf("the target owes %d combat skips, want 1", n)
	}
	if n := g.NextCombatPhaseSkipsForLocked(me.ID); n != 0 {
		t.Errorf("the controller owes %d combat skips, want 0", n)
	}
}

// Targeted by any spell, even its controller's own Stave Off, the
// Illusion is sacrificed before the spell resolves.
func TestM12B1PhantasmalCreaturesAreSacrificedWhenTargeted(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Phantasmal Bear", "Creature — Bear Illusion", m12B1PhantasmalBearOracle},
		{"Phantasmal Dragon", "Creature — Dragon Illusion", m12B1PhantasmalDragonOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.CastCatalogSpell(t, g, "Stave Off", "Instant", m12B1StaveOffOracle, effectstest.CardTarget(id))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
				t.Errorf("%s was not sacrificed", tc.name)
			}
		})
	}
}

// --- instants and sorceries -------------------------------------------------------

func TestM12B1SorinsThirstDrainsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m12B1BearType, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sorin's Thirst", "Instant", m12B1SorinsThirstOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestM12B1DrainsAPlayer(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Sorin's Vengeance", m12B1SorinsVengeanceOracle, 10},
		{"Taste of Blood", m12B1TasteOfBloodOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			mine, theirs := me.Life, opp.Life
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != theirs-tc.n {
				t.Errorf("the target is at %d, want %d", opp.Life, theirs-tc.n)
			}
			if me.Life != mine+tc.n {
				t.Errorf("the caster is at %d, want %d", me.Life, mine+tc.n)
			}
		})
	}
}

func TestM12B1WringFleshShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Wring Flesh", "Instant", m12B1WringFleshOracle, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, giant); p != 1 || tough != 3 {
		t.Errorf("the target is %d/%d, want 1/3", p, tough)
	}
}

func TestM12B1GuardiansPledgePumpsOnlyYourWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	white := effectstest.B12Creature(g, me.ID, "White", "Creature — Human", 2, 2)
	effectstest.HmlSetColors(g, white, "W")
	green := effectstest.B12Creature(g, me.ID, "Green", m12B1BearType, 2, 2)
	effectstest.HmlSetColors(g, green, "G")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human", 2, 2)
	effectstest.HmlSetColors(g, theirs, "W")
	effectstest.CastCatalogSpell(t, g, "Guardians' Pledge", "Instant", m12B1GuardiansPledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, white); p != 4 || tough != 4 {
		t.Errorf("your white creature is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, green); p != 2 {
		t.Errorf("your green creature has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's white creature has power %d, want 2", p)
	}
}

func TestM12B1HideousVisageGivesYourCreaturesIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m12B1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m12B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hideous Visage", "Sorcery", m12B1HideousVisageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "intimidate") {
		t.Error("your creature has no intimidate")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "intimidate") {
		t.Error("an opponent's creature has intimidate")
	}
}

func TestM12B1StaveOffGivesProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m12B1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stave Off", "Instant", m12B1StaveOffOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "protection from red") {
		t.Error("the target did not gain protection from red")
	}
}

func TestM12B1TectonicRiftDestroysALandAndGroundsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	walker := effectstest.B12Creature(g, me.ID, "Walker", m12B1BearType, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m12B1BearType, 2, 2)
	flyer := m12B1Flyer(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Tectonic Rift", "Sorcery", m12B1TectonicRiftOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land survived")
	}
	effectstest.BrAttack(t, g, walker)
	if err := g.DeclareBlocker(bear, walker); err == nil {
		t.Error("a creature without flying blocked")
	}
	if err := g.DeclareBlocker(flyer, walker); err != nil {
		t.Errorf("a creature with flying can't block: %v", err)
	}
}

// --- planeswalkers ---------------------------------------------------------------

func TestM12B1GarrukPrimalHunterPlusOneMakesABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	garruk := effectstest.PushWalkerForTest(g, me.ID, "Garruk, Primal Hunter", m12B1GarrukPrimalHunterOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	before := effectstest.CountNamed(g, "Beast")
	effectstest.LrwB1Activate(t, g, me, garruk, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Beast") != before+1 {
		t.Error("no Beast token")
	}
	if got := effectstest.LoyaltyOf(g, garruk); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestM12B1GarrukPrimalHunterMinusThreeDrawsByGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	garruk := effectstest.PushWalkerForTest(g, me.ID, "Garruk, Primal Hunter", m12B1GarrukPrimalHunterOracle, 5)
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
	effectstest.B12Creature(g, me.ID, "Small", m12B1BearType, 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := me.Hand.Size()
	effectstest.LrwB1Activate(t, g, me, garruk, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d, want %d", got, hand+4)
	}
}

func TestM12B1GarrukPrimalHunterMinusSixMakesAWurmPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	garruk := effectstest.PushWalkerForTest(g, me.ID, "Garruk, Primal Hunter", m12B1GarrukPrimalHunterOracle, 6)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.LrwB1Activate(t, g, me, garruk, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Wurm"); got != 3 {
		t.Errorf("%d Wurms, want 3", got)
	}
}

func TestM12B1JaceMemoryAdeptPlusOneDrawsAndMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace, Memory Adept", m12B1JaceMemoryAdeptOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand, library := me.Hand.Size(), opp.Library.Size()
	effectstest.LrwB1Activate(t, g, me, jace, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := library - opp.Library.Size(); got != 1 {
		t.Errorf("the target milled %d, want 1", got)
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestM12B1JaceMemoryAdeptZeroMillsTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	m12B1FillLibrary(opp, 10)
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace, Memory Adept", m12B1JaceMemoryAdeptOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	library := opp.Library.Size()
	effectstest.LrwB1Activate(t, g, me, jace, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 10 {
		t.Errorf("the target milled %d, want 10", got)
	}
}

func TestM12B1JaceMemoryAdeptMinusSevenEachTargetDrawsTwenty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	other := g.Seats[(seat+2)%len(g.Seats)]
	m12B1FillLibrary(opp, 25)
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace, Memory Adept", m12B1JaceMemoryAdeptOracle, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand, untargeted := opp.Hand.Size(), other.Hand.Size()
	effectstest.LrwB1Activate(t, g, me, jace, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand+20 {
		t.Errorf("the target's hand %d, want %d", got, hand+20)
	}
	if got := other.Hand.Size(); got != untargeted {
		t.Errorf("a player not targeted drew %d", got-untargeted)
	}
}
