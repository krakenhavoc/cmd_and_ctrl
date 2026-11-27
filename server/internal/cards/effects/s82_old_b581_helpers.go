package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b581_helpers.go — S82 OLD batch 581 (Ravnica: City of Guilds, #5600).

// CreaturesBlockedWithThisTurn is "creatures that blocked or were blocked by
// <id> this turn", read from the turn's block declarations (one EventBlock per
// pair: CardID the blocker, Target the attacker), in the order they were
// declared. Objects, not names (CR 400.7): a battlefield entry by `id` starts
// a new object that nothing has blocked, and any zone move by the other
// creature ends that object. Caller holds g.mu.
func CreaturesBlockedWithThisTurn(g *game.Game, id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.EventsThisTurn() {
		switch ev.Kind {
		case game.EventZoneMove:
			if ev.CardID == id && ev.NewZone == game.ZoneBattlefield {
				out = out[:0]
				continue
			}
			if i := slices.Index(out, ev.CardID); i >= 0 {
				out = slices.Delete(out, i, i+1)
			}
		case game.EventBlock:
			other := uuid.Nil
			switch id {
			case ev.CardID:
				other = ev.Target
			case ev.Target:
				other = ev.CardID
			}
			if other != uuid.Nil && other != id && !slices.Contains(out, other) {
				out = append(out, other)
			}
		}
	}
	return out
}

// GazeOfTheGorgonBody is the delayed trigger Gaze of the Gorgon schedules: "at
// this turn's next end of combat, destroy all creatures that blocked or were
// blocked by it this turn". The trigger's target is the creature, its Amount
// the end of combat steps that had begun when it was scheduled, so it acts only
// at the first one after (a later extra combat is not "next"; Triton Tactics'
// count). The set is read as the trigger resolves, not as the spell did.
var GazeOfTheGorgonBody = game.SimpleDelayedBody("s82-old-b581/gaze-of-the-gorgon-destroy", gazeOfTheGorgonDestroy)

func gazeOfTheGorgonDestroy(g *game.Game, item *game.StackItem) error {
	if EndCombatsBegunThisTurn(g) != item.Params.Amount+1 {
		return nil
	}
	set := map[uuid.UUID]bool{}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		for _, id := range CreaturesBlockedWithThisTurn(g, t.ID) {
			set[id] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return set[c.InstanceID] && c.IsCreature() },
	}.Apply(NewContext(g, item))
}
