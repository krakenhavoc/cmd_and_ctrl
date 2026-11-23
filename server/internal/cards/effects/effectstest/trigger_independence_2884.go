package effectstest

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_2884_test.go — #2884: "Ask only when the order
// matters" skips the CR 603.3b prompt for a batch whose items are
// pairwise independent, read off each row's declared footprint
// (footprint.go, game/trigger_independence.go). ADR 0018's #2884
// amendment.

const (
	EssenceWardenOracle       = "6ca2a89e-7032-4864-b4e9-66f3178f90ab"
	FirebrandArcherOracle     = "2e9289d6-dbc6-456d-88cf-d1f534e731d6"
	IndependenceTokenProbe    = "test-2884-token-probe"
	IndependenceCountingProbe = "test-2884-counting-probe"
)

func init() {
	// "Whenever you cast a noncreature spell, create a 1/1 colorless
	// Construct token": a declared token step, with a token that has no
	// text.
	effects.Register(effects.Spec{
		OracleID: IndependenceTokenProbe,
		Name:     "Independence Token Probe",
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Independence Token Probe — create a 1/1 Construct",
				effects.Do(effects.CreateToken{Template: game.Card{
					Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1,
				}, N: 1})),
		},
	})
	// "Whenever you cast a noncreature spell, put a +1/+1 counter on this
	// creature for each creature you control": a count of the board, in
	// a hand-written effect the registry cannot read.
	effects.Register(effects.Spec{
		OracleID: IndependenceCountingProbe,
		Name:     "Independence Counting Probe",
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Independence Counting Probe — a counter for each creature",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.IsCreature() {
							n++
						}
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// settleUntilOrderPrompt passes priority until a trigger_order prompt
// opens or the stack is empty, and returns the prompt (or nil).
func SettleUntilOrderPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16 && TriggerOrderPrompt(g) == nil && !StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	return TriggerOrderPrompt(g)
}
