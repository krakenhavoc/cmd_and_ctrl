package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dundoolin Weaver — {1}{G} Creature — Kithkin Druid 2/1 (Lorwyn Eclipsed,
// #4458):
//
//	"When this creature enters, if you control three or more creatures,
//	 return target permanent card from your graveyard to your hand."
//
// Hologram Stalker's return behind an intervening "if" (CR 603.4), Human
// Torch's shape: the count is read as the Weaver enters, so with fewer
// than three creatures it does not trigger at all, and again as the
// trigger resolves, so losing creatures in response stops the return. The
// Weaver counts itself while it is on the battlefield. The target is chosen
// as the trigger goes on the stack (CR 603.3d), so with no permanent card
// in the graveyard it is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28564d10-bfc5-4beb-b0e2-b9aa5c03f740",
		Name:         "Dundoolin Weaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(CondBoard{
				Name: "you control three or more creatures",
				Test: func(ctx *Context) bool { return dundoolinWeaverThreeCreatures(ctx.Game, ctx.Controller()) },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && dundoolinWeaverThreeCreatures(g, source.Controller)
			}, "Dundoolin Weaver — return target permanent card from your graveyard to your hand", returnTargetGraveyardCardToHand)),
				TargetCardInGraveyard("target permanent card from your graveyard", Permanent(), YouOwn())),
		},
	})
}

// dundoolinWeaverThreeCreatures is "you control three or more creatures".
func dundoolinWeaverThreeCreatures(g *game.Game, controller uuid.UUID) bool {
	return len(b23CreaturesYouControl(g, controller)) >= 3
}
