package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vedalken Infuser — Creature — Vedalken Wizard {3}{U}, 1/4 (Mirrodin
// Besieged, S82 OLD batch 464, #5576):
//
//	"At the beginning of your upkeep, you may put a charge counter on
//	 target artifact."
//
// The target is chosen as the trigger goes on the stack; the "may" is
// asked as it resolves, and only for a target still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68cce1e6-1043-41b2-abd6-7e27ae0125e3",
		Name:         "Vedalken Infuser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Vedalken Infuser — you may put a charge counter on target artifact",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.MayChoice{
							Question: "Vedalken Infuser — put a charge counter on " + effects.CardNameFor(g, id) + "?",
							Then:     vedalkenInfuserCounter,
							Carry:    []uuid.UUID{id},
						}.Apply(ctx)
					}),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}

var vedalkenInfuserCounter = effects.MayYes("confirm/s82-old-b464-vedalken-infuser-counter", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	id := effects.CarryAt(r, 0)
	if id == uuid.Nil || !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterCharge, N: 1}.Apply(ctx)
})
