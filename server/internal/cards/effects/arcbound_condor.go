package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcbound Condor — Artifact Creature — Bird {2}{B}{B}, 0/0 (Modern Horizons
// 3, #5948):
//
//	"Flying
//	 Modular 3 (This creature enters with three +1/+1 counters on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)
//	 Whenever another artifact you control enters, target creature an
//	 opponent controls gets -1/-1 until end of turn."
//
// Modular is the engine's keyword (game/modular.go, CR 702.43a): the printed
// token carries both halves. The trigger is Machinesmith Automaton's artifact
// watch (enteredUnderYourControl, another) with a target the controller chooses
// as it goes on the stack: with no creature an opponent controls the trigger
// is removed (CR 603.3d), and one that has left by resolution does nothing
// (CR 608.2b). The Condor does not trigger on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d484be42-5dd4-49e0-afcd-2d3c1a2c3037",
		Name:            "Arcbound Condor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "modular 3"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && c.IsArtifact()
				},
				"Arcbound Condor — target creature an opponent controls gets -1/-1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: -1, Toughness: -1,
						Label: "Arcbound Condor — -1/-1 until end of turn"}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
