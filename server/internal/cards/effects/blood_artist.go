package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Artist — 0/1 Creature — Vampire for {1}{B}:
//
//	"Whenever this creature or another creature dies, target player
//	loses 1 life and you gain 1 life."
//
// The aristocrats payoff. Note "this creature OR ANOTHER" — Blood
// Artist triggers on its own death too, and on creatures ANY player
// controls, which is what makes it a board-wipe payoff as well as a
// sacrifice payoff.
//
// The drain is targeted (CR: "target player"), so every death queues
// a pick_target prompt for Blood Artist's controller. That's how the
// card really plays — tedious across a long combo turn, and correct.
func init() {
	Register(Spec{
		OracleID:     "310f141c-7f37-4729-aed6-dd9c09db448d",
		Name:         "Blood Artist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			},
			Targets: TargetPlayer("target player"),
			Key:     "Blood Artist — drain 1",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				if err := g.ChangePlayerLifeForEffect(ctx.Source(), item.Targets[0].ID, -1); err != nil {
					return err
				}
				return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
