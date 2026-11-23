package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "80e5d768-cc1f-48f8-b0fe-ed974e4d1a56",
		Name:         "Vault Plunderer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Vault Plunderer — target player draws a card and loses 1 life",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return effects.DrawCards{Player: t.ID, N: 1}.Apply(ctx)
							}
						}
						return nil
					}), effects.TargetPlayerLosesLife{N: 1})),
				effects.TargetPlayer("target player")),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1})),
		},
	})
}
