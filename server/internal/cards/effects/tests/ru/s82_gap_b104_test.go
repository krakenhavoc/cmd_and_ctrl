package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// Yisan: each activation puts a verse counter first, then fetches a creature
// at exactly that mana value; the second one looks for mana value 2.
func TestGapB104YisanFetchesAtTheNumberOfVerseCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yisan := effectstest.PushCatalogPermanent(g, me.ID, "Yisan, the Wanderer Bard", "Legendary Creature — Human Rogue Bard", gapB104YisanOracle, false)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "One Drop", TypeLine: "Creature — Elf", ManaCost: "{G}"},
		game.Card{Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}"},
		game.Card{Name: "One Rock", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	effectstest.ApaMana(me, "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if n := effectstest.CountersOn(g, yisan, "verse"); n != 1 {
		t.Fatalf("%d verse counters after the cost, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Two Drop") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "One Rock") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Three Drop") != uuid.Nil {
		t.Fatal("only a creature card with mana value 1 is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "One Drop")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "One Drop") == uuid.Nil {
		t.Fatal("the creature is not on the battlefield")
	}
	effectstest.Untap(g, yisan)
	effectstest.ApaMana(me, "C", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || effectstest.SearchOptionNamed(g, c, "Two Drop") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "One Rock") != uuid.Nil {
		t.Fatal("the second search should offer the mana value 2 creature only")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Two Drop")
	if n := effectstest.CountersOn(g, yisan, "verse"); n != 2 {
		t.Errorf("%d verse counters, want 2", n)
	}
}

func TestGapB104YisanCostNeedsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yisan := effectstest.PushCatalogPermanent(g, me.ID, "Yisan, the Wanderer Bard", "Legendary Creature — Human Rogue Bard", gapB104YisanOracle, false)
	effectstest.ApaMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, yisan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for two mana")
	}
	if n := effectstest.CountersOn(g, yisan, "verse"); n != 0 {
		t.Errorf("a refused activation left %d verse counters", n)
	}
}

// Anax: power is the devotion to red, and a nontoken creature of mine
// dying (Anax himself too) leaves a Satyr that can't block, two if its
// power was 4 or more.
func TestGapB104AnaxPowerIsDevotionToRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anax := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Anax, Hardened in the Forge", OracleID: gapB104AnaxOracle,
		TypeLine: "Legendary Enchantment Creature — Demigod", ManaCost: gapB104RedCreatureCost, Power: 0, Toughness: 3})
	if got := effectstest.ApaLive(g, anax).PowerForComparison(); got != 2 {
		t.Fatalf("Anax's own {R}{R} make power %d, want 2", got)
	}
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red Thing", TypeLine: "Artifact", ManaCost: "{R}"})
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Red Thing", TypeLine: "Artifact", ManaCost: "{R}{R}"})
	if got := effectstest.ApaLive(g, anax).PowerForComparison(); got != 3 {
		t.Fatalf("power %d with three red symbols of mine, want 3 (an opponent's do not count)", got)
	}
}

func TestGapB104AnaxMakesSatyrsForNontokenCreaturesAndTwoForBigOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Anax, Hardened in the Forge", "Legendary Enchantment Creature — Demigod", gapB104AnaxOracle, 2, 3)
	token := effectstest.B12Creature(g, me.ID, "Saproling", "Token Creature — Saproling", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	giant := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)

	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(token))
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(theirs))
	if n := effectstest.CountBattlefieldByName(g, "Satyr"); n != 0 {
		t.Fatalf("%d Satyrs after a token and an opponent's creature died, want 0", n)
	}
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(bear))
	if n := effectstest.B27CountNamed(g, me.ID, "Satyr"); n != 1 {
		t.Fatalf("%d Satyrs after my 2/2 died, want 1", n)
	}
	effectstest.B27Kill(g, giant)
	effectstest.MonSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B27CountNamed(g, me.ID, "Satyr"); n != 3 {
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
			return effectstest.AuraRestrictions(t, g, c.InstanceID)&game.CantBlock != 0
		}
	}
	return false
}

func TestGapB104AnaxTriggersForItselfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anax := effectstest.B12Push(g, me.ID, "Anax, Hardened in the Forge", "Legendary Enchantment Creature — Demigod", gapB104AnaxOracle, 2, 1)
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(anax))
	if n := effectstest.B27CountNamed(g, me.ID, "Satyr"); n != 1 {
		t.Fatalf("%d Satyrs after Anax died, want 1", n)
	}
}

// Avatar of Slaughter: every creature has double strike and must attack.
func TestGapB104AvatarGivesEveryCreatureDoubleStrikeAndForcesAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Avatar of Slaughter", "Creature — Avatar", gapB104AvatarOracle, 8, 8)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	theirs := effectstest.ReqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	for _, id := range []uuid.UUID{bear, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
			t.Errorf("%s has no double strike", id)
		}
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aspirant := effectstest.B12Push(g, me.ID, "Blood Aspirant", "Creature — Satyr Berserker", gapB104AspirantOracle, 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Test Rock", "Artifact")
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, aspirant, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after sacrificing an artifact, want 1", n)
	}
}

func TestGapB104BloodAspirantPingsAndStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	aspirant := effectstest.PushCatalogPermanent(g, me.ID, "Blood Aspirant", "Creature — Satyr Berserker", gapB104AspirantOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Test Rock", "Artifact")
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.ApaMana(me, "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, aspirant, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTarget(wall), SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact was sacrificed for \"a creature or enchantment\"")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "R", "C")
	effectstest.ApaActivate(t, g, me, aspirant, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(wall), SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	if n := effectstest.CountersOn(g, aspirant, game.CounterPlusOne); n != 1 {
		t.Errorf("the Aspirant has %d counters after a sacrifice, want 1", n)
	}
	if dmg := effectstest.B12Card(t, g, wall).DamageMarked; dmg != 1 {
		t.Errorf("the Wall has %d damage, want 1", dmg)
	}
	cantBlock := effectstest.AuraRestrictions(t, g, wall)&game.CantBlock != 0
	if !cantBlock {
		t.Error("the Wall can still block")
	}
}
