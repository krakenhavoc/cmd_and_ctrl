package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative.go is the card vocabulary of the initiative (CR 726) and
// of dungeons (CR 309), ADR 0203 (#4096). The engine runs the
// designation itself — its three inherent triggered abilities and its
// hand-on when the holder leaves (game/initiative.go) — and the
// venture into Undercity it causes (game/dungeon.go). A card only has
// to say "you take the initiative" or ask about it:
//
//	TakeTheInitiative{}                  "you take the initiative"
//	WhenThisEntersYouTakeTheInitiative   the ETB most initiative cards print
//	YouHaveTheInitiative(g, you)         "if you have the initiative"
//	YouCompletedADungeon(g, you)         "if you've completed a dungeon"
//	WheneverYouCompleteADungeon          "Whenever you complete a dungeon"
//
// Where a condition is read follows the card, as for the monarch (ADR
// 0096 §5): "At the beginning of your end step, if you have the
// initiative" is an intervening "if" (CR 603.4), checked as the
// ability triggers and again as it resolves — IfYouHaveTheInitiative
// builds both halves.

// TakeTheInitiative is "<player> takes the initiative" (CR 726). Player
// defaults to the effect's controller. Taking it triggers the venture
// into Undercity, even for a player who already has it (CR 726.5).
type TakeTheInitiative struct {
	Player uuid.UUID
}

func (t TakeTheInitiative) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if p := ctx.Game.PlayerByIDForEffect(player); p == nil || p.Eliminated {
		return nil
	}
	return ctx.Game.TakeInitiativeForEffect(player, ctx.Source())
}

// WhenThisEntersYouTakeTheInitiative is "When this <permanent> enters,
// you take the initiative."
func WhenThisEntersYouTakeTheInitiative(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — you take the initiative", Do(TakeTheInitiative{}))
}

// YouHaveTheInitiative reports whether `you` have the initiative.
func YouHaveTheInitiative(g *game.Game, you uuid.UUID) bool {
	return game.HasInitiative(g, you)
}

// YouCompletedADungeon reports whether `you` have completed a dungeon
// this game (CR 309.7).
func YouCompletedADungeon(g *game.Game, you uuid.UUID) bool {
	return game.CompletedADungeon(g, you)
}

// youCompletedADungeonEvent is "whenever you complete a dungeon" as a
// trigger condition.
func youCompletedADungeonEvent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Kind == game.EventDungeonCompleted && ev.Actor == source.Controller
}

// WheneverYouCompleteADungeon — "Whenever you complete a dungeon, …"
// (Loot Dispute's Loud Ruckus).
func WheneverYouCompleteADungeon(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDungeonCompleted, youCompletedADungeonEvent, label, effect)
}

// IfYouHaveTheInitiative narrows a trigger to the intervening "if you
// have the initiative" (CR 603.4): it triggers only while its
// controller has the initiative, and does nothing if they have lost it
// by the time it resolves (InterveningIf's resolution check).
func IfYouHaveTheInitiative(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && YouHaveTheInitiative(g, source.Controller) && when(ev, source, lki, g)
	}
	// A Do keeps its steps visible to the footprint (If's footprint is
	// the condition's read followed by its steps); a hand-written
	// Effect goes through InterveningIf.
	if steps, ok := doStepsOf(t.Effect); ok {
		t.Effect = Do(If{Cond: CondYouHaveTheInitiative{}, Then: steps})
		return t
	}
	return InterveningIf(CondYouHaveTheInitiative{}, t)
}

// CondYouHaveTheInitiative is "if you have the initiative": the item's
// controller has it as the item resolves.
type CondYouHaveTheInitiative struct{}

func (CondYouHaveTheInitiative) Holds(ctx *Context) bool {
	return YouHaveTheInitiative(ctx.Game, ctx.Controller())
}

func (CondYouHaveTheInitiative) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondInitiative, 0), true
}
