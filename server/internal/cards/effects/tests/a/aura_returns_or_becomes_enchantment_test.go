package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_returns_or_becomes_enchantment_test.go — #6235, Takklemaggot's
// return: to a creature the dead creature's controller chooses, or as a
// non-Aura enchantment that damages that player each upkeep.

const (
	maggotOracle = "test-6235-maggot"
	maggotGrant  = "test-6235-maggot/upkeep-damage"
)

// registerMaggot registers the test card once for the package: a grant key
// is catalog-wide and cannot be registered twice.
func registerMaggot(t *testing.T) {
	t.Helper()
	if _, ok := effects.Lookup(maggotOracle); ok {
		return
	}
	effects.Register(effects.Spec{
		OracleID:     maggotOracle,
		Name:         "Test Maggot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants:       []effects.AbilityGrant{effects.ThatPlayersUpkeepDealsDamageGrant(maggotGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Test Maggot — return it",
				effects.ReturnToNewHostOrAsEnchantment("Test Maggot", maggotGrant)),
		},
	})
}

// maggotTable: seat 0 controls the Aura on seat 1's creature.
func maggotTable(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	registerMaggot(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	maggot := effectstest.B12Push(g, me.ID, "Test Maggot", "Enchantment — Aura", maggotOracle, 0, 0)
	effectstest.AttachForTest(t, g, maggot, victim)
	return g, maggot, victim
}

// killHost destroys the enchanted creature and runs the state-based
// actions, which put the Aura into the graveyard (CR 704.5m) before the
// dies trigger goes on the stack, as at a real table.
func killHost(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	effectstest.DestroyNow(t, g, victim)
	g.RunStateChecksForTest()
}

func maggotPermanent(g *game.Game, id uuid.UUID) (game.Card, bool) {
	return effectstest.BattlefieldCard(g, id)
}

// With no creature it could enchant, it returns as an enchantment that
// is no Aura and deals 1 damage to the dead creature's controller at the
// beginning of that player's upkeep, and only theirs.
func TestAuraReturnsAsNonAuraEnchantmentWithNoHost(t *testing.T) {
	g, maggot, victim := maggotTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	killHost(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := maggotPermanent(g, maggot)
	if !ok {
		t.Fatal("the Aura did not return to the battlefield")
	}
	if c.IsAura() || !c.IsEnchantment() {
		t.Fatalf("it returned with types %v / %v, want a non-Aura enchantment", c.TypeLine, effectstest.EffectiveSubtypes(t, g, maggot))
	}
	if c.Controller != me.ID {
		t.Errorf("it returned under %s, want its owner's control", c.Controller)
	}
	if c.AttachedTo.ID != uuid.Nil {
		t.Errorf("it returned attached to %v", c.AttachedTo)
	}
	if got := effects.ChosenPlayerOf(g, maggot); got != opp.ID {
		t.Fatalf("\"that player\" is %s, want the dead creature's controller", got)
	}

	before := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, opp.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("that player's upkeep: life %d → %d, want 1 damage", before, opp.Life)
	}
	if _, ok := maggotPermanent(g, maggot); !ok {
		t.Error("the enchantment was put into the graveyard: it is not held to an enchant clause")
	}
	mine := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, me.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("another player's upkeep dealt damage: %d → %d", mine, me.Life)
	}
}

// With a creature to choose, the dead creature's controller picks and
// the Aura returns attached to it, under its controller's control.
func TestAuraReturnsAttachedToTheChosenCreature(t *testing.T) {
	g, maggot, victim := maggotTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	killHost(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the Aura's controller was asked; the dead creature's controller chooses")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, other)
	c, ok := maggotPermanent(g, maggot)
	if !ok {
		t.Fatal("the Aura did not return")
	}
	if !c.IsAura() || c.AttachedTo.ID != other || c.Controller != me.ID {
		t.Errorf("returned aura=%v attached=%v controller=%s, want an Aura on the chosen creature under seat 0", c.IsAura(), c.AttachedTo, c.Controller)
	}
}

// Declining the choice is "if they don't": it returns as the enchantment.
func TestAuraReturnsAsEnchantmentWhenTheChoiceIsDeclined(t *testing.T) {
	g, maggot, victim := maggotTable(t)
	opp := g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	killHost(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID)
	c, ok := maggotPermanent(g, maggot)
	if !ok || c.IsAura() {
		t.Fatalf("declined: on battlefield=%v aura=%v, want a non-Aura enchantment", ok, c.IsAura())
	}
	if got := effects.ChosenPlayerOf(g, maggot); got != opp.ID {
		t.Errorf("\"that player\" is %s, want the decliner", got)
	}
}
