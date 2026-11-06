package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beetle-Headed Merchants — Creature — Human Citizen {4}{B}, 5/4:
//
//	"Whenever this creature attacks, you may sacrifice another creature or
//	 artifact. If you do, draw a card and put a +1/+1 counter on this
//	 creature."
//
// Automated Warfare System's attack trigger (`maySacrificeOne`, asked at
// resolution over what you control then, the source excluded) with a
// counter beside the draw. The counter goes on this creature only while
// it is still the object that attacked (CR 400.7): one that left and
// came back gets none, though the card is still drawn.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "0e6609b4-f359-4bf1-a06e-f4317a0e7b3c",
		Name:         "Beetle-Headed Merchants",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Beetle-Headed Merchants — you may sacrifice another creature or artifact to draw a card and grow",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Beetle-Headed Merchants — sacrifice another creature or artifact to draw a card and put a +1/+1 counter on this creature?",
						Or(Creature(), Artifact()), true, beetleHeadedMerchantsSacrificed)
				}),
		},
	})
}

var beetleHeadedMerchantsSacrificed = RegisterSacrificedOne("prompt-run/beetle-headed-merchants-draw-and-counter", func(ctx *Context, _ uuid.UUID) error {
	if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
