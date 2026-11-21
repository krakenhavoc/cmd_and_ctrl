package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestNamelessInversionIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.NamelessInversionOracle)
	if !ok {
		t.Fatal("Nameless Inversion is not registered")
	}
	if spec.Name != "Nameless Inversion" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// The +3/-3 kills a 3-toughness creature (CR 704.5f).
func TestNamelessInversionKillsAThreeToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.PushTribalCreature(g, g.Seats[1].ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastNamelessInversion(t, g, victim)
	if effects.OnBattlefield(g, victim) {
		t.Fatal("the 3/3 survived +3/-3")
	}
}

// A Goblin stops counting for a Goblin lord until end of turn, and
// counts again after it.
func TestNamelessInversionTurnsOffAGoblinLordUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: effectstest.GoblinChieftainOracle, Owner: mine, Controller: mine,
	})
	goblin := effectstest.PushTribalCreature(g, mine, "Goblin Brute", "Creature — Goblin", 2, 4)
	if p, tg := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 3 || tg != 5 {
		t.Fatalf("control: the Goblin is %d/%d, want 3/5 under the Chieftain", p, tg)
	}

	effectstest.CastNamelessInversion(t, g, goblin)
	if p, tg := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 5 || tg != 1 {
		t.Errorf("the Goblin is %d/%d, want 5/1: +3/-3 and no Chieftain bonus", p, tg)
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, goblin), "haste") {
		t.Error("the Chieftain still grants haste to a creature with no creature type")
	}
	if subs := effectstest.EffectiveSubtypes(t, g, goblin); len(subs) != 0 {
		t.Errorf("subtypes = %v, want none", subs)
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if p, tg := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 3 || tg != 5 {
		t.Errorf("after the turn the Goblin is %d/%d, want 3/5 again", p, tg)
	}
}

// A changeling under it is not an Elf, and keeps the keyword.
func TestNamelessInversionMakesAChangelingNoCreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	shifter := effectstest.PushTribalCreature(g, g.Seats[1].ID, "Woodland Changeling", "Creature — Shapeshifter", 2, 4, game.KeywordChangeling)
	effectstest.CastNamelessInversion(t, g, shifter)
	c := effectstest.LayeredCard(t, g, shifter)
	if c.HasSubtype("Elf") || game.HasAllCreatureTypes(&c) {
		t.Error("the changeling is still every creature type")
	}
	if !game.HasKeyword(&c, game.KeywordChangeling) {
		t.Error("the changeling lost its keyword; only its types go")
	}
}

// A later "gains all creature types" wins by timestamp (CR 613.7).
func TestNamelessInversionLosesToALaterAllCreatureTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := effectstest.PushTribalCreature(g, mine, "Goblin Brute", "Creature — Goblin", 2, 4)
	effectstest.CastNamelessInversion(t, g, goblin)
	if c := effectstest.LayeredCard(t, g, goblin); c.HasSubtype("Goblin") {
		t.Fatal("control: the Goblin kept its type")
	}
	effectstest.CastCatalogSpell(t, g, "Shields of Velis Vel", "Kindred Instant — Shapeshifter",
		effectstest.ShieldsOfVelisVelOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, goblin); !c.HasSubtype("Elf") || !c.HasSubtype("Goblin") {
		t.Error("the later all-creature-types effect did not win")
	}
}
