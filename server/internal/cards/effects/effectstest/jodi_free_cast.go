package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_free_cast_test.go — S58 PR 5: Ugin's Binding, Sunbird's
// Invocation, Emergent Ultimatum and Nicol Bolas, God-Pharaoh.

// castFromExile casts `id` from exile with the free-cast grant,
// non-strict as the harness does everywhere.
func CastFromExile(t *testing.T, g *game.Game, caster, id uuid.UUID) error {
	t.Helper()
	return g.CastSpell(caster, id, game.CastSpellParams{FromZone: "exile"})
}

// pushGraveyardBinding puts Ugin's Binding in `p`'s graveyard.
func PushGraveyardBinding(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Ugin's Binding", TypeLine: "Instant",
		OracleID: UginsBindingOracle, ManaCost: "{2}{U}", Colors: nil, Owner: p.ID, Controller: p.ID})
	return id
}

// triggerOnStackFrom reports whether a triggered ability whose source
// is a card named `name` is waiting on the stack.
func TriggerOnStackFrom(g *game.Game, name string) bool {
	for _, item := range g.StackMeta {
		if item == nil || item.Kind != game.StackItemTriggered {
			continue
		}
		if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Name == name {
			return true
		}
	}
	return false
}

// --- Sunbird's Invocation ----------------------------------------------

// sunbirdLibrary stacks the top of `p`'s library, top card last in the
// argument order given (the first argument ends up on top).
func SunbirdLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

// --- Nicol Bolas, God-Pharaoh -------------------------------------------

func ActivateBolas(t *testing.T, g *game.Game, me *game.Player, bolas uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, bolas, idx, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
}
