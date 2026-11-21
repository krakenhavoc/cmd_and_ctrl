package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hotshot Investigators — Creature — Vedalken Detective {5}{U}, 4/4
// (Murders at Karlov Manor, #4469):
//
//	"When this creature enters, return up to one other target creature
//	 to its owner's hand. If you controlled it, investigate. (Create a
//	 Clue token. It's an artifact with "{2}, Sacrifice this token: Draw
//	 a card.")"
//
// Boomerang Basics' shape on an enters trigger with Azog's "up to one
// other target creature". "If you controlled it" is read BEFORE the
// bounce: in a hand the card has no controller (CR 108.4). The Clue is
// the bounce's continuation, so a returned commander's command-zone
// question is answered first; it does not ask whether the card arrived,
// since the condition is about who controlled the creature, not about
// the move (the line Boomerang Basics draws). With no target, or a
// target gone by resolution, nothing happens and there is no Clue (the
// ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "288fcf5c-f5aa-4f62-9dd8-ba295df48f7a",
		Name:         "Hotshot Investigators",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hotshot Investigators — return up to one other target creature to its owner's hand; investigate if you controlled it",
				hotshotInvestigatorsBounce),
				effects.Another(effects.TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

func hotshotInvestigatorsBounce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	bounce := effects.BounceToHand{Target: id}
	if controller, ok := effects.ControllerOfTarget(ctx, id); ok && controller == ctx.Controller() {
		bounce.ThenTail = ctx.ItemTail(hotshotInvestigatorsClue)
	}
	return bounce.Apply(ctx)
}

// hotshotInvestigatorsClue is "investigate", run once the bounce has
// settled.
var hotshotInvestigatorsClue = effects.CardTail("hotshot-investigators/investigate", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Investigate{N: 1}.Apply(ctx)
})
