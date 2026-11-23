package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathbringer Regent — Creature — Dragon {5}{B}{B}, 5/6:
//
//	"Flying
//	 When this creature enters, if you cast it from your hand and there
//	 are five or more other creatures on the battlefield, destroy all
//	 other creatures."
//
// An intervening "if" (CR 603.4): both halves are checked as the
// ability would trigger and again as it resolves. "Cast from your hand"
// is read off the cast event's origin zone (b30CastFromHand); the other
// creatures are counted across every battlefield, the Regent excluded
// by object. The destroy is one simultaneous event, so dies-triggers
// see the whole batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "71b5fdf2-4475-4102-85eb-9be164a9ae29",
		Name:     "Deathbringer Regent",
		// ADR 0126 §6: only when cast from your hand with five or more other creatures out.
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && deathbringerRegentCondition(g, source.InstanceID)
			}, "Deathbringer Regent — destroy all other creatures", func(g *game.Game, item *game.StackItem) error {
				if !deathbringerRegentCondition(g, item.SourceCardID) {
					return nil
				}
				return effects.DestroyAllMatching{
					Match: effects.And(effects.Creature(), effects.OtherThan(item.SourceCardID)),
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// deathbringerRegentCondition is "you cast it from your hand and there
// are five or more other creatures on the battlefield".
func deathbringerRegentCondition(g *game.Game, self uuid.UUID) bool {
	if !effects.B30CastFromHand(g, self) {
		return false
	}
	return deathbringerRegentOtherCreatures(g, self) >= 5
}

func deathbringerRegentOtherCreatures(g *game.Game, self uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.IsCreature() {
			n++
		}
	}
	return n
}
