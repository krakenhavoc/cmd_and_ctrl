package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Scribe — Creature — Human Artificer {1}{W}, 1/3:
//
//	"Metalcraft — {T}: You get {E} (an energy counter). Activate only if
//	 you control three or more artifacts.
//	 Whenever you get one or more {E} during your turn, creatures you
//	 control get +1/+1 until end of turn."
//
// ADR 0129 §6 (#1995): one trigger per placement of energy during its
// controller's turn (CR 603.2c). The creatures are locked in as the
// trigger resolves (CR 611.2c), so one that arrives later this turn
// gets nothing. The metalcraft gate is Inventor's Fair's
// ControlsAtLeast(3, artifacts).
//
// No simplification.
func init() {
	const label = "Brotherhood Scribe — creatures you control get +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "e22bb590-9c9a-4a99-8cd7-1dbe742f4cdd",
		Name:         "Brotherhood Scribe",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Metalcraft — {T}: You get {E}. Activate only if you control three or more artifacts.",
			Cost:      TapCost(),
			Condition: ControlsAtLeast(3, MatchArtifact),
			Purpose:   game.Purpose{Answers: game.AnswerValue, Energy: 1},
			Effect:    Do(GetEnergy{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventPlayerCounterPlaced, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return YouGotEnergy(ev, source, lki, g) && IsYourTurn(g, source.Controller)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
