package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inferno Titan — Creature — Giant {4}{R}{R}, 6/6 (EDHREC rank 3118):
//
//	"{R}: This creature gets +1/+0 until end of turn.
//	 Whenever this creature enters or attacks, it deals 3 damage
//	 divided as you choose among one, two, or three targets."
//
// The red Titan: three points of removal on the way in and three more
// every attack. The trigger is ONE printed ability with two trigger
// conditions, so it is one TriggeredAbility watching both EventETB and
// EventAttack (Sun Titan's shape).
//
// The division is the controller's (#1563, CR 601.2d via CR 603.3d):
// the pick_target prompt asks for one to three targets and each one's
// share, every target at least 1 and the shares summing to 3. A target
// that leaves in response takes nothing and its share is lost, not
// moved to the others (CR 608.2b).
//
// The firebreathing is an ordinary activated ability, pumping the
// ability's own source (item.SourceCardID, CR 113.7a).
func init() {
	Register(Spec{
		OracleID:     "0ce47c8b-1e1f-463f-94f0-35ca00be89e6",
		Name:         "Inferno Titan",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Inferno Titan — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b21SelfEnteredOrAttacked(ev, source)
			},
			Targets: TargetAny().WithCount(1, 3).Dividing(Divide(3)),
			Key:     "Inferno Titan — 3 damage divided among one, two, or three targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
