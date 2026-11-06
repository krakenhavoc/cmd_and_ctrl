package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Lord Azula — Legendary Creature — Human Noble {1}{U}{B}{R}, 4/4:
//
//	"Firebending 2 (Whenever this creature attacks, add {R}{R}. This
//	 mana lasts until end of combat.)
//	 Whenever you cast a spell while Fire Lord Azula is attacking, copy
//	 that spell. You may choose new targets for the copy. (A copy of a
//	 permanent spell becomes a token.)"
//
// Firebending is Avatar Roku's. The cast trigger is Reflections of
// Littjara's "copy that spell" (CR 707.10, with CR 707.10c's new
// targets), gated on Azula being an attacking creature as the spell is
// cast. The copy is made from last-known information, so it is created
// even if the spell was countered in response (the ruling); X, and the
// effects of additional costs paid, are copied, while choices made on
// resolution are made separately for the copy. A copy is not cast.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "2631130e-63e2-43a3-914a-648e76722203",
		Name:         "Fire Lord Azula",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fire Lord Azula — firebending 2", Do(firebending(2))),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && source.IsCreature() && source.AttackingTarget != uuid.Nil
			}, "Fire Lord Azula — copy that spell", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CopySpell{StackID: ctx.Trigger().Event.CardID, ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
			}),
		},
	})
}
