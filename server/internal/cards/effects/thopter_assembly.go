package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Assembly — Artifact Creature — Thopter {6}, 5/5:
//
//	"Flying
//	 At the beginning of your upkeep, if you control no Thopters other
//	 than this creature, return this creature to its owner's hand and
//	 create five 1/1 colorless Thopter artifact creature tokens with
//	 flying."
//
// An intervening "if" (CR 603.4): checked when the upkeep begins and
// again as the trigger resolves, so a Thopter that arrives in response
// stops it. The tokens are made even if the Assembly has already left
// (it does as much as it can, CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "007c110d-4364-41e5-a8bc-6ac05ecbbd00",
		Name:            "Thopter Assembly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && s82RprB09NoOtherThopter(g, source.Controller, source.InstanceID)
			}, "Thopter Assembly — return this creature to its owner's hand and create five 1/1 Thopters",
				func(g *game.Game, item *game.StackItem) error {
					if !s82RprB09NoOtherThopter(g, item.Controller, item.SourceCardID) {
						return nil
					}
					return thopterAssemblyBody(g, item)
				}),
		},
	})
}

var thopterAssemblyBody = Do(
	Instruction(func(ctx *Context) error {
		return BounceToHand{Target: ctx.Item.SourceCardID}.Apply(ctx)
	}),
	CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 5},
)
