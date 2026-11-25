package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lucy MacLean, Positively Armed — Legendary Creature — Human Survivor
// {4}{W}, 3/5 (Fallout):
//
//	"Golden Rule — Whenever a token enters, you may have target player
//	 other than its controller create a token that's a copy of it, then you
//	 draw a card if an opponent created a token this way. Do this only once
//	 each turn."
//
// One trigger per token entering, under anyone's control. The target is a
// player other than the entering token's controller, read as the ability
// goes on the stack (a TargetsFrom clause over the triggering event). The
// "you may" is asked as it resolves (CR 603.5); "yes" has the target player
// create the copy, so that player is its owner and controller (CR 111.2),
// copying the token as it is now or as it last existed if it has left
// (CR 707.2, 608.2h). The draw follows the creation (a keyed tail) and
// happens only when a token was created and its creator is one of Lucy's
// controller's opponents (a teammate's copy draws nothing).
//
// "Do this only once each turn" limits the action, not the trigger, as
// Terrasymbiosis's does: the first "yes" uses up the turn and a "no" does
// not. The record is per object (NoteDoneThisTurn), and it is also read as
// the ability triggers, so no trigger is made once the turn's copy is made.
//
// No simplification.
func init() {
	t := effects.On(game.EventETB, lucyMacLeanATokenEntered, lucyMacLeanLabel, lucyMacLeanCopy)
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return when(ev, source, ch, g) && !effects.B28DoneThisTurn(g, source.InstanceID, lucyMacLeanDone)
	}
	t = effects.WithTargetsFrom(t, lucyMacLeanTarget)
	effects.Register(effects.Spec{
		OracleID:     "1196a4cb-544c-4e7b-91cd-f0820b21a80d",
		Name:         "Lucy MacLean, Positively Armed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(t, "Golden Rule — have the target player create a token that's a copy of it?"),
		},
	})
}

const (
	lucyMacLeanLabel = "Golden Rule — target player other than its controller creates a token that's a copy of it, then you draw a card if an opponent created it"
	// lucyMacLeanDone is the once-each-turn record's label.
	lucyMacLeanDone = "Golden Rule — a copy was made this turn"
)

// lucyMacLeanATokenEntered is "whenever a token enters".
func lucyMacLeanATokenEntered(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsToken()
}

// lucyMacLeanTarget is "target player other than its controller": the
// entering token's controller as the ability goes on the stack.
func lucyMacLeanTarget(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	var owner uuid.UUID
	if c, ok := g.LookupCardForEffect(tc.Event.CardID); ok {
		owner = c.Controller
	} else if tc.Object != nil {
		owner = tc.Object.Controller
	}
	return effects.TargetPlayer("target player other than its controller", func(_ *game.Game, _ uuid.UUID, p *game.Player) bool {
		return p.ID != owner
	})
}

// lucyMacLeanCopy is the "yes": the target player creates a token copy of
// the token that entered, then the draw is judged.
func lucyMacLeanCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.B28DoneThisTurn(g, item.SourceCardID, lucyMacLeanDone) {
		return nil
	}
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			target = t.ID
		}
	}
	if target == uuid.Nil {
		return nil
	}
	effects.B28NoteDoneThisTurn(g, item.SourceCardID, lucyMacLeanDone)
	ref, ok := g.PermanentRefForEffect(ctx.Trigger().Event.CardID)
	if !ok {
		return nil
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
	if !ok {
		return nil
	}
	then := ctx.Tail(lucyMacLeanDraw)
	then.Carry = []uuid.UUID{target}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: target, Source: ctx.Source(),
		Groups: []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: 1}},
	}, then)
}

// lucyMacLeanDraw is "then you draw a card if an opponent created a token
// this way".
var lucyMacLeanDraw = effects.CardTail("lucy-maclean/draw-if-an-opponent-created", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	creator := o.Tail.Carry[0]
	for _, opp := range ctx.Opponents() {
		if opp == creator {
			return effects.DrawCards{N: 1}.Apply(ctx)
		}
	}
	return nil
})
