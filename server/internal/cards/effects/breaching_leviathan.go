package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breaching Leviathan — Creature — Leviathan {7}{U}{U}, 9/9 (Commander
// 2014, #5553):
//
//	"When this creature enters, if you cast it from your hand, tap all
//	 nonblue creatures. Those creatures don't untap during their
//	 controllers' next untap steps."
//
// An intervening "if" (CR 603.4) read off the cast record (b30CastFromHand,
// Wakening Sun's Avatar's and Deathbringer Regent's condition), checked as
// it enters and again as the trigger resolves: a Leviathan reanimated, put
// onto the battlefield or cast from the command zone does nothing. The
// nonblue creatures are the ones on the battlefield as it resolves, every
// controller's, and each is tapped and held through its controller's next
// untap step (Dragon Turtle's TapAndFreeze), an already tapped one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d249b16-33ef-463e-9e5d-0b84d434f093",
		Name:         "Breaching Leviathan",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && b30CastFromHand(g, source.InstanceID)
			}, "Breaching Leviathan — tap all nonblue creatures; they don't untap during their controllers' next untap steps",
				breachingLeviathanTapNonblue),
		},
	})
}

// breachingLeviathanTapNonblue re-checks the cast from hand, then taps every
// nonblue creature and holds it through its controller's next untap step.
func breachingLeviathanTapNonblue(g *game.Game, item *game.StackItem) error {
	if !b30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(Creature(), Not(OfColor("U")))) {
		ids = append(ids, c.InstanceID)
	}
	return TapAndFreeze{Targets: ids, Label: "Breaching Leviathan — they don't untap during their controllers' next untap steps"}.Apply(ctx)
}
