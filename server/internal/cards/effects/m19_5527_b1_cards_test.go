package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m19_5527_b1_cards_test.go — Core Set 2019, batch 1 (#5527): three
// vanilla bodies, nine creatures whose whole text is keywords, and
// thirty-eight cards whose text an existing catalog card already prints
// (or that compose its helpers). Each test plays the real catalog entry
// through the engine and asserts what its text says.

const (
	m19B1BogstomperOracle      = "7c2f1915-8d91-47c6-bab4-be325348673a"
	m19B1LoxodonOracle         = "57f0f55d-4cf9-4a35-96dd-05c24c2a9a4f"
	m19B1OnakkeOgreOracle      = "fa10cacb-ab14-447a-b411-9d74bc3772fb"
	m19B1HavocDevilsOracle     = "be1c2607-091a-41e7-ae4c-73c7e5689739"
	m19B1HiredBladeOracle      = "fed32c9c-7458-4143-b0b2-45ca23ba25ee"
	m19B1HostileMinotaurOracle = "e050a0a3-f48e-4cba-9cd4-692a92261c9f"
	m19B1KnightOfTuskOracle    = "524d81fb-33f5-4e77-ae41-477b91e67be4"
	m19B1VigilantBalothOracle  = "caf1bc69-a8d7-454b-9cc6-2e4d003776a8"
	m19B1RustwingFalconOracle  = "438ee544-78a2-4e32-9343-70838b60ac16"
	m19B1SilverbeakOracle      = "e9f410c1-d93d-4291-95fc-513832fb36e0"
	m19B1TwoHeadedZombieOracle = "7e280770-9a8f-4445-84d2-3b9cac659136"
	m19B1WallOfMistOracle      = "b2ca8824-6bd8-4aa3-b156-618118eb98a4"

	m19B1GhastbarkOracle       = "bdd0d01d-b9f2-4c3f-9b3b-04e122013de5"
	m19B1SerrasGuardianOracle  = "c50cd7e7-8d53-44af-ac96-3e104587bdb3"
	m19B1MaraudersAxeOracle    = "5d0e49fa-5dfb-48c2-af97-fdfb788d5f40"
	m19B1CatalystOracle        = "7779525d-080d-4669-bb65-af532bf0983e"
	m19B1GoblinMotivatorOracle = "df67597f-42a8-4fa7-b426-d7db8b384445"
	m19B1RootSnareOracle       = "cbf743a5-123f-4747-826d-7f8bb929a52c"
	m19B1TakeVengeanceOracle   = "c60b8edd-aa3f-48ea-b202-d10599f91d9d"
	m19B1CageOracle            = "207c7f93-3abf-44c5-ace6-3f86359c9745"
	m19B1AviationOracle        = "ad11ff81-acaa-4529-ba15-718dadbea259"
	m19B1GallantCavalryOracle  = "b500c460-2814-47a4-b307-17ca519f3565"
	m19B1RhoxOracleOracle      = "d26d1cce-3bcf-48d4-abce-8b12ca7b7432"
	m19B1SalvagerOracle        = "3e400e13-6cae-45e6-b5cc-5b5f9d74dec5"
	m19B1StranglingOracle      = "444493a6-0f90-4847-9619-4ea8fbf7cf6f"
	m19B1AegisOracle           = "f8e1b38a-2a77-4cf1-8bef-31acb61ede10"
	m19B1FieryFinishOracle     = "afa3eb61-624c-4abe-bf1e-58fa3b4c14a6"
	m19B1LichsCaressOracle     = "01b3af6d-ab7d-4137-a94a-1cf2cbfd9d44"
	m19B1SovereignsBiteOracle  = "13718d93-fe46-4a18-a511-88649ef1c9de"
	m19B1VampireSovOracle      = "49c7b7c6-c8dc-4871-9103-b057d2034c3e"
	m19B1PlagueMareOracle      = "4cd096ac-dd67-4208-a3ad-8e6061b442c6"
	m19B1VineMareOracle        = "499c7ee3-08a3-4f40-809a-3f7087ab9b24"
	m19B1LightningMareOracle   = "7c7948c6-5144-4df0-8e33-117710ae00af"
	m19B1BristlingBoarOracle   = "f4aed3d2-04f5-49a4-a6be-d3cf097903f8"
	m19B1EncyclopediaOracle    = "d91ea728-2d69-42cb-bb5d-e2b058f0d7b1"
	m19B1FellSpecterOracle     = "d655a1c9-b84d-4069-a2ca-3aa0dacab3e1"
	m19B1FountainOracle        = "df024c55-c008-48ec-a1a5-02ce336e3de6"
	m19B1KnightsPledgeOracle   = "63604320-68de-415e-b1f9-dd0f85c5d8a3"
	m19B1GearsmithProdOracle   = "589d6b67-17ad-4ca7-9b0c-1b919bf03a27"
	m19B1AerialEngineerOracle  = "27f5969e-d613-4cbc-87af-98899bb5c58a"
	m19B1GearsmithGuardOracle  = "0c0f7092-3fe6-4a6e-8b4e-803810b43e50"
	m19B1RavenousHarpyOracle   = "e9ccec56-ec2b-4ab8-b929-6e4795711db9"
	m19B1SiegebreakerOracle    = "fcd281e3-f336-449f-a887-4b98dc14dfe7"
	m19B1SkallaWolfOracle      = "22da0659-7cd0-4277-b6d2-7d954dcd7729"
	m19B1MilitiaBuglerOracle   = "edb6ae9e-a41c-4d82-97bf-4be00e98364c"
	m19B1DrillmasterOracle     = "d89bbfe7-e7ef-4f1c-a6b5-6d8ef4079daa"
	m19B1AngelOfDawnOracle     = "f63ea0e5-7769-4a22-bbb3-658281141d8c"
	m19B1DwarvenPriestOracle   = "6a077a9b-d725-4b14-957f-1586db2adf59"
	m19B1ScholarOracle         = "0753aee4-33db-48c5-9854-16a9d91535b2"
	m19B1PackbeastOracle       = "12c7289f-da53-403a-a607-227f43c7e171"

	m19B1BearType = "Creature — Bear"
)

// m19B1Vanilla are the batch's bodies with no rules text.
var m19B1Vanilla = map[string]string{
	m19B1BogstomperOracle: "Bogstomper",
	m19B1LoxodonOracle:    "Loxodon Line Breaker",
	m19B1OnakkeOgreOracle: "Onakke Ogre",
}

// m19B1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var m19B1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	m19B1HavocDevilsOracle:     {"Havoc Devils", "Creature — Devil", []string{"trample"}},
	m19B1HiredBladeOracle:      {"Hired Blade", "Creature — Human Assassin", []string{"flash"}},
	m19B1HostileMinotaurOracle: {"Hostile Minotaur", "Creature — Minotaur", []string{"haste"}},
	m19B1KnightOfTuskOracle:    {"Knight of the Tusk", "Creature — Human Knight", []string{"vigilance"}},
	m19B1VigilantBalothOracle:  {"Vigilant Baloth", "Creature — Beast", []string{"vigilance"}},
	m19B1RustwingFalconOracle:  {"Rustwing Falcon", "Creature — Bird", []string{"flying"}},
	m19B1SilverbeakOracle:      {"Silverbeak Griffin", "Creature — Griffin", []string{"flying"}},
	m19B1TwoHeadedZombieOracle: {"Two-Headed Zombie", "Creature — Zombie", []string{"menace"}},
	m19B1WallOfMistOracle:      {"Wall of Mist", "Creature — Wall", []string{"defender"}},
}

func TestM19B1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		m19B1GhastbarkOracle:       "Ghastbark Twins",
		m19B1SerrasGuardianOracle:  "Serra's Guardian",
		m19B1MaraudersAxeOracle:    "Marauder's Axe",
		m19B1CatalystOracle:        "Catalyst Elemental",
		m19B1GoblinMotivatorOracle: "Goblin Motivator",
		m19B1RootSnareOracle:       "Root Snare",
		m19B1TakeVengeanceOracle:   "Take Vengeance",
		m19B1CageOracle:            "Hieromancer's Cage",
		m19B1AviationOracle:        "Aviation Pioneer",
		m19B1GallantCavalryOracle:  "Gallant Cavalry",
		m19B1RhoxOracleOracle:      "Rhox Oracle",
		m19B1SalvagerOracle:        "Salvager of Secrets",
		m19B1StranglingOracle:      "Strangling Spores",
		m19B1AegisOracle:           "Aegis of the Heavens",
		m19B1FieryFinishOracle:     "Fiery Finish",
		m19B1LichsCaressOracle:     "Lich's Caress",
		m19B1SovereignsBiteOracle:  "Sovereign's Bite",
		m19B1VampireSovOracle:      "Vampire Sovereign",
		m19B1PlagueMareOracle:      "Plague Mare",
		m19B1VineMareOracle:        "Vine Mare",
		m19B1LightningMareOracle:   "Lightning Mare",
		m19B1BristlingBoarOracle:   "Bristling Boar",
		m19B1EncyclopediaOracle:    "Arcane Encyclopedia",
		m19B1FellSpecterOracle:     "Fell Specter",
		m19B1FountainOracle:        "Fountain of Renewal",
		m19B1KnightsPledgeOracle:   "Knight's Pledge",
		m19B1GearsmithProdOracle:   "Gearsmith Prodigy",
		m19B1AerialEngineerOracle:  "Aerial Engineer",
		m19B1GearsmithGuardOracle:  "Gearsmith Guardian",
		m19B1RavenousHarpyOracle:   "Ravenous Harpy",
		m19B1SiegebreakerOracle:    "Siegebreaker Giant",
		m19B1SkallaWolfOracle:      "Skalla Wolf",
		m19B1MilitiaBuglerOracle:   "Militia Bugler",
		m19B1DrillmasterOracle:     "Cavalry Drillmaster",
		m19B1AngelOfDawnOracle:     "Angel of the Dawn",
		m19B1DwarvenPriestOracle:   "Dwarven Priest",
		m19B1ScholarOracle:         "Scholar of Stars",
		m19B1PackbeastOracle:       "Trusty Packbeast",
	}
	for oracle, name := range m19B1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range m19B1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(all))
	}
	for oracle, name := range all {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- helpers -------------------------------------------------------------------

// m19B1CastCreature casts a creature card with a body from the active
// player's hand and leaves it on the stack.
func m19B1CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	return castFromHand(t, g, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness},
		game.CastSpellParams{})
}

// m19B1Tokens counts the tokens p controls on the battlefield.
func m19B1Tokens(g *game.Game, p uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.IsToken() {
			n++
		}
	}
	return n
}

// m19B1LibraryTop stacks cards on top of p's library, the first listed
// on top.
func m19B1LibraryTop(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// --- bodies and keywords -----------------------------------------------------

func TestM19B1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range m19B1Vanilla {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestM19B1OnakkeOgreIsAPlainCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Push(g, me.ID, "Onakke Ogre", "Creature — Ogre Warrior", m19B1OnakkeOgreOracle, 4, 2)
	bear := b12Creature(g, opp.ID, "Bear", m19B1BearType, 2, 2)
	if err := porBlock(t, g, ogre, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Onakke Ogre couldn't be blocked: %v", err)
	}
}

func TestM19B1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range m19B1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 1, 1)
			for _, kw := range k.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// The keywords printed beside the other abilities of the batch.
func TestM19B1OtherCardsKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Ghastbark Twins", "Creature — Treefolk", m19B1GhastbarkOracle, []string{"trample"}},
		{"Serra's Guardian", "Creature — Angel", m19B1SerrasGuardianOracle, []string{"flying", "vigilance"}},
		{"Gallant Cavalry", "Creature — Human Knight", m19B1GallantCavalryOracle, []string{"vigilance"}},
		{"Vampire Sovereign", "Creature — Vampire Noble", m19B1VampireSovOracle, []string{"flying"}},
		{"Vine Mare", "Creature — Elemental Horse", m19B1VineMareOracle, []string{"hexproof"}},
		{"Fell Specter", "Creature — Specter", m19B1FellSpecterOracle, []string{"flying"}},
		{"Ravenous Harpy", "Creature — Harpy", m19B1RavenousHarpyOracle, []string{"flying"}},
		{"Siegebreaker Giant", "Creature — Giant Warrior", m19B1SiegebreakerOracle, []string{"trample"}},
		{"Militia Bugler", "Creature — Human Soldier", m19B1MilitiaBuglerOracle, []string{"vigilance"}},
		{"Angel of the Dawn", "Creature — Angel", m19B1AngelOfDawnOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Flash: Hired Blade is cast at the beginning of combat.
func TestM19B1HiredBladeHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Hired Blade", "Creature — Human Assassin", m19B1HiredBladeOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Hired Blade at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Hired Blade did not resolve")
	}
}

// Haste: Hostile Minotaur attacks the turn it is cast (CR 702.10b).
func TestM19B1HostileMinotaurAttacksTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	id := castHoldCreature(t, g, "Hostile Minotaur", "Creature — Minotaur", m19B1HostileMinotaurOracle, 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Errorf("Hostile Minotaur couldn't attack the turn it arrived: %v", err)
	}
}

// Vigilance: attacking doesn't tap it (CR 702.20b).
func TestM19B1VigilantCreaturesAttackWithoutTapping(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Knight of the Tusk", "Creature — Human Knight", m19B1KnightOfTuskOracle},
		{"Vigilant Baloth", "Creature — Beast", m19B1VigilantBalothOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			declareAttack(t, g, opp.ID, id)
			if alliTapped(t, g, id) {
				t.Errorf("%s tapped to attack", tc.name)
			}
		})
	}
}

// Menace: one blocker is not enough (CR 702.110b).
func TestM19B1TwoHeadedZombieNeedsTwoBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := b12Push(g, me.ID, "Two-Headed Zombie", "Creature — Zombie", m19B1TwoHeadedZombieOracle, 4, 2)
	a := b12Creature(g, opp.ID, "Bear A", m19B1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", m19B1BearType, 2, 2)
	brAttack(t, g, zombie)
	brRefusal(t, g.DeclareBlocker(a, zombie), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: zombie}, {Blocker: b, Attacker: zombie}}); err != nil {
		t.Errorf("two blockers were refused: %v", err)
	}
}

// Defender: Wall of Mist can't attack (CR 702.3b).
func TestM19B1WallOfMistCantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Push(g, me.ID, "Wall of Mist", "Creature — Wall", m19B1WallOfMistOracle, 0, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("Wall of Mist attacked")
	}
}

// --- statics and block rules ------------------------------------------------

func TestM19B1GhastbarkTwinsBlocksTwo(t *testing.T) {
	g := newCatalogGame(t)
	twins := b12Push(g, g.Seats[1].ID, "Ghastbark Twins", "Creature — Treefolk", m19B1GhastbarkOracle, 7, 7)
	if got := mbCapacity(t, g, twins); got != 2 {
		t.Errorf("Ghastbark Twins can block %d creatures, want 2", got)
	}
}

func TestM19B1SerrasGuardianGivesYourOtherCreaturesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Serra's Guardian", "Creature — Angel", m19B1SerrasGuardianOracle, 5, 5)
	mine := b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", m19B1BearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("my other creature has no vigilance")
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature got vigilance")
	}
}

func TestM19B1MaraudersAxeGivesPlusTwoZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", m19B1BearType, 2, 2)
	axe := b12Push(g, me.ID, "Marauder's Axe", "Artifact — Equipment", m19B1MaraudersAxeOracle, 0, 0)
	equipTo(t, g, me.ID, axe, bear)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestM19B1KnightsPledgeGivesPlusTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	bear := b12Creature(g, me.ID, "Bear", m19B1BearType, 2, 2)
	pledge := castCatalogSpell(t, g, "Knight's Pledge", "Enchantment — Aura", m19B1KnightsPledgeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pledge) {
		t.Fatal("Knight's Pledge did not resolve")
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestM19B1GearsmithProdigyNeedsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prodigy := b12Push(g, me.ID, "Gearsmith Prodigy", "Creature — Human Artificer", m19B1GearsmithProdOracle, 1, 2)
	if p, tough := ptOf(t, g, prodigy); p != 1 || tough != 2 {
		t.Errorf("without an artifact it is %d/%d, want 1/2", p, tough)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	if p, tough := ptOf(t, g, prodigy); p != 2 || tough != 2 {
		t.Errorf("with an artifact it is %d/%d, want 2/2", p, tough)
	}
}

func TestM19B1AerialEngineerFliesWithAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	engineer := b12Push(g, me.ID, "Aerial Engineer", "Creature — Human Artificer", m19B1AerialEngineerOracle, 2, 4)
	if p, _ := ptOf(t, g, engineer); p != 2 || hasEffectiveKeyword(t, g, engineer, "flying") {
		t.Errorf("without an artifact: power %d, flying %v; want 2 and no flying", p, hasEffectiveKeyword(t, g, engineer, "flying"))
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	if p, tough := ptOf(t, g, engineer); p != 4 || tough != 4 {
		t.Errorf("with an artifact it is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, engineer, "flying") {
		t.Error("with an artifact it has no flying")
	}
}

func TestM19B1GearsmithGuardianNeedsABlueCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := b12Push(g, me.ID, "Gearsmith Guardian", "Artifact Creature — Construct", m19B1GearsmithGuardOracle, 3, 5)
	red := b12Creature(g, me.ID, "Red Bear", m19B1BearType, 2, 2)
	hmlSetColors(g, red, "R")
	if p, _ := ptOf(t, g, guardian); p != 3 {
		t.Errorf("without a blue creature its power is %d, want 3", p)
	}
	blue := b12Creature(g, me.ID, "Blue Bear", m19B1BearType, 2, 2)
	hmlSetColors(g, blue, "U")
	if p, tough := ptOf(t, g, guardian); p != 5 || tough != 5 {
		t.Errorf("with a blue creature it is %d/%d, want 5/5", p, tough)
	}
}

func TestM19B1MaresCantBeBlockedByTheirColor(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, forbidden, allowed string
	}{
		{"Plague Mare", "Creature — Nightmare Horse", m19B1PlagueMareOracle, "W", "G"},
		{"Vine Mare", "Creature — Elemental Horse", m19B1VineMareOracle, "B", "W"},
		{"Lightning Mare", "Creature — Elemental Horse", m19B1LightningMareOracle, "U", "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, color := range []string{tc.forbidden, tc.allowed} {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				mare := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
				blocker := b12Creature(g, opp.ID, "Blocker", m19B1BearType, 2, 2)
				hmlSetColors(g, blocker, color)
				err := porBlock(t, g, mare, opp.ID, blocker)
				if color == tc.forbidden && err == nil {
					t.Errorf("a %s creature blocked %s", color, tc.name)
				}
				if color == tc.allowed && err != nil {
					t.Errorf("a %s creature couldn't block %s: %v", color, tc.name, err)
				}
			}
		})
	}
}

func TestM19B1BristlingBoarRefusesASecondBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boar := b12Push(g, me.ID, "Bristling Boar", "Creature — Boar", m19B1BristlingBoarOracle, 4, 3)
	a := b12Creature(g, opp.ID, "Bear A", m19B1BearType, 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", m19B1BearType, 2, 2)
	brAttack(t, g, boar)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: boar}, {Blocker: b, Attacker: boar},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlocker(a, boar); err != nil {
		t.Errorf("one blocker is legal: %v", err)
	}
}

// --- spells ------------------------------------------------------------------

func TestM19B1RootSnarePreventsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	defender := pushBear(g, me.ID, "Defender", 2)
	attacker := pushBear(g, opp.ID, "Attacker", 2)
	castCatalogSpell(t, g, "Root Snare", "Instant", m19B1RootSnareOracle, nil)
	passPriorityAroundTable(t, g)
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := damageMarkedOn(g, defender); got != 0 {
		t.Errorf("combat damage marked = %d, want 0", got)
	}
	if err := g.MarkDamage(defender, 1); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if got := damageMarkedOn(g, defender); got != 1 {
		t.Errorf("non-combat damage marked = %d, want 1", got)
	}
}

func TestM19B1TakeVengeanceDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	untapped := b12Creature(g, opp.ID, "Untapped", m19B1BearType, 2, 2)
	tapped := b12Creature(g, opp.ID, "Tapped", m19B1BearType, 2, 2)
	tapForTest(t, g, tapped)
	if err := castCatalogSpellErr(t, g, "Take Vengeance", "Sorcery", m19B1TakeVengeanceOracle, cardTarget(untapped)); err == nil {
		t.Fatal("Take Vengeance targeted an untapped creature")
	}
	castCatalogSpell(t, g, "Take Vengeance", "Sorcery", m19B1TakeVengeanceOracle, cardTarget(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestM19B1TargetPumpsAndShrinks(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, tough           int
	}{
		{"Strangling Spores", m19B1StranglingOracle, "Instant", 1, 1},
		{"Aegis of the Heavens", m19B1AegisOracle, "Instant", 5, 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := seatsForCombat(g)
			bear := b12Creature(g, me.ID, "Big Bear", m19B1BearType, 4, 4)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.tough {
				t.Errorf("the 4/4 is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
		})
	}
}

func TestM19B1FieryFinishDealsSeven(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	big := b12Creature(g, opp.ID, "Seven", "Creature — Giant", 7, 7)
	castCatalogSpell(t, g, "Fiery Finish", "Sorcery", m19B1FieryFinishOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("a 7/7 survived 7 damage")
	}
}

func TestM19B1LichsCaressDestroysAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	bear := b12Creature(g, opp.ID, "Their Bear", m19B1BearType, 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Lich's Caress", "Sorcery", m19B1LichsCaressOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestM19B1SovereignsBiteDrainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Sovereign's Bite", "Sorcery", m19B1SovereignsBiteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("life me %d->%d opp %d->%d, want +3 / -3", mine, me.Life, theirs, opp.Life)
	}
}

// --- enters triggers -----------------------------------------------------------

func TestM19B1VampireSovereignDrainsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine, theirs := me.Life, opp.Life
	m19B1CastCreature(t, g, "Vampire Sovereign", "Creature — Vampire Noble", m19B1VampireSovOracle, 3, 4)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("the trigger targeted its own controller")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("life me %d->%d opp %d->%d, want +3 / -3", mine, me.Life, theirs, opp.Life)
	}
}

func TestM19B1PlagueMareShrinksOnlyTheOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", m19B1BearType, 2, 2)
	castHoldCreature(t, g, "Plague Mare", "Creature — Nightmare Horse", m19B1PlagueMareOracle, 2, 2)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their bear is %d/%d, want 1/1", p, tough)
	}
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my bear is %d/%d, want 2/2", p, tough)
	}
}

func TestM19B1AngelOfTheDawnPumpsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", m19B1BearType, 2, 2)
	castHoldCreature(t, g, "Angel of the Dawn", "Creature — Angel", m19B1AngelOfDawnOracle, 3, 3)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("my bear has no vigilance")
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tough)
	}
}

func TestM19B1DwarvenPriestGainsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	life := me.Life
	castHoldCreature(t, g, "Dwarven Priest", "Creature — Dwarf Cleric", m19B1DwarvenPriestOracle, 2, 4)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 (the Bear and the Priest)", life, me.Life)
	}
}

func TestM19B1TokenMakers(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Aviation Pioneer", "Creature — Human Artificer", m19B1AviationOracle},
		{"Gallant Cavalry", "Creature — Human Knight", m19B1GallantCavalryOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := seatsForCombat(g)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if got := m19B1Tokens(g, me.ID); got != 1 {
				t.Errorf("%s made %d tokens, want 1", tc.name, got)
			}
		})
	}
}

func TestM19B1RhoxOracleDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Rhox Oracle", "Creature — Rhino Monk", m19B1RhoxOracleOracle, 4, 2)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestM19B1ScholarOfStarsDrawsOnlyWithAnArtifact(t *testing.T) {
	for _, withArtifact := range []bool{true, false} {
		g := newCatalogGame(t)
		me, _ := seatsForCombat(g)
		seedLibrary(me, "A", "B")
		if withArtifact {
			b12Permanent(g, me.ID, "My Rock", "Artifact")
		}
		m19B1CastCreature(t, g, "Scholar of Stars", "Creature — Human Artificer", m19B1ScholarOracle, 3, 2)
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		want := 0
		if withArtifact {
			want = 1
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("with an artifact %v: drew %d, want %d", withArtifact, got, want)
		}
	}
}

func TestM19B1SalvagerOfSecretsReturnsAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	bolt := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	bear := pushGraveyardCardTyped(me, "Dead Bear", m19B1BearType)
	m19B1CastCreature(t, g, "Salvager of Secrets", "Creature — Merfolk Wizard", m19B1SalvagerOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	if pick := latestPickTarget(g, me.ID); hasID(pick.PickTargetCards, bear) {
		t.Error("a creature card was offered for 'target instant or sorcery card'")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not come back to hand")
	}
}

func TestM19B1TrustyPackbeastReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	relic := pushGraveyardCardTyped(me, "Old Relic", "Artifact")
	spell := pushGraveyardCardTyped(me, "Old Spell", "Sorcery")
	m19B1CastCreature(t, g, "Trusty Packbeast", "Creature — Beast", m19B1PackbeastOracle, 2, 3)
	b04WaitForPick(t, g, me.ID)
	if pick := latestPickTarget(g, me.ID); hasID(pick.PickTargetCards, spell) {
		t.Error("a sorcery card was offered for 'target artifact card'")
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not come back to hand")
	}
}

func TestM19B1CavalryDrillmasterGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	bear := b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	m19B1CastCreature(t, g, "Cavalry Drillmaster", "Creature — Human Knight", m19B1DrillmasterOracle, 2, 1)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear has no first strike")
	}
}

func TestM19B1FellSpecterDiscardAndDrain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	card := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	hand, life := opp.Hand.Size(), opp.Life
	m19B1CastCreature(t, g, "Fell Specter", "Creature — Specter", m19B1FellSpecterOracle, 1, 3)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil {
		answerDiscard(t, g, opp.ID, card)
	}
	passPriorityAroundTable(t, g)
	if got := hand - opp.Hand.Size(); got != 1 {
		t.Errorf("opponent discarded %d, want 1", got)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want -2 for the discard", life, opp.Life)
	}
}

func TestM19B1HieromancersCageExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", m19B1BearType, 4, 4)
	cage := castCatalogSpell(t, g, "Hieromancer's Cage", "Enchantment", m19B1CageOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cage) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after the Cage left, %d copies under its owner's control, want 1", n)
	}
}

func TestM19B1LookAtTheTopAndTakeOne(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		depth                  int
		good, bad              game.Card
	}{
		{"Skalla Wolf", "Creature — Wolf Spirit", m19B1SkallaWolfOracle, 3, 3, 5,
			game.Card{Name: "Green Card", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}},
			game.Card{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}}},
		{"Militia Bugler", "Creature — Human Soldier", m19B1MilitiaBuglerOracle, 2, 3, 4,
			game.Card{Name: "Scout", TypeLine: "Creature — Human", Power: 2, Toughness: 1},
			game.Card{Name: "Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := seatsForCombat(g)
			cards := []game.Card{tc.bad, tc.good}
			for len(cards) < tc.depth {
				cards = append(cards, game.Card{Name: "Land", TypeLine: "Land"})
			}
			cards = append(cards, tc.good) // below the look: never offered
			ids := m19B1LibraryTop(me, cards...)
			bad, good, deep := ids[0], ids[1], ids[len(ids)-1]
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			pick := chooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no prompt to take a card")
			}
			for _, id := range pick.ChooseCards {
				if id == bad || id == deep {
					t.Errorf("offered a card it may not take: %s", id)
				}
			}
			answerChooseCards(t, g, me.ID, good)
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(good) {
				t.Error("the chosen card is not in hand")
			}
		})
	}
}

// --- activated abilities ---------------------------------------------------------

func TestM19B1CatalystElementalSacrificesForRedRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := seatsForCombat(g)
	cat := b12Push(g, me.ID, "Catalyst Elemental", "Creature — Elemental", m19B1CatalystOracle, 2, 2)
	if err := g.ActivateManaAbility(me.ID, cat, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("pool has %d red, want 2", got)
	}
	if g.Battlefield.Contains(cat) {
		t.Error("Catalyst Elemental was not sacrificed")
	}
}

func TestM19B1GoblinMotivatorGrantsHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := seatsForCombat(g)
	motivator := b12Push(g, me.ID, "Goblin Motivator", "Creature — Goblin Warrior", m19B1GoblinMotivatorOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", m19B1BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, motivator, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear has no haste")
	}
}

func TestM19B1LightningMareCantBeCounteredAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seatsForCombat(g)
	id := m19B1CastCreature(t, g, "Lightning Mare", "Creature — Elemental Horse", m19B1LightningMareOracle, 3, 1)
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("Lightning Mare can be countered")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Lightning Mare did not resolve")
	}
	floatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 4 || tough != 1 {
		t.Errorf("Lightning Mare is %d/%d, want 4/1", p, tough)
	}
}

func TestM19B1ArcaneEncyclopediaDrawsForThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := seatsForCombat(g)
	seedLibrary(me, "A", "B")
	book := b12Push(g, me.ID, "Arcane Encyclopedia", "Artifact — Book", m19B1EncyclopediaOracle, 0, 0)
	floatForTest(g, me, "CCC")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestM19B1FountainOfRenewalSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := seatsForCombat(g)
	seedLibrary(me, "A", "B")
	fountain := b12Push(g, me.ID, "Fountain of Renewal", "Artifact", m19B1FountainOracle, 0, 0)
	floatForTest(g, me, "CCC")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fountain) {
		t.Error("the Fountain was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestM19B1FountainOfRenewalGainsOneAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Fountain of Renewal", "Artifact", m19B1FountainOracle, false)
	life := owner.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if owner.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, owner.Life)
	}
}

func TestM19B1RavenousHarpyEatsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := seatsForCombat(g)
	harpy := b12Push(g, me.ID, "Ravenous Harpy", "Creature — Harpy", m19B1RavenousHarpyOracle, 1, 2)
	fodder := b12Creature(g, me.ID, "Fodder", m19B1BearType, 2, 2)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, harpy, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
	if p, tough := ptOf(t, g, harpy); p != 2 || tough != 3 {
		t.Errorf("Ravenous Harpy is %d/%d, want 2/3", p, tough)
	}
}

func TestM19B1SiegebreakerGiantStopsOneCreatureBlocking(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	giant := b12Push(g, me.ID, "Siegebreaker Giant", "Creature — Giant Warrior", m19B1SiegebreakerOracle, 6, 3)
	bear := b12Creature(g, them.ID, "Their Bear", m19B1BearType, 2, 2)
	other := b12Creature(g, them.ID, "Their Other", m19B1BearType, 2, 2)
	attacker := b12Creature(g, me.ID, "Attacker", m19B1BearType, 2, 2)
	floatForTest(g, me, "RRRR")
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	attackIntoBlocks(t, g, them.ID, attacker)
	if blockRefusal(t, g, bear, attacker) == "" {
		t.Error("the target blocked")
	}
	if got := blockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}
