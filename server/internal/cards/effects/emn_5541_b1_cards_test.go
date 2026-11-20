package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emn_5541_b1_cards_test.go — Eldritch Moon, batch 1 (#5541). Every test
// plays the real catalog entry through the engine.

const (
	emnB1DawnGryffOracle         = "1c34b2e4-e600-4688-86cb-8a1c11fdc770"
	emnB1DerangedWhelpOracle     = "6f246109-8c8b-4909-9888-561bc46ad776"
	emnB1FaithbearerOracle       = "725dc451-cae4-4345-9741-6c58e5a1a1dc"
	emnB1FalkenrathReaverOracle  = "6cfd9d88-f450-4408-bbe2-94a2d295b24a"
	emnB1FieldCreeperOracle      = "e6584e39-099c-437f-9239-fa718cac85af"
	emnB1SwiftSpinnerOracle      = "91585b64-f10d-47c6-a563-384ceb944710"
	emnB1WoodlandPatrolOracle    = "ce31cfc4-a109-4cf5-b8f9-210267a98114"
	emnB1VampireCutthroatOracle  = "2e163679-7303-4e05-9d20-86b76dfa1407"
	emnB1WeirdedVampireOracle    = "3af07cd3-36c3-4792-bc93-3ec82e1606bb"
	emnB1BloodbriarOracle        = "7aef2a40-ae18-46d6-9873-0e5056345b81"
	emnB1BrazenWolvesOracle      = "0b5a3d7b-6b2c-4e91-8218-2b957277855c"
	emnB1SteadfastCatharOracle   = "00895741-f94f-4d80-8ca4-c436bf3c8d56"
	emnB1CultistsStaffOracle     = "3e211792-2656-4bee-80ac-60eac0c45c2d"
	emnB1ExultantCultistOracle   = "e802442e-6a78-4c15-9034-fc8872470208"
	emnB1GavonyUnhallowedOracle  = "95ccb4a6-6e45-4c00-b7c6-25384022c02f"
	emnB1IronwrightsOracle       = "599faf5b-88fd-4373-b6ce-6047a67b30ea"
	emnB1SigardianPriestOracle   = "f2e22c9d-b0ec-42bb-a520-3c36df546ba0"
	emnB1SomberwaldStagOracle    = "b1b67851-23d3-4799-ab29-15bce7bb2c39"
	emnB1SpringsageRitualOracle  = "14945b80-10ec-4205-98df-cc3e119ca57c"
	emnB1TatteredHaunterOracle   = "754a72fd-e837-44f7-9bd7-5bf55ee8134c"
	emnB1AlchemistsOracle        = "9aaf0df7-5e2d-4f49-a8cb-66523be15ad6"
	emnB1DistemperOracle         = "4d3e7cde-a4ed-428f-bd37-c6dda2104221"
	emnB1ChillingGraspOracle     = "45c517d4-944c-43f6-8ead-a47965377575"
	emnB1InsatiableGorgersOracle = "e4534f33-8784-40a5-af1a-0f9975741aa2"
	emnB1BackwoodsOracle         = "d2aecca6-fe03-4686-9daa-68f1f6c73ab4"
	emnB1FoulbloodsOracle        = "08adacab-96c3-4532-8ab3-0c8064fa4150"
	emnB1BoldImpalerOracle       = "374e6c1b-5a9a-4d3f-8bff-2f6bc0419391"
	emnB1ChokingOracle           = "42bc33dd-3bba-462d-b272-2e8bd58c649e"
	emnB1DisplaceOracle          = "334e2147-5944-4558-b2bc-2f2bfb96dfa5"
	emnB1FiendBinderOracle       = "060c7999-ab81-4e37-af65-7cd219b693c8"
	emnB1StensiaInnkeeperOracle  = "8a7f446b-557a-41cc-b9ef-369e220a0eeb"
	emnB1IngeniousSkaabOracle    = "9107ece7-a263-4203-ab09-069193cc4902"
	emnB1MercurialGeistsOracle   = "1f6bff6c-a2f0-4ed7-8169-acb4d31a7187"
	emnB1LaboratoryBruteOracle   = "dc92bbee-a445-4927-932b-15985bb93f19"
	emnB1LilianasEliteOracle     = "b6492082-d0ef-4f7c-a4b1-f8ecf6eb0890"
	emnB1SpectralReservesOracle  = "628e5c1d-c7f9-4a98-98fb-13d7d5f81f64"
	emnB1CertainDeathOracle      = "1d4b8a31-6500-4e3d-bcaf-f6f184d7743e"
	emnB1ContingencyPlanOracle   = "67bc4116-539e-4748-95e1-033df78173ec"
	emnB1EnlightenedManiacOracle = "4310524d-dc0e-48ad-ae66-4c64aaa39340"
	emnB1DesperateSentryOracle   = "72e0f7f2-1d04-4ea8-a6e8-4d6a13aae9c7"
	emnB1GeistArchivesOracle     = "345964ea-12b6-44be-b1a6-50178768b8f4"
	emnB1SuccumbOracle           = "c8e21bc4-ea57-46d6-9e7e-ad95e17ea0fd"
	emnB1WoodcuttersGritOracle   = "7756d2b9-7b10-4d05-8ad3-8d8f39c1eb23"
	emnB1PrimalDruidOracle       = "d031d7d2-1e41-4201-9073-c4c1be182ba4"
	emnB1ThaliasLancersOracle    = "58b349ec-9987-4832-a8f3-8e92027ea75b"
	emnB1SpreadingFlamesOracle   = "95bedd11-ed01-4cb1-bcdb-c5f9fdba8c79"
	emnB1CrossroadsOracle        = "a4f41c7b-b4ae-46cd-9fdc-8b8e6c29a62d"
	emnB1WolfkinBondOracle       = "49e22e88-d4d9-4781-81e3-bf77b76b132b"
	emnB1SkirsdagOracle          = "c7d97311-6e34-4679-a678-e107022c785c"
	emnB1StandardBearerOracle    = "679d5a1d-8da8-4142-924b-97b359f55309"

	emnB1BearType  = "Creature — Bear"
	emnB1HumanType = "Creature — Human Soldier"
	emnB1AuraType  = "Enchantment — Aura"
)

// --- every card is in the catalog ----------------------------------------

func TestEmn5541B1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		emnB1DawnGryffOracle: "Dawn Gryff", emnB1DerangedWhelpOracle: "Deranged Whelp",
		emnB1FaithbearerOracle: "Faithbearer Paladin", emnB1FalkenrathReaverOracle: "Falkenrath Reaver",
		emnB1FieldCreeperOracle: "Field Creeper", emnB1SwiftSpinnerOracle: "Swift Spinner",
		emnB1WoodlandPatrolOracle: "Woodland Patrol", emnB1VampireCutthroatOracle: "Vampire Cutthroat",
		emnB1WeirdedVampireOracle: "Weirded Vampire", emnB1BloodbriarOracle: "Bloodbriar",
		emnB1BrazenWolvesOracle: "Brazen Wolves", emnB1SteadfastCatharOracle: "Steadfast Cathar",
		emnB1CultistsStaffOracle: "Cultist's Staff", emnB1ExultantCultistOracle: "Exultant Cultist",
		emnB1GavonyUnhallowedOracle: "Gavony Unhallowed", emnB1IronwrightsOracle: "Ironwright's Cleansing",
		emnB1SigardianPriestOracle: "Sigardian Priest", emnB1SomberwaldStagOracle: "Somberwald Stag",
		emnB1SpringsageRitualOracle: "Springsage Ritual", emnB1TatteredHaunterOracle: "Tattered Haunter",
		emnB1AlchemistsOracle: "Alchemist's Greeting", emnB1DistemperOracle: "Distemper of the Blood",
		emnB1ChillingGraspOracle: "Chilling Grasp", emnB1InsatiableGorgersOracle: "Insatiable Gorgers",
		emnB1BackwoodsOracle: "Backwoods Survivalists", emnB1FoulbloodsOracle: "Thraben Foulbloods",
		emnB1BoldImpalerOracle: "Bold Impaler", emnB1ChokingOracle: "Choking Restraints",
		emnB1DisplaceOracle: "Displace", emnB1FiendBinderOracle: "Fiend Binder",
		emnB1StensiaInnkeeperOracle: "Stensia Innkeeper",
		emnB1IngeniousSkaabOracle:   "Ingenious Skaab", emnB1MercurialGeistsOracle: "Mercurial Geists",
		emnB1LaboratoryBruteOracle: "Laboratory Brute", emnB1LilianasEliteOracle: "Liliana's Elite",
		emnB1SpectralReservesOracle: "Spectral Reserves", emnB1CertainDeathOracle: "Certain Death",
		emnB1ContingencyPlanOracle: "Contingency Plan", emnB1EnlightenedManiacOracle: "Enlightened Maniac",
		emnB1DesperateSentryOracle: "Desperate Sentry", emnB1GeistArchivesOracle: "Geist of the Archives",
		emnB1SuccumbOracle: "Succumb to Temptation", emnB1WoodcuttersGritOracle: "Woodcutter's Grit",
		emnB1PrimalDruidOracle: "Primal Druid", emnB1ThaliasLancersOracle: "Thalia's Lancers",
		emnB1SpreadingFlamesOracle: "Spreading Flames", emnB1CrossroadsOracle: "Crossroads Consecrator",
		emnB1WolfkinBondOracle: "Wolfkin Bond", emnB1SkirsdagOracle: "Skirsdag Supplicant",
		emnB1StandardBearerOracle: "Thraben Standard Bearer",
	}
	if len(want) != 50 {
		t.Fatalf("the batch registers 50 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- vanilla bodies and printed keywords -----------------------------------

func TestEmn5541B1VanillaBodiesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{emnB1FalkenrathReaverOracle, emnB1FieldCreeperOracle} {
		spec, _ := Lookup(oracle)
		if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 ||
			len(spec.Static) != 0 || spec.OnResolve != nil {
			t.Errorf("%s registers an ability; it is a vanilla creature", spec.Name)
		}
	}
}

func TestEmn5541B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Dawn Gryff", "Creature — Hippogriff", emnB1DawnGryffOracle, []string{"flying"}},
		{"Deranged Whelp", "Creature — Wolf", emnB1DerangedWhelpOracle, []string{"menace"}},
		{"Faithbearer Paladin", "Creature — Human Knight", emnB1FaithbearerOracle, []string{"lifelink"}},
		{"Swift Spinner", "Creature — Spider", emnB1SwiftSpinnerOracle, []string{"flash", "reach"}},
		{"Woodland Patrol", "Creature — Human Scout", emnB1WoodlandPatrolOracle, []string{"vigilance"}},
		{"Vampire Cutthroat", "Creature — Vampire Rogue", emnB1VampireCutthroatOracle, []string{"skulk", "lifelink"}},
		{"Tattered Haunter", "Creature — Spirit", emnB1TatteredHaunterOracle, []string{"flying"}},
		{"Ingenious Skaab", "Creature — Zombie Horror", emnB1IngeniousSkaabOracle, []string{"prowess"}},
		{"Mercurial Geists", "Creature — Spirit", emnB1MercurialGeistsOracle, []string{"flying"}},
		{"Geist of the Archives", "Creature — Spirit", emnB1GeistArchivesOracle, []string{"defender"}},
		{"Thalia's Lancers", "Creature — Human Knight", emnB1ThaliasLancersOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Flash: Swift Spinner is cast at the beginning of combat.
func TestEmn5541B1SwiftSpinnerHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Swift Spinner", "Creature — Spider", emnB1SwiftSpinnerOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Swift Spinner at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Swift Spinner did not resolve")
	}
}

// Skulk (CR 702.118b): a creature with greater power can't block the
// Cutthroat; one with equal power can.
func TestEmn5541B1VampireCutthroatHasSkulk(t *testing.T) {
	for _, tc := range []struct {
		power int
		want  bool
	}{{2, false}, {1, true}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cutthroat := b12Push(g, me.ID, "Vampire Cutthroat", "Creature — Vampire Rogue", emnB1VampireCutthroatOracle, 1, 1)
		blocker := b12Creature(g, opp.ID, "Blocker", emnB1BearType, tc.power, 4)
		err := porBlock(t, g, cutthroat, opp.ID, blocker)
		if tc.want && err != nil {
			t.Errorf("a %d-power creature couldn't block the Cutthroat: %v", tc.power, err)
		}
		if !tc.want && err == nil {
			t.Errorf("a %d-power creature blocked the Cutthroat", tc.power)
		}
	}
}

func TestEmn5541B1TatteredHaunterBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", emnB1HumanType, 2, 2)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	haunter := b12Push(g, opp.ID, "Tattered Haunter", "Creature — Spirit", emnB1TatteredHaunterOracle, 2, 1)
	brAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(haunter, walker); err == nil {
		t.Error("Tattered Haunter blocked a creature without flying")
	}
	if err := g.DeclareBlocker(haunter, flyer); err != nil {
		t.Errorf("Tattered Haunter can't block a flyer: %v", err)
	}
}

// --- madness ----------------------------------------------------------------

// Each madness card grows both halves of the keyword from its one
// Madness string: the discard replacement and the exile trigger.
func TestEmn5541B1MadnessCardsGrowBothHalves(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
	}{
		{"Weirded Vampire", emnB1WeirdedVampireOracle, "{2}{B}"},
		{"Alchemist's Greeting", emnB1AlchemistsOracle, "{1}{R}"},
		{"Distemper of the Blood", emnB1DistemperOracle, "{R}"},
		{"Chilling Grasp", emnB1ChillingGraspOracle, "{3}{U}"},
		{"Insatiable Gorgers", emnB1InsatiableGorgersOracle, "{3}{R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := Lookup(tc.oracle)
			if spec.Madness != tc.cost {
				t.Errorf("madness cost %q, want %q", spec.Madness, tc.cost)
			}
			reps, trigs := madnessAbilitiesFor(tc.oracle)
			if reps != 1 || trigs != 1 {
				t.Errorf("%d discard replacements and %d exile triggers, want 1 and 1", reps, trigs)
			}
		})
	}
}

func TestEmn5541B1AlchemistsGreetingDealsFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4)
	if err := castCatalogSpellErr(t, g, "Alchemist's Greeting", "Sorcery", emnB1AlchemistsOracle, []game.TargetRef{playerTarget(opp.ID)}); err == nil {
		t.Fatal("Alchemist's Greeting targeted a player")
	}
	castCatalogSpell(t, g, "Alchemist's Greeting", "Sorcery", emnB1AlchemistsOracle, cardTarget(ox))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ox) {
		t.Error("a 4-toughness creature survived 4 damage")
	}
}

func TestEmn5541B1DistemperOfTheBloodPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", emnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Distemper of the Blood", "Sorcery", emnB1DistemperOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

func TestEmn5541B1ChillingGraspTapsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", emnB1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", emnB1BearType, 2, 2)
	c := b12Creature(g, opp.ID, "C", emnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Chilling Grasp", "Instant", emnB1ChillingGraspOracle, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	if !tappedNow(g, a) || !tappedNow(g, b) {
		t.Error("a target was left untapped")
	}
	if tappedNow(g, c) {
		t.Error("a creature that was not targeted was tapped")
	}
}

func TestEmn5541B1InsatiableGorgersMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	gorgers := b12Push(g, me.ID, "Insatiable Gorgers", "Creature — Vampire Berserker", emnB1InsatiableGorgersOracle, 5, 3)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != gorgers {
		t.Fatalf("pass with the Gorgers home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(gorgers, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Gorgers attacking: %v", err)
	}
}

// --- delirium -----------------------------------------------------------------

func TestEmn5541B1DeliriumBodiesGrowAndGainTheirKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, tough                    int
	}{
		{"Backwoods Survivalists", "Creature — Human Warrior", emnB1BackwoodsOracle, "trample", 4, 3},
		{"Thraben Foulbloods", "Creature — Zombie Dog", emnB1FoulbloodsOracle, "menace", 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			dskRT2Delirium(g, me, 3)
			if p, tt := ptOf(t, g, id); p != tc.power || tt != tc.tough {
				t.Errorf("three card types: %d/%d, want %d/%d", p, tt, tc.power, tc.tough)
			}
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("three card types: %s", tc.keyword)
			}
			dskRT2Delirium(g, me, 4)
			if p, tt := ptOf(t, g, id); p != tc.power+1 || tt != tc.tough+1 {
				t.Errorf("delirium: %d/%d, want %d/%d", p, tt, tc.power+1, tc.tough+1)
			}
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("delirium: no %s", tc.keyword)
			}
		})
	}
}

func TestEmn5541B1DesperateSentryGrowsWithDeliriumAndLeavesAHorror(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sentry := b12Push(g, me.ID, "Desperate Sentry", emnB1HumanType, emnB1DesperateSentryOracle, 1, 2)
	dskRT2Delirium(g, me, 3)
	if p, _ := ptOf(t, g, sentry); p != 1 {
		t.Errorf("three card types: power %d, want 1", p)
	}
	dskRT2Delirium(g, me, 4)
	if p, tt := ptOf(t, g, sentry); p != 4 || tt != 2 {
		t.Errorf("delirium: %d/%d, want 4/2", p, tt)
	}
	b15Destroy(t, g, sentry)
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens after it died, want 1", n)
	}
}

// --- combat triggers ------------------------------------------------------------

func TestEmn5541B1AttackPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		wantP, wantT           int
	}{
		{"Brazen Wolves", "Creature — Wolf", emnB1BrazenWolvesOracle, 2, 3, 4, 3},
		{"Steadfast Cathar", emnB1HumanType, emnB1SteadfastCatharOracle, 2, 1, 2, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			declareAttack(t, g, them.ID, id)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d while attacking, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestEmn5541B1FiendBinderTapsADefendersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := b12Push(g, me.ID, "Fiend Binder", emnB1HumanType, emnB1FiendBinderOracle, 3, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	bystander := b12Creature(g, g.Seats[2].ID, "Other Bear", emnB1BearType, 2, 2)
	declareAttack(t, g, opp.ID, binder)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("the defending player's creature was not tapped")
	}
	if tappedNow(g, bystander) {
		t.Error("a creature of a player who is not defending was tapped")
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestEmn5541B1BloodbriarGrowsOnOtherSacrificesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	briar := b12Push(g, me.ID, "Bloodbriar", "Creature — Plant Elemental", emnB1BloodbriarOracle, 2, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Theirs", emnB1BearType, 2, 2)
	broSac(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, briar, game.CounterPlusOne); n != 0 {
		t.Fatalf("an opponent's sacrifice gave %d counters", n)
	}
	broSac(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, briar, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after you sacrificed a permanent, want 1", n)
	}
}

func TestEmn5541B1GavonyUnhallowedGrowsWhenYourOtherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := b12Push(g, me.ID, "Gavony Unhallowed", "Creature — Zombie", emnB1GavonyUnhallowedOracle, 2, 4)
	mine := b12Creature(g, me.ID, "My Bear", emnB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	b15Destroy(t, g, theirs)
	if n := countersOn(g, zombie, game.CounterPlusOne); n != 0 {
		t.Fatalf("an opponent's creature dying gave %d counters", n)
	}
	b15Destroy(t, g, mine)
	if n := countersOn(g, zombie, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after your creature died, want 1", n)
	}
}

func TestEmn5541B1ExultantCultistDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	cultist := b12Push(g, me.ID, "Exultant Cultist", "Creature — Human Wizard", emnB1ExultantCultistOracle, 2, 2)
	hand := me.Hand.Size()
	b15Destroy(t, g, cultist)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestEmn5541B1SomberwaldStagMayFight(t *testing.T) {
	for _, yes := range []bool{true, false} {
		name := "declined"
		if yes {
			name = "fights"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
			stag := castHoldCreature(t, g, "Somberwald Stag", "Creature — Elk", emnB1SomberwaldStagOracle, 4, 3)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, victim)
			answerLatestTriggerPrompt(t, g, me.ID, yes)
			passPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if got := g.Battlefield.Contains(victim); got == yes {
				t.Errorf("the 2/2 is on the battlefield: %v", got)
			}
			want := 0
			if yes {
				want = 2
			}
			if got := damageMarkedOn(g, stag); got != want {
				t.Errorf("the Stag has %d damage, want %d", got, want)
			}
		})
	}
}

func TestEmn5541B1EnlightenedManiacMakesAHorror(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Enlightened Maniac", "Creature — Human", emnB1EnlightenedManiacOracle, 0, 2)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", n)
	}
}

func TestEmn5541B1LaboratoryBruteMillsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	castHoldCreature(t, g, "Laboratory Brute", "Creature — Zombie Horror", emnB1LaboratoryBruteOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != 4 {
		t.Errorf("%d cards reached the graveyard, want 4", got)
	}
}

func TestEmn5541B1StensiaInnkeeperTapsAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	bear := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	castHoldCreature(t, g, "Stensia Innkeeper", "Creature — Vampire", emnB1StensiaInnkeeperOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	if hasID(latestPickTarget(g, me.ID).PickTargetCards, bear) {
		t.Error("a creature is offered as the target")
	}
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, land) {
		t.Error("the opponent's land was not tapped")
	}
}

func TestEmn5541B1PrimalDruidMayFetchABasicTappedWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	druid := b12Push(g, me.ID, "Primal Druid", "Creature — Human Druid", emnB1PrimalDruidOracle, 0, 3)
	b15Destroy(t, g, druid)
	takeSearchPrompts(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was not fetched")
	}
	if !tappedNow(g, forest) {
		t.Error("the fetched land entered untapped")
	}
}

func TestEmn5541B1ThaliasLancersMayTutorALegendaryCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := stapleLibraryCard(me, "Plain Bear", emnB1BearType)
	legend := stapleLibraryCard(me, "Old Legend", "Legendary Creature — Human Knight")
	castHoldCreature(t, g, "Thalia's Lancers", "Creature — Human Knight", emnB1ThaliasLancersOracle, 4, 4)
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(legend) {
		t.Error("the legendary card is not in your hand")
	}
	if me.Hand.Contains(plain) {
		t.Error("a nonlegendary card was tutored")
	}
}

func TestEmn5541B1GeistOfTheArchivesScriesAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E")
	b12Push(g, me.ID, "Geist of the Archives", "Creature — Spirit", emnB1GeistArchivesOracle, 0, 4)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry at the upkeep")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
}

func TestEmn5541B1MercurialGeistsGrowOnAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	geists := b12Push(g, me.ID, "Mercurial Geists", "Creature — Spirit", emnB1MercurialGeistsOracle, 1, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Think", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, geists); p != 4 || tt != 3 {
		t.Errorf("after a sorcery: %d/%d, want 4/3", p, tt)
	}
}

// --- activated abilities -------------------------------------------------------------

func TestEmn5541B1IngeniousSkaabProwessAndPump(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	skaab := b12Push(g, me.ID, "Ingenious Skaab", "Creature — Zombie Horror", emnB1IngeniousSkaabOracle, 2, 3)
	castCatalogSpell(t, g, "Think", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, skaab); p != 3 || tt != 4 {
		t.Fatalf("after a noncreature spell: %d/%d, want 3/4", p, tt)
	}
	apaMana(me, "U")
	apaActivate(t, g, me, skaab, 0, game.ActivateAbilityParams{})
	if p, tt := ptOf(t, g, skaab); p != 4 || tt != 3 {
		t.Errorf("after {U}: %d/%d, want 4/3", p, tt)
	}
}

func TestEmn5541B1BoldImpalerPumpsForTwoAndRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	impaler := b12Push(g, me.ID, "Bold Impaler", "Creature — Vampire Knight", emnB1BoldImpalerOracle, 1, 2)
	apaMana(me, "C", "C", "R")
	apaActivate(t, g, me, impaler, 0, game.ActivateAbilityParams{})
	if p, tt := ptOf(t, g, impaler); p != 3 || tt != 2 {
		t.Errorf("after the pump: %d/%d, want 3/2", p, tt)
	}
}

func TestEmn5541B1SigardianPriestTapsOnlyANonHuman(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	priest := b12Push(g, me.ID, "Sigardian Priest", "Creature — Human Cleric", emnB1SigardianPriestOracle, 1, 2)
	human := b12Creature(g, opp.ID, "Their Soldier", emnB1HumanType, 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(human)}); err == nil {
		t.Fatal("the Priest targeted a Human")
	}
	apaActivate(t, g, me, priest, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !tappedNow(g, bear) {
		t.Error("the non-Human creature is untapped")
	}
}

func TestEmn5541B1CrossroadsConsecratorPumpsAnAttackingHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cleric := b12Push(g, me.ID, "Crossroads Consecrator", "Creature — Human Cleric", emnB1CrossroadsOracle, 1, 2)
	human := b12Creature(g, me.ID, "Soldier", emnB1HumanType, 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	brAttack(t, g, human, elf)
	apaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(elf)}); err == nil {
		t.Fatal("the Consecrator targeted an attacker that isn't a Human")
	}
	apaActivate(t, g, me, cleric, 0, game.ActivateAbilityParams{Targets: cardTarget(human)})
	if p, tt := ptOf(t, g, human); p != 3 || tt != 3 {
		t.Errorf("the attacking Human is %d/%d, want 3/3", p, tt)
	}
}

func TestEmn5541B1SkirsdagSupplicantDiscardsForEachPlayerToLoseTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	cleric := b12Push(g, me.ID, "Skirsdag Supplicant", "Creature — Human Cleric", emnB1SkirsdagOracle, 2, 3)
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	lives := b27Lives(g)
	apaMana(me, "B")
	apaActivate(t, g, me, cleric, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}})
	if !me.Graveyard.Contains(spare) {
		t.Error("the card was not discarded")
	}
	for i, l := range b27Lives(g) {
		if l != lives[i]-2 {
			t.Errorf("seat %d life %d -> %d, want -2", i, lives[i], l)
		}
	}
}

func TestEmn5541B1ThrabenStandardBearerMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bearer := b12Push(g, me.ID, "Thraben Standard Bearer", emnB1HumanType, emnB1StandardBearerOracle, 1, 1)
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	apaMana(me, "C", "W")
	apaActivate(t, g, me, bearer, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}})
	if !me.Graveyard.Contains(spare) {
		t.Error("the card was not discarded")
	}
	if n, _ := countTokensNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

// --- equipment and auras ----------------------------------------------------------

func TestEmn5541B1CultistsStaffGivesTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	staff := seedEquipment(g, me.ID, "Cultist's Staff", emnB1CultistsStaffOracle)
	if got := equipPrice(t, g, me.ID, staff, bear); got != "{3}" {
		t.Errorf("equip costs %s, want {3}", got)
	}
	equipTo(t, g, me.ID, staff, bear)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 4/4", p, tt)
	}
}

func TestEmn5541B1ChokingRestraintsStopsCombatAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std51Seats(g)
	victim := pushVanillaCreature(g, me.ID, "Enchanted", 2, 2)
	aura := castCatalogSpell(t, g, "Choking Restraints", emnB1AuraType, emnB1ChokingOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Choking Restraints did not resolve")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(victim, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	std51Pay(g, me, "{3}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if !g.Exile.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the enchanted creature was not exiled")
	}
}

func TestEmn5541B1WolfkinBondMakesAWolfAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", emnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Wolfkin Bond", emnB1AuraType, emnB1WolfkinBondOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Errorf("%d Wolf tokens, want 1", n)
	}
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tt)
	}
}

func TestEmn5541B1LilianasEliteCountsCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gyCard(me, "Old Bear", emnB1BearType, "{1}{G}", "")
	gyCard(me, "Old Wolf", "Creature — Wolf", "{1}{R}", "")
	gyCard(me, "Old Bolt", "Instant", "{R}", "")
	elite := b12Push(g, me.ID, "Liliana's Elite", "Creature — Zombie", emnB1LilianasEliteOracle, 1, 1)
	s29Bump(g)
	if p, tt := ptOf(t, g, elite); p != 3 || tt != 3 {
		t.Errorf("with two creature cards: %d/%d, want 3/3", p, tt)
	}
}

// --- the spells --------------------------------------------------------------------

func TestEmn5541B1ArtifactOrEnchantmentRemoval(t *testing.T) {
	t.Run("Ironwright's Cleansing exiles", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
		bear := b12Creature(g, opp.ID, "Bear", emnB1BearType, 2, 2)
		if err := castCatalogSpellErr(t, g, "Ironwright's Cleansing", "Sorcery", emnB1IronwrightsOracle, cardTarget(bear)); err == nil {
			t.Fatal("it targeted a creature")
		}
		castCatalogSpell(t, g, "Ironwright's Cleansing", "Sorcery", emnB1IronwrightsOracle, cardTarget(rock))
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(rock) {
			t.Error("the artifact was not exiled")
		}
	})
	t.Run("Springsage Ritual destroys and gains four", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		charm := b12Permanent(g, opp.ID, "Charm", "Enchantment")
		life := me.Life
		castCatalogSpell(t, g, "Springsage Ritual", "Instant", emnB1SpringsageRitualOracle, cardTarget(charm))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(charm) {
			t.Error("the enchantment survived")
		}
		if me.Life != life+4 {
			t.Errorf("life %d -> %d, want +4", life, me.Life)
		}
	})
}

func TestEmn5541B1CertainDeathDestroysAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	mine, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Certain Death", "Sorcery", emnB1CertainDeathOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("its controller %d -> %d (want -2), you %d -> %d (want +2)", theirs, opp.Life, mine, me.Life)
	}
}

func TestEmn5541B1SpectralReservesMakesTwoSpiritsAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Spectral Reserves", "Sorcery", emnB1SpectralReservesOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 2 {
		t.Errorf("%d Spirit tokens, want 2", n)
	}
	if id := findBattlefieldByName(g, "Spirit"); id == uuid.Nil || !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the Spirit has no flying")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestEmn5541B1ContingencyPlanSurveilsFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	castCatalogSpell(t, g, "Contingency Plan", "Sorcery", emnB1ContingencyPlanOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 5 {
		t.Errorf("surveil looks at %d cards, want 5", len(c.ScryCards))
	}
}

func TestEmn5541B1SuccumbToTemptationDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Succumb to Temptation", "Instant", emnB1SuccumbOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
}

func TestEmn5541B1WoodcuttersGritPumpsYourCreatureWithHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", emnB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Woodcutter's Grit", "Instant", emnB1WoodcuttersGritOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Woodcutter's Grit targeted an opponent's creature")
	}
	castCatalogSpell(t, g, "Woodcutter's Grit", "Instant", emnB1WoodcuttersGritOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the Bear has no hexproof")
	}
}

func TestEmn5541B1SpreadingFlamesDividesSix(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Spreading Flames", "Instant", emnB1SpreadingFlamesOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 4},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a: %d damage, want 2", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 4 {
		t.Errorf("b: %d damage, want 4", d)
	}
}

// Displace blinks your creatures together; a creature you control but
// don't own comes back under its owner's control.
func TestEmn5541B1DisplaceBlinksUnderTheirOwners(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := roonaStolen(g, me.ID, me.ID, "My Bear")
	stolen := roonaStolen(g, opp.ID, me.ID, "Stolen Bear")
	theirs := b12Creature(g, opp.ID, "Their Bear", emnB1BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Displace", "Instant", emnB1DisplaceOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Displace targeted an opponent's creature")
	}
	castCatalogSpell(t, g, "Displace", "Instant", emnB1DisplaceOracle, cardTargets(mine, stolen))
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, mine) || onBattlefieldByID(g, stolen) {
		t.Fatal("a target is still the same object (CR 400.7)")
	}
	if c := roonaOnly(t, g, "My Bear"); c.Controller != me.ID {
		t.Errorf("your Bear returned under %v, want %v", c.Controller, me.ID)
	}
	if c := roonaOnly(t, g, "Stolen Bear"); c.Controller != opp.ID {
		t.Errorf("the stolen Bear returned under %v, want its owner %v", c.Controller, opp.ID)
	}
}
