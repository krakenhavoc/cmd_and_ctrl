package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB104YisanOracle     = "7f4b0233-a2b0-41a2-8448-2df1f845148c"
	gapB104AnaxOracle      = "e23cd06d-1360-4e5d-aecc-adf902bede69"
	gapB104AvatarOracle    = "2fc6305e-e75e-478a-826b-49f655ed832c"
	gapB104AspirantOracle  = "6db51087-6995-4ed4-978a-3d17d60019af"
	gapB104SatyrTokenType  = "Token Creature — Satyr"
	gapB104RedCreatureCost = "{1}{R}{R}"
)

func TestGapB104CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB104YisanOracle:    "Yisan, the Wanderer Bard",
		gapB104AnaxOracle:     "Anax, Hardened in the Forge",
		gapB104AvatarOracle:   "Avatar of Slaughter",
		gapB104AspirantOracle: "Blood Aspirant",
	} {
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

// Yisan: each activation puts a verse counter first, then fetches a creature
// at exactly that mana value; the second one looks for mana value 2.
func TestGapB104YisanFetchesAtTheNumberOfVerseCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yisan := pushCatalogPermanent(g, me.ID, "Yisan, the Wanderer Bard", "Legendary Creature — Human Rogue Bard", gapB104YisanOracle, false)
	seedSearchLibrary(me,
		game.Card{Name: "One Drop", TypeLine: "Creature — Elf", ManaCost: "{G}"},
		game.Card{Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}"},
		game.Card{Name: "One Rock", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	apaMana(me, "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if n := countersOn(g, yisan, "verse"); n != 1 {
		t.Fatalf("%d verse counters after the cost, want 1", n)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Two Drop") != uuid.Nil || searchOptionNamed(g, c, "One Rock") != uuid.Nil || searchOptionNamed(g, c, "Three Drop") != uuid.Nil {
		t.Fatal("only a creature card with mana value 1 is offered")
	}
	answerSearchNamed(t, g, me.ID, "One Drop")
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "One Drop") == uuid.Nil {
		t.Fatal("the creature is not on the battlefield")
	}
	untap(g, yisan)
	apaMana(me, "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	c = searchChoiceFor(g, me.ID)
	if c == nil || searchOptionNamed(g, c, "Two Drop") == uuid.Nil || searchOptionNamed(g, c, "One Rock") != uuid.Nil {
		t.Fatal("the second search should offer the mana value 2 creature only")
	}
	answerSearchNamed(t, g, me.ID, "Two Drop")
	if n := countersOn(g, yisan, "verse"); n != 2 {
		t.Errorf("%d verse counters, want 2", n)
	}
}

func TestGapB104YisanCostNeedsMana(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yisan := pushCatalogPermanent(g, me.ID, "Yisan, the Wanderer Bard", "Legendary Creature — Human Rogue Bard", gapB104YisanOracle, false)
	apaMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for two mana")
	}
	if n := countersOn(g, yisan, "verse"); n != 0 {
		t.Errorf("a refused activation left %d verse counters", n)
	}
}

// Anax: power is the devotion to red, and a nontoken creature of mine
// dying (Anax himself too) leaves a Satyr that can't block, two if its
// power was 4 or more.
func TestGapB104AnaxPowerIsDevotionToRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anax := apaPush(g, me.ID, me.ID, game.Card{Name: "Anax, Hardened in the Forge", OracleID: gapB104AnaxOracle,
		TypeLine: "Legendary Enchantment Creature — Demigod", ManaCost: gapB104RedCreatureCost, Power: 0, Toughness: 3})
	if got := apaLive(g, anax).PowerForComparison(); got != 2 {
		t.Fatalf("Anax's own {R}{R} make power %d, want 2", got)
	}
	apaPush(g, me.ID, me.ID, game.Card{Name: "Red Thing", TypeLine: "Artifact", ManaCost: "{R}"})
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Red Thing", TypeLine: "Artifact", ManaCost: "{R}{R}"})
	if got := apaLive(g, anax).PowerForComparison(); got != 3 {
		t.Fatalf("power %d with three red symbols of mine, want 3 (an opponent's do not count)", got)
	}
}

func TestGapB104AnaxMakesSatyrsForNontokenCreaturesAndTwoForBigOnes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, me.ID, "Anax, Hardened in the Forge", "Legendary Enchantment Creature — Demigod", gapB104AnaxOracle, 2, 3)
	token := b12Creature(g, me.ID, "Saproling", "Token Creature — Saproling", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	giant := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)

	midB17Bolt(t, g, rprB30Ref(token))
	midB17Bolt(t, g, rprB30Ref(theirs))
	if n := countBattlefieldByName(g, "Satyr"); n != 0 {
		t.Fatalf("%d Satyrs after a token and an opponent's creature died, want 0", n)
	}
	midB17Bolt(t, g, rprB30Ref(bear))
	if n := b27CountNamed(g, me.ID, "Satyr"); n != 1 {
		t.Fatalf("%d Satyrs after my 2/2 died, want 1", n)
	}
	b27Kill(g, giant)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Satyr"); n != 3 {
		t.Fatalf("%d Satyrs after my 4/4 died, want 1 + 2 = 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Satyr" {
			continue
		}
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" {
			t.Errorf("Satyr is %d/%d %v, want a 1/1 red", c.Power, c.Toughness, c.Colors)
		}
	}
	// "This token can't block."
	if !satyrCantBlock(t, g) {
		t.Error("a Satyr can block")
	}
}

func satyrCantBlock(t *testing.T, g *game.Game) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Satyr" {
			return auraRestrictions(t, g, c.InstanceID)&game.CantBlock != 0
		}
	}
	return false
}

func TestGapB104AnaxTriggersForItselfDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anax := b12Push(g, me.ID, "Anax, Hardened in the Forge", "Legendary Enchantment Creature — Demigod", gapB104AnaxOracle, 2, 1)
	midB17Bolt(t, g, rprB30Ref(anax))
	if n := b27CountNamed(g, me.ID, "Satyr"); n != 1 {
		t.Fatalf("%d Satyrs after Anax died, want 1", n)
	}
}

// Avatar of Slaughter: every creature has double strike and must attack.
func TestGapB104AvatarGivesEveryCreatureDoubleStrikeAndForcesAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Avatar of Slaughter", "Creature — Avatar", gapB104AvatarOracle, 8, 8)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	theirs := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	for _, id := range []uuid.UUID{bear, theirs} {
		if !hasEffectiveKeyword(t, g, id, "double strike") {
			t.Errorf("%s has no double strike", id)
		}
	}
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with a haste Bear at home = %v, want an attack-requirement refusal", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

// Blood Aspirant grows for each permanent I sacrifice and pings for a
// sacrifice, keeping the target from blocking.
func TestGapB104BloodAspirantGrowsWhenYouSacrificeAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aspirant := b12Push(g, me.ID, "Blood Aspirant", "Creature — Satyr Berserker", gapB104AspirantOracle, 1, 1)
	rock := b12Permanent(g, me.ID, "Test Rock", "Artifact")
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, aspirant, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after sacrificing an artifact, want 1", n)
	}
}

func TestGapB104BloodAspirantPingsAndStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	aspirant := pushCatalogPermanent(g, me.ID, "Blood Aspirant", "Creature — Satyr Berserker", gapB104AspirantOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := b12Permanent(g, me.ID, "Test Rock", "Artifact")
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	apaMana(me, "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, aspirant, 0, game.ActivateAbilityParams{Strict: true,
		Targets: cardTarget(wall), SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact was sacrificed for \"a creature or enchantment\"")
	}
	me.ManaPool = game.ManaPool{}
	apaMana(me, "R", "C")
	apaActivate(t, g, me, aspirant, 0, game.ActivateAbilityParams{Targets: cardTarget(wall), SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	if n := countersOn(g, aspirant, game.CounterPlusOne); n != 1 {
		t.Errorf("the Aspirant has %d counters after a sacrifice, want 1", n)
	}
	if dmg := b12Card(t, g, wall).DamageMarked; dmg != 1 {
		t.Errorf("the Wall has %d damage, want 1", dmg)
	}
	cantBlock := auraRestrictions(t, g, wall)&game.CantBlock != 0
	if !cantBlock {
		t.Error("the Wall can still block")
	}
}
