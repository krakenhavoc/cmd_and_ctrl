package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// strict is the payment every test here makes: the engine spends the
// pool, so the tokens — and their riders — are recorded.
var Strict = game.CastSpellParams{Strict: true}

// castFromHandErr puts a card in `p`'s hand and casts it, returning the
// error rather than failing — the refusal tests need it.
func CastFromHandErr(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
	return id, g.CastSpell(p.ID, id, params)
}

// counterByEffect runs the counter verb every catalog counterspell
// calls, and reports whether the spell is still on the stack after it.
func CounterByEffect(t *testing.T, g *game.Game, spell uuid.UUID) (stillOnStack bool) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
		stillOnStack = g.Stack.Contains(spell)
	})
	return stillOnStack
}

// settleKeepingCopyTargets passes priority until the stack is empty,
// answering every CR 707.10c "choose new targets" prompt by keeping the
// copy aimed at `victim` and every trigger-ordering prompt as offered.
func SettleKeepingCopyTargets(t *testing.T, g *game.Game, caster, victim uuid.UUID, copies int) {
	t.Helper()
	answered := 0
	for i := 0; i < 48; i++ {
		if p := TriggerOrderPromptFor(g, caster); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, caster, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := LatestPickTarget(g, caster); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
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
}
