package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cobra Trap — Instant — Trap {4}{G}{G} (Commander 2015, S82 OLD #5546):
//
//	"If a noncreature permanent under your control was destroyed this turn by
//	 a spell or ability an opponent controlled, you may pay {G} rather than
//	 pay this spell's mana cost.
//	 Create four 1/1 green Snake creature tokens."
//
// The trap clause is an alternative cost (CR 118.9), Mindbreak Trap's shape,
// offered while this turn's log holds a battlefield exit stamped as a
// destruction (leave_cause.go: LeaveCauseDestroy, to the graveyard or by a
// replacement to exile) made by a resolving spell or ability another player
// controlled (Event.LeaveController, Karmic Justice's reading). The permanent
// is judged as it last existed on the battlefield (CR 603.10a): under your
// control and not a creature, so an animated land doesn't count. Sacrifices,
// state-based deaths and exile are not destruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3953f48-01c0-4647-9e4a-d70708d01e33",
		Name:         "Cobra Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If a noncreature permanent under your control was destroyed this turn by a spell or ability an opponent controlled, pay {G}",
			ManaCost:  "{G}",
			Condition: ANoncreaturePermanentOfYoursWasDestroyedByAnOpponentThisTurn,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 4}.Apply(ctx)
		},
	})
}

// ANoncreaturePermanentOfYoursWasDestroyedByAnOpponentThisTurn is Cobra
// Trap's condition, read off this turn's events.
func ANoncreaturePermanentOfYoursWasDestroyedByAnOpponentThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneBattlefield || ev.LeaveCause != game.LeaveCauseDestroy {
			continue
		}
		if ev.LeaveController == uuid.Nil || ev.LeaveController == controller {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || effects.LeftAsType(ev, c, "creature") {
			continue
		}
		if effects.LeftUnderControlOf(ev, c) == controller {
			return true
		}
	}
	return false
}
