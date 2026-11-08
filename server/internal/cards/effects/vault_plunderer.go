package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault Plunderer — Creature — Human Rogue {2}{B}, 3/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, target player draws a card and loses 1
//	 life."
//
// Bloodgift Demon's trigger on entry: "target player", so the card can
// go to whoever the controller likes, and the 1 life comes off whoever
// draws. A target that has left the game by resolution does neither
// (CR 608.2b). The life loss is a step of its own, so it waits for any
// prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80e5d768-cc1f-48f8-b0fe-ed974e4d1a56",
		Name:         "Vault Plunderer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Vault Plunderer — target player draws a card and loses 1 life",
					Do(Instruction(func(ctx *Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
							}
						}
						return nil
					}), TargetPlayerLosesLife{N: 1})),
				TargetPlayer("target player")),
				ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1})),
		},
	})
}
