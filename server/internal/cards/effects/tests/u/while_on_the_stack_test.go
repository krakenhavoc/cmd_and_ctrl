package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// while_on_the_stack_test.go — #6222: statics of a spell that function
// while it is on the stack. CR 113.6b: "An ability that states which
// zones it functions in functions only from those zones." Kaervek's
// Torch and Torrent of Lava are their batch's; the shapes are registered
// here.

const (
	torchTestOracle    = "test-6222-torch"
	torrentTestOracle  = "test-6222-torrent"
	counterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"
	shockTestOracle    = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

func registerTorch(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: torchTestOracle,
		Name:     "Torch Test",
		Targets:  effects.TargetAny(),
		CostModifiers: []game.CostModifier{effects.SpellsThatTargetThisCostMore(2,
			"As long as Torch Test is on the stack, spells that target it cost {2} more to cast.")},
		OnResolve: func(*game.StackItem, *effects.Context) error { return nil },
	})
}

// torchOnTheStack casts the Torch from seat 0 at seat 1 and passes
// priority to seat 1.
func torchOnTheStack(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	registerTorch(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	torch := effectstest.CastCatalogSpell(t, g, "Torch Test", "Instant", torchTestOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	return g, me, opp, torch
}

func castStrict(g *game.Game, p *game.Player, id uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

// The issue's acceptance: a counterspell targeting the Torch costs {2}
// more. {U}{U} alone no longer casts it; {U}{U} and two more does.
func TestACounterspellAtKaerveksTorchCostsTwoMore(t *testing.T) {
	g, _, opp, torch := torchOnTheStack(t)
	cs := effectstest.HandCardFull(opp, "Counterspell", "Instant", "{U}{U}", counterspellOracle, []string{"U"})
	at := game.TargetRef{Kind: game.TargetCard, ID: torch}

	effectstest.FloatForTest(g, opp, "UU")
	if err := castStrict(g, opp, cs, at); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("Counterspell at the Torch for {U}{U} = %v, want ErrInsufficientMana", err)
	}
	effectstest.FloatForTest(g, opp, "CC")
	if err := castStrict(g, opp, cs, at); err != nil {
		t.Fatalf("Counterspell at the Torch for {2}{U}{U}: %v", err)
	}
	if !g.Stack.Contains(cs) {
		t.Fatal("the Counterspell should be on the stack")
	}
	if n := opp.ManaPool.Total(); n != 0 {
		t.Errorf("the pool should be spent, %d mana left", n)
	}
}

// A spell that doesn't target the Torch pays its own cost.
func TestASpellThatDoesNotTargetTheTorchIsNotTaxed(t *testing.T) {
	g, me, opp, _ := torchOnTheStack(t)
	shock := effectstest.HandCardFull(opp, "Shock", "Instant", "{R}", shockTestOracle, []string{"R"})
	effectstest.FloatForTest(g, opp, "R")
	if err := castStrict(g, opp, shock, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("Shock at a player while the Torch is on the stack: %v", err)
	}
}

// A static that declares the stack applies while its spell is there and
// stops when it leaves; "you" is the spell's controller.
func TestAStaticFunctionsWhileItsSpellIsOnTheStack(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: torrentTestOracle,
		Name:     "Torrent Test",
		Static: []game.StaticAbility{{
			Label: "As long as Torrent Test is on the stack, creatures its controller controls have flying.",
			Layer: game.Layer6Ability,
			Zones: []game.ZoneKind{game.ZoneStack},
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "flying")
			},
		}},
		OnResolve: func(*game.StackItem, *effects.Context) error { return nil },
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	flies := func(id uuid.UUID) bool {
		var out bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					out = game.HasKeyword(&g.Battlefield.Cards[i], "flying")
				}
			}
		})
		return out
	}
	if flies(mine) {
		t.Fatal("no creature should fly before the spell is cast")
	}
	spell := effectstest.CastCatalogSpell(t, g, "Torrent Test", "Sorcery", torrentTestOracle, nil)
	if !flies(mine) {
		t.Error("seat 0's creature should fly while seat 0's spell is on the stack")
	}
	if flies(theirs) {
		t.Error("seat 1's creature is not the spell's controller's")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) {
		t.Fatal("the spell should have resolved")
	}
	if flies(mine) {
		t.Error("the static must stop when its spell leaves the stack")
	}
}
