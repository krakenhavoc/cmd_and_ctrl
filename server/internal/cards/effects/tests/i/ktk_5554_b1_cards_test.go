package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ktk_5554_b1_cards_test.go — Khans of Tarkir, batch 1 (#5554): five
// vanilla bodies, nine creatures whose whole text is keywords, fifteen
// morphs, five outlast creatures and sixteen cards whose text a
// catalogued card already prints. Each test plays the real catalog entry
// through the engine.

const (
	ktkB1AlpineGrizzlyOracle    = "86e3c86c-d50e-4aa9-b080-f68b6afee87c"
	ktkB1RottingMastodonOracle  = "d6e90e75-ef12-49d0-8dd8-0c6cb718ce28"
	ktkB1SummitProwlerOracle    = "c63a9a3d-13df-4a94-a8e1-18071c351a8d"
	ktkB1TuskedColossodonOracle = "bba11c0e-01a3-45ad-8de4-2148cf660ba8"
	ktkB1WetlandSambarOracle    = "8a50b8f7-1af8-4e00-9734-66a87e05cc6f"

	ktkB1AlabasterKirinOracle      = "640fa591-ba79-4148-86a6-9de994a9e67c"
	ktkB1HighspireMantisOracle     = "6adff11d-5126-4a4b-b182-e2700d406995"
	ktkB1MantisRiderOracle         = "5c1a0a7a-dea4-4b8e-b9c0-8c19942a01ef"
	ktkB1VenerableLammasuOracle    = "359c6272-fcb5-4d3f-a4b5-4209e354a044"
	ktkB1DragonStyleTwinsOracle    = "d150f456-c33a-4b19-ac73-b1d818d5db8f"
	ktkB1JeskaiStudentOracle       = "c8ebfc8b-3cee-491b-a568-8a8dcb22541d"
	ktkB1JeskaiWindscoutOracle     = "c35f4b80-90b6-467a-a666-1cc3437d49ab"
	ktkB1RiverwheelAerialistOracle = "f9221367-24a5-41a3-8943-c6a3cb5d605d"
	ktkB1WhirlwindAdeptOracle      = "0c7e261f-45d5-4c86-8cba-9b6edce8ef49"

	ktkB1CanyonLurkersOracle    = "a744c5d7-c903-446a-bb30-30788058d9db"
	ktkB1GlacialStalkerOracle   = "49d06291-c8cb-4e22-b178-d0598582b2ee"
	ktkB1KrumarBondKinOracle    = "8e34f487-aa75-4afb-af75-3349de456c79"
	ktkB1WarBehemothOracle      = "4792a05e-b557-4132-8303-235feefe0d99"
	ktkB1WitnessOfTheAgesOracle = "afeb9de8-7030-4797-a9c8-5a99de3c410f"
	ktkB1WoollyLoxodonOracle    = "b588df7a-64d8-449e-b573-13badc8cdf56"
	ktkB1AbzanGuideOracle       = "85b2fb89-5ebe-4c13-bc12-38d1417e5880"
	ktkB1SageEyeHarrierOracle   = "589d1c6e-5808-424c-a6e4-dadef90d119e"
	ktkB1SaguArcherOracle       = "cf8a8482-e243-4155-bc21-a1e474c6f1e7"
	ktkB1SaguMaulerOracle       = "8d877ee9-1751-48a1-9f2f-02897dae153b"
	ktkB1SidisisPetOracle       = "9b183d76-1dee-4ecf-9226-e1aa600b1247"
	ktkB1SnowhornRiderOracle    = "3d5a753b-ed2d-4646-ade3-5daeba2a7270"
	ktkB1MonasteryFlockOracle   = "fd05f23e-8974-445d-9059-d706a3740d89"
	ktkB1MysticHiddenWayOracle  = "530853b4-68a0-4814-b0aa-ffeb5585cd43"
	ktkB1KinTreeWardenOracle    = "1ea9517a-c2ed-4aec-8fd3-e603b0e7dec9"

	ktkB1DisownedAncestorOracle = "192603a4-720e-4c53-bee5-4f2826c1ed7f"
	ktkB1SaltRoadPatrolOracle   = "b101943b-23f1-450f-a722-13761dc33419"
	ktkB1MerEkNightbladeOracle  = "56a8554c-d427-44ca-a01d-9e12c89acd1e"
	ktkB1TuskguardCaptainOracle = "aa149597-087a-4e1f-953c-2cd04e90a93b"
	ktkB1LongshotSquadOracle    = "c153e202-1d93-4044-b457-b8ebd5b36ec3"

	ktkB1MarduHatebladeOracle      = "acff2573-cf8a-4aee-96a3-28721a4128f6"
	ktkB1RakshasaDeathdealerOracle = "acdc17cf-8f8c-42ee-bede-481ce00cf4bf"
	ktkB1ChiefOfTheEdgeOracle      = "636a9364-54a6-4f67-ad96-731646e8bde7"
	ktkB1ChiefOfTheScaleOracle     = "32e4de02-8437-4ffc-acc9-a3736652d282"
	ktkB1ValleyDasherOracle        = "aad16b50-a866-4871-843f-ec8813ad5dfd"
	ktkB1HighlandGameOracle        = "0e0d40a6-e0a4-4fb2-be72-314f7d66b7eb"
	ktkB1SiegeRhinoOracle          = "e1087ce2-e3c2-4ab1-bd08-1dc9ab70f12f"
	ktkB1BearsCompanionOracle      = "930248b3-44e6-4c0b-bd7d-4b48323a7272"

	ktkB1WeaveFateOracle          = "ff568a29-e31a-4bc3-b8a3-8fa56f91d52d"
	ktkB1TakeUpArmsOracle         = "61e2d557-f08d-489b-94c8-aff87e5189f9"
	ktkB1ThrottleOracle           = "01823ab7-b9f6-48bc-b45c-9b0fabc6f657"
	ktkB1WaterwhirlOracle         = "77636154-962f-4ca8-85e7-c0c9c0c3cd2d"
	ktkB1AwakenTheBearOracle      = "53dc9795-d26a-4821-afad-8d77d279ffe3"
	ktkB1DutifulReturnOracle      = "c8f958da-3a31-4064-9a78-66e9b65b85ab"
	ktkB1DebilitatingInjuryOracle = "52eab77d-9a07-4e14-8872-72681d3b3d0e"
	ktkB1SiegecraftOracle         = "a49a9418-72d1-4ab1-89d9-a1f65ae23d35"

	ktkB1BearType    = "Creature — Bear"
	ktkB1WarriorType = "Creature — Human Warrior"
	ktkB1AuraType    = "Enchantment — Aura"
)

// ktkB1Vanilla are the batch's bodies with no rules text.
var ktkB1Vanilla = map[string]string{
	ktkB1AlpineGrizzlyOracle:    "Alpine Grizzly",
	ktkB1RottingMastodonOracle:  "Rotting Mastodon",
	ktkB1SummitProwlerOracle:    "Summit Prowler",
	ktkB1TuskedColossodonOracle: "Tusked Colossodon",
	ktkB1WetlandSambarOracle:    "Wetland Sambar",
}

// ktkB1Keyworded are the creatures whose whole text is printed keywords.
var ktkB1Keyworded = map[string]struct {
	name, typeLine string
	power, tough   int
	keywords       []string
}{
	ktkB1AlabasterKirinOracle:      {"Alabaster Kirin", "Creature — Kirin", 2, 3, []string{"flying", "vigilance"}},
	ktkB1HighspireMantisOracle:     {"Highspire Mantis", "Creature — Insect", 3, 3, []string{"flying", "trample"}},
	ktkB1MantisRiderOracle:         {"Mantis Rider", "Creature — Human Monk", 3, 3, []string{"flying", "vigilance", "haste"}},
	ktkB1VenerableLammasuOracle:    {"Venerable Lammasu", "Creature — Lammasu", 5, 4, []string{"flying"}},
	ktkB1DragonStyleTwinsOracle:    {"Dragon-Style Twins", "Creature — Human Monk", 3, 3, []string{"double strike", "prowess"}},
	ktkB1JeskaiStudentOracle:       {"Jeskai Student", "Creature — Human Monk", 1, 3, []string{"prowess"}},
	ktkB1JeskaiWindscoutOracle:     {"Jeskai Windscout", "Creature — Bird Scout", 2, 1, []string{"flying", "prowess"}},
	ktkB1RiverwheelAerialistOracle: {"Riverwheel Aerialists", "Creature — Djinn Monk", 4, 5, []string{"flying", "prowess"}},
	ktkB1WhirlwindAdeptOracle:      {"Whirlwind Adept", "Creature — Djinn Monk", 4, 2, []string{"hexproof", "prowess"}},
}

// ktkB1Morph is one morph creature: its printed body, the keywords it has
// only face up, its morph cost and the mana that pays it.
type ktkB1Morph struct {
	name, typeLine, oracle, faceUp string
	mana                           []string
	power, tough                   int
	keywords                       []string
}

var ktkB1Morphs = []ktkB1Morph{
	{"Canyon Lurkers", "Creature — Human Rogue", ktkB1CanyonLurkersOracle, "{3}{R}", []string{"C", "C", "C", "R"}, 5, 2, nil},
	{"Glacial Stalker", "Creature — Elemental", ktkB1GlacialStalkerOracle, "{4}{U}", []string{"C", "C", "C", "C", "U"}, 4, 5, nil},
	{"Krumar Bond-Kin", "Creature — Orc Warrior", ktkB1KrumarBondKinOracle, "{4}{B}", []string{"C", "C", "C", "C", "B"}, 5, 3, nil},
	{"War Behemoth", "Creature — Beast", ktkB1WarBehemothOracle, "{4}{W}", []string{"C", "C", "C", "C", "W"}, 3, 6, nil},
	{"Witness of the Ages", "Artifact Creature — Golem", ktkB1WitnessOfTheAgesOracle, "{5}", []string{"C", "C", "C", "C", "C"}, 4, 4, nil},
	{"Woolly Loxodon", "Creature — Elephant Warrior", ktkB1WoollyLoxodonOracle, "{5}{G}", []string{"C", "C", "C", "C", "C", "G"}, 6, 7, nil},
	{"Abzan Guide", ktkB1WarriorType, ktkB1AbzanGuideOracle, "{2}{W}{B}{G}", []string{"C", "C", "W", "B", "G"}, 4, 4, []string{"lifelink"}},
	{"Sage-Eye Harrier", "Creature — Bird Warrior", ktkB1SageEyeHarrierOracle, "{3}{W}", []string{"C", "C", "C", "W"}, 1, 5, []string{"flying"}},
	{"Sagu Archer", "Creature — Snake Archer", ktkB1SaguArcherOracle, "{4}{G}", []string{"C", "C", "C", "C", "G"}, 2, 5, []string{"reach"}},
	{"Sagu Mauler", "Creature — Beast", ktkB1SaguMaulerOracle, "{3}{G}{U}", []string{"C", "C", "C", "G", "U"}, 6, 6, []string{"trample", "hexproof"}},
	{"Sidisi's Pet", "Creature — Zombie Ape", ktkB1SidisisPetOracle, "{1}{B}", []string{"C", "B"}, 1, 4, []string{"lifelink"}},
	{"Snowhorn Rider", ktkB1WarriorType, ktkB1SnowhornRiderOracle, "{2}{G}{U}{R}", []string{"C", "C", "G", "U", "R"}, 5, 5, []string{"trample"}},
	{"Monastery Flock", "Creature — Bird", ktkB1MonasteryFlockOracle, "{U}", []string{"U"}, 0, 5, []string{"defender", "flying"}},
	{"Mystic of the Hidden Way", "Creature — Human Monk", ktkB1MysticHiddenWayOracle, "{2}{U}", []string{"C", "C", "U"}, 3, 2, nil},
	{"Kin-Tree Warden", ktkB1WarriorType, ktkB1KinTreeWardenOracle, "{G}", []string{"G"}, 1, 1, nil},
}

// ktkB1Outlast are the outlast creatures and the keyword each grants to
// creatures you control with a +1/+1 counter ("" for none).
var ktkB1Outlast = []struct {
	name, typeLine, oracle, keyword string
	power, tough                    int
}{
	{"Disowned Ancestor", "Creature — Spirit Warrior", ktkB1DisownedAncestorOracle, "", 0, 4},
	{"Salt Road Patrol", "Creature — Human Scout", ktkB1SaltRoadPatrolOracle, "", 2, 5},
	{"Mer-Ek Nightblade", "Creature — Orc Assassin", ktkB1MerEkNightbladeOracle, "deathtouch", 2, 3},
	{"Tuskguard Captain", ktkB1WarriorType, ktkB1TuskguardCaptainOracle, "trample", 2, 3},
	{"Longshot Squad", "Creature — Dog Archer", ktkB1LongshotSquadOracle, "reach", 3, 3},
}

func TestKtkB1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		ktkB1MarduHatebladeOracle:      "Mardu Hateblade",
		ktkB1RakshasaDeathdealerOracle: "Rakshasa Deathdealer",
		ktkB1ChiefOfTheEdgeOracle:      "Chief of the Edge",
		ktkB1ChiefOfTheScaleOracle:     "Chief of the Scale",
		ktkB1ValleyDasherOracle:        "Valley Dasher",
		ktkB1HighlandGameOracle:        "Highland Game",
		ktkB1SiegeRhinoOracle:          "Siege Rhino",
		ktkB1BearsCompanionOracle:      "Bear's Companion",
		ktkB1WeaveFateOracle:           "Weave Fate",
		ktkB1TakeUpArmsOracle:          "Take Up Arms",
		ktkB1ThrottleOracle:            "Throttle",
		ktkB1WaterwhirlOracle:          "Waterwhirl",
		ktkB1AwakenTheBearOracle:       "Awaken the Bear",
		ktkB1DutifulReturnOracle:       "Dutiful Return",
		ktkB1DebilitatingInjuryOracle:  "Debilitating Injury",
		ktkB1SiegecraftOracle:          "Siegecraft",
	}
	for oracle, name := range ktkB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range ktkB1Keyworded {
		all[oracle] = k.name
	}
	for _, m := range ktkB1Morphs {
		all[m.oracle] = m.name
	}
	for _, o := range ktkB1Outlast {
		all[o.oracle] = o.name
	}
	if len(all) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(all))
	}
	for oracle, name := range all {
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

// --- bodies and keywords -----------------------------------------------------

func TestKtkB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range ktkB1Vanilla {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 || len(spec.AlternativeCosts) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestKtkB1AlpineGrizzlyIsAPlainCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grizzly := effectstest.B12Push(g, me.ID, "Alpine Grizzly", ktkB1BearType, ktkB1AlpineGrizzlyOracle, 4, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", ktkB1BearType, 2, 2)
	if err := effectstest.PorBlock(t, g, grizzly, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Alpine Grizzly couldn't be blocked: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, grizzly); p != 4 || tough != 2 {
		t.Errorf("Alpine Grizzly is %d/%d, want 4/2", p, tough)
	}
}

func TestKtkB1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range ktkB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := effects.Lookup(oracle)
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, k.power, k.tough)
			for _, kw := range k.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Prowess (CR 702.108a): casting a noncreature spell gives each of them
// +1/+1 until end of turn.
func TestKtkB1ProwessCreaturesGrowOnANoncreatureSpell(t *testing.T) {
	for _, oracle := range []string{
		ktkB1DragonStyleTwinsOracle, ktkB1JeskaiStudentOracle, ktkB1JeskaiWindscoutOracle,
		ktkB1RiverwheelAerialistOracle, ktkB1WhirlwindAdeptOracle,
	} {
		k := ktkB1Keyworded[oracle]
		t.Run(k.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, k.name, k.typeLine, oracle, k.power, k.tough)
			effectstest.AdvanceToMain(t, g)
			effectstest.CastCatalogSpell(t, g, "Weave Fate", "Instant", ktkB1WeaveFateOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != k.power+1 || tough != k.tough+1 {
				t.Errorf("%s is %d/%d after a noncreature spell, want %d/%d", k.name, p, tough, k.power+1, k.tough+1)
			}
		})
	}
}

// Hexproof (CR 702.11b): an opponent's Throttle can't target Whirlwind
// Adept or a face-up Sagu Mauler, but can target their Bear.
func TestKtkB1HexproofRefusesAnOpponentsSpell(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Whirlwind Adept", "Creature — Djinn Monk", ktkB1WhirlwindAdeptOracle},
		{"Sagu Mauler", "Creature — Beast", ktkB1SaguMaulerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			bear := effectstest.B12Creature(g, me.ID, "Bear", ktkB1BearType, 2, 2)
			throttle := effectstest.PushCatalogHandCard(opp, "Throttle", "Instant", ktkB1ThrottleOracle)
			if err := g.CastSpell(opp.ID, throttle, game.CastSpellParams{Targets: effectstest.CardTarget(id)}); err == nil {
				t.Fatalf("an opponent's spell targeted %s", tc.name)
			}
			if err := g.CastSpell(opp.ID, throttle, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
				t.Fatalf("the opponent couldn't target a creature without hexproof: %v", err)
			}
		})
	}
}

// Haste: Mantis Rider attacks the turn it arrives (CR 702.10b), and
// vigilance leaves it untapped (CR 702.20b).
func TestKtkB1MantisRiderAttacksAtOnceWithoutTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := effectstest.CastHoldCreature(t, g, "Mantis Rider", "Creature — Human Monk", ktkB1MantisRiderOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	if effectstest.AlliTapped(t, g, id) {
		t.Error("Mantis Rider tapped to attack")
	}
}

// --- morph -------------------------------------------------------------------

// Every morph casts face down for the keyword's {3} and turns face up for
// the card's own printed cost (CR 702.37a).
func TestKtkB1MorphCostsAreThePrintedOnes(t *testing.T) {
	for _, m := range ktkB1Morphs {
		ac := game.AlternativeCostByKey(m.oracle, "morph")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no morph cast", m.name)
			continue
		}
		if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownMorphed || ac.FaceDown.FaceUpCounter {
			t.Errorf("%s morph = %q %q counter=%v", m.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCounter)
		}
		if ac.FaceDown.FaceUpCost != m.faceUp {
			t.Errorf("%s turns face up for %q, want %q", m.name, ac.FaceDown.FaceUpCost, m.faceUp)
		}
	}
}

// Face down each is a 2/2 with no keyword (CR 708.2a); its morph cost turns
// it into the printed creature, keywords and all (CR 708.8).
func TestKtkB1MorphsTurnFaceUpIntoThePrintedCreature(t *testing.T) {
	for _, m := range ktkB1Morphs {
		t.Run(m.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.S36CastFaceDown(t, g, m.name, m.typeLine, m.oracle, "morph", m.power, m.tough)
			if p, tough := effectstest.MkmMPPT(t, g, id); p != 2 || tough != 2 {
				t.Errorf("face down it is %d/%d, want 2/2", p, tough)
			}
			for _, kw := range m.keywords {
				if effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
					t.Errorf("face down it has %s (CR 708.2a)", kw)
				}
			}
			effectstest.S36TurnUp(t, g, me, id, m.mana...)
			if p, tough := effectstest.MkmMPPT(t, g, id); p != m.power || tough != m.tough {
				t.Errorf("face up it is %d/%d, want %d/%d", p, tough, m.power, m.tough)
			}
			for _, kw := range m.keywords {
				if !effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
					t.Errorf("face up it has no %s", kw)
				}
			}
		})
	}
}

func TestKtkB1MysticOfTheHiddenWayCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Mystic of the Hidden Way", "Creature — Human Monk", ktkB1MysticHiddenWayOracle, 3, 2)
	effectstest.AssertRestrictions(t, g, id, game.CantBeBlocked)
}

// Defender: Monastery Flock can't attack (CR 702.3b).
func TestKtkB1MonasteryFlockCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flock := effectstest.B12Push(g, me.ID, "Monastery Flock", "Creature — Bird", ktkB1MonasteryFlockOracle, 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(flock, opp.ID); err == nil {
		t.Error("Monastery Flock attacked")
	}
}

// ktkB1ShieldThenDestroy activates row `row` of `id` for `mana`, then
// destroys it, and reports whether it is still on the battlefield.
func ktkB1ShieldThenDestroy(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, row int, mana ...string) bool {
	t.Helper()
	effectstest.ApaMana(me, mana...)
	effectstest.ApaActivate(t, g, me, id, row, game.ActivateAbilityParams{})
	effectstest.B27Kill(g, id)
	return g.Battlefield.Contains(id)
}

func TestKtkB1KinTreeWardenRegeneratesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	warden := effectstest.B12Push(g, me.ID, "Kin-Tree Warden", ktkB1WarriorType, ktkB1KinTreeWardenOracle, 1, 1)
	if !ktkB1ShieldThenDestroy(t, g, me, warden, 0, "C", "C") {
		t.Error("the shielded Warden was destroyed")
	}
}

// --- outlast -------------------------------------------------------------------

func TestKtkB1OutlastPutsACounterAndGrantsTheKeyword(t *testing.T) {
	for _, c := range ktkB1Outlast {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			self := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle,
				Power: c.power, Toughness: c.tough, Owner: me.ID, Controller: me.ID})
			bear := effectstest.B12Creature(g, me.ID, "Bear", ktkB1BearType, 2, 2)
			theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", ktkB1BearType, 2, 2)
			if err := effectstest.RprB04ActivateOutlast(t, g, me, self); err != nil {
				t.Fatalf("outlast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.BattlefieldPlusOnes(g, self); n != 1 {
				t.Fatalf("%d counters after outlast, want 1", n)
			}
			if !effectstest.FindBattlefieldCardForTest(g, self).Tapped {
				t.Error("outlast did not tap the creature")
			}
			if p, tough := effectstest.PtOf(t, g, self); p != c.power+1 || tough != c.tough+1 {
				t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.power+1, c.tough+1)
			}
			if c.keyword == "" {
				return
			}
			if !effectstest.HasEffectiveKeyword(t, g, self, c.keyword) {
				t.Errorf("%s with a counter has no %s", c.name, c.keyword)
			}
			if effectstest.HasEffectiveKeyword(t, g, bear, c.keyword) {
				t.Errorf("a creature with no counter has %s", c.keyword)
			}
			effectstest.AddCountersForTest(g, bear, game.CounterPlusOne, 1)
			if !effectstest.HasEffectiveKeyword(t, g, bear, c.keyword) {
				t.Errorf("your creature with a counter has no %s", c.keyword)
			}
			effectstest.AddCountersForTest(g, theirs, game.CounterPlusOne, 1)
			if effectstest.HasEffectiveKeyword(t, g, theirs, c.keyword) {
				t.Errorf("an opponent's creature with a counter has %s", c.keyword)
			}
		})
	}
}

func TestKtkB1OutlastIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	self := effectstest.PushCatalogPermanent(g, me.ID, "Salt Road Patrol", "Creature — Human Scout", ktkB1SaltRoadPatrolOracle, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, self, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("outlast was activated outside a main phase")
	}
}

// --- activated abilities --------------------------------------------------------

func TestKtkB1MarduHatebladeGainsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	blade := effectstest.B12Push(g, me.ID, "Mardu Hateblade", ktkB1WarriorType, ktkB1MarduHatebladeOracle, 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, blade, "deathtouch") {
		t.Fatal("Mardu Hateblade has deathtouch before paying")
	}
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, blade, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, blade, "deathtouch") {
		t.Error("Mardu Hateblade has no deathtouch after paying {B}")
	}
}

func TestKtkB1RakshasaDeathdealerPumpsAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	demon := effectstest.B12Push(g, me.ID, "Rakshasa Deathdealer", "Creature — Demon", ktkB1RakshasaDeathdealerOracle, 2, 2)
	effectstest.ApaMana(me, "B", "G")
	effectstest.ApaActivate(t, g, me, demon, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, demon); p != 4 || tough != 4 {
		t.Errorf("the Deathdealer is %d/%d, want 4/4", p, tough)
	}
	if !ktkB1ShieldThenDestroy(t, g, me, demon, 1, "B", "G") {
		t.Error("the shielded Deathdealer was destroyed")
	}
}

// --- statics -------------------------------------------------------------------

func TestKtkB1ChiefsPumpOnlyYourOtherWarriors(t *testing.T) {
	for _, tc := range []struct {
		name, oracle         string
		power, tough, dp, dt int
	}{
		{"Chief of the Edge", ktkB1ChiefOfTheEdgeOracle, 3, 2, 1, 0},
		{"Chief of the Scale", ktkB1ChiefOfTheScaleOracle, 2, 3, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			chief := effectstest.B12Push(g, me.ID, tc.name, ktkB1WarriorType, tc.oracle, tc.power, tc.tough)
			warrior := effectstest.B12Creature(g, me.ID, "Warrior", ktkB1WarriorType, 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", ktkB1BearType, 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", ktkB1WarriorType, 2, 2)
			if p, tough := effectstest.PtOf(t, g, warrior); p != 2+tc.dp || tough != 2+tc.dt {
				t.Errorf("your Warrior is %d/%d, want %d/%d", p, tough, 2+tc.dp, 2+tc.dt)
			}
			if p, tough := effectstest.PtOf(t, g, chief); p != tc.power || tough != tc.tough {
				t.Errorf("the Chief pumped itself: %d/%d", p, tough)
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("a non-Warrior is %d/%d, want 2/2", p, tough)
			}
			if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
				t.Errorf("an opponent's Warrior is %d/%d, want 2/2", p, tough)
			}
		})
	}
}

func TestKtkB1ValleyDasherMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	dasher := effectstest.B12Push(g, me.ID, "Valley Dasher", "Creature — Human Berserker", ktkB1ValleyDasherOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, dasher, "haste") {
		t.Error("Valley Dasher has no haste")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != dasher {
		t.Fatalf("pass with the Dasher home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(dasher, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Dasher attacking: %v", err)
	}
}

// --- triggers ------------------------------------------------------------------

func TestKtkB1HighlandGameGainsTwoWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elk := effectstest.B12Push(g, me.ID, "Highland Game", "Creature — Elk", ktkB1HighlandGameOracle, 2, 1)
	life := me.Life
	effectstest.B27Kill(g, elk)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestKtkB1SiegeRhinoDrainsEachOpponentForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	rhino := effectstest.CastHoldCreature(t, g, "Siege Rhino", "Creature — Rhino", ktkB1SiegeRhinoOracle, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 3
		if s.ID == me.ID {
			want = lives[s.ID] + 3
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, rhino, "trample") {
		t.Error("Siege Rhino has no trample")
	}
}

func TestKtkB1BearsCompanionMakesAFourFourBear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Bear's Companion", ktkB1WarriorType, ktkB1BearsCompanionOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	n, bear := effectstest.B14Tokens(g, me.ID, "Bear")
	if n != 1 || bear.Power != 4 || bear.Toughness != 4 || !bear.HasColor("G") {
		t.Fatalf("want one 4/4 green Bear: %d, %+v", n, bear)
	}
}

// --- spells --------------------------------------------------------------------

func TestKtkB1WeaveFateDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Weave Fate", "Instant", ktkB1WeaveFateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestKtkB1TakeUpArmsMakesThreeWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastCatalogSpell(t, g, "Take Up Arms", "Instant", ktkB1TakeUpArmsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n, w := effectstest.B14Tokens(g, me.ID, "Warrior")
	if n != 3 || w.Power != 1 || w.Toughness != 1 || !w.HasColor("W") {
		t.Fatalf("want three 1/1 white Warriors: %d, %+v", n, w)
	}
}

func TestKtkB1ThrottleKillsAFourToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Throttle", "Instant", ktkB1ThrottleOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 4/4 survived -4/-4")
	}
}

func TestKtkB1WaterwhirlBouncesTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", ktkB1BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", ktkB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Waterwhirl", "Instant", ktkB1WaterwhirlOracle, effectstest.CardRefs(theirs, mine))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) || !me.Hand.Contains(mine) {
		t.Error("the two creatures did not return to their owners' hands")
	}
}

func TestKtkB1AwakenTheBearPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", ktkB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Awaken the Bear", "Instant", ktkB1AwakenTheBearOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

func TestKtkB1DutifulReturnReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	one := effectstest.PushGraveyardCardTyped(me, "One", ktkB1BearType)
	two := effectstest.PushGraveyardCardTyped(me, "Two", ktkB1BearType)
	rock := effectstest.PushGraveyardCardTyped(me, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Dutiful Return", "Sorcery", ktkB1DutifulReturnOracle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("Dutiful Return targeted an artifact card")
	}
	effectstest.CastCatalogSpell(t, g, "Dutiful Return", "Sorcery", ktkB1DutifulReturnOracle, effectstest.CardRefs(one, two))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Error("the creature cards did not return to hand")
	}
}

func TestKtkB1AurasChangeTheEnchantedCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Debilitating Injury", ktkB1DebilitatingInjuryOracle, 1, 2},
		{"Siegecraft", ktkB1SiegecraftOracle, 5, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			host := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 4)
			effectstest.CastCatalogSpell(t, g, tc.name, ktkB1AuraType, tc.oracle, effectstest.CardTarget(host))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, host); p != tc.power || tough != tc.tough {
				t.Errorf("the enchanted creature is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
		})
	}
}
