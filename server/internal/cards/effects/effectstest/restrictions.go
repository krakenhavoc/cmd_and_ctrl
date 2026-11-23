package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restrictions_test.go — the S24 restriction cards, driven through a
// real cast or a real activation. The vocabulary itself is pinned in
// server/internal/game/restrictions_test.go and the "the bot is never
// offered a move the engine refuses" invariant in
// server/internal/legal/restrictions_test.go; these are the cards.

const (
	PacifismOracle      = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	ArrestOracle        = "81728b98-8cf9-4734-a318-69184bb4d15c"
	FaithsFettersOracle = "2b2d76f5-4c9b-49dc-b202-68095e2d9b29"
	RoguesPassageOracle = "f29dc596-2121-4421-8463-15f6c2e8b9b3"
)

// pushRestrictionBear is a 2/2 on the battlefield under `owner`,
// stamped with a battlefield timestamp so the layer engine sorts it.
func PushRestrictionBear(g *game.Game, owner *game.Player, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// restrictionsOf reads a battlefield card's post-layer restriction
// set, forcing a recompute first — the same read every engine gate
// makes.
func RestrictionsOf(t *testing.T, g *game.Game, id uuid.UUID) game.Restriction {
	t.Helper()
	var out game.Restriction
	found := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = game.RestrictionsOn(&g.Battlefield.Cards[i])
				found = true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}

func AssertRestrictions(t *testing.T, g *game.Game, id uuid.UUID, want game.Restriction) {
	t.Helper()
	if got := RestrictionsOf(t, g, id); got != want {
		t.Errorf("restrictions = %v, want %v", got.Names(), want.Names())
	}
}

// advanceToStepInTurn walks the cursor forward to `step` within the
// current turn.
func AdvanceToStepInTurn(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 30; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %v: %v", step, err)
		}
	}
	t.Fatalf("never reached %v (at %v)", step, g.Turn.Step)
}
