package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Restoration Specialist — Creature — Dwarf Artificer {1}{W}, 2/1 (Aether
// Revolt, S82 older sets, #5536):
//
//	"{W}, Sacrifice this creature: Return up to one target artifact card and
//	 up to one target enchantment card from your graveyard to your hand."
//
// Two target clauses, each up to one, in printed order. Each card is
// returned if it is still in the graveyard when the ability resolves
// (CR 608.2b); the sacrifice is a cost, so a countered ability still loses
// the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36f71d49-0f64-4b5d-b1ee-bb0af5e9db9a",
		Name:         "Restoration Specialist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this creature: Return up to one target artifact card and up to one target enchantment card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("up to one target artifact card from your graveyard", effects.YouOwn(), effects.Artifact()).WithCount(0, 1),
				effects.TargetCardInGraveyard("up to one target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment()).WithCount(0, 1),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
			},
		}},
	})
}
