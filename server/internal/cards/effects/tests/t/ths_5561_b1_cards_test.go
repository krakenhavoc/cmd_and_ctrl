package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ths_5561_b1_cards_test.go — a batch of Theros cards (#5561). Every
// test plays the real catalog entry through the engine.

const (
	thsBorderlandMinotaurOracle    = "03a9ac3f-f6e4-4ade-9770-cf723434b7e8"
	thsBronzeSableOracle           = "f27e5e11-0ad0-448b-8760-75ed1b97e7d8"
	thsFelhideMinotaurOracle       = "10871545-1c44-4e31-a49a-5b16c3b5cac5"
	thsPheresBandCentaursOracle    = "4aebea34-a05b-4626-97bc-f15fa177ef05"
	thsSilentArtisanOracle         = "2298eb28-425d-4e87-990b-6bc0ef9bcfda"
	thsTravelingPhilosopherOracle  = "1adc4691-20cf-4f55-af85-3132fe1ff2a8"
	thsTritonShorethiefOracle      = "d8f6f61a-39c2-4bac-9f16-a5e6340f6301"
	thsYokedOxOracle               = "19b4cf0b-e1e4-4e50-80bc-673bc6ef8bb8"
	thsAnvilwroughtRaptorOracle    = "d87ebd8c-ad2c-45c9-8c97-f046aeda861c"
	thsBenthicGiantOracle          = "43c449d1-51a9-45e0-8102-823a6811ff54"
	thsGuardiansOfMeletisOracle    = "7604aac7-82ae-4f5b-888d-a008754f716a"
	thsInsatiableHarpyOracle       = "6774c0ce-103e-471b-8de0-f325805674ca"
	thsSatyrRamblerOracle          = "6ea0d274-564b-4785-801d-b3272a7b6bfa"
	thsVulpineGoliathOracle        = "e4c96980-31ed-494e-89db-c0a0f1fd23fd"
	thsSedgeScorpionOracle         = "fbadf7a3-0d54-40fe-a763-f641dd448e56"
	thsTwoHeadedCerberusOracle     = "7b2bbd87-f288-44bd-a15c-1432a3e90cb3"
	thsVaporkinOracle              = "169e415a-9ff6-43bb-bb68-fc49e69f95a8"
	thsHorizonScholarOracle        = "29933a32-3738-4ec5-aa6d-ce3684e2e5a3"
	thsOmenspeakerOracle           = "43d7fcf3-acc2-4e9b-a466-c73b1b6c58af"
	thsBloodTollHarpyOracle        = "98f9619e-585f-486f-b143-fb42085eb583"
	thsMinotaurSkullcleaverOracle  = "dd36aea8-9f1c-4daf-86f6-ab9307279ce7"
	thsReturnedCentaurOracle       = "223c97fb-dc96-41ba-be9d-bad9b5e4876f"
	thsAgentOfHorizonsOracle       = "668a6a9d-80fc-437b-99b8-6807e92eabca"
	thsAsphodelWandererOracle      = "a6ffcbc9-0990-4d95-a995-1b2f321256d7"
	thsDeathbellowRaiderOracle     = "3b14e7ea-e01c-421d-acd2-0b70436ca6aa"
	thsOpalineUnicornOracle        = "ca5ca2c9-2ddd-4e70-b738-8de61aa82570"
	thsUnknownShoresOracle         = "82c5ec8e-27be-475a-9921-ad61209fd022"
	thsVoyagingSatyrOracle         = "bce18f34-848e-4dd2-9a38-6084c0695f78"
	thsIllTemperedCyclopsOracle    = "07f62413-92cc-4f44-a827-a7c3a7318602"
	thsNessianAspOracle            = "8b46e077-f667-432f-ab5b-8256a7378311"
	thsKeepsakeGorgonOracle        = "3aa58e5a-1b5e-4fb3-bf2d-9bad8126b6e6"
	thsFabledHeroOracle            = "f9fd9b1c-4c00-4db7-bbfe-7fd5d8bb9258"
	thsCentaurBattlemasterOracle   = "4b854749-f3a5-4199-bbfa-83090b8ccc25"
	thsStaunchHeartedWarriorOracle = "aa6244fb-e1f1-4b53-b530-a8c9d0712a5b"
	thsTritonFortuneHunterOracle   = "f84637f0-1a04-4fd1-85d9-8e8f5d444800"
	thsNimbusNaiadOracle           = "1b3b14b9-088d-48df-8d21-54ee68989ac0"
	thsObservantAlseidOracle       = "0eb0c9d2-bc5a-4ba3-bd83-57af1da51c3b"
	thsLeafcrownDryadOracle        = "01914751-af8e-4aea-a254-6fa4f9fd3457"
	thsNyleasEmissaryOracle        = "ec30e221-ffc8-4867-ba9c-f39712b3aa93"
	thsSpearpointOreadOracle       = "2fa7bb26-18eb-4ff0-8558-f6f3981a84ce"
	thsFeralInvocationOracle       = "5b3c069b-5bed-4b70-b04c-b7143a4a6e96"
	thsScourgemarkOracle           = "66efb37f-9eab-4ff4-a1d4-65fe9ca7332b"
	thsFleetfeatherSandalsOracle   = "9811d1fa-1398-49a2-bcbe-0e40d9779ef1"
	thsPharikasCureOracle          = "06d77149-52b6-4d2c-bfb9-1c25f6b1b0bb"
	thsLashOfTheWhipOracle         = "e2ac369a-f6d4-4b58-8695-a7c5b906e999"
	thsMarchOfTheReturnedOracle    = "53a40369-0e62-4ad2-a6bd-65a5ff0da65e"
	thsSipOfHemlockOracle          = "861096be-dbb3-429e-820f-d82f6ed2aa24"
	thsDarkBetrayalOracle          = "d3ebe9ae-108f-4d23-a236-35d632889cf8"
	thsGlareOfHeresyOracle         = "6e127e6e-b292-4398-b3fa-8b520a7d26a3"
	thsShreddingWindsOracle        = "9f914186-082e-4a85-a1cc-e9788074637b"

	thsBearType  = "Creature — Bear"
	thsAuraType  = "Enchantment — Aura"
	thsNymphType = "Enchantment Creature — Nymph"
)

// --- every card is in the catalog ------------------------------------------

func TestThs5561B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		thsBorderlandMinotaurOracle, thsBronzeSableOracle, thsFelhideMinotaurOracle, thsPheresBandCentaursOracle,
		thsSilentArtisanOracle, thsTravelingPhilosopherOracle, thsTritonShorethiefOracle, thsYokedOxOracle,
		thsAnvilwroughtRaptorOracle, thsBenthicGiantOracle, thsGuardiansOfMeletisOracle, thsInsatiableHarpyOracle,
		thsSatyrRamblerOracle, thsVulpineGoliathOracle, thsSedgeScorpionOracle, thsTwoHeadedCerberusOracle,
		thsVaporkinOracle, thsHorizonScholarOracle, thsOmenspeakerOracle, thsBloodTollHarpyOracle,
		thsMinotaurSkullcleaverOracle, thsReturnedCentaurOracle, thsAgentOfHorizonsOracle, thsAsphodelWandererOracle,
		thsDeathbellowRaiderOracle, thsOpalineUnicornOracle, thsUnknownShoresOracle, thsVoyagingSatyrOracle,
		thsIllTemperedCyclopsOracle, thsNessianAspOracle, thsKeepsakeGorgonOracle, thsFabledHeroOracle,
		thsCentaurBattlemasterOracle, thsStaunchHeartedWarriorOracle, thsTritonFortuneHunterOracle, thsNimbusNaiadOracle,
		thsObservantAlseidOracle, thsLeafcrownDryadOracle, thsNyleasEmissaryOracle, thsSpearpointOreadOracle,
		thsFeralInvocationOracle, thsScourgemarkOracle, thsFleetfeatherSandalsOracle, thsPharikasCureOracle,
		thsLashOfTheWhipOracle, thsMarchOfTheReturnedOracle, thsSipOfHemlockOracle, thsDarkBetrayalOracle,
		thsGlareOfHeresyOracle, thsShreddingWindsOracle,
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

// --- printed keywords ------------------------------------------------------

func TestThs5561B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Anvilwrought Raptor", "Artifact Creature — Bird", thsAnvilwroughtRaptorOracle, []string{"flying", "first strike"}},
		{"Benthic Giant", "Creature — Giant", thsBenthicGiantOracle, []string{"hexproof"}},
		{"Guardians of Meletis", "Artifact Creature — Golem", thsGuardiansOfMeletisOracle, []string{"defender"}},
		{"Insatiable Harpy", "Creature — Harpy", thsInsatiableHarpyOracle, []string{"flying", "lifelink"}},
		{"Satyr Rambler", "Creature — Satyr", thsSatyrRamblerOracle, []string{"trample"}},
		{"Vulpine Goliath", "Creature — Fox", thsVulpineGoliathOracle, []string{"trample"}},
		{"Sedge Scorpion", "Creature — Scorpion", thsSedgeScorpionOracle, []string{"deathtouch"}},
		{"Two-Headed Cerberus", "Creature — Dog", thsTwoHeadedCerberusOracle, []string{"double strike"}},
		{"Vaporkin", "Creature — Elemental", thsVaporkinOracle, []string{"flying"}},
		{"Horizon Scholar", "Creature — Sphinx", thsHorizonScholarOracle, []string{"flying"}},
		{"Blood-Toll Harpy", "Creature — Harpy", thsBloodTollHarpyOracle, []string{"flying"}},
		{"Minotaur Skullcleaver", "Creature — Minotaur Berserker", thsMinotaurSkullcleaverOracle, []string{"haste"}},
		{"Ill-Tempered Cyclops", "Creature — Cyclops", thsIllTemperedCyclopsOracle, []string{"trample"}},
		{"Nessian Asp", "Creature — Snake", thsNessianAspOracle, []string{"reach"}},
		{"Keepsake Gorgon", "Creature — Gorgon", thsKeepsakeGorgonOracle, []string{"deathtouch"}},
		{"Fabled Hero", "Creature — Human Soldier", thsFabledHeroOracle, []string{"double strike"}},
		{"Nimbus Naiad", thsNymphType, thsNimbusNaiadOracle, []string{"flying"}},
		{"Observant Alseid", thsNymphType, thsObservantAlseidOracle, []string{"vigilance"}},
		{"Leafcrown Dryad", "Enchantment Creature — Nymph Dryad", thsLeafcrownDryadOracle, []string{"reach"}},
		{"Nylea's Emissary", "Enchantment Creature — Cat", thsNyleasEmissaryOracle, []string{"trample"}},
		{"Spearpoint Oread", thsNymphType, thsSpearpointOreadOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Vaporkin blocks a flyer and refuses an attacker without flying.
func TestThs5561B1VaporkinBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", thsBearType, 2, 2)
	flyer := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	vaporkin := effectstest.B12Push(g, opp.ID, "Vaporkin", "Creature — Elemental", thsVaporkinOracle, 2, 1)
	effectstest.BrAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(vaporkin, walker); err == nil {
		t.Error("Vaporkin blocked a creature without flying")
	}
	if err := g.DeclareBlocker(vaporkin, flyer); err != nil {
		t.Errorf("Vaporkin can't block a flyer: %v", err)
	}
}

// --- enters triggers ---------------------------------------------------------

func TestThs5561B1EntersAndScriesTwo(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Horizon Scholar", "Creature — Sphinx", thsHorizonScholarOracle, 4, 4},
		{"Omenspeaker", "Creature — Human Wizard", thsOmenspeakerOracle, 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			c := effectstest.ScryChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no scry prompt")
			}
			if len(c.ScryCards) != 2 {
				t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
			}
		})
	}
}

func TestThs5561B1BloodTollHarpyDrainsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.CastHoldCreature(t, g, "Blood-Toll Harpy", "Creature — Harpy", thsBloodTollHarpyOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]-1 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]-1)
		}
	}
}

func TestThs5561B1MinotaurSkullcleaverPumpsItselfOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Minotaur Skullcleaver", "Creature — Minotaur Berserker", thsMinotaurSkullcleaverOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 2 {
		t.Errorf("the Minotaur is %d/%d, want 4/2", p, tough)
	}
}

func TestThs5561B1ReturnedCentaurMillsTheTargetPlayerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	library := opp.Library.Size()
	effectstest.CastHoldCreature(t, g, "Returned Centaur", "Creature — Zombie Centaur", thsReturnedCentaurOracle, 2, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
}

// --- activated abilities -------------------------------------------------------

func TestThs5561B1AgentOfHorizonsBecomesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	agent := effectstest.B12Push(g, me.ID, "Agent of Horizons", "Creature — Human Rogue", thsAgentOfHorizonsOracle, 3, 2)
	effectstest.AssertRestrictions(t, g, agent, 0)
	if err := g.ActivateCatalogAbility(me.ID, agent, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.ApaMana(me, "C", "C", "U")
	effectstest.ApaActivate(t, g, me, agent, 0, game.ActivateAbilityParams{})
	effectstest.AssertRestrictions(t, g, agent, game.CantBeBlocked)
}

func TestThs5561B1RegeneratorsSurviveDestruction(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Asphodel Wanderer", "Creature — Skeleton Soldier", thsAsphodelWandererOracle},
		{"Deathbellow Raider", "Creature — Minotaur Berserker", thsDeathbellowRaiderOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.ApaMana(me, "C", "C", "B")
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			effectstest.B27Kill(g, id)
			if !g.Battlefield.Contains(id) {
				t.Errorf("the shielded %s was destroyed", tc.name)
			}
		})
	}
}

func TestThs5561B1DeathbellowRaiderMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	raider := effectstest.B12Push(g, me.ID, "Deathbellow Raider", "Creature — Minotaur Berserker", thsDeathbellowRaiderOracle, 2, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != raider {
		t.Fatalf("pass with the Raider home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Raider attacking: %v", err)
	}
}

func TestThs5561B1OpalineUnicornTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	unicorn := effectstest.PushCatalogPermanent(g, me.ID, "Opaline Unicorn", "Artifact Creature — Unicorn", thsOpalineUnicornOracle, false)
	for _, colour := range []string{"W", "G"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(unicorn) })
		me.ManaPool = nil
		effectstest.B28TapForMana(t, g, me.ID, unicorn, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("the Unicorn taps for {%s}: pool %v", colour, got)
		}
	}
}

func TestThs5561B1UnknownShoresTapsForColorlessOrFiltersAnyColor(t *testing.T) {
	g, me, _ := effectstest.C2LandTable(t)
	land := effectstest.C2Land(g, me.ID, "Unknown Shores", "Land", thsUnknownShoresOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("pool %d after tapping for {C}, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	me.ManaPool = nil
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the filter ability paid its {1} from nothing")
	}
	effectstest.C2Pay(t, g, me, "{C}")
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("filter: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want one {B}", got)
	}
}

func TestThs5561B1VoyagingSatyrUntapsOnlyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	satyr := effectstest.B12Push(g, me.ID, "Voyaging Satyr", "Creature — Satyr Druid", thsVoyagingSatyrOracle, 1, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", thsBearType, 2, 2)
	effectstest.FindBattlefieldCardForTest(g, land).Tapped = true
	effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true
	if err := g.ActivateCatalogAbility(me.ID, satyr, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("the Satyr targeted a creature")
	}
	effectstest.ApaActivate(t, g, me, satyr, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)})
	if effectstest.HmlTapped(t, g, land) {
		t.Error("the land is still tapped")
	}
	if !effectstest.HmlTapped(t, g, satyr) {
		t.Error("the Satyr did not tap")
	}
}

// --- monstrosity ----------------------------------------------------------------

func TestThs5561B1MonstrosityPutsItsCounters(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Ill-Tempered Cyclops", thsIllTemperedCyclopsOracle, 3},
		{"Nessian Asp", thsNessianAspOracle, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushMonster(g, me.ID, tc.name, tc.oracle, 3, 3)
			effectstest.ActivateMonstrosity(t, g, me.ID, id, 0)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.PlusOneCounters(g, id); got != tc.n {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.n)
			}
			if !effectstest.MonstrousNow(g, id) {
				t.Error("not monstrous")
			}
		})
	}
}

func TestThs5561B1KeepsakeGorgonDestroysANonGorgonOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorgon := effectstest.PushMonster(g, me.ID, "Keepsake Gorgon", thsKeepsakeGorgonOracle, 2, 5)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", thsBearType, 2, 2)
	theirGorgon := effectstest.B12Creature(g, opp.ID, "Their Gorgon", "Creature — Gorgon", 2, 2)
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", thsBearType, 2, 2)
	effectstest.ActivateMonstrosity(t, g, me.ID, gorgon, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, theirGorgon) || effectstest.HasID(p.PickTargetCards, myBear) {
		t.Errorf("illegal targets offered: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBear) {
		t.Error("the opponent's creature survived")
	}
	if got := effectstest.PlusOneCounters(g, gorgon); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// --- heroic -----------------------------------------------------------------------

func TestThs5561B1HeroicCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Fabled Hero", "Creature — Human Soldier", thsFabledHeroOracle, 1},
		{"Centaur Battlemaster", "Creature — Centaur Warrior", thsCentaurBattlemasterOracle, 3},
		{"Staunch-Hearted Warrior", "Creature — Human Warrior", thsStaunchHeartedWarriorOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			c := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			other := effectstest.B12Creature(g, me.ID, "Other", thsBearType, 2, 2)
			effectstest.BnsB25Heroic(t, g, other)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != 0 {
				t.Fatalf("a spell targeting another creature put %d counters on it", got)
			}
			effectstest.BnsB25Heroic(t, g, c)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.n)
			}
		})
	}
}

func TestThs5561B1TritonFortuneHunterDrawsWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Triton Fortune Hunter", "Creature — Merfolk Soldier", thsTritonFortuneHunterOracle, 2, 2)
	hand := me.Hand.Size()
	effectstest.BnsB25Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

// --- bestow -------------------------------------------------------------------------

func TestThs5561B1BestowedPumpsAndGrantsItsKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		bonus                           int
	}{
		{"Nimbus Naiad", thsNymphType, thsNimbusNaiadOracle, "flying", 2},
		{"Observant Alseid", thsNymphType, thsObservantAlseidOracle, "vigilance", 2},
		{"Leafcrown Dryad", "Enchantment Creature — Nymph Dryad", thsLeafcrownDryadOracle, "reach", 2},
		{"Nylea's Emissary", "Enchantment Creature — Cat", thsNyleasEmissaryOracle, "trample", 3},
		{"Spearpoint Oread", thsNymphType, thsSpearpointOreadOracle, "first strike", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.BestowCard(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", thsBearType, 2, 2)
			effectstest.CastBestowed(t, g, id, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if s := effectstest.ReadBestowState(g, id); !s.OnBattlefield || s.AttachedTo != bear {
				t.Fatalf("the bestowed Aura is not attached to the Bear: %+v", s)
			}
			if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2+tc.bonus || tough != 2+tc.bonus {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, 2+tc.bonus, 2+tc.bonus)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

// --- Auras and Equipment ----------------------------------------------------------

func TestThs5561B1FeralInvocationHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", thsBearType, 2, 2)
	id := effectstest.B13HandCard(me, "Feral Invocation", thsAuraType, thsFeralInvocationOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("Feral Invocation at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestThs5561B1ScourgemarkPumpsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", thsBearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Scourgemark", thsAuraType, thsScourgemarkOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
}

func TestThs5561B1FleetfeatherSandalsGrantFlyingAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	sandals := effectstest.PushCatalogPermanent(g, me.ID, "Fleetfeather Sandals", "Artifact — Equipment", thsFleetfeatherSandalsOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", thsBearType, 2, 2)
	effectstest.EquipTo(t, g, me.ID, sandals, bear)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped Bear has no %s", kw)
		}
	}
}

// --- the spells ---------------------------------------------------------------------

func TestThs5561B1PharikasCureDamagesAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Pharika's Cure", "Instant", thsPharikasCureOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the Ogre has %d damage, want 2", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestThs5561B1LashOfTheWhipShrinksFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Lash of the Whip", "Instant", thsLashOfTheWhipOracle, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, giant); p != 1 || tough != 1 {
		t.Errorf("the Giant is %d/%d, want 1/1", p, tough)
	}
}

func TestThs5561B1MarchOfTheReturnedReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	one := effectstest.WthGraveyardCard(me, "One", thsBearType)
	two := effectstest.WthGraveyardCard(me, "Two", thsBearType)
	rock := effectstest.WthGraveyardCard(me, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "March of the Returned", "Sorcery", thsMarchOfTheReturnedOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("March of the Returned targeted an artifact card")
	}
	effectstest.CastCatalogSpell(t, g, "March of the Returned", "Sorcery", thsMarchOfTheReturnedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
			t.Error("a creature card did not return to hand")
		}
	}
}

func TestThs5561B1SipOfHemlockDestroysAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", thsBearType, 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Sip of Hemlock", "Sorcery", thsSipOfHemlockOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestThs5561B1DarkBetrayalDestroysOnlyABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.HmlSetColors(g, elf, "G")
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, zombie, "B")
	if err := effectstest.CastCatalogSpellErr(t, g, "Dark Betrayal", "Instant", thsDarkBetrayalOracle, effectstest.CardTarget(elf)); err == nil {
		t.Fatal("Dark Betrayal targeted a green creature")
	}
	effectstest.CastCatalogSpell(t, g, "Dark Betrayal", "Instant", thsDarkBetrayalOracle, effectstest.CardTarget(zombie))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zombie) {
		t.Error("the black creature survived")
	}
}

func TestThs5561B1GlareOfHeresyExilesOnlyAWhitePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	red := effectstest.B12Permanent(g, opp.ID, "Red Rune", "Enchantment")
	effectstest.HmlSetColors(g, red, "R")
	white := effectstest.B12Permanent(g, opp.ID, "White Rune", "Enchantment")
	effectstest.HmlSetColors(g, white, "W")
	if err := effectstest.CastCatalogSpellErr(t, g, "Glare of Heresy", "Sorcery", thsGlareOfHeresyOracle, effectstest.CardTarget(red)); err == nil {
		t.Fatal("Glare of Heresy targeted a red permanent")
	}
	effectstest.CastCatalogSpell(t, g, "Glare of Heresy", "Sorcery", thsGlareOfHeresyOracle, effectstest.CardTarget(white))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || opp.Graveyard.Contains(white) {
		t.Error("the white permanent was not exiled")
	}
}

func TestThs5561B1ShreddingWindsHitsOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", thsBearType, 8, 8)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 8, Toughness: 8, Keywords: []string{"flying"}})
	if err := effectstest.CastCatalogSpellErr(t, g, "Shredding Winds", "Instant", thsShreddingWindsOracle, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("Shredding Winds targeted a creature without flying")
	}
	effectstest.CastCatalogSpell(t, g, "Shredding Winds", "Instant", thsShreddingWindsOracle, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 7 {
		t.Errorf("the flyer has %d damage, want 7", got)
	}
}
