package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sower of Temptation — 2/2 Faerie Wizard for {2}{U}{U}:
//
//	"Flying
//	 When this creature enters, gain control of target creature for as
//	 long as this creature remains on the battlefield."
//
// The CR 611.2b card: a duration that is a CONDITION rather than a
// clock. The engine re-evaluates it on every layer pass, and every
// board change already bumps the layer version, so the creature goes
// home the moment the Sower does — killed in response to the trigger,
// killed years later, bounced, exiled, sacrificed.
//
// Two subtleties the duration model gets right and a "remember the
// Sower's instance ID" implementation would not:
//
//   - **Flicker gives the creature back (CR 400.7).** A Sower blinked
//     out and back is a NEW object, so the effect that watched the old
//     one ends. The condition is keyed on instance ID AND
//     battlefield-entry stamp, so the returning Sower does not inherit
//     the theft — it just enters and its own ETB trigger goes on the
//     stack, taking a creature afresh.
//   - **The Sower dying in response to its own trigger takes nothing.**
//     CR 611.2b: an effect whose condition is already false as it
//     would begin never begins. The duration constructor returns false
//     and nothing is registered, rather than an effect that would be
//     swept on its first recompute after having briefly applied.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "5881aac2-4f31-45cb-bdb5-64a29ec23316",
		Name:            "Sower of Temptation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("target creature"),
			Key:     "Sower of Temptation — gain control of target creature for as long as this remains",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				ctx := NewContext(g, item)
				d, ok := DurationWhileSourceRemains(ctx, item.SourceCardID)
				if !ok {
					// CR 611.2b: the Sower is already gone, so
					// the effect never begins.
					return nil
				}
				return GainControl{
					Target:     item.Targets[0].ID,
					Controller: item.Controller,
					Duration:   d,
					Label:      "Sower of Temptation — control for as long as the Sower remains",
				}.Apply(ctx)
			},
		}},
	})
}
