package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const sizzlingSoloistLabel = "Sizzling Soloist — alliance: target creature an opponent controls can't block this turn; the second time this resolves this turn, it attacks during its controller's next combat phase if able"

// Sizzling Soloist — Creature — Human Citizen {3}{R}, 3/2 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, target creature
//	 an opponent controls can't block this turn. If this is the second time
//	 this ability has resolved this turn, that creature attacks during its
//	 controller's next combat phase if able."
//
// The can't-block is Ahn-Crop Crasher's restriction. The "second time" is
// Captain Michael Burnham's resolution tally. The attack requirement is
// pinned to the creature (CR 400.7) and read live against its controller, as
// Gideon, Battle-Forged's is: it lasts through that controller's next turn,
// or to the end of this turn when the creature's controller is the active
// player and combat has not begun.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8975d81-d373-49b1-be00-9513b92877d9",
		Name:         "Sizzling Soloist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl, sizzlingSoloistLabel,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: sizzlingSoloistLabel}).Apply(ctx); err != nil {
							return err
						}
						if !effects.SncSecondResolutionThisTurn(g, item, sizzlingSoloistLabel) {
							return nil
						}
						return effects.SncCreatureAttacksDuringItsControllersNextCombat(ctx, id,
							"Sizzling Soloist — attacks during its controller's next combat phase if able")
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
