package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_d_helpers_test.go — the shared scaffolding for the ADR 0103
// PR 3 Room tests (group D): a real Room card, the real cast and unlock
// actions, and a settle loop that also answers trigger targets.

// roomsDCard builds a Room card the way the importer settles one: both
// faces, materialised as the whole card (the state in a hand or library).
func roomsDCard(owner uuid.UUID, oracle, left, leftCost, right, rightCost, color string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: left, TypeLine: "Enchantment — Room", ManaCost: leftCost, Colors: []string{color}},
			{Name: right, TypeLine: "Enchantment — Room", ManaCost: rightCost, Colors: []string{color}},
		},
	}
	c.SettleImported()
	return c
}

// roomsDOnBattlefield puts a Room on the battlefield with `mask`
// unlocked, named for its unlocked doors (CR 709.5), a stand-in for one
// the player cast and unlocked earlier.
func roomsDOnBattlefield(g *game.Game, owner uuid.UUID, oracle, left, right string, mask game.DoorMask) uuid.UUID {
	name := ""
	switch {
	case mask.Full():
		name = left + " // " + right
	case mask.Has(game.DoorLeft):
		name = left
	case mask.Has(game.DoorRight):
		name = right
	}
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Owner: owner, Controller: owner,
		Name: name, TypeLine: "Enchantment — Room", Layout: game.LayoutSplit, Unlocked: mask,
		Faces: []game.Face{
			{Name: left, TypeLine: "Enchantment — Room"},
			{Name: right, TypeLine: "Enchantment — Room"},
		},
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// roomsDCast casts one half of a Room from hand and settles the stack.
func roomsDCast(t *testing.T, g *game.Game, me *game.Player, c game.Card, face int) {
	t.Helper()
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
	roomsDSettle(t, g, me.ID)
}

// roomsDUnlock takes the unlock special action on a door and settles.
func roomsDUnlock(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, door game.DoorSide) {
	t.Helper()
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionUnlock, game.SpecialActionParams{Door: door}); err != nil {
		t.Fatalf("unlock door %v: %v", door, err)
	}
	roomsDSettle(t, g, me.ID)
}

// roomsDSettle resolves everything on the stack, answering each trigger
// order prompt in the order offered and each trigger-target prompt with
// its first card.
func roomsDSettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil || c.Chooser != chooser {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoicePickTarget:
				if len(c.PickTargetCards) == 0 {
					t.Fatalf("a target prompt with no candidates: %+v", c)
				}
				if err := g.ResolvePickTarget(c.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: c.PickTargetCards[0]}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
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
		if len(g.PendingChoices) > 0 && !hasAnswerableChoice(g, chooser) {
			return
		}
	}
	t.Fatal("the stack never settled")
}

// hasAnswerableChoice reports whether the settle loop can answer a
// pending choice itself; any other kind is the test's to answer.
func hasAnswerableChoice(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser &&
			(c.Kind == game.PendingChoiceTriggerOrder || c.Kind == game.PendingChoicePickTarget) {
			return true
		}
	}
	return false
}

// roomsDCountNamed counts battlefield permanents with a name.
func roomsDCountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}
