package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint Hawk — Creature — Bird {W}, 2/2 (Scars of Mirrodin, S82 OLD batch 472,
// #5577):
//
//	"Flying
//	 When this creature enters, sacrifice it unless you return an artifact you
//	 control to its owner's hand."
//
// Quickling's enter trigger (SacrificeUnlessYouReturnOne) over the artifacts
// you control: with none it is sacrificed at once, otherwise the controller
// names one to return or none, which sacrifices the Hawk. It is not an
// artifact itself, so it can't return itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d24525f-dbac-434b-bda7-ff8af55245ff",
		Name:            "Glint Hawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glint Hawk — sacrifice it unless you return an artifact you control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificeUnlessYouReturnOne(effects.NewContext(g, item), item,
						effects.OldB472ArtifactsYouControl(g, item.Controller),
						"Glint Hawk — return an artifact you control to its owner's hand (or sacrifice Glint Hawk)", glintHawkPicked)
				}),
		},
	})
}

var glintHawkPicked = effects.AskedPick("choose-cards/s82-old-b472-glint-hawk-return", func(ctx *effects.Context, r game.CardsChosen) error {
	return effects.ReturnPickedOrSacrifice(ctx, r, glintHawkBounced)
})

var glintHawkBounced = effects.CardTail("s82-old-b472/glint-hawk-after-bounce", effects.SacrificeSourceUnlessBounced)
