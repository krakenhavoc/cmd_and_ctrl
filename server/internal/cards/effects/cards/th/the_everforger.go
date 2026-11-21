package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Everforger — Legendary Artifact Creature — Construct {6}, 4/5 (Mystery
// Booster Commander, #5948):
//
//	"Rulebreaker — A deck with this commander can have artifact creature and
//	 Equipment cards of any color identity and any basic land cards.
//	 Whenever you cast an artifact creature or Equipment spell, you may copy
//	 that spell. Do this only once each turn. (The copy becomes a token.)"
//
// The deck clause is read off the commander's text by internal/deck
// (rulebreaker.go). The trigger is Wandering Archaic's copy of a spell that
// names it without targeting it (CR 707.10, from last-known information when
// it has left the stack), made by The Everforger's controller; the copy of a
// permanent spell becomes a token as it resolves. "Do this only once each
// turn" is Riveteers Ascendancy's record: the turn is marked only when a copy
// is really made, so a "no" leaves a later spell free to be copied, and a
// second trigger already on the stack does nothing once the first has copied.
//
// No simplification.
func init() {
	const label = "The Everforger — you may copy that spell (once each turn)"
	effects.Register(effects.Spec{
		OracleID:     "b8354ac6-5285-43dd-9ba5-93064def2e96",
		Name:         "The Everforger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || g.ReplacementAppliedThisTurnForEffect(source, everforgerDone) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && ((c.IsArtifact() && c.IsCreature()) || c.HasSubtype("Equipment"))
			}, label, everforgerAsk),
		},
	})
}

func everforgerAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || item.Trigger == nil || g.ReplacementAppliedThisTurnForEffect(&src, everforgerDone) {
		return nil
	}
	spellID := item.Trigger.Event.CardID
	// The question does not stop the table, so the spell may resolve before the
	// answer: hold its last-known information for the copy (#4276).
	g.HoldSpellForEffect(spellID)
	return effects.MayChoice{
		Player:   item.Controller,
		Question: "The Everforger — copy that spell?",
		Then:     everforgerCopy,
		Carry:    []uuid.UUID{spellID},
	}.Apply(ctx)
}

const everforgerDone = "The Everforger — you may copy that spell (once each turn) (done this turn)" // label + " (done this turn)"

var everforgerCopy = effects.MayAnswer("confirm/the-everforger-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || ctx.Game.ReplacementAppliedThisTurnForEffect(&src, everforgerDone) {
		return nil
	}
	ctx.Game.MarkReplacementAppliedThisTurnForEffect(&src, everforgerDone)
	return effects.CopySpell{StackID: r.Carry[0], Controller: ctx.Controller(), FromLastKnown: true}.Apply(ctx)
})
