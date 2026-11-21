package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// next_cast_copy_test.go — the catalog half of #663: the two cards
// whose printed ability is a delayed trigger with an EVENT condition.
// The mechanism (fires once, expires with the turn, takes the
// harvester's dispatch, survives a clone) is pinned in
// game/delayed_on_event_test.go.

const (
	DoublecastOracle        = "946675f5-8998-4f1b-934b-85ffe6e2f002"
	GalvanicIterationOracle = "c51ed6e3-813b-49c4-b1de-7f92a77f8f6e"
)

// castBoltAtAndKeepTheCopysTarget casts a Lightning Bolt at `victim`,
// then answers the CR 707.10c "you may choose new targets" prompt for
// every copy the cast produced by keeping the original target, and
// settles the stack.
//
// Keeping the target is the interesting answer here: the point of
// these cards is the second three damage, not the re-aim, and
// Reverberate's test already pins the re-aim.
func CastBoltAtAndKeepTheCopysTarget(t *testing.T, g *game.Game, caster, victim uuid.UUID, copies int) {
	t.Helper()
	CastCatalogSpell(t, g, "Lightning Bolt", "Instant", LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})

	answered := 0
	for i := 0; i < 48; i++ {
		// Two owed triggers go on the stack at one boundary, which is
		// an ordinary CR 603.3b ordering prompt — ordinary because by
		// then these ARE ordinary triggers.
		if p := TriggerOrderPromptFor(g, caster); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, caster, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := LatestPickTarget(g, caster); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster,
				game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget for copy %d: %v", answered+1, err)
			}
			answered++
			continue
		}
		if StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != copies {
		t.Fatalf("answered %d copy-target prompts, want %d", answered, copies)
	}
	PassPriorityAroundTable(t, g)
}

// triggerOrderPromptFor is the outstanding CR 603.3b ordering prompt
// addressed to `chooser`, or nil.
func TriggerOrderPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
