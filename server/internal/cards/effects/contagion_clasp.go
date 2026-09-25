package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contagion Clasp — Artifact {2}:
//
//	"When this artifact enters, put a -1/-1 counter on target
//	 creature.
//	 {4}, {T}: Proliferate."
//
// The card the proliferate auto-pick was written against: the Clasp
// puts a -1/-1 counter on something of THEIRS, and every later
// activation grows it. That is the whole engine, and it is also the
// asymmetry the pick has to get right — proliferating this board
// must choose the opponent's shrinking creature and must not choose
// your own creature if it happens to be carrying a -1/-1 counter
// too.
//
// The ETB is a targeted trigger rather than an OnETB hook: it
// targets, so it needs the announce-time legality check and the
// resolution re-check a trigger's target clause gets. With no
// creature on the board the trigger is dropped entirely (CR 603.3d)
// rather than prompting for an impossible target.
func init() {
	Register(Spec{
		OracleID:     "43f2d81e-aa01-4fa9-9046-6a27a05dbd2d",
		Name:         "Contagion Clasp",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You don't choose what to proliferate — the game picks for you, adding every counter that helps you and every counter that hurts an opponent."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("target creature"),
			Key:     "Contagion Clasp — put a -1/-1 counter on target creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
					return nil
				}
				return AddCounter{
					Target: item.Targets[0].ID,
					Kind:   game.CounterMinusOne,
					N:      1,
				}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}: Proliferate",
			Cost:  Plus(ManaCost("{4}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Proliferate{}.Apply(NewContext(g, item))
			},
		}},
	})
}
