package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 4. Every test drives
// the registered catalog row through the engine.

// hobB4Activate floats `mana`, then activates ability `idx` of `id` with
// strict payment (the cost must really be paid) at `targets`.
func HobB4Activate(t *testing.T, g *game.Game, who, id uuid.UUID, idx int, mana string, targets []game.TargetRef) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, idx, game.ActivateAbilityParams{Targets: targets, Strict: true})
}

// hobB4Creature puts a creature with the given type line onto the
// battlefield, untapped and ready to attack.
func HobB4Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// hobB4NoPickPrompt reports that no pick_target prompt is open.
func HobB4NoPickPrompt(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return false
		}
	}
	return true
}
