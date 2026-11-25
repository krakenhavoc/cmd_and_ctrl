package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// when_player_loses_with_counter.go — #6123: a delayed trigger on a marked
// player losing the game, and the token copies of what they controlled.
//
//	Miss Highwater: "… they draw seven cards and get a contract counter.
//	For as long as they have a contract counter, when they lose the game,
//	for each artifact and creature they controlled, create a token that's
//	a copy of it."
//
// The "when they lose the game" is a delayed triggered ability the
// resolving ability creates (CR 603.7a), controlled by that ability's
// controller (CR 603.7d) and owing nothing to its source: it keeps working
// after Miss Highwater has left the battlefield. It hangs off the PLAYER
// and the counter ("for as long as they have a contract counter"): the
// condition reads the counter as they leave, so a player whose counter was
// removed first does not trigger it. It fires once (CR 603.7c).
//
// It fires on game.EventPlayerEliminated, which the engine emits before CR
// 800.4a takes the player's objects out of the game, and it resolves
// afterwards; "each artifact and creature they controlled" is the
// engine's record of their battlefield as they left
// (game.Player.DepartedPermanents), each copied as it was then (CR 707.2)
// and typed as it was then (CR 608.2h).

// WhenPlayerLosesWithCounter schedules the delayed trigger: when `Player`
// loses the game while they have at least one `Counter` counter, `Body`
// resolves under the scheduling ability's controller with Params.Player
// set to the player who lost. Label is the stack label.
type WhenPlayerLosesWithCounter struct {
	Player  uuid.UUID
	Counter string
	Label   string
	Body    game.BodyRef
}

func (w WhenPlayerLosesWithCounter) Apply(ctx *Context) error {
	if w.Player == uuid.Nil || w.Counter == "" {
		return nil
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        w.Label,
		On:           []game.EventKind{game.EventPlayerEliminated},
		Condition:    playerLosesWithCounterCondition,
		CondParams:   game.EffectParams{Player: w.Player, Name: w.Counter},
		Body:         w.Body,
		Params:       game.EffectParams{Player: w.Player},
		Duration:     ptrDuration(game.IndefiniteDuration()),
	})
	return nil
}

func ptrDuration(d game.Duration) *game.Duration { return &d }

// playerLosesWithCounterCondition is the trigger event: the marked player
// left the game holding the counter (CR 122.1: a counter on a player is
// read off the player). An on-disk key.
var playerLosesWithCounterCondition = game.DelayedCondition("departed/player-loses-with-counter",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
		if ev.Kind != game.EventPlayerEliminated || ev.Actor != p.Player {
			return false
		}
		pl := g.PlayerByIDForEffect(p.Player)
		return pl != nil && pl.Counters[p.Name] > 0
	})

// TokenCopiesOfDepartedArtifactsAndCreaturesBody is "for each artifact and
// creature they controlled, create a token that's a copy of it", under the
// trigger's controller, for the player in Params.Player. Every copy is
// created at once (CR 608.2f). An on-disk key.
var TokenCopiesOfDepartedArtifactsAndCreaturesBody = game.DelayedBody("departed/token-copies-of-artifacts-and-creatures",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		return CreateTokenCopiesOfDeparted{Player: p.Player, Types: []string{"artifact", "creature"}}.Apply(NewContext(g, item))
	})

// CreateTokenCopiesOfDeparted creates, under Controller (the context's
// controller when Nil), one token copy of each permanent Player controlled
// when they left the game that had any of Types (all of them when Types is
// empty), as one creation (CR 608.2f). A player still in the game has no
// record, so nothing is created.
type CreateTokenCopiesOfDeparted struct {
	Controller uuid.UUID
	Player     uuid.UUID
	Types      []string
}

func (c CreateTokenCopiesOfDeparted) Apply(ctx *Context) error {
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	var groups []game.TokenGroup
	for _, d := range ctx.Game.DepartedPermanentsForEffect(c.Player) {
		if !departedHasAnyType(d, c.Types) {
			continue
		}
		groups = append(groups, game.TokenGroup{Template: CopiedToken(TokenCopyTemplateFromCard(d.Source)), Count: 1})
	}
	if len(groups) == 0 {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: controller, Source: ctx.Source(), Groups: groups, Together: true,
	}, game.KeyedTail{})
}

func departedHasAnyType(d game.DepartedPermanent, types []string) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if d.HasType(t) {
			return true
		}
	}
	return false
}
