package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 3. Every test drives
// the registered catalog row through the engine.

// hobB3Cast puts the named spell in `who`'s hand, floats `mana`, and casts it
// under the strict mana gate, so a cast that costs more than the pool holds
// is refused. It returns the spell's id and the cast's error.
func HobB3Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := HandCardFull(who, name, typeLine, cost, HobOracle(t, name), nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

// hobB3Price is what the named spell costs to cast with `targets` chosen.
func HobB3Price(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost string, targets ...uuid.UUID) int {
	t.Helper()
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return SelfPriced(t, g, who, HobOracle(t, name), typeLine, cost, game.ZoneHand, refs).ManaValue()
}

// hobB3Activate floats `mana` and activates ability 0 of `id` under the
// strict mana gate.
func HobB3Activate(t *testing.T, g *game.Game, who, id uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{Strict: true})
}
