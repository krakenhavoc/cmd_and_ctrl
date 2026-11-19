package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contagion Engine — Artifact {6}:
//
//	"When this artifact enters, put a -1/-1 counter on each creature
//	 target player controls.
//	 {4}, {T}: Proliferate twice."
//
// The creatures are the ones the chosen player controls as the trigger
// resolves; the player is re-checked at resolution (CR 608.2b). The
// counters go on one creature after another through the keyed counter
// path, so a replacement on one is answered before the next. "Proliferate
// twice" is two proliferates in a row, the second asked once the first is
// answered (so "whenever you proliferate" triggers twice).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "482bd55a-ba73-4722-8643-94fa12818515",
		Name:         "Contagion Engine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters("Contagion Engine — put a -1/-1 counter on each creature target player controls",
					contagionEngineCounters),
					TargetPlayer("target player")),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Proliferate twice.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return proliferateTimesKeyed(g, item.Controller, item.SourceCardID, 2)
			},
		}},
	})
}

// contagionEngineCounters is the enters trigger's body.
func contagionEngineCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		var ids []uuid.UUID
		for _, c := range MatchingBattlefield(ctx, And(Creature(), ControlledBy(t.ID))) {
			ids = append(ids, c.InstanceID)
		}
		return avB24PutCountersOnEachThen(ctx, ids, game.CounterMinusOne, 1, game.KeyedTail{})
	}
	return nil
}
