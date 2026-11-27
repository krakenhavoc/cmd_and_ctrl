package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gustrider Exuberant — Creature — Human Wizard {2}{W}, 1/2 (Shards of
// Alara, S82 OLD batch 510, #5588):
//
//	"Flying
//	 Sacrifice this creature: Creatures you control with power 5 or greater
//	 gain flying until end of turn."
//
// The set is fixed as the ability resolves (CR 611.2c): the creatures with
// power 5 or greater you control then, without the sacrificed Exuberant.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a358fe8f-6ec2-4b42-8851-ef65308ccb1d",
		Name:            "Gustrider Exuberant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Creatures you control with power 5 or greater gain flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.ControlledBy(item.Controller), effects.PowerGE(5)),
					Keywords: []string{"flying"},
					Label:    "Gustrider Exuberant — creatures with power 5 or greater gain flying",
				}.Apply(ctx)
			},
		}},
	})
}
