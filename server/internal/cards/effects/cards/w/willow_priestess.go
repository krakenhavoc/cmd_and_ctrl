package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Willow Priestess — Creature — Faerie Druid {2}{G}{G}, 2/2 (Homelands, S82 OLD
// b821, #5635):
//
//	"{T}: You may put a Faerie permanent card from your hand onto the battlefield.
//	{2}{G}: Target green creature gains protection from black until end of turn."
//
// Goblin Lackey's put from hand with a tap cost, and Obsidian Acolyte's
// protection grant (CR 702.16) on a green creature.
//
// No simplification.
func init() {
	const putLabel = "Willow Priestess — you may put a Faerie permanent card from your hand onto the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "630eab2b-c1fd-46bb-8ae0-ff837e46db1c",
		Name:         "Willow Priestess",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: You may put a Faerie permanent card from your hand onto the battlefield.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.Subtype("Faerie"),
						Optional: true,
						Label:    putLabel,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{G}: Target green creature gains protection from black until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, game.ProtectionFromColor("B"))),
				Cost:    effects.ManaCost("{2}{G}"),
				Targets: effects.TargetCreature("target green creature", effects.OfColor("G")),
				Effect:  effects.EbTargetCreatureGainsUntilEOT(game.ProtectionFromColor("B"), "Willow Priestess — protection from black until end of turn"),
			},
		},
	})
}
