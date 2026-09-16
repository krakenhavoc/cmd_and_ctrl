package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overseer of the Damned — Creature — Demon {5}{B}{B}, 5/5 (EDHREC
// rank 3512):
//
//	"Flying
//	 When this creature enters, you may destroy target creature.
//	 Whenever a nontoken creature an opponent controls dies, create a
//	 tapped 2/2 black Zombie creature token."
//
// A removal spell on a body that turns every opposing death into a
// Zombie — its own entry kill included, when the target was a
// nontoken creature an opponent controlled. The entry trigger is
// optional and targeted (the CR 603.3d prompt, then the pick); the
// dies trigger is diedCreature narrowed to a nontoken creature under
// another player's control, so an opponent's token dying makes
// nothing and one of the controller's own creatures dying makes
// nothing. The Zombie enters tapped through the token entry options.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1f085ce-5b52-45b3-aef0-f77f36b3da36",
		Name:            "Overseer of the Damned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetCreature("target creature"),
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: "Overseer of the Damned — destroy target creature?",
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Overseer of the Damned — destroy target creature", destroyFirstLegalTarget)
				},
			},
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33OpponentsNontokenCreatureDied(ev, source, g)
			}, "Overseer of the Damned — create a tapped 2/2 black Zombie", createTappedZombie),
		},
	})
}
