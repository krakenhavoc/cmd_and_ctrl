package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// transform.go — the card-facing half of CR 701.27 (ADR 0079).
//
// Two primitives, and picking the wrong one is the mistake the ADR was
// written to prevent, because both compile and both put a permanent on
// its other face:
//
//	Transform / TransformThis        the permanent turns over IN PLACE.
//	                                 Same object (CR 712.18): counters,
//	                                 damage, attachments, the CR 613.7
//	                                 timestamp and summoning sickness
//	                                 all stay, and nothing re-triggers.
//
//	ExileAndReturnTransformed        "exile it, then return it to the
//	                                 battlefield transformed" — TWO zone
//	                                 changes and a NEW object (CR 400.7):
//	                                 counters and damage fall off, the
//	                                 timestamp is fresh, it is summoning
//	                                 sick again, and every ETB on the new
//	                                 face fires.
//
// Read the printed card. "Transform Storm the Vault" is the first;
// "Exile this Saga, then return it to the battlefield transformed under
// your control" is the second.

// Transform turns the named permanent over (CR 701.27a).
//
// A permanent that cannot transform — single-faced, an adventure or
// split card, a meld card, a face-down permanent, or one whose other
// face is an instant or sorcery — does NOTHING, and that is not an
// error (CR 701.27c, CR 701.27d). So a "transform target creature"
// effect aimed at an ordinary creature resolves and changes nothing,
// exactly as it does in paper.
type Transform struct {
	Target uuid.UUID
}

func (t Transform) Apply(ctx *Context) error {
	if t.Target == uuid.Nil {
		return nil
	}
	return ctx.Game.TransformPermanentForEffect(t.Target)
}

// TransformThis turns the SOURCE permanent over — "transform Storm the
// Vault", "transform Aang", the back half of every printed
// transforming permanent's own trigger or ability.
//
// Reads the source off the Context rather than capturing a *game.Card,
// for the reason every trigger Effect does: undo restores a cloned
// game and a captured pointer resolves against the wrong one.
type TransformThis struct{}

func (TransformThis) Apply(ctx *Context) error {
	return Transform{Target: ctx.Source()}.Apply(ctx)
}

// ExileAndReturnTransformed is the OTHER shape: "exile this permanent,
// then return it to the battlefield transformed under your control"
// (Fable of the Mirror-Breaker's chapter III, CR 712.14a).
//
// Leave Controller zero for "under its owner's control"; set it for
// "under your control", which is what the printed cards say.
//
// The exile can PAUSE — a commander is offered the command zone
// (CR 903.9) — so the return is a continuation of it rather than the
// next line, exactly as Flicker's is (#894). If the card never reaches
// exile, nothing returns and the permanent stays where the replacement
// put it: the sentence's second half is conditional on its first.
type ExileAndReturnTransformed struct {
	Target     uuid.UUID
	Controller uuid.UUID
}

func (e ExileAndReturnTransformed) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	return ctx.Game.ExileAndReturnTransformedForEffect(e.Target, e.Controller)
}

// TransformedPermanent reports whether the card is a "transformed
// permanent" in the CR 701.27g sense: a double-faced permanent on the
// battlefield with its back face up.
//
// Exported as a predicate rather than left as `c.ActiveFace != 0` at
// each call site because the rule has a second half that is easy to
// drop — "a permanent with its front face up is never considered a
// transformed permanent, even if it had its back face up previously" —
// and because the zone matters: CR 712.8a puts every card outside the
// battlefield and the stack on its front face anyway, so a card that
// reads non-zero off the battlefield is a bug rather than a
// transformed permanent.
func TransformedPermanent(c game.Card) bool {
	return c.ActiveFace != 0 && len(c.Faces) > 1
}
