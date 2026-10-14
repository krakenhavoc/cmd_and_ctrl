package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreadfeast Demon — Creature — Demon {5}{B}{B}, 6/6:
//
//	"Flying
//	 At the beginning of your end step, sacrifice a non-Demon creature.
//	 If you do, create a token that's a copy of this creature."
//
// The sacrifice is chosen as the trigger resolves (CR 608.2c), over the
// non-Demon creatures its controller controls then, and is mandatory:
// with one on the battlefield the prompt has no "none" answer. With none
// there is nothing to sacrifice, nobody is asked and no token is made.
//
// "If you do" reads what really left (mustSacrificeOne), and the token
// is a copy of the Demon's copiable values (CR 707.2) — counters and
// Auras on it are not copied. Every copy carries the trigger too, so
// two Demons make two sacrifices and two tokens each end step, as
// printed. A Demon that has left the battlefield by then is copied from
// the card where it now sits; a token Demon that has left has ceased to
// exist and is copied from nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aa364a33-7c72-4096-b8ed-ca9d5515c3fe",
		Name:            "Dreadfeast Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Dreadfeast Demon — sacrifice a non-Demon creature, then copy this creature",
				func(g *game.Game, item *game.StackItem) error {
					return mustSacrificeOne(g, item, "Dreadfeast Demon — sacrifice a non-Demon creature",
						And(Creature(), Not(Subtype("Demon"))),
						func(ctx *Context, _ uuid.UUID) error {
							return CreateTokenCopy{
								Controller: ctx.Item.Controller,
								Copy:       ctx.Item.SourceCardID,
								N:          1,
							}.Apply(ctx)
						})
				}),
		},
	})
}
