package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Peace in Our Time — Enchantment {3}{W} (Star Trek, #4973):
//
//	"At the beginning of combat on each player's turn, that player may
//	 draw a card. If they do, they can't attack this turn and you put a
//	 peace counter on this enchantment.
//	 At the beginning of your upkeep, if this enchantment has twelve or
//	 more peace counters on it, you win the game."
//
// The combat trigger fires on every player's turn, yours included, and
// asks the active player (Web of Inertia's begin-combat shape). The
// question is a draw (MayChoice.Draws), so a player a draw limit stops
// from drawing is not asked and nothing else happens. On a yes they draw,
// then the creatures they control can't attack for the rest of the turn
// (Oriss, Samite Guardian's scoped rule, read live, so a creature they
// gain control of later this turn can't attack either), and the counter
// goes on this enchantment only while it is still the object that
// triggered (AddCounter, CR 400.7).
//
// The upkeep trigger is an intervening "if" (CR 603.4): twelve or more
// peace counters as your upkeep begins, and again as it resolves, read
// from the enchantment's last-known counters if it has left the
// battlefield in the meantime (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b19ce6f2-7778-4779-8173-973eb7ee0b02",
		Name:         "Peace in Our Time",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, StepBegan(game.StepBeginCombat, false),
				"Peace in Our Time — that player may draw a card; if they do, they can't attack this turn", peaceInOurTimeAsk),
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && source.Counters[peaceCounter] >= 12
			}, "Peace in Our Time — if this enchantment has twelve or more peace counters on it, you win the game", peaceInOurTimeWin),
		},
	})
}

// peaceCounter is the counter kind Peace in Our Time counts.
const peaceCounter = "peace"

// peaceInOurTimeAsk asks the player whose turn it is.
func peaceInOurTimeAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	if player == uuid.Nil {
		return nil
	}
	return MayChoice{
		Player:   player,
		Question: "Peace in Our Time — draw a card? If you do, you can't attack this turn.",
		YesLabel: "Draw, and don't attack this turn",
		NoLabel:  "Don't draw",
		Draws:    1,
		Then:     peaceInOurTimeAnswer,
	}.Apply(NewContext(g, item))
}

// peaceInOurTimeAnswer is the yes: the draw, the attack ban on the
// player who drew, and the counter on the enchantment.
var peaceInOurTimeAnswer = MayYes("confirm/peace-in-our-time-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	if err := (DrawCards{Player: r.Chooser, N: 1}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, r.Chooser,
		[]game.Mod{game.AddRestrictionsMod(game.CantAttack)}, DurationUntilEndOfTurn(ctx),
		"Peace in Our Time — that player can't attack this turn")
	return AddCounter{Target: ctx.Source(), Kind: peaceCounter, N: 1}.Apply(ctx)
})

// peaceInOurTimeWin re-checks the counters as the trigger resolves.
func peaceInOurTimeWin(g *game.Game, item *game.StackItem) error {
	n := counterCountOn(g, item.SourceCardID, peaceCounter)
	if !onBattlefield(g, item.SourceCardID) {
		n = g.LastKnownCountersForEffect(item.SourceCardID)[peaceCounter]
	}
	if n < 12 {
		return nil
	}
	return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
}
