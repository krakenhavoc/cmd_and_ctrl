package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Woman, Secret Agent — Legendary Creature — Spider Human Spy Hero
// {3}{W/U}, 1/4:
//
//	"Flash
//	 When Spider-Woman enters, tap target creature an opponent controls.
//	 That creature can't become untapped for as long as you control
//	 Spider-Woman."
//
// Dungeon Geists' shape with a stronger lock (#4574): the creature stays
// tapped through every untap, so an untap effect does nothing to it either
// (CR 614.17). The lock is a "can't become untapped" hold timed by CR
// 611.2b's "for as long as you control ~": if Spider-Woman has already left
// or changed control by the time the trigger resolves, the creature is
// tapped and nothing more (TapAndLockWhileYouControlThis).
func init() {
	Register(Spec{
		OracleID:        "04e53f36-9f16-4678-9caf-9afda85f76b4",
		Name:            "Spider-Woman, Secret Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Spider-Woman — tap target creature; it can't become untapped while you control Spider-Woman",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TapAndLockWhileYouControlThis(ctx, holdTargetIDs(ctx))
				}),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		)},
	})
}
