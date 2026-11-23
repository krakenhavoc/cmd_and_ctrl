package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_effects_test.go — the four Clone-class cards, end to end
// through the real cast → resolve → replacement → prompt → entry
// pipeline. What is asserted is deliberately the OBSERVABLE result
// (name, P/T, types, counters, catalog behaviour), because that is
// the only thing that tells "the copy landed" apart from "the copy
// landed on a card nobody can see".

const (
	OracleClone      = "42226b87-0746-4ebf-9fd0-108d508462af"
	OracleMetamorph  = "340bbe8b-e987-4c3e-ab4e-9dee63e57d4f"
	OracleSakashima  = "a7243d25-22a2-4df5-adaf-1f40f5330ec1"
	OracleSparkDoubl = "8dcb35e5-ae44-455f-86e3-4a77d496ff34"
)

// copyPrompt returns the queued copy-target choice, or nil.
func CopyPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCopyTarget {
			return c
		}
	}
	return nil
}

// resolveWithCopyChoice passes priority until the copy picker opens,
// answers it with `pick` (uuid.Nil declines), then settles the
// stack. Fails the test if no picker ever appears — every card in
// this file is supposed to ask.
func ResolveWithCopyChoice(t *testing.T, g *game.Game, pick uuid.UUID) {
	t.Helper()
	asked := false
	for i := 0; i < 32; i++ {
		if c := CopyPrompt(g); c != nil {
			asked = true
			if err := g.ResolveCopyTarget(c.ID, c.Chooser, pick); err != nil {
				t.Fatalf("ResolveCopyTarget: %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	if !asked {
		t.Fatal("no copy-target prompt was ever queued")
	}
}

// battlefieldCard returns the named battlefield card by value.
func CopyBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return game.Card{}
}

// seedCreature puts a vanilla creature on the battlefield under the
// active seat's control, through the zone-move event so the layer
// listener stamps it.
func SeedCopyableCreature(g *game.Game, controller uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   "oracle-" + name,
		ManaCost:   "{1}{G}",
		Power:      p,
		Toughness:  tough,
		Owner:      controller,
		Controller: controller,
	})
}

// hasCopySupertype reports whether the card's EFFECTIVE type line
// carries the supertype.
func HasCopySupertype(c game.Card, s string) bool {
	for _, got := range c.Effective().Supertypes {
		if got == s {
			return true
		}
	}
	return false
}
