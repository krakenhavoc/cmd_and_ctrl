package testsru

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Test helpers the GAP roll-up 35 tests (#5948) used from other test
// packages while the catalog was one package. A test package cannot reach
// another's unexported helpers, so this package keeps its own copies.

const r35PacifismOracle = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"

// r35AnswerCopyTarget is answerCopyTarget from so/spinerock_tyrant_test.go, copied for the GAP roll-up 35 tests.
func r35AnswerCopyTarget(t *testing.T, g *game.Game, chooser uuid.UUID, to game.TargetRef) {
	t.Helper()
	p := effectstest.CopyPromptFor(g, chooser)
	if p == nil {
		t.Fatal("no new-targets prompt for the copy is open")
	}
	if err := g.ResolvePickTarget(p.ID, chooser, to); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// r35B23Cast is b23Cast from av/batch23_test.go, copied for the GAP roll-up 35 tests.
func r35B23Cast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, colors []string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(active, name, typeLine, manaCost, oracle, colors)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// r35B23Draw is b23Draw from av/batch23_test.go, copied for the GAP roll-up 35 tests.
func r35B23Draw(g *game.Game, p uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p, n) })
}

// r35BbdAnswerOption is bbdAnswerOption from av/bbd_5529_b1_cards_test.go, copied for the GAP roll-up 35 tests.
func r35BbdAnswerOption(t *testing.T, g *game.Game, chooser uuid.UUID, label string) {
	t.Helper()
	c := effectstest.LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s", chooser)
	}
	for i, opt := range c.PickOptions {
		if opt.Label == label {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick(%q): %v", label, err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered: %v", label, effectstest.OptionWords(c))
}

// r35BlBorrowedActivated is blBorrowedActivated from av/borrowed_lender_cards_test.go, copied for the GAP roll-up 35 tests.
func r35BlBorrowedActivated(t *testing.T, g *game.Game, id uuid.UUID) (idx []int, origins game.AbilityOrigins) {
	t.Helper()
	_, all := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, id))
	for i, o := range all {
		if o.Borrowed {
			idx = append(idx, i)
			origins = append(origins, o)
		}
	}
	return idx, origins
}

// r35EnlistDeclare is enlistDeclare from ds/enlist_cards_test.go, copied for the GAP roll-up 35 tests.
func r35EnlistDeclare(t *testing.T, g *game.Game, attacker, enlisted uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: attacker, Target: g.Seats[1].ID, Enlist: enlisted}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare with enlist: %v", err)
	}
}

// r35EnlistSettle is enlistSettle from ds/enlist_cards_test.go, copied for the GAP roll-up 35 tests.
func r35EnlistSettle(t *testing.T, g *game.Game, first string) {
	t.Helper()
	g.RunStateChecksForTest()
	for i := 0; i < 64; i++ {
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
			order := append([]uuid.UUID(nil), c.TriggerOrderIDs...)
			for j, id := range order {
				if first != "" && strings.Contains(r35EnlistItemLabel(g, id), first) {
					order[0], order[j] = order[j], order[0]
					break
				}
			}
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if c := effectstest.ScryChoiceFor(g, g.Seats[0].ID); c != nil {
			if err := g.ResolveScry(c.ID, c.Chooser, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the declaration's triggers never settled")
}

// r35EnlistItemLabel is enlistItemLabel from ds/enlist_cards_test.go, copied for the GAP roll-up 35 tests.
func r35EnlistItemLabel(g *game.Game, id uuid.UUID) string {
	for _, it := range g.PendingTriggers {
		if it != nil && it.ID == id {
			return it.Label
		}
	}
	if it := g.StackMeta[id]; it != nil {
		return it.Label
	}
	return ""
}

// r35GptB2CardRef is gptB2CardRef from g/gpt_5599_b2_cards_test.go, copied for the GAP roll-up 35 tests.
func r35GptB2CardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// r35GtcAttackWith is gtcAttackWith from g/gtc_5565_b1_cards_test.go, copied for the GAP roll-up 35 tests.
func r35GtcAttackWith(t *testing.T, g *game.Game, id uuid.UUID, others int) {
	t.Helper()
	me, opp := effectstest.BnsActive(g)
	attackers := []uuid.UUID{id}
	for i := 0; i < others; i++ {
		attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1))
	}
	effectstest.DeclareAttack(t, g, opp.ID, attackers...)
}

// r35KitsunePacifism is kitsunePacifism from i/kitsune_mystic_test.go, copied for the GAP roll-up 35 tests.
func r35KitsunePacifism(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.ApaPush(g, p.ID, p.ID, game.Card{
		Name: "Pacifism", OracleID: r35PacifismOracle,
		TypeLine: "Enchantment — Aura", Colors: []string{"W"},
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatal(err)
		}
	})
	return id
}
