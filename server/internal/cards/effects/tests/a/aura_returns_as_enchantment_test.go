package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_returns_as_enchantment_test.go — #6235: Takklemaggot's dies
// trigger (effects.AuraReturnsOnHostDeath) and its gained upkeep trigger
// (effects.ChosenPlayerUpkeepDamageGrant), on a test card shaped like
// it. Takklemaggot itself ships with its OLD batch.

const (
	returnsAsEnchantmentOracle = "test-6235-takklemaggot"
	returnsAsEnchantmentGrant  = "test-6235-takklemaggot/upkeep-damage"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: returnsAsEnchantmentOracle,
		Name:     "Test Takklemaggot",
		Targets:  effects.EnchantCreature(),
		Grants:   []effects.AbilityGrant{effects.ChosenPlayerUpkeepDamageGrant(returnsAsEnchantmentGrant, 1, "Test Takklemaggot")},
		Triggered: []game.TriggeredAbility{
			effects.AuraReturnsOnHostDeath("Test Takklemaggot", returnsAsEnchantmentGrant),
		},
	})
}

func returnsSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16 && !effectstest.StackFullyEmpty(g); i++ {
		if effectstest.ChooseCardsChoiceFor(g, g.Seats[0].ID) != nil || effectstest.ChooseCardsChoiceFor(g, g.Seats[1].ID) != nil {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func returnsCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok = g.LookupCardForEffect(id)
	})
	if !ok {
		t.Fatalf("card %v is nowhere", id)
	}
	return c
}

// The acceptance: with no creature it could enchant, it returns under its
// controller's control as an enchantment that is no Aura, has no enchant,
// enchants nothing, and deals 1 damage to the dead creature's controller
// at the beginning of that player's upkeep (and not at its controller's).
func TestAuraReturnsAsAnEnchantmentWithNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Test Takklemaggot", returnsAsEnchantmentOracle, bear)
	effectstest.KillCreature(t, g, opp.ID, bear)
	returnsSettle(t, g)

	if !g.Battlefield.Contains(aura) {
		t.Fatal("it did not return to the battlefield")
	}
	c := returnsCard(t, g, aura)
	if c.Controller != me.ID {
		t.Errorf("controller %v, want the trigger's controller", c.Controller)
	}
	if c.IsAura() || !c.IsEnchantment() {
		t.Errorf("types %v %v: want a non-Aura enchantment", c.Effective().Types, c.Effective().Subtypes)
	}
	if c.AttachedTo.ID != uuid.Nil {
		t.Errorf("attached to %v, want nothing", c.AttachedTo)
	}
	for _, a := range c.Effective().Abilities {
		if len(a) >= 7 && a[:7] == "enchant" {
			t.Errorf("it kept %q", a)
		}
	}
	if c.ChosenPlayer != opp.ID {
		t.Errorf("that player = %v, want the dead creature's controller", c.ChosenPlayer)
	}

	life := opp.Life
	myLife := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("that player's life %d at their upkeep, want %d", opp.Life, life-1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("its controller took damage at their own upkeep: %d, want %d", me.Life, myLife)
	}
}

// With a creature it could enchant, the dead creature's controller chooses
// one and it returns attached to that creature, still an Aura.
func TestAuraReturnsAttachedToTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	other := effectstest.PushTypedCard(g, me.ID, "Other Bear", "Creature — Bear", "{1}{G}")
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Test Takklemaggot", returnsAsEnchantmentOracle, bear)
	effectstest.KillCreature(t, g, opp.ID, bear)
	returnsSettle(t, g)

	if p := effectstest.ChooseCardsChoiceFor(g, me.ID); p != nil {
		t.Fatal("the Aura's controller was asked; the dead creature's controller chooses")
	}
	p := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != other || p.ChooseMin != 1 {
		t.Fatalf("prompt %+v, want the opponent choosing the one creature", p)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, other)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("it did not return")
	}
	c := returnsCard(t, g, aura)
	if c.AttachedTo.ID != other || !c.IsAura() || c.Controller != me.ID {
		t.Errorf("attached to %v, Aura %v, controller %v: want an Aura of mine on the chosen creature", c.AttachedTo.ID, c.IsAura(), c.Controller)
	}
}
