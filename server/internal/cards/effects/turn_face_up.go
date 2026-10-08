package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_face_up.go — #2590, ADR 0082's 2026-10-07 second amendment: an
// EFFECT that turns a face-down permanent face up, for free.
//
// morph.go and turn_face_down.go cover the keywords that put a card
// face down and the four cards that turn a face-up permanent over.
// This is the missing direction of the second: "turn it face up" as
// an instruction, not the CR 116.2g special action that pays a morph
// or manifest cost. Hauntwoods Shrieker, Zimone, Mystery Unraveler and
// Experimental Lab // Staff Room are the first cards to say it.
//
// Nothing in here is a rule. The copiable values reverting, the object
// staying the same object, the timestamp, the knowers and the "when
// turned face up" triggers are all the engine's, behind
// game.TurnFaceUpForEffect, which shares its tail with the special
// action. What a card file says is only WHICH permanent and WHO did it.

// TurnFaceUp turns the named battlefield permanents face up as part of
// an effect, paying nothing (CR 708.8, CR 701.40b).
//
// A permanent that is not on the battlefield any more, is not face
// down, or that the rules refuse (CR 701.40b: a manifested or cloaked
// card that is not a creature card stays face down) is skipped. Those
// are "nothing happens" in the rules, so none is an error here: the
// ability that asked has still resolved.
//
// `Actor` is the player whose effect it is and defaults to the
// effect's controller, not the permanent's: the Shrieker turns an
// opponent's creature card up. `Source` is the object doing it and
// defaults to the effect's source; the log names it.
type TurnFaceUp struct {
	Source  uuid.UUID
	Actor   uuid.UUID
	Targets []uuid.UUID
}

func (t TurnFaceUp) Apply(ctx *Context) error {
	source := t.Source
	if source == uuid.Nil {
		source = ctx.Source()
	}
	actor := t.Actor
	if actor == uuid.Nil {
		actor = ctx.Controller()
	}
	for _, id := range ctx.withoutNewSourceObject(t.Targets) {
		ctx.Game.TurnFaceUpForEffect(source, actor, id)
	}
	return nil
}

// canTurnFaceUpByEffect reads the permanent `id` right now and asks
// the engine's one rule for it (game.CanTurnFaceUpForEffect). False for
// a permanent that is gone, is face up, or is a manifested non-creature
// card. A card that offers a CHOICE (Staff Room) asks this before it
// asks the question, so it never offers a turn that would do nothing.
func canTurnFaceUpByEffect(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && game.CanTurnFaceUpForEffect(c)
}

// faceDownPermanentsThatCanTurnUp lists, in battlefield order, the
// face-down permanents `player` controls that an effect could turn
// face up. Zimone's "a permanent you control".
func faceDownPermanentsThatCanTurnUp(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && game.CanTurnFaceUpForEffect(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
