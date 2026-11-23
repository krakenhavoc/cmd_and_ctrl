package testsds

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dtk_5551_b1_cards_test.go — Dragons of Tarkir, batch 1 (#5551). Every
// test plays the real catalog entry through the engine.

const (
	dtkAncientCarpOracle         = "84122d1e-7657-42fa-8724-c4a3a19729a6"
	dtkColossodonYearlingOracle  = "a322ad80-e64b-4e9a-b542-3a036dc14d28"
	dtkDromokaWarriorOracle      = "f62998a7-898f-494f-89a1-624f6236033b"
	dtkWanderingTombshellOracle  = "768c78f3-f09c-4b77-974d-514925e8df5e"
	dtkChampionOfArashinOracle   = "f688863f-b8eb-4143-8ea5-0b234cc91bd0"
	dtkConiferStriderOracle      = "6175197f-cf3c-41df-9d24-b4fffa09631a"
	dtkHandOfSilumgarOracle      = "e2714e1f-c883-41e0-bc2f-60e15462ef98"
	dtkScionOfUginOracle         = "7422f43d-0c1e-42cb-864d-1c63f2ec533a"
	dtkTerritorialRocOracle      = "adb5883d-44cc-455e-94dc-20e1c1f97d68"
	dtkUpdraftElementalOracle    = "a33982e4-d7d2-4319-bf04-534b8f253dfb"
	dtkDragonsEyeSentryOracle    = "661b23fc-78cd-44f7-ba53-0cb89ccca41f"
	dtkAvenSunstrikerOracle      = "60b0fd7c-a4cc-4a41-8455-debcfeccff81"
	dtkDirgurNemesisOracle       = "c08b444b-6617-4199-b9da-5d4e81cd7597"
	dtkMarshHulkOracle           = "979a581b-22a0-4461-8638-d219587323d7"
	dtkMisthoofKirinOracle       = "678eee7a-549d-4f40-a81c-70d036fc299d"
	dtkOjutaiInterceptorOracle   = "ecc42076-5cce-4e00-a21b-34846ea41df2"
	dtkSandstormChargerOracle    = "aa7a9689-e79a-4562-a621-eadc45966ce8"
	dtkSegmentedKrotiqOracle     = "05221ba6-53fd-4f45-b8b4-493724f75963"
	dtkStormcragElementalOracle  = "f28c18a1-13f8-4360-adc7-9d828ab1e933"
	dtkGudulLurkerOracle         = "bde976c6-afed-44bb-8e3d-4103289dff5e"
	dtkMarangRiverSkeletonOracle = "0cfbbb26-9e1c-4b79-8b32-dcc05739f5ea"
	dtkAcidSpewerDragonOracle    = "f4378682-359f-48b0-adb7-b996b8d972db"
	dtkBelltollDragonOracle      = "d0e529e3-0f77-4a55-ba35-46591f48d960"
	dtkHerdchaserDragonOracle    = "9c7073e9-c676-4629-814c-a96e7af905cf"
	dtkShieldhideDragonOracle    = "4f5fee21-bf29-4873-baf4-04911c309d81"
	dtkStormwingDragonOracle     = "10796fe6-edf8-427b-986c-e7666a72ec8d"
	dtkAtarkaEfreetOracle        = "177c1f2d-abb9-40ea-afe4-141b08fdeabb"
	dtkStratusDancerOracle       = "5b6a821c-51e4-44dc-8353-f4d10eab7fcb"
	dtkSilumgarSpellEaterOracle  = "34d90f98-7af5-4c67-8fb3-620237d94d14"
	dtkKolaghanSkirmisherOracle  = "ade4ffa5-6609-44eb-8f1a-45e7203d90e3"
	dtkScreamreachBrawlerOracle  = "ffc84ff4-22eb-4652-9723-2f40d637a427"
	dtkLightningBerserkerOracle  = "1e474772-3fe4-46be-8c5f-6e477b9fd74e"
	dtkRecklessImpOracle         = "e8e88e9a-0097-49ac-8b23-482a3190f21b"
	dtkPitilessHordeOracle       = "d5726b60-60f6-4481-bc76-759fbb55bb49"
	dtkSilumgarButcherOracle     = "77216efa-4f34-4534-9df3-918f0b888ce1"
	dtkQarsiSadistOracle         = "69e15fd8-9531-4392-8b85-b539c138ea91"
	dtkVulturousAvenOracle       = "83c0c648-f000-470d-bcc3-9ec81257d0cd"
	dtkAvenTacticianOracle       = "f2d5949b-3cdc-4efd-9841-7d450e12dbbe"
	dtkSandcrafterMageOracle     = "6bf9c91f-f278-4b97-8517-d062118c4e56"
	dtkSandsteppeScavengerOracle = "747ee321-f2dc-4d27-90cd-c4e610086958"
	dtkDromokasGiftOracle        = "1264da2f-4767-4bec-b4ed-1c625ac3cf95"
	dtkDefeatOracle              = "21645448-ecbd-4036-b9ed-47cd903079e4"
	dtkFateForgottenOracle       = "7f591bfb-a241-4980-9cb3-fb5af3355f6b"
	dtkFlattenOracle             = "bc5d5084-77de-4dc5-8769-300eafcbfee7"
	dtkResupplyOracle            = "129ec223-e76d-491d-b299-07ac91168d31"
	dtkRoastOracle               = "b5e60a85-2f31-47ec-a95b-2e69b2d2c00e"
	dtkGlintOracle               = "2fc07680-3ffb-4507-a1e1-b9580f5ce878"
	dtkPalaceFamiliarOracle      = "dab9e40f-b73a-4c55-8bb3-63b32c23f07b"
	dtkCunningBreezedancerOracle = "ba997e44-fc80-4463-ad3f-6497be26df0f"

	dtkBearType   = "Creature — Bear"
	dtkDragonType = "Creature — Dragon"
)

// dtkManaFor turns a printed cost such as "{4}{W}" into pool tokens: a
// colourless token per generic mana and one per coloured symbol.
func dtkManaFor(cost string) []string {
	var out []string
	for _, sym := range strings.Split(strings.Trim(cost, "{}"), "}{") {
		n := 0
		generic := sym != ""
		for _, r := range sym {
			if r < '0' || r > '9' {
				generic = false
				break
			}
			n = n*10 + int(r-'0')
		}
		if generic {
			for i := 0; i < n; i++ {
				out = append(out, "C")
			}
			continue
		}
		out = append(out, sym)
	}
	return out
}

// dtkResupplyOnStack casts Resupply (a catalog instant from this batch)
// from the active player's hand and leaves it on the stack: a spell whose
// six life shows whether it resolved.
func dtkResupplyOnStack(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, "Resupply", "Instant", dtkResupplyOracle, nil)
}

// --- every card is in the catalog ----------------------------------------

func TestDtk5551B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		dtkAncientCarpOracle, dtkColossodonYearlingOracle, dtkDromokaWarriorOracle, dtkWanderingTombshellOracle,
		dtkChampionOfArashinOracle, dtkConiferStriderOracle, dtkHandOfSilumgarOracle, dtkScionOfUginOracle,
		dtkTerritorialRocOracle, dtkUpdraftElementalOracle, dtkDragonsEyeSentryOracle, dtkAvenSunstrikerOracle,
		dtkDirgurNemesisOracle, dtkMarshHulkOracle, dtkMisthoofKirinOracle, dtkOjutaiInterceptorOracle,
		dtkSandstormChargerOracle, dtkSegmentedKrotiqOracle, dtkStormcragElementalOracle, dtkGudulLurkerOracle,
		dtkMarangRiverSkeletonOracle, dtkAcidSpewerDragonOracle, dtkBelltollDragonOracle, dtkHerdchaserDragonOracle,
		dtkShieldhideDragonOracle, dtkStormwingDragonOracle, dtkAtarkaEfreetOracle, dtkStratusDancerOracle,
		dtkSilumgarSpellEaterOracle, dtkKolaghanSkirmisherOracle, dtkScreamreachBrawlerOracle, dtkLightningBerserkerOracle,
		dtkRecklessImpOracle, dtkPitilessHordeOracle, dtkSilumgarButcherOracle, dtkQarsiSadistOracle,
		dtkVulturousAvenOracle, dtkAvenTacticianOracle, dtkSandcrafterMageOracle, dtkSandsteppeScavengerOracle,
		dtkDromokasGiftOracle, dtkDefeatOracle, dtkFateForgottenOracle, dtkFlattenOracle,
		dtkResupplyOracle, dtkRoastOracle, dtkGlintOracle, dtkPalaceFamiliarOracle,
		dtkCunningBreezedancerOracle,
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

// --- printed keywords -------------------------------------------------------

func TestDtk5551B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Champion of Arashin", "Creature — Dog Warrior", dtkChampionOfArashinOracle, []string{"lifelink"}},
		{"Conifer Strider", "Creature — Elemental", dtkConiferStriderOracle, []string{"hexproof"}},
		{"Hand of Silumgar", "Creature — Human Warrior", dtkHandOfSilumgarOracle, []string{"deathtouch"}},
		{"Scion of Ugin", "Creature — Dragon Spirit", dtkScionOfUginOracle, []string{"flying"}},
		{"Territorial Roc", "Creature — Bird", dtkTerritorialRocOracle, []string{"flying"}},
		{"Updraft Elemental", "Creature — Elemental", dtkUpdraftElementalOracle, []string{"flying"}},
		{"Dragon's Eye Sentry", "Creature — Human Monk", dtkDragonsEyeSentryOracle, []string{"defender", "first strike"}},
		{"Aven Sunstriker", "Creature — Bird Warrior", dtkAvenSunstrikerOracle, []string{"flying", "double strike"}},
		{"Dirgur Nemesis", "Creature — Serpent", dtkDirgurNemesisOracle, []string{"defender"}},
		{"Misthoof Kirin", "Creature — Kirin", dtkMisthoofKirinOracle, []string{"flying", "vigilance"}},
		{"Ojutai Interceptor", "Creature — Bird Soldier", dtkOjutaiInterceptorOracle, []string{"flying"}},
		{"Stormcrag Elemental", "Creature — Elemental", dtkStormcragElementalOracle, []string{"trample"}},
		{"Acid-Spewer Dragon", dtkDragonType, dtkAcidSpewerDragonOracle, []string{"flying", "deathtouch"}},
		{"Belltoll Dragon", dtkDragonType, dtkBelltollDragonOracle, []string{"flying", "hexproof"}},
		{"Herdchaser Dragon", dtkDragonType, dtkHerdchaserDragonOracle, []string{"flying", "trample"}},
		{"Shieldhide Dragon", dtkDragonType, dtkShieldhideDragonOracle, []string{"flying", "lifelink"}},
		{"Stormwing Dragon", dtkDragonType, dtkStormwingDragonOracle, []string{"flying", "first strike"}},
		{"Stratus Dancer", "Creature — Djinn Monk", dtkStratusDancerOracle, []string{"flying"}},
		{"Reckless Imp", "Creature — Imp", dtkRecklessImpOracle, []string{"flying"}},
		{"Vulturous Aven", "Creature — Bird Shaman", dtkVulturousAvenOracle, []string{"flying"}},
		{"Aven Tactician", "Creature — Bird Soldier", dtkAvenTacticianOracle, []string{"flying"}},
		{"Palace Familiar", "Creature — Bird", dtkPalaceFamiliarOracle, []string{"flying"}},
		{"Cunning Breezedancer", dtkDragonType, dtkCunningBreezedancerOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- megamorph ----------------------------------------------------------------

// Every megamorph card casts face down for {3} and turns face up for its
// megamorph cost with a +1/+1 counter (CR 702.37b).
func TestDtk5551B1MegamorphTurnsUpWithACounter(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
	}{
		{"Aven Sunstriker", "Creature — Bird Warrior", dtkAvenSunstrikerOracle, "{4}{W}"},
		{"Dirgur Nemesis", "Creature — Serpent", dtkDirgurNemesisOracle, "{6}{U}"},
		{"Marsh Hulk", "Creature — Zombie Ogre", dtkMarshHulkOracle, "{6}{B}"},
		{"Misthoof Kirin", "Creature — Kirin", dtkMisthoofKirinOracle, "{1}{W}"},
		{"Ojutai Interceptor", "Creature — Bird Soldier", dtkOjutaiInterceptorOracle, "{3}{U}"},
		{"Sandstorm Charger", "Creature — Beast", dtkSandstormChargerOracle, "{4}{W}"},
		{"Segmented Krotiq", "Creature — Insect", dtkSegmentedKrotiqOracle, "{6}{G}"},
		{"Stormcrag Elemental", "Creature — Elemental", dtkStormcragElementalOracle, "{4}{R}{R}"},
		{"Gudul Lurker", "Creature — Salamander", dtkGudulLurkerOracle, "{U}"},
		{"Marang River Skeleton", "Creature — Skeleton", dtkMarangRiverSkeletonOracle, "{3}{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alt := game.FaceDownCastFor(tc.oracle)
			if alt == nil || alt.FaceDown == nil || !alt.FaceDown.FaceUpCounter {
				t.Fatalf("%s has no megamorph declaration", tc.name)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.S36CastFaceDown(t, g, tc.name, tc.typeLine, tc.oracle, "megamorph", 3, 3)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.S36TurnUp(t, g, me, id, dtkManaFor(tc.cost)...)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 1 {
				t.Errorf("megamorph counters = %d, want 1", n)
			}
		})
	}
}

// The five Dragons put a +1/+1 counter on each OTHER Dragon creature you
// control as they turn face up, and on nothing else.
func TestDtk5551B1DragonsCounterEachOtherDragon(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Acid-Spewer Dragon", dtkAcidSpewerDragonOracle, "{5}{B}{B}"},
		{"Belltoll Dragon", dtkBelltollDragonOracle, "{5}{U}{U}"},
		{"Herdchaser Dragon", dtkHerdchaserDragonOracle, "{5}{G}{G}"},
		{"Shieldhide Dragon", dtkShieldhideDragonOracle, "{5}{W}{W}"},
		{"Stormwing Dragon", dtkStormwingDragonOracle, "{5}{R}{R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.S36CastFaceDown(t, g, tc.name, dtkDragonType, tc.oracle, "megamorph", 3, 3)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			whelp := effectstest.B12Creature(g, me.ID, "Whelp", dtkDragonType, 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", dtkBearType, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Dragon", dtkDragonType, 2, 2)
			effectstest.S36TurnUp(t, g, me, id, dtkManaFor(tc.cost)...)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, whelp, game.CounterPlusOne); n != 1 {
				t.Errorf("the other Dragon has %d counters, want 1", n)
			}
			if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 0 {
				t.Errorf("a non-Dragon got %d counters", n)
			}
			if n := effectstest.CounterCount(g, theirs, game.CounterPlusOne); n != 0 {
				t.Errorf("an opponent's Dragon got %d counters", n)
			}
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 1 {
				t.Errorf("the Dragon itself has %d counters, want only megamorph's 1", n)
			}
		})
	}
}

func TestDtk5551B1AtarkaEfreetPingsWhenTurnedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.S36CastFaceDown(t, g, "Atarka Efreet", "Creature — Efreet Shaman", dtkAtarkaEfreetOracle, "megamorph", 5, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	effectstest.S36TurnUp(t, g, me, id, "R", "C", "C")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestDtk5551B1StratusDancerCountersAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.S36CastFaceDown(t, g, "Stratus Dancer", "Creature — Djinn Monk", dtkStratusDancerOracle, "megamorph", 2, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	spell := dtkResupplyOnStack(t, g)
	effectstest.S36TurnUp(t, g, me, id, "U", "C")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the countered instant still gained life: %d, want %d", me.Life, life)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the countered instant is not in its owner's graveyard")
	}
}

func TestDtk5551B1SilumgarSpellEaterCountersUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.S36CastFaceDown(t, g, "Silumgar Spell-Eater", "Creature — Snake Wizard", dtkSilumgarSpellEaterOracle, "megamorph", 2, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	spell := dtkResupplyOnStack(t, g)
	effectstest.S36TurnUp(t, g, me, id, "U", "C", "C", "C", "C")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, spell)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the spell's controller was not asked to pay {3}")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the countered instant still gained life: %d, want %d", me.Life, life)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the countered instant is not in its owner's graveyard")
	}
}

// Gudul Lurker can't be blocked.
func TestDtk5551B1GudulLurkerCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lurker := effectstest.B12Push(g, me.ID, "Gudul Lurker", "Creature — Salamander", dtkGudulLurkerOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", dtkBearType, 2, 2)
	effectstest.BrAttack(t, g, lurker)
	if err := g.DeclareBlocker(bear, lurker); err == nil {
		t.Error("Gudul Lurker was blocked")
	}
}

func TestDtk5551B1MarangRiverSkeletonRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	skeleton := effectstest.B12Push(g, me.ID, "Marang River Skeleton", "Creature — Skeleton", dtkMarangRiverSkeletonOracle, 1, 1)
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, skeleton, 0, game.ActivateAbilityParams{})
	effectstest.B27Kill(g, skeleton)
	if !g.Battlefield.Contains(skeleton) {
		t.Error("the shielded Skeleton was destroyed")
	}
}

// --- dash -------------------------------------------------------------------------

// Each dash card cast for its dash cost has haste and returns to its
// owner's hand at the beginning of the next end step.
func TestDtk5551B1DashHasteThenReturnsToHand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, pool string
		power, toughness             int
	}{
		{"Kolaghan Skirmisher", "Creature — Human Warrior", dtkKolaghanSkirmisherOracle, "CCB", 2, 2},
		{"Screamreach Brawler", "Creature — Orc Berserker", dtkScreamreachBrawlerOracle, "CR", 2, 3},
		{"Lightning Berserker", "Creature — Human Berserker", dtkLightningBerserkerOracle, "R", 1, 1},
		{"Reckless Imp", "Creature — Imp", dtkRecklessImpOracle, "CB", 2, 2},
		{"Pitiless Horde", "Creature — Orc Berserker", dtkPitilessHordeOracle, "CCBB", 5, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.ToMainForCost(t, g)
			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
				Power: tc.power, Toughness: tc.toughness, Owner: me.ID, Controller: me.ID,
			})
			effectstest.FloatForTest(g, me, tc.pool)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
				t.Fatalf("dash %s: %v", tc.name, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("dashed %s is not on the battlefield", tc.name)
			}
			if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
				t.Error("a dashed creature has no haste")
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
				t.Error("the dashed creature did not return to its owner's hand at the end step")
			}
		})
	}
}

func TestDtk5551B1LightningBerserkerPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Lightning Berserker", "Creature — Human Berserker", dtkLightningBerserkerOracle, 1, 1)
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 1 {
		t.Errorf("the Berserker is %d/%d, want 2/1", p, tough)
	}
}

func TestDtk5551B1RecklessImpCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", dtkBearType, 2, 2)
	imp := effectstest.B12Push(g, opp.ID, "Reckless Imp", "Creature — Imp", dtkRecklessImpOracle, 2, 2)
	effectstest.BrAttack(t, g, walker)
	if err := g.DeclareBlocker(imp, walker); err == nil {
		t.Error("Reckless Imp blocked")
	}
}

func TestDtk5551B1PitilessHordeCostsTwoLifeEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Pitiless Horde", "Creature — Orc Berserker", dtkPitilessHordeOracle, 5, 3)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- exploit --------------------------------------------------------------------

func TestDtk5551B1SilumgarButcherShrinksTheTarget(t *testing.T) {
	g, me, opp, fodder := effectstest.VowExploitTable(t)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", dtkBearType, 3, 3)
	effectstest.ExploitCast(t, g, me, "Silumgar Butcher", "Creature — Zombie Djinn", dtkSilumgarButcherOracle)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived -3/-3")
	}
}

func TestDtk5551B1QarsiSadistDrainsTwo(t *testing.T) {
	g, me, opp, fodder := effectstest.VowExploitTable(t)
	effectstest.ExploitCast(t, g, me, "Qarsi Sadist", "Creature — Human Cleric", dtkQarsiSadistOracle)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	mine, theirs := me.Life, opp.Life
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("life: me %d (want %d), opponent %d (want %d)", me.Life, mine+2, opp.Life, theirs-2)
	}
}

func TestDtk5551B1VulturousAvenDrawsTwoAndLosesTwo(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	effectstest.PushLibraryTopForTest(me, "One")
	effectstest.PushLibraryTopForTest(me, "Two")
	effectstest.ExploitCast(t, g, me, "Vulturous Aven", "Creature — Bird Shaman", dtkVulturousAvenOracle)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestDtk5551B1DecliningTheExploitRaisesNoPayoff(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Silumgar Butcher", "Creature — Zombie Djinn", dtkSilumgarButcherOracle},
		{"Qarsi Sadist", "Creature — Human Cleric", dtkQarsiSadistOracle},
		{"Vulturous Aven", "Creature — Bird Shaman", dtkVulturousAvenOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, _ := effectstest.VowExploitTable(t)
			effectstest.VowDeclined(t, g, me, tc.name, tc.typeLine, tc.oracle)
		})
	}
}

// --- bolster --------------------------------------------------------------------

// The creature with the least toughness gets the counters.
func TestDtk5551B1EntersAndBolsters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, n    int
	}{
		{"Aven Tactician", "Creature — Bird Soldier", dtkAvenTacticianOracle, 2, 3, 1},
		{"Sandcrafter Mage", "Creature — Human Wizard", dtkSandcrafterMageOracle, 2, 2, 1},
		{"Sandsteppe Scavenger", "Creature — Dog Scout", dtkSandsteppeScavengerOracle, 2, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			small := effectstest.B12Creature(g, me.ID, "Small", dtkBearType, 1, 1)
			big := effectstest.B12Creature(g, me.ID, "Big", dtkBearType, 5, 5)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, small, game.CounterPlusOne); got != tc.n {
				t.Errorf("the least-toughness creature has %d counters, want %d", got, tc.n)
			}
			if effectstest.CounterCount(g, big, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, id, game.CounterPlusOne) != 0 {
				t.Error("a creature that is not the least tough got a counter")
			}
		})
	}
}

func TestDtk5551B1DromokasGiftBolstersFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.B12Creature(g, me.ID, "Small", dtkBearType, 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", dtkBearType, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Dromoka's Gift", "Instant", dtkDromokasGiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, small, game.CounterPlusOne); got != 4 {
		t.Errorf("the least-toughness creature has %d counters, want 4", got)
	}
	if got := effectstest.CounterCount(g, big, game.CounterPlusOne); got != 0 {
		t.Errorf("the tougher creature has %d counters, want 0", got)
	}
}

// --- the spells -------------------------------------------------------------------

func TestDtk5551B1DefeatDestroysPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", dtkBearType, 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", dtkBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Defeat", "Sorcery", dtkDefeatOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("Defeat targeted a creature with power 3")
	}
	effectstest.CastCatalogSpell(t, g, "Defeat", "Sorcery", dtkDefeatOracle, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the power-2 creature survived")
	}
}

func TestDtk5551B1FateForgottenExilesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", dtkBearType, 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := effectstest.CastCatalogSpellErr(t, g, "Fate Forgotten", "Instant", dtkFateForgottenOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Fate Forgotten targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Fate Forgotten", "Instant", dtkFateForgottenOracle, effectstest.CardTarget(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || opp.Graveyard.Contains(ench) {
		t.Error("the enchantment was not exiled")
	}
}

func TestDtk5551B1FlattenGivesMinusFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", dtkBearType, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Flatten", "Instant", dtkFlattenOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, tough)
	}
}

func TestDtk5551B1ResupplyGainsSixAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushLibraryTopForTest(me, "Top")
	life := me.Life
	dtkResupplyOnStack(t, g)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d", me.Life, life+6)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestDtk5551B1RoastHitsOnlyAGroundCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", dtkBearType, 6, 6)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 6, Toughness: 6, Keywords: []string{"flying"}})
	if err := effectstest.CastCatalogSpellErr(t, g, "Roast", "Sorcery", dtkRoastOracle, effectstest.CardTarget(flyer)); err == nil {
		t.Fatal("Roast targeted a creature with flying")
	}
	effectstest.CastCatalogSpell(t, g, "Roast", "Sorcery", dtkRoastOracle, effectstest.CardTarget(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, walker).DamageMarked; got != 5 {
		t.Errorf("the creature has %d damage, want 5", got)
	}
}

func TestDtk5551B1GlintProtectsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", dtkBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", dtkBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Glint", "Instant", dtkGlintOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Glint targeted a creature you don't control")
	}
	effectstest.CastCatalogSpell(t, g, "Glint", "Instant", dtkGlintOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 5 {
		t.Errorf("the creature is %d/%d, want 2/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the creature has no hexproof")
	}
}

// --- the rest ------------------------------------------------------------------------

func TestDtk5551B1PalaceFamiliarDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryTopForTest(me, "Top")
	familiar := effectstest.B12Push(g, me.ID, "Palace Familiar", "Creature — Bird", dtkPalaceFamiliarOracle, 1, 1)
	hand := me.Hand.Size()
	effectstest.B27Kill(g, familiar)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestDtk5551B1CunningBreezedancerGrowsOnNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dancer := effectstest.B12Push(g, me.ID, "Cunning Breezedancer", dtkDragonType, dtkCunningBreezedancerOracle, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Bear", dtkBearType, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, dancer); p != 4 || tough != 4 {
		t.Fatalf("a creature spell pumped it to %d/%d", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, dancer); p != 6 || tough != 6 {
		t.Errorf("after a noncreature spell %d/%d, want 6/6", p, tough)
	}
}

// The vanilla bodies are known to the catalog and carry no abilities of
// their own.
func TestDtk5551B1VanillaBodies(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Ancient Carp", "Creature — Fish", dtkAncientCarpOracle},
		{"Colossodon Yearling", "Creature — Beast", dtkColossodonYearlingOracle},
		{"Dromoka Warrior", "Creature — Human Warrior", dtkDromokaWarriorOracle},
		{"Wandering Tombshell", "Creature — Zombie Turtle", dtkWanderingTombshellOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if got := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})); got != 0 {
				t.Errorf("%s has %d activated abilities", tc.name, got)
			}
		})
	}
}
