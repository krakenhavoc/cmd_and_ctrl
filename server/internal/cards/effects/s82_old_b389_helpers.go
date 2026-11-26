package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// endCombatsBegunThisTurn counts the end of combat steps that have begun
// this turn (an extra combat makes a second one).
func EndCombatsBegunThisTurn(g *game.Game) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventStepBegan && ev.Step == game.StepEndCombat {
			n++
		}
	}
	return n
}

// TritonTacticsBody is the delayed trigger Triton Tactics schedules: "at
// this turn's next end of combat, tap each creature that was blocked by one
// of those creatures this turn and it doesn't untap during its controller's
// next untap step". The trigger's targets are the pumped creatures, its
// Amount the end of combat steps that had begun when it was scheduled, so it
// acts only at the first one after (a later extra combat is not "next"). A
// block pair is one EventBlock: CardID the blocker, Target the attacker.
var TritonTacticsBody = game.SimpleDelayedBody("s82-old-b389/triton-tactics-tap-blocked", tritonTacticsTapBlocked)

func tritonTacticsTapBlocked(g *game.Game, item *game.StackItem) error {
	if EndCombatsBegunThisTurn(g) != item.Params.Amount+1 {
		return nil
	}
	pumped := map[uuid.UUID]bool{}
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard {
			pumped[t.ID] = true
		}
	}
	ctx := NewContext(g, item)
	seen := map[uuid.UUID]bool{}
	var tapped []uuid.UUID
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventBlock || !pumped[ev.CardID] || seen[ev.Target] || !OnBattlefield(g, ev.Target) {
			continue
		}
		seen[ev.Target] = true
		if err := (TapTarget{Target: ev.Target}).Apply(ctx); err != nil {
			return err
		}
		tapped = append(tapped, ev.Target)
	}
	if len(tapped) == 0 {
		return nil
	}
	return DoesntUntapNextUntapStep{Targets: tapped, Label: "Triton Tactics — doesn't untap during its controller's next untap step"}.Apply(ctx)
}
