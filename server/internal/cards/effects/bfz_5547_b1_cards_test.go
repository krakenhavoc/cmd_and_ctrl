package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bfz_5547_b1_cards_test.go — Battle for Zendikar, batch 1 (#5547): two
// vanilla bodies, nine creatures whose whole text is keywords, ten rally
// Allies, seven landfall creatures, and cards whose text a catalogued
// card already prints. Each test plays the real catalog entry through
// the engine and asserts what its text says.

const (
	bfzB1BroodhunterWurmOracle   = "02d50c11-51e3-4b3f-bf3d-3b1775d514f6"
	bfzB1ExpeditionEnvoyOracle   = "8bc7c769-33e0-4482-95fc-b344b772e174"
	bfzB1CloudMantaOracle        = "59b6bf51-d15d-4032-85ac-29b234b59fc9"
	bfzB1KitesailScoutOracle     = "b365a7dd-fed1-40f8-9544-2b5bc1af761c"
	bfzB1ShadowGliderOracle      = "b92e0156-ea0f-4b13-b3bc-7e84a31445fc"
	bfzB1GhostlySentinelOracle   = "a16f4b53-c8cf-45b0-89e7-507acd094a64"
	bfzB1EldraziDevastatorOracle = "bc08ac66-969c-4905-a2d6-16ff565be475"
	bfzB1FortifiedRampartOracle  = "a55202a9-bddb-40df-b93f-06678ae1e0ea"
	bfzB1ShatterskullOracle      = "e2bc1600-eeec-443b-99e4-dd61497edd5b"
	bfzB1PlatedCrusherOracle     = "411f3ea6-36ad-4bbc-8cde-e79662a08ee9"
	bfzB1VestigeOracle           = "55187899-c445-48e9-80b9-235eaf42f803"
	bfzB1ChannelerOracle         = "33826f91-7680-4a4c-be39-db6c2c0bd465"
	bfzB1LifespringDruidOracle   = "165c1b17-2bde-4c96-892d-4a8fdc7a57b7"
	bfzB1CliffsideLookoutOracle  = "97671aac-7322-44ae-8908-d573734db3af"
	bfzB1CoralhelmGuideOracle    = "979e670e-bbd8-4595-8d9d-ad1c0bc422e3"
	bfzB1LavastepRaiderOracle    = "3a96782b-e9c8-4bc5-845d-45f6bf7b8dc4"
	bfzB1DemonsGraspOracle       = "13228a22-d63a-4952-a482-4f84da696e88"
	bfzB1GideonsReproachOracle   = "ec43d67c-2254-4782-8bd4-3318b25907e6"
	bfzB1NaturalConnectionOracle = "dc85c659-f2ed-4b13-807c-d853f74de410"
	bfzB1ReclaimingVinesOracle   = "3ec3d99d-2c49-469a-81d5-f1609b6e49ca"
	bfzB1StonefuryOracle         = "577554de-78c3-4fa2-a866-06b80280bc97"
	bfzB1VolcanicUpheavalOracle  = "4d628f52-183a-4a89-b05f-50463a82513d"
	bfzB1TurnAgainstOracle       = "634d6357-af76-4e25-8c63-81522bc726f5"
	bfzB1CompleteDisregardOracle = "b0bd842c-d518-42b3-850c-22a70f7f1436"
	bfzB1ScourOracle             = "241a50c5-f65f-4847-89c7-5c0ef6025dc1"
	bfzB1CourierGriffinOracle    = "2bb0a5e3-ccb0-4f87-b021-e60ebfa35b2b"
	bfzB1RotShamblerOracle       = "88c187fd-b3d3-4218-8e63-380a9ca9f1b4"
	bfzB1AngelOfRenewalOracle    = "c604fda4-164c-4966-94f8-1e3a96d5ef9d"
	bfzB1DefiantBloodlordOracle  = "7bf5196d-a16c-418f-943c-2128e40c55da"
	bfzB1GeopedeOracle           = "78cd6d20-033f-471f-a86e-59ec8307ae1f"
	bfzB1EndlessOneOracle        = "7a51780a-fa28-4ee0-94c7-4330800ca9cb"
	bfzB1TajuruStalwartOracle    = "ac87cc33-4a71-471b-917d-6a92721ea0b9"
	bfzB1WoodlandWandererOracle  = "d5f3c711-edfa-4105-9844-e3f444942a03"
	bfzB1ChasmGuideOracle        = "446d3207-bff8-4edf-9b20-62abcc56427e"
	bfzB1TajuruBeastmasterOracle = "cc72ec48-9e8b-4ded-b445-f3a839fd1abe"
	bfzB1TajuruWarcallerOracle   = "8b89c6f1-8c9e-43ef-ad6b-11424de090e4"

	bfzB1WurmType  = "Creature — Wurm"
	bfzB1EnvoyType = "Creature — Human Scout Ally"
)

// bfzB1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var bfzB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	bfzB1CloudMantaOracle:        {"Cloud Manta", "Creature — Fish", []string{"flying"}},
	bfzB1KitesailScoutOracle:     {"Kitesail Scout", "Creature — Kor Scout", []string{"flying"}},
	bfzB1ShadowGliderOracle:      {"Shadow Glider", "Creature — Kor Soldier", []string{"flying"}},
	bfzB1GhostlySentinelOracle:   {"Ghostly Sentinel", "Creature — Kor Spirit", []string{"flying", "vigilance"}},
	bfzB1EldraziDevastatorOracle: {"Eldrazi Devastator", "Creature — Eldrazi", []string{"trample"}},
	bfzB1FortifiedRampartOracle:  {"Fortified Rampart", "Creature — Wall", []string{"defender"}},
	bfzB1ShatterskullOracle:      {"Shatterskull Recruit", "Creature — Giant Warrior Ally", []string{"menace"}},
	bfzB1PlatedCrusherOracle:     {"Plated Crusher", "Creature — Beast", []string{"trample", "hexproof"}},
	bfzB1VestigeOracle:           {"Vestige of Emrakul", "Creature — Eldrazi Drone", []string{game.KeywordDevoid, "trample"}},
}

// bfzB1Landfall are the batch's landfall pumps: base body, the pump, and
// the printed keyword beside it ("" for none).
var bfzB1Landfall = []struct {
	name, typeLine, oracle string
	power, tough, n        int
	keyword                string
}{
	{"Scythe Leopard", "Creature — Cat", "ab539b16-d2a0-404d-8ad3-7ff33f2f1abe", 1, 1, 1, ""},
	{"Valakut Predator", "Creature — Elemental", "b86a5d90-0820-46cb-b43c-c03a423a3a9f", 2, 2, 2, ""},
	{"Grove Rumbler", "Creature — Elemental", "5dcbed75-bafc-492a-85a1-44f3ff06fa78", 3, 3, 2, "trample"},
	{"Makindi Sliderunner", "Creature — Beast", "4754e9c6-1790-4c01-87c2-3a16ce7e65f5", 2, 1, 1, "trample"},
	{"Wave-Wing Elemental", "Creature — Elemental", "35044f8a-c8e0-41ae-ba61-43109e51f7e8", 3, 4, 2, "flying"},
	{"Geyserfield Stalker", "Creature — Elemental", "d93d98ef-a5b6-4d6f-9828-fd5943471d0f", 3, 2, 2, "menace"},
	{"Ondu Greathorn", "Creature — Beast", "9d061ec5-1edf-46e7-9e6a-4ddc9ced4114", 2, 3, 2, "first strike"},
}

// bfzB1RallyGrants are the rally Allies that grant a keyword to the team.
var bfzB1RallyGrants = []struct {
	name, typeLine, oracle string
	power, tough           int
	keyword                string
}{
	{"Chasm Guide", "Creature — Goblin Scout Ally", bfzB1ChasmGuideOracle, 3, 2, "haste"},
	{"Firemantle Mage", "Creature — Human Shaman Ally", "27599eff-19df-48e2-9a1e-0c087f42a075", 2, 2, "menace"},
	{"Hero of Goma Fada", "Creature — Human Knight Ally", "f15908ea-2be0-4035-933d-4412fe032164", 4, 3, "indestructible"},
	{"Kor Bladewhirl", "Creature — Kor Soldier Ally", "7024d9db-df2e-4eba-9fa0-76ebe3df05a7", 2, 2, "first strike"},
	{"Lantern Scout", "Creature — Human Scout Ally", "803f2f95-533e-4b5a-8ede-d00fad779430", 3, 2, "lifelink"},
	{"Makindi Patrol", "Creature — Human Knight Ally", "d3f58efb-ba75-4caf-9e33-74236c707b83", 2, 3, "vigilance"},
	{"Ondu Champion", "Creature — Minotaur Warrior Ally", "b638407b-4749-41fa-8054-f9db57b849ea", 4, 3, "trample"},
	{"Resolute Blademaster", "Creature — Human Soldier Ally", "fec7b50a-7ed3-461d-9e91-e5c6ca9f774f", 2, 2, "double strike"},
}

func TestBfz5547B1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		bfzB1BroodhunterWurmOracle:   "Broodhunter Wurm",
		bfzB1ExpeditionEnvoyOracle:   "Expedition Envoy",
		bfzB1ChannelerOracle:         "Kozilek's Channeler",
		bfzB1LifespringDruidOracle:   "Lifespring Druid",
		bfzB1CliffsideLookoutOracle:  "Cliffside Lookout",
		bfzB1CoralhelmGuideOracle:    "Coralhelm Guide",
		bfzB1LavastepRaiderOracle:    "Lavastep Raider",
		bfzB1DemonsGraspOracle:       "Demon's Grasp",
		bfzB1GideonsReproachOracle:   "Gideon's Reproach",
		bfzB1NaturalConnectionOracle: "Natural Connection",
		bfzB1ReclaimingVinesOracle:   "Reclaiming Vines",
		bfzB1StonefuryOracle:         "Stonefury",
		bfzB1VolcanicUpheavalOracle:  "Volcanic Upheaval",
		bfzB1TurnAgainstOracle:       "Turn Against",
		bfzB1CompleteDisregardOracle: "Complete Disregard",
		bfzB1ScourOracle:             "Scour from Existence",
		bfzB1CourierGriffinOracle:    "Courier Griffin",
		bfzB1RotShamblerOracle:       "Rot Shambler",
		bfzB1AngelOfRenewalOracle:    "Angel of Renewal",
		bfzB1DefiantBloodlordOracle:  "Defiant Bloodlord",
		bfzB1GeopedeOracle:           "Tunneling Geopede",
		bfzB1EndlessOneOracle:        "Endless One",
		bfzB1TajuruStalwartOracle:    "Tajuru Stalwart",
		bfzB1WoodlandWandererOracle:  "Woodland Wanderer",
		bfzB1TajuruBeastmasterOracle: "Tajuru Beastmaster",
		bfzB1TajuruWarcallerOracle:   "Tajuru Warcaller",
	}
	for oracle, k := range bfzB1Keyworded {
		all[oracle] = k.name
	}
	for _, l := range bfzB1Landfall {
		all[l.oracle] = l.name
	}
	for _, r := range bfzB1RallyGrants {
		all[r.oracle] = r.name
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

// --- bodies and keywords -----------------------------------------------------

func TestBfz5547B1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range map[string]string{
		bfzB1BroodhunterWurmOracle: "Broodhunter Wurm",
		bfzB1ExpeditionEnvoyOracle: "Expedition Envoy",
	} {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
	g := newCatalogGame(t)
	wurm := b12Push(g, g.Seats[0].ID, "Broodhunter Wurm", bfzB1WurmType, bfzB1BroodhunterWurmOracle, 4, 3)
	if p, tough := ptOf(t, g, wurm); p != 4 || tough != 3 {
		t.Errorf("Broodhunter Wurm is %d/%d, want 4/3", p, tough)
	}
}

func TestBfz5547B1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range bfzB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 2, 2)
			for _, kw := range k.keywords {
				if kw == game.KeywordDevoid {
					continue
				}
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Devoid: Vestige of Emrakul is colourless whatever its cost (CR 702.114a).
func TestBfz5547B1VestigeOfEmrakulIsColourless(t *testing.T) {
	c := game.Card{Name: "Vestige of Emrakul", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{3}{R}", OracleID: bfzB1VestigeOracle}
	if !c.IsColorless() {
		t.Error("Vestige of Emrakul has a colour")
	}
}

// Flying: a ground creature can't block Cloud Manta (CR 702.9b).
func TestBfz5547B1CloudMantaIsBlockedOnlyByFlyingOrReach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	manta := b12Push(g, me.ID, "Cloud Manta", "Creature — Fish", bfzB1CloudMantaOracle, 3, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := porBlock(t, g, manta, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Cloud Manta")
	}
}

// Defender: Fortified Rampart can't attack (CR 702.3b).
func TestBfz5547B1FortifiedRampartCantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Push(g, me.ID, "Fortified Rampart", "Creature — Wall", bfzB1FortifiedRampartOracle, 0, 6)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Error("Fortified Rampart attacked")
	}
}

// Vigilance: Ghostly Sentinel attacks without tapping (CR 702.20b).
func TestBfz5547B1GhostlySentinelAttacksWithoutTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Ghostly Sentinel", "Creature — Kor Spirit", bfzB1GhostlySentinelOracle, 3, 3)
	declareAttack(t, g, opp.ID, id)
	if alliTapped(t, g, id) {
		t.Error("Ghostly Sentinel tapped to attack")
	}
}

// Hexproof: an opponent's Demon's Grasp can't target Plated Crusher; it
// kills a creature with 5 toughness.
func TestBfz5547B1DemonsGraspKillsButCantTargetPlatedCrusher(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	crusher := b12Push(g, opp.ID, "Plated Crusher", "Creature — Beast", bfzB1PlatedCrusherOracle, 7, 6)
	if err := castCatalogSpellErr(t, g, "Demon's Grasp", "Sorcery", bfzB1DemonsGraspOracle, cardRefs(crusher)); err == nil {
		t.Fatal("Demon's Grasp targeted an opponent's hexproof creature")
	}
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castCatalogSpell(t, g, "Demon's Grasp", "Sorcery", bfzB1DemonsGraspOracle, cardRefs(ogre))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !opp.Graveyard.Contains(ogre) {
		t.Error("a 5/5 survived -5/-5")
	}
}

// --- mana abilities ----------------------------------------------------------

func TestBfz5547B1KozileksChannelerTapsForTwoColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Kozilek's Channeler", "Creature — Eldrazi", bfzB1ChannelerOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool holds %d mana, want 2", me.ManaPool.Total())
	}
	if !layeredCard(t, g, id).Tapped {
		t.Error("the {T} cost did not tap Kozilek's Channeler")
	}
}

func TestBfz5547B1LifespringDruidTapsForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Lifespring Druid", "Creature — Elf Druid", bfzB1LifespringDruidOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("want a five-colour pick, got %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("pool = %+v, want one {R}", me.ManaPool)
	}
}

// --- activated abilities -----------------------------------------------------

func TestBfz5547B1CliffsideLookoutPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lookout := b12Push(g, me.ID, "Cliffside Lookout", "Creature — Kor Scout Ally", bfzB1CliffsideLookoutOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, lookout, 0, "{C}{C}{C}{C}{W}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, lookout); p != 2 || tough != 2 {
		t.Errorf("the Lookout is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tough)
	}
}

func TestBfz5547B1CoralhelmGuideMakesATargetUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := b12Push(g, me.ID, "Coralhelm Guide", "Creature — Merfolk Scout Ally", bfzB1CoralhelmGuideOracle, 2, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, guide, 0, "{C}{C}{C}{C}{U}", cardRefs(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := porBlock(t, g, bear, opp.ID, blocker); err == nil {
		t.Error("the chosen creature was blocked")
	}
}

func TestBfz5547B1LavastepRaiderFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raider := b12Push(g, me.ID, "Lavastep Raider", "Creature — Goblin Warrior", bfzB1LavastepRaiderOracle, 1, 2)
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		if err := hobB4Activate(t, g, me.ID, raider, 0, "{C}{C}{R}", nil); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := ptOf(t, g, raider); p != 5 || tough != 2 {
		t.Errorf("Lavastep Raider is %d/%d after two activations, want 5/2", p, tough)
	}
}

// --- spells ------------------------------------------------------------------

func TestBfz5547B1GideonsReproachHitsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Ogre", 4, 4)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Ogre", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	first := b13HandCard(me, "Gideon's Reproach", "Instant", bfzB1GideonsReproachOracle)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("Gideon's Reproach targeted a creature that is not in combat")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the 4/4 attacker survived 4 damage")
	}
}

func TestBfz5547B1NaturalConnectionFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
	)
	castCatalogSpell(t, g, "Natural Connection", "Instant", bfzB1NaturalConnectionOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bayou") != uuid.Nil {
		t.Error("a nonbasic was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest A")
	card, ok := searchFetchedCard(g, "Forest A")
	if !ok || !card.Tapped {
		t.Errorf("the basic should be on the battlefield tapped, got %+v (found %v)", card, ok)
	}
}

// Reclaiming Vines takes an artifact, enchantment or land; Volcanic
// Upheaval only a land. Neither takes a creature.
func TestBfz5547B1LandDestructionTakesOnlyItsKinds(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		hits                   []string
	}{
		{"Reclaiming Vines", "Sorcery", bfzB1ReclaimingVinesOracle, []string{"Artifact", "Enchantment", "Land"}},
		{"Volcanic Upheaval", "Instant", bfzB1VolcanicUpheavalOracle, []string{"Land"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			if err := castCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, cardRefs(bear)); err == nil {
				t.Errorf("%s targeted a creature", tc.name)
			}
			for _, kind := range tc.hits {
				id := b12Permanent(g, opp.ID, "Their "+kind, kind)
				castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardRefs(id))
				passPriorityAroundTable(t, g)
				if g.Battlefield.Contains(id) || !opp.Graveyard.Contains(id) {
					t.Errorf("%s did not destroy the %s", tc.name, kind)
				}
			}
		})
	}
}

func TestBfz5547B1StonefuryCountsYourLands(t *testing.T) {
	for _, tc := range []struct {
		tough int
		dies  bool
	}{{3, true}, {4, false}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		for i := 0; i < 3; i++ {
			b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
		}
		b12Permanent(g, opp.ID, "Their Mountain", "Basic Land — Mountain")
		victim := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, tc.tough)
		castCatalogSpell(t, g, "Stonefury", "Instant", bfzB1StonefuryOracle, cardRefs(victim))
		passPriorityAroundTable(t, g)
		if dead := !g.Battlefield.Contains(victim); dead != tc.dies {
			t.Errorf("a 3/%d dealt damage by three lands: died %v, want %v", tc.tough, dead, tc.dies)
		}
	}
}

func TestBfz5547B1TurnAgainstStealsUntapsAndHastes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	avB26Tapped(t, g, theirs, true)
	castCatalogSpell(t, g, "Turn Against", "Instant", bfzB1TurnAgainstOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Errorf("controller %s, want the caster %s", got, me.ID)
	}
	if layeredCard(t, g, theirs).Tapped {
		t.Error("the stolen creature was not untapped")
	}
	if !hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the stolen creature has no haste")
	}
}

func TestBfz5547B1CompleteDisregardExilesOnlyPowerThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Ogre", 3, 3)
	if err := castCatalogSpellErr(t, g, "Complete Disregard", "Instant", bfzB1CompleteDisregardOracle, cardRefs(big)); err == nil {
		t.Fatal("Complete Disregard targeted a creature with power 4")
	}
	castCatalogSpell(t, g, "Complete Disregard", "Instant", bfzB1CompleteDisregardOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Exile.Contains(small) {
		t.Error("the 3/3 was not exiled")
	}
}

func TestBfz5547B1ScourFromExistenceExilesAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Scour from Existence", "Instant", bfzB1ScourOracle, cardRefs(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Exile.Contains(land) {
		t.Error("the land was not exiled")
	}
}

// --- triggers ----------------------------------------------------------------

func TestBfz5547B1CourierGriffinGainsTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	griffin := castHoldCreature(t, g, "Courier Griffin", "Creature — Griffin", bfzB1CourierGriffinOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if !hasEffectiveKeyword(t, g, griffin, "flying") {
		t.Error("Courier Griffin has no flying")
	}
}

func TestBfz5547B1AngelOfRenewalGainsOnePerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castHoldCreature(t, g, "Angel of Renewal", "Creature — Angel Ally", bfzB1AngelOfRenewalOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 (two Bears and the Angel)", life, me.Life)
	}
}

func TestBfz5547B1DefiantBloodlordDrainsTheTargetOnLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Defiant Bloodlord", "Creature — Vampire", bfzB1DefiantBloodlordOracle, 4, 5)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the target lost %d, want 3", before-opp.Life)
	}
}

func TestBfz5547B1RotShamblerGrowsOnlyForAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shambler := b12Push(g, me.ID, "Rot Shambler", "Creature — Fungus", bfzB1RotShamblerOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Mouse", "Creature — Mouse", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	midB17Bolt(t, g, rprB30Ref(theirs))
	if n := countersOn(g, shambler, game.CounterPlusOne); n != 0 {
		t.Fatalf("an opponent's creature dying gave %d counters", n)
	}
	midB17Bolt(t, g, rprB30Ref(mine))
	if n := countersOn(g, shambler, game.CounterPlusOne); n != 1 {
		t.Fatalf("my creature died and the Shambler has %d counters, want 1", n)
	}
}

func TestBfz5547B1TunnelingGeopedeHitsEachOpponentOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Tunneling Geopede", "Creature — Insect", bfzB1GeopedeOracle, 3, 2)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 1
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
}

func TestBfz5547B1LandfallCreaturesGrowOnYourLand(t *testing.T) {
	for _, tc := range bfzB1Landfall {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
			playLandFromHand(t, g, "Forest", "")
			passPriorityAroundTable(t, g)
			if p, tough := rfPT(t, g, id); p != tc.power+tc.n || tough != tc.tough+tc.n {
				t.Errorf("after a land: %d/%d, want %d/%d", p, tough, tc.power+tc.n, tc.tough+tc.n)
			}
		})
	}
}

// A land entering under an opponent's control is not landfall for you.
func TestBfz5547B1LandfallIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	leopard := b12Push(g, me.ID, "Scythe Leopard", "Creature — Cat", bfzB1Landfall[0].oracle, 1, 1)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, leopard); p != 1 || tough != 1 {
		t.Errorf("Scythe Leopard is %d/%d after an opponent's land, want 1/1", p, tough)
	}
}

// --- entry counters ----------------------------------------------------------

func TestBfz5547B1EndlessOneEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	one := b12PlayFromHand(t, g, "Endless One", "Creature — Eldrazi", bfzB1EndlessOneOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if n := counterOn(g, one, game.CounterPlusOne); n != 3 {
		t.Errorf("X = 3 put %d +1/+1 counters on it, want 3", n)
	}
}

// bfzB1ConvergeCast casts a converge creature from the active seat's hand
// under the strict mana gate, paying with exactly `mana`.
func bfzB1ConvergeCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle, mana string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, mana)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness,
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestBfz5547B1ConvergeCountsTheColoursSpent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle, mana string
		power, tough, want                 int
	}{
		{"Tajuru Stalwart", "Creature — Elf Scout Ally", "{2}{G}", bfzB1TajuruStalwartOracle, "GWU", 0, 1, 3},
		{"Tajuru Stalwart", "Creature — Elf Scout Ally", "{2}{G}", bfzB1TajuruStalwartOracle, "GCC", 0, 1, 1},
		{"Woodland Wanderer", "Creature — Elemental", "{3}{G}", bfzB1WoodlandWandererOracle, "GWUB", 2, 2, 4},
	} {
		t.Run(tc.name+" "+tc.mana, func(t *testing.T) {
			g := newCatalogGame(t)
			id := bfzB1ConvergeCast(t, g, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.mana, tc.power, tc.tough)
			if n := counterCount(g, id, game.CounterPlusOne); n != tc.want {
				t.Errorf("%d counters for %s, want %d", n, tc.mana, tc.want)
			}
		})
	}
}

// --- rally -------------------------------------------------------------------

// Each rally Ally grants its keyword to your creatures when it enters, and
// not to an opponent's.
func TestBfz5547B1RallyGrantsTheTeamItsKeyword(t *testing.T) {
	for _, tc := range bfzB1RallyGrants {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			ally := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			passPriorityAroundTable(t, g)
			if !hasEffectiveKeyword(t, g, mine, tc.keyword) {
				t.Errorf("my creature did not gain %s", tc.keyword)
			}
			if !hasEffectiveKeyword(t, g, ally, tc.keyword) {
				t.Errorf("%s did not gain %s itself", tc.name, tc.keyword)
			}
			if hasEffectiveKeyword(t, g, theirs, tc.keyword) {
				t.Errorf("an opponent's creature gained %s", tc.keyword)
			}
		})
	}
}

// Rally fires for another Ally entering under your control, and not for a
// creature that is not an Ally.
func TestBfz5547B1RallyFiresForAnotherAllyOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Chasm Guide", "Creature — Goblin Scout Ally", bfzB1ChasmGuideOracle, 3, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Broodhunter Wurm", bfzB1WurmType, bfzB1BroodhunterWurmOracle, 4, 3)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("a Wurm entering triggered rally")
	}
	castHoldCreature(t, g, "Expedition Envoy", bfzB1EnvoyType, bfzB1ExpeditionEnvoyOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("another Ally entering did not trigger rally")
	}
}

func TestBfz5547B1RallyTeamPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Tajuru Beastmaster", "Creature — Elf Warrior Ally", bfzB1TajuruBeastmasterOracle, 5, 5, 1},
		{"Tajuru Warcaller", "Creature — Elf Warrior Ally", bfzB1TajuruWarcallerOracle, 2, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, mine); p != 2+tc.n || tough != 2+tc.n {
				t.Errorf("my bear is %d/%d, want %d/%d", p, tough, 2+tc.n, 2+tc.n)
			}
			if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
				t.Errorf("their bear is %d/%d, want 2/2", p, tough)
			}
		})
	}
}
