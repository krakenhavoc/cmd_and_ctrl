package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nameless_inversion_test.go — Nameless Inversion (#4691, Lorwyn
// Eclipsed #4458).

const namelessInversionOracle = "d6911456-42cb-4b23-a560-dc98784530f5"

func castNamelessInversion(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Nameless Inversion", "Kindred Instant — Shapeshifter",
		namelessInversionOracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func TestNamelessInversionIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(namelessInversionOracle)
	if !ok {
		t.Fatal("Nameless Inversion is not registered")
	}
	if spec.Name != "Nameless Inversion" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// The +3/-3 kills a 3-toughness creature (CR 704.5f).
func TestNamelessInversionKillsAThreeToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := pushTribalCreature(g, g.Seats[1].ID, "Hill Giant", "Creature — Giant", 3, 3)
	castNamelessInversion(t, g, victim)
	if onBattlefield(g, victim) {
		t.Fatal("the 3/3 survived +3/-3")
	}
}

// A Goblin stops counting for a Goblin lord until end of turn, and
// counts again after it.
func TestNamelessInversionTurnsOffAGoblinLordUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: goblinChieftainOracle, Owner: mine, Controller: mine,
	})
	goblin := pushTribalCreature(g, mine, "Goblin Brute", "Creature — Goblin", 2, 4)
	if p, tg := effectivePower(t, g, goblin), effectiveToughness(t, g, goblin); p != 3 || tg != 5 {
		t.Fatalf("control: the Goblin is %d/%d, want 3/5 under the Chieftain", p, tg)
	}

	castNamelessInversion(t, g, goblin)
	if p, tg := effectivePower(t, g, goblin), effectiveToughness(t, g, goblin); p != 5 || tg != 1 {
		t.Errorf("the Goblin is %d/%d, want 5/1: +3/-3 and no Chieftain bonus", p, tg)
	}
	if hasAbility(effectiveAbilities(t, g, goblin), "haste") {
		t.Error("the Chieftain still grants haste to a creature with no creature type")
	}
	if subs := effectiveSubtypes(t, g, goblin); len(subs) != 0 {
		t.Errorf("subtypes = %v, want none", subs)
	}

	advancePastCleanupForTest(t, g)
	if p, tg := effectivePower(t, g, goblin), effectiveToughness(t, g, goblin); p != 3 || tg != 5 {
		t.Errorf("after the turn the Goblin is %d/%d, want 3/5 again", p, tg)
	}
}

// A changeling under it is not an Elf, and keeps the keyword.
func TestNamelessInversionMakesAChangelingNoCreatureType(t *testing.T) {
	g := newCatalogGame(t)
	shifter := pushTribalCreature(g, g.Seats[1].ID, "Woodland Changeling", "Creature — Shapeshifter", 2, 4, game.KeywordChangeling)
	castNamelessInversion(t, g, shifter)
	c := layeredCard(t, g, shifter)
	if c.HasSubtype("Elf") || game.HasAllCreatureTypes(&c) {
		t.Error("the changeling is still every creature type")
	}
	if !game.HasKeyword(&c, game.KeywordChangeling) {
		t.Error("the changeling lost its keyword; only its types go")
	}
}

// A later "gains all creature types" wins by timestamp (CR 613.7).
func TestNamelessInversionLosesToALaterAllCreatureTypes(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := pushTribalCreature(g, mine, "Goblin Brute", "Creature — Goblin", 2, 4)
	castNamelessInversion(t, g, goblin)
	if c := layeredCard(t, g, goblin); c.HasSubtype("Goblin") {
		t.Fatal("control: the Goblin kept its type")
	}
	castCatalogSpell(t, g, "Shields of Velis Vel", "Kindred Instant — Shapeshifter",
		shieldsOfVelisVelOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: mine}})
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, goblin); !c.HasSubtype("Elf") || !c.HasSubtype("Goblin") {
		t.Error("the later all-creature-types effect did not win")
	}
}
