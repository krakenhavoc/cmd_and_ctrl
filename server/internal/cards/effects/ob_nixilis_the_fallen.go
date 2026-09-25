package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ob Nixilis, the Fallen — Legendary Creature — Demon {3}{B}{B}, 3/3
// (EDHREC rank 1956):
//
//	"Landfall — Whenever a land you control enters, you may have
//	 target player lose 3 life. If you do, put three +1/+1 counters
//	 on Ob Nixilis."
//
// The landfall drain that grows. An optional, targeted landfall
// trigger: the "may" is the prompt, the target is picked when the
// trigger goes on the stack, and "if you do" is read at resolution
// — a target who left the game loses nothing and the counters do
// not come. The counters need Ob Nixilis still on the battlefield;
// a trigger that resolves after he has died still drains the
// player, as printed, and puts counters on nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "154104b7-4125-4276-8bc7-9a6edfe48cb9",
		Name:         "Ob Nixilis, the Fallen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Ob Nixilis, the Fallen — have target player lose 3 life? (He gets three +1/+1 counters.)"},
			Targets:        TargetPlayer("target player"),
			Key:            "Ob Nixilis, the Fallen — target player loses 3 life, three +1/+1 counters",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				ctx := NewContext(g, item)
				if p := ctx.PlayerByID(item.Targets[0].ID); p == nil || p.Eliminated {
					return nil
				}
				if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Targets[0].ID, -3); err != nil {
					return err
				}
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 3}.Apply(ctx)
			},
		}},
	})
}
