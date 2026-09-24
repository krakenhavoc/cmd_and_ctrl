package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prepare.go — the catalog's two words for preparation cards (CR 722,
// ADR 0090). The engine owns the designation, the copy in exile, the
// cast and its sweep (game/prepare.go); a card file only ever says
// WHEN its permanent becomes prepared, and what its prepare spell
// does.
//
// A preparation card is two registrations, the adventure shape:
//
//	Register(Spec{OracleID: id,        …})  // the permanent — face 0
//	Register(Spec{OracleID: id + "#1", …})  // the prepare spell — face 1
//
// The prepare spell's Spec is an ordinary instant or sorcery Spec. It
// is never cast from the card (CR 722.3); the copy the engine casts
// out of exile carries the card's oracle ID with face 1 active, so it
// resolves through the "#1" entry exactly as an Adventure half does.

// SelfEntersPrepared is "This creature enters prepared" (CR 722.3a) —
// a CR 614.1d replacement on the permanent's own entry, the
// SelfEntersTapped shape. The landing gives the permanent the
// designation and makes the copy of its prepare spell in exile before
// EventETB, so the permanent is never on the battlefield unprepared
// and an ETB trigger already finds it prepared.
func SelfEntersPrepared() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersPrepared = true
			return nil
		},
	}
}

// BecomePrepared is "[this creature / target creature] becomes
// prepared" (CR 722.3a). Target is the permanent.
//
// Not an error when nothing happens, because CR 722.3a says nothing
// does: a permanent with no prepare spell and a permanent that is
// already prepared both simply stay as they are, and a permanent that
// left the battlefield before the ability resolved is not there to
// become anything (CR 400.7).
type BecomePrepared struct {
	Target uuid.UUID
}

func (b BecomePrepared) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(b.Target) { // #1432
		return nil
	}
	if _, err := ctx.Game.BecomePreparedForEffect(b.Target); err != nil && !errors.Is(err, game.ErrCardNotFound) {
		return err
	}
	return nil
}
