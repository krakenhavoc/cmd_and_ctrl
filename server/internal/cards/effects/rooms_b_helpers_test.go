package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_b_helpers_test.go — the harness the ADR 0103 group B Room tests
// share: a real Room card (its real oracle ID, so the catalog entry
// drives it), a real cast of either half, a real unlock, and a settle
// loop that answers the prompts a door trigger raises.

func roomsBCard(owner uuid.UUID, oracle, leftName, leftCost, rightName, rightCost string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: leftName, TypeLine: "Enchantment — Room", ManaCost: leftCost},
			{Name: rightName, TypeLine: "Enchantment — Room", ManaCost: rightCost},
		},
	}
	c.SettleImported()
	return c
}

// roomsBCast puts the Room in hand and casts the given half (0 left,
// 1 right) in the main phase. Nothing resolves yet.
func roomsBCast(t *testing.T, g *game.Game, me *game.Player, c game.Card, face int) {
	t.Helper()
	advanceToMain(t, g)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
}

// roomsBUnlock takes the CR 709.5e special action for one door.
func roomsBUnlock(t *testing.T, g *game.Game, me *game.Player, c game.Card, door game.DoorSide) {
	t.Helper()
	if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionUnlock, game.SpecialActionParams{Door: door}); err != nil {
		t.Fatalf("unlock door %v: %v", door, err)
	}
}

// roomsBSettle resolves everything, ordering simultaneous triggers as
// offered and picking `pick` (a card ID, or none) on any target or
// choose-cards prompt `me` gets.
func roomsBSettle(t *testing.T, g *game.Game, me *game.Player, pick ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 96; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil || c.Chooser != me.ID {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, me.ID, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoicePickTarget:
				if len(pick) == 0 {
					t.Fatalf("unexpected pick_target prompt")
				}
				if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: pick[0]}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				answered = true
			case game.PendingChoiceChooseCards:
				if len(pick) == 0 {
					t.Fatalf("unexpected choose_cards prompt")
				}
				if err := g.ResolveChooseCards(c.ID, me.ID, pick[:1]); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// roomsBCountNamed counts battlefield permanents with the given name.
func roomsBCountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == name {
			n++
		}
	}
	return n
}
