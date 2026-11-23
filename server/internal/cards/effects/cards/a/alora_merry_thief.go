package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alora, Merry Thief — Legendary Creature — Halfling Rogue {2}{U}, 3/2
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Whenever you attack, up to one target attacking creature can't be
//	 blocked this turn. Return that creature to its owner's hand at the
//	 beginning of the next end step.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// "Whenever you attack" is one trigger per attack declaration (Bast's
// OncePerBatch over EventAttack). The target is an attacking creature, any
// player's, picked as the trigger goes on the stack. At resolution a target
// that is no longer an attacking creature makes the whole ability fizzle
// (CR 608.2b); otherwise it can't be blocked this turn (Minamo Sightbender's
// RestrictUntilEOT) and a delayed trigger bounces it at the beginning of the
// next end step (Sakashima's body: a card that has left the battlefield by
// then is skipped).
//
// Choose a Background is a deck-construction rule (CR 702.124k) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	const label = "Alora, Merry Thief — up to one target attacking creature can't be blocked this turn; return it to its owner's hand at the beginning of the next end step"
	effects.Register(effects.Spec{
		OracleID:     "a26432ff-ae60-454d-a6de-fec3feafa669",
		Name:         "Alora, Merry Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Alora, Merry Thief — can't be blocked",
						}).Apply(ctx); err != nil {
							return err
						}
						return effects.ScheduleDelayedTrigger{
							Label: "Alora, Merry Thief — return it to its owner's hand",
							Cards: []uuid.UUID{t.ID},
							Body:  effects.S82GapB221ReturnListedToHandBody,
						}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetCreature("up to one target attacking creature", effects.AttackingCreature()).WithCount(0, 1))),
		},
	})
}
