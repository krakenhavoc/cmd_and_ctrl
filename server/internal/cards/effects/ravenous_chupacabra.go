package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Chupacabra — Creature — Beast Horror {2}{B}{B}, 2/2:
//
//	"When this creature enters, destroy target creature an opponent
//	controls."
//
// Unconditional creature removal stapled to a body, and the removal
// is repeatable with any blink or recursion — which is why it's a
// staple in every black deck that can cast it.
//
// MANDATORY, unlike Reclamation Sage: no OptionalPrompt, so the
// trigger always fires when a legal target exists. When none does the
// engine drops the trigger entirely (CR 603.3d) rather than prompting,
// so an empty opponent board is a clean no-op.
//
// "an opponent controls" is enforced by the predicate, so your own
// creatures — including the Chupacabra itself — are never offered.
func init() {
	Register(Spec{
		OracleID: "7b459306-149b-4f43-abc1-2dd70c748c0e",
		Name:     "Ravenous Chupacabra",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return destroyChosenTargetTrigger(source, "Ravenous Chupacabra — destroy target creature an opponent controls")
			},
		}},
	})
}
