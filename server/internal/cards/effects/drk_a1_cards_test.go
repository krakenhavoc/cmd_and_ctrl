package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// drk_a1_cards_test.go — The Dark, first batch (#5638): three vanilla
// bodies, four creatures with one printed keyword, a mana creature,
// three regenerators, six activated removal and value rows, and eight
// spells. Every seam already exists; each test plays the real catalog
// entry and asserts what the text says.

const (
	drkA1SistersOracle        = "389a9d46-d3fb-47f0-91cd-f6d487636916"
	drkA1DrownedOracle        = "3da87e58-6a4b-48c3-873c-bc654277d4df"
	drkA1DiabolicOracle       = "afd9b232-3f46-40c5-b1cf-2a7e8c955eba"
	drkA1GhostShipOracle      = "09de1136-8a8f-4539-92a2-611dcdd0790b"
	drkA1ExorcistOracle       = "2a1ca3e2-e1d1-4525-a07e-13d5ba1ac431"
	drkA1MerfolkAssassinOrc   = "7edd536c-c043-4189-aa57-83d22eed3103"
	drkA1DiggingTeamOracle    = "8408f2a1-e321-43f5-a7d1-1911eba9d706"
	drkA1ScavengerFolkOracle  = "74f361f8-a0d8-4824-9380-61469d6a19cf"
	drkA1FountainOracle       = "b906923c-9997-4da5-a05e-53eb4d2dff32"
	drkA1BookOfRassOracle     = "a5e3ca74-ec33-4679-9366-8b1a45048783"
	drkA1FissureOracle        = "c8b1e9f3-b014-4e57-b278-6d84a7e88b23"
	drkA1DustToDustOracle     = "b17eb586-0272-4d02-be75-de06ea044b6f"
	drkA1InfernoOracle        = "69e2df4e-c7f3-4c3a-be5b-1c4afb29cead"
	drkA1MarshGasOracle       = "08ae9f91-48c8-4721-a82d-063defc1b727"
	drkA1HolyLightOracle      = "8f7dd328-b367-4546-ba79-1efc1e97f2ab"
	drkA1MoraleOracle         = "8af9d623-efa3-4cbc-a903-aaec32d78b0a"
	drkA1TivadarsCrusadeOrc   = "5cf4d98e-f2f4-435b-a633-990b2e654f3a"
	drkA1RiptideOracle        = "b4541f38-23c1-4f51-a35b-0b222cdaed2c"
	drkA1MarshGoblinsOracle   = "a2c5a504-5c8f-4c1b-a8b1-4e045aee4673"
	drkA1CarnivorousPlantOrcl = "3a7478d7-b33e-4227-8cc3-b30643624d56"
)

// drkA1Vanilla are the batch's bodies with no rules text.
var drkA1Vanilla = map[string]string{
	"aa6cdcc7-f5ea-47bf-9448-1c63e36b18d1": "Squire",
	"5314e57b-107c-4478-9cdb-51d1732f9468": "Scarwood Goblins",
	"ee969637-a20e-4163-97c0-9fd5cb17b741": "Goblin Hero",
}

// drkA1Keyworded are the batch's creatures whose whole text is one
// printed keyword.
var drkA1Keyworded = map[string]struct {
	name    string
	keyword string
}{
	"45b94e3c-a905-435b-aee5-bec9239fd24c": {"Bog Imp", "flying"},
	drkA1CarnivorousPlantOrcl:              {"Carnivorous Plant", "defender"},
	"228f0afa-146b-4564-a380-71595cdf1ef4": {"Land Leeches", "first strike"},
	drkA1MarshGoblinsOracle:                {"Marsh Goblins", "swampwalk"},
}

func TestDrkA1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		drkA1SistersOracle:       "Sisters of the Flame",
		drkA1DrownedOracle:       "Drowned",
		drkA1DiabolicOracle:      "Diabolic Machine",
		drkA1GhostShipOracle:     "Ghost Ship",
		drkA1ExorcistOracle:      "Exorcist",
		drkA1MerfolkAssassinOrc:  "Merfolk Assassin",
		drkA1DiggingTeamOracle:   "Goblin Digging Team",
		drkA1ScavengerFolkOracle: "Scavenger Folk",
		drkA1FountainOracle:      "Fountain of Youth",
		drkA1BookOfRassOracle:    "Book of Rass",
		drkA1FissureOracle:       "Fissure",
		drkA1DustToDustOracle:    "Dust to Dust",
		drkA1InfernoOracle:       "Inferno",
		drkA1MarshGasOracle:      "Marsh Gas",
		drkA1HolyLightOracle:     "Holy Light",
		drkA1MoraleOracle:        "Morale",
		drkA1TivadarsCrusadeOrc:  "Tivadar's Crusade",
		drkA1RiptideOracle:       "Riptide",
	}
	for oracle, name := range drkA1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range drkA1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 25 {
		t.Fatalf("the batch lists %d cards, want 25", len(all))
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

func TestDrkA1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range drkA1Vanilla {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 || len(spec.ManaAbilities) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestDrkA1KeywordedCreaturesPrintOneKeyword(t *testing.T) {
	for oracle, k := range drkA1Keyworded {
		spec, _ := Lookup(oracle)
		if !slices.Equal(spec.PrintedKeywords, []string{k.keyword}) {
			t.Errorf("%s: printed keywords %v, want [%s]", k.name, spec.PrintedKeywords, k.keyword)
		}
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
			t.Errorf("%s declares more than its keyword", k.name)
		}
	}
}

func TestDrkA1MarshGoblinsHasSwampwalkOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblins := b12Push(g, me.ID, "Marsh Goblins", "Creature — Goblin", drkA1MarshGoblinsOracle, 1, 1)
	if !effectiveAbilitiesContain(t, g, goblins, "swampwalk") {
		t.Error("Marsh Goblins should have swampwalk on the battlefield")
	}
}

func TestDrkA1SistersOfTheFlameTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sisters := b21Push(g, me.ID, "Sisters of the Flame", "Creature — Human Shaman", drkA1SistersOracle, 2, 2, "R")
	activateManaFor(t, g, me.ID, sisters, 0, game.ManaAbilityParams{})
	red := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "R" {
			red += tok.N()
		}
	}
	if red != 1 {
		t.Errorf("pool holds %d red, want 1 (%v)", red, me.ManaPool)
	}
	if !b16Tapped(t, g, sisters) {
		t.Error("the Sisters tap as the cost")
	}
}

// Each regenerator pays its own cost and survives the next destruction
// this turn, tapped.
func TestDrkA1RegeneratorsSurviveADestroy(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		mana                   []string
		p, tough               int
	}{
		{"Drowned", "Creature — Zombie", drkA1DrownedOracle, []string{"B"}, 1, 1},
		{"Diabolic Machine", "Artifact Creature — Construct", drkA1DiabolicOracle, []string{"C", "C", "C"}, 4, 4},
		{"Ghost Ship", "Creature — Spirit", drkA1GhostShipOracle, []string{"U", "U", "U"}, 2, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, c.name, c.typeLine, c.oracle, c.p, c.tough)
			advanceToMain(t, g)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("the regeneration cost must be paid")
			}
			b06AddMana(me, c.mana...)
			b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
			g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
			if !g.Battlefield.Contains(id) {
				t.Fatal("the shield should replace the destruction")
			}
			if !b16Tapped(t, g, id) {
				t.Error("a regenerated creature is tapped")
			}
		})
	}
}

func TestDrkA1GhostShipHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	ship := b12Push(g, g.Seats[0].ID, "Ghost Ship", "Creature — Spirit", drkA1GhostShipOracle, 2, 4)
	if !effectiveAbilitiesContain(t, g, ship, "flying") {
		t.Error("Ghost Ship should have flying on the battlefield")
	}
}

func TestDrkA1ExorcistDestroysOnlyABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exorcist := b21Push(g, me.ID, "Exorcist", "Creature — Human Cleric", drkA1ExorcistOracle, 1, 1, "W")
	green := b16Creature(g, opp.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	black := b16Creature(g, opp.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	advanceToMain(t, g)
	b06AddMana(me, "W", "C")
	if err := g.ActivateCatalogAbility(me.ID, exorcist, 0, game.ActivateAbilityParams{Targets: b16TargetCard(green)}); err == nil {
		t.Fatal("a green creature is not a legal target")
	}
	b16Activate(t, g, me.ID, exorcist, 0, game.ActivateAbilityParams{Targets: b16TargetCard(black)})
	if g.Battlefield.Contains(black) || !opp.Graveyard.Contains(black) {
		t.Error("the black creature should be destroyed")
	}
	if !b16Tapped(t, g, exorcist) {
		t.Error("the Exorcist taps as part of the cost")
	}
}

func TestDrkA1MerfolkAssassinDestroysOnlyAnIslandwalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := b21Push(g, me.ID, "Merfolk Assassin", "Creature — Merfolk Assassin", drkA1MerfolkAssassinOrc, 1, 2, "U")
	plain := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	walker := pushTribalCreature(g, opp.ID, "Walker", "Creature — Merfolk", 2, 2, "islandwalk")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Targets: b16TargetCard(plain)}); err == nil {
		t.Fatal("a creature without islandwalk is not a legal target")
	}
	b16Activate(t, g, me.ID, assassin, 0, game.ActivateAbilityParams{Targets: b16TargetCard(walker)})
	if g.Battlefield.Contains(walker) || !opp.Graveyard.Contains(walker) {
		t.Error("the islandwalker should be destroyed")
	}
}

func TestDrkA1GoblinDiggingTeamSacrificesToDestroyAWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	team := b21Push(g, me.ID, "Goblin Digging Team", "Creature — Goblin", drkA1DiggingTeamOracle, 1, 1, "R")
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	wall := b16Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8, "R")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, team, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a non-Wall is not a legal target")
	}
	b16Activate(t, g, me.ID, team, 0, game.ActivateAbilityParams{Targets: b16TargetCard(wall)})
	if g.Battlefield.Contains(wall) || !opp.Graveyard.Contains(wall) {
		t.Error("the Wall should be destroyed")
	}
	if g.Battlefield.Contains(team) || !me.Graveyard.Contains(team) {
		t.Error("the Digging Team is sacrificed as part of the cost")
	}
}

func TestDrkA1ScavengerFolkSacrificesToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	folk := b21Push(g, me.ID, "Scavenger Folk", "Creature — Human", drkA1ScavengerFolkOracle, 1, 1, "G")
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	advanceToMain(t, g)
	b06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, folk, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a nonartifact creature is not a legal target")
	}
	b16Activate(t, g, me.ID, folk, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if g.Battlefield.Contains(folk) || !me.Graveyard.Contains(folk) {
		t.Error("Scavenger Folk is sacrificed as part of the cost")
	}
}

func TestDrkA1FountainOfYouthGainsOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fountain := b12Push(g, me.ID, "Fountain of Youth", "Artifact", drkA1FountainOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, fountain, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{2} must be paid")
	}
	b06AddMana(me, "C", "C")
	life := me.Life
	b16Activate(t, g, me.ID, fountain, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if !b16Tapped(t, g, fountain) {
		t.Error("the Fountain taps as part of the cost")
	}
}

func TestDrkA1BookOfRassPaysTwoLifeForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	book := b12Push(g, me.ID, "Book of Rass", "Artifact — Book", drkA1BookOfRassOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, book, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// No tap: it can be activated again, but not without 2 life to pay.
	me.Life = 1
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("at 1 life there is no 2 life to pay")
	}
}

func TestDrkA1FissureDestroysALandOrAnUnregenerableCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Fissure", "Instant", drkA1FissureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("a noncreature, nonland artifact is not a legal target")
	}
	castCatalogSpell(t, g, "Fissure", "Instant", drkA1FissureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the land should be destroyed")
	}
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Fissure", "Instant", drkA1FissureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature can't be regenerated and should be destroyed")
	}
}

func TestDrkA1DustToDustExilesTwoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	b := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Dust to Dust", "Sorcery", drkA1DustToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("a nonartifact creature is not a legal target")
	}
	castCatalogSpell(t, g, "Dust to Dust", "Sorcery", drkA1DustToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
			t.Errorf("%v should be in exile", id)
		}
	}
}

func TestDrkA1InfernoHitsEveryCreatureAndEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	six := b12Creature(g, opp.ID, "Six", "Creature — Giant", 6, 6)
	seven := b12Creature(g, me.ID, "Seven", "Creature — Giant", 7, 7)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Inferno", "Instant", drkA1InfernoOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(six) {
		t.Error("a 6-toughness creature should die")
	}
	if !g.Battlefield.Contains(seven) {
		t.Error("a 7-toughness creature should survive")
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-6 {
			t.Errorf("seat %v life %d → %d, want -6", p.ID, lives[p.ID], p.Life)
		}
	}
}

func TestDrkA1MarshGasShrinksEveryCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 2)
	castCatalogSpell(t, g, "Marsh Gas", "Instant", drkA1MarshGasOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 1 || tt != 3 {
		t.Errorf("mine is %d/%d, want 1/3", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != -1 || tt != 2 {
		t.Errorf("theirs is %d/%d, want -1/2", p, tt)
	}
}

func TestDrkA1HolyLightSparesWhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	white := b16Creature(g, opp.ID, "White Bear", "Creature — Bear", 2, 2, "W")
	green := b16Creature(g, opp.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	elf := b16Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1, "G")
	castCatalogSpell(t, g, "Holy Light", "Instant", drkA1HolyLightOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, white); p != 2 || tt != 2 {
		t.Errorf("the white creature is %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, green); p != 1 || tt != 1 {
		t.Errorf("the green creature is %d/%d, want 1/1", p, tt)
	}
	if g.Battlefield.Contains(elf) {
		t.Error("a nonwhite 1/1 should die")
	}
}

func TestDrkA1MoralePumpsOnlyAttackingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attacker {
			g.Battlefield.Cards[i].AttackingTarget = opp.ID
		}
	}
	castCatalogSpell(t, g, "Morale", "Instant", drkA1MoraleOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, attacker); p != 3 || tt != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3", p, tt)
	}
	if p, tt := ptOf(t, g, idle); p != 2 || tt != 2 {
		t.Errorf("the creature not attacking is %d/%d, want 2/2", p, tt)
	}
}

func TestDrkA1TivadarsCrusadeDestroysEveryGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Goblin Raider", "Creature — Goblin Warrior", 1, 1)
	mine := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	kindred := b12Permanent(g, opp.ID, "Goblin Shrine", "Kindred Enchantment — Goblin")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Tivadar's Crusade", "Sorcery", drkA1TivadarsCrusadeOrc, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, mine, kindred} {
		if g.Battlefield.Contains(id) {
			t.Errorf("Goblin %v should be destroyed", id)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a non-Goblin survives")
	}
}

func TestDrkA1RiptideTapsEveryBlueCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 1, 1, "U")
	theirs := b16Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1, "U")
	green := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Riptide", "Instant", drkA1RiptideOracle, nil)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, mine) || !b16Tapped(t, g, theirs) {
		t.Error("every blue creature should be tapped, whoever controls it")
	}
	if b16Tapped(t, g, green) {
		t.Error("a green creature stays untapped")
	}
}
