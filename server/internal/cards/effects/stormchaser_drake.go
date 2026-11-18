package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormchaser Drake — Creature — Drake {1}{U}, 2/1 (Innistrad: Crimson
// Vow, S82):
//
//	"Flying
//	 Whenever this creature becomes the target of a spell you control,
//	 draw a card."
//
// EventBecomesTarget does not say spell or ability, so the spell half is
// read off the stack (b40TargetedByASpell) as Illuminator Virtuoso does;
// the event's Actor is the targeting spell's controller and must be the
// Drake's controller. An ability you control that targets it does not
// trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9530416-d5fb-4ddd-ae6f-0eefb9f276c5",
		Name:            "Stormchaser Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID &&
					ev.Actor == source.Controller && b40TargetedByASpell(ev, g)
			}, "Stormchaser Drake — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
