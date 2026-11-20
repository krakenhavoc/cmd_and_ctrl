package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eld_5516_b1_cards_test.go — Throne of Eldraine, batch 1 (#5516). Every
// test plays the real catalog entry through the engine.

const (
	eldB1KnightOfTheKeep      = "ab1d04ef-5f39-4a95-969e-b0872e6db0d6"
	eldB1OkosAccomplices      = "93cd12da-a7d5-4d01-a3bc-77e441cb0947"
	eldB1PrizedGriffin        = "d383902c-63ba-4714-bfeb-377829950b0d"
	eldB1TomeRaider           = "ecd9aa9a-06b2-4d8d-9264-fda2036ff142"
	eldB1LostLegion           = "b8f3b5b5-8835-400f-a626-94ff0a007e56"
	eldB1ProphetOfThePeak     = "eaa1634e-6ba1-41c2-91c0-967b36d13e14"
	eldB1CorridorMonitor      = "07b6ee55-a77e-4ba7-9ab7-8a8375c7acc9"
	eldB1BarrowWitches        = "9d4b34cc-0452-42a4-90e4-2a48e7ffadc0"
	eldB1BlacklanceParagon    = "258a78cc-a9f3-459a-9837-93a9856f72db"
	eldB1BurningYardTrainer   = "6ea7a921-45b2-4589-833b-75b6dc6aed07"
	eldB1VenerableKnight      = "ef027846-be81-4959-a6b5-56bd01b1e68a"
	eldB1GarrisonGriffin      = "2ae828d1-0c8b-4558-bd0b-512760020913"
	eldB1InspiringVeteran     = "aa1a63dd-acb1-465f-8970-667b8d7c57c9"
	eldB1JoustingDummy        = "96eea161-8af0-452f-b992-576efdf58d88"
	eldB1LocthwainGargoyle    = "d8a04142-1d28-4c2e-b2d5-da88b2199678"
	eldB1FirebornKnight       = "b77695da-86fb-4675-b2d3-51732ad0325f"
	eldB1MaraleafPixie        = "013e4281-4e3e-47c2-ac9a-f570668bcb16"
	eldB1SignpostScarecrow    = "5b6c6dab-0687-4297-b215-89e699d65681"
	eldB1SpinningWheel        = "e0935a0f-7087-40cd-a094-89f200da6ef0"
	eldB1ScaldingCauldron     = "dceb96a4-d3c5-43a2-93f6-50316e1e7a1b"
	eldB1TemptingWitch        = "82284fe4-ecb3-4b27-bb03-27290ad15b17"
	eldB1CrystalSlipper       = "8c2d2888-38be-41f9-aaf2-8337e87d2aed"
	eldB1BelovedPrincess      = "682cbf2a-2723-4590-ae52-7c5d8b2573cf"
	eldB1FortifyingProvisions = "f00dc657-4695-497d-aad3-af7b6a3ccf94"
	eldB1BloodhazeWolverine   = "601fd61b-4939-459a-84d1-d947d29c9564"
	eldB1SteelgazeGriffin     = "3fdc082b-1abf-4715-9d66-105981c1f989"
	eldB1MadRatter            = "779fa24d-e9d2-475d-b3aa-dc0388b18307"
	eldB1IrencragPyromancer   = "5abcc262-fbe1-4e81-897c-da5a8fb1f1f4"
	eldB1ThunderousSnapper    = "f826ac35-7dde-414c-94ec-0bdfe04327b4"
	eldB1FellThePheasant      = "c78c8f7d-45fd-46c0-9ed4-4d0e85aada2e"
	eldB1LashOfThorns         = "642d3494-7a6a-4aec-ba64-7c0c467ada99"
	eldB1TrueLovesKiss        = "9506aba2-5e1d-48c5-ace8-4beb824ad8af"
	eldB1ForebodingFruit      = "33a162f3-e781-494f-ab8f-12a84801a694"
	eldB1UnexplainedVision    = "9f693483-f4df-497a-a6c7-fc269fc60ea8"
	eldB1TurnIntoAPumpkin     = "cfda6f65-3755-40e5-8061-a21941c4b1bd"

	eldB1CuriousPair         = "e67bd8eb-66db-48af-82f8-74f5595f1928"
	eldB1EmberethShieldbreak = "52e439d0-a263-4d53-9ae3-fa1f8aa8293e"
	eldB1GarenbrigCarver     = "496c9633-ff17-48df-9b24-1d8cbfd06126"
	eldB1LonesomeUnicorn     = "c8f6cc1f-b7e4-470d-8d48-ef848d7c1116"
	eldB1MerfolkSecretkeeper = "4f5096d4-bbb5-4a48-9dd0-9dabdd2ac125"
	eldB1RimrockKnight       = "24db950b-e4fe-42d0-977b-4957eb8229d9"
	eldB1ShepherdOfTheFlock  = "6d611800-5b0e-40ca-82a2-5b294a1ebf6d"
	eldB1SilverflameSquire   = "93e4308e-e701-440e-861a-a30324654095"
	eldB1TuinvaleTreefolk    = "09193c40-538c-4154-a671-076eece35e54"
	eldB1FaerieGuidemother   = "04fbb7aa-0dd4-4ac6-837d-822227fb6356"
	eldB1ArdenvaleTactician  = "ea27c8bf-33c5-443f-b8b2-51235efc2491"
	eldB1OakhameRanger       = "9c247a17-361c-435e-90c8-ce9c6b71f7db"
	eldB1GiantKiller         = "91e71279-c9d2-4873-916a-59da03a65741"

	eldB1KnightType = "Creature — Human Knight"
	eldB1BearType   = "Creature — Bear"
)

var eldB1Adventures = []string{
	eldB1CuriousPair, eldB1EmberethShieldbreak, eldB1GarenbrigCarver, eldB1LonesomeUnicorn,
	eldB1MerfolkSecretkeeper, eldB1RimrockKnight, eldB1ShepherdOfTheFlock, eldB1SilverflameSquire,
	eldB1TuinvaleTreefolk, eldB1FaerieGuidemother, eldB1ArdenvaleTactician, eldB1OakhameRanger,
	eldB1GiantKiller,
}

// eldB1GraveyardCard puts a card of the given type line into p's graveyard.
func eldB1GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// eldB1DrawToSecond has p draw until it has drawn two cards this turn, so
// the last draw is the turn's second.
func eldB1DrawToSecond(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	for g.TurnTallyFor(p.ID).CardsDrawn < 2 {
		seedLibraryTop(p, "Plains", "Basic Land — Plains", "")
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
	}
}

// eldB1CastPaid casts a catalog spell from the active player's hand,
// strictly paid from `pool`, and resolves it.
func eldB1CastPaid(t *testing.T, g *game.Game, name, typeLine, cost, oracle, pool string, targets []game.TargetRef) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	floatForTest(g, me, pool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

// eldB1Food counts the Food tokens p controls.
func eldB1Food(g *game.Game, p *game.Player) int {
	return tokensNamedControlledBy(g, "Food", p.ID)
}

// --- every card is in the catalog ----------------------------------------

func TestEld5516EveryCardIsRegisteredFull(t *testing.T) {
	oracles := []string{
		eldB1KnightOfTheKeep, eldB1OkosAccomplices, eldB1PrizedGriffin, eldB1TomeRaider,
		eldB1LostLegion, eldB1ProphetOfThePeak, eldB1CorridorMonitor, eldB1BarrowWitches,
		eldB1BlacklanceParagon, eldB1BurningYardTrainer, eldB1VenerableKnight, eldB1GarrisonGriffin,
		eldB1InspiringVeteran, eldB1JoustingDummy, eldB1LocthwainGargoyle, eldB1FirebornKnight,
		eldB1MaraleafPixie, eldB1SignpostScarecrow, eldB1SpinningWheel, eldB1ScaldingCauldron,
		eldB1TemptingWitch, eldB1CrystalSlipper, eldB1BelovedPrincess, eldB1FortifyingProvisions,
		eldB1BloodhazeWolverine, eldB1SteelgazeGriffin, eldB1MadRatter, eldB1IrencragPyromancer,
		eldB1ThunderousSnapper, eldB1FellThePheasant, eldB1LashOfThorns, eldB1TrueLovesKiss,
		eldB1ForebodingFruit, eldB1UnexplainedVision, eldB1TurnIntoAPumpkin,
	}
	for _, a := range eldB1Adventures {
		oracles = append(oracles, a, a+"#1")
	}
	for _, oracle := range oracles {
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

// --- printed keywords -----------------------------------------------------------

func TestEld5516KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Oko's Accomplices", "Creature — Faerie", eldB1OkosAccomplices, []string{"flying"}},
		{"Prized Griffin", "Creature — Griffin", eldB1PrizedGriffin, []string{"flying"}},
		{"Tome Raider", "Creature — Faerie", eldB1TomeRaider, []string{"flying"}},
		{"Blacklance Paragon", eldB1KnightType, eldB1BlacklanceParagon, []string{"flash"}},
		{"Burning-Yard Trainer", eldB1KnightType, eldB1BurningYardTrainer, []string{"trample", "haste"}},
		{"Garrison Griffin", "Creature — Griffin", eldB1GarrisonGriffin, []string{"flying"}},
		{"Fireborn Knight", eldB1KnightType, eldB1FirebornKnight, []string{"double strike"}},
		{"Maraleaf Pixie", "Creature — Faerie", eldB1MaraleafPixie, []string{"flying"}},
		{"Signpost Scarecrow", "Artifact Creature — Scarecrow", eldB1SignpostScarecrow, []string{"vigilance"}},
		{"Beloved Princess", "Creature — Human Noble", eldB1BelovedPrincess, []string{"lifelink"}},
		{"Steelgaze Griffin", "Creature — Griffin", eldB1SteelgazeGriffin, []string{"flying"}},
		{"Lonesome Unicorn", "Creature — Unicorn", eldB1LonesomeUnicorn, []string{"vigilance"}},
		{"Faerie Guidemother", "Creature — Faerie", eldB1FaerieGuidemother, []string{"flying"}},
		{"Ardenvale Tactician", eldB1KnightType, eldB1ArdenvaleTactician, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- enters triggers ----------------------------------------------------------

func TestEld5516TomeRaiderDrawsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Tome Raider", "Creature — Faerie", eldB1TomeRaider, 1, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestEld5516EntersAndScriesTwo(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Lost Legion", "Creature — Spirit Knight", eldB1LostLegion},
		{"Prophet of the Peak", "Artifact Creature — Cat", eldB1ProphetOfThePeak},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 3)
			passPriorityAroundTable(t, g)
			c := scryChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no scry prompt")
			}
			if len(c.ScryCards) != 2 {
				t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
			}
		})
	}
}

func TestEld5516CorridorMonitorUntapsYourArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Tapped: true})
	castHoldCreature(t, g, "Corridor Monitor", "Artifact Creature — Construct", eldB1CorridorMonitor, 1, 4)
	answerPickTargetIfOpen(t, g, rock)
	passPriorityAroundTable(t, g)
	if hmlTapped(t, g, rock) {
		t.Error("the artifact is still tapped")
	}
}

func TestEld5516BarrowWitchesReturnsAKnightCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := eldB1GraveyardCard(me, "Old Knight", eldB1KnightType)
	eldB1GraveyardCard(me, "Old Bear", eldB1BearType)
	castHoldCreature(t, g, "Barrow Witches", "Creature — Human Warlock", eldB1BarrowWitches, 3, 4)
	answerPickTargetIfOpen(t, g, knight)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(knight) || me.Graveyard.Contains(knight) {
		t.Error("the Knight card did not return to hand")
	}
}

func TestEld5516BlacklanceParagonGivesAKnightDeathtouchAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	knight := b12Creature(g, opp.ID, "Their Knight", eldB1KnightType, 2, 2)
	castHoldCreature(t, g, "Blacklance Paragon", eldB1KnightType, eldB1BlacklanceParagon, 3, 1)
	answerPickTargetIfOpen(t, g, knight)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !hasEffectiveKeyword(t, g, knight, kw) {
			t.Errorf("the Knight has no %s", kw)
		}
	}
}

func TestEld5516BurningYardTrainerPumpsAnotherKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := b12Creature(g, me.ID, "My Knight", eldB1KnightType, 2, 2)
	castHoldCreature(t, g, "Burning-Yard Trainer", eldB1KnightType, eldB1BurningYardTrainer, 3, 3)
	answerPickTargetIfOpen(t, g, knight)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, knight); p != 4 || tough != 4 {
		t.Errorf("the Knight is %d/%d, want 4/4", p, tough)
	}
	for _, kw := range []string{"trample", "haste"} {
		if !hasEffectiveKeyword(t, g, knight, kw) {
			t.Errorf("the Knight has no %s", kw)
		}
	}
}

func TestEld5516VenerableKnightCountersAKnightWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	venerable := b12Push(g, me.ID, "Venerable Knight", eldB1KnightType, eldB1VenerableKnight, 2, 1)
	knight := b12Creature(g, me.ID, "My Knight", eldB1KnightType, 2, 2)
	b27Kill(g, venerable)
	g.RunStateChecksForTest()
	answerPickTargetIfOpen(t, g, knight)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, knight, game.CounterPlusOne); got != 1 {
		t.Errorf("the Knight has %d +1/+1 counters, want 1", got)
	}
}

func TestEld5516GarrisonGriffinGivesAKnightFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	griffin := b12Push(g, me.ID, "Garrison Griffin", "Creature — Griffin", eldB1GarrisonGriffin, 2, 2)
	knight := b12Creature(g, me.ID, "My Knight", eldB1KnightType, 2, 2)
	declareAttack(t, g, opp.ID, griffin)
	answerPickTargetIfOpen(t, g, knight)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, knight, "flying") {
		t.Error("the Knight has no flying")
	}
}

// --- statics -------------------------------------------------------------------

func TestEld5516InspiringVeteranPumpsOtherKnightsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	veteran := b12Push(g, me.ID, "Inspiring Veteran", eldB1KnightType, eldB1InspiringVeteran, 2, 2)
	mine := b12Creature(g, me.ID, "My Knight", eldB1KnightType, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", eldB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Knight", eldB1KnightType, 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		what string
	}{{mine, 3, "your other Knight"}, {veteran, 2, "the Veteran"}, {bear, 2, "your Bear"}, {theirs, 2, "an opponent's Knight"}} {
		if p, tough := ptOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s is %d/%d, want %d/%d", tc.what, p, tough, tc.want, tc.want)
		}
	}
}

func TestEld5516BelovedPrincessEvadesBigBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	princess := b12Push(g, me.ID, "Beloved Princess", "Creature — Human Noble", eldB1BelovedPrincess, 1, 1)
	big := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	small := b12Creature(g, opp.ID, "Bear", eldB1BearType, 2, 2)
	brAttack(t, g, princess)
	if err := g.DeclareBlocker(big, princess); err == nil {
		t.Error("a creature with power 3 blocked the Princess")
	}
	if err := g.DeclareBlocker(small, princess); err != nil {
		t.Errorf("a creature with power 2 can't block the Princess: %v", err)
	}
}

func TestEld5516FortifyingProvisionsMakesFoodAndToughens(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", eldB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", eldB1BearType, 2, 2)
	castCatalogSpell(t, g, "Fortifying Provisions", "Enchantment", eldB1FortifyingProvisions, nil)
	passPriorityAroundTable(t, g)
	if got := eldB1Food(g, me); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
	if _, tough := ptOf(t, g, mine); tough != 3 {
		t.Errorf("your creature has toughness %d, want 3", tough)
	}
	if _, tough := ptOf(t, g, theirs); tough != 2 {
		t.Errorf("an opponent's creature has toughness %d, want 2", tough)
	}
}

func TestEld5516CrystalSlipperGivesPowerAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	slipper := pushCatalogPermanent(g, me.ID, "Crystal Slipper", "Artifact — Equipment", eldB1CrystalSlipper, false)
	bear := b12Creature(g, me.ID, "Bear", eldB1BearType, 2, 2)
	apaMana(me, "C")
	equipTo(t, g, me.ID, slipper, bear)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear has no haste")
	}
}

// --- activated abilities ----------------------------------------------------

func TestEld5516SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		wantP, wantT           int
		keyword                string
	}{
		{"Jousting Dummy", "Artifact Creature — Scarecrow Knight", eldB1JoustingDummy, 2, 1, []string{"C", "C", "C"}, 3, 1, ""},
		{"Fireborn Knight", eldB1KnightType, eldB1FirebornKnight, 2, 3, []string{"R", "W", "R", "W"}, 3, 4, ""},
		{"Locthwain Gargoyle", "Artifact Creature — Gargoyle", eldB1LocthwainGargoyle, 0, 3, []string{"C", "C", "C", "C"}, 2, 3, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestEld5516ManaAbilities(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		pay                    []string
		color                  string
	}{
		{"Maraleaf Pixie", "Creature — Faerie", eldB1MaraleafPixie, nil, "U"},
		{"Signpost Scarecrow", "Artifact Creature — Scarecrow", eldB1SignpostScarecrow, []string{"C", "C"}, "R"},
		{"Spinning Wheel", "Artifact", eldB1SpinningWheel, nil, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			apaMana(me, tc.pay...)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{tc.color}}); err != nil {
				t.Fatalf("mana ability: %v", err)
			}
			if got := me.ManaPool.Total(); got != 1 {
				t.Errorf("pool holds %d mana, want 1", got)
			}
		})
	}
}

func TestEld5516SpinningWheelTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	wheel := pushCatalogPermanent(g, me.ID, "Spinning Wheel", "Artifact", eldB1SpinningWheel, false)
	bear := b12Creature(g, opp.ID, "Bear", eldB1BearType, 2, 2)
	apaMana(me, "C", "C", "C", "C", "C")
	apaActivate(t, g, me, wheel, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the creature is untapped")
	}
}

func TestEld5516ScaldingCauldronDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	cauldron := pushCatalogPermanent(g, me.ID, "Scalding Cauldron", "Artifact", eldB1ScaldingCauldron, false)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	apaMana(me, "C", "C", "C")
	apaActivate(t, g, me, cauldron, 0, game.ActivateAbilityParams{Targets: cardTarget(ogre)})
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
	if g.Battlefield.Contains(cauldron) {
		t.Error("the Cauldron was not sacrificed")
	}
}

func TestEld5516TemptingWitchMakesFoodOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Tempting Witch", "Creature — Human Warlock", eldB1TemptingWitch, 1, 3)
	passPriorityAroundTable(t, g)
	if got := eldB1Food(g, me); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

func TestEld5516TemptingWitchSacrificesAFoodForThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	witch := b12Push(g, me.ID, "Tempting Witch", "Creature — Human Warlock", eldB1TemptingWitch, 1, 3)
	var food uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, FoodToken(), 1, game.TokenEntryOptions{})
		if err == nil && len(ids) == 1 {
			food = ids[0]
		}
	})
	if food == uuid.Nil {
		t.Fatal("no Food token")
	}
	apaMana(me, "C", "C")
	life := opp.Life
	apaActivate(t, g, me, witch, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{food},
	})
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-3)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if !hmlTapped(t, g, witch) {
		t.Error("the Witch did not tap")
	}
}

// --- "whenever you draw your second card each turn" -------------------------------

func TestEld5516SecondDrawPumpsThisCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		wantP, wantT           int
		keyword                string
	}{
		{"Bloodhaze Wolverine", "Creature — Wolverine", eldB1BloodhazeWolverine, 2, 1, 3, 2, "first strike"},
		{"Steelgaze Griffin", "Creature — Griffin", eldB1SteelgazeGriffin, 2, 4, 4, 4, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			eldB1DrawToSecond(t, g, me)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestEld5516MadRatterMakesTwoRats(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	b12Push(g, me.ID, "Mad Ratter", "Creature — Goblin", eldB1MadRatter, 1, 2)
	eldB1DrawToSecond(t, g, me)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Rat", me.ID); got != 2 {
		t.Errorf("%d Rats, want 2", got)
	}
}

func TestEld5516IrencragPyromancerDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Irencrag Pyromancer", "Creature — Human Wizard", eldB1IrencragPyromancer, 0, 4)
	life := opp.Life
	eldB1DrawToSecond(t, g, me)
	answerPickTargetPlayerIfOpen(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-3)
	}
}

func TestEld5516ThunderousSnapperDrawsOnBigSpells(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	b12Push(g, me.ID, "Thunderous Snapper", "Creature — Turtle Hydra", eldB1ThunderousSnapper, 4, 4)
	for _, tc := range []struct {
		cost  string
		draws int
	}{{"{1}{G}", 0}, {"{4}{G}", 1}} {
		hand := me.Hand.Size()
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell " + tc.cost, TypeLine: "Sorcery", ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", tc.cost, err)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("a %s spell drew %d, want %d", tc.cost, got, tc.draws)
		}
	}
}

// --- instants and sorceries ------------------------------------------------------

func TestEld5516FellThePheasantHitsAFlyerAndMakesFood(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	walker := b12Creature(g, opp.ID, "Walker", eldB1BearType, 6, 6)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 6, Toughness: 6, Keywords: []string{"flying"}})
	if err := castCatalogSpellErr(t, g, "Fell the Pheasant", "Instant", eldB1FellThePheasant, cardTarget(walker)); err == nil {
		t.Fatal("Fell the Pheasant targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Fell the Pheasant", "Instant", eldB1FellThePheasant, cardTarget(flyer))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, flyer).DamageMarked; got != 5 {
		t.Errorf("the flyer has %d damage, want 5", got)
	}
	if got := eldB1Food(g, me); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

func TestEld5516LashOfThornsPumpsAndGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", eldB1BearType, 2, 2)
	castCatalogSpell(t, g, "Lash of Thorns", "Instant", eldB1LashOfThorns, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 4/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the Bear has no deathtouch")
	}
}

func TestEld5516TrueLovesKissExilesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", eldB1BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "True Love's Kiss", "Instant", eldB1TrueLovesKiss, cardTarget(bear)); err == nil {
		t.Fatal("True Love's Kiss targeted a creature")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "True Love's Kiss", "Instant", eldB1TrueLovesKiss, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestEld5516ForebodingFruitAdamantFood(t *testing.T) {
	for _, tc := range []struct {
		pool string
		food int
	}{{"BBB", 1}, {"BCC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			hand, life := opp.Hand.Size(), opp.Life
			eldB1CastPaid(t, g, "Foreboding Fruit", "Sorcery", "{2}{B}", eldB1ForebodingFruit, tc.pool, []game.TargetRef{playerTarget(opp.ID)})
			if got := opp.Hand.Size() - hand; got != 2 {
				t.Errorf("the target drew %d, want 2", got)
			}
			if opp.Life != life-2 {
				t.Errorf("the target is at %d, want %d", opp.Life, life-2)
			}
			if got := eldB1Food(g, me); got != tc.food {
				t.Errorf("%d Food, want %d", got, tc.food)
			}
		})
	}
}

func TestEld5516UnexplainedVisionAdamantScry(t *testing.T) {
	for _, tc := range []struct {
		pool string
		scry bool
	}{{"UUUCC", true}, {"UUCCC", false}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			hand := me.Hand.Size()
			eldB1CastPaid(t, g, "Unexplained Vision", "Sorcery", "{4}{U}", eldB1UnexplainedVision, tc.pool, nil)
			if got := me.Hand.Size() - hand; got != 3 {
				t.Errorf("drew %d, want 3", got)
			}
			c := scryChoiceFor(g, me.ID)
			if (c != nil) != tc.scry {
				t.Fatalf("scry prompt open = %v, want %v", c != nil, tc.scry)
			}
			if c != nil && len(c.ScryCards) != 3 {
				t.Errorf("scrying %d cards, want 3", len(c.ScryCards))
			}
		})
	}
}

func TestEld5516TurnIntoAPumpkinBouncesDrawsAndAdamantFood(t *testing.T) {
	for _, tc := range []struct {
		pool string
		food int
	}{{"UUUC", 1}, {"UCCC", 0}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			advanceToMain(t, g)
			bear := b12Creature(g, opp.ID, "Bear", eldB1BearType, 2, 2)
			hand := me.Hand.Size()
			eldB1CastPaid(t, g, "Turn into a Pumpkin", "Instant", "{3}{U}", eldB1TurnIntoAPumpkin, tc.pool, cardTarget(bear))
			if !opp.Hand.Contains(bear) {
				t.Error("the permanent did not return to its owner's hand")
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
			if got := eldB1Food(g, me); got != tc.food {
				t.Errorf("%d Food, want %d", got, tc.food)
			}
		})
	}
}

// --- Adventures -----------------------------------------------------------------------

// eldB1Adventure seats an Adventure card in the active player's hand at a
// main phase: a creature face 0 and an Adventure face 1.
func eldB1Adventure(t *testing.T, g *game.Game, oracle, creature, creatureType, adventure, adventureType string) (*game.Player, uuid.UUID) {
	t.Helper()
	return hobA6HandAdventure(t, g, oracle,
		hobA6Face(creature, creatureType, "{1}", 2, 2),
		hobA6Face(adventure, adventureType, "{1}", 0, 0))
}

func TestEld5516AdventuresThatMakeTokens(t *testing.T) {
	for _, tc := range []struct {
		creature, creatureType, adventure, oracle string
		token                                     string
		n                                         int
	}{
		{"Curious Pair", "Creature — Human Peasant", "Treats to Share", eldB1CuriousPair, "Food", 1},
		{"Lonesome Unicorn", "Creature — Unicorn", "Rider in Need", eldB1LonesomeUnicorn, "Knight", 1},
		{"Oakhame Ranger", "Creature — Elf Knight Ranger", "Bring Back", eldB1OakhameRanger, "Human", 2},
	} {
		t.Run(tc.adventure, func(t *testing.T) {
			g := newCatalogGame(t)
			me, id := eldB1Adventure(t, g, tc.oracle, tc.creature, tc.creatureType, tc.adventure, "Sorcery — Adventure")
			hobA6MustCast(t, g, me, id, 1, nil)
			if got := tokensNamedControlledBy(g, tc.token, me.ID); got != tc.n {
				t.Errorf("%d %s tokens, want %d", got, tc.token, tc.n)
			}
			hobA6WentOnAnAdventure(t, g, id)
		})
	}
}

func TestEld5516RiderInNeedKnightHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1LonesomeUnicorn, "Lonesome Unicorn", "Creature — Unicorn", "Rider in Need", "Sorcery — Adventure")
	hobA6MustCast(t, g, me, id, 1, nil)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && IsToken(c) && !hasEffectiveKeyword(t, g, c.InstanceID, "vigilance") {
			t.Error("the Knight token has no vigilance")
		}
	}
}

func TestEld5516BattleDisplayDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1EmberethShieldbreak, "Embereth Shieldbreaker", eldB1KnightType, "Battle Display", "Sorcery — Adventure")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", eldB1BearType, 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := hobA6Cast(g, me, id, 1, cardTarget(bear)); err == nil {
		t.Fatal("Battle Display targeted a creature")
	}
	hobA6MustCast(t, g, me, id, 1, cardTarget(rock))
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516AdventurePumps(t *testing.T) {
	for _, tc := range []struct {
		creature, creatureType, adventure, oracle string
		wantP, wantT                              int
		keyword                                   string
	}{
		{"Garenbrig Carver", "Creature — Human Warrior", "Shield's Might", eldB1GarenbrigCarver, 4, 4, ""},
		{"Rimrock Knight", "Creature — Dwarf Knight", "Boulder Rush", eldB1RimrockKnight, 4, 2, ""},
		{"Faerie Guidemother", "Creature — Faerie", "Gift of the Fae", eldB1FaerieGuidemother, 4, 3, "flying"},
	} {
		t.Run(tc.adventure, func(t *testing.T) {
			g := newCatalogGame(t)
			me, id := eldB1Adventure(t, g, tc.oracle, tc.creature, tc.creatureType, tc.adventure, "Instant — Adventure")
			bear := b12Creature(g, me.ID, "Bear", eldB1BearType, 2, 2)
			hobA6MustCast(t, g, me, id, 1, cardTarget(bear))
			if p, tough := ptOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
			hobA6WentOnAnAdventure(t, g, id)
		})
	}
}

func TestEld5516RimrockKnightCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", eldB1BearType, 2, 2)
	rimrock := b12Push(g, opp.ID, "Rimrock Knight", "Creature — Dwarf Knight", eldB1RimrockKnight, 3, 1)
	brAttack(t, g, attacker)
	if err := g.DeclareBlocker(rimrock, attacker); err == nil {
		t.Error("Rimrock Knight blocked")
	}
}

func TestEld5516VentureDeeperMillsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1MerfolkSecretkeeper, "Merfolk Secretkeeper", "Creature — Merfolk Wizard", "Venture Deeper", "Sorcery — Adventure")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	graveyard := opp.Graveyard.Size()
	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{playerTarget(opp.ID)})
	if got := opp.Graveyard.Size() - graveyard; got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516UsherToSafetyReturnsYourPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1ShepherdOfTheFlock, "Shepherd of the Flock", "Creature — Human Peasant", "Usher to Safety", "Instant — Adventure")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Theirs", eldB1BearType, 2, 2)
	mine := b12Creature(g, me.ID, "Mine", eldB1BearType, 2, 2)
	if err := hobA6Cast(g, me, id, 1, cardTarget(theirs)); err == nil {
		t.Fatal("Usher to Safety targeted an opponent's permanent")
	}
	hobA6MustCast(t, g, me, id, 1, cardTarget(mine))
	if !me.Hand.Contains(mine) {
		t.Error("your permanent did not return to your hand")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516OnAlertPumpsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1SilverflameSquire, "Silverflame Squire", "Creature — Human Soldier", "On Alert", "Instant — Adventure")
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: eldB1BearType, Power: 2, Toughness: 2, Tapped: true})
	hobA6MustCast(t, g, me, id, 1, cardTarget(bear))
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if hmlTapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516OakenBoonPutsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1TuinvaleTreefolk, "Tuinvale Treefolk", "Creature — Treefolk Druid", "Oaken Boon", "Sorcery — Adventure")
	bear := b12Creature(g, me.ID, "Bear", eldB1BearType, 2, 2)
	hobA6MustCast(t, g, me, id, 1, cardTarget(bear))
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516DizzyingSwoopTapsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1ArdenvaleTactician, "Ardenvale Tactician", eldB1KnightType, "Dizzying Swoop", "Instant — Adventure")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "A", eldB1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", eldB1BearType, 2, 2)
	hobA6MustCast(t, g, me, id, 1, cardTargets(a, b))
	for _, c := range []uuid.UUID{a, b} {
		if !hmlTapped(t, g, c) {
			t.Error("a target is untapped")
		}
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516ChopDownDestroysOnlyABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, id := eldB1Adventure(t, g, eldB1GiantKiller, "Giant Killer", "Creature — Human Peasant", "Chop Down", "Instant — Adventure")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Small", eldB1BearType, 3, 3)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	if err := hobA6Cast(g, me, id, 1, cardTarget(small)); err == nil {
		t.Fatal("Chop Down targeted a creature with power 3")
	}
	hobA6MustCast(t, g, me, id, 1, cardTarget(giant))
	if g.Battlefield.Contains(giant) {
		t.Error("the Giant survived")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestEld5516GiantKillerTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	killer := b12Push(g, me.ID, "Giant Killer", "Creature — Human Peasant", eldB1GiantKiller, 1, 2)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	apaMana(me, "C", "W")
	apaActivate(t, g, me, killer, 0, game.ActivateAbilityParams{Targets: cardTarget(giant)})
	if !hmlTapped(t, g, giant) {
		t.Error("the target is untapped")
	}
}

func TestEld5516OakhameRangerPumpsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	ranger := b12Push(g, me.ID, "Oakhame Ranger", "Creature — Elf Knight Ranger", eldB1OakhameRanger, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", eldB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", eldB1BearType, 2, 2)
	apaActivate(t, g, me, ranger, 0, game.ActivateAbilityParams{})
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{ranger, 3}, {bear, 3}, {theirs, 2}} {
		if p, tough := ptOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("a creature is %d/%d, want %d/%d", p, tough, tc.want, tc.want)
		}
	}
}
