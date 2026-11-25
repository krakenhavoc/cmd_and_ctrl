package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oracleMysticReflection = "491dc110-b459-412e-8a0b-45e5744ee8d1"

// mysticReflectionCastCreature casts a vanilla creature from the active
// player's hand and resolves it.
func mysticReflectionCastCreature(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		PrintedPTKnown: true, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// TestMysticReflectionCopiesTheNextEntryOnly — #6160's acceptance: a
// creature spell resolving after Mystic Reflection enters as a copy of the
// chosen creature (CR 707.2); two tokens created later are a separate
// entry and are themselves.
func TestMysticReflectionCopiesTheNextEntryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	model := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Mystic Reflection", "Instant", oracleMysticReflection,
		[]game.TargetRef{{Kind: game.TargetCard, ID: model}})
	effectstest.PassPriorityAroundTable(t, g)

	bear := mysticReflectionCastCreature(t, g, "Grizzly Bears")
	if c := effectstest.DcCard(t, g, bear); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Fatalf("the next creature entered as %q (power %d), want a Hill Giant", c.Name, c.Effective().Power)
	}

	var made []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		made, err = g.CreateTokensForEffect(me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier",
			Power: 1, Toughness: 1, PrintedPTKnown: true}, 2, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	for _, id := range made {
		if c := effectstest.DcCard(t, g, id); c.Name != "Soldier" {
			t.Errorf("a token created after the spent entry became %q", c.Name)
		}
	}
}

// TestMysticReflectionTwoTokensTogetherBothCopy — "one or more … they
// enter as copies": every permanent of the one simultaneous entry.
func TestMysticReflectionTwoTokensTogetherBothCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	model := effectstest.SeedCopyableCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Mystic Reflection", "Instant", oracleMysticReflection,
		[]game.TargetRef{{Kind: game.TargetCard, ID: model}})
	effectstest.PassPriorityAroundTable(t, g)

	var made []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		made, err = g.CreateTokensForEffect(me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier",
			Power: 1, Toughness: 1, PrintedPTKnown: true}, 2, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	if len(made) != 2 {
		t.Fatalf("made %d tokens, want 2", len(made))
	}
	for _, id := range made {
		c := effectstest.DcCard(t, g, id)
		if c.Name != "Hill Giant" || !c.IsToken() || c.Controller != me.ID {
			t.Errorf("token entered as %q (token %v, controller %v), want my token copy of Hill Giant", c.Name, c.IsToken(), c.Controller)
		}
	}
	bear := mysticReflectionCastCreature(t, g, "Grizzly Bears")
	if c := effectstest.DcCard(t, g, bear); c.Name != "Grizzly Bears" {
		t.Errorf("a creature cast after the tokens became %q", c.Name)
	}
}

// TestMysticReflectionCannotTargetALegendaryCreature — "target
// nonlegendary creature".
func TestMysticReflectionCannotTargetALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := effectstest.SeedCopyableCreature(g, me.ID, "Isamaru", "Legendary Creature — Dog", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Mystic Reflection", "Instant", oracleMysticReflection,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Mystic Reflection on a legendary creature: err = %v, want ErrIllegalTarget", err)
	}
}

// TestMysticReflectionCanBeForetold — the Foretell {U} special action is
// declared.
func TestMysticReflectionCanBeForetold(t *testing.T) {
	spec, ok := effects.Lookup(oracleMysticReflection)
	if !ok || len(spec.SpecialActions) != 1 {
		t.Fatalf("Mystic Reflection declares %d special actions, want foretell", len(spec.SpecialActions))
	}
}
