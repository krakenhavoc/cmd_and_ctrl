package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b13_test.go — Adventures in the Forgotten Realms, batch 13 (S82,
// #5499): six Classes, Monk of the Open Hand and Circle of the Moon Druid.
// Each test plays the real catalog entry and asserts what its text says.

const (
	afrB13DruidOracle   = "dcbcbf42-4654-487a-acad-21f2606d229b"
	afrB13ClericOracle  = "f0ead7f5-5c82-40fd-8e89-3e3c5d31fbad"
	afrB13PaladinOracle = "ac9d52b2-78a5-4b17-8aaf-1c1cb936a508"
	afrB13RangerOracle  = "d1586808-efc2-469a-8384-ba4acf231e69"
	afrB13WarlockOracle = "766c2c67-d1ee-44ff-b9bb-355361646ef1"
	afrB13FighterOracle = "8a280ba8-6b7b-46dd-9c2b-6d0027cfff30"
	afrB13MonkOracle    = "2d3998fc-a68a-4923-a456-b3e43ab6decb"
	afrB13CircleOracle  = "90569b99-8b63-4d2e-b7de-b578b8adbdcb"
	afrB13ClassType     = "Enchantment — Class"
)

func TestAfrB13EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB13DruidOracle:   "Druid Class",
		afrB13ClericOracle:  "Cleric Class",
		afrB13PaladinOracle: "Paladin Class",
		afrB13RangerOracle:  "Ranger Class",
		afrB13WarlockOracle: "Warlock Class",
		afrB13FighterOracle: "Fighter Class",
		afrB13MonkOracle:    "Monk of the Open Hand",
		afrB13CircleOracle:  "Circle of the Moon Druid",
	} {
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

// afrB13SetLevel sets a Class's level the way a level-up resolves: the
// designation changes and the level event is announced.
func afrB13SetLevel(t *testing.T, g *game.Game, id uuid.UUID, level int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.SetClassLevelForEffect(id, level) })
	if err != nil {
		t.Fatalf("SetClassLevelForEffect(%d): %v", level, err)
	}
}

// afrB13Named lists a controller's battlefield cards with a name.
func afrB13Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Monk of the Open Hand ------------------------------------------------

func TestAfrB13MonkGrowsOnTheSecondSpellEachTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.B12Push(g, me.ID, "Monk of the Open Hand", "Creature — Elf Monk", afrB13MonkOracle, 1, 1)
	effectstest.ToMainForCost(t, g)
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	if n := effects.CounterCountOn(g, monk, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after the first spell, want 0", n)
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "Second")
	if n := effects.CounterCountOn(g, monk, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after the second spell, want 1", n)
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "Third")
	if n := effects.CounterCountOn(g, monk, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after the third spell, want 1 (only the second counts)", n)
	}
}

// --- Circle of the Moon Druid ---------------------------------------------

func TestAfrB13CircleIsABear4x2OnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := effectstest.B12Push(g, me.ID, "Circle of the Moon Druid", "Creature — Human Elf Druid", afrB13CircleOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, druid); p != 4 || tough != 2 {
		t.Errorf("on your turn the Druid is %d/%d, want 4/2", p, tough)
	}
	c := effectstest.LayeredCard(t, g, druid)
	if !c.HasSubtype("Bear") || c.HasSubtype("Human") || c.HasSubtype("Elf") || c.HasSubtype("Druid") || !c.IsCreature() {
		t.Errorf("on your turn the type line is %q, want a Bear that lost its other creature types", c.TypeLine)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, druid); p != 2 || tough != 4 {
		t.Errorf("on another player's turn the Druid is %d/%d, want 2/4", p, tough)
	}
	c = effectstest.LayeredCard(t, g, druid)
	if c.HasSubtype("Bear") || !c.HasSubtype("Human") || !c.HasSubtype("Elf") || !c.HasSubtype("Druid") {
		t.Errorf("on another player's turn the type line is %q, want Human Elf Druid", c.TypeLine)
	}
}

// --- Druid Class ------------------------------------------------------------

func TestAfrB13DruidClassLandfallGainsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushClass(g, me.ID, "Druid Class", afrB13ClassType, afrB13DruidOracle)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.B31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestAfrB13DruidClassLevelTwoAllowsAnAdditionalLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushClass(g, me.ID, "Druid Class", afrB13ClassType, afrB13DruidOracle)
	effectstest.ToMain(t, g)
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the ordinary land drop: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err == nil {
		t.Fatal("a second land at level 1 was allowed")
	}
	afrB13SetLevel(t, g, id, 2)
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the additional land at level 2 was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land at level 2 was allowed")
	}
}

func TestAfrB13DruidClassLevelThreeAnimatesALandThatKeepsCounting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushClass(g, me.ID, "Druid Class", afrB13ClassType, afrB13DruidOracle)
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	other := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	theirs := effectstest.B12Permanent(g, g.Seats[1].ID, "Plains", "Basic Land — Plains")

	afrB13SetLevel(t, g, id, 2)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		t.Fatal("the level 3 ability asked for a target at level 2")
	}
	afrB13SetLevel(t, g, id, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at level 3")
	}
	if slices.Contains(pick.PickTargetCards, theirs) || slices.Contains(pick.PickTargetCards, id) || !slices.Contains(pick.PickTargetCards, target) {
		t.Errorf("legal set %v: want lands you control only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.LayeredCard(t, g, target)
	if !c.IsLand() || !c.IsCreature() {
		t.Fatalf("the target's type line is %q, want a creature that is still a land", c.TypeLine)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, target); p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d with three lands, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, target, "haste") {
		t.Error("the land has no haste")
	}
	if c := effectstest.LayeredCard(t, g, other); c.IsCreature() {
		t.Error("a land that was not targeted became a creature")
	}
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	if p, tough := effectstest.EffectivePTOf(t, g, target); p != 4 || tough != 4 {
		t.Errorf("a fourth land: the animated land is %d/%d, want 4/4 (counted live)", p, tough)
	}
}

// --- Cleric Class -----------------------------------------------------------

func TestAfrB13ClericClassGainsOneMoreLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushClass(g, me.ID, "Cleric Class", afrB13ClassType, afrB13ClericOracle)
	effectstest.AdvanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2)
	})
	if me.Life != life+3 {
		t.Errorf("your gain of 2: life %d -> %d, want +3", life, me.Life)
	}
	if opp.Life != oppLife+2 {
		t.Errorf("an opponent's gain of 2: life %d -> %d, want +2", oppLife, opp.Life)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the level 2 trigger asked for a target at level 1")
	}
}

func TestAfrB13ClericClassLevelTwoPutsACounterWheneverYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushClass(g, me.ID, "Cleric Class", afrB13ClassType, afrB13ClericOracle)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	afrB13SetLevel(t, g, id, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt after gaining life at level 2")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != bear {
		t.Errorf("legal set %v: want only the creature you control", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.CounterCountOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters on the Bear, want 1", n)
	}
}

func TestAfrB13ClericClassLevelThreeReturnsACreatureAndGainsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushClass(g, me.ID, "Cleric Class", afrB13ClassType, afrB13ClericOracle)
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.PushToGraveyard(g, me, game.Card{Name: "Grave Wall", TypeLine: "Creature — Wall", ManaCost: "{1}{W}", Power: 0, Toughness: 3})
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Grave Spell", TypeLine: "Sorcery", ManaCost: "{1}"})
	life := me.Life
	afrB13SetLevel(t, g, id, 2)
	afrB13SetLevel(t, g, id, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at level 3")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != dead {
		t.Errorf("legal set %v: want only the creature card in your graveyard", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	// The returned creature is a creature you control: the level 2 trigger
	// (the gain was life gained) asks for its own target.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, p.PickTargetCards[0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	if walls := afrB13Named(g, "Grave Wall"); len(walls) != 1 {
		t.Fatalf("%d Grave Walls on the battlefield, want 1", len(walls))
	}
	// Its toughness 3, plus the level 1 replacement's 1.
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (toughness 3 plus 1)", life, me.Life)
	}
}

// --- Paladin Class ----------------------------------------------------------

func TestAfrB13PaladinClassTaxesOpponentsSpellsOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushClass(g, me.ID, "Paladin Class", afrB13ClassType, afrB13PaladinOracle)
	if got := effectstest.PriceInHand(t, g, opp, "Their Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("an opponent's spell during your turn: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("your own spell: %d, want 1", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if got := effectstest.PriceInHand(t, g, opp, "Their Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("an opponent's spell during their own turn: %d, want 1", got)
	}
}

func TestAfrB13PaladinClassLevelTwoPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushClass(g, me.ID, "Paladin Class", afrB13ClassType, afrB13PaladinOracle)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("at level 1 the Bear has power %d, want 2", p)
	}
	afrB13SetLevel(t, g, id, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your Bear is %d/%d at level 2, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestAfrB13PaladinClassLevelThreeGrowsOneAttackerForEachOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushClass(g, me.ID, "Paladin Class", afrB13ClassType, afrB13PaladinOracle)
	afrB13SetLevel(t, g, id, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 1, 1)
	stay := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if slices.Contains(pick.PickTargetCards, stay) || !slices.Contains(pick.PickTargetCards, a) {
		t.Errorf("legal set %v: want attacking creatures only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, a)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("three attackers triggered the Class more than once")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Level 2's anthem still applies at level 3 (+1/+1), and the trigger adds
	// +1/+1 for each of the two other attackers: 1/1 becomes 4/4.
	if p, tough := effectstest.EffectivePTOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "double strike") {
		t.Error("the target did not gain double strike")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, b); p != 2 || effectstest.HasEffectiveKeyword(t, g, b, "double strike") {
		t.Error("another attacker was changed")
	}
}

// --- Ranger Class -----------------------------------------------------------

func TestAfrB13RangerClassMakesAWolfWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Ranger Class", afrB13ClassType, afrB13RangerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	wolves := effectstest.SocB8Named(g, me.ID, "Wolf")
	if len(wolves) != 1 {
		t.Fatalf("%d Wolves, want 1", len(wolves))
	}
	if w := wolves[0]; w.Power != 2 || w.Toughness != 2 || len(w.Colors) != 1 || w.Colors[0] != "G" {
		t.Errorf("the Wolf is %d/%d %v, want a 2/2 green Wolf", w.Power, w.Toughness, w.Colors)
	}
}

func TestAfrB13RangerClassLevelTwoPutsACounterOnAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushClass(g, me.ID, "Ranger Class", afrB13ClassType, afrB13RangerOracle)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	afrB13SetLevel(t, g, id, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	effectstest.PickCard(t, g, me.ID, b)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("two attackers triggered the Class twice")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.CounterCountOn(g, b, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", n)
	}
	if n := effects.CounterCountOn(g, a, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters on the other attacker, want 0", n)
	}
}

func TestAfrB13RangerClassLevelOneDoesNotTriggerOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushClass(g, me.ID, "Ranger Class", afrB13ClassType, afrB13RangerOracle)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the level 2 ability triggered at level 1")
	}
}

func TestAfrB13RangerClassLevelThreeCastsCreaturesFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Ranger Class", afrB13ClassType, afrB13RangerOracle)
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.TopVisibleTo(g, me.ID, me.ID) {
		t.Error("the top card is visible before level 3")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("cast a creature off the top before level 3")
	}
	afrB13SetLevel(t, g, id, 2)
	if effectstest.TopVisibleTo(g, me.ID, me.ID) {
		t.Error("the top card is visible at level 2")
	}
	afrB13SetLevel(t, g, id, 3)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to the Class's controller at level 3")
	}
	sorcery := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("cast a sorcery off the top; only creature spells may be")
	}
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("played a land off the top; only creature spells may be cast")
	}
	bear = effectstest.SeedTopOfLibrary(me, game.Card{Name: "Second Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a creature off the top at level 3: %v", err)
	}
}

// --- Warlock Class ----------------------------------------------------------

func TestAfrB13WarlockClassDrainsOneAtYourEndStepIfACreatureDied(t *testing.T) {
	for _, tc := range []struct {
		name   string
		whose  int
		drains bool
	}{
		{"your creature died", 0, true},
		{"an opponent's creature died", 1, true},
		{"no creature died", -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.PushClass(g, me.ID, "Warlock Class", afrB13ClassType, afrB13WarlockOracle)
			effectstest.AdvanceToMain(t, g)
			if tc.whose >= 0 {
				owner := g.Seats[tc.whose]
				bear := effectstest.B12Creature(g, owner.ID, "Bear", "Creature — Bear", 2, 2)
				effectstest.KillCreature(t, g, owner.ID, bear)
				effectstest.PassPriorityAroundTable(t, g)
			}
			before := effectstest.LifeOfOpponents(g)
			myLife := me.Life
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			for i, life := range effectstest.LifeOfOpponents(g) {
				want := before[i]
				if tc.drains {
					want--
				}
				if life != want {
					t.Errorf("opponent %d life %d -> %d, want %d", i, before[i], life, want)
				}
			}
			if me.Life != myLife {
				t.Errorf("your life %d -> %d, want unchanged", myLife, me.Life)
			}
		})
	}
}

func TestAfrB13WarlockClassLevelTwoTakesOneOfTheTopThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushClass(g, me.ID, "Warlock Class", afrB13ClassType, afrB13WarlockOracle)
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	afrB13SetLevel(t, g, id, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 || slices.Contains(pick.ChooseCards, ids[3]) {
		t.Fatalf("offered %v in [%d,%d], want exactly one of the top three", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, rest := range []uuid.UUID{ids[0], ids[2]} {
		if !me.Graveyard.Contains(rest) {
			t.Error("a card that was not chosen is not in the graveyard")
		}
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card left the library")
	}
}

func TestAfrB13WarlockClassLevelThreeDrainsTheLifeTheyLostThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushClass(g, me.ID, "Warlock Class", afrB13ClassType, afrB13WarlockOracle)
	effectstest.AdvanceToMain(t, g)
	opp1, opp2 := g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp1.ID, -3) })
	// Level 2 has no end-step drain.
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	afrB13SetLevel(t, g, id, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:1]); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	afrB13SetLevel(t, g, id, 3)
	l1, l2, my := opp1.Life, opp2.Life, me.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp1.Life != l1-3 {
		t.Errorf("the opponent who lost 3: life %d -> %d, want -3 more", l1, opp1.Life)
	}
	if opp2.Life != l2 {
		t.Errorf("an opponent who lost nothing: life %d -> %d, want unchanged", l2, opp2.Life)
	}
	if me.Life != my {
		t.Errorf("your life %d -> %d, want unchanged", my, me.Life)
	}
}

// --- Fighter Class ----------------------------------------------------------

func TestAfrB13FighterClassSearchesForAnEquipmentWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Test Sword", "Artifact — Equipment"),
		effectstest.SearchTestLand("Test Bauble", "Artifact"),
	)
	effectstest.CastCatalogSpell(t, g, "Fighter Class", afrB13ClassType, afrB13FighterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Class's entry asked for no search")
	}
	for _, bad := range []string{"Forest", "Test Bauble"} {
		if effectstest.SearchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("%s was offered; only an Equipment card may be found", bad)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Test Sword")
	if _, ok := effectstest.HandCardNamed(me, "Test Sword"); !ok {
		t.Error("the Equipment is not in hand")
	}
}

func TestAfrB13FighterClassLevelTwoMakesEquipCostTwoLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushClass(g, me.ID, "Fighter Class", afrB13ClassType, afrB13FighterOracle)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
		t.Errorf("equip at level 1: %s, want {3}", got)
	}
	afrB13SetLevel(t, g, id, 2)
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{1}" {
		t.Errorf("equip at level 2: %s, want {1}", got)
	}
}

func TestAfrB13FighterClassLevelThreeForcesABlockThisCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	id := effectstest.PushClass(g, me.ID, "Fighter Class", afrB13ClassType, afrB13FighterOracle)
	afrB13SetLevel(t, g, id, 3)
	hunter := effectstest.ReqCreature(g, me.ID, "Hunter Bear", "", 2, 2, "haste")
	victim := effectstest.ReqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	free := effectstest.ReqCreature(g, opp.ID, "Free Bear", "", 2, 2)

	effectstest.AttackAndTrigger(t, g, hunter)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever a creature you control attacks")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.BlocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != hunter {
		t.Fatalf("records = %+v, want one naming the attacker", recs)
	}
	effectstest.OnToBlockers(t, g)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got == "" {
		t.Error("the refusal has no sentence")
	}
	if err := effectstest.ReqBlock(t, g, hunter, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	_ = free
	for i := 0; i < 12 && g.Turn.Step != game.StepPostcombatMain; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if recs := effectstest.BlocksAttackerRecords(g); len(recs) != 0 {
		t.Errorf("the requirement outlived the combat: %+v", recs)
	}
}

func TestAfrB13FighterClassLevelTwoDoesNotForceBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	id := effectstest.PushClass(g, me.ID, "Fighter Class", afrB13ClassType, afrB13FighterOracle)
	afrB13SetLevel(t, g, id, 2)
	hunter := effectstest.ReqCreature(g, me.ID, "Hunter Bear", "", 2, 2, "haste")
	effectstest.ReqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	effectstest.AttackAndTrigger(t, g, hunter)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the level 3 ability triggered at level 2")
	}
}
